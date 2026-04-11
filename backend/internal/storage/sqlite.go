package storage

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"strings"
	"time"
	"SvcWatch/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// SqliteStorage implements SQLite storage.
type SqliteStorage struct {
	db *sql.DB
}

// NewSqliteStorage creates a new SqliteStorage.
func NewSqliteStorage(dbPath string) *SqliteStorage {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to open sqlite database: %v", err)
	}

	return &SqliteStorage{
		db: db,
	}
}

// InitTable initializes a mapped log table in the database.
func (s *SqliteStorage) InitTable(tableName string, clearOnStartup bool) {
	if clearOnStartup {
		// Drop table to clear previous data on startup
		dropTableSQL := fmt.Sprintf(`DROP TABLE IF EXISTS %s;`, tableName)
		_, err := s.db.Exec(dropTableSQL)
		if err != nil {
			log.Fatalf("Failed to drop table %s: %v", tableName, err)
		}
	}

	// Create table if not exists
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		remote_addr TEXT,
		remote_user TEXT,
		time_local DATETIME,
		request TEXT,
		status INTEGER,
		body_bytes_sent INTEGER,
		http_referer TEXT,
		http_user_agent TEXT,
		request_time REAL
	);
	`, tableName)
	_, err := s.db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Failed to create table %s: %v", tableName, err)
	}
}

// Save saves a log entry to SQLite.
func (s *SqliteStorage) Save(tableName string, entry *model.LogEntry) error {
	insertSQL := fmt.Sprintf(`
	INSERT INTO %s (
		remote_addr, remote_user, time_local, request, status, body_bytes_sent, http_referer, http_user_agent, request_time
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tableName)
	_, err := s.db.Exec(insertSQL,
		entry.RemoteAddr,
		entry.RemoteUser,
		entry.TimeLocal,
		entry.Request,
		entry.Status,
		entry.BodyBytesSent,
		entry.HttpReferer,
		entry.HttpUserAgent,
		entry.RequestTime,
	)
	if err != nil {
		log.Printf("Failed to insert log entry into %s: %v", tableName, err)
		return err
	}
	return nil
}

// GetTotalCount returns the total number of logs from a SQLite table.
func (s *SqliteStorage) GetTotalCount(tableName string) int {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := s.db.QueryRow(query).Scan(&count)
	if err != nil {
		log.Printf("Failed to get total count from %s: %v", tableName, err)
		return 0
	}
	return count
}

// MetricValue represents a value and its percentage change compared to a previous period.
type MetricValue struct {
	Value          float64 `json:"value"`
	ComparePercent float64 `json:"compare_percent"`
}

// OverviewStats contains the calculated statistics for the business overview.
type OverviewStats struct {
	TotalRequests   MetricValue `json:"total_requests"`
	SuccessRate     MetricValue `json:"success_rate"`
	ErrorRate       MetricValue `json:"error_rate"`
	AvgResponseTime MetricValue `json:"avg_response_time"`
	CompareType     string      `json:"compare_type"` // e.g., "vs yesterday" or "vs previous period"
}

