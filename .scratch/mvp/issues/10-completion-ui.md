Status: ready-for-agent

# 10: 完了UI（開く / フォルダで表示）

## Parent

`.scratch/mvp/PRD.md`

## What to build

Capture Session が正常完了 or 停止で終了したあと、Output Document をすぐ確認できるアクションをバー上に提供する。

- フロント：完了状態時に「PDF を開く」「フォルダで表示」ボタンを表示。
- `app.go`：
  - `OpenOutputDocument()` メソッド：保存パスを `open` コマンドまたは Wails の `runtime.BrowserOpenURL` で開く（macOS のデフォルト PDF ビューア = Preview.app）。
  - `RevealOutputDocument()` メソッド：保存パスを `open -R <path>` で Finder ハイライト表示。
- 次の撮影に進むための「リセット」ボタンも併設（設定 = Region / ClickPoint をクリアして撮影前状態に戻す）。

## Acceptance criteria

- [ ] 撮影完了 or 停止後、「PDF を開く」「フォルダで表示」「リセット」ボタンが表示される
- [ ] 「PDF を開く」で Preview.app が立ち上がり生成 PDF が表示される
- [ ] 「フォルダで表示」で Finder が立ち上がり該当 PDF がハイライトされる
- [ ] 「リセット」で Region / ClickPoint がクリアされ、再度それらを選択しないと開始できない状態になる

## Blocked by

- 08: 進捗表示（同イベント機構を使うため）
