# Changelog

All notable changes to this project are documented here. The format follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and this project
adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `lnav-cli version` subcommand plus `--version` flag; build metadata
  (`Version`, `Commit`, `Date`) injected via `-ldflags` from the Makefile and
  GoReleaser.
- `.goreleaser.yml` producing darwin/linux/windows × amd64/arm64 archives with
  README, LICENSE, CHANGELOG and `skills/` bundled in each archive.
- GitHub Actions `ci.yml` PR gate: build · vet · gofmt · `go mod tidy` check ·
  unit + dry-run E2E with coverage summary · golangci-lint · skills-check, all
  rolled up into a single `results` gate suitable for branch protection.
- `.golangci.yml` (v2 schema) enabling the standard set of correctness linters
  (`errcheck`, `errorlint`, `govet`, `unused`, `gocritic`, …) plus
  `gofmt`/`goimports` formatters.
- `CONTRIBUTING.md`, `CLAUDE.md` (AI-agent pre-PR checklist + conventions),
  `.licenserc.yaml` (skywalking-eyes SPDX MIT header config).
- Makefile targets: `install`, `uninstall`, `clean`, `coverage`, `help`,
  `release-snapshot`.

### Changed
- Release workflow now delegates packaging to GoReleaser (was a hand-rolled
  build matrix + `softprops/action-gh-release`).
- README / README.zh rewritten around a human vs AI-agent quick-start split,
  commands table, advanced section (named sources, time windows, dry-run,
  envelope format), security notes, and project layout — mirrored structure
  from the sibling `lark-cli` repo.
- `LICENSE` corrected from Apache-2.0 text to MIT so it matches the SPDX
  identifier used throughout the source tree.

## [v1.0.x] — pre-changelog baseline

These changes landed on `main` before this CHANGELOG existed. Refer to
`git log` for the individual commits.

### Added
- cobra root with global flags (`--source/--since/--until/--format/--limit/--dry-run`).
- `doctor`: detects `lnav` on PATH and prints its version.
- `setup`: prints platform-specific lnav install guidance.
- `+search`: regex + level search with optional time window, backed by
  `internal/lnavexec` argv construction.
- `+sql`: SQLite query over log files with `--show-schema` short-circuit.
- `+summary`: level distribution + top-N errors + histogram merged into a
  single `{data, _meta}` envelope.
- `+tail`: bounded follow — requires `--duration` or `--max-events`; rejects
  `unbounded_tail` otherwise.
- `source add/ls/show/rm` persisted to `~/.lnav-cli/sources.yaml`
  (`LNAV_CLI_CONFIG_DIR` honored by tests).
- Time-window parser (`1h`, RFC3339, `2006-01-02T15:04:05`, space variant, date
  only).
- Structured output envelope + `output.Err{Code,Message,Hint,ConsoleURL}`.
- Dry-run E2E tests for every shortcut; live E2E for `+search` against a
  bundled nginx fixture.
- Claude Code skills: `lnav-shared`, `lnav-search`, `lnav-sql`, `lnav-summary`,
  `lnav-tail` with references; `scripts/skills-check.sh` validates their
  frontmatter.
- Initial `.github/workflows/release.yml` auto-tagging every push to `main`.

[Unreleased]: https://github.com/MonsterChenzhuo/lnav-cli/compare/HEAD...HEAD
