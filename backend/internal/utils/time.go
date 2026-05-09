package utils

import (
	"fmt"
	"time"
)

const (
	// MaxTimeRangeLimit is the default maximum allowed range for statistics queries (1 year).
	MaxTimeRangeLimit = 366 * 24 * time.Hour
)

// ParseTime parses a time string into a time.Time object.
// It supports RFC3339Nano (JS toISOString), RFC3339, datetime-local, and "2006-01-02 15:04:05".
func ParseTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,       // "2026-05-09T12:00:00.000Z" (from JS toISOString)
		time.RFC3339,           // "2026-05-09T12:00:00Z"
		"2006-01-02T15:04:05",  // datetime-local without timezone
		"2006-01-02T15:04",     // datetime-local short
		"2006-01-02 15:04:05",  // legacy format
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time format: %q", s)
}

// ValidateTimeRange checks if the time range is logical and within limits.
func ValidateTimeRange(start, end time.Time, maxRange time.Duration) error {
	now := time.Now()
	
	// Basic logic checks
	if start.After(now) {
		return fmt.Errorf("start_time cannot be in the future")
	}
	if !end.After(start) {
		return fmt.Errorf("end_time must be after start_time")
	}
	
	// Optional range limit check
	if maxRange > 0 && end.Sub(start) > maxRange {
		return fmt.Errorf("time range cannot exceed %v", maxRange)
	}
	
	return nil
}

// ParseAndValidateRange is a helper that parses two time strings and validates the resulting range.
func ParseAndValidateRange(startStr, endStr string, maxRange time.Duration) (time.Time, time.Time, error) {
	startT, err := ParseTime(startStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_time format")
	}

	endT, err := ParseTime(endStr)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_time format")
	}

	if err := ValidateTimeRange(startT, endT, maxRange); err != nil {
		return time.Time{}, time.Time{}, err
	}

	return startT, endT, nil
}
