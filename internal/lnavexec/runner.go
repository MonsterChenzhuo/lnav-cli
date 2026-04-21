package lnavexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
)

// Runner executes lnav as a subprocess; DryRun true prints argv instead of running.
type Runner struct {
	Binary string
	DryRun bool
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run executes `lnav args...`.
func (r Runner) Run(args []string) error {
	if r.Binary == "" {
		r.Binary = "lnav"
	}
	if r.Stdout == nil {
		r.Stdout = os.Stdout
	}
	if r.Stderr == nil {
		r.Stderr = os.Stderr
	}
	if r.DryRun {
		fmt.Fprintln(r.Stdout, prettyArgv(r.Binary, args))
		return nil
	}
	if _, err := exec.LookPath(r.Binary); err != nil {
		return output.Errorf("lnav_not_found", "lnav executable %q not found on PATH", r.Binary).
			WithHint("run: lnav-cli setup, or install manually (brew install lnav)")
	}
	cmd := exec.Command(r.Binary, args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if err := cmd.Run(); err != nil {
		return output.Errorf("lnav_exec_failed", "%v", err)
	}
	return nil
}

func prettyArgv(bin string, args []string) string {
	parts := []string{bin}
	for _, a := range args {
		if strings.ContainsAny(a, " \t\"'") {
			parts = append(parts, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
			continue
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}