// StatusDistributionEntry represents a single status code class distribution.
type StatusDistributionEntry struct {
	CodeClass  string  `json:"code_class"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// StatusDistributionResult contains the total record count and the distribution of status codes.
type StatusDistributionResult struct {
	Total        int                       `json:"total"`
	Distribution []StatusDistributionEntry `json:"distribution"`
}

// TimeSeriesPoint represents a single data point in a time series chart.
type TimeSeriesPoint struct {
	Timestamp string  `json:"ts"`
	Value     float64 `json:"value"`
}

// TimeSeriesResult represents the complete response for a time series query.
type TimeSeriesResult struct {
	Metric   string            `json:"metric"`
	Interval string            `json:"interval"`
	Points   []TimeSeriesPoint `json:"points"`
}

// BaseMetrics contains raw metric counts for a specific time period.
type BaseMetrics struct {
	TotalRequests   float64
	SuccessCount    float64
	ErrorCount      float64
	AvgResponseTime float64
}

// GetBaseMetrics queries the raw metrics for a specific time range.
func (s *SqliteStorage) GetBaseMetrics(tableName string, startTime, endTime time.Time) (*BaseMetrics, error) {
	// The query calculates Total Request, Success Count (<400), Error Count (>=400) and Average Request Time.
	query := fmt.Sprintf(`
		SELECT 
			CAST(COUNT(*) AS REAL) as total_requests,
			CAST(SUM(CASE WHEN status >= 200 AND status < 300 THEN 1 ELSE 0 END) AS REAL) as success_count,
			CAST(SUM(CASE WHEN status >= 400 THEN 1 ELSE 0 END) AS REAL) as error_count,
			COALESCE(AVG(request_time), 0.0) as avg_response_time
		FROM %s 
		WHERE time_local >= ? AND time_local <= ?
	`, tableName)

	var metrics BaseMetrics
	var successCount sql.NullFloat64
	var errorCount sql.NullFloat64

	err := s.db.QueryRow(query, startTime, endTime).Scan(
		&metrics.TotalRequests,
		&successCount,
		&errorCount,
		&metrics.AvgResponseTime,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return &metrics, nil
		}
		return nil, fmt.Errorf("failed to query base metrics: %w", err)
	}

	if successCount.Valid {
		metrics.SuccessCount = successCount.Float64
	}
	if errorCount.Valid {
		metrics.ErrorCount = errorCount.Float64
	}

	return &metrics, nil
}

func calculateComparePercent(current, previous float64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0 // 100% increase if previous was 0 and current is some
		}
		return 0.0 // 0% change if both are 0
	}
	diff := ((current - previous) / previous) * 100.0
	// Round to 2 decimal places
	return math.Round(diff*100) / 100
}

// GetOverviewWithCompare calculates metrics for the given timeframe and compares against the same timeframe 24 hours prior.
func (s *SqliteStorage) GetOverviewWithCompare(tableName string, startTimeStr, endTimeStr string) (*OverviewStats, error) {
	// 1. Parse times to calculate the previous period
	layout := time.RFC3339
	if len(startTimeStr) == 19 {
		layout = time.DateTime // "2006-01-02 15:04:05"
	}

	startTime, err := time.Parse(layout, startTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_time format: %w", err)
	}
	endTime, err := time.Parse(layout, endTimeStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_time format: %w", err)
	}

	// 2. Get current period metrics
	currentMetrics, err := s.GetBaseMetrics(tableName, startTime, endTime)
	if err != nil {
		return nil, err
	}

	// Calculate range duration
	rangeDuration := endTime.Sub(startTime)
	offset := 24 * time.Hour
	compareType := "vs yesterday"
	if rangeDuration > 24*time.Hour {
		offset = rangeDuration
		compareType = "vs previous period"
	}

	prevStartTime := startTime.Add(-offset)
	prevEndTime := endTime.Add(-offset)

	// 3. Get previous period metrics
	prevMetrics, err := s.GetBaseMetrics(tableName, prevStartTime, prevEndTime)
	if err != nil {
		// If getting previous metrics fails, we will still return current metrics but with 0% compare.
		log.Printf("Warning: Failed to get previous metrics for %s: %v", tableName, err)
		prevMetrics = &BaseMetrics{} 
	}

	// 4. Calculate Rates for Current Period
	var currSuccessRate, currErrorRate float64
	if currentMetrics.TotalRequests > 0 {
		currSuccessRate = (currentMetrics.SuccessCount / currentMetrics.TotalRequests) * 100
		currErrorRate = (currentMetrics.ErrorCount / currentMetrics.TotalRequests) * 100
	}

	// 5. Calculate Rates for Previous Period
	var prevSuccessRate, prevErrorRate float64
	if prevMetrics.TotalRequests > 0 {
		prevSuccessRate = (prevMetrics.SuccessCount / prevMetrics.TotalRequests) * 100
		prevErrorRate = (prevMetrics.ErrorCount / prevMetrics.TotalRequests) * 100
	}

	// 6. Assemble Final Overview Stats
	stats := &OverviewStats{
		TotalRequests: MetricValue{
			Value:          currentMetrics.TotalRequests,
			ComparePercent: calculateComparePercent(currentMetrics.TotalRequests, prevMetrics.TotalRequests),
		},
		SuccessRate: MetricValue{
			Value:          math.Round(currSuccessRate*100) / 100, // percentage 0-100 rounded to 2 decimals
			ComparePercent: calculateComparePercent(currSuccessRate, prevSuccessRate),
		},
		ErrorRate: MetricValue{
			Value:          math.Round(currErrorRate*100) / 100, // percentage 0-100 rounded to 2 decimals
			ComparePercent: calculateComparePercent(currErrorRate, prevErrorRate),
		},
		AvgResponseTime: MetricValue{
			Value:          math.Round(currentMetrics.AvgResponseTime*100) / 100, // rounded to 2 decimals
			ComparePercent: calculateComparePercent(currentMetrics.AvgResponseTime, prevMetrics.AvgResponseTime),
		},
		CompareType: compareType,
	}

	return stats, nil
}

// GetStatusDistribution calculates the distribution of status codes (1xx, 2xx, 3xx, 4xx, 5xx) for a given time range.
func (s *SqliteStorage) GetStatusDistribution(tableName string, startTime, endTime time.Time) (*StatusDistributionResult, error) {
	query := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total,
			SUM(CASE WHEN status >= 100 AND status < 200 THEN 1 ELSE 0 END) as s1xx,
			SUM(CASE WHEN status >= 200 AND status < 300 THEN 1 ELSE 0 END) as s2xx,
			SUM(CASE WHEN status >= 300 AND status < 400 THEN 1 ELSE 0 END) as s3xx,
			SUM(CASE WHEN status >= 400 AND status < 500 THEN 1 ELSE 0 END) as s4xx,
			SUM(CASE WHEN status >= 500 AND status < 600 THEN 1 ELSE 0 END) as s5xx
		FROM %s 
		WHERE time_local >= ? AND time_local <= ?
	`, tableName)

	var total int
	var s1xx, s2xx, s3xx, s4xx, s5xx sql.NullInt64

	err := s.db.QueryRow(query, startTime, endTime).Scan(&total, &s1xx, &s2xx, &s3xx, &s4xx, &s5xx)
	if err != nil {
		return nil, fmt.Errorf("failed to query status distribution: %w", err)
	}

	result := &StatusDistributionResult{
		Total: total,
		Distribution: []StatusDistributionEntry{
			{CodeClass: "1xx", Count: int(s1xx.Int64)},
			{CodeClass: "2xx", Count: int(s2xx.Int64)},
			{CodeClass: "3xx", Count: int(s3xx.Int64)},
			{CodeClass: "4xx", Count: int(s4xx.Int64)},
			{CodeClass: "5xx", Count: int(s5xx.Int64)},
		},
	}

	if total > 0 {
		for i := range result.Distribution {
			perc := (float64(result.Distribution[i].Count) / float64(total)) * 100
			result.Distribution[i].Percentage = math.Round(perc*100) / 100
		}
	}

	return result, nil
}

