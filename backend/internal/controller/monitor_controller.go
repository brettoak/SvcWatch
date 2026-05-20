package controller

import (
	"SvcWatch/internal/service"
	"SvcWatch/internal/storage"
	"SvcWatch/internal/utils"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nxadm/tail"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins
	},
}

// MonitorController handles HTTP requests for monitor statistics.
type MonitorController struct {
	svc *service.MonitorService
}

// NewMonitorController creates a new instance of MonitorController.
func NewMonitorController(svc *service.MonitorService) *MonitorController {
	return &MonitorController{
		svc: svc,
	}
}

// TimeRangeRequest represents common query parameters for statistics endpoints.
type TimeRangeRequest struct {
	StartTime string `form:"start_time" binding:"required"`
	EndTime   string `form:"end_time"   binding:"required"`
	LogFile   string `form:"log_file"`
}

// ==========================================
// 1. System & Utility Endpoints
// ==========================================

// PingHandler Health Check
// @Summary Health Check
// @Description Returns pong message to check if API is alive
// @Tags 1. System
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/sev/ping [get]
func (ctrl *MonitorController) PingHandler(c *gin.Context) {
	utils.Success(c, gin.H{
		"message": "pong",
	})
}

// ==========================================
// 2. Overview Dashboard Endpoints
// ==========================================

// OverviewHandler Get business overview key metrics
// @Summary Get business overview key metrics
// @Description Get overview statistics with comparison for a time range
// @Tags 2. Overview Dashboard
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start Time" default(2026-03-19 00:00:00)
// @Param end_time query string true "End Time" default(2026-03-20 00:00:00)
// @Param log_file query string false "Log File or Source ID (optional)" default(access.log)
// @Success 200 {object} OverviewResponseWrapper
// @Router /api/v1/sev/overview [get]
func (ctrl *MonitorController) OverviewHandler(c *gin.Context) {
	var req TimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, "start_time and end_time are required")
		return
	}

	if req.LogFile == "" {
		req.LogFile = "access.log"
	}

	if _, _, err := utils.ParseAndValidateRange(req.StartTime, req.EndTime, utils.MaxTimeRangeLimit); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	aggregated, err := ctrl.svc.GetOverview(req.StartTime, req.EndTime, req.LogFile)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, aggregated)
}

