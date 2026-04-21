# Triage Workflow

```
+summary → 找热点错误
       ↓
+search --level error "<hotspot 关键字>" → 找原始日志
       ↓
+sql → 关联维度（user_id / host / status）定位根因
```

## 经验

- 若 `levels` 里 warning 远多于 error，往往是**上游系统降级**信号，而不是应用 bug。
- `histogram` 显示错误在某时刻突增，先看同期的部署记录 / 流量图。
- `top_errors` 若全是同一行文本，考虑用 `+search` + `--limit` 翻看原始日志，而不是只看聚合。
