# lnav-cli

[English](./README.md) | [中文](./README.zh.md)

为 AI Agent 封装 [lnav](https://lnav.org) 的日志分析能力 —— Claude Code 一行命令完成日志检索与分析。

## 状态

MVP：`lnav-cli doctor` + `lnav-cli +search` + skills `lnav-shared`, `lnav-search`。
路线图（v1.0）：`+sql`、`+summary`、`+tail`、sources 别名、`lnav-cli setup`。

## 快速开始

```bash
# 1. 安装 lnav
brew install lnav               # macOS
# sudo apt install lnav         # Ubuntu

# 2. 构建 lnav-cli
git clone git@github.com:MonsterChenzhuo/lnav-cli.git
cd lnav-cli && make build

# 3. 自检
./lnav-cli doctor

# 4. 搜索日志
./lnav-cli +search -s /var/log/nginx/error.log --since 1h --level error "timeout"
```

## 给 Claude Code 使用

把 `skills/` 目录下的 skill 放到 Claude Code 的 skills 路径
（通常是 `~/.claude/skills/`），Agent 即可加载 `lnav-shared` 与 `lnav-search`。

## 开发约定

见 [AGENTS.md](./AGENTS.md)。