// GetTimeSeries retrieves trend data for a specific metric and interval across multiple tables.
func (s *SqliteStorage) GetTimeSeries(tableNames []string, metric string, interval string, startTime, endTime time.Time) (*TimeSeriesResult, error) {
	if len(tableNames) == 0 {
		return &TimeSeriesResult{Metric: metric, Interval: interval, Points: []TimeSeriesPoint{}}, nil
	}

	var timeFormat string
	var intervalSec float64
	switch interval {
	case "1m":
		timeFormat = "%Y-%m-%dT%H:%M:00Z"
		intervalSec = 60
	case "5m":
		// Custom format for 5m rounding
		timeFormat = "%Y-%m-%dT%H:"
		intervalSec = 300
	case "1h":
		timeFormat = "%Y-%m-%dT%H:00:00Z"
		intervalSec = 3600
	case "6h":
		// Custom format for 6h rounding
		intervalSec = 21600
	case "1d":
		timeFormat = "%Y-%m-%dT00:00:00Z"
		intervalSec = 86400
	case "1w":
		// Custom format for weekly rounding (starts on Monday)
		intervalSec = 604800
	case "1M":
		// Custom format for monthly rounding
		intervalSec = 2592000
	default:
		return nil, fmt.Errorf("unsupported interval: %s", interval)
	}

	// For custom intervals, we need a special grouping expression
	var timeExpr string
	if interval == "5m" {
		// Example: 2024-01-01T00:07:00Z -> 2024-01-01T00:05:00Z
		timeExpr = "strftime('%Y-%m-%dT%H:', time_local) || printf('%02d', (strftime('%M', time_local) / 5) * 5) || ':00Z'"
	} else if interval == "6h" {
		// Example: 2024-01-01T07:00:00Z -> 2024-01-01T06:00:00Z
		timeExpr = "strftime('%Y-%m-%dT', time_local) || printf('%02d', (strftime('%H', time_local) / 6) * 6) || ':00:00Z'"
	} else if interval == "1w" {
		// Group by week starting on Monday
		timeExpr = "date(time_local, 'weekday 0', '-6 days') || 'T00:00:00Z'"
	} else if interval == "1M" {
		// Group by the first day of the month
		timeExpr = "strftime('%Y-%m-01T00:00:00Z', time_local)"
	} else {
		timeExpr = fmt.Sprintf("strftime('%s', time_local)", timeFormat)
	}

	// Construct UNION query for multiple tables
	var unions []string
	for _, tableName := range tableNames {
		unions = append(unions, fmt.Sprintf("SELECT time_local, status, body_bytes_sent, request_time FROM %s WHERE time_local >= ? AND time_local <= ?", tableName))
	}
	unionQuery := strings.Join(unions, " UNION ALL ")

	var finalQuery string
	switch metric {
	case "qps":
		finalQuery = fmt.Sprintf(`
			SELECT ts, COUNT(*) / %f as val
			FROM (SELECT %s as ts FROM (%s))
			GROUP BY ts ORDER BY ts
		`, intervalSec, timeExpr, unionQuery)
	case "error_rate":
		finalQuery = fmt.Sprintf(`
			SELECT ts, (SUM(CASE WHEN status >= 400 THEN 1 ELSE 0 END) * 100.0 / COUNT(*)) as val
			FROM (SELECT %s as ts, status FROM (%s))
			GROUP BY ts ORDER BY ts
		`, timeExpr, unionQuery)
	case "bandwidth":
		// Bandwidth as bytes/sec
		finalQuery = fmt.Sprintf(`
			SELECT ts, SUM(body_bytes_sent) / %f as val
			FROM (SELECT %s as ts, body_bytes_sent FROM (%s))
			GROUP BY ts ORDER BY ts
		`, intervalSec, timeExpr, unionQuery)
	case "latency_p99":
		finalQuery = fmt.Sprintf(`
			SELECT ts, MAX(request_time) as val
			FROM (
				SELECT ts, request_time, 
				       ROW_NUMBER() OVER (PARTITION BY ts ORDER BY request_time) as rn,
				       COUNT(*) OVER (PARTITION BY ts) as cnt
				FROM (SELECT %s as ts, request_time FROM (%s))
			)
			WHERE rn >= (cnt * 0.99)
			GROUP BY ts ORDER BY ts
		`, timeExpr, unionQuery)
	default:
		return nil, fmt.Errorf("unsupported metric: %s", metric)
	}

	// Prepare arguments (each table needs startTime and endTime)
	var args []interface{}
	for range tableNames {
		args = append(args, startTime, endTime)
	}

	rows, err := s.db.Query(finalQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}
	defer rows.Close()

	var points []TimeSeriesPoint
	for rows.Next() {
		var p TimeSeriesPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, fmt.Errorf("scan failed: %w", err)
		}
		// Round value to 2 decimal places
		p.Value = math.Round(p.Value*100) / 100
		points = append(points, p)
	}

	return &TimeSeriesResult{
		Metric:   metric,
		Interval: interval,
		Points:   points,
	}, nil
}

// Close closes the database connection.
func (s *SqliteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
