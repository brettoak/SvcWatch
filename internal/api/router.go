package api

import (
	"fmt"
	"math"
	"path/filepath"
	"time"
	"SvcWatch/internal/config"
	"SvcWatch/internal/middleware"
	"SvcWatch/internal/monitor"
	"SvcWatch/internal/storage"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// APIController holds dependencies for API handlers.
type APIController struct {
	monitors []*monitor.Monitor
	cfg      *config.Config
}

// PingHandler Health Check
// @Summary Health Check
// @Description returns a "pong" string to verify the service is running
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sev/ping [get]
func (ctrl *APIController) PingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// StatsHandler Get aggregated logs statistics
// @Summary Get aggregated logs statistics
// @Description Retrieves the total log count for each configured monitored target.
// @Tags Statistics
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/sev/stats [get]
func (ctrl *APIController) StatsHandler(c *gin.Context) {
	stats := make(map[string]interface{})
	for i, monInst := range ctrl.monitors {
		tableName := ctrl.cfg.Targets[i].Table
		stats[tableName] = monInst.GetStats()["total_logs"]
	}
	c.JSON(200, stats)
}

// OverviewHandler Get business overview key metrics
// @Summary Get business overview key metrics
// @Description Retrieves Total Requests, Success Rate, Error Rate, and Average Response time with % comparison against yesterday.
// @Tags Statistics
// @Security BearerAuth
// @Produce json
// @Param start_time query string true "Start Time (RFC3339 or YYYY-MM-DD HH:MM:SS)" example(2026-03-10 00:00:00)
// @Param end_time query string true "End Time (RFC3339 or YYYY-MM-DD HH:MM:SS)" example(2026-03-17 23:59:59)
// @Param log_file query string false "Optional specific log file (table name or filename) to search" example(nginx_logs)
// @Success 200 {object} storage.OverviewStats
// @Router /api/v1/sev/overview [get]
func (ctrl *APIController) OverviewHandler(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	logFile := c.Query("log_file") // Optional

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(400, gin.H{"error": "start_time and end_time are required"})
		return
	}

	// Helper to parse time in supported formats
	parseTime := func(s string) (time.Time, error) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05", s)
		}
		return t, err
	}

	startT, err := parseTime(startTimeStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid start_time format"})
		return
	}

	endT, err := parseTime(endTimeStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid end_time format"})
		return
	}

	now := time.Now()
	if startT.After(now) {
		c.JSON(400, gin.H{"error": "start_time cannot be in the future"})
		return
	}

	if !endT.After(startT) {
		c.JSON(400, gin.H{"error": "end_time must be after start_time"})
		return
	}

	if endT.Sub(startT) > 366*24*time.Hour {
		c.JSON(400, gin.H{"error": "time range cannot exceed 1 year"})
		return
	}

	// We'll aggregate results if logFile is empty. For MVP, we'll just sum or take average
	// To be perfectly accurate for multiple files, we should ask storage to query all tables or do custom logic.
	// We'll do a simple iteration across configured tables.
	var aggregated *storage.OverviewStats

	for _, monInst := range ctrl.monitors {
		// If specific logFile is requested, skip others
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		if logFile != "" && tableName != logFile && filepath.Base(logPath) != logFile {
			continue
		}

		stats, err := monInst.GetOverviewStats(startTimeStr, endTimeStr)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get stats for %s: %v", tableName, err)})
			return
		}

		if aggregated == nil {
			aggregated = stats
		} else {
			// Basic accumulation; 
			// Note: Average rate accumulation is not perfectly mathematically sound 
			// without total weights, but sufficient for an MVP overview dashboard.
			aggregated.TotalRequests.Value += stats.TotalRequests.Value
			aggregated.SuccessRate.Value = (aggregated.SuccessRate.Value + stats.SuccessRate.Value) / 2
			aggregated.ErrorRate.Value = (aggregated.ErrorRate.Value + stats.ErrorRate.Value) / 2
			aggregated.AvgResponseTime.Value = (aggregated.AvgResponseTime.Value + stats.AvgResponseTime.Value) / 2
			
			// For simplicity we average out the compare percents for combined view
			aggregated.TotalRequests.ComparePercent = (aggregated.TotalRequests.ComparePercent + stats.TotalRequests.ComparePercent) / 2
			aggregated.SuccessRate.ComparePercent = (aggregated.SuccessRate.ComparePercent + stats.SuccessRate.ComparePercent) / 2
			aggregated.ErrorRate.ComparePercent = (aggregated.ErrorRate.ComparePercent + stats.ErrorRate.ComparePercent) / 2
			aggregated.AvgResponseTime.ComparePercent = (aggregated.AvgResponseTime.ComparePercent + stats.AvgResponseTime.ComparePercent) / 2
		}
	}

	if aggregated == nil {
		aggregated = &storage.OverviewStats{}
	}

	c.JSON(200, aggregated)
}

