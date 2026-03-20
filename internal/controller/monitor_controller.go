package controller

import (
	"SvcWatch/internal/service"
	"SvcWatch/internal/utils"
	"fmt"
	"time"

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

// ParsedTimeRange holds parsed and validated time objects.
type ParsedTimeRange struct {
	StartT time.Time
	EndT   time.Time
}

// PingHandler Health Check
func (ctrl *MonitorController) PingHandler(c *gin.Context) {
	utils.Success(c, gin.H{
		"message": "pong",
	})
}

// StatsHandler Get aggregated logs statistics
func (ctrl *MonitorController) StatsHandler(c *gin.Context) {
	stats := ctrl.svc.GetStats()
	utils.Success(c, stats)
}

// OverviewHandler Get business overview key metrics
func (ctrl *MonitorController) OverviewHandler(c *gin.Context) {
	var req TimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, "start_time and end_time are required")
		return
	}

	if _, err := ctrl.validateTimeRange(req.StartTime, req.EndTime); err != nil {
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
func (ctrl *MonitorController) StatusDistributionHandler(c *gin.Context) {
	var req TimeRangeRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		utils.Error(c, 400, "start_time and end_time are required")
		return
	}

	parsed, err := ctrl.validateTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		utils.Error(c, 400, err.Error())
		return
	}

	result, err := ctrl.svc.GetStatusDistribution(parsed.StartT, parsed.EndT, req.LogFile)
	if err != nil {
		utils.Error(c, 500, err.Error())
		return
	}

	utils.Success(c, result)
}

// validateTimeRange parses and validates the time strings.
func (ctrl *MonitorController) validateTimeRange(startStr, endStr string) (*ParsedTimeRange, error) {
	parseTime := func(s string) (time.Time, error) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t, err = time.Parse("2006-01-02 15:04:05", s)
		}
		return t, err
	}

	startT, err := parseTime(startStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format")
	}

	endT, err := parseTime(endStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format")
	}

	now := time.Now()
	if startT.After(now) {
		return nil, fmt.Errorf("start_time cannot be in the future")
	}
	if !endT.After(startT) {
		return nil, fmt.Errorf("end_time must be after start_time")
	}
	if endT.Sub(startT) > 366*24*time.Hour {
		return nil, fmt.Errorf("time range cannot exceed 1 year")
	}

	return &ParsedTimeRange{StartT: startT, EndT: endT}, nil
}
