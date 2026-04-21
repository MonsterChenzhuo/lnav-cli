package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the lnav-cli environment (lnav on PATH, version, permissions)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			path, err := exec.LookPath("lnav")
			if err != nil {
				fmt.Fprintln(out, "lnav: MISSING")
				fmt.Fprintln(out, "hint: run `lnav-cli setup` or `brew install lnav`")
				return nil
			}
			ver, verr := exec.Command("lnav", "-V").Output()
			fmt.Fprintf(out, "lnav: OK (%s)\n", path)
			if verr == nil {
				fmt.Fprintf(out, "version: %s", string(ver))
			}
			return nil
		},
	}
}
