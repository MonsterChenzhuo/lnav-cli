package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSummary_DryRun_EmitsThreeQueries(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+summary", "--dry-run", "-s", "/tmp/a.log"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"SELECT log_level, count(*) AS c",
		"LIMIT 10",
		"strftime(",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q: %s", want, got)
		}
	}
}
