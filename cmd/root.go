// Package cmd hosts the cobra command tree for lnav-cli.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type GlobalOpts struct {
	Sources []string
	Since   string
	Until   string
	Format  string
	Limit   int
	DryRun  bool
}

var globalOpts GlobalOpts

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "lnav-cli",
		Short:         "Agent-friendly wrapper around lnav",
		Long:          "lnav-cli — drive lnav from Claude Code with one-line commands.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringSliceVarP(&globalOpts.Sources, "source", "s", nil, "log source alias, file, or glob (repeatable)")
	root.PersistentFlags().StringVar(&globalOpts.Since, "since", "", "start of time window (e.g. 1h, 2026-04-21T10:00:00Z)")
	root.PersistentFlags().StringVar(&globalOpts.Until, "until", "", "end of time window")
	root.PersistentFlags().StringVar(&globalOpts.Format, "format", "ndjson", "output format: ndjson|json|table|pretty")
	root.PersistentFlags().IntVar(&globalOpts.Limit, "limit", 0, "max rows returned (0 = unbounded)")
	root.PersistentFlags().BoolVar(&globalOpts.DryRun, "dry-run", false, "print the lnav argv instead of executing")

	root.AddCommand(newDoctorCmd())
	root.AddCommand(newSearchCmd())
	return root
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}
