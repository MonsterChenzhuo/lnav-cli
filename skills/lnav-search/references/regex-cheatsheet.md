# lnav PCRE Regex Cheatsheet

lnav 的 `:filter-in` 使用 PCRE。常用写法：

| 目的 | 正则 | 备注 |
|------|------|------|
| 数字 | `\d+` | shell 里常要写成 `"\\d+"` 避免吃转义 |
| HTTP 5xx | `5\d\d` | 访问日志匹配服务端错误 |
| 多关键词 OR | `timeout\|refused` | 注意反引号与 shell 转义 |
| 反向过滤 | 在 lnav 里用 `:filter-out`；lnav-cli v0.1 暂未暴露 | v1.0 追加 `--exclude` |
| 字面量字符 | 用 `\` 转义 `.`/`(`/`)`/`?`/`*`/`+` | 不转义会变成元字符 |

## 安全

- 绝不要把用户原始字符串拼成正则。Go 里先 `regexp.QuoteMeta(s)` 再传给 lnav-cli。
- lnav-cli 会拒绝含换行的 pattern（panic 防御）。
