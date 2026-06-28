Status: ready-for-agent

# 01: Tracer bullet — 最小e2e Capture Session

## Parent

`.scratch/mvp/PRD.md`

## What to build

PRD の中核である Capture Session のループを、**ハードコード設定で1ボタンから動かせる最小e2e**を構築する。これにより以降のスライスは「UIを足す」「入力経路を足す」だけになる。

実装する内容：

- `internal/session` パッケージ：`CaptureSession` 型を定義し、コンストラクタで CaptureRegion / AdvanceClickPoint / RepeatCount / StepInterval / 出力パス / コラボレータ群（Screener, Clicker, PdfWriter, Clock の各インターフェース）を受け取り、`Start(ctx) error` と `Stop()` を提供する。インターフェースは本パッケージ内で宣言する（依存性逆転）。
- `internal/screener` パッケージ：macOS で `CaptureRegion` の `image.Image` を返す実装。`kbinani/screenshot` か `robotgo` を選定。
- `internal/clicker` パッケージ：macOS で `AdvanceClickPoint` 座標に左クリックを送る実装。`robotgo` または CGo 経由の `CGEventPost`。
- `internal/pdfwriter` パッケージ：ADR-0001 に従う**逐次追記方式**。`AppendPage(image.Image) error` と `Close() error`。`signintech/gopdf` を利用予定。
- `app.go`：仮の `RunTestSession()` メソッドを追加し、上記を組み合わせてハードコード値（例：画面全体・画面中央クリック・RepeatCount=3・StepInterval=1秒・出力 `~/Desktop/pasha-tracer.pdf`）でセッションを実行する。
- フロントエンド：既存テンプレートのボタンを「テスト撮影」ボタンに置き換え、`RunTestSession` を呼ぶ。

テスト（fakes-only、`internal/session/session_test.go`）：

1. 正常系 N=5 回ループ、Screener/Clicker/PdfWriter.AppendPage が各5回呼ばれ、Close が最後に1回
2. 1ステップ内で Screener → AppendPage → Clicker → Sleep の順序
3. Step Interval と等しい duration で Clock.Sleep が呼ばれる
4. 別ゴルーチンから Stop() を呼ぶと現ステップ完了後に終了 + Close 呼ばれる
5. Screener が error を返すと即時中断 + Close + Start が error 返却
6. Clicker が error を返すと即時中断（同上）
7. PdfWriter.AppendPage が error を返すと即時中断（同上）
8. context キャンセルで中断 + Close

`internal/session/fakes_test.go` に Screener/Clicker/PdfWriter/Clock のフェイク。

## Acceptance criteria

- [ ] `internal/session` の8シナリオ Go テストが全て通る
- [ ] `wails dev` でアプリ起動 → 「テスト撮影」ボタンを押すと、3秒程度後に `~/Desktop/pasha-tracer.pdf` が生成される
- [ ] 生成された PDF を Preview.app で開くと3ページあり、各ページに画面のスクリーンショットが入っている
- [ ] 撮影中に画面中央でマウスカーソルが3回クリックされる（観察可能）
- [ ] `internal/session` パッケージが `internal/screener` 等の具体実装パッケージを import していない
- [ ] `CONTEXT.md` の用語（CaptureSession, CaptureRegion, AdvanceClickPoint, CaptureStep, RepeatCount, StepInterval, OutputDocument）が型名・関数名で一貫使用されている

## Blocked by

None — can start immediately.
