# AGENTS.md

This file exists for AI coding tools that look for `AGENTS.md` by convention
(OpenAI Codex, Gemini CLI, etc.). The authoritative pre-PR checklist,
conventions, and source-layout reference for this repo live in
[CLAUDE.md](./CLAUDE.md) — read that first.

## TL;DR for agents in a hurry

1. Run `make test && make lint` before opening a PR.
2. `stdout = data`, `stderr = errors`. Errors must be `*output.Err` with a
   stable snake_case `code` and a `hint`.
3. All lnav argv goes through `internal/lnavexec`. No shell. Newline-injected
   input panics — that's intentional.
4. Every new shortcut needs a matching dry-run E2E under `tests/e2e/dryrun/`
   pinning the expected lnav argv.
5. Follow-style commands must refuse to run without `--duration` or
   `--max-events` (see `cmd/tail.go`).
6. Conventional Commits: `feat(cmd): …`, `fix(lnavexec): …`, `docs(skill): …`.

For the full version — source layout table, code conventions, shortcut
checklist, common traps — open [CLAUDE.md](./CLAUDE.md).
