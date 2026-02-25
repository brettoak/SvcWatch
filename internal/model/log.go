package model

import "time"

// LogEntry represents a single parsed Nginx log line.
// All comments are in English as requested.
type LogEntry struct {
	RemoteAddr    string    `json:"remote_addr"`
	RemoteUser    string    `json:"remote_user"`
	TimeLocal     time.Time `json:"time_local"`
	Request       string    `json:"request"`
	Status        int       `json:"status"`
	BodyBytesSent int       `json:"body_bytes_sent"`
	HttpReferer   string    `json:"http_referer"`
	HttpUserAgent string    `json:"http_user_agent"`
}
