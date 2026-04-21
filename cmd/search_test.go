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

func TestSearch_ResolvesAlias(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LNAV_CLI_CONFIG_DIR", dir)

	add := NewRootCmd()
	addBuf := &bytes.Buffer{}
	add.SetOut(addBuf)
	add.SetErr(addBuf)
	add.SetArgs([]string{"source", "add", "app", "--paths", "/tmp/app1.log,/tmp/app2.log"})
	if err := add.Execute(); err != nil {
		t.Fatalf("seed add: %v (%s)", err, addBuf.String())
	}

	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+search", "--dry-run", "-s", "app", "panic"})
	if err := root.Execute(); err != nil {
		t.Fatalf("search: %v", err)
	}
	got := out.String()
	for _, want := range []string{"/tmp/app1.log", "/tmp/app2.log", ":filter-in panic"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in argv:\n%s", want, got)
		}
	}
}
