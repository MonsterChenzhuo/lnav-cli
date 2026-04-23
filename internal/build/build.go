// Copyright (c) 2026 lnav-cli authors
// SPDX-License-Identifier: MIT

// Package build exposes version metadata injected at link time.
package build

import "runtime/debug"

// Version is set via -ldflags "-X ...internal/build.Version=..." at build time.
// Falls back to module version (for `go install`), then "DEV".
var Version = "DEV"

// Commit is the short git SHA, set via -ldflags.
var Commit = ""

// Date is the build date in YYYY-MM-DD, set via -ldflags.
var Date = ""

func init() {
	if Version == "DEV" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "(devel)" && info.Main.Version != "" {
			Version = info.Main.Version
		}
	}
	if Version == "" {
		Version = "DEV"
	}
}
