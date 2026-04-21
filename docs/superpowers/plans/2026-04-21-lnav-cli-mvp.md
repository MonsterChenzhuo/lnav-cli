# lnav-cli MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付 lnav-cli MVP — 打通「Claude Code → `lnav-cli +search` → lnav 子进程 → JSON envelope」完整链路，并附带 `lnav-shared` + `lnav-search` 两个 Skill，使 Claude Code 可通过一行命令完成日志搜索。

**Architecture:** Go + cobra 命令行；`cmd/*` 暴露子命令；`internal/*` 按职责拆分（`lnavexec` 构造 argv 并调用 lnav、`timerange` 解析时间、`output` 做 JSON 封包与结构化错误）；skills 目录以 cli/ 同构的 SKILL.md + references 形式提供给 Claude Code。

**Tech Stack:** Go 1.23、`github.com/spf13/cobra`、`github.com/stretchr/testify`、`github.com/tidwall/gjson`（解析 lnav JSON 输出）、外部 `lnav` 可执行文件。

---

## File Structure

MVP 创建或修改的文件：

- `go.mod`, `go.sum` — 模块声明与依赖
- `main.go` — 入口，调用 `cmd.Execute()`
- `.gitignore` — 忽略构建产物
- `Makefile` — `build/unit-test/lint/e2e-dryrun/e2e-live/skills-check`
- `AGENTS.md` — 精简自 cli/AGENTS.md 的贡献守则
- `README.md`, `README.zh.md` — 快速开始
- `cmd/root.go` — cobra root、全局 flags、Execute 入口
- `cmd/doctor.go` — `lnav-cli doctor`
- `cmd/search.go` — `lnav-cli +search` shortcut
- `internal/lnavexec/builder.go` — 构造 lnav argv（无状态）
- `internal/lnavexec/builder_test.go`
- `internal/lnavexec/runner.go` — 真正 exec + `--dry-run` 支持
- `internal/lnavexec/runner_test.go`
- `internal/output/envelope.go` — `{data,_meta}` JSON / NDJSON / structured error envelope
- `internal/output/envelope_test.go`
- `internal/timerange/parse.go` — `--since/--until`（相对时长 & 绝对 RFC3339）
- `internal/timerange/parse_test.go`
- `skills/lnav-shared/SKILL.md`
- `skills/lnav-search/SKILL.md`
- `skills/lnav-search/references/regex-cheatsheet.md`
- `skills/lnav-search/references/time-window.md`
- `scripts/skills-check.sh` — frontmatter 与 references 存在性校验
- `tests/fixtures/nginx.log` — Live E2E fixture
- `tests/e2e/dryrun/search_dryrun_test.go`
- `tests/e2e/live/search_live_test.go`（当 `lnav` 不在 PATH 时 `t.Skip`）

---

## Task 1 — Go 项目骨架

**Files:**
- Create: `go.mod`, `main.go`, `.gitignore`, `Makefile`, `AGENTS.md`

- [ ] **Step 1: 初始化 go mod**

```bash
cd /Users/opay-20240095/IdeaProjects/createcli/lnav-cli
go mod init github.com/MonsterChenzhuo/lnav-cli
```

- [ ] **Step 2: 写 `main.go`**

```go
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
```

- [ ] **Step 3: 写 `.gitignore`**

```
/lnav-cli
/dist/
*.test
*.out
.idea/
.vscode/
```

- [ ] **Step 4: 写 `Makefile`**

```makefile
.PHONY: build unit-test lint e2e-dryrun e2e-live skills-check test fmt vet

GO ?= go
BINARY ?= lnav-cli

build:
	$(GO) build -o $(BINARY) .

unit-test:
	$(GO) test -race -count=1 ./internal/... ./cmd/...

e2e-dryrun:
	$(GO) test -race -count=1 ./tests/e2e/dryrun/...

e2e-live:
	$(GO) test -count=1 -tags=live ./tests/e2e/live/...

skills-check:
	bash scripts/skills-check.sh

fmt:
	gofmt -l -w .

vet:
	$(GO) vet ./...

lint:
	$(GO) run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6 run

test: unit-test e2e-dryrun skills-check
```

- [ ] **Step 5: 写 `AGENTS.md`**

```markdown
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
```

- [ ] **Step 6: 验证**

```bash
go vet ./... ; echo "vet exit=$?"
gofmt -l .
```
Expected: `vet exit=0`，`gofmt` 无输出（此时仓库只有 `main.go`，会因 `cmd` 包缺失而 vet 失败；**允许失败**，下一任务补齐）。

- [ ] **Step 7: Commit + push**

```bash
git add go.mod main.go .gitignore Makefile AGENTS.md
git commit -m "chore: scaffold go module, makefile, agents doc"
git push origin main
```

---

## Task 2 — cobra root 与全局 flags

**Files:**
- Create: `cmd/root.go`, `cmd/root_test.go`

- [ ] **Step 1: 写 `cmd/root_test.go`（先写测试）**

```go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootCommand_HelpIncludesShortcuts(t *testing.T) {
	buf := &bytes.Buffer{}
	root := NewRootCmd()
	root.SetOut(buf)
	root.SetArgs([]string{"--help"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"+search", "doctor", "--format", "--dry-run"} {
		if !strings.Contains(out, want) {
			t.Errorf("root --help missing %q; got:\n%s", want, out)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./cmd -run TestRootCommand_HelpIncludesShortcuts
```
Expected: FAIL — `NewRootCmd` 未定义。

