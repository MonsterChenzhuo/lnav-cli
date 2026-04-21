package dryrun_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_SQL_DryRun(t *testing.T) {
	out, err := exec.Command("go", "run", "../../..",
		"+sql", "--dry-run", "-s", "/tmp/a.log", "SELECT 1").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(out)
	for _, want := range []string{"lnav", "-q", "SELECT 1", "/tmp/a.log"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q: %s", want, s)
		}
	}
}
