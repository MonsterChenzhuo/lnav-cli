package dryrun_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Search_DryRun(t *testing.T) {
	out, err := exec.Command("go", "run", "../../..",
		"+search", "--dry-run", "--since", "30m", "-s", "/var/log/app.log", "panic").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\noutput: %s", err, out)
	}
	s := string(bytes.TrimSpace(out))
	for _, want := range []string{"lnav", ":filter-in panic", ":write-json-to -", "/var/log/app.log"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in output:\n%s", want, s)
		}
	}
}
