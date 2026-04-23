# Contributing to lnav-cli

Thanks for helping out. This doc is the short version aimed at humans; AI
agents should read [CLAUDE.md](./CLAUDE.md) instead (same ideas, more
prescriptive).

## Ground rules

- **One intent per PR.** Keep refactors, features, and fixes separate — the
  reviewer's job shouldn't be untangling them.
- **Output discipline.** `stdout` is data (JSON/NDJSON); `stderr` is logs and
  the error envelope. Don't mix.
- **Errors are structured.** Everything returned from a `cobra.Command.RunE`
  must be an `*output.Err`, never a bare `fmt.Errorf`. Give it a stable
  machine-readable `Code`, a human `Message`, and — whenever possible — a
  `Hint` that tells the agent how to fix its input.
- **Argv construction is pure.** New lnav integrations go through
  `internal/lnavexec`. No `os/exec.Command("sh", "-c", ...)`. No string
  interpolation into shell.
- **Every shortcut needs a dry-run E2E test.** Before shipping a new `+foo`
  shortcut, add a dry-run assertion under `tests/e2e/dryrun/` that pins the
  expected lnav argv. This is how we prevent silent argv regressions.
- **Bounded by default.** If you add a follow-style command, it must refuse to
  run without a time/event budget — same contract as `+tail`.

## Local workflow

```bash
git checkout -b feat/my-shortcut
# hack, hack, hack
make test        # vet + unit + dry-run E2E + skills-check
make lint        # golangci-lint v2.1.6
gofmt -l .       # must produce no output
go mod tidy      # must not change go.mod / go.sum
```

If you touched anything that shells out to `lnav`, also run:

```bash
make e2e-live    # requires lnav on PATH
```

## Commit style

[Conventional Commits](https://www.conventionalcommits.org/), scoped by the
package or command you touched:

```
feat(cmd): add +grep shortcut
fix(lnavexec): reject carriage-return in patterns
test(e2e): pin +summary argv shape
docs(skill): tighten lnav-sql schema-first rule
```

Types in use: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`, `ci`,
`build`, `perf`.

## Adding a shortcut

1. Add the argv builder under `internal/lnavexec/builder.go` (panic on
   newline-injected input — existing builders show the pattern).
2. Add the cobra command in `cmd/<name>.go`. Validate inputs, resolve sources,
   return `*output.Err`, honor `--dry-run`.
3. Register it in `cmd/root.go`.
4. Add a unit test covering the happy path and the rejection paths (missing
   source, bad time, unbounded follow, etc.).
5. Add a dry-run E2E under `tests/e2e/dryrun/` asserting the generated argv.
6. Add or update the matching skill under `skills/`. At minimum: `name`,
   `version`, `description`, `metadata.requires.bins` (must include `lnav`),
   and a `references/` file explaining the contract.
7. Run `make skills-check` to validate the frontmatter.
8. Update `README.md` command table and `CHANGELOG.md` under `[Unreleased]`.

## Adding a skill

Skills live under `skills/<name>/SKILL.md` with optional `references/*.md`.
Frontmatter is validated by `scripts/skills-check.sh`:

```yaml
---
name: lnav-foo
version: 0.1.0
description: what this skill does in one sentence
metadata:
  requires:
    bins: [lnav, lnav-cli]
  cliHelp: lnav-cli +foo --help
---
```

Every `references/*.md` referenced in `SKILL.md` must exist.

## Releasing

Releases are automated. Every push to `main`:

1. `ci.yml` runs the full PR gate.
2. `release.yml` tags `v1.0.${GITHUB_RUN_NUMBER}` and invokes GoReleaser.
3. GoReleaser produces darwin/linux/windows × amd64/arm64 archives plus a
   `SHA256SUMS` checksum file, publishes them to GitHub Releases, and derives
   release notes from conventional commits since the last tag.

To cut a tag manually (e.g. a backport), push an annotated `v*` tag and the
same release job fires.

## Questions

Open a GitHub issue. For design-level discussion, put a short proposal under
`docs/superpowers/specs/` first — that's where the MVP and v1.0 specs live.
