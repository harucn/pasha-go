# pasha-go

[Wails v2](https://wails.io/) + React + TypeScript で作るデスクトップアプリ。

## 必要環境

- Go（`.tool-versions`: `golang 1.26.4`）
- Node.js（`.tool-versions`: `nodejs 26.4.0`）
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

```bash
# asdf を使っているなら
asdf install

# Wails CLI
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

## セットアップ

```bash
# Go の依存
go mod download

# フロントエンドの依存
cd frontend && npm install
```

## 開発（dev server）

リポジトリのルートで:

```bash
wails dev
```

- Wails が Go バックエンドと Vite フロントエンドを同時起動
- フロントは `http://localhost:34115` でブラウザからも開ける（DevTools 利用可）
- `app.go` / `frontend/src/**` 変更でホットリロード

## ビルド

```bash
# プロダクションビルド（build/bin/ に配置）
wails build

# フロントエンドだけビルドしたい場合
cd frontend && npm run build
```

## テスト

### Go

```bash
go test ./...
```

### フロントエンド（Vitest + React Testing Library）

```bash
cd frontend
npm test          # 1 回実行
npm run test:watch  # ウォッチモード
```

Wails が生成する `wailsjs/go/main/App` は `vi.mock` でモックする（例: `src/__tests__/App.test.tsx`）。

## Lint / Format

### Go（golangci-lint）

```bash
golangci-lint run ./...      # 静的解析
golangci-lint fmt ./...      # フォーマット適用
golangci-lint fmt --diff ./... # 差分のみ表示
```

### フロントエンド（Biome）

```bash
cd frontend
npm run check       # 静的解析 + フォーマット差分チェック
npm run check:fix   # 自動修正適用
```

## ディレクトリ構成

```
.
├── app.go              # Wails で公開する Go の API
├── main.go             # エントリポイント
├── frontend/           # React + TypeScript (Vite)
│   ├── src/
│   └── wailsjs/        # wails dev/build が自動生成（コミット対象）
├── build/              # ビルド成果物 / アイコン等
└── docs/               # ADR・ドメイン文書など
```

詳細な開発ガイドラインは [`AGENTS.md`](AGENTS.md) と [`CONTEXT.md`](CONTEXT.md) を参照。

