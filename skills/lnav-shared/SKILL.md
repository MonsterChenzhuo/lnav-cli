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
