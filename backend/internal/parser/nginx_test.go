package parser

import (
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	// Sample log line
	line := `127.0.0.1 - - [24/Jan/2026:15:00:00 +1100] "GET /ping HTTP/1.1" 200 13 "-" "Go-http-client/1.1"`

	entry, err := Parse(line)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if entry.RemoteAddr != "127.0.0.1" {
		t.Errorf("Expected RemoteAddr 127.0.0.1, got %s", entry.RemoteAddr)
	}
	if entry.Status != 200 {
		t.Errorf("Expected Status 200, got %d", entry.Status)
	}
	if entry.Request != "GET /ping HTTP/1.1" {
		t.Errorf("Expected Request 'GET /ping HTTP/1.1', got '%s'", entry.Request)
	}
	
	expectedTime, _ := time.Parse("02/Jan/2006:15:04:05 -0700", "24/Jan/2026:15:00:00 +1100")
	if !entry.TimeLocal.Equal(expectedTime) {
		t.Errorf("Expected TimeLocal %v, got %v", expectedTime, entry.TimeLocal)
	}
}

func TestParseInvalid(t *testing.T) {
	line := "invalid log line"
	_, err := Parse(line)
	if err == nil {
		t.Error("Expected error for invalid log line, got nil")
	}
}
