# AGENTS.md

## Pre-PR Checklist

1. `make unit-test`
2. `go vet ./...`
3. `gofmt -l .` — must produce no output
4. `go mod tidy` — must not change `go.mod`/`go.sum`
5. `make skills-check`
6. `make e2e-dryrun`

## Conventions

- `stdout = data`, `stderr = logs/errors`.
- All errors returned from `RunE` must be `*output.Err`, never bare `fmt.Errorf`.
- All filesystem paths received from CLI flags must be validated before use.
- Every shortcut requires at least one dry-run E2E test asserting the generated `lnav` argv.

## Commit Style

Conventional Commits (English): `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`, `ci:`.
