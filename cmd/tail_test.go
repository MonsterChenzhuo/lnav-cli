package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestTail_RequiresBound(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+tail", "--dry-run", "-s", "/tmp/a.log", "panic"})
	err := root.Execute()
	if err == nil || !strings.Contains(err.Error(), "unbounded_tail") {
		t.Fatalf("expected unbounded_tail error, got: %v", err)
	}
}

func TestTail_DryRun_WithDuration(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+tail", "--dry-run", "--duration", "10s", "-s", "/tmp/a.log", "panic"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{"lnav", ":filter-in panic", ":goto 0"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
}
