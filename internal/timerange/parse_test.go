package timerange

import (
	"testing"
	"time"
)

func TestParse_Relative(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	got, err := parseAt("1h", now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := now.Add(-time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_RFC3339(t *testing.T) {
	got, err := parseAt("2026-04-20T10:30:00Z", time.Now())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.UTC().Format(time.RFC3339) != "2026-04-20T10:30:00Z" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestParse_Empty(t *testing.T) {
	if _, err := parseAt("", time.Now()); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestParse_Invalid(t *testing.T) {
	if _, err := parseAt("tomorrow", time.Now()); err == nil {
		t.Fatal("expected error on unsupported input")
	}
}
