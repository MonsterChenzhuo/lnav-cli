package live_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI_Search_Live_5xx(t *testing.T) {
	if _, err := exec.LookPath("lnav"); err != nil {
		t.Skip("lnav not installed; skipping live E2E")
	}
	_, thisFile, _, _ := runtime.Caller(0)
	fixture := filepath.Join(filepath.Dir(thisFile), "..", "..", "fixtures", "nginx.log")

	out, err := exec.Command("go", "run", "../../..",
		"+search", "-s", fixture, `5\d\d`).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(bytes.TrimSpace(out))
	if !strings.Contains(s, "500") || !strings.Contains(s, "502") {
		t.Fatalf("expected 500 and 502 in output, got:\n%s", s)
	}
}
