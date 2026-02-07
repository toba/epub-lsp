---
name: upstream-check
description: Check upstream repos for updates. Use when user says "/upstream", "check upstream", "sync with upstream", or wants to see what's changed in the source repos this project was derived from.
---

# Upstream Check

Compare this project against its upstream sources to find valuable updates.

## Upstream Repos

| Source | Repo | What we use |
|--------|------|-------------|
| LSP | `yayolande/go-template-lsp` | LSP server architecture, gota template parser |

## State Tracking

Last checked SHA is stored in `.last-checked` in this skill's directory.

## Workflow

1. Read the last checked SHA (if exists):
```bash
cat "$(dirname "$0")/.last-checked" 2>/dev/null || echo "none"
```

2. Get commits from upstream since last check:
```bash
# If we have a last SHA, get commits since then:
gh api repos/yayolande/go-template-lsp/commits --jq '.[0:20] | .[] | "\(.sha[0:7]) \(.commit.message | split("\n")[0])"'
```
Stop at the last checked SHA. If no state file exists, just show recent 10.

3. For interesting commits, fetch the diff:
```bash
gh api repos/yayolande/go-template-lsp/commits/SHA --jq '.files[] | "\(.filename)\n\(.patch)"'
```

4. After evaluation, update the state file with the latest SHA:
```bash
gh api repos/yayolande/go-template-lsp/commits --jq '.[0].sha' > /Users/jason/Developer/pacer/go-template-lsp/.claude/skills/upstream-check/.last-checked
```

## What to Look For

- Bug fixes in template parsing (gota parser)
- New LSP features (diagnostics, hover, completion, etc.)
- Analyzer improvements

## Output

Summarize findings as:
- Commits worth porting (with rationale)
- Commits to skip (already have, not relevant, etc.)

Then update `.last-checked` with the latest upstream SHA.
