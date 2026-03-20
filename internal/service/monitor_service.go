package service

import (
	"fmt"
	"math"
	"path/filepath"
	"time"
	"SvcWatch/internal/config"
	"SvcWatch/internal/monitor"
	"SvcWatch/internal/storage"
)

// MonitorService handles the business logic for monitor statistics.
type MonitorService struct {
	monitors []*monitor.Monitor
	cfg      *config.Config
}

// NewMonitorService creates a new instance of MonitorService.
func NewMonitorService(monitors []*monitor.Monitor, cfg *config.Config) *MonitorService {
	return &MonitorService{
		monitors: monitors,
		cfg:      cfg,
	}
}

// GetStats returns aggregated log counts for all monitors.
func (s *MonitorService) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	for i, monInst := range s.monitors {
		tableName := s.cfg.Targets[i].Table
		stats[tableName] = monInst.GetStats()["total_logs"]
	}
	return stats
}

// GetOverview aggregates key metrics across monitors.
func (s *MonitorService) GetOverview(startTime, endTime, logFile string) (*storage.OverviewStats, error) {
	var aggregated *storage.OverviewStats

	for _, monInst := range s.monitors {
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		if logFile != "" && tableName != logFile && filepath.Base(logPath) != logFile {
			continue
		}

		stats, err := monInst.GetOverviewStats(startTime, endTime)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for %s: %v", tableName, err)
		}

		if aggregated == nil {
			aggregated = stats
		} else {
			aggregated.TotalRequests.Value += stats.TotalRequests.Value
			aggregated.SuccessRate.Value = (aggregated.SuccessRate.Value + stats.SuccessRate.Value) / 2
			aggregated.ErrorRate.Value = (aggregated.ErrorRate.Value + stats.ErrorRate.Value) / 2
			aggregated.AvgResponseTime.Value = (aggregated.AvgResponseTime.Value + stats.AvgResponseTime.Value) / 2
			aggregated.TotalRequests.ComparePercent = (aggregated.TotalRequests.ComparePercent + stats.TotalRequests.ComparePercent) / 2
			aggregated.SuccessRate.ComparePercent = (aggregated.SuccessRate.ComparePercent + stats.SuccessRate.ComparePercent) / 2
			aggregated.ErrorRate.ComparePercent = (aggregated.ErrorRate.ComparePercent + stats.ErrorRate.ComparePercent) / 2
			aggregated.AvgResponseTime.ComparePercent = (aggregated.AvgResponseTime.ComparePercent + stats.AvgResponseTime.ComparePercent) / 2
		}
	}

	if aggregated == nil {
		aggregated = &storage.OverviewStats{}
	}
	return aggregated, nil
}

// GetStatusDistribution retrieves HTTP status code distribution across monitors.
func (s *MonitorService) GetStatusDistribution(startT, endT time.Time, logFile string) (*storage.StatusDistributionResult, error) {
	var aggregated *storage.StatusDistributionResult

	for _, monInst := range s.monitors {
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		if logFile != "" && tableName != logFile && filepath.Base(logPath) != logFile {
			continue
		}

		result, err := monInst.GetStatusDistribution(startT, endT)
		if err != nil {
			return nil, fmt.Errorf("failed to get distribution for %s: %v", tableName, err)
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
		for i := range aggregated.Distribution {
			perc := float64(aggregated.Distribution[i].Count) / float64(aggregated.Total)
			aggregated.Distribution[i].Percentage = math.Round(perc*1000) / 1000
		}
	}
	return aggregated, nil
}
