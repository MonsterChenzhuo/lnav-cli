// Package lnavexec builds and executes lnav subprocess invocations.
package lnavexec

import "strings"

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

func assertNoNewline(s, field string) {
	if strings.ContainsAny(s, "\r\n") {
		panic("lnavexec: newline in " + field + " — input validation missing up-stream")
	}
}
