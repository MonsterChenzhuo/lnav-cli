package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/lnavexec"
	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/timerange"
)

type searchOpts struct {
	pattern string
	level   string
}

func newSearchCmd() *cobra.Command {
	var local searchOpts
	c := &cobra.Command{
		Use:   "+search [pattern]",
		Short: "regex/level search across one or more log sources",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				local.pattern = args[0]
			}
			if len(globalOpts.Sources) == 0 {
				return output.Errorf("missing_source", "at least one --source / -s is required").
					WithHint("pass a path, glob, or registered alias")
			}
			so := lnavexec.SearchOpts{
				Pattern: local.pattern,
				Level:   local.level,
				Files:   globalOpts.Sources,
			}
			if globalOpts.Since != "" {
				ts, err := timerange.Parse(globalOpts.Since)
				if err != nil {
					return output.Errorf("bad_since", "%v", err)
				}
				so.SinceTS = ts.UTC().Format("2006-01-02T15:04:05Z")
			}
			if globalOpts.Until != "" {
				ts, err := timerange.Parse(globalOpts.Until)
				if err != nil {
					return output.Errorf("bad_until", "%v", err)
				}
				so.UntilTS = ts.UTC().Format("2006-01-02T15:04:05Z")
			}
			runner := lnavexec.Runner{
				DryRun: globalOpts.DryRun,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			}
			return runner.Run(lnavexec.BuildSearchArgs(so))
		},
	}
	c.Flags().StringVar(&local.pattern, "pattern", "", "regex (alternative to positional)")
	c.Flags().StringVar(&local.level, "level", "", "minimum log level (info|warning|error|fatal)")
	return c
}
