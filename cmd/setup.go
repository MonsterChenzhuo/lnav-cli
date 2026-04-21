package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

func newSetupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "setup",
		Short: "print the recommended installation command for lnav on this platform",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			if _, err := exec.LookPath("lnav"); err == nil {
				fmt.Fprintln(out, "lnav already installed — nothing to do")
				fmt.Fprintln(out, "re-run: lnav-cli doctor")
				return nil
			}
			switch runtime.GOOS {
			case "darwin":
				fmt.Fprintln(out, "install lnav via Homebrew:")
				fmt.Fprintln(out, "  brew install lnav")
			case "linux":
				fmt.Fprintln(out, "install lnav via your package manager, e.g.:")
				fmt.Fprintln(out, "  sudo apt install lnav     # Debian/Ubuntu")
				fmt.Fprintln(out, "  sudo dnf install lnav     # Fedora")
			case "windows":
				fmt.Fprintln(out, "windows support is best-effort; recommended path:")
				fmt.Fprintln(out, "  winget install lnav       # or download from https://lnav.org")
			default:
				fmt.Fprintln(out, "platform not auto-detected; see https://lnav.org/downloads.html")
			}
			fmt.Fprintln(out, "after installing, re-run: lnav-cli doctor")
			return nil
		},
	}
}
