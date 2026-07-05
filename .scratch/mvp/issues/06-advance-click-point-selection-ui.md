Status: ready-for-agent

# 06: Advance Click Point 選択 UI

## Parent

`.scratch/mvp/PRD.md`

## What to build

01 でハードコードされていた Advance Click Point を、視覚的に画面上から1点だけ指定できるようにする。

### 前提：#05 のピボットを踏襲する

#05 では当初「全画面半透明オーバーレイでドラッグ選択」実装を書いたが、UX 上の問題（別ウィンドウ内しか選べない・選択箇所が見えない等）から捨て、**「透過ウィンドウ自体を範囲として使う」方式**にピボットした（コミット `cada8d8`）。#06 も同じ思想で設計する。

### 実装オプション（実装前に選ぶ）

以下 3 案から選択する。**A を推奨**。

**A. window-as-cursor（推奨）**
- 「クリック位置選択」ボタン押下で、透過な小さいウィンドウ（例：40×40 pt、十字マーク表示）に切り替わる。
- ユーザーはそのウィンドウをドラッグして目的の点へ移動 →「確定」ボタンで確定。
- 確定位置 = ウィンドウの中心（Screen Space の点、`internal/appwindow` で取得）。
- **利点**：#05 の `beginRegionSelection` / `restoreWindow` と実装機構を共有できる。座標系変換も同じパスに乗る。

**B. オーバーレイ復活**
- 全画面透過ウィンドウを新規に開き、その上でクリックを受け付ける。
- **利点**：直感的で「クリックしたら決定」の1ステップで済む。
- **欠点**：#05 で捨てた実装を再導入することになり、実装コストが増える。透過ウィンドウの座標系は Wails の別ウィンドウ扱いになり、`internal/appwindow` の拡張が必要。

**C. Skip：region 中心固定**
- 独立した Click Point 選択 UI を作らず、`Capture Region` の中心を Advance Click Point として使う（現状の実装のまま）。
- **利点**：追加実装ゼロで既に動いている。
- **欠点**：AC「1点をユーザーが指定できる」を諦める（PRD 上の妥協）。

### 座標系

- Advance Click Point は **Screen Space**（プライマリ左上原点、logical points）で扱う。詳細は `docs/adr/0003-canonical-screen-coordinate-space.md`。
- 案 A を採る場合、`internal/appwindow.GetMainWindowRect()` の返す矩形の中心を Point として使えばよい。新たな座標系変換は不要。
- `internal/clicker` の `robotgo.Move(x, y)` は macOS では `CGEventPost` を経由し、同じ Screen Space（top-left global points）を要求するため、Screen Space の値をそのまま渡してよい。

## Acceptance criteria

- [ ] 「クリック位置選択」ボタンを押すと（A 案）バーが小さな透過ターゲットに切り替わる／（B 案）全画面透過オーバーレイが出る／（C 案）該当なし
- [ ] （A/B 案）確定操作でバー上に「クリック位置指定済み」表示が出る
- [ ] （A/B 案）Esc でキャンセルすると元の状態に戻る
- [ ] 確定した座標で開始すると、各ステップでその座標がクリックされる
- [ ] クリック位置未指定の状態では開始ボタンが押せない（C 案では「Region 選択済み」を代替条件にする）
- [ ] 再選択で上書きできる（A/B 案）
- [ ] **マルチディスプレイ確認**：セカンダリディスプレイ上でクリック位置を指定 → 実際にそのディスプレイの当該位置がクリックされることを HITL で確認（robotgo の multi-display 挙動を実機で検証）

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
- 05: Capture Region 選択 UI（実装機構を共有するため）

## Design notes

- 案 A を選んだ場合、`internal/appwindow` に点変換ヘルパー（例：`GetMainWindowCenterPoint() image.Point`）を追加すると呼び出し側がシンプル。既存の `GetMainWindowRect()` を呼んで中心を計算するのでも十分。
- robotgo の macOS 実装は歴史的にマルチディスプレイでの挙動が版によって差異があるため、AC の HITL 検証項目を必ず通す。