- [ ] **Step 3: 写 `cmd/root.go`**

```go
// Package cmd hosts the cobra command tree for lnav-cli.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type GlobalOpts struct {
	Sources []string
	Since   string
	Until   string
	Format  string
	Limit   int
	DryRun  bool
}

var globalOpts GlobalOpts

func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "lnav-cli",
		Short:         "Agent-friendly wrapper around lnav",
		Long:          "lnav-cli — drive lnav from Claude Code with one-line commands.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.PersistentFlags().StringSliceVarP(&globalOpts.Sources, "source", "s", nil, "log source alias, file, or glob (repeatable)")
	root.PersistentFlags().StringVar(&globalOpts.Since, "since", "", "start of time window (e.g. 1h, 2026-04-21T10:00:00Z)")
	root.PersistentFlags().StringVar(&globalOpts.Until, "until", "", "end of time window")
	root.PersistentFlags().StringVar(&globalOpts.Format, "format", "ndjson", "output format: ndjson|json|table|pretty")
	root.PersistentFlags().IntVar(&globalOpts.Limit, "limit", 0, "max rows returned (0 = unbounded)")
	root.PersistentFlags().BoolVar(&globalOpts.DryRun, "dry-run", false, "print the lnav argv instead of executing")

	root.AddCommand(newDoctorCmd())
	root.AddCommand(newSearchCmd())
	return root
}

// Execute runs the root command and returns a process exit code.
func Execute() int {
	root := NewRootCmd()
	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return 1
	}
	return 0
}

// placeholder stubs for command constructors defined in sibling files.
func newDoctorCmd() *cobra.Command { return &cobra.Command{Use: "doctor", RunE: func(*cobra.Command, []string) error { return nil }} }
func newSearchCmd() *cobra.Command {
	return &cobra.Command{Use: "+search", Short: "regex/level search", RunE: func(*cobra.Command, []string) error { return nil }}
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go mod tidy
go test ./cmd -run TestRootCommand_HelpIncludesShortcuts -v
```
Expected: PASS.

- [ ] **Step 5: Commit + push**

```bash
git add cmd/root.go cmd/root_test.go go.mod go.sum
git commit -m "feat(cmd): add cobra root with global flags"
git push origin main
```

---

## Task 3 — `internal/timerange` 时间解析

**Files:**
- Create: `internal/timerange/parse.go`, `internal/timerange/parse_test.go`

- [ ] **Step 1: 写 `internal/timerange/parse_test.go`**

```go
package timerange

import (
	"testing"
	"time"
)

func TestParse_Relative(t *testing.T) {
	now := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	got, err := parseAt("1h", now)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	want := now.Add(-time.Hour)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestParse_RFC3339(t *testing.T) {
	got, err := parseAt("2026-04-20T10:30:00Z", time.Now())
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if got.UTC().Format(time.RFC3339) != "2026-04-20T10:30:00Z" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestParse_Empty(t *testing.T) {
	if _, err := parseAt("", time.Now()); err == nil {
		t.Fatal("expected error on empty input")
	}
}

func TestParse_Invalid(t *testing.T) {
	if _, err := parseAt("tomorrow", time.Now()); err == nil {
		t.Fatal("expected error on unsupported input")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/timerange -run TestParse
```
Expected: FAIL (`parseAt` undefined).

- [ ] **Step 3: 写 `internal/timerange/parse.go`**

```go
// Package timerange parses --since/--until values into concrete instants.
package timerange

import (
	"fmt"
	"time"
)

// Parse converts a user-facing time expression into an absolute time, using time.Now as the anchor.
func Parse(s string) (time.Time, error) {
	return parseAt(s, time.Now())
}

func parseAt(s string, now time.Time) (time.Time, error) {
	if s == "" {
		return time.Time{}, fmt.Errorf("empty time expression")
	}
	if d, err := time.ParseDuration(s); err == nil {
		return now.Add(-d), nil
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time expression %q (use e.g. 1h, 2026-04-21T10:00:00Z)", s)
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go test ./internal/timerange -v
```
Expected: PASS (4 tests).

- [ ] **Step 5: Commit + push**

```bash
git add internal/timerange
git commit -m "feat(timerange): parse --since/--until values"
git push origin main
```

---

## Task 4 — `internal/output` envelope & 结构化错误

**Files:**
- Create: `internal/output/envelope.go`, `internal/output/envelope_test.go`

- [ ] **Step 1: 写 `internal/output/envelope_test.go`**

```go
package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestWriteJSON_Envelope(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteJSON(buf, []map[string]any{{"ts": "2026-04-21T00:00:00Z", "body": "hello"}}, Meta{Source: "nginx-prod"})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	var got struct {
		Data []map[string]any `json:"data"`
		Meta map[string]any   `json:"_meta"`
	}
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal: %v\nraw=%s", err, buf.String())
	}
	if len(got.Data) != 1 || got.Data[0]["body"] != "hello" {
		t.Fatalf("data wrong: %+v", got.Data)
	}
	if got.Meta["source"] != "nginx-prod" {
		t.Fatalf("meta wrong: %+v", got.Meta)
	}
}

func TestWriteNDJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	err := WriteNDJSON(buf, []map[string]any{{"a": 1}, {"a": 2}})
	if err != nil {
		t.Fatalf("write: %v", err)
	}
	lines := strings.Split(strings.TrimRight(buf.String(), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %q", len(lines), buf.String())
	}
}

func TestErr_JSONShape(t *testing.T) {
	e := Errorf("lnav_not_found", "lnav executable not found on PATH").WithHint("run: lnav-cli setup")
	raw, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(raw), "\"code\":\"lnav_not_found\"") ||
		!strings.Contains(string(raw), "\"hint\":\"run: lnav-cli setup\"") {
		t.Fatalf("shape wrong: %s", raw)
	}
	if e.Error() == "" {
		t.Fatal("Error() empty")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/output
```
Expected: FAIL (types not defined).

