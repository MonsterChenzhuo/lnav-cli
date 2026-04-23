// Copyright (c) 2026 lnav-cli authors
// SPDX-License-Identifier: MIT

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/build"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "lnav-cli %s\n", build.Version)
			if build.Commit != "" {
				fmt.Fprintf(out, "  commit:   %s\n", build.Commit)
			}
			if build.Date != "" {
				fmt.Fprintf(out, "  built:    %s\n", build.Date)
			}
			fmt.Fprintf(out, "  go:       %s\n", runtime.Version())
			fmt.Fprintf(out, "  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			return nil
		},
	}
}
