# 座標系は「primary top-left global points」に一本化する

pasha-go でユーザー空間の座標を扱うすべてのモジュールは、**プライマリディスプレイの左上を原点 (0, 0) とし、x 軸を右、y 軸を下向き、単位は logical points（物理 pixel ではない）** の座標系を唯一の共通言語として使う。この空間を **Screen Space** と呼ぶ。

- `Capture Region`（画面上の矩形）と `Advance Click Point`（画面上の点）はどちらも Screen Space の値として扱う。
- Wails ↔ Go ↔ フロント（TS/React）の境界を跨ぐ座標も Screen Space。フロント側で `devicePixelRatio` を掛ける／割る等の変換は **禁止**。
- 複数ディスプレイ配置（右／左／上／下）に伴う負値やプライマリ幅を超える値は正当な Screen Space の値であり、無条件に許容する。

## この座標系を選ぶ理由

- キャプチャ実装（`internal/screener` → `kbinani/screenshot.Capture`）と自動クリック実装（`internal/clicker` → `robotgo.Move` on macOS `CGEventPost`）が **共に** この空間を要求するため、外側のあらゆる座標をここに集約すれば OS 依存の変換が 1 箇所で済む。
- macOS ネイティブの `NSScreen`（bottom-left global）や Wails の `runtime.WindowGetPosition`（現在のディスプレイのローカル）と異なる空間を採用する誘惑があるが、いずれも下流のキャプチャ／クリック層と噛み合わず、multi-display で必ず破綻する（本 ADR 制定のきっかけとなった bug: マルチディスプレイでプライマリしか撮れない現象）。

## 境界の変換責務

OS 由来の座標を Screen Space に変換する処理は **`internal/appwindow` パッケージに閉じ込める**。

- macOS: `appwindow.GetMainWindowRect()` が cgo 経由で `NSWindow.frame` と `CGDisplayBounds(CGMainDisplayID())` を読み、pure 関数 `nsScreenToKbinani` で y 軸反転して Screen Space に変換する。
- 将来 Windows / Linux 対応時も、同パッケージ内に platform-specific ファイルを追加し、外への API は `Screen Space の image.Rectangle を返す` 契約を保つ。

これにより、`app.go` 以降の全レイヤーは「Screen Space の値しか扱わない」前提を安全に置ける。

## Considered Options

- **NSScreen 座標をそのまま使う（bottom-left global, points）**: Cocoa API と直接繋がるが、y 軸方向が下流の実装（kbinani, robotgo, PDF ライブラリ等）と全て逆で境界の変換コストが分散する
- **物理 pixel 空間に統一（DPR 適用済み）**: Retina 対応が明示的になるが、kbinani/robotgo が points ベースのためどの層でも DPR を意識することになり、二重掛けによる bug が起きやすい
- **primary top-left global points（採用）**: キャプチャ／クリックの両ライブラリが要求する空間そのものなので、外側の変換を一箇所に閉じ込められる
