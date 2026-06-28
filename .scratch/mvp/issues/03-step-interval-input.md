Status: ready-for-agent

# 03: Step Interval を UI で指定可能に

## Parent

`.scratch/mvp/PRD.md`

## What to build

01 でハードコードされていた Step Interval を、フロントエンドの入力欄から指定できるようにする。

- フロント：数値入力欄（単位は秒、小数許可、デフォルト1.0秒、min=0.1）。
- `app.go`：フロントから渡された秒数を `time.Duration` に変換し CaptureSession に渡す。
- 0や負の値は受け付けない。

## Acceptance criteria

- [ ] UI で Step Interval を変更し、開始すると各 Capture Step 間の待機が指定秒に変わることが体感で確認できる
- [ ] 0や負の値を入れた状態では開始されない
- [ ] CaptureSession の既存テストが通り続ける

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
