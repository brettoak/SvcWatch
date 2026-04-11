package controller

import (
	"SvcWatch/internal/service"
	"SvcWatch/internal/storage"
	"SvcWatch/internal/utils"

	"github.com/gin-gonic/gin"
)

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

// TimeSeriesRequest represents query parameters for trend data.
type TimeSeriesRequest struct {
	Metric    string   `form:"metric" binding:"required,oneof=qps error_rate latency_p99 bandwidth"`
	Interval  string   `form:"interval" binding:"required,oneof=1m 5m 1h 6h 1d 1w 1M"`
	StartTime string   `form:"start_time" binding:"required"`
	EndTime   string   `form:"end_time" binding:"required"`
	SourceIDs []string `form:"source_ids"`
}

// PingHandler Health Check
// @Summary Health Check
// @Description Returns pong message to check if API is alive
// @Tags System
// @Produce json
// @Success 200 {object} utils.Response
// @Router /api/v1/sev/ping [get]
func (ctrl *MonitorController) PingHandler(c *gin.Context) {
	utils.Success(c, gin.H{
		"message": "pong",
	})
}

// StatsHandler Get aggregated logs statistics
// @Summary Get aggregated logs statistics
// @Description Get total logs count for all monitored tables
// @Tags Monitor
// @Produce json
// @Security BearerAuth
// @Success 200 {object} StatsResponseWrapper
// @Router /api/v1/sev/stats [get]
func (ctrl *MonitorController) StatsHandler(c *gin.Context) {
	stats := ctrl.svc.GetStats()
	utils.Success(c, stats)
}

// OverviewHandler Get business overview key metrics
// @Summary Get business overview key metrics
// @Description Get overview statistics with comparison for a time range
// @Tags Monitor
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

// StatusDistributionHandler Get distribution of HTTP status codes
// @Summary Get HTTP status code distribution
// @Description Get distribution of status codes for a time range
// @Tags Monitor
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
// @Tags Monitor
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

// TimeSeriesHandler Get trend data for charts
// @Summary Get trend data for charts
// @Description Get time-series data for a metric (qps, error_rate, latency_p99, bandwidth). Range cannot exceed 1 year. Returns max 20 points.
// @Tags Monitor
// @Produce json
// @Security BearerAuth
// @Param metric query string true "Metric type" Enums(qps, error_rate, latency_p99, bandwidth)
// @Param interval query string true "Aggregation interval (may be overridden if unreasonable)" Enums(1m, 5m, 1h, 6h, 1d, 1w, 1M)
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

	result, err := ctrl.svc.GetTimeSeriesStats(req.Metric, req.Interval, req.StartTime, req.EndTime, req.SourceIDs)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

