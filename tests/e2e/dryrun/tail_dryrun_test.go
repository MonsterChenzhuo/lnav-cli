package dryrun_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Tail_DryRun_WithDuration(t *testing.T) {
	out, err := exec.Command("go", "run", "../../..",
		"+tail", "--dry-run", "--duration", "5s", "-s", "/tmp/a.log", "panic").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(out)
	for _, want := range []string{"lnav", ":filter-in panic", ":goto 0"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q: %s", want, s)
		}
	}
}

func TestCLI_Tail_RejectsUnbounded(t *testing.T) {
	cmd := exec.Command("go", "run", "../../..",
		"+tail", "--dry-run", "-s", "/tmp/a.log", "panic")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatalf("expected non-zero exit; got success with: %s", out)
	}
	if !strings.Contains(string(out), "unbounded_tail") {
		t.Fatalf("expected unbounded_tail in stderr: %s", out)
	}
}
