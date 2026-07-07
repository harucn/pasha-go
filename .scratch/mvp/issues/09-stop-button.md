Status: ready-for-agent

# 09: 停止ボタン

## Parent

`.scratch/mvp/PRD.md`

## What to build

撮影中にユーザーが任意のタイミングで Capture Session を中断できる「停止」ボタンを実装する。

- フロント：撮影中は「テスト撮影」ボタンを「停止」ボタンに差し替える（バーの横幅に余裕がないため、開始と停止は排他表示にする）。
- `app.go`：ボタンから呼べる `StopSession()` メソッドを追加し、内部で保持している CaptureSession の `Stop()` を呼ぶ。
- Q3 の方針に従い、**現在実行中の Capture Step を最後まで終えてから**ループを終了。途中の不完全な状態にはしない。
- 停止後、Output Document は確定保存される（CaptureSession の Close 呼び出しは 01 で実装済み）。
- 完了通知（`session:completed` イベント、08 のチャネルと同型）をフロントへ送信し、UI は撮影終了状態に遷移する。

## Acceptance criteria

- [ ] 撮影中に停止ボタンを押すと、現在の Capture Step が完了してからループが止まる
- [ ] 停止時点までの PDF が正しく保存され、Preview.app で開ける
- [ ] 停止後、バーは「撮影終了」状態を示す
- [ ] 撮影していない時は停止ボタンが disabled

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
- 08: 進捗表示
