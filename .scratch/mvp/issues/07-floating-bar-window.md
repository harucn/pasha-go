Status: ready-for-agent

# 07: フローティングバー化（Wails ウィンドウ構成）

## Parent

`.scratch/mvp/PRD.md`

## What to build

ADR-0002 に従い、Wails のメインウィンドウを**コンパクトな常時最前面フローティングバー**として再構成する。

- `main.go` の Wails `options.App`：
  - `Frameless: true`
  - `AlwaysOnTop: true`
  - サイズを小さく（例：480×80、コンテンツに応じて要調整）
  - ウィンドウタイトル不要
- フロント：フローティングバーらしいレイアウト（横長または2段組）に再構成。02〜06 で追加された入力ウィジェット（Region/ClickPoint ボタン、Count、Interval、出力先、ファイル名、開始ボタン）を全て1本のバー内に収める。
- バーをドラッグで移動可能にする（CSS `--wails-draggable: drag` または Wails の `runtime.WindowSetPosition` + マウスイベント）。
- **開始ボタンの有効化条件**：CaptureRegion、AdvanceClickPoint、RepeatCount、StepInterval、保存先フォルダ、ファイル名が全て揃った時のみ enabled。

## Acceptance criteria

- [ ] アプリ起動時、デスクトップ上に枠なし・常時最前面の小さなバーが1本だけ表示される
- [ ] バーをドラッグして任意の位置に移動できる
- [ ] 他のアプリ（Preview, Finder 等）を前面化してもバーは隠れない
- [ ] 設定が1つでも欠けていると開始ボタンが disabled
- [ ] 全て揃うと開始ボタンが enabled になる
- [ ] **範囲選択との協調**：`Frameless: true` + `AlwaysOnTop: true` の状態でも、#05 の `beginRegionSelection` によるバー → 大きな透過フレームへのリサイズ、および `restoreWindow` による元のバーサイズ／位置への復元が正しく動作する（HITL 検証）
- [ ] **座標系ドキュメントとの整合**：範囲選択後に取得される Capture Region が Screen Space（`docs/adr/0003-canonical-screen-coordinate-space.md`）で保持されている（既に #05 で担保されているが、フローティングバー化後も回帰していないことを確認）

## Blocked by

- 02, 03, 04, 05, 06（UI 要素が出揃ってからバーレイアウトを最終調整するのが効率的）
