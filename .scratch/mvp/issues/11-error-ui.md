Status: ready-for-agent

# 11: エラーUI（赤色表示）

## Parent

`.scratch/mvp/PRD.md`

## What to build

CaptureSession 中に発生したエラーをフロントへ伝え、バー上で目立つ形（赤色）で表示する。エラー時のループ即時中断と部分PDF確定ロジック自体は 01 の CaptureSession で実装済み。本スライスは**通知と UI 表示**のみ。

- Go 側：CaptureSession から返された error / 中断時の最終ステータスを `app.go` で捕捉し、`runtime.EventsEmit("session:error", {message})` をフロントへ送信。
- フロント：`session:error` を受信したらバー上に赤色背景のメッセージブロックを表示。「閉じる」ボタンで消せる。
- エラーメッセージは技術的詳細でなくユーザーが理解できる文章にする（例：「スクリーンキャプチャに失敗しました。Screen Recording 権限が無効になっている可能性があります」「PDF の書き込みに失敗しました。ディスク空き容量を確認してください」）。

## Acceptance criteria

- [ ] CaptureSession 中に Screener/Clicker/PdfWriter のいずれかが error を返すと、ループが即時中断され、バー上に赤色エラー表示が出る
- [ ] 部分PDF（その時点まで）はディスクに保存されており Preview.app で開ける
- [ ] エラーメッセージは原因に応じた人間可読な文章になっている
- [ ] 「閉じる」ボタンでエラー表示を消せる

## Blocked by

- 08: 進捗表示（同イベント機構を使うため）
