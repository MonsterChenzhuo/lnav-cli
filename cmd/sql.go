package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/lnavexec"
	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/source"
)

func newSQLCmd() *cobra.Command {
	var showSchema bool
	c := &cobra.Command{
		Use:   "+sql [query]",
		Short: "run a SQLite query against log files via lnav",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
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
				return output.Errorf("unsupported_stdin_source", "command-backed source not yet wired into +sql")
			}
			var query string
			switch {
			case showSchema:
				query = ".schema"
			case len(args) == 1:
				query = args[0]
			default:
				return output.Errorf("missing_query", "either provide a query positional arg or --show-schema")
			}
			runner := lnavexec.Runner{
				DryRun: globalOpts.DryRun,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			}
			return runner.Run(lnavexec.BuildSQLArgs(lnavexec.SQLOpts{Query: query, Files: resolved.Files}))
		},
	}
	c.Flags().BoolVar(&showSchema, "show-schema", false, "print lnav SQLite schema instead of running a query")
	return c
}
