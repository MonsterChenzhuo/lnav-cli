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
