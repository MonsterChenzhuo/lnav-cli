# lnav-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue.svg)](https://go.dev/)
[![CI](https://github.com/MonsterChenzhuo/lnav-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/MonsterChenzhuo/lnav-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/MonsterChenzhuo/lnav-cli.svg)](https://github.com/MonsterChenzhuo/lnav-cli/releases)

[English](./README.md) | [中文](./README.zh.md)

An **agent-friendly wrapper** around [lnav](https://lnav.org) (the terminal log viewer). `lnav-cli` turns lnav's headless mode into a handful of structured, one-line commands designed for Claude Code and other AI agents — so triaging an incident becomes a chat turn, not a tmux session.

[Install](#installation) · [Quick Start](#quick-start) · [Commands](#commands) · [Agent Skills](#agent-skills) · [Advanced](#advanced-usage) · [Security](#security--limits) · [Contributing](#contributing)

## Why lnav-cli?

- **Agent-native output** — every command emits either lnav's JSON stream (ndjson) or a structured envelope `{data, _meta}` on stdout; errors go to stderr as `{code, message, hint}`. No ANSI, no pagers, no prompts.
- **Bounded by default** — `+tail` refuses to run without `--duration` or `--max-events`, so an agent can never hang a session on a live log stream.
- **Safe argv construction** — lnav arguments are assembled in pure Go (`internal/lnavexec`), with explicit rejection of newline-injected input. No shell, no template expansion.
- **Three shortcuts cover 90% of triage** — `+search` (regex + level + time window), `+sql` (schema-first SQLite over logs), `+summary` (level distribution + top-N errors + histogram), plus `+tail` for bounded follow.
- **Named sources** — `lnav-cli source add backend --paths ...` turns a messy path list into an alias you can hand to every downstream command.
- **Claude Code skills bundled** — five skills (`lnav-shared`, `lnav-search`, `lnav-sql`, `lnav-summary`, `lnav-tail`) ship in this repo so an agent knows exactly how to use the tool.

## Features

| Category        | Capabilities                                                                                    |
| --------------- | ----------------------------------------------------------------------------------------------- |
| Regex search    | `+search` with `--level`, `--since`, `--until`; returns ndjson events                            |
| SQL over logs   | `+sql` with `--show-schema` short-circuit; runs SQLite queries against lnav's log tables         |
| Triage summary  | `+summary` merges level distribution, top-N error bodies, and time histogram into one envelope   |
| Bounded follow  | `+tail` requires `--duration` or `--max-events`; context-cancels cleanly on deadline             |
| Named sources   | `source add/ls/show/rm`, persisted to `~/.lnav-cli/sources.yaml`                                 |
| Time windows    | Relative (`1h`, `30m`) and absolute (RFC3339, `2006-01-02T15:04:05`, date-only)                  |
| Health check    | `doctor` probes `lnav` on `PATH` and prints its version; `setup` prints platform install hints  |
| Dry-run mode    | `--dry-run` prints the computed lnav argv instead of exec'ing — ideal for agent verification     |

## Installation

### Requirements

- `lnav` ≥ 0.12 on `PATH` (the tool wraps lnav; it does not embed it)
- Go `1.25+` (only for building from source)

### Install lnav first

```bash
brew install lnav                      # macOS
sudo apt-get install -y lnav           # Debian / Ubuntu
sudo dnf install -y lnav               # Fedora
scoop install lnav                     # Windows (scoop)
```

Or see the full matrix at <https://lnav.org/downloads>.

### Install lnav-cli

**Option 1 — From a release tarball (recommended)**

```bash
# pick your platform; see https://github.com/MonsterChenzhuo/lnav-cli/releases
curl -L -o lnav-cli.tar.gz \
  https://github.com/MonsterChenzhuo/lnav-cli/releases/latest/download/lnav-cli-<ver>-<os>-<arch>.tar.gz
tar -xzf lnav-cli.tar.gz
sudo install -m755 lnav-cli-*/lnav-cli /usr/local/bin/lnav-cli
lnav-cli version
```

**Option 2 — From source**

```bash
git clone https://github.com/MonsterChenzhuo/lnav-cli.git
cd lnav-cli
make install                 # installs to /usr/local/bin by default
#  or: make install PREFIX=$HOME/.local
lnav-cli version
```

**Option 3 — `go install`**

```bash
go install github.com/MonsterChenzhuo/lnav-cli@latest
```

## Quick Start

### Quick Start (Human Users)

> **AI assistants:** skip to [Quick Start (AI Agent)](#quick-start-ai-agent).

```bash
# 1. Check environment
lnav-cli doctor

# 2. Register a log source (optional but recommended)
lnav-cli source add nginx \
  --paths "/var/log/nginx/access.log,/var/log/nginx/error.log" \
  --default-level warning

# 3. Search last hour for "timeout" at WARN or above
lnav-cli +search -s nginx --since 1h --level warning "timeout"

# 4. Triage: top errors + histogram for the last 6h
lnav-cli +summary -s nginx --since 6h --histogram 5m

# 5. SQL drill-down (schema first!)
lnav-cli +sql -s nginx --show-schema
lnav-cli +sql -s nginx \
  "SELECT log_level, count(*) c FROM access_log GROUP BY 1 ORDER BY c DESC"
```

### Quick Start (AI Agent)

> The following is what Claude Code (or any SKILL-aware agent) should execute. All output is JSON or ndjson on stdout; errors are `{code, message, hint}` on stderr.

**Step 1 — Verify environment**

```bash
lnav-cli doctor
```

**Step 2 — Dry-run the argv you plan to submit**

```bash
lnav-cli +search -s /var/log/app.log --since 30m --level error --dry-run "panic|fatal"
```

Inspect the printed lnav command. Only execute once the argv is what you expect.

**Step 3 — Execute**

```bash
lnav-cli +search -s /var/log/app.log --since 30m --level error "panic|fatal"
```

**Step 4 — When following a live stream, ALWAYS bound the tail**

```bash
# Either --duration or --max-events is required; the command refuses without one.
lnav-cli +tail -s /var/log/app.log --duration 30s --level error
```

**Step 5 — Install the bundled skills so the agent knows the command contracts**

```bash
cp -r skills/* ~/.claude/skills/
```

See [`skills/lnav-shared/SKILL.md`](./skills/lnav-shared/SKILL.md) for the shared contract that every other skill references.

## Commands

| Command                 | Purpose                                                                   |
| ----------------------- | ------------------------------------------------------------------------- |
| `lnav-cli doctor`       | Probe `lnav` on PATH, print its version and `PATH` entry                   |
| `lnav-cli setup`        | Print platform-specific install hints for lnav (no network, no sudo)       |
| `lnav-cli version`      | Print `lnav-cli` version, commit, build date, platform                     |
| `lnav-cli +search`      | Regex search with `--level` and `--since/--until`                          |
| `lnav-cli +sql`         | SQLite query over log files; `--show-schema` dumps the table layout        |
| `lnav-cli +summary`     | Triage envelope: level counts + top-N errors + histogram                   |
| `lnav-cli +tail`        | Bounded follow; requires `--duration` or `--max-events`                    |
| `lnav-cli source add`   | Register a named source (paths or command)                                 |
| `lnav-cli source ls`    | List registered sources                                                    |
| `lnav-cli source show`  | Print a source's raw definition                                            |
| `lnav-cli source rm`    | Remove a source                                                            |

### Global flags

| Flag             | Default   | Description                                                    |
| ---------------- | --------- | -------------------------------------------------------------- |
| `-s, --source`   | (none)    | Alias, file path, or glob — repeatable, required by shortcuts  |
| `--since`        | (none)    | Start of the time window (`1h`, `2026-04-22T10:00:00Z`, ...)   |
| `--until`        | (none)    | End of the time window                                          |
| `--format`       | `ndjson`  | `ndjson\|json\|table\|pretty` (structured envelopes use json)   |
| `--limit`        | `0`       | Max rows returned — parsed; strict enforcement tracked for v1.x |
| `--dry-run`      | `false`   | Print the generated lnav argv instead of executing             |

## Agent Skills

Five skills ship under [`skills/`](./skills/). Copy them into the Claude Code skills directory (`~/.claude/skills/`) or publish them via your normal skill distribution flow.

| Skill             | Description                                                                                  |
| ----------------- | -------------------------------------------------------------------------------------------- |
| `lnav-shared`     | Gateway doc every other skill references — output contract, error envelope, time windows     |
| `lnav-search`     | `+search` contract + regex cheatsheet + time-window patterns                                  |
| `lnav-sql`        | `+sql` contract with a **schema-first** rule (run `--show-schema` before authoring queries)   |
| `lnav-summary`    | `+summary` triage workflow (levels → top errors → histogram zoom)                             |
| `lnav-tail`       | `+tail` bounded-follow rules (`--duration` / `--max-events` required)                         |

## Advanced Usage

### Named sources

A source can be either a **paths** list (files/globs) or a **command** (shell command whose stdout feeds lnav). Stored in `~/.lnav-cli/sources.yaml` (override with `LNAV_CLI_CONFIG_DIR`).

```bash
lnav-cli source add api \
  --paths "/var/log/app/*.log" \
  --default-level warning

lnav-cli source show api
```

> `command:` sources are persisted today but the stdin plumbing into lnav is scheduled for v1.x. `+search/+sql/+summary/+tail` currently error with `unsupported_stdin_source` when they hit a command source.

### Time windows

Accepted formats (UTC is assumed when no zone is provided):

- Relative: `30s`, `5m`, `2h`
- RFC3339: `2026-04-22T10:00:00Z`, `2026-04-22T10:00:00+08:00`
- Space variant: `2026-04-22 10:00:00`
- Date only: `2026-04-22`

### Dry-run

Every shortcut honors `--dry-run`, printing the exact lnav argv. This is the recommended way for an agent to self-check before issuing a real run.

```bash
$ lnav-cli +sql -s nginx --show-schema --dry-run
lnav -n -q .schema /var/log/nginx/access.log /var/log/nginx/error.log
```

### Output envelopes

Commands that return structured JSON use:

```json
{
  "data": {...},
  "_meta": {"count": 3}
}
```

Errors:

```json
{
  "code": "unbounded_tail",
  "message": "+tail requires --duration or --max-events to prevent hanging the agent",
  "hint": "pass --duration 30s or --max-events 500"
}
```

## Security & Limits

- **No shell, no eval.** All lnav invocations are built as `[]string` argv via `internal/lnavexec`. Inputs containing newlines are rejected with a panic — agents must sanitize or split first.
- **No privilege escalation.** `lnav-cli` never calls `sudo`, never touches systemd, and never edits files outside `$LNAV_CLI_CONFIG_DIR || ~/.lnav-cli/`.
- **Path validation.** Source aliases resolve to actual files before exec; missing files surface as `resolve_source` errors.
- **Bounded follow is mandatory.** `+tail` requires `--duration` or `--max-events` and a context timeout is installed so the process cannot outlive its deadline.
- **Known gaps (tracked for v1.x):** strict `--limit` enforcement, `command:` source stdin plumbing, `--format table|pretty` for non-summary commands.

## Project Layout

```
lnav-cli/
├── cmd/                    cobra command tree (root, doctor, setup, version,
│                           +search, +sql, +summary, +tail, source)
├── internal/
│   ├── build/              version metadata (ldflags-injected)
│   ├── lnavexec/           pure argv construction + subprocess runner
│   ├── output/             structured Err + JSON/NDJSON envelope helpers
│   ├── source/             sources.yaml load/save/resolve
│   └── timerange/          relative + absolute timestamp parsing
├── skills/                 Claude Code skills (lnav-shared + 4 shortcuts)
├── scripts/                skills-check.sh frontmatter validator
├── tests/
│   ├── e2e/dryrun/         dry-run E2E per shortcut (no lnav required)
│   ├── e2e/live/           live E2E (skipped when lnav is missing)
│   └── fixtures/           nginx.log sample
├── docs/superpowers/       design spec + MVP/v1.0 implementation plans
└── .github/workflows/      ci.yml (PR gate) + release.yml (goreleaser)
```

## Contributing

1. Read [CONTRIBUTING.md](./CONTRIBUTING.md) for workflow, branch policy, and commit conventions.
2. Read [CLAUDE.md](./CLAUDE.md) if you are (or are driving) an AI agent — it's the authoritative pre-PR checklist and code-convention reference for this repo.
3. Run the full local gate before pushing:
   ```bash
   make test       # vet + unit + dry-run E2E + skills-check
   make lint       # golangci-lint v2.1.6
   ```
4. Commit with [Conventional Commits](https://www.conventionalcommits.org/) (`feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`).
5. PRs must pass the `ci` workflow; releases are cut automatically from `main` via `release.yml` + GoReleaser.

## License

[MIT](./LICENSE). lnav itself is licensed separately — see <https://github.com/tstack/lnav>.