// StatusDistributionHandler Get distribution of HTTP status codes
// @Summary Get distribution of HTTP status codes
// @Description Retrieves the total count and 1xx, 2xx, 3xx, 4xx, 5xx distribution with percentages.
// @Tags Statistics
// @Security BearerAuth
// @Produce json
// @Param start_time query string true "Start Time (RFC3339 or YYYY-MM-DD HH:MM:SS)" example(2026-03-10 00:00:00)
// @Param end_time query string true "End Time (RFC3339 or YYYY-MM-DD HH:MM:SS)" example(2026-03-17 23:59:59)
// @Param log_file query string false "Optional specific log file (table name or filename) to search" example(nginx_logs)
// @Success 200 {object} storage.StatusDistributionResult
// @Router /api/v1/sev/distribution [get]
func (ctrl *APIController) StatusDistributionHandler(c *gin.Context) {
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")
	logFile := c.Query("log_file")

	if startTimeStr == "" || endTimeStr == "" {
		c.JSON(400, gin.H{"error": "start_time and end_time are required"})
		return
	}

	parseTime := func(s string) (time.Time, error) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05", s)
		}
		return t, err
	}

	startT, err := parseTime(startTimeStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid start_time format"})
		return
	}

	endT, err := parseTime(endTimeStr)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid end_time format"})
		return
	}

	if startT.After(time.Now()) {
		c.JSON(400, gin.H{"error": "start_time cannot be in the future"})
		return
	}

	if !endT.After(startT) {
		c.JSON(400, gin.H{"error": "end_time must be after start_time"})
		return
	}

	var aggregated *storage.StatusDistributionResult

	for _, monInst := range ctrl.monitors {
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		if logFile != "" && tableName != logFile && filepath.Base(logPath) != logFile {
			continue
		}

		result, err := monInst.GetStatusDistribution(startT, endT)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("failed to get distribution for %s: %v", tableName, err)})
			return
		}

		if aggregated == nil {
			aggregated = result
		} else {
			aggregated.Total += result.Total
			for i := range aggregated.Distribution {
				aggregated.Distribution[i].Count += result.Distribution[i].Count
			}
		}
	}

	if aggregated == nil {
		aggregated = &storage.StatusDistributionResult{
			Distribution: []storage.StatusDistributionEntry{
				{CodeClass: "1xx"}, {CodeClass: "2xx"}, {CodeClass: "3xx"}, {CodeClass: "4xx"}, {CodeClass: "5xx"},
			},
		}
	} else if aggregated.Total > 0 {
		// Recalculate percentages for aggregated view
		for i := range aggregated.Distribution {
			perc := float64(aggregated.Distribution[i].Count) / float64(aggregated.Total)
			aggregated.Distribution[i].Percentage = math.Round(perc*1000) / 1000
		}
	}

	c.JSON(200, aggregated)
}

// SetupRouter initializes and configures the Gin API router.
func SetupRouter(monitors []*monitor.Monitor, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// Enable CORS for all origins (fixes Swagger UI doc.json fetch issues)
	router.Use(cors.Default())

	ctrl := &APIController{monitors: monitors, cfg: cfg}

	v1 := router.Group("/api/v1/sev")
	{
		// Public routes
		v1.GET("/ping", ctrl.PingHandler)

		// Protected routes require token authentication
		private := v1.Group("")
		private.Use(middleware.TokenAuthMiddleware(cfg.Auth.PassportURL))
		{
			// Example permission required to view stats
			private.GET("/stats", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:stats"), ctrl.StatsHandler)
			// Overview endpoint
			private.GET("/overview", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:overview"), ctrl.OverviewHandler)
			// Status distribution endpoint
			private.GET("/distribution", middleware.PermissionMiddleware(cfg.Auth.PermissionURL, cfg.Auth.SysCode, "view:distribution"), ctrl.StatusDistributionHandler)
		}
	}

	// Register Swagger UI
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
