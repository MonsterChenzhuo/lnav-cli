package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestDoctor_Reports(t *testing.T) {
	out := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := newDoctorCmd()
	cmd.SetOut(out)
	cmd.SetErr(stderr)
	_ = cmd.Execute()
	report := out.String()
	if _, err := exec.LookPath("lnav"); err == nil {
		if !strings.Contains(report, "lnav: OK") {
			t.Fatalf("expected 'lnav: OK' when lnav is on PATH, got:\n%s", report)
		}
	} else {
		if !strings.Contains(report, "lnav: MISSING") {
			t.Fatalf("expected 'lnav: MISSING' when lnav absent, got:\n%s", report)
		}
	}
}
