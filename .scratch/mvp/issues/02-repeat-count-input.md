Status: ready-for-agent

# 02: Repeat Count を UI で指定可能に

## Parent

`.scratch/mvp/PRD.md`

## What to build

01 でハードコードされていた Repeat Count を、フロントエンドの入力欄から指定できるようにする。

- フロント：数値入力欄（min=1, デフォルト値10程度）を追加。
- `app.go`：`RunTestSession` を改修し、フロントから渡された Repeat Count を CaptureSession コンストラクタに渡す。
- 入力欄が空・無効値（0以下、非整数）の場合は開始ボタンを押せない（または送信されない）。

## Acceptance criteria

- [ ] UI で Repeat Count を変更し、開始すると指定した枚数のページが PDF に生成される
- [ ] 0や負の値を入れた状態では開始されない
- [ ] CaptureSession のテストには影響しない（既存テストが通り続ける）

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
