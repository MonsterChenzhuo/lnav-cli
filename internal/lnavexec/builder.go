// Package lnavexec builds and executes lnav subprocess invocations.
package lnavexec

import (
	"fmt"
	"strings"
)

type SearchOpts struct {
	Pattern string // regex passed to :filter-in; empty means no filter
	Level   string // one of info|warning|error|fatal; empty means any
	SinceTS string // RFC3339 / lnav-compatible absolute timestamp
	UntilTS string
	Files   []string // positional args passed to lnav
}

// BuildSearchArgs produces the argv (minus the leading "lnav" binary) for a +search call.
// Inputs must not contain newlines; the function panics on violation because callers
// are expected to have run validation up-stream.
func BuildSearchArgs(o SearchOpts) []string {
	assertNoNewline(o.Pattern, "pattern")
	assertNoNewline(o.Level, "level")
	assertNoNewline(o.SinceTS, "since")
	assertNoNewline(o.UntilTS, "until")

	args := []string{"-n"}
	if o.Pattern != "" {
		args = append(args, "-c", ":filter-in "+o.Pattern)
	}
	if o.Level != "" {
		args = append(args, "-c", ":set-min-log-level "+o.Level)
	}
	if o.SinceTS != "" {
		args = append(args, "-c", ":hide-unmarked-lines-before "+o.SinceTS)
	}
	if o.UntilTS != "" {
		args = append(args, "-c", ":hide-unmarked-lines-after "+o.UntilTS)
	}
	args = append(args, "-c", ":write-json-to -", "--")
	args = append(args, o.Files...)
	return args
}

type SQLOpts struct {
	Query string
	Files []string
}

// BuildSQLArgs produces argv for `lnav -n -q <query> -- <files...>`.
func BuildSQLArgs(o SQLOpts) []string {
	assertNoNewline(o.Query, "query")
	args := []string{"-n", "-q", o.Query, "--"}
	args = append(args, o.Files...)
	return args
}

type TailOpts struct {
	Pattern string
	Level   string
	Files   []string
}

// BuildTailArgs produces argv for bounded follow; :goto 0 forces a single pass so
// lnav emits existing matches, and caller is expected to add -c :goto or rely on
// write-json-to streaming + context timeout.
func BuildTailArgs(o TailOpts) []string {
	assertNoNewline(o.Pattern, "pattern")
	assertNoNewline(o.Level, "level")
	args := []string{"-n"}
	if o.Pattern != "" {
		args = append(args, "-c", ":filter-in "+o.Pattern)
	}
	if o.Level != "" {
		args = append(args, "-c", ":set-min-log-level "+o.Level)
	}
	args = append(args, "-c", ":write-json-to -", "-c", ":goto 0", "--")
	args = append(args, o.Files...)
	return args
}

// BuildSummaryQueries returns the three SQL queries used by +summary:
// level distribution, top-N error bodies, and a time histogram.
func BuildSummaryQueries(topN int, histogramBucket string) []string {
	if topN <= 0 {
		topN = 10
	}
	if histogramBucket == "" {
		histogramBucket = "5m"
	}
	bucket := "%Y-%m-%d %H:%M"
	if histogramBucket == "1h" {
		bucket = "%Y-%m-%d %H:00"
	}
	return []string{
		"SELECT log_level, count(*) AS c FROM all_logs WHERE log_level IN ('error','warning','fatal') GROUP BY log_level ORDER BY c DESC",
		fmt.Sprintf("SELECT log_level, log_body, count(*) AS c FROM all_logs WHERE log_level IN ('error','warning','fatal') GROUP BY log_body ORDER BY c DESC LIMIT %d", topN),
		fmt.Sprintf("SELECT strftime('%s', log_time) AS bin, count(*) AS c FROM all_logs WHERE log_level IN ('error','fatal') GROUP BY bin ORDER BY bin", bucket),
	}
}

func assertNoNewline(s, field string) {
	if strings.ContainsAny(s, "\r\n") {
		panic("lnavexec: newline in " + field + " — input validation missing up-stream")
	}
}