// StatsWebSocketHandler handles real-time statistics push via WebSocket
// @Summary Real-time statistics push via WebSocket
// @Description Upgrade connection to WebSocket and stream real-time request counts and error counts. Pushes initial history as 'init' event and periodic updates as 'update' event.
// @Tags 2. Overview Dashboard
// @Param log_file query string false "Log File or Source ID (optional)" default(access.log)
// @Param interval query string false "Refresh interval (e.g., 1s, 2s, 5s)" default(1s)
// @Param simulate query bool false "Enable mock simulation data" default(false)
// @Success 101 {object} StatsWSInitResponse "Switching Protocols (Handshake Success). Initial history data frame pushed immediately."
// @Success 200 {object} StatsWSUpdateResponse "Periodic real-time update data frame pushed at the specified interval."
// @Router /api/v1/sev/stats/ws [get]
func (ctrl *MonitorController) StatsWebSocketHandler(c *gin.Context) {
	logFile := c.Query("log_file")
	sourceIDs := []string{}
	if logFile != "" {
		sourceIDs = append(sourceIDs, logFile)
	}

	intervalStr := c.Query("interval")
	if intervalStr == "" {
		intervalStr = "1s"
	}
	interval, err := time.ParseDuration(intervalStr)
	if err != nil || interval <= 0 {
		interval = 1 * time.Second
	}

	simulate := c.Query("simulate") == "true"

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade to websocket: %v\n", err)
		return
	}
	defer conn.Close()

	var writeMu sync.Mutex

	// done channel signals all goroutines to stop on client disconnect
	done := make(chan struct{})

	// Handle client disconnection
	conn.SetPongHandler(func(string) error { return nil })
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()

	// Heartbeat ping
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeMu.Lock()
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// 1. Send initial history (last 30 points)
	var historyPoints []storage.RealTimeHistoryPoint
	now := time.Now()
	historyLength := 30
	intervalSec := interval.Seconds()
	startT := now.Add(-time.Duration(historyLength) * interval)

	if simulate {
		// Generate random mock history
		historyPoints = make([]storage.RealTimeHistoryPoint, historyLength)
		for i := 0; i < historyLength; i++ {
			ptTime := startT.Add(time.Duration(float64(i)*intervalSec) * time.Second)
			total := 5 + rand.Intn(16) // between 5 and 20
			errors := 0
			if rand.Intn(10) > 7 {
				errors = rand.Intn(3) // 0 to 2 errors
			}
			if errors > total {
				errors = total
			}
			historyPoints[i] = storage.RealTimeHistoryPoint{
				Timestamp: ptTime.Format(time.RFC3339),
				Total:     total,
				Errors:    errors,
			}
		}
	} else {
		// Get real history from SQLite
		historyPoints, err = ctrl.svc.GetRealTimeHistory(startT, now, intervalSec, sourceIDs)
		if err != nil {
			fmt.Printf("Error fetching real-time history: %v\n", err)
			historyPoints = []storage.RealTimeHistoryPoint{}
		}
	}

	initMsg := gin.H{
		"type": "init",
		"data": historyPoints,
	}
	writeMu.Lock()
	err = conn.WriteJSON(initMsg)
	writeMu.Unlock()
	if err != nil {
		return
	}

	// 2. Start ticker for real-time updates
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Track the last pushed time to avoid overlap or missing logs
	lastTime := now

	for {
		select {
		case <-done:
			return
		case tickTime := <-ticker.C:
			var point storage.RealTimeHistoryPoint
			if simulate {
				total := 5 + rand.Intn(16)
				errors := 0
				if rand.Intn(10) > 7 {
					errors = rand.Intn(3)
				}
				if errors > total {
					errors = total
				}
				point = storage.RealTimeHistoryPoint{
					Timestamp: tickTime.Format(time.RFC3339),
					Total:     total,
					Errors:    errors,
				}
			} else {
				// Query database for logs between lastTime and tickTime
				total, errors, err := ctrl.svc.GetRealTimeStats(lastTime, tickTime, sourceIDs)
				if err != nil {
					fmt.Printf("Error querying real-time stats: %v\n", err)
					total = 0
					errors = 0
				}
				point = storage.RealTimeHistoryPoint{
					Timestamp: tickTime.Format(time.RFC3339),
					Total:     total,
					Errors:    errors,
				}
				lastTime = tickTime
			}

			updateMsg := gin.H{
				"type": "update",
				"data": point,
			}
			writeMu.Lock()
			err = conn.WriteJSON(updateMsg)
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// StatusDistributionHandler Get distribution of HTTP status codes
// @Summary Get HTTP status code distribution
// @Description Get distribution of status codes for a time range
// @Tags 2. Overview Dashboard
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start Time" default(2026-03-19 00:00:00)
// @Param end_time query string true "End Time" default(2026-03-20 00:00:00)
// @Param log_file query string false "Log File or Source ID (optional)" default(access.log)
// @Success 200 {object} StatusDistributionResponseWrapper
// @Router /api/v1/sev/distribution [get]
func (ctrl *MonitorController) StatusDistributionHandler(c *gin.Context) {
	var req TimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, "start_time and end_time are required")
		return
	}

	if req.LogFile == "" {
		req.LogFile = "access.log"
	}

	startT, endT, err := utils.ParseAndValidateRange(req.StartTime, req.EndTime, utils.MaxTimeRangeLimit)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	result, err := ctrl.svc.GetStatusDistribution(startT, endT, req.LogFile)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

// LogsWebSocketHandler Real-time logs streaming via WebSocket
// @Summary Real-time logs streaming via WebSocket
// @Description Upgrade connection to WebSocket and stream raw logs in real-time. Pushes standard Nginx log lines as text frames.
// @Tags 2. Overview Dashboard
// @Param log_file query string false "Log File or Source ID (optional)" default(access.log)
// @Success 101 {string} string "Switching Protocols (Handshake Success). Streams raw nginx log lines as text frames."
// @Router /api/v1/sev/logs/ws [get]
func (ctrl *MonitorController) LogsWebSocketHandler(c *gin.Context) {
	logFile := c.Query("log_file")
	if logFile == "" {
		logFile = "access.log"
	}

	// Resolve the actual file path from the monitor configuration
	actualPath := ctrl.svc.ResolveLogPath(logFile)

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Printf("Failed to upgrade to websocket: %v\n", err)
		return
	}
	defer conn.Close()

	var writeMu sync.Mutex

	// Calculate offset for the last 10 lines
	seekInfo := getLastNLinesOffset(actualPath, 10)

	// Configure tail to read from the calculated offset
	config := tail.Config{
		ReOpen:    true,
		Follow:    true,
		MustExist: false,
		Location:  seekInfo,
	}

	t, err := tail.TailFile(actualPath, config)
	if err != nil {
		writeMu.Lock()
		_ = conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error tailing file: %v", err)))
		writeMu.Unlock()
		return
	}
	defer t.Stop()

	// done channel signals all goroutines to stop on client disconnect
	done := make(chan struct{})

	// Handle client disconnection and pong responses
	conn.SetPongHandler(func(string) error { return nil })
	go func() {
		defer close(done)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				// Client disconnected or error
				t.Stop()
				return
			}
		}
	}()

	// Heartbeat: send a WebSocket ping every 30s to keep the connection alive
	// through Nginx and other reverse proxies that close idle connections.
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				writeMu.Lock()
				err := conn.WriteMessage(websocket.PingMessage, nil)
				writeMu.Unlock()
				if err != nil {
					return
				}
			case <-done:
				return
			}
		}
	}()

	// Send lines to client
	for {
		select {
		case <-done:
			return
		case line, ok := <-t.Lines:
			if !ok {
				return
			}
			if line.Err != nil {
				fmt.Printf("Tail error: %v\n", line.Err)
				continue
			}
			writeMu.Lock()
			err := conn.WriteMessage(websocket.TextMessage, []byte(line.Text))
			writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

// TopPathsRequest represents query parameters for the top paths endpoint.
type TopPathsRequest struct {
	StartTime string `form:"start_time" binding:"required"`
	EndTime   string `form:"end_time" binding:"required"`
	SourceID  string `form:"source_id"`
	Limit     int    `form:"limit,default=10" binding:"min=1,max=100"`
}

// TopPathsHandler Get top requested paths
// @Summary Get top requested paths
// @Description Get the top requested interface URIs along with their request count, average response time, and error rate.
// @Tags 2. Overview Dashboard
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start Time" default(2026-03-19 00:00:00)
// @Param end_time query string true "End Time" default(2026-03-20 00:00:00)
// @Param source_id query string false "Log File or Source ID" default(access.log)
// @Param limit query int false "Number of top paths to return (default 10, max 100)" default(10)
// @Success 200 {object} TopPathsResponseWrapper
// @Router /api/v1/sev/stats/top-paths [get]
func (ctrl *MonitorController) TopPathsHandler(c *gin.Context) {
	var req TopPathsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	result, err := ctrl.svc.GetTopPaths(req.StartTime, req.EndTime, req.SourceID, req.Limit)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

// TimeSeriesRequest represents query parameters for trend data.
type TimeSeriesRequest struct {
	Metric    string   `form:"metric" binding:"required,oneof=qps error_rate latency_p99 bandwidth"`
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
	SourceIDs []string `form:"source_ids"`
}

// TimeSeriesHandler Get trend data for charts
// @Summary Get trend data for charts
// @Description Get time-series data for a metric (qps, error_rate, latency_p99, bandwidth). Range cannot exceed 1 year. Returns exactly 30 points.
// @Tags 2. Overview Dashboard
// @Produce json
// @Security BearerAuth
// @Param metric query string true "Metric type" Enums(qps, error_rate, latency_p99, bandwidth)
// @Param start_time query string true "Start Time" default(2026-03-19 00:00:00)
// @Param end_time query string true "End Time" default(2026-03-20 00:00:00)
// @Param source_ids query []string false "List of Source IDs or Log Files to aggregate"
// @Success 200 {object} TimeSeriesResponseWrapper
// @Router /api/v1/sev/stats/timeseries [get]
func (ctrl *MonitorController) TimeSeriesHandler(c *gin.Context) {
	var req TimeSeriesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	if _, _, err := utils.ParseAndValidateRange(req.StartTime, req.EndTime, utils.MaxTimeRangeLimit); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	result, err := ctrl.svc.GetTimeSeriesStats(req.Metric, req.StartTime, req.EndTime, req.SourceIDs)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

// ==========================================
// 3. Logs Analysis Endpoints
// ==========================================

// LogQueryRequest represents query parameters for detailed log querying.
type LogQueryRequest struct {
	Page        int    `form:"page,default=1" binding:"min=1"`
	Size        int    `form:"size,default=50" binding:"min=1,max=500"`
	StartTime   string `form:"start_time"`
	EndTime     string `form:"end_time"`
	SourceID    string `form:"source_id"`
	IP          string `form:"ip"`
	Method      string `form:"method"`
	Status      *int   `form:"status"`
	StatusClass string `form:"status_class"`
	PathKeyword string `form:"path_keyword"`
	MinLatency  *int   `form:"min_latency"`
	MaxLatency  *int   `form:"max_latency"`
	Sort        string `form:"sort"`
}

// LogsHandler queries log details
// @Summary Query detailed logs
// @Description Query parsed Nginx logs with comprehensive filtering, sorting, and pagination
// @Tags 3. Logs Analysis
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default 1)" default(1)
// @Param size query int false "Page size (default 50, max 500)" default(50)
// @Param start_time query string false "Start Time" example(2026-03-19 00:00:00)
// @Param end_time query string false "End Time" example(2026-03-20 00:00:00)
// @Param source_id query string false "Log File or Source ID"
// @Param ip query string false "IP address (supports prefix match)" example(192.168.1.1)
// @Param method query string false "HTTP Method (e.g. GET)" example(GET)
// @Param status query int false "Exact HTTP Status (e.g. 500)" example(200)
// @Param status_class query string false "HTTP Status Class (e.g. 5xx)" example(5xx)
// @Param path_keyword query string false "Keyword to search in URL path" example(api)
// @Param min_latency query int false "Minimum Latency in ms" example(100)
// @Param max_latency query int false "Maximum Latency in ms" example(5000)
// @Param sort query string false "Sort order (time_desc or latency_desc)" Enums(time_desc, latency_desc) default(time_desc)
// @Success 200 {object} LogsResponseWrapper
// @Router /api/v1/sev/logs [get]
func (ctrl *MonitorController) LogsHandler(c *gin.Context) {
	var req LogQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	filter := storage.LogQueryFilter{
		Page:        req.Page,
		Size:        req.Size,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		IP:          req.IP,
		Method:      req.Method,
		Status:      req.Status,
		StatusClass: req.StatusClass,
		PathKeyword: req.PathKeyword,
		MinLatency:  req.MinLatency,
		MaxLatency:  req.MaxLatency,
		Sort:        req.Sort,
	}

	resp, err := ctrl.svc.GetLogs(req.SourceID, filter)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, resp)
}

// ==========================================
// 4. Geo Analysis Endpoints
// ==========================================

// GeoDistributionRequest represents query parameters for the geo distribution endpoint.
type GeoDistributionRequest struct {
	StartTime string `form:"start_time" binding:"required"`
	EndTime   string `form:"end_time" binding:"required"`
	SourceID  string `form:"source_id"`
	Limit     int    `form:"limit,default=100" binding:"min=1,max=1000"`
}

// GeoDistributionHandler Get geographical distribution of requests
// @Summary Get geographical distribution of requests
// @Description Get geographical distribution of IP addresses from logs.
// @Tags 4. Geo Analysis
// @Produce json
// @Security BearerAuth
// @Param start_time query string true "Start Time" default(2026-03-19 00:00:00)
// @Param end_time query string true "End Time" default(2026-03-20 00:00:00)
// @Param source_id query string false "Log File or Source ID" default(access.log)
// @Param limit query int false "Number of locations to return (default 100, max 1000)" default(100)
// @Success 200 {array} storage.GeoDistributionItem
// @Router /api/v1/sev/stats/geo-distribution [get]
func (ctrl *MonitorController) GeoDistributionHandler(c *gin.Context) {
	var req GeoDistributionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	result, err := ctrl.svc.GetGeoDistribution(req.StartTime, req.EndTime, req.SourceID, req.Limit)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

// ==========================================
// 5. System Stats Endpoints
// ==========================================

// StatsHandler Get aggregated logs statistics
// @Summary Get aggregated logs statistics
// @Description Get total logs count for all monitored tables
// @Tags 5. System Stats
// @Produce json
// @Security BearerAuth
// @Success 200 {object} StatsResponseWrapper
// @Router /api/v1/sev/stats [get]
func (ctrl *MonitorController) StatsHandler(c *gin.Context) {
	stats := ctrl.svc.GetStats()
	utils.Success(c, stats)
}

// ==========================================
// Helper functions
// ==========================================

// getLastNLinesOffset finds the offset of the Nth line from the end of the file.
func getLastNLinesOffset(filename string, n int) *tail.SeekInfo {
	file, err := os.Open(filename)
	if err != nil {
		return &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	}

	filesize := stat.Size()
	if filesize == 0 {
		return &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	}

	// Read a chunk from the end of the file
	// 8KB should be more than enough for 10 lines of typical logs
	const maxChunk = 8192
	var readSize int64 = maxChunk
	if filesize < readSize {
		readSize = filesize
	}

	buf := make([]byte, readSize)
	_, err = file.ReadAt(buf, filesize-readSize)
	if err != nil {
		return &tail.SeekInfo{Offset: 0, Whence: os.SEEK_END}
	}

	count := 0
	pos := int64(len(buf)) - 1

	// If the file ends with a newline, skip it so we don't count it as one of the N lines
	if pos >= 0 && buf[pos] == '\n' {
		pos--
	}

	for ; pos >= 0; pos-- {
		if buf[pos] == '\n' {
			count++
			if count == n {
				// Found the start of the Nth line from the end
				return &tail.SeekInfo{Offset: filesize - readSize + pos + 1, Whence: os.SEEK_SET}
			}
		}
	}

	// If we didn't find N lines:
	// If the whole file was read, start from the beginning
	if filesize <= readSize {
		return &tail.SeekInfo{Offset: 0, Whence: os.SEEK_SET}
	}

	// If the file is large and we didn't find N lines in the chunk, 
	// just start from the beginning of the chunk as a best effort
	return &tail.SeekInfo{Offset: filesize - readSize, Whence: os.SEEK_SET}
}
