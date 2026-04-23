# CLAUDE.md

Pre-PR checklist and code conventions for AI agents (Claude Code and
equivalents) working in this repository. If you are a human contributor you
probably want [CONTRIBUTING.md](./CONTRIBUTING.md) instead, which covers the
same material but is less mechanical.

## Goal

One intent per PR. If you find yourself halfway through a feature and want to
"just also clean up this other thing", **stop** and open a second PR for the
cleanup. Reviewers here optimize for small, single-intent diffs.

## What this tool is

`lnav-cli` is a thin, agent-friendly wrapper around [lnav](https://lnav.org).
The design premise — enforced by tests and by the output contract — is that
every invocation is a one-shot, bounded, pure-argv subprocess call that emits
structured JSON. Anything that weakens those properties (shell-outs, unbounded
tails, `fmt.Printf` bypassing the envelope) is a regression, not a feature.

## Build & test

Run these locally; they are also what `ci.yml` runs on every PR.

```bash
make build              # builds ./lnav-cli with version metadata
make unit-test          # -race -count=1 ./internal/... ./cmd/...
make e2e-dryrun         # tests/e2e/dryrun/ — no lnav required
make e2e-live           # tests/e2e/live/ — lnav must be on PATH
make lint               # golangci-lint v2.1.6 (same version as CI)
make skills-check       # validates skills/*/SKILL.md frontmatter
make test               # vet + unit + dry-run + skills-check
```

## Pre-PR checklist

Run these in order. **Do not open a PR until all six are green.**

1. `make unit-test` — every package test passes.
2. `go vet ./...` — no complaints.
3. `gofmt -l .` — must print nothing.
4. `go mod tidy` — must not change `go.mod` / `go.sum`.
5. `make skills-check` — skill frontmatter still valid.
6. `make e2e-dryrun` — dry-run argv regressions caught.

If you touched anything that shells out to `lnav`, also run `make e2e-live`
(it skips automatically when `lnav` is not on PATH, so the green-tick-on-CI
story does **not** cover it).

## Source layout

| Path                          | Purpose                                                                      |
| ----------------------------- | ---------------------------------------------------------------------------- |
| `main.go`                     | One-liner: `os.Exit(cmd.Execute())`. Do not grow it.                         |
| `cmd/`                        | cobra command tree. One file per command (`+search.go`, `+sql.go`, …).        |
| `cmd/root.go`                 | `GlobalOpts`, persistent flags, `Execute()`, subcommand registration.        |
| `cmd/version.go`              | `lnav-cli version` — reads from `internal/build`.                             |
| `internal/build/`             | Version/Commit/Date, injected via `-ldflags` at build time.                  |
| `internal/lnavexec/`          | **The only place** that constructs lnav argv or execs a subprocess.          |
| `internal/lnavexec/builder.go`| Pure `[]string` builders. Panic on newline-injected input.                   |
| `internal/lnavexec/runner.go` | `Runner.Run` / `RunCtx`. `--dry-run` short-circuit. Deadline → nil error.    |
| `internal/output/`            | `Err{Code,Message,Hint}`, `WriteJSON`, `WriteNDJSON`. The envelope contract. |
| `internal/source/`            | `sources.yaml` load/save/resolve. Honors `LNAV_CLI_CONFIG_DIR`.               |
| `internal/timerange/`         | `Parse` — relative + RFC3339 + `2006-01-02T15:04:05` + space + date-only.    |
| `skills/`                     | Claude Code skills (`lnav-shared` + 4 shortcuts).                            |
| `scripts/skills-check.sh`     | Skill frontmatter validator.                                                 |
| `tests/e2e/dryrun/`           | Per-shortcut argv-shape tests. **Add one for every new shortcut.**           |
| `tests/e2e/live/`             | Live E2E; skipped when `lnav` is missing.                                    |
| `docs/superpowers/`           | Design spec + MVP/v1.0 implementation plans. Append here, don't rewrite.     |

## Code conventions

- **Output discipline.** `stdout = data`, `stderr = logs/errors`. Never
  `fmt.Print*` straight to `os.Stdout` — use `cmd.OutOrStdout()` and
  `cmd.ErrOrStderr()` so tests can capture it.
- **Errors are structured.** Every `RunE` returns an `*output.Err`:
  ```go
  return output.Errorf("missing_source", "at least one --source / -s is required").
      WithHint("pass a path, glob, or registered alias")
  ```
  Codes are snake_case and stable — agents grep them.
- **Never shell out.** All lnav invocations go through `lnavexec.Runner` with
  a pre-built `[]string` argv. No `sh -c`. No `fmt.Sprintf` into a command
  string.
- **Reject newline injection.** The builders in `internal/lnavexec/builder.go`
  panic when an input contains `\n` or `\r`. Keep this behavior; it's load-bearing.
- **Bounded follow.** Any command that tails a source must refuse to run
  without `--duration` or `--max-events`. See `cmd/tail.go` for the exact
  rejection path (`output.Errorf("unbounded_tail", …)`).
- **Validate paths before exec.** Source aliases must resolve through
  `source.Config.Resolve` — don't pass raw user strings to lnav.
- **Dry-run parity.** Every shortcut must honor `--dry-run` and print the argv
  it *would* have executed. A shortcut that executes work even in dry-run mode
  is a bug.
- **Table-driven tests.** Builder and timerange tests already use this style;
  follow it when extending.

## Adding a new shortcut (checklist)

1. Add the argv builder under `internal/lnavexec/builder.go`, plus a unit
   test in `builder_test.go` covering the happy path and the newline-rejection
   panic.
2. Add `cmd/<name>.go`. Validate inputs, resolve sources, return `*output.Err`,
   honor `--dry-run`.
3. Register in `cmd/root.go`.
4. Add `cmd/<name>_test.go` — at minimum a dry-run assertion and the missing-
   source rejection.
5. Add `tests/e2e/dryrun/<name>_test.go` pinning the generated argv.
6. Add / update `skills/lnav-<name>/SKILL.md` with valid frontmatter (name,
   version, description, `metadata.requires.bins`, `metadata.cliHelp`) and at
   least one `references/*.md` explaining the contract.
7. `make skills-check` must pass.
8. Update `README.md` command table + global flag table if applicable, and add
   a bullet to `CHANGELOG.md` under `[Unreleased]`.

## Commit style

[Conventional Commits](https://www.conventionalcommits.org/), scoped by
package or command:

```
feat(cmd): add +grep shortcut
fix(lnavexec): reject carriage-return in patterns
test(e2e): pin +summary argv shape
docs(skill): tighten lnav-sql schema-first rule
ci: run coverage on PRs
```

Types in use: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`,
`build`, `perf`.

## Who uses this CLI

An **AI agent** is the primary consumer. That drives several conventions the
code would otherwise not need:

- Stable error codes (agents branch on them).
- Hints on errors (agents self-correct from the hint).
- Dry-run on every shortcut (agents verify argv before executing).
- Bounded tails (agents otherwise hang their session).
- JSON/NDJSON only (agents can't parse tables reliably).

When you're about to add a helpful prompt, a colored warning, an interactive
confirm, or a "press any key to continue" — **stop**. Emit a structured error
with a hint instead.

## Common traps

- **Adding `fmt.Println` inside a shortcut.** It bypasses the envelope and
  breaks JSON parsing for the caller. Use `cmd.OutOrStdout()` and encode JSON.
- **Forgetting the dry-run E2E.** Unit tests catch argv composition bugs only
  for the cases you remembered; the dry-run E2E catches argv *shape* drift for
  the whole pipeline.
- **Returning bare `fmt.Errorf`.** `cobra` will print it, but the agent will
  not get a `code` to branch on. Always wrap with `output.Errorf`.
- **Reading `os.Stdin/Stdout/Stderr` directly in `cmd/`.** Use the cobra
  accessors; tests depend on them.
- **Committing the built binary.** `.gitignore` lists `/lnav-cli`; confirm it
  stays ignored. Don't `git add -f` it.

## When in doubt

Grep a sibling command. The shortcut skeleton is deliberately repetitive so
adding a new one is mostly pattern-matching. If the pattern doesn't fit, open
a short design note under `docs/superpowers/specs/` before touching code.
