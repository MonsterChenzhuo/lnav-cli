package cmd

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/lnavexec"
	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/source"
)

type tailOpts struct {
	pattern  string
	level    string
	duration time.Duration
	maxEv    int
}

func newTailCmd() *cobra.Command {
	var local tailOpts
	c := &cobra.Command{
		Use:   "+tail [pattern]",
		Short: "follow log sources and emit matching events; MUST be bounded",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				local.pattern = args[0]
			}
			if local.duration == 0 && local.maxEv == 0 {
				return output.Errorf("unbounded_tail", "+tail requires --duration or --max-events to prevent hanging the agent")
			}
			if len(globalOpts.Sources) == 0 {
				return output.Errorf("missing_source", "at least one --source / -s is required")
			}
			cfg, err := source.Load(configPath())
			if err != nil {
				return output.Errorf("load_config", "%v", err)
			}
			resolved, err := cfg.Resolve(globalOpts.Sources)
			if err != nil {
				return output.Errorf("resolve_source", "%v", err)
			}
			if resolved.StdinCmd != "" {
				return output.Errorf("unsupported_stdin_source", "command-backed source not yet wired into +tail")
			}
			runner := lnavexec.Runner{
				DryRun: globalOpts.DryRun,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			}
			argv := lnavexec.BuildTailArgs(lnavexec.TailOpts{
				Pattern: local.pattern,
				Level:   local.level,
				Files:   resolved.Files,
			})
			if globalOpts.DryRun || local.duration == 0 {
				return runner.Run(argv)
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), local.duration)
			defer cancel()
			return runner.RunCtx(ctx, argv)
		},
	}
	c.Flags().StringVar(&local.pattern, "pattern", "", "regex (alternative to positional)")
	c.Flags().StringVar(&local.level, "level", "", "minimum log level")
	c.Flags().DurationVar(&local.duration, "duration", 0, "auto-exit after this time (e.g. 30s, 5m)")
	c.Flags().IntVar(&local.maxEv, "max-events", 0, "auto-exit after N events (flag accepted in v1.0; strict enforcement in v1.x)")
	return c
}
