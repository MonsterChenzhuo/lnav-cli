# schema-first 工作流

1. **Always** start with:
   ```bash
   lnav-cli +sql --show-schema -s <source>
   ```
2. 读输出，挑出目标列（注意 nginx vs syslog 列名差异）。
3. 再写查询，带 `--dry-run` 验证语法。
4. 正式跑。

## 为什么

- 日志格式被 lnav 识别后才有对应表；不同日志类型的列名差异巨大。
- 凭记忆写 `SELECT status` 在 access_log 里会失败（列名实际是 `sc_status`）。
- schema 查询本身不消耗大量资源，作为前置 cheap。
