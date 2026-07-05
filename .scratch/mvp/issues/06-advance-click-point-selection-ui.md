Status: ready-for-agent

# 06: Advance Click Point 選択 UI

## Parent

`.scratch/mvp/PRD.md`

## What to build

01 でハードコードされていた Advance Click Point を、視覚的に画面上から1点だけ指定できるようにする。

### 前提：#05 のピボットを踏襲する

#05 では当初「全画面半透明オーバーレイでドラッグ選択」実装を書いたが、UX 上の問題（別ウィンドウ内しか選べない・選択箇所が見えない等）から捨て、**「透過ウィンドウ自体を範囲として使う」方式**にピボットした（コミット `cada8d8`）。#06 も同じ思想で設計する。

### 採用 UX: In-frame draggable marker（範囲選択と一括指定）

実装 UX は当初 A/B/C の 3 案を検討したが、実装中のフィードバックで **D. 範囲選択の透過フレーム内にドラッグ可能なマーカーを置く方式** に確定した（`.copilot/session-state/1481c6bf-.../checkpoints/` 参照）。

- 「範囲選択」ボタンで従来通り透過フレーム（デフォルト 500×400 pt）に切り替わる。
- **フレーム内にドラッグ可能なマーカー（円形＋十字）** が表示される。初期位置はフレーム中央。
- ユーザーはフレームを移動・リサイズして範囲を合わせつつ、マーカーをドラッグしてクリック位置を決める。
- 「確定」で **範囲とクリック位置を同時にコミット** する。フロントで
  - `region = GetSelectedRegion()`（既存 cgo helper）
  - `clickPoint = { x: region.x + markerX, y: region.y + markerY }`
    （フレームは Frameless で 100vw/100vh・macOS では CSS px = points なので単純加算）
- 別モードのクリック位置選択ボタン・ダイアログは廃止。

### 座標系

- Advance Click Point は **Screen Space**（プライマリ左上原点、logical points）で扱う。詳細は `docs/adr/0003-canonical-screen-coordinate-space.md`。
- `internal/appwindow.GetMainWindowRect()` が Screen Space の region を返し、フロントで markerPos を加算するだけで clickPoint が得られる。新たな座標系変換は不要。
- `internal/clicker` の `robotgo.Move(x, y)` は macOS では `CGEventPost` を経由し、同じ Screen Space（top-left global points）を要求するため、Screen Space の値をそのまま渡してよい。

## Acceptance criteria

- [x] 範囲選択ダイアログ内にクリック位置マーカーが表示される
- [x] マーカーをドラッグしてフレーム内で移動できる
- [x] 確定操作で「範囲指定済み」と「クリック位置指定済み (x,y)」の両方が表示される
- [x] Esc / キャンセルで元の状態に戻り、範囲もクリック位置も保存されない
- [x] 範囲もしくはクリック位置が未指定の状態では開始ボタンが押せない（実際は同時にセットされる）
- [x] 再選択で上書きできる（マーカーは毎回中央にリセット）
- [ ] **マルチディスプレイ確認**：セカンダリディスプレイ上でクリック位置を指定 → 実際にそのディスプレイの当該位置がクリックされることを HITL で確認（robotgo の multi-display 挙動を実機で検証）

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
- 05: Capture Region 選択 UI（実装機構を共有するため）

## Design notes

- Go 側は既存の `GetSelectedRegion` のみを使う。当初検討した `GetSelectedClickPoint` は不要になった。
- マーカー位置はフロントの React state（CSS px = points on macOS）。ドラッグは pointer capture + delta 方式で jsdom テストも動く。
- robotgo の macOS 実装は歴史的にマルチディスプレイでの挙動が版によって差異があるため、AC の HITL 検証項目を必ず通す。
