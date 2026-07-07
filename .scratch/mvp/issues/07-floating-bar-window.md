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
- フロント：フローティングバーらしいレイアウト（横長または2段組）に再構成。02〜06 で追加された入力ウィジェットを全て1本のバー内に収める。
  - **範囲＆クリック位置選択ボタン 1 つ**（#06 のピボットにより、範囲選択ダイアログ内でクリック位置マーカーも同時に指定するため、ボタンは 1 つで済む）
  - RepeatCount、StepInterval、出力先フォルダ、ファイル名、開始ボタン
  - 範囲指定済みインジケータ、クリック位置指定済みインジケータ（両方 region 確定時に同時にセット）
- バーをドラッグで移動可能にする（CSS `--wails-draggable: drag` または Wails の `runtime.WindowSetPosition` + マウスイベント）。
- **開始ボタンの有効化条件**：CaptureRegion、AdvanceClickPoint、RepeatCount、StepInterval、保存先フォルダ、ファイル名が全て揃った時のみ enabled。実装上は region ↔ clickPoint が同時にセット／クリアされるため、片方チェックで足りるが、意味を明確に保つため両方を条件式に残しておく。

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
