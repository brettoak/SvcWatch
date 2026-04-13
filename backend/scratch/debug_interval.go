package main

import (
	"fmt"
	"math"
	"time"
)

func main() {
	startStr := "2026-01-13T02:30:00.000Z"
	endStr := "2026-04-13T02:30:00.000Z"

	startT, _ := time.Parse(time.RFC3339, startStr)
	endT, _ := time.Parse(time.RFC3339, endStr)

	duration := endT.Sub(startT)
	fmt.Printf("Duration: %v (%.0f seconds)\n", duration, duration.Seconds())

	type tier struct {
		name string
		sec  float64
	}
	tiers := []tier{
		{"1m", 60}, {"2m", 120}, {"5m", 300}, {"10m", 600}, {"30m", 1800},
		{"1h", 3600}, {"2h", 7200}, {"6h", 21600}, {"12h", 43200},
		{"1d", 86400}, {"2d", 172800}, {"3d", 259200}, {"4d", 345600}, {"5d", 432000},
		{"1w", 604800}, {"2w", 1209600}, {"1M", 2592000},
	}

	bestInterval := tiers[len(tiers)-1].name
	minDiff := math.MaxFloat64
	for _, t := range tiers {
		pointsCount := duration.Seconds() / t.sec
		diff := math.Abs(pointsCount - 20)
		fmt.Printf("Tier: %s, Points: %.2f, Diff: %.2f\n", t.name, pointsCount, diff)
		if diff < minDiff {
			minDiff = diff
			bestInterval = t.name
		}
	}
	fmt.Printf("Best Interval: %s\n", bestInterval)
}
