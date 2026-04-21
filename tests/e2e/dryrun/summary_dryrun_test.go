package dryrun_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Summary_DryRun(t *testing.T) {
	out, err := exec.Command("go", "run", "../../..",
		"+summary", "--dry-run", "-s", "/tmp/a.log").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(out)
	if !strings.Contains(s, "SELECT log_level") || !strings.Contains(s, "strftime(") {
		t.Fatalf("unexpected output: %s", s)
	}
}
