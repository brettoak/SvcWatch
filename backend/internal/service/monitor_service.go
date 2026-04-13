package service

import (
	"fmt"
	"math"
	"path/filepath"
	"time"
	"SvcWatch/internal/config"
	"SvcWatch/internal/monitor"
	"SvcWatch/internal/storage"
	"SvcWatch/internal/utils"
)

// MonitorService handles the business logic for monitor statistics.
type MonitorService struct {
	monitors []*monitor.Monitor
	cfg      *config.Config
	store    *storage.SqliteStorage
}

// NewMonitorService creates a new instance of MonitorService.
func NewMonitorService(monitors []*monitor.Monitor, cfg *config.Config, store *storage.SqliteStorage) *MonitorService {
	return &MonitorService{
		monitors: monitors,
		cfg:      cfg,
		store:    store,
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
	} else {
		// Round final values to 2 decimal places after aggregation
		aggregated.SuccessRate.Value = math.Round(aggregated.SuccessRate.Value*100) / 100
		aggregated.ErrorRate.Value = math.Round(aggregated.ErrorRate.Value*100) / 100
		aggregated.AvgResponseTime.Value = math.Round(aggregated.AvgResponseTime.Value*100) / 100
		
		aggregated.TotalRequests.ComparePercent = math.Round(aggregated.TotalRequests.ComparePercent*100) / 100
		aggregated.SuccessRate.ComparePercent = math.Round(aggregated.SuccessRate.ComparePercent*100) / 100
		aggregated.ErrorRate.ComparePercent = math.Round(aggregated.ErrorRate.ComparePercent*100) / 100
		aggregated.AvgResponseTime.ComparePercent = math.Round(aggregated.AvgResponseTime.ComparePercent*100) / 100
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
			perc := (float64(aggregated.Distribution[i].Count) / float64(aggregated.Total)) * 100
			aggregated.Distribution[i].Percentage = math.Round(perc*100) / 100
		}
	}
	return aggregated, nil
}

// GetLogs queries detailed logs with filtering, sorting, and pagination.
func (s *MonitorService) GetLogs(sourceID string, filter storage.LogQueryFilter) (*storage.LogQueryResponse, error) {
	var tableNames []string

	// Determine which tables to query based on SourceID
	for _, monInst := range s.monitors {
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		
		// If a generic SourceID was provided, skip non-matching tables
		if sourceID != "" && tableName != sourceID && filepath.Base(logPath) != sourceID {
			continue
		}
		tableNames = append(tableNames, tableName)
	}

	if len(tableNames) == 0 {
		return &storage.LogQueryResponse{Total: 0, Page: filter.Page, Size: filter.Size, Items: []storage.LogQueryItem{}}, nil
	}

	return s.store.QueryLogs(tableNames, filter)
}

// GetTimeSeriesStats aggregates trend data for a specific metric and interval across multiple source IDs.
func (s *MonitorService) GetTimeSeriesStats(metric, startTime, endTime string, sourceIDs []string) (*storage.TimeSeriesResult, error) {
	var tableNames []string

	// Map sourceIDs to table names for efficient lookup
	sourceMap := make(map[string]bool)
	for _, id := range sourceIDs {
		sourceMap[id] = true
	}

	for _, monInst := range s.monitors {
		tableName := monInst.GetTableName()
		logPath := monInst.GetLogPath()
		
		// If sourceIDs are provided, filter by matching table name or log file base name
		if len(sourceIDs) > 0 {
			if !sourceMap[tableName] && !sourceMap[filepath.Base(logPath)] {
				continue
			}
		}
		tableNames = append(tableNames, tableName)
	}

	if len(tableNames) == 0 {
		return &storage.TimeSeriesResult{Metric: metric, Interval: "auto", Points: []storage.TimeSeriesPoint{}}, nil
	}

	startT, endT, err := utils.ParseAndValidateRange(startTime, endTime, utils.MaxTimeRangeLimit)
	if err != nil {
		return nil, err
	}

	duration := endT.Sub(startT)

	type tier struct {
		name string
		sec  float64
	}
	tiers := []tier{
		{"1m", 60},
		{"2m", 120},
		{"5m", 300},
		{"10m", 600},
		{"30m", 1800},
		{"1h", 3600},
		{"2h", 7200},
		{"6h", 21600},
		{"12h", 43200},
		{"1d", 86400},
		{"2d", 172800},
		{"3d", 259200},
		{"4d", 345600},
		{"5d", 432000},
		{"1w", 604800},
		{"2w", 1209600},
		{"1M", 2592000},
	}

	// Find the interval that results in points count closest to 20
	bestInterval := tiers[len(tiers)-1].name
	minDiff := math.MaxFloat64
	for _, t := range tiers {
		pointsCount := duration.Seconds() / t.sec
		diff := math.Abs(pointsCount - 20)
		if diff < minDiff {
			minDiff = diff
			bestInterval = t.name
		}
	}

	return s.store.GetTimeSeries(tableNames, metric, bestInterval, startT, endT)
}

