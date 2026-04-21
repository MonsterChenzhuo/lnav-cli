package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/source"
)

// configPath returns the sources.yaml path, honoring LNAV_CLI_CONFIG_DIR for tests.
func configPath() string {
	if d := os.Getenv("LNAV_CLI_CONFIG_DIR"); d != "" {
		return filepath.Join(d, "sources.yaml")
	}
	return source.DefaultPath()
}

func newSourceCmd() *cobra.Command {
	c := &cobra.Command{Use: "source", Short: "manage named log sources"}
	c.AddCommand(newSourceAddCmd(), newSourceLsCmd(), newSourceShowCmd(), newSourceRmCmd())
	return c
}

func newSourceAddCmd() *cobra.Command {
	var paths, command, level string
	cmd := &cobra.Command{
		Use:   "add NAME",
		Short: "register a named source (use --paths or --command)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if paths == "" && command == "" {
				return output.Errorf("missing_input", "either --paths or --command is required")
			}
			cfg, err := source.Load(configPath())
			if err != nil {
				return output.Errorf("load_config", "%v", err)
			}
			s := source.Source{DefaultLevel: level}
			if command != "" {
				s.Command = command
			} else {
				for _, p := range strings.Split(paths, ",") {
					p = strings.TrimSpace(p)
					if p != "" {
						s.Paths = append(s.Paths, p)
					}
				}
			}
			cfg.Sources[name] = s
			if err := source.Save(configPath(), cfg); err != nil {
				return output.Errorf("save_config", "%v", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "added source %q\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&paths, "paths", "", "comma-separated paths/globs")
	cmd.Flags().StringVar(&command, "command", "", "shell command whose stdout feeds lnav")
	cmd.Flags().StringVar(&level, "default-level", "", "default --level when omitted")
	return cmd
}

func newSourceLsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "ls",
		Short: "list registered sources",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := source.Load(configPath())
			if err != nil {
				return output.Errorf("load_config", "%v", err)
			}
			names := make([]string, 0, len(cfg.Sources))
			for n := range cfg.Sources {
				names = append(names, n)
			}
			sort.Strings(names)
			out := cmd.OutOrStdout()
			for _, n := range names {
				s := cfg.Sources[n]
				switch {
				case s.Command != "":
					fmt.Fprintf(out, "%s\tcommand=%s\n", n, s.Command)
				default:
					fmt.Fprintf(out, "%s\tpaths=%s\n", n, strings.Join(s.Paths, ","))
				}
			}
			return nil
		},
	}
}

func newSourceShowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "show NAME",
		Short: "show a source's raw definition",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := source.Load(configPath())
			if err != nil {
				return output.Errorf("load_config", "%v", err)
			}
			s, ok := cfg.Sources[args[0]]
			if !ok {
				return output.Errorf("unknown_source", "no source named %q", args[0])
			}
			out := cmd.OutOrStdout()
			if s.Command != "" {
				fmt.Fprintf(out, "command: %s\n", s.Command)
			}
			if len(s.Paths) > 0 {
				fmt.Fprintf(out, "paths:\n")
				for _, p := range s.Paths {
					fmt.Fprintf(out, "  - %s\n", p)
				}
			}
			if s.DefaultLevel != "" {
				fmt.Fprintf(out, "default_level: %s\n", s.DefaultLevel)
			}
			return nil
		},
	}
}

func newSourceRmCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rm NAME",
		Short: "remove a source",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := source.Load(configPath())
			if err != nil {
				return output.Errorf("load_config", "%v", err)
			}
			if _, ok := cfg.Sources[args[0]]; !ok {
				return output.Errorf("unknown_source", "no source named %q", args[0])
			}
			delete(cfg.Sources, args[0])
			if err := source.Save(configPath(), cfg); err != nil {
				return output.Errorf("save_config", "%v", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "removed source %q\n", args[0])
			return nil
		},
	}
}
