package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSearch_DryRun_PrintsLnavArgv(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+search", "--dry-run", "--since", "1h", "--level", "error", "-s", "/tmp/foo.log", "timeout"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"lnav",
		":filter-in timeout",
		":set-min-log-level error",
		":hide-unmarked-lines-before",
		":write-json-to -",
		"/tmp/foo.log",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in argv; got:\n%s", want, got)
		}
	}
}

func TestSearch_RequiresSource(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+search", "--dry-run", "anything"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when no --source given")
	}
}
