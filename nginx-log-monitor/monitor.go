package monitor

import (
	"fmt"
	"nginx-log-monitor/internal/collector"
	"nginx-log-monitor/internal/parser"
	"nginx-log-monitor/internal/storage"
)

// Monitor is the main entry point for the log monitoring service.
type Monitor struct {
	collector *collector.LogCollector
	storage   *storage.SqliteStorage
}

// NewMonitor creates a new Monitor instance.
func NewMonitor(logPath string, clearOnStartup bool) (*Monitor, error) {
	coll, err := collector.NewLogCollector(logPath)
	if err != nil {
		return nil, err
	}
	store := storage.NewSqliteStorage("nginx_logs.db", clearOnStartup)
	return &Monitor{
		collector: coll,
		storage:   store,
	}, nil
}

// Start begins collecting and processing logs.
// It runs in the background.
func (m *Monitor) Start() {
	m.collector.Start()
	go func() {
		for line := range m.collector.DataChannel {
			entry, err := parser.Parse(line)
			if err != nil {
				// simple error logging, could use a proper logger
				fmt.Printf("Parse error: %v, line: %s\n", err, line)
				continue
			}
			m.storage.Save(entry)
		}
	}()
}

// GetStats returns current statistics.
func (m *Monitor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_logs": m.storage.GetTotalCount(),
	}
}
