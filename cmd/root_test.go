package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_HelpIncludesShortcuts(t *testing.T) {
	buf := &bytes.Buffer{}
	root := NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"+search", "doctor", "--format", "--dry-run"} {
		if !strings.Contains(out, want) {
			t.Errorf("root --help missing %q; got:\n%s", want, out)
		}
	}
}
