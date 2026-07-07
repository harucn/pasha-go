Status: done

# 07: フローティングバー化（Wails ウィンドウ構成）

## Parent

`.scratch/mvp/PRD.md`

## What to build

ADR-0002 に従い、Wails のメインウィンドウを**コンパクトな常時最前面フローティングバー**として再構成する。

- `main.go` の Wails `options.App`：
  - `AlwaysOnTop: true`
  - サイズを小さく（940×76 に決定。7 個の control が 1 行に収まる横幅、上部 24px は macOS traffic light 用の余白）
  - `Width`/`MinWidth`/`MaxWidth` を同値、`Height`/`MinHeight`/`MaxHeight` を同値にしてバーサイズを固定（ユーザ手動リサイズ不可）
  - `mac.TitleBar: mac.TitleBarHiddenInset()` で閉じる/最小化/フルスクリーンの traffic light を残しつつタイトルバー背景は透過
  - `AlwaysOnTop` と併用。`Frameless: true` は traffic light も消えてしまうため採用しない
- フロント：フローティングバーらしいレイアウト（横長 1 行 + 状態行）に再構成。02〜06 で追加された入力ウィジェットを全て1本のバー内に収める。
  - **範囲＆クリック位置選択ボタン 1 つ**（#06 のピボットにより、範囲選択ダイアログ内でクリック位置マーカーも同時に指定するため、ボタンは 1 つで済む）
  - RepeatCount、StepInterval、出力先フォルダ、ファイル名、開始ボタン
  - 範囲指定済みインジケータ、クリック位置指定済みインジケータ（両方 region 確定時に同時にセット）
- バーをドラッグで移動可能にする。バー本体は `--wails-draggable: drag` + `-webkit-app-region: drag`、入力/ボタン群は `no-drag`。
- 範囲選択に遷移する時は `WindowSetMinSize(200, 150)` / `WindowSetMaxSize(0, 0)` でサイズ制約を緩めてから `WindowSetSize(500, 400)`。復帰時に元のバーサイズへ再ロック。
- **開始ボタンの有効化条件**：CaptureRegion、AdvanceClickPoint、RepeatCount、StepInterval、保存先フォルダ、ファイル名が全て揃った時のみ enabled。実装上は region ↔ clickPoint が同時にセット／クリアされるため、片方チェックで足りるが、意味を明確に保つため両方を条件式に残しておく。

## Acceptance criteria

- [x] アプリ起動時、デスクトップ上に常時最前面の小さなバーが1本だけ表示される
- [x] バーをドラッグして任意の位置に移動できる
- [x] バーは手動リサイズできない（サイズ固定）
- [x] 閉じる / Dock 最小化ボタンは動作する
- [x] 他のアプリ（Preview, Finder 等）を前面化してもバーは隠れない
- [x] 設定が1つでも欠けていると開始ボタンが disabled
- [x] 全て揃うと開始ボタンが enabled になる
- [x] **範囲選択との協調**：`AlwaysOnTop: true` の状態でも、#05 の `beginRegionSelection` によるバー → 大きな透過フレームへのリサイズ、および `restoreWindow` による元のバーサイズ／位置への復元が正しく動作する（HITL 検証済）
- [x] **範囲選択ウィンドウは自由にリサイズ可能**（バーサイズロックが範囲選択時のみ動的に解除される）
- [x] **座標系ドキュメントとの整合**：範囲選択後に取得される Capture Region が Screen Space（`docs/adr/0003-canonical-screen-coordinate-space.md`）で保持されている（HITL 検証済、#04 の multi-display 回帰も確認）

## Notes

- ADR-0002 の「バーのみ UI」思想を維持。ロゴ画像はバーに収まらないため削除。ステータス行は横幅を保ちつつコンパクトに残置。
- Wails のデフォルト CSS drag プロパティは `--wails-draggable`。当初 `-webkit-app-region` を `CSSDragProperty` で上書きしようとしたがバー Frameless 化と併用で挙動が不安定だったため Wails デフォルトに戻し、旧 `.region-frame` にも `--wails-draggable: drag` を併記。
- 実装 PR: feat/floating-bar-window ブランチ。

## Blocked by

- 02, 03, 04, 05, 06（UI 要素が出揃ってからバーレイアウトを最終調整するのが効率的）
