---
name: lnav-tail
version: 0.1.0
description: "lnav-cli 实时跟随技能。当用户要‘盯一会日志’、‘等新错误出现’、‘运行中捕获告警’时触发。核心命令 +tail。CRITICAL: MUST 提供 --duration 或 --max-events，否则命令会卡住 Agent。首次使用必须先读 ../lnav-shared/SKILL.md。"
metadata:
  requires:
    bins: ["lnav-cli", "lnav"]
  cliHelp: "lnav-cli +tail --help"
---

# lnav-cli +tail — 实时跟随

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../lnav-shared/SKILL.md`](../lnav-shared/SKILL.md)。**
**BLOCKING — `+tail` MUST 同时带 `--duration` 或 `--max-events`；否则命令永不返回，会挂住 Agent。**

## 典型用法

```bash
# 跟 30 秒，把所有 error 及以上流出来
lnav-cli +tail -s app --duration 30s --level error

# 跟 10 分钟，只看匹配 5xx 的请求
lnav-cli +tail -s nginx-prod --duration 10m "5\\d\\d"
```

## 参数

| Flag | 说明 |
|------|------|
| `--duration` | 自动退出时长：`30s` / `5m` / `1h` |
| `--max-events` | 收集到 N 条后退出（v1.0：cobra 接收，严格实现留 v1.x） |
| `--level` / 正则位置参数 | 与 `+search` 一致 |

## 输出

每一行一条 JSON（NDJSON）。超过 `--duration` 时 lnav 子进程被 context cancel，正常退出。

详见 [bounded-follow.md](references/bounded-follow.md)。
