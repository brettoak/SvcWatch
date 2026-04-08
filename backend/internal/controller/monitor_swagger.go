package controller

import (
	"SvcWatch/internal/storage"
)

// OverviewResponseWrapper is the Swagger representation of the Overview API response.
type OverviewResponseWrapper struct {
	Code    int                  `json:"code" example:"200"`
	Message string               `json:"message" example:"success"`
	Data    storage.OverviewStats `json:"data"`
}

// StatusDistributionResponseWrapper is the Swagger representation of the Status Distribution API response.
type StatusDistributionResponseWrapper struct {
	Code    int                              `json:"code" example:"200"`
	Message string                           `json:"message" example:"success"`
	Data    storage.StatusDistributionResult `json:"data"`
}

// TimeSeriesResponseWrapper is the Swagger representation of the Time Series API response.
type TimeSeriesResponseWrapper struct {
	Code    int                       `json:"code" example:"200"`
	Message string                    `json:"message" example:"success"`
	Data    storage.TimeSeriesResult `json:"data"`
}

// LogsResponseWrapper is the Swagger representation of the Logs API response.
type LogsResponseWrapper struct {
	Code    int                      `json:"code" example:"200"`
	Message string                   `json:"message" example:"success"`
	Data    storage.LogQueryResponse `json:"data"`
}

// StatsResponseWrapper is the Swagger representation of the generic Stats API response.
type StatsResponseWrapper struct {
	Code    int                    `json:"code" example:"200"`
	Message string                 `json:"message" example:"success"`
	Data    map[string]interface{} `json:"data"`
}
