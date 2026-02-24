package monitor

import (
	"fmt"
	"nginx-log-monitor/internal/collector"
	"nginx-log-monitor/internal/parser"
	storage "nginx-log-monitor/storagePkg"
)

// Monitor is the main entry point for the log monitoring service for a single table.
type Monitor struct {
	collector *collector.LogCollector
	storage   *storage.SqliteStorage
	tableName string
}

// NewMonitor creates a new Monitor instance.
func NewMonitor(logPath string, store *storage.SqliteStorage, tableName string) (*Monitor, error) {
	coll, err := collector.NewLogCollector(logPath)
	if err != nil {
		return nil, err
	}
	return &Monitor{
		collector: coll,
		storage:   store,
		tableName: tableName,
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
			m.storage.Save(m.tableName, entry)
		}
	}()
}

// GetStats returns current statistics for this monitor's table.
func (m *Monitor) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_logs": m.storage.GetTotalCount(m.tableName),
	}
}
