// Copyright (c) 2026 lnav-cli authors
// SPDX-License-Identifier: MIT
//
// lnav-cli — lnav wrapper for AI agents.
package main

import (
	"os"

	"github.com/MonsterChenzhuo/lnav-cli/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
