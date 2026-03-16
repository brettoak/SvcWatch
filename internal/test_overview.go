package main

import (
	"fmt"
	"log"
	"time"

	"SvcWatch/internal/model"
	"SvcWatch/internal/storage"
)

func main() {
	dbPath := "./test_overview.db"
	store := storage.NewSqliteStorage(dbPath)
	defer store.Close()

	tableName := "test_logs"
	store.InitTable(tableName, true)

	// Insert test data
	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// Yesterday: 2 total requests, 1 success (50%), 1 error (50%), Avg time: 1.5s
	store.Save(tableName, &model.LogEntry{Status: 200, RequestTime: 1.0, TimeLocal: yesterday})
	store.Save(tableName, &model.LogEntry{Status: 500, RequestTime: 2.0, TimeLocal: yesterday})

	// Today: 4 total requests, 3 success (75%), 1 error (25%), Avg time: 0.5s
	store.Save(tableName, &model.LogEntry{Status: 200, RequestTime: 0.2, TimeLocal: now})
	store.Save(tableName, &model.LogEntry{Status: 200, RequestTime: 0.5, TimeLocal: now})
	store.Save(tableName, &model.LogEntry{Status: 200, RequestTime: 0.8, TimeLocal: now})
	store.Save(tableName, &model.LogEntry{Status: 404, RequestTime: 0.5, TimeLocal: now})

	startStr := now.Add(-1 * time.Hour).Format(time.RFC3339)
	endStr := now.Add(1 * time.Hour).Format(time.RFC3339)

	stats, err := store.GetOverviewWithCompare(tableName, startStr, endStr)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total Requests: %.0f (Compare: %.2f%%)\n", stats.TotalRequests.Value, stats.TotalRequests.ComparePercent)
	fmt.Printf("Success Rate: %.2f%% (Compare: %.2f%%)\n", stats.SuccessRate.Value, stats.SuccessRate.ComparePercent)
	fmt.Printf("Error Rate: %.2f%% (Compare: %.2f%%)\n", stats.ErrorRate.Value, stats.ErrorRate.ComparePercent)
	fmt.Printf("Avg Response Time: %.3fs (Compare: %.2f%%)\n", stats.AvgResponseTime.Value, stats.AvgResponseTime.ComparePercent)
}
