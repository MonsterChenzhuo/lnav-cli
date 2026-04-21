# lnav-cli 设计文档

- Date: 2026-04-21
- Status: Draft（待用户复审）
- Owner: lnav-cli maintainers

## 1. 目标

把 [lnav](https://lnav.org) 的日志检索与分析能力，封装成一个 Agent 友好的命令行工具 `lnav-cli`，并以 Claude Code Skills 的形式交付，使得 Claude Code 通过**一行命令**就能完成日常日志检索、SQL 分析、错误摘要、实时跟随等工作。

风格对齐同仓库下的 `cli/`（lark-cli）：Go + cobra、三段式输出约定（stdout = 数据 / stderr = 日志 / 结构化错误）、Shortcut `+verb` 形态、Skills 分发模型。

## 2. 范围（MVP / v1.0 / v1.x）

### MVP（首个可跑通的版本）
- Go 项目骨架：`cmd/root.go` + `main.go` + `go.mod` + `Makefile` + `AGENTS.md` + `README.md/README.zh.md`。
- `internal/lnavexec`：最朴素的 lnav 子进程封装 + `--dry-run`。
- `lnav-cli doctor`：只检测 `lnav` 是否在 PATH。
- `lnav-cli +search`：打通「参数 → 构造 lnav argv → 执行 → JSON envelope」完整链路；`--source` 暂只支持裸路径/glob。
- `skills/lnav-shared/SKILL.md` + `skills/lnav-search/SKILL.md`。
- Dry-run E2E 若干 + 一条 Live E2E（search nginx fixture）。

### v1.0（完整发布）
- 其余 3 个 shortcut：`+sql`、`+summary`、`+tail`。
- `lnav-cli source add/ls/rm/show`，基于 `~/.lnav-cli/sources.yaml`。
- `lnav-cli setup`：macOS brew / Linux apt / Windows winget + GitHub Release fallback。
- 其余 3 个 skills 及对应 `references/*.md`。
- `--format json/ndjson/table/pretty` 全支持。
- `package.json` + `bin/` shim，跑通 `npx skills add lnav-cli/lnav-cli -y -g`。

### v1.x（后续，非本次交付）
- 索引缓存层（方案 C）。
- `kubectl logs` / `journalctl` 源模板的深度集成。
- Windows CI 矩阵转正。
- 长驻 lnav daemon（方案 B）作为可选后端。

## 3. 使用场景（四类核心场景，决定了 4 个 shortcut）

1. **搜索/过滤**：按 regex + 时间窗 + 日志级别检索，替代 `grep + awk + less`。
2. **SQL 分析**：利用 lnav 内建 SQLite 对日志跑聚合、TopN、GroupBy。
3. **错误/异常摘要**：一键扫描，输出错误级别分布、TopN 错误体、按时间桶的直方图。
4. **实时 tail/跟随**：跟随文件或目录，把新增的匹配事件以事件流方式回吐给 Agent；必须是有界跟随（`--duration` 或 `--max-events`），防止 Agent 阻塞。

## 4. 顶层架构

```
lnav-cli (Go, cobra)
├── cmd/
│   ├── root.go               # 全局 flags: --source, --format, --since/--until, --dry-run
│   ├── setup.go              # lnav-cli setup  (检测/安装 lnav: brew/apt/winget/GH release fallback)
│   ├── doctor.go             # lnav-cli doctor (环境自检: lnav 版本/PATH/权限)
│   ├── source/               # lnav-cli source add/ls/rm/show
│   └── shortcuts/            # +verb agent-friendly 命令
│       ├── search.go
│       ├── sql.go
│       ├── summary.go
│       └── tail.go
├── internal/
│   ├── lnavexec/             # lnav 子进程封装: 构造 argv、-c/-q 编排、超时、stderr 解析
│   ├── source/               # sources 配置加载
│   ├── output/               # JSON/NDJSON/table/pretty 格式化 + 结构化错误 envelope
│   ├── timerange/            # --since/--until 解析（相对/绝对）
│   └── install/              # 各平台 lnav 安装器
├── skills/
│   ├── lnav-shared/SKILL.md
│   ├── lnav-search/SKILL.md  + references/
│   ├── lnav-sql/SKILL.md     + references/
│   ├── lnav-summary/SKILL.md + references/
│   └── lnav-tail/SKILL.md    + references/
├── skill-template/           # 同构于 cli/skill-template
├── main.go
├── go.mod
├── Makefile
├── AGENTS.md
├── README.md / README.zh.md
└── package.json              # 用于 npx skills add 分发
```

### 数据流（单次调用）

```
Claude Code
  └─ lnav-cli +verb --source nginx-prod --since 1h --format ndjson 'pattern'
      └─ Go: 解析 flags + sources → 构造 argv
          └─ exec: lnav -n -c ":filter-in pattern" -c ":set-min-log-level error" \
                         -c ":write-json-to -" -- <resolved files>
              └─ 解析 lnav stdout → 按 --format 输出 {data:[...], _meta:{...}} → stdout
              └─ lnav stderr / Go 错误 → structured error envelope → stderr
```

### 关键约定

- `stdout = 数据，stderr = 日志/错误`。
- 所有错误走结构化 envelope（`code` / `message` / `hint` / 可选 `console_url`），参考 cli/AGENTS.md。
- 接收的路径 / glob 先经 `validate.SafeInputPath`，防止注入到 lnav 的 `-c` 命令里。
- `--dry-run` 不执行 lnav，仅打印完整的 lnav argv 到 stdout（供 Agent 预审与调试）。

## 5. 调用 lnav 的方式（选定方案 A：无状态 headless）

每次 `lnav-cli` 命令都 `fork` 一个 `lnav -n -q ...` / `lnav -n -c ...` 子进程，结束即退出。

- 优点：简单、可预测、零残留状态、匹配 Agent「一行命令」的心智模型。
- 缺点：重复索引大文件有开销；通过 sources 别名 + lnav 自带的 `~/.local/share/lnav/index` 缓存缓解。
- 未来可能性：v1.x 增加独立的索引缓存层（方案 C）；长驻 daemon（方案 B）留作未来选项。

## 6. Shortcuts 详细映射

所有 shortcut 共享全局 flags：`--source/-s`、`--since/--until`、`--format {json|ndjson|table|pretty}`、`--limit N`、`--dry-run`。

### 6.1 `lnav-cli +search` — regex 搜索/过滤

```bash
lnav-cli +search --source nginx-prod --since 1h --level error "timeout|refused"
lnav-cli +search /var/log/app/*.log --pattern "user_id=\d+" --limit 200 --format ndjson
```

映射：
```
lnav -n \
  -c ":filter-in <regex>" \
  -c ":set-min-log-level <level>" \
  -c ":hide-unmarked-lines-before <ts>" \
  -c ":write-json-to -" \
  -- <files>
```

输出（NDJSON 每行一条）：
```json
{"ts":"2026-04-21T10:00:00Z","level":"error","body":"upstream timed out","file":"/var/log/nginx/error.log","line":1234,"fields":{"client":"10.0.0.1"}}
```

### 6.2 `lnav-cli +sql` — SQLite 查询

```bash
lnav-cli +sql -s k8s-api --since 24h \
  'SELECT log_level, count(*) c FROM all_logs GROUP BY log_level ORDER BY c DESC'

lnav-cli +sql -s nginx-prod --query-file queries/top_ips.sql --format table
```

映射：`lnav -n -q "<sql>" -- <files>`。lnav 执行 SQL 后默认以 JSON 写 stdout；Go 层再按 `--format` 转换。

约束（在 Skill 里强制）：调用前必须先跑
```bash
lnav-cli +sql --show-schema -s <source>
```
查看可用表（如 `all_logs` / `syslog_log` / `access_log`）再写 SQL，禁止凭空猜列名。

### 6.3 `lnav-cli +summary` — 错误/异常摘要

```bash
lnav-cli +summary --source app --since 6h
lnav-cli +summary /var/log/syslog --top 10 --histogram 5m
```

由 3 次 lnav 调用组合：
1. 错误级别分布：`SELECT log_level, count(*) FROM all_logs WHERE log_level IN ('error','warning','fatal') GROUP BY 1`
2. TopN 错误体：`SELECT log_level, log_body, count(*) c FROM all_logs WHERE log_level >= 'warning' GROUP BY log_body ORDER BY c DESC LIMIT <top>`
3. 错误时间直方图：`SELECT strftime('%Y-%m-%d %H:%M', log_time) bin, count(*) FROM all_logs WHERE log_level >= 'error' GROUP BY 1`

输出：聚合 JSON `{levels, top_errors, histogram}`。

### 6.4 `lnav-cli +tail` — 实时跟随

```bash
lnav-cli +tail -s nginx-prod --level error --pattern "5\d\d"
lnav-cli +tail -s app --duration 30s
```

映射：`lnav -n -c ":filter-in ..." -c ":set-min-log-level ..." -c ":write-json-to -" -c ":goto 0" --` + lnav 自带 follow。

**关键约束**：必须提供 `--duration` 或 `--max-events` 之一；Skill 中标注为 CRITICAL 阻塞性要求，否则 Agent 会被 tail 卡死。

### 6.5 Source 别名解析

`~/.lnav-cli/sources.yaml`：

```yaml
sources:
  nginx-prod:
    paths: ["/var/log/nginx/access.log*", "/var/log/nginx/error.log"]
    default_level: warning
  k8s-api:
    command: "kubectl logs -n kube-system deploy/kube-apiserver --tail=-1"
```

- `--source` 既可以接别名也可以接裸路径；
- `paths` 型直接作为 lnav 的位置参数；
- `command` 型通过 stdin 喂给 lnav：`<cmd> | lnav -n ...`。

## 7. Skills 内容结构

### 7.1 文件布局

```
skills/
├── lnav-shared/SKILL.md
├── lnav-search/
│   ├── SKILL.md
│   └── references/{regex-cheatsheet.md, time-window.md}
├── lnav-sql/
│   ├── SKILL.md
│   └── references/{schema-introspection.md, common-queries.md, json-log-columns.md}
├── lnav-summary/
│   ├── SKILL.md
│   └── references/triage-workflow.md
└── lnav-tail/
    ├── SKILL.md
    └── references/bounded-follow.md
```

### 7.2 frontmatter 模板

```yaml
---
name: lnav-search
version: 1.0.0
description: "lnav-cli 日志搜索/过滤技能。当用户要按关键词、正则、时间窗、日志级别查找日志行时触发（grep/tail 场景）。核心命令 +search，支持别名 source、--since/--until、--level，输出 NDJSON。首次使用或遇到 'lnav not found' 错误时必须先读 ../lnav-shared/SKILL.md。"
metadata:
  requires:
    bins: ["lnav-cli", "lnav"]
  cliHelp: "lnav-cli +search --help"
---
```

### 7.3 lnav-shared/SKILL.md 骨架

1. 前置检查：第一次使用引导 `lnav-cli doctor`；`lnav` 缺失时用 `lnav-cli setup` 安装。
2. 通用参数速查：`--source` / `--since` / `--format` 语义与示例。
3. Sources 概念：`lnav-cli source add` 注册后 Agent 调用用别名。
4. 错误处理约定：stderr JSON envelope 字段 + 常见 code（`lnav_not_found`、`permission_denied`、`parse_error`）的修复路径。
5. 输出约定：stdout = 数据；区分数据行与 `_meta`。
6. 禁止事项：不允许无上限 `+tail`；禁止把用户原始字符串直接拼入正则，推荐 `--pattern-file`。

### 7.4 场景 Skill 的固定段落模板

借鉴 `lark-calendar` 的结构：**CRITICAL 前置** → **核心场景** → **工作流** → **Shortcut 参数速查表** → **常见错误**。

- `lnav-sql`：CRITICAL「调用 `+sql` 前 MUST 先跑 `lnav-cli +sql --show-schema -s <source>`」。
- `lnav-tail`：CRITICAL「MUST 同时/任一提供 `--duration` 或 `--max-events`，否则命令会卡住」。

### 7.5 Claude Code 学习与分发路径

- 开发期：`--skill-dir ./skills` 直接加载仓内 skills。
- 正式分发：参照 lark-cli，`npx skills add lnav-cli/lnav-cli -y -g` 将 `skills/` 拷入用户 `~/.claude/skills/`。
- `package.json` 配 `files: ["skills/**"]` 与 `bin: {lnav-cli: ...}`（v1 的 `bin` 可仅为从 GitHub Release 下载原生 Go 二进制的 shim 脚本）。

### 7.6 目标使用体验（一行命令）

```bash
# 安装一次
brew install lnav            # 或 lnav-cli setup 自动装
go install .../lnav-cli@latest
npx skills add lnav-cli/lnav-cli -y -g

# 注册日志源一次
lnav-cli source add nginx-prod --paths '/var/log/nginx/*.log'

# Claude Code 一行命令直接用
lnav-cli +summary -s nginx-prod --since 2h
lnav-cli +search  -s nginx-prod --since 30m --level error "upstream timed out"
lnav-cli +sql     -s nginx-prod 'SELECT status, count(*) FROM access_log GROUP BY status'
lnav-cli +tail    -s nginx-prod --level error --duration 60s
```

## 8. 测试与质量门禁

### 8.1 Pre-PR 门禁（对齐 cli/AGENTS.md）

1. `make unit-test`（`-race`）
2. `go vet ./...`
3. `gofmt -l .` 无输出
4. `go mod tidy` 不改动 `go.mod`/`go.sum`
5. `golangci-lint run --new-from-rev=origin/main`

Makefile 目标：`build` / `unit-test` / `e2e-dryrun` / `e2e-live` / `lint` / `skills-check`。

`skills-check` 额外断言：
- 每个 `skills/*/SKILL.md` 的 frontmatter 齐全（`name`、`version`、`description`、`metadata`）。
- 引用到的 `references/*.md` 文件真实存在。
- `cliHelp` 字段里的命令可执行（`--help` 非零退出则失败）。

### 8.2 测试层级

| 层级 | 范围 | 实现 |
|---|---|---|
| 单元测试 | `internal/lnavexec` argv 构造、`internal/timerange` 时间解析、`internal/source` 解析、`internal/output` 格式化 | 表驱动，`t.TempDir()` 隔离 |
| Dry-run E2E（每个 shortcut 必写）| `lnav-cli +search --dry-run ...` 断言生成的 argv | 无需 lnav，CI fork PR 可跑 |
| Live E2E | fixture 日志（nginx / syslog / json app log）真跑 lnav 并断言结果 | CI 安装 lnav；置于 `tests/e2e/` |
| 安装器测试 | `internal/install` 各平台分支 | mock 包管理器调用 |

CI 矩阵：macOS + Ubuntu；Windows v1 降级为 best-effort + 手测，v1.x 再转正。

## 9. 非目标 / 明确排除

- 不复刻 lnav 的 TUI；lnav-cli 永远是 headless。
- 不在 v1 做 daemon / session 复用。
- 不在 v1 深度集成 k8s/journalctl（仅通过 sources 的 `command:` 通用通路支持）。
- 不提供 lnav 之外的索引器或解析器；日志格式识别全权交给 lnav。

## 10. 风险与缓解

| 风险 | 缓解 |
|---|---|
| 大日志每次重新索引慢 | 依赖 lnav 自身的持久索引；v1.x 加缓存层 |
| 用户输入的 regex/路径被注入到 lnav `-c` | `validate.SafeInputPath` + `-c` 参数独立分发、不做字符串拼接 |
| `+tail` 阻塞 Agent | 强制 `--duration` 或 `--max-events`；Skill 标为 CRITICAL |
| Windows lnav 支持不稳 | v1 降级；doctor 明确提示；Windows 用户可用 WSL |
| 多平台 lnav 安装复杂 | setup 命令走包管理器优先，GH Release 二进制 fallback；doctor 负责自检 |

## 11. 后续动作

本设计通过后，用 `superpowers:writing-plans` 技能产出分阶段实施计划（MVP → v1.0 → v1.x 之间的具体任务）。
