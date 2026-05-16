package parser

import (
	"fmt"
	"SvcWatch/internal/model"
	"SvcWatch/internal/utils"
	"regexp"
	"strconv"
	"time"
)

// Default Nginx log format regex
// $remote_addr - $remote_user [$time_local] "$request" $status $body_bytes_sent "$http_referer" "$http_user_agent" optionally ending with $request_time
var logRegex = regexp.MustCompile(`^(\S+) \- (\S+) \[(.*?)\] "(.*?)" (\d+) (\d+) "(.*?)" "(.*?)"(?:\s+(\d+(?:\.\d+)?))?`)

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
	// Normalize to UTC for consistent storage and querying
	timeLocal = timeLocal.UTC()

	status, _ := strconv.Atoi(matches[5])
	bodyBytes, _ := strconv.Atoi(matches[6])

	reqTime := 0.0
	if len(matches) > 9 && matches[9] != "" {
		reqTime, _ = strconv.ParseFloat(matches[9], 64)
	}

	entry := &model.LogEntry{
		RemoteAddr:    matches[1],
		RemoteUser:    matches[2],
		TimeLocal:     timeLocal,
		Request:       matches[4],
		Status:        status,
		BodyBytesSent: bodyBytes,
		HttpReferer:   matches[7],
		HttpUserAgent: matches[8],
		RequestTime:   reqTime,
	}

	// Lookup Geo Location
	geo := utils.LookupIP(entry.RemoteAddr)
	if geo != nil {
		entry.Country = geo.Country
		entry.Region = geo.Region
		entry.City = geo.City
		entry.Latitude = geo.Latitude
		entry.Longitude = geo.Longitude
	}

	return entry, nil
}
