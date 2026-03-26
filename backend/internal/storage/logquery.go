package storage

import (
	"SvcWatch/internal/model"
)

// LogQueryFilter defines the criteria for querying logs.
type LogQueryFilter struct {
	Page        int
	Size        int
	StartTime   string
	EndTime     string
	IP          string
	Method      string
	Status      *int
	StatusClass string
	PathKeyword string
	MinLatency  *int
	MaxLatency  *int
	Sort        string
}

// LogQueryItem represents a single log entry returned in queries with its source.
type LogQueryItem struct {
	SourceID string         `json:"source_id"`
	Entry    model.LogEntry `json:"entry"`
}

// LogQueryResponse represents the paginated response for log queries.
type LogQueryResponse struct {
	Total int            `json:"total"`
	Page  int            `json:"page"`
	Size  int            `json:"size"`
	Items []LogQueryItem `json:"items"`
}
