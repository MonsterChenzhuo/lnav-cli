package lnavexec

import (
	"reflect"
	"testing"
)

func TestBuildSearchArgs_Minimal(t *testing.T) {
	got := BuildSearchArgs(SearchOpts{Files: []string{"/tmp/a.log"}})
	want := []string{"-n", "-c", ":write-json-to -", "--", "/tmp/a.log"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildSearchArgs_Full(t *testing.T) {
	got := BuildSearchArgs(SearchOpts{
		Pattern: "timeout|refused",
		Level:   "error",
		SinceTS: "2026-04-21T10:00:00Z",
		Files:   []string{"/var/log/nginx/error.log"},
	})
	want := []string{
		"-n",
		"-c", ":filter-in timeout|refused",
		"-c", ":set-min-log-level error",
		"-c", ":hide-unmarked-lines-before 2026-04-21T10:00:00Z",
		"-c", ":write-json-to -",
		"--",
		"/var/log/nginx/error.log",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v\nwant=%v", got, want)
	}
}

func TestBuildSearchArgs_RejectsShellInjection(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on newline in pattern")
		}
	}()
	_ = BuildSearchArgs(SearchOpts{Pattern: "bad\npattern", Files: []string{"/tmp/a.log"}})
}
