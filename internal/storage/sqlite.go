package storage

import (
	"database/sql"
	"fmt"
	"log"
	"math"
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

// BaseMetrics contains raw metric counts for a specific time period.
type BaseMetrics struct {
	TotalRequests   float64
	SuccessCount    float64
	ErrorCount      float64
	AvgResponseTime float64
}

// GetBaseMetrics queries the raw metrics for a specific time range.
func (s *SqliteStorage) GetBaseMetrics(tableName string, startTime, endTime string) (*BaseMetrics, error) {
	// The query calculates Total Request, Success Count (<400), Error Count (>=400) and Average Request Time.
	query := fmt.Sprintf(`
		SELECT 
			CAST(COUNT(*) AS REAL) as total_requests,
			CAST(SUM(CASE WHEN status < 400 THEN 1 ELSE 0 END) AS REAL) as success_count,
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
	// 1. Get current period metrics
	currentMetrics, err := s.GetBaseMetrics(tableName, startTimeStr, endTimeStr)
	if err != nil {
		return nil, err
	}

	// 2. Parse times to calculate the previous period
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

	// Calculate range duration
	rangeDuration := endTime.Sub(startTime)
	offset := 24 * time.Hour
	compareType := "vs yesterday"
	if rangeDuration > 24*time.Hour {
		offset = rangeDuration
		compareType = "vs previous period"
	}

	prevStartTime := startTime.Add(-offset).Format(layout)
	prevEndTime := endTime.Add(-offset).Format(layout)

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
			Value:          math.Round(currentMetrics.AvgResponseTime*1000) / 1000, // rounded to 3 decimals
			ComparePercent: calculateComparePercent(currentMetrics.AvgResponseTime, prevMetrics.AvgResponseTime),
		},
		CompareType: compareType,
	}

	return stats, nil
}

// Close closes the database connection.
func (s *SqliteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
