Status: ready-for-agent

# 12: macOS 権限プリフライト

## Parent

`.scratch/mvp/PRD.md`

## What to build

アプリ起動時に macOS の Screen Recording 権限と Accessibility 権限の付与状況を検知し、不足があれば誘導ダイアログを表示する。

- `internal/permissions` パッケージ：
  - `HasScreenRecording() bool`：CGo 経由で `CGPreflightScreenCaptureAccess()` を呼ぶ。
  - `HasAccessibility() bool`：CGo 経由で `AXIsProcessTrustedWithOptions()` を呼ぶ。
  - `OpenSettings(kind)`：それぞれの権限ペインの URL を `open` で開く（例：`x-apple.systempreferences:com.apple.preference.security?Privacy_ScreenCapture`）。
- `app.go`：`startup` で両権限をチェック、不足があれば `runtime.EventsEmit("permission:missing", {kinds: ["screen-recording", "accessibility"]})` をフロントへ送信。
- フロント：受信時にダイアログ（モーダル）を表示。各権限ごとに「System Settings を開く」ボタン、「権限再チェック」ボタン。
- 再チェックは Go 側に都度問い合わせ、両方 OK になったらダイアログを閉じる。

## Acceptance criteria

- [ ] 起動時に両権限が揃っていれば通常通り起動（ダイアログ出ない）
- [ ] どちらか欠けていると起動直後にダイアログが出る
- [ ] 「System Settings を開く」で対応する権限ペインが表示される
- [ ] 権限を付与した後「再チェック」を押すとダイアログが閉じる
- [ ] 権限なしの状態でも、ダイアログは閉じずに撮影開始が物理的に不可能な状態が保たれる

## Blocked by

None — 完全独立、1 と並行可。
