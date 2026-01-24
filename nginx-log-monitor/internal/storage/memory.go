package storage

import (
	"nginx-log-monitor/internal/model"
	"sync"
)

// MemoryStorage implements in-memory storage.
type MemoryStorage struct {
	mu   sync.RWMutex
	logs []*model.LogEntry
}

// NewMemoryStorage creates a new MemoryStorage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		logs: make([]*model.LogEntry, 0),
	}
}

// Save saves a log entry to memory.
func (s *MemoryStorage) Save(entry *model.LogEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.logs = append(s.logs, entry)
	return nil
}

// GetTotalCount returns the total number of logs.
func (s *MemoryStorage) GetTotalCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.logs)
}
