# 为什么 +tail 必须有界

Agent 模型（Claude Code 等）调用 `+tail` 时，命令不会主动结束 = Agent 永远等不到结果 = 会话卡死。

lnav-cli 强制要求至少一项：
- `--duration <time>`: 到点退出（v1.0 已实现；用 `context.WithTimeout` + `exec.CommandContext`）。
- `--max-events <N>`: 收满 N 条退出（v1.0 仅接收 flag，v1.x 实现严格计数；没有 `--duration` 只给 `--max-events` 当前版本仍返回错误，建议同时加 `--duration` 兜底）。

## 经验

- 调试线上问题：先用 `+summary` 看样貌，再用 `+tail --duration 60s` 持续观察。
- 发布期盯错误：`+tail --duration 15m --level error`，够用且不至于卡太久。
