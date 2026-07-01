# AGENTS.md

## 全般ルール

### 迷ったら別のエージェントに聞く

ライブラリ選定・設計判断・実装方針・エラーメッセージの解釈・ツール選定など、**自信が持てない時は独断で進めずに別のサブエージェントへ相談する**。Copilot CLI なら以下のいずれか：

- `rubber-duck` agent — 計画・実装に対する高シグナルなフィードバック
- `general-purpose` agent — 別コンテキストで深く調査・提案
- `research` agent — 複数ソースを横断調査して根拠付きで報告
- **別 CLI への shell out**（例：`claude -p "..."` で Claude CLI に問う）— Copilot CLI 内のサブエージェントが同じ思考バイアスで詰まっている時、別ツール由来のセカンドオピニオンとして特に有効

モデルは状況に応じて選ぶ（`task` tool の `model` パラメータで Claude / GPT / Gemini など切替可能）。「とりあえず書いて動けばOK」で済ませず、判断に迷ったら手を止めて聞く。オーナーは Go や周辺ツールの専門家ではないので、エージェント側が技術的不確実性を解消する責任を負う。

---

## Go 開発時のルール

エージェントが「もっともらしいが古い／誤った Go の知識」で実装するリスクを構造的に防ぐため、以下のルールに従うこと。

### 1. 非自明な判断は必ず1次情報を参照する

Go の標準ライブラリ・サードパーティライブラリ・イディオムについて、自信がない判断は推測で進めず、以下を **web fetch / GitHub 検索で取りに行く** こと：

| 判断種別 | 参照先 |
|---|---|
| 標準ライブラリの使い方 | `pkg.go.dev/<package>` |
| サードパーティライブラリの選定・使い方 | `pkg.go.dev` + GitHub README（star数・最終更新・open issue 数を確認） |
| イディオム・パターン | `go.dev/blog`、Effective Go、[Google Go Style Guide](https://google.github.io/styleguide/go/) |
| 実用例 | GitHub `search_code` で高 star リポジトリの実装を確認 |
| バージョン固有挙動 | Go release notes（`go.dev/doc/devel/release`） |

「自分の学習データには〜と書かれている」では不十分。**現時点の公式情報で裏取り**してから実装する。

### 2. 依存追加時のゲート

`go.mod` に新しいライブラリを追加する時は、以下を確認しコミットメッセージ or ADR に記録：

- ライセンス（MIT / Apache-2.0 / BSD など、商用利用に問題ないか）
- 最終リリース日（1年以上更新なしなら代替を検討、もしくは採用理由を明記）
- Star 数・コミュニティ活発度・代替候補との比較
- メンテナの活動性（直近の commit / issue 対応）

### 3. レビュー段階でセカンドオピニオン

実装が一定量まとまったタイミング（issue 1つ分くらい）で、Copilot CLI の `code-review` サブエージェントを別コンテキストで走らせ、Go ベストプラクティス観点の指摘を受ける。指摘に対しては「対応」「却下＋理由」のいずれかを記録する。

### 4. オーナーへの解説責任

オーナーは Go（および標準ライブラリ・エコシステム）の専門家ではない。エージェントは実装するだけでなく、**オーナーがコードを読めるようになる**ところまで責任を持つ。

- Go 実装（コミット・PR）を提示した後、以下のいずれかに該当する要素があれば、実装ターンの末尾で日本語で簡潔に補足する：
  - **標準ライブラリの用途**（`fmt`, `os`, `context`, `errors`, `sync`, `io`, `time` などの、そのファイルで使われている package が何をしているか）
  - **Go 固有のイディオム**（early return / guard clause、error wrapping の `%w`、`defer`、`context.Context` を第1引数で受け取る慣習、table-driven test、interface の依存性逆転、goroutine と channel、`sync.Mutex` の zero-value 使用など）
  - **型・シンタックスの非自明な点**（`iota`、埋め込み、レシーバの pointer/value 使い分け、named return values、`any`/`interface{}` の互換性など）
  - **`go.mod` に加わったサードパーティ package の役割**
- 「毎行コメント」ではなく、**新しく登場した要素だけ**を要点3〜7個程度で。既に前セッションで説明済みの内容は繰り返さなくてよい。
- 質問されなくても先に補足する（オーナーが「何を知らないかを知らない」ケースをカバーするため）。

---

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