- [ ] **Step 3: 写 `internal/output/envelope.go`**

```go
// Package output formats lnav-cli stdout data and stderr errors.
package output

import (
	"encoding/json"
	"fmt"
	"io"
)

// Meta is attached to every JSON response as "_meta".
type Meta struct {
	Source string `json:"source,omitempty"`
	Since  string `json:"since,omitempty"`
	Until  string `json:"until,omitempty"`
	Count  int    `json:"count"`
}

type envelope struct {
	Data any  `json:"data"`
	Meta Meta `json:"_meta"`
}

// WriteJSON writes a single JSON envelope containing all rows and metadata.
func WriteJSON(w io.Writer, rows []map[string]any, meta Meta) error {
	meta.Count = len(rows)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(envelope{Data: rows, Meta: meta})
}

// WriteNDJSON writes one JSON object per line (no envelope) for streaming consumers.
func WriteNDJSON(w io.Writer, rows []map[string]any) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, r := range rows {
		if err := enc.Encode(r); err != nil {
			return err
		}
	}
	return nil
}

// Err is the structured error envelope written to stderr.
type Err struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Hint       string `json:"hint,omitempty"`
	ConsoleURL string `json:"console_url,omitempty"`
}

func (e *Err) Error() string { return fmt.Sprintf("%s: %s", e.Code, e.Message) }

// Errorf builds a new structured error.
func Errorf(code, format string, a ...any) *Err {
	return &Err{Code: code, Message: fmt.Sprintf(format, a...)}
}

// WithHint attaches a human-actionable hint to the error.
func (e *Err) WithHint(hint string) *Err { e.Hint = hint; return e }

// WriteErr emits the error as JSON to stderr.
func WriteErr(w io.Writer, e *Err) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc.Encode(e)
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go test ./internal/output -v
```
Expected: PASS (3 tests).

- [ ] **Step 5: Commit + push**

```bash
git add internal/output
git commit -m "feat(output): add JSON envelope and structured error"
git push origin main
```

---

## Task 5 — `internal/lnavexec` argv builder（纯函数）

**Files:**
- Create: `internal/lnavexec/builder.go`, `internal/lnavexec/builder_test.go`

- [ ] **Step 1: 写 `internal/lnavexec/builder_test.go`**

```go
package lnavexec

import (
	"reflect"
	"testing"
)

func TestBuildSearchArgs_Minimal(t *testing.T) {
	got := BuildSearchArgs(SearchOpts{Files: []string{"/tmp/a.log"}})
	want := []string{"-n", "-c", ":write-json-to -", "--", "/tmp/a.log"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v want=%v", got, want)
	}
}

func TestBuildSearchArgs_Full(t *testing.T) {
	got := BuildSearchArgs(SearchOpts{
		Pattern:  "timeout|refused",
		Level:    "error",
		SinceTS:  "2026-04-21T10:00:00Z",
		Files:    []string{"/var/log/nginx/error.log"},
	})
	want := []string{
		"-n",
		"-c", ":filter-in timeout|refused",
		"-c", ":set-min-log-level error",
		"-c", ":hide-unmarked-lines-before 2026-04-21T10:00:00Z",
		"-c", ":write-json-to -",
		"--",
		"/var/log/nginx/error.log",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got=%v\nwant=%v", got, want)
	}
}

func TestBuildSearchArgs_RejectsShellInjection(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on newline in pattern")
		}
	}()
	_ = BuildSearchArgs(SearchOpts{Pattern: "bad\npattern", Files: []string{"/tmp/a.log"}})
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/lnavexec
```
Expected: FAIL.

- [ ] **Step 3: 写 `internal/lnavexec/builder.go`**

```go
// Package lnavexec builds and executes lnav subprocess invocations.
package lnavexec

import "strings"

type SearchOpts struct {
	Pattern string   // regex passed to :filter-in; empty means no filter
	Level   string   // one of info|warning|error|fatal; empty means any
	SinceTS string   // RFC3339 / lnav-compatible absolute timestamp
	UntilTS string
	Files   []string // positional args passed to lnav
}

// BuildSearchArgs produces the argv (minus the leading "lnav" binary) for a +search call.
// Inputs must not contain newlines; the function panics on violation because callers
// are expected to have run validation up-stream.
func BuildSearchArgs(o SearchOpts) []string {
	assertNoNewline(o.Pattern, "pattern")
	assertNoNewline(o.Level, "level")
	assertNoNewline(o.SinceTS, "since")
	assertNoNewline(o.UntilTS, "until")

	args := []string{"-n"}
	if o.Pattern != "" {
		args = append(args, "-c", ":filter-in "+o.Pattern)
	}
	if o.Level != "" {
		args = append(args, "-c", ":set-min-log-level "+o.Level)
	}
	if o.SinceTS != "" {
		args = append(args, "-c", ":hide-unmarked-lines-before "+o.SinceTS)
	}
	if o.UntilTS != "" {
		args = append(args, "-c", ":hide-unmarked-lines-after "+o.UntilTS)
	}
	args = append(args, "-c", ":write-json-to -", "--")
	args = append(args, o.Files...)
	return args
}

func assertNoNewline(s, field string) {
	if strings.ContainsAny(s, "\r\n") {
		panic("lnavexec: newline in " + field + " — input validation missing up-stream")
	}
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go test ./internal/lnavexec -v
```
Expected: PASS (3 tests).

