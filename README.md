# lnav-cli

[English](./README.md) | [中文](./README.zh.md)

Agent-friendly wrapper around [lnav](https://lnav.org) — drive log search & analysis from Claude Code with one-line commands.

## Status

MVP: `lnav-cli doctor` + `lnav-cli +search` + skills `lnav-shared`, `lnav-search`.
Roadmap (v1.0): `+sql`, `+summary`, `+tail`, sources aliases, `lnav-cli setup`.

## Quick Start

```bash
# 1. Install lnav
brew install lnav               # macOS
# sudo apt install lnav         # Ubuntu

# 2. Build lnav-cli
git clone git@github.com:MonsterChenzhuo/lnav-cli.git
cd lnav-cli && make build

# 3. Verify
./lnav-cli doctor

# 4. Search
./lnav-cli +search -s /var/log/nginx/error.log --since 1h --level error "timeout"
```

## For Claude Code

Copy the `skills/` directory into your Claude Code skills path
(`~/.claude/skills/` in typical installs) so the agent can discover
`lnav-shared` and `lnav-search`.
See each skill's `SKILL.md` for the contract.

## Development

See [AGENTS.md](./AGENTS.md) for the pre-PR checklist and conventions.
