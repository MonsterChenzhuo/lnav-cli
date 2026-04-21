package lnavexec

import (
	"reflect"
	"strings"
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

func TestBuildSQLArgs(t *testing.T) {
	got := BuildSQLArgs(SQLOpts{Query: "SELECT 1", Files: []string{"/tmp/a.log"}})
	want := []string{"-n", "-q", "SELECT 1", "--", "/tmp/a.log"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildSQLArgs_RejectsNewline(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on newline in query")
		}
	}()
	_ = BuildSQLArgs(SQLOpts{Query: "SELECT\n1", Files: []string{"/tmp/a.log"}})
}

func TestBuildTailArgs(t *testing.T) {
	got := BuildTailArgs(TailOpts{Pattern: "5\\d\\d", Level: "error", Files: []string{"/tmp/a.log"}})
	want := []string{
		"-n",
		"-c", ":filter-in 5\\d\\d",
		"-c", ":set-min-log-level error",
		"-c", ":write-json-to -",
		"-c", ":goto 0",
		"--",
		"/tmp/a.log",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v\nwant=%v", got, want)
	}
}

func TestBuildSummaryQueries_Shape(t *testing.T) {
	qs := BuildSummaryQueries(10, "1m")
	if len(qs) != 3 {
		t.Fatalf("expected 3 queries, got %d", len(qs))
	}
	if !strings.Contains(qs[0], "log_level") || !strings.Contains(qs[1], "LIMIT 10") || !strings.Contains(qs[2], "strftime") {
		t.Fatalf("unexpected queries: %+v", qs)
	}
}