- [ ] **Step 5: Commit + push**

```bash
git add internal/lnavexec
git commit -m "feat(lnavexec): build +search argv"
git push origin main
```

---

## Task 6 — `internal/lnavexec` runner（exec + dry-run）

**Files:**
- Create: `internal/lnavexec/runner.go`, `internal/lnavexec/runner_test.go`

- [ ] **Step 1: 写 `internal/lnavexec/runner_test.go`**

```go
package lnavexec

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunner_DryRun_PrintsArgv(t *testing.T) {
	out := &bytes.Buffer{}
	r := Runner{Binary: "lnav", DryRun: true, Stdout: out}
	args := []string{"-n", "-c", ":filter-in foo", "--", "/tmp/x.log"}
	if err := r.Run(args); err != nil {
		t.Fatalf("run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "lnav -n -c ':filter-in foo' -- /tmp/x.log") {
		t.Fatalf("unexpected dry-run output: %q", got)
	}
}

func TestRunner_MissingBinary_ReturnsNotFoundErr(t *testing.T) {
	r := Runner{Binary: "definitely-not-a-real-binary-xyz", DryRun: false}
	err := r.Run([]string{"-n"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "lnav_not_found") {
		t.Fatalf("wrong error code: %v", err)
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./internal/lnavexec -run TestRunner
```
Expected: FAIL.

- [ ] **Step 3: 写 `internal/lnavexec/runner.go`**

```go
package lnavexec

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
)

// Runner executes lnav as a subprocess; DryRun true prints argv instead of running.
type Runner struct {
	Binary string
	DryRun bool
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// Run executes `lnav args...`.
func (r Runner) Run(args []string) error {
	if r.Binary == "" {
		r.Binary = "lnav"
	}
	if r.Stdout == nil {
		r.Stdout = os.Stdout
	}
	if r.Stderr == nil {
		r.Stderr = os.Stderr
	}
	if r.DryRun {
		fmt.Fprintln(r.Stdout, prettyArgv(r.Binary, args))
		return nil
	}
	if _, err := exec.LookPath(r.Binary); err != nil {
		return output.Errorf("lnav_not_found", "lnav executable %q not found on PATH", r.Binary).
			WithHint("run: lnav-cli setup, or install manually (brew install lnav)")
	}
	cmd := exec.Command(r.Binary, args...)
	cmd.Stdin = r.Stdin
	cmd.Stdout = r.Stdout
	cmd.Stderr = r.Stderr
	if err := cmd.Run(); err != nil {
		return output.Errorf("lnav_exec_failed", "%v", err)
	}
	return nil
}

func prettyArgv(bin string, args []string) string {
	parts := []string{bin}
	for _, a := range args {
		if strings.ContainsAny(a, " \t\"'") {
			parts = append(parts, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
			continue
		}
		parts = append(parts, a)
	}
	return strings.Join(parts, " ")
}
```

- [ ] **Step 4: 运行测试，确认通过**

```bash
go test ./internal/lnavexec -v
```
Expected: PASS.

- [ ] **Step 5: Commit + push**

```bash
git add internal/lnavexec/runner.go internal/lnavexec/runner_test.go
git commit -m "feat(lnavexec): add runner with dry-run support"
git push origin main
```

---

## Task 7 — `cmd/doctor` 命令

**Files:**
- Create: `cmd/doctor.go`, `cmd/doctor_test.go`
- Modify: `cmd/root.go`（删除 doctor 占位符，改为引用真实构造器）

- [ ] **Step 1: 写 `cmd/doctor_test.go`**

```go
package cmd

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestDoctor_Reports(t *testing.T) {
	out := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd := newDoctorCmd()
	cmd.SetOut(out)
	cmd.SetErr(stderr)
	_ = cmd.Execute()
	report := out.String()
	if _, err := exec.LookPath("lnav"); err == nil {
		if !strings.Contains(report, "lnav: OK") {
			t.Fatalf("expected 'lnav: OK' when lnav is on PATH, got:\n%s", report)
		}
	} else {
		if !strings.Contains(report, "lnav: MISSING") {
			t.Fatalf("expected 'lnav: MISSING' when lnav absent, got:\n%s", report)
		}
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./cmd -run TestDoctor
```
Expected: FAIL.

- [ ] **Step 3: 写 `cmd/doctor.go`**

```go
package cmd

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Check the lnav-cli environment (lnav on PATH, version, permissions)",
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			path, err := exec.LookPath("lnav")
			if err != nil {
				fmt.Fprintln(out, "lnav: MISSING")
				fmt.Fprintln(out, "hint: run `lnav-cli setup` or `brew install lnav`")
				return nil
			}
			ver, verr := exec.Command("lnav", "-V").Output()
			fmt.Fprintf(out, "lnav: OK (%s)\n", path)
			if verr == nil {
				fmt.Fprintf(out, "version: %s", string(ver))
			}
			return nil
		},
	}
}
```

- [ ] **Step 4: 从 `cmd/root.go` 删除 `newDoctorCmd` 占位符**

