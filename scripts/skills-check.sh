#!/usr/bin/env bash
set -euo pipefail

fail=0
shopt -s nullglob
for skill in skills/*/SKILL.md; do
  for field in '^name:' '^version:' '^description:'; do
    if ! head -20 "$skill" | grep -q "$field"; then
      echo "[skills-check] $skill missing field matching $field" >&2
      fail=1
    fi
  done
  dir="$(dirname "$skill")"
  while IFS= read -r ref; do
    target="$dir/$ref"
    if [[ ! -f "$target" ]]; then
      echo "[skills-check] $skill references missing file: $target" >&2
      fail=1
    fi
  done < <(grep -oE 'references/[A-Za-z0-9_.-]+\.md' "$skill" | sort -u)
done
exit "$fail"
