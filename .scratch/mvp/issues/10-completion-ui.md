Status: ready-for-agent

# 10: 完了UI（開く / フォルダで表示）

## Parent

`.scratch/mvp/PRD.md`

## What to build

Capture Session が正常完了 or 停止で終了したあと、Output Document をすぐ確認できるアクションをバー上に提供する。

- フロント：完了状態時に「PDF を開く」「フォルダで表示」「リセット」ボタンをバー内に表示。バー横幅の制約から、開始／停止／完了アクションは**排他表示**にする（状態に応じてボタンが差し替わる）。
- `app.go`：
  - `OpenOutputDocument()` メソッド：保存パスを `open` コマンドまたは Wails の `runtime.BrowserOpenURL` で開く（macOS のデフォルト PDF ビューア = Preview.app）。
  - `RevealOutputDocument()` メソッド：保存パスを `open -R <path>` で Finder ハイライト表示。
- 次の撮影に進むための「リセット」ボタンも併設（設定 = Region / ClickPoint をクリアして撮影前状態に戻す。#06 のピボットにより両者は同時に管理されるため、実装上は一括クリアでよい）。

## Acceptance criteria

- [ ] 撮影完了 or 停止後、「PDF を開く」「フォルダで表示」「リセット」ボタンが表示される
- [ ] 「PDF を開く」で Preview.app が立ち上がり生成 PDF が表示される
- [ ] 「フォルダで表示」で Finder が立ち上がり該当 PDF がハイライトされる
- [ ] 「リセット」で Region / ClickPoint がクリアされ、範囲選択ダイアログを再度実行しないと開始できない状態になる

## Blocked by

- 08: 進捗表示（同イベント機構を使うため）
