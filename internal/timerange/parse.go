// Package timerange parses --since/--until values into concrete instants.
package timerange

import (
	"fmt"
	"time"
)

// Parse converts a user-facing time expression into an absolute time, using time.Now as the anchor.
func Parse(s string) (time.Time, error) {
	return parseAt(s, time.Now())
}

func parseAt(s string, now time.Time) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time expression")
	}
	if d, err := time.ParseDuration(s); err == nil {
		return now.Add(-d), nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time expression %q (use e.g. 1h, 2026-04-21T10:00:00Z)", s)
}
