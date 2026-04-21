---
name: lnav-summary
version: 0.1.0
description: "lnav-cli 错误/异常摘要技能。当用户要快速了解一段日志里‘出了什么错’、错误级别分布、TopN 错误体、错误随时间分布（直方图）时触发。核心命令 +summary，返回聚合 JSON。首次使用必须先读 ../lnav-shared/SKILL.md。"
metadata:
  requires:
    bins: ["lnav-cli", "lnav"]
  cliHelp: "lnav-cli +summary --help"
---

# lnav-cli +summary — 错误摘要

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lnav-shared/SKILL.md`](../lnav-shared/SKILL.md)。**

## 输出结构

```json
{
  "data": {
    "levels":     [ {"log_level":"error", "c": 120} ],
    "top_errors": [ {"log_level":"error", "log_body":"upstream timed out", "c": 80} ],
    "histogram":  [ {"bin":"2026-04-21 10:00", "c": 12} ]
  },
  "_meta": {"count": 3}
}
```

## 参数

| Flag | 默认 | 说明 |
|------|------|------|
| `--top` | 10 | TopN 错误体 |
| `--histogram` | `5m` | 直方图桶：`1m` / `5m` / `1h` |

## 工作流（triage）

1. `+summary -s <source> --since 6h` 先看全貌。
2. 从 `levels` / `top_errors` 锁定目标。
3. 用 `+search -s <source> --since 6h --level error "<top_body 关键字>"` 深挖具体日志。
4. 必要时用 `+sql` 做更细粒度聚合。

详见 [triage-workflow.md](references/triage-workflow.md)。