打开 `cmd/root.go`，移除文件末尾占位的：
```go
func newDoctorCmd() *cobra.Command { return &cobra.Command{Use: "doctor", RunE: func(*cobra.Command, []string) error { return nil }} }
```

- [ ] **Step 5: 运行测试，确认通过**

```bash
go test ./cmd -v
```
Expected: PASS（root + doctor）。

- [ ] **Step 6: Commit + push**

```bash
git add cmd/doctor.go cmd/doctor_test.go cmd/root.go
git commit -m "feat(cmd): add doctor command reporting lnav availability"
git push origin main
```

---

## Task 8 — `cmd/search` shortcut（`+search`）

**Files:**
- Create: `cmd/search.go`, `cmd/search_test.go`
- Modify: `cmd/root.go`（删除 search 占位符）

- [ ] **Step 1: 写 `cmd/search_test.go`**

```go
package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestSearch_DryRun_PrintsLnavArgv(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+search", "--dry-run", "--since", "1h", "--level", "error", "-s", "/tmp/foo.log", "timeout"})
	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	got := out.String()
	for _, want := range []string{
		"lnav",
		":filter-in timeout",
		":set-min-log-level error",
		":hide-unmarked-lines-before",
		":write-json-to -",
		"/tmp/foo.log",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in argv; got:\n%s", want, got)
		}
	}
}

func TestSearch_RequiresSource(t *testing.T) {
	root := NewRootCmd()
	out := &bytes.Buffer{}
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs([]string{"+search", "--dry-run", "anything"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected error when no --source given")
	}
}
```

- [ ] **Step 2: 运行测试，确认失败**

```bash
go test ./cmd -run TestSearch
```
Expected: FAIL.

- [ ] **Step 3: 写 `cmd/search.go`**

```go
package cmd

import (
	"github.com/spf13/cobra"

	"github.com/MonsterChenzhuo/lnav-cli/internal/lnavexec"
	"github.com/MonsterChenzhuo/lnav-cli/internal/output"
	"github.com/MonsterChenzhuo/lnav-cli/internal/timerange"
)

type searchOpts struct {
	pattern string
	level   string
}

func newSearchCmd() *cobra.Command {
	var local searchOpts
	c := &cobra.Command{
		Use:   "+search [pattern]",
		Short: "regex/level search across one or more log sources",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 1 {
				local.pattern = args[0]
			}
			if len(globalOpts.Sources) == 0 {
				return output.Errorf("missing_source", "at least one --source / -s is required").
					WithHint("pass a path, glob, or registered alias")
			}
			so := lnavexec.SearchOpts{
				Pattern: local.pattern,
				Level:   local.level,
				Files:   globalOpts.Sources,
			}
			if globalOpts.Since != "" {
				ts, err := timerange.Parse(globalOpts.Since)
				if err != nil {
					return output.Errorf("bad_since", "%v", err)
				}
				so.SinceTS = ts.UTC().Format("2006-01-02T15:04:05Z")
			}
			if globalOpts.Until != "" {
				ts, err := timerange.Parse(globalOpts.Until)
				if err != nil {
					return output.Errorf("bad_until", "%v", err)
				}
				so.UntilTS = ts.UTC().Format("2006-01-02T15:04:05Z")
			}
			runner := lnavexec.Runner{
				DryRun: globalOpts.DryRun,
				Stdout: cmd.OutOrStdout(),
				Stderr: cmd.ErrOrStderr(),
			}
			return runner.Run(lnavexec.BuildSearchArgs(so))
		},
	}
	c.Flags().StringVar(&local.pattern, "pattern", "", "regex (alternative to positional)")
	c.Flags().StringVar(&local.level, "level", "", "minimum log level (info|warning|error|fatal)")
	return c
}
```

- [ ] **Step 4: 删除 `cmd/root.go` 的 search 占位符**

移除文件末尾：
```go
func newSearchCmd() *cobra.Command {
	return &cobra.Command{Use: "+search", Short: "regex/level search", RunE: func(*cobra.Command, []string) error { return nil }}
}
```

- [ ] **Step 5: 运行测试，确认通过**

```bash
go test ./cmd -v
```
Expected: PASS。

- [ ] **Step 6: Commit + push**

```bash
git add cmd/search.go cmd/search_test.go cmd/root.go
git commit -m "feat(cmd): add +search shortcut backed by lnavexec"
git push origin main
```

---

## Task 9 — Dry-run E2E

**Files:**
- Create: `tests/e2e/dryrun/search_dryrun_test.go`

- [ ] **Step 1: 写 E2E 测试**

```go
package dryrun_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestCLI_Search_DryRun(t *testing.T) {
	out, err := exec.Command("go", "run", "../../..",
		"+search", "--dry-run", "--since", "30m", "-s", "/var/log/app.log", "panic").CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\noutput: %s", err, out)
	}
	s := string(bytes.TrimSpace(out))
	for _, want := range []string{"lnav", ":filter-in panic", ":write-json-to -", "/var/log/app.log"} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %q in output:\n%s", want, s)
		}
	}
}
```

- [ ] **Step 2: 运行**

```bash
go test ./tests/e2e/dryrun -v
```
Expected: PASS。

- [ ] **Step 3: Commit + push**

```bash
git add tests/e2e/dryrun
git commit -m "test(e2e): add dry-run E2E for +search"
git push origin main
```

---

## Task 10 — Live E2E with nginx fixture（可选，lnav 不在 PATH 时 skip）

