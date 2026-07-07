Status: ready-for-agent

# 08: 進捗表示

## Parent

`.scratch/mvp/PRD.md`

## What to build

CaptureSession が現在何ステップ目を実行中かを、Go→フロントへリアルタイム通知し、バー上に表示する。

- Go 側：CaptureSession に「ステップ完了通知」のフックを追加（コンストラクタで progress コールバックを受け取る、または `internal/session` で event 型を返すチャネルを公開）。`app.go` がこれを購読し、Wails の `runtime.EventsEmit("session:progress", {current, total})` でフロントへ送信。
- フロント：バー上部のステータス行（`.floating-bar .result`）に「N / M ステップ完了」を表示。バー横幅（940px）に余裕がないため、進捗バーではなくテキストで開始。`runtime.EventsOn("session:progress", ...)` で受信。
- CaptureSession のテストには progress フックのテストも追加（呼び出し回数とパラメータの順序）。

## Acceptance criteria

- [ ] 撮影開始後、バー上に「N / M」または進捗が動的に更新される
- [ ] 完了時に N === M を表示
- [ ] CaptureSession のテストで progress フックが各ステップ完了時に呼ばれることを検証

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
