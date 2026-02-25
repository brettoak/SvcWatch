package parser

import (
	"fmt"
	"SvcWatch/internal/model"
	"regexp"
	"strconv"
	"time"
)

// Default Nginx log format regex
// $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent"
var logRegex = regexp.MustCompile(`^(\S+) \- (\S+) \[(.*?)\] "(.*?)" (\d+) (\d+) "(.*?)" "(.*?)"`)

// Parse parses a raw log line into a LogEntry.
func Parse(line string) (*model.LogEntry, error) {
	matches := logRegex.FindStringSubmatch(line)
	if len(matches) < 9 {
		return nil, fmt.Errorf("invalid log format: %s", line)
	}

	// Parse Time
	// Layout for [24/Jan/2026:15:00:00 +1100]
	layout := "02/Jan/2006:15:04:05 -0700"
	timeLocal, err := time.Parse(layout, matches[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse time: %v", err)
	}

	status, _ := strconv.Atoi(matches[5])
	bodyBytes, _ := strconv.Atoi(matches[6])

	return &model.LogEntry{
		RemoteAddr:    matches[1],
		RemoteUser:    matches[2],
		TimeLocal:     timeLocal,
		Request:       matches[4],
		Status:        status,
		BodyBytesSent: bodyBytes,
		HttpReferer:   matches[7],
		HttpUserAgent: matches[8],
	}, nil
}