**Files:**
- Create: `tests/fixtures/nginx.log`, `tests/e2e/live/search_live_test.go`

- [ ] **Step 1: 写 fixture `tests/fixtures/nginx.log`**

```
192.168.0.1 - - [21/Apr/2026:10:00:00 +0000] "GET / HTTP/1.1" 200 512 "-" "curl/8"
192.168.0.2 - - [21/Apr/2026:10:00:05 +0000] "GET /api HTTP/1.1" 500 128 "-" "curl/8"
192.168.0.3 - - [21/Apr/2026:10:00:10 +0000] "POST /login HTTP/1.1" 502 0 "-" "curl/8"
192.168.0.4 - - [21/Apr/2026:10:00:15 +0000] "GET / HTTP/1.1" 200 512 "-" "curl/8"
```

- [ ] **Step 2: 写 Live E2E**

```go
package live_test

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCLI_Search_Live_5xx(t *testing.T) {
	if _, err := exec.LookPath("lnav"); err != nil {
		t.Skip("lnav not installed; skipping live E2E")
	}
	_, thisFile, _, _ := runtime.Caller(0)
	fixture := filepath.Join(filepath.Dir(thisFile), "..", "..", "fixtures", "nginx.log")

	out, err := exec.Command("go", "run", "../../..",
		"+search", "-s", fixture, `5\d\d`).CombinedOutput()
	if err != nil {
		t.Fatalf("run: %v\n%s", err, out)
	}
	s := string(bytes.TrimSpace(out))
	if !strings.Contains(s, "500") || !strings.Contains(s, "502") {
		t.Fatalf("expected 500 and 502 in output, got:\n%s", s)
	}
}
```

- [ ] **Step 3: 运行**

```bash
go test ./tests/e2e/live -v
```
Expected: PASS 或 SKIP。

- [ ] **Step 4: Commit + push**

```bash
git add tests/fixtures tests/e2e/live
git commit -m "test(e2e): add live +search against nginx fixture"
git push origin main
```

---

## Task 11 — Skills: `lnav-shared`

**Files:**
- Create: `skills/lnav-shared/SKILL.md`

- [ ] **Step 1: 写 `skills/lnav-shared/SKILL.md`**

