package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSQL_DryRun(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+sql", "--dry-run", "-s", "/tmp/a.log", "SELECT 1"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{"lnav", "-q", "SELECT 1", "/tmp/a.log"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
}

func TestSQL_ShowSchema(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+sql", "--dry-run", "--show-schema", "-s", "/tmp/a.log"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if !strings.Contains(out.String(), ".schema") {
		t.Fatalf("expected .schema in argv: %s", out.String())
	}
}

func TestSQL_RequiresQueryOrSchema(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+sql", "--dry-run", "-s", "/tmp/a.log"})
	if err := root.Execute(); err == nil || !strings.Contains(err.Error(), "missing_query") {
		t.Fatalf("expected missing_query, got: %v", err)
	}
}
