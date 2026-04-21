package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSetup_PrintsGuidance(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"setup"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "lnav") || !strings.Contains(got, "doctor") {
		t.Fatalf("unexpected guidance: %s", got)
	}
}
