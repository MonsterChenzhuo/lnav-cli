---
name: lnav-sql
version: 0.1.0
description: "lnav-cli SQL 分析技能。当用户要对日志做聚合、TopN、分布、GroupBy、JOIN 时触发。核心命令 +sql，底层用 lnav 内建 SQLite。CRITICAL: 写 SQL 前 MUST 先跑 +sql --show-schema 查看可用列，禁止凭空写列名。首次使用必须先读 ../lnav-shared/SKILL.md。"
metadata:
  requires:
    bins: ["lnav-cli", "lnav"]
  cliHelp: "lnav-cli +sql --help"
---

# lnav-cli +sql — 日志 SQL 分析

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lnav-shared/SKILL.md`](../lnav-shared/SKILL.md)。**
**BLOCKING — 写任何 SQL 之前，必须先 `lnav-cli +sql --show-schema -s <source>` 查看列；禁止凭记忆猜列名。**

## 常用表（lnav 自动暴露）

| 表 | 用途 |
|----|------|
| `all_logs` | 统一视图，列：`log_time`、`log_level`、`log_body`、`log_part` 等 |
| `access_log` | nginx/apache 风格访问日志（有 `cs_method`、`cs_uri_stem`、`sc_status`、`sc_bytes` 等） |
| `syslog_log` | syslog（有 `syslog_procname`、`syslog_pid`） |
| 其他格式自动注册 | 取决于 lnav 识别到的日志格式 |

## 工作流

1. `lnav-cli +sql --show-schema -s <source>` 拿到表 & 列名。
2. 先跑 `--dry-run` 预审：
   ```bash
   lnav-cli +sql --dry-run -s nginx-prod 'SELECT sc_status, count(*) FROM access_log GROUP BY 1'
   ```
3. 去掉 `--dry-run` 实跑。建议 `--format table` 用于人类可读，`--format json` 用于 Agent 二次解析（默认）。

## 示例

```bash
# HTTP 状态码分布
lnav-cli +sql -s nginx-prod \
  'SELECT sc_status, count(*) c FROM access_log GROUP BY sc_status ORDER BY c DESC'

# TopN 错误体
lnav-cli +sql -s app --since 24h \
  'SELECT log_body, count(*) c FROM all_logs WHERE log_level="error" GROUP BY log_body ORDER BY c DESC LIMIT 10'
```

## 常见错误

| 错误码 | 含义 | 修复 |
|--------|------|------|
| `missing_query` | 既没 query 也没 --show-schema | 至少给其中一个 |
| `missing_source` | 未指定 `-s` | 加 `-s <source>` |
| `lnav_exec_failed` | SQL 错或列名不存在 | **先跑 `--show-schema` 确认列** |

## 参考

- [schema-first 工作流](references/schema-first.md)