```markdown
---
name: lnav-shared
version: 0.1.0
description: "lnav-cli 共享基础：环境自检（doctor）、lnav 安装（setup）、通用参数语义（--source/--since/--format）、结构化错误处理、sources 别名概念。所有 lnav-* 场景 skill 必须先读本文件。当 Agent 遇到 lnav_not_found、parse_error 或首次使用 lnav-cli 时必读。"
metadata:
  requires:
    bins: ["lnav-cli"]
  cliHelp: "lnav-cli --help"
---

# lnav-cli 共享规则

## 环境准备

第一次使用先跑 `lnav-cli doctor`：
- 输出 `lnav: OK` 即可进入具体场景 skill。
- 输出 `lnav: MISSING`：引导用户 `brew install lnav`（macOS）或 `apt install lnav`（Linux）；v1.0 起可用 `lnav-cli setup` 自动安装。

## 通用参数

| Flag | 语义 | 示例 |
|------|------|------|
| `-s/--source` | 日志源：别名、文件路径、或 glob；可重复 | `-s nginx-prod` / `-s /var/log/*.log` |
| `--since` | 开始时间：`15m`、`2h`、`2026-04-21T10:00:00Z` | `--since 30m` |
| `--until` | 结束时间；缺省=现在 | `--until 2026-04-21T11:00:00Z` |
| `--format` | `ndjson`（默认）/`json`/`table`/`pretty` | `--format json` |
| `--limit` | 最多返回行数；0=不限 | `--limit 200` |
| `--dry-run` | 只打印将要执行的 lnav argv，不真正执行 | `--dry-run` |

## 输出约定

- **stdout = 数据**：JSON envelope `{"data":[...],"_meta":{"source":"...","count":N}}` 或 NDJSON。
- **stderr = 日志/错误**：任何非数据内容走 stderr。
- 错误为结构化 JSON：

```json
{"code":"lnav_not_found","message":"lnav executable not found on PATH","hint":"run: lnav-cli setup"}
```

## 常见错误码

| code | 含义 | 修复 |
|------|------|------|
| `lnav_not_found` | 系统未安装 lnav | `brew install lnav` 或 `lnav-cli setup` |
| `missing_source` | 未提供 `--source` | 追加 `-s <path-or-alias>` |
| `bad_since` / `bad_until` | 时间表达式无法解析 | 使用 `1h` / `30m` / RFC3339 |
| `lnav_exec_failed` | lnav 子进程非零退出 | 去掉 `--dry-run` 看 stderr；必要时加 `--limit` 缩小输入 |

## 禁止事项

- **禁止**在未加 `--duration` 或 `--max-events` 的情况下调用 `+tail`（v1.0 起提供）。
- **禁止**把用户原始字符串直接拼进 regex；遇到含特殊字符的字面量，请先用 `regexp.QuoteMeta` 或改走 `--pattern-file`（v1.0）。

## 参考 skill

- [`../lnav-search/SKILL.md`](../lnav-search/SKILL.md) — 搜索/过滤
- 后续补齐：`lnav-sql` / `lnav-summary` / `lnav-tail`
```

- [ ] **Step 2: Commit + push**

```bash
git add skills/lnav-shared
git commit -m "docs(skill): add lnav-shared SKILL"
git push origin main
```

---

## Task 12 — Skills: `lnav-search` + references

**Files:**
- Create: `skills/lnav-search/SKILL.md`
- Create: `skills/lnav-search/references/regex-cheatsheet.md`
- Create: `skills/lnav-search/references/time-window.md`

- [ ] **Step 1: 写 `skills/lnav-search/SKILL.md`**

```markdown
---
name: lnav-search
version: 0.1.0
description: "lnav-cli 日志搜索/过滤技能。当用户要按关键词、正则、时间窗、日志级别查找日志行时触发（grep/tail 替代场景）。核心命令 +search，输出 NDJSON。首次使用或遇到 lnav_not_found 必须先读 ../lnav-shared/SKILL.md。"
metadata:
  requires:
    bins: ["lnav-cli", "lnav"]
  cliHelp: "lnav-cli +search --help"
---

# lnav-cli +search — 日志搜索/过滤

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lnav-shared/SKILL.md`](../lnav-shared/SKILL.md)。**

## 核心场景

1. 按关键词/正则查找：`lnav-cli +search -s <source> "pattern"`
2. 只看某级别及以上：`--level error`
3. 限定时间窗：`--since 1h --until 30m`
4. 限制输出量（Agent 上下文保护）：`--limit 200`

## Shortcut 速查

| Flag | 说明 | 示例 |
|------|------|------|
| 位置参数 / `--pattern` | 正则 | `"timeout|refused"` |
| `--level` | `info` / `warning` / `error` / `fatal` | `--level error` |
| `-s/--source` | 日志源（见 lnav-shared） | `-s /var/log/nginx/error.log` |
| `--since` / `--until` | 时间窗 | `--since 30m` |
| `--limit` | 最多行数 | `--limit 200` |
| `--dry-run` | 打印 lnav argv | `--dry-run` |

## 推荐工作流

1. **先 dry-run 预审**（第一次跑或不确定时）：
   ```bash
   lnav-cli +search --dry-run -s /var/log/app.log --since 10m --level error "panic|fatal"
   ```
2. **确认 argv 后再实跑**（去掉 `--dry-run`）。
3. **Agent 上下文保护**：默认加 `--limit`；必要时二次缩窄 `--since`。

## 常见陷阱

- lnav `:filter-in` 使用 PCRE 正则，`\d` 需转义：示例 `"5\\d\\d"`（shell 里还要再转一层）。
- `--since` 支持相对值（如 `2h`），也支持 RFC3339；详见 [time-window.md](references/time-window.md)。
- **禁止**把用户原样字符串拼进正则；字面量用 `regexp.QuoteMeta` 处理后再传入。

## 常见错误

| 错误码 | 解释 | 修复 |
|---------|------|------|
| `missing_source` | 忘了 `-s` | 追加 `-s <path-or-alias>` |
| `bad_since` | 时间表达式 | 改用 `1h` / `30m` / RFC3339 |
| `lnav_exec_failed` | lnav 非零退出 | 看 stderr；通常是文件不可读或 regex 非法 |

## 参考文档

- [regex cheatsheet](references/regex-cheatsheet.md)
- [time window 语义](references/time-window.md)
```

- [ ] **Step 2: 写 `references/regex-cheatsheet.md`**

```markdown
# lnav PCRE Regex Cheatsheet

lnav 的 `:filter-in` 使用 PCRE。常用写法：

| 目的 | 正则 | 备注 |
|------|------|------|
| 数字 | `\d+` | shell 里常要写成 `"\\d+"` 避免吃转义 |
| HTTP 5xx | `5\d\d` | 访问日志匹配服务端错误 |
| 多关键词 OR | `timeout|refused` | 注意反引号与 shell 转义 |
| 反向过滤 | 在 lnav 里用 `:filter-out`；lnav-cli v0.1 暂未暴露 | v1.0 追加 `--exclude` |
| 字面量字符 | 用 `\` 转义 `.`/`(`/`)`/`?`/`*`/`+` | 不转义会变成元字符 |

## 安全

- 绝不要把用户原始字符串拼成正则。Go 里先 `regexp.QuoteMeta(s)` 再传给 lnav-cli。
- lnav-cli 会拒绝含换行的 pattern（panic 防御）。
```

- [ ] **Step 3: 写 `references/time-window.md`**

```markdown
# --since / --until 语义

## 相对时长

基于"当前时间"回退：

| 输入 | 含义 |
|------|------|
| `30m` | 最近 30 分钟 |
| `1h` | 最近 1 小时 |
| `24h` | 最近 1 天 |

## 绝对时间

任一格式均可：

- RFC3339（推荐）：`2026-04-21T10:00:00Z`
- `2006-01-02T15:04:05`（无时区=本地）
- `2006-01-02 15:04:05`
- `2006-01-02`（整天）

## 语义

`--since` 是**下界**（该时刻之后的日志）；`--until` 是**上界**。缺省 `--until` 等同于"现在"。

底层映射到 lnav `:hide-unmarked-lines-before` / `:hide-unmarked-lines-after`。

## 示例

```bash
lnav-cli +search -s app --since 2h                       # 最近两小时
lnav-cli +search -s app --since 2026-04-20T10:00:00Z \
                           --until 2026-04-20T11:00:00Z  # 一小时窗口
```
```

- [ ] **Step 4: Commit + push**

```bash
git add skills/lnav-search
git commit -m "docs(skill): add lnav-search SKILL and references"
git push origin main
```

---

## Task 13 — `scripts/skills-check.sh` + README

**Files:**
- Create: `scripts/skills-check.sh`
- Create: `README.md`
- Create: `README.zh.md`

- [ ] **Step 1: 写 `scripts/skills-check.sh`**

```bash
#!/usr/bin/env bash
set -euo pipefail

fail=0
shopt -s nullglob
for skill in skills/*/SKILL.md; do
  for field in '^name:' '^version:' '^description:'; do
    if ! head -20 "$skill" | grep -q "$field"; then
      echo "[skills-check] $skill missing field matching $field" >&2
      fail=1
    fi
  done
  dir="$(dirname "$skill")"
  # Validate references/*.md links inside SKILL.md exist
  while IFS= read -r ref; do
    target="$dir/$ref"
    if [[ ! -f "$target" ]]; then
      echo "[skills-check] $skill references missing file: $target" >&2
      fail=1
    fi
  done < <(grep -oE 'references/[A-Za-z0-9_.-]+\.md' "$skill" | sort -u)
