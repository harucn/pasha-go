# AGENTS.md

## Lint & Format

**Go**: `goimports`（formatter）+ `golangci-lint`（linter）

```bash
# 静的解析
golangci-lint run ./...

# フォーマット適用
golangci-lint fmt ./...

# フォーマット差分のみ表示
golangci-lint fmt --diff ./...
```

設定は `.golangci.yml`。

**TypeScript / React**: `biome`（formatter + linter）

```bash
cd frontend

# 静的解析 + フォーマット差分チェック
npm run check

# 自動修正適用
npm run check:fix
```

設定は `frontend/biome.json`。コミット前に上記が通ることを確認。

## Agent skills

### Issue tracker

Issues live as markdown files under `.scratch/<feature>/` in this repo. See `docs/agents/issue-tracker.md`.

### Triage labels

Uses the default canonical vocabulary (`needs-triage`, `needs-info`, `ready-for-agent`, `ready-for-human`, `wontfix`). See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at the repo root. See `docs/agents/domain.md`.

