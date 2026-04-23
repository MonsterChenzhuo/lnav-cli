# lnav-cli

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-%3E%3D1.25-blue.svg)](https://go.dev/)
[![CI](https://github.com/MonsterChenzhuo/lnav-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/MonsterChenzhuo/lnav-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/MonsterChenzhuo/lnav-cli.svg)](https://github.com/MonsterChenzhuo/lnav-cli/releases)

[English](./README.md) | [中文](./README.zh.md)

一个面向 **AI Agent** 的 [lnav](https://lnav.org)（终端日志查看器）封装。`lnav-cli` 把 lnav 的 headless 模式暴露成几个结构化、一行就能跑完的命令，让 Claude Code 等 Agent 排障从"开 tmux"变成"发一条消息"。

[安装](#安装) · [快速开始](#快速开始) · [命令](#命令) · [Agent Skills](#agent-skills) · [高级用法](#高级用法) · [安全与限制](#安全与限制) · [贡献](#贡献)

## 为什么选 lnav-cli？

- **Agent 友好的输出**：所有命令 stdout 只吐 lnav 的 JSON 流（ndjson）或结构化信封 `{data, _meta}`；错误走 stderr 的 `{code, message, hint}`。没有 ANSI、没有分页器、没有交互提示。
- **默认有边界**：`+tail` 必须带 `--duration` 或 `--max-events`，Agent 不可能因此挂住会话。
- **安全的 argv 构造**：lnav 参数在 Go 里纯结构化拼装（`internal/lnavexec`），遇到换行注入直接 panic。没有 shell，没有模板展开。
- **3 个 shortcut 覆盖 90% 排障场景**：`+search`（正则 + 级别 + 时间窗）、`+sql`（SQLite 查询日志，强制先看 schema）、`+summary`（级别分布 + Top-N 错误 + 直方图），再加 `+tail` 做有界跟踪。
- **命名源（source）**：`lnav-cli source add backend --paths ...` 把一堆路径收敛成一个别名，下游命令直接引用。
- **内置 Claude Code skills**：五个 skill（`lnav-shared`、`lnav-search`、`lnav-sql`、`lnav-summary`、`lnav-tail`）随仓库发布，Agent 拿来即用。

## 能力一览

| 类别         | 能力                                                                                      |
| ------------ | ----------------------------------------------------------------------------------------- |
| 正则搜索     | `+search` 支持 `--level`、`--since`、`--until`；输出 ndjson 事件                            |
| SQL 日志查询 | `+sql` 带 `--show-schema` 短路；用 SQLite 语法查询 lnav 的日志表                            |
| 排障信封     | `+summary` 合并级别分布、Top-N 错误内容、时间直方图为一个 envelope                          |
| 有界跟踪     | `+tail` 必须 `--duration` 或 `--max-events`；到期自动 context cancel                        |
| 命名源       | `source add/ls/show/rm`，持久化到 `~/.lnav-cli/sources.yaml`                                |
| 时间窗       | 相对（`1h`、`30m`）+ 绝对（RFC3339、`2006-01-02T15:04:05`、纯日期）                          |
| 健康检查     | `doctor` 检测 PATH 上的 `lnav` 与版本；`setup` 打印平台安装提示（不联网、不 sudo）          |
| Dry-run      | `--dry-run` 打印即将执行的 lnav argv —— Agent 自检首选                                       |

## 安装

### 前置条件

- `lnav` ≥ 0.12 必须在 PATH 上（本工具只做封装，不内置 lnav）
- Go `1.25+`（仅源码编译需要）

### 先装 lnav

```bash
brew install lnav                      # macOS
sudo apt-get install -y lnav           # Debian / Ubuntu
sudo dnf install -y lnav               # Fedora
scoop install lnav                     # Windows (scoop)
```

更完整的列表见 <https://lnav.org/downloads>。

### 安装 lnav-cli

**方式 1 —— 下载 release 产物（推荐）**

```bash
# 平台/架构见 https://github.com/MonsterChenzhuo/lnav-cli/releases
curl -L -o lnav-cli.tar.gz \
  https://github.com/MonsterChenzhuo/lnav-cli/releases/latest/download/lnav-cli-<ver>-<os>-<arch>.tar.gz
tar -xzf lnav-cli.tar.gz
sudo install -m755 lnav-cli-*/lnav-cli /usr/local/bin/lnav-cli
lnav-cli version
```

**方式 2 —— 源码构建**

```bash
git clone https://github.com/MonsterChenzhuo/lnav-cli.git
cd lnav-cli
make install                 # 默认安装到 /usr/local/bin
#  或：make install PREFIX=$HOME/.local
lnav-cli version
```

**方式 3 —— `go install`**

```bash
go install github.com/MonsterChenzhuo/lnav-cli@latest
```

## 快速开始

### 面向人类用户

> **AI 助手请直接看** [面向 AI Agent](#面向-ai-agent)。

```bash
# 1. 检查环境
lnav-cli doctor

# 2. 注册一个命名源（可选但推荐）
lnav-cli source add nginx \
  --paths "/var/log/nginx/access.log,/var/log/nginx/error.log" \
  --default-level warning

# 3. 搜索最近 1 小时内 WARN 及以上、含 "timeout" 的日志
lnav-cli +search -s nginx --since 1h --level warning "timeout"

# 4. 排障：最近 6 小时的级别分布 + 直方图
lnav-cli +summary -s nginx --since 6h --histogram 5m

# 5. SQL 下钻（先看 schema！）
lnav-cli +sql -s nginx --show-schema
lnav-cli +sql -s nginx \
  "SELECT log_level, count(*) c FROM access_log GROUP BY 1 ORDER BY c DESC"
```

### 面向 AI Agent

> 以下步骤是 Claude Code（或任何支持 SKILL 的 Agent）应当执行的范式。所有结果 stdout 只有 JSON/NDJSON；错误 stderr 只有 `{code, message, hint}`。

**第 1 步 —— 验证环境**

```bash
lnav-cli doctor
```

**第 2 步 —— 先 dry-run 确认要执行的 argv**

```bash
lnav-cli +search -s /var/log/app.log --since 30m --level error --dry-run "panic|fatal"
```

检查打印出来的 lnav 命令，符合预期再真跑。

**第 3 步 —— 执行**

```bash
lnav-cli +search -s /var/log/app.log --since 30m --level error "panic|fatal"
```

**第 4 步 —— 跟踪实时流时，必须加边界**

```bash
# 不加 --duration 或 --max-events 会被拒绝。
lnav-cli +tail -s /var/log/app.log --duration 30s --level error
```

**第 5 步 —— 安装内置 skills，让 Agent 知道命令契约**

```bash
cp -r skills/* ~/.claude/skills/
```

[`skills/lnav-shared/SKILL.md`](./skills/lnav-shared/SKILL.md) 是所有其他 skill 都引用的公共契约。

## 命令

| 命令                    | 作用                                                                |
| ----------------------- | ------------------------------------------------------------------- |
| `lnav-cli doctor`       | 探测 PATH 上的 `lnav`，打印版本                                       |
| `lnav-cli setup`        | 打印平台安装提示（不联网、不 sudo）                                   |
| `lnav-cli version`      | 打印版本、commit、构建日期、平台                                      |
| `lnav-cli +search`      | 正则搜索，支持 `--level` 和 `--since/--until`                         |
| `lnav-cli +sql`         | 对日志文件跑 SQLite 查询；`--show-schema` 先看表结构                   |
| `lnav-cli +summary`     | 排障信封：级别计数 + Top-N 错误 + 直方图                              |
| `lnav-cli +tail`        | 有界跟踪；必须带 `--duration` 或 `--max-events`                        |
| `lnav-cli source add`   | 注册命名源（paths 或 command）                                        |
| `lnav-cli source ls`    | 列出已注册的源                                                        |
| `lnav-cli source show`  | 查看源的原始定义                                                      |
| `lnav-cli source rm`    | 删除源                                                                |

### 全局 flag

| Flag             | 默认值    | 说明                                                          |
| ---------------- | --------- | ------------------------------------------------------------- |
| `-s, --source`   | 无        | 别名/文件/glob，可重复；shortcut 必填                            |
| `--since`        | 无        | 时间窗起点（`1h`、`2026-04-22T10:00:00Z`、…）                   |
| `--until`        | 无        | 时间窗终点                                                      |
| `--format`       | `ndjson`  | `ndjson\|json\|table\|pretty`（结构化信封走 json）               |
| `--limit`        | `0`       | 最大行数 —— 已解析；严格执行列入 v1.x                            |
| `--dry-run`      | `false`   | 打印 lnav argv 而不执行                                        |

## Agent Skills

五个 skill 放在 [`skills/`](./skills/)。拷贝到 Claude Code 的 skills 目录（`~/.claude/skills/`）或走你自己的 skill 分发流程。

| Skill             | 说明                                                                                        |
| ----------------- | ------------------------------------------------------------------------------------------- |
| `lnav-shared`     | 所有 skill 都引用的公共契约 —— 输出格式、错误 envelope、时间窗                                 |
| `lnav-search`     | `+search` 契约 + 正则 cheatsheet + 时间窗模板                                                  |
| `lnav-sql`        | `+sql` 契约，强制 **schema-first**（先 `--show-schema` 再写 SQL）                              |
| `lnav-summary`    | `+summary` 排障流程（级别 → Top-N 错误 → 直方图放大）                                          |
| `lnav-tail`       | `+tail` 有界跟踪规则（`--duration` / `--max-events` 必填）                                     |

## 高级用法

### 命名源

源可以是 **paths**（文件/glob 列表）或 **command**（shell 命令，其 stdout 喂给 lnav）。存储在 `~/.lnav-cli/sources.yaml`（可用 `LNAV_CLI_CONFIG_DIR` 覆盖）。

```bash
lnav-cli source add api \
  --paths "/var/log/app/*.log" \
  --default-level warning

lnav-cli source show api
```

> `command:` 类型源当前已持久化，但把 stdout 喂给 lnav 的管道计划在 v1.x 完成。目前 `+search/+sql/+summary/+tail` 遇到 command 源会返回 `unsupported_stdin_source` 错误。

### 时间窗

支持的格式（不带时区默认 UTC）：

- 相对：`30s`、`5m`、`2h`
- RFC3339：`2026-04-22T10:00:00Z`、`2026-04-22T10:00:00+08:00`
- 空格变种：`2026-04-22 10:00:00`
- 纯日期：`2026-04-22`

### Dry-run

所有 shortcut 支持 `--dry-run`，打印完整 lnav argv。Agent 自检首选方式。

```bash
$ lnav-cli +sql -s nginx --show-schema --dry-run
lnav -n -q .schema /var/log/nginx/access.log /var/log/nginx/error.log
```

### 输出信封

返回结构化 JSON 的命令用：

```json
{
  "data": {...},
  "_meta": {"count": 3}
}
```

错误：

```json
{
  "code": "unbounded_tail",
  "message": "+tail requires --duration or --max-events to prevent hanging the agent",
  "hint": "pass --duration 30s or --max-events 500"
}
```

## 安全与限制

- **无 shell、无 eval**：lnav 调用都是 `[]string` argv，经 `internal/lnavexec` 构造。输入含换行直接 panic —— Agent 必须先清洗或切分。
- **不提权**：`lnav-cli` 不会 `sudo`、不碰 systemd、不写 `$LNAV_CLI_CONFIG_DIR || ~/.lnav-cli/` 以外的路径。
- **路径校验**：源别名在 exec 前解析成真实文件；缺失会以 `resolve_source` 错误暴露。
- **有界跟踪强制**：`+tail` 必须 `--duration` 或 `--max-events`；context 带 deadline，进程不可能超时存活。
- **已知缺口（规划 v1.x）**：严格 `--limit` 强制；`command:` 源的 stdin 管道；非 summary 命令的 `--format table|pretty`。

## 项目结构

```
lnav-cli/
├── cmd/                    cobra 命令树（root / doctor / setup / version /
│                           +search / +sql / +summary / +tail / source）
├── internal/
│   ├── build/              版本元数据（ldflags 注入）
│   ├── lnavexec/           纯 argv 构造 + 子进程 runner
│   ├── output/             结构化 Err + JSON/NDJSON envelope 助手
│   ├── source/             sources.yaml 加载/保存/解析
│   └── timerange/          相对 + 绝对时间解析
├── skills/                 Claude Code skills（lnav-shared + 4 个 shortcut）
├── scripts/                skills-check.sh frontmatter 校验
├── tests/
│   ├── e2e/dryrun/         每个 shortcut 的 dry-run E2E（无需 lnav）
│   ├── e2e/live/           live E2E（lnav 不在时自动 skip）
│   └── fixtures/           nginx.log 样例
├── docs/superpowers/       设计稿 + MVP/v1.0 实施计划
└── .github/workflows/      ci.yml（PR 门禁） + release.yml（goreleaser）
```

## 贡献

1. 先读 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解工作流、分支策略、commit 规约。
2. 若你是 AI Agent 或在驱动 Agent，请读 [CLAUDE.md](./CLAUDE.md) —— 这是本仓库面向 Agent 的 pre-PR checklist 与代码约束。
3. 推前跑完本地完整闸：
   ```bash
   make test       # vet + unit + dry-run E2E + skills-check
   make lint       # golangci-lint v2.1.6
   ```
4. 用 [Conventional Commits](https://www.conventionalcommits.org/) 提交（`feat:`、`fix:`、`docs:`、`test:`、`refactor:`、`chore:`、`ci:`）。
5. PR 必须通过 `ci` workflow；`main` 上通过 `release.yml` + GoReleaser 自动发版。

## 许可

[MIT](./LICENSE)。lnav 本身另有许可，见 <https://github.com/tstack/lnav>。