done
exit "$fail"
```

- [ ] **Step 2: 赋可执行权限并运行**

```bash
chmod +x scripts/skills-check.sh
bash scripts/skills-check.sh && echo "skills-check OK"
```
Expected: `skills-check OK`。

- [ ] **Step 3: 写 `README.md`**

```markdown
# lnav-cli

[English](./README.md) | [中文](./README.zh.md)

Agent-friendly wrapper around [lnav](https://lnav.org) — drive log search & analysis from Claude Code with one-line commands.

## Status

MVP: `lnav-cli doctor` + `lnav-cli +search` + skills `lnav-shared`, `lnav-search`.
Roadmap (v1.0): `+sql`, `+summary`, `+tail`, sources aliases, `lnav-cli setup`.

## Quick Start

```bash
# 1. Install lnav
brew install lnav               # macOS
# sudo apt install lnav         # Ubuntu

# 2. Build lnav-cli
git clone git@github.com:MonsterChenzhuo/lnav-cli.git
cd lnav-cli && make build

# 3. Verify
./lnav-cli doctor

# 4. Search
./lnav-cli +search -s /var/log/nginx/error.log --since 1h --level error "timeout"
```

## For Claude Code

Copy the `skills/` directory into your Claude Code skills path
(`~/.claude/skills/` in typical installs) so the agent can discover
`lnav-shared` and `lnav-search`.
See each skill's `SKILL.md` for the contract.

## Development

See [AGENTS.md](./AGENTS.md) for the pre-PR checklist and conventions.
```

- [ ] **Step 4: 写 `README.zh.md`**

```markdown
# lnav-cli

[English](./README.md) | [中文](./README.zh.md)

为 AI Agent 封装 [lnav](https://lnav.org) 的日志分析能力 —— Claude Code 一行命令完成日志检索与分析。

## 状态

MVP：`lnav-cli doctor` + `lnav-cli +search` + skills `lnav-shared`, `lnav-search`。
路线图（v1.0）：`+sql`、`+summary`、`+tail`、sources 别名、`lnav-cli setup`。

## 快速开始

```bash
# 1. 安装 lnav
brew install lnav               # macOS
# sudo apt install lnav         # Ubuntu

# 2. 构建 lnav-cli
git clone git@github.com:MonsterChenzhuo/lnav-cli.git
cd lnav-cli && make build

# 3. 自检
./lnav-cli doctor

# 4. 搜索日志
./lnav-cli +search -s /var/log/nginx/error.log --since 1h --level error "timeout"
```

## 给 Claude Code 使用

把 `skills/` 目录下的 skill 放到 Claude Code 的 skills 路径
（通常是 `~/.claude/skills/`），Agent 即可加载 `lnav-shared` 与 `lnav-search`。

## 开发约定

见 [AGENTS.md](./AGENTS.md)。
```

- [ ] **Step 5: Commit + push**

```bash
git add scripts/skills-check.sh README.md README.zh.md
git commit -m "docs: add README, zh README, and skills-check script"
git push origin main
```

---

## Task 14 — 最终质量门禁一次性验证

- [ ] **Step 1: 运行完整 pre-PR 流程**

```bash
gofmt -l .                     # expect no output
go vet ./...
go mod tidy && git diff --exit-code go.mod go.sum || { echo "go mod tidy changed files"; exit 1; }
make unit-test
make e2e-dryrun
make skills-check
```
Expected: 全部通过（如有 gofmt 输出或 go mod tidy 改动，先修复再重跑）。

- [ ] **Step 2: 提交任何修正**

```bash
git status
git add -A
git commit -m "chore: tidy up after pre-PR gate" || echo "nothing to commit"
git push origin main
```

---

## Self-Review

**Spec coverage:**
- MVP Goal（打通 +search 全链路 + 2 个 skills）：Task 5-12 覆盖。
- 非目标/v1.0 范围：显式排除在本计划外，留到下一份 plan。
- 测试层级（单元 + dry-run E2E + live E2E）：Task 3/4/5/6/9/10 覆盖。
- 质量门禁（gofmt/vet/mod tidy/unit-test/skills-check）：Makefile + Task 14 覆盖。

**Placeholder scan:** 无 TODO/TBD；每一步都给出完整代码或具体命令。

**Type consistency:** `SearchOpts` / `Runner` / `Err` / `Meta` / `GlobalOpts` 在后续任务中的用法与 Task 4-6 定义一致；`Errorf(code, msg).WithHint(...)` 调用链在 runner.go、search.go 中一致。
