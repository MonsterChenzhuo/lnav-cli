package lnavexec

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunner_DryRun_PrintsArgv(t *testing.T) {
	out := &bytes.Buffer{}
	r := Runner{Binary: "lnav", DryRun: true, Stdout: out}
	args := []string{"-n", "-c", ":filter-in foo", "--", "/tmp/x.log"}
	if err := r.Run(args); err != nil {
		t.Fatalf("run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "lnav -n -c ':filter-in foo' -- /tmp/x.log") {
		t.Fatalf("unexpected dry-run output: %q", got)
	}
}

func TestRunner_MissingBinary_ReturnsNotFoundErr(t *testing.T) {
	r := Runner{Binary: "definitely-not-a-real-binary-xyz", DryRun: false}
	err := r.Run([]string{"-n"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "lnav_not_found") {
		t.Fatalf("wrong error code: %v", err)
	}
}
