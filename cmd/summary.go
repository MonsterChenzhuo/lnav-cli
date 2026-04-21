package cmd

import (
	"bytes"
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/lnavexec"
	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/source"
)

func newSummaryCmd() *cobra.Command {
	var topN int
	var bucket string
	c := &cobra.Command{
		Use:   "+summary",
		Short: "error-level distribution + top-N errors + histogram",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
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
				return output.Errorf("unsupported_stdin_source", "command-backed source not yet wired into +summary")
			}
			queries := lnavexec.BuildSummaryQueries(topN, bucket)
			if globalOpts.DryRun {
				runner := lnavexec.Runner{
					DryRun: true,
					Stdout: cmd.OutOrStdout(),
					Stderr: cmd.ErrOrStderr(),
				}
				for _, q := range queries {
					if err := runner.Run(lnavexec.BuildSQLArgs(lnavexec.SQLOpts{Query: q, Files: resolved.Files})); err != nil {
						return err
					}
				}
				return nil
			}
			labels := []string{"levels", "top_errors", "histogram"}
			result := map[string]any{}
			for i, q := range queries {
				buf := &bytes.Buffer{}
				sub := lnavexec.Runner{Stdout: buf, Stderr: cmd.ErrOrStderr()}
				if err := sub.Run(lnavexec.BuildSQLArgs(lnavexec.SQLOpts{Query: q, Files: resolved.Files})); err != nil {
					return err
				}
				var rows any
				if len(bytes.TrimSpace(buf.Bytes())) == 0 {
					rows = []any{}
				} else if err := json.Unmarshal(buf.Bytes(), &rows); err != nil {
					return output.Errorf("bad_lnav_json", "lnav stdout not JSON: %v", err)
				}
				result[labels[i]] = rows
			}
			enc := json.NewEncoder(cmd.OutOrStdout())
			enc.SetEscapeHTML(false)
			return enc.Encode(map[string]any{"data": result, "_meta": map[string]any{"count": 3}})
		},
	}
	c.Flags().IntVar(&topN, "top", 10, "top N error bodies")
	c.Flags().StringVar(&bucket, "histogram", "5m", "histogram bucket: 1m|5m|1h")
	return c
}
