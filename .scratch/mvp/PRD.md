Status: ready-for-agent

# PRD: pasha-go MVP — Capture Session 基盤

## Problem Statement

紙資料・PDFビューア・プレゼン資料などを1ページずつデジタル化したい場面で、ユーザーは「画面の一部をスクショ → ビューアの『次へ』をクリック → 再びスクショ…」を**100ページ以上に渡って手動で繰り返す**ことが多い。手作業は単調で時間がかかり、注意が逸れるとページ抜けや重複が発生する。既存の汎用スクリーンキャプチャツールはこの「キャプチャ→クリック→待機」のループを自動化してくれない。

## Solution

pasha-go はデスクトップ上にフローティングバー1本だけを表示する macOS アプリで、ユーザーが事前に **Capture Region**（取得する画面矩形）と **Advance Click Point**（次ページへ進めるクリック位置）を指定し、**Repeat Count** と **Step Interval** を入れて「開始」を押すだけで、自動的に **Capture Step**（キャプチャ → クリック → 待機）を N 回繰り返す。各 Capture Step の画像は **Output Document**（マルチページPDF）に逐次追記され、撮影中・撮影後にいつでもそのPDFを開ける。

## User Stories

1. As a pasha-go ユーザー, I want アプリ起動時に macOS の Screen Recording 権限と Accessibility 権限の付与状況が自動でチェックされる, so that 権限不足で撮影に失敗してから慌てずに済む。
2. As a pasha-go ユーザー, I want 必要な権限が不足している場合、「System Settings を開く」ボタン付きのダイアログが表示される, so that どこで何を許可すればよいか自分で調べなくて済む。
3. As a pasha-go ユーザー, I want アプリ起動後、画面上に小さなフローティングバーが1本だけ表示される, so that デスクトップの大半を覆われずに対象アプリを見ながら作業できる。
4. As a pasha-go ユーザー, I want フローティングバーをドラッグで好きな位置に動かせる, so that ターゲットアプリの邪魔にならない場所にどけられる。
5. As a pasha-go ユーザー, I want フローティングバーが常に他のウィンドウより前面に表示される, so that 撮影中に対象アプリの裏に隠れてしまわない。
6. As a pasha-go ユーザー, I want フローティングバーの「範囲選択」ボタンを押すと、画面全体に半透明のオーバーレイが出る, so that マウスドラッグで Capture Region を視覚的に指定できる。
7. As a pasha-go ユーザー, I want 範囲選択中にドラッグした矩形が半透明にハイライトされる, so that 確定前に範囲が正しいか確認できる。
8. As a pasha-go ユーザー, I want 範囲選択をキャンセル（Esc キー等）できる, so that 間違えてボタンを押してしまっても元の状態に戻れる。
9. As a pasha-go ユーザー, I want 範囲選択確定後、フローティングバーに「範囲指定済み」のような状態表示が出る, so that 設定が完了したことが分かる。
10. As a pasha-go ユーザー, I want 範囲選択を再実行して上書きできる, so that 間違った範囲を選んでもやり直せる。
11. As a pasha-go ユーザー, I want フローティングバーの「クリック位置選択」ボタンを押すと、画面全体に半透明のオーバーレイが出る, so that 1点クリックで Advance Click Point を視覚的に指定できる。
12. As a pasha-go ユーザー, I want クリック位置選択をキャンセルできる, so that 間違えてボタンを押してしまっても元の状態に戻れる。
13. As a pasha-go ユーザー, I want クリック位置選択確定後、フローティングバーに「クリック位置指定済み」のような状態表示が出る, so that 設定が完了したことが分かる。
14. As a pasha-go ユーザー, I want クリック位置選択を再実行して上書きできる, so that 違う位置に変更したい場合にやり直せる。
15. As a pasha-go ユーザー, I want フローティングバーで Repeat Count（数値）を入力できる, so that 何ページ撮影するか指定できる。
16. As a pasha-go ユーザー, I want フローティングバーで Step Interval（秒数）を入力できる, so that ターゲットアプリのページ送り処理にかかる時間に応じた待機時間を指定できる。
17. As a pasha-go ユーザー, I want フローティングバーで Output Document の保存先フォルダを選択できる, so that 任意の場所に PDF を保存できる。
18. As a pasha-go ユーザー, I want Output Document のファイル名がデフォルトでタイムスタンプ（例: `pasha-2026-06-28_15-30.pdf`）になっている, so that 何も入力しなくても撮影を開始できる。
19. As a pasha-go ユーザー, I want Output Document のファイル名を手動で編集できる, so that 用途に応じた名前を付けられる。
20. As a pasha-go ユーザー, I want 保存先フォルダに同名ファイルがあった場合、自動で `-2`, `-3` などの連番が付与される, so that 撮影中に確認ダイアログで止められずに済む。
21. As a pasha-go ユーザー, I want Capture Region・Advance Click Point・Repeat Count・Step Interval・保存先が全て揃って初めて「開始」ボタンが有効になる, so that 設定漏れで撮影が失敗するのを防げる。
22. As a pasha-go ユーザー, I want 「開始」を押すと、その瞬間から Capture Session が始まる, so that 余計なカウントダウンを待たされない（ターゲットアプリの前面化はバーが常時表示なので事前に済ませられる）。
23. As a pasha-go ユーザー, I want Capture Session 実行中、フローティングバーに「N / Repeat Count ステップ完了」の進捗が表示される, so that あとどれくらいで終わるか分かる。
24. As a pasha-go ユーザー, I want Capture Session 実行中、フローティングバーの「停止」ボタンがいつでも押せる状態になっている, so that 中断したくなったらすぐ止められる。
25. As a pasha-go ユーザー, I want 「停止」ボタンを押すと、現在実行中の Capture Step を最後まで終えた後にループが止まる, so that 中断時の PDF が中途半端な状態にならない。
26. As a pasha-go ユーザー, I want 「停止」した時点までの Output Document が確定保存される, so that 途中まで取れた分は無駄にならない。
27. As a pasha-go ユーザー, I want Capture Session 中にエラー（スクショ失敗、クリック失敗、PDF書き込み失敗、ディスプレイ構成変化等）が起きたら、ループが即時中断される, so that 不完全な状態で撮り続けて欠落ページのある PDF を生む事故を避けられる。
28. As a pasha-go ユーザー, I want エラー中断時、その時点までの Output Document が部分PDFとして保存される, so that 失敗してもそこまでの成果が残る。
29. As a pasha-go ユーザー, I want エラー中断時、フローティングバー上に分かりやすいエラーメッセージが赤色で表示される, so that 何が起きたか即座に分かる。
30. As a pasha-go ユーザー, I want Repeat Count に到達して Capture Session が正常完了した時、フローティングバーに完了表示が出る, so that 撮影が成功したことが分かる。
31. As a pasha-go ユーザー, I want 完了後、Output Document を「開く」または「フォルダで表示」できるボタンがフローティングバーに出る, so that 完成PDF をすぐ確認できる。
32. As a pasha-go ユーザー, I want Capture Session 完了/中断後、設定（Region・ClickPoint等）をクリアして次の撮影に進める, so that 連続して別の資料を撮りたい時にスムーズに移れる。
33. As a pasha-go ユーザー, I want アプリ終了時に未完了の Capture Session があれば、その時点までの Output Document が保存された状態でクローズされる, so that 想定外の終了でも成果がロストしない。

## Implementation Decisions

### モジュール構成（Go側）

- **`internal/session` パッケージ — `CaptureSession` 型**: 本機能の中核。Repeat Count・Step Interval・CaptureRegion・AdvanceClickPoint・出力パスを受け取り、ループを実行する。コラボレータ（Screener / Clicker / PdfWriter / Clock）は全てインターフェースとして注入され、本パッケージは OS API に直接依存しない。`Start(ctx context.Context) error` と `Stop()` を提供する。**唯一のテスト seam**。
- **`internal/screener` パッケージ**: macOS 画面キャプチャの実装。`Screener` インターフェースを満たす。`robotgo` または `kbinani/screenshot` を利用予定。
- **`internal/clicker` パッケージ**: macOS 合成クリックの実装。`Clicker` インターフェースを満たす。`robotgo` または CGo 経由の `CGEventPost` を利用予定。
- **`internal/pdfwriter` パッケージ**: PDF逐次追記の実装。`PdfWriter` インターフェース（`AppendPage(image.Image) error` / `Close() error`）を満たす。`signintech/gopdf` を利用予定。ADR-0001 の方針に従う。
- **`internal/permissions` パッケージ**: macOS の Screen Recording と Accessibility 権限の検知。`HasScreenRecording() bool` / `HasAccessibility() bool` / `OpenSettings(kind PermissionKind) error` を提供。
- **`app.go` の `App` 型**: Wails のバインドエントリ。フロントエンドが呼ぶ操作（権限チェック、Region/ClickPoint 選択開始、Capture Session 開始/停止、Output Document を開く）を公開する。

### Wails ウィンドウ構成

- メインウィンドウは廃止（ADR-0002）。代わりに `Frameless: true` + `AlwaysOnTop: true` + 小サイズ（例: 480x80）のフローティングバー1つを開く。
- 範囲選択・クリック位置選択時は、別ウィンドウとして全画面透明オーバーレイを動的に開閉する（`runtime.WindowReload` ではなく、必要に応じて 2nd Wails ウィンドウを `wails.Run` の Options で予め用意するか、フロント側で全画面 div として実装）。
- 実装方式は最終的に「フロント側で全画面 div を一時的に表示し、ポインタイベントを取る」案を第一候補とする。マウス座標は Wails 経由でグローバル座標に変換する必要があり、必要なら `runtime` API を介して Go 側からスクリーン座標を取得する。

### フロントエンド構成（React + TypeScript + Vite）

- 既存テンプレートのまま React で実装。
- フローティングバーは単一ページコンポーネントとして実装。状態管理はローカルステートで十分（永続化は MVP では行わない、ADR と Q5 の方針）。
- バー内部のレイアウト：横長または2段組で、Region/ClickPoint ボタン・Repeat Count 入力・Step Interval 入力・出力先選択・開始/停止ボタン・進捗表示を配置。

### ドメイン用語の遵守

`CONTEXT.md` で定義された用語（Capture Region / Advance Click Point / Capture Step / Capture Session / Repeat Count / Step Interval / Output Document）を Go の型名・関数名・UI ラベル全てで一貫して使用する。

### エラー処理

- Screener・Clicker・PdfWriter が返した error は CaptureSession 内で捕捉し、ループを break、Output Document を Close、フロントへ通知。
- Wails のイベント機構（`runtime.EventsEmit`）で `session:error` イベントを発行し、フローティングバーが受信して赤色エラー表示する。
- 進捗も同様に `session:progress` イベントで通知。

### ファイル命名と衝突解消

- デフォルト名は Go の `time.Now().Format("pasha-2006-01-02_15-04")` ベース。
- 保存パス確定時に同名チェックし、存在すれば `-2`, `-3` ... を末尾に付与（PDF 拡張子の前）。CaptureSession 開始前にファイル名を確定して PdfWriter に渡す。

### 対応プラットフォーム

macOS のみ（Q6 / ADR の方針）。`wails.json` のビルドターゲット、依存ライブラリ選定は macOS 前提で進める。

## Testing Decisions

### テストの基本方針

- **外部挙動のみテストする**（実装詳細はテストしない）。CaptureSession の入力（設定 + コラボレータの振る舞い）に対する出力（コラボレータへの呼び出し順序・回数、返却 error、Output Document の Close 状態）を検証する。
- **OS 境界を跨ぐ層（実 Screener・実 Clicker・実 PdfWriter・実 permissions）は自動テスト対象外**。これらは macOS 上での手動スモークテストで担保。

### 主seam: `internal/session.CaptureSession`（Go）

唯一のテスト seam。`internal/session_test.go` に以下のシナリオを Go の標準 `testing` パッケージで実装：

1. **正常系：N 回ループ実行**
   - Repeat Count=5 を渡し、Screener.Capture が 5 回呼ばれ、Clicker.Click が 5 回呼ばれ、PdfWriter.AppendPage が 5 回呼ばれ、PdfWriter.Close が最後に 1 回呼ばれることを検証。
2. **ステップ順序の検証**
   - 1 ステップ内で Screener → PdfWriter.AppendPage → Clicker → Clock.Sleep の順に呼ばれることを検証。
3. **Step Interval の遵守**
   - フェイク Clock の Sleep が Step Interval と等しい duration で呼ばれることを検証。
4. **Stop() で中断**
   - 別ゴルーチンから途中で Stop() を呼んだ時、現在のステップ完了後にループが終了し、PdfWriter.Close が呼ばれることを検証。
5. **Screener エラーで即時中断**
   - 3 回目の Screener.Capture が error を返したら、それ以降の Clicker/AppendPage が呼ばれず、PdfWriter.Close が呼ばれ、CaptureSession.Start が error を返すことを検証。
6. **Clicker エラーで即時中断**（同上の Clicker 版）
7. **PdfWriter.AppendPage エラーで即時中断**（同上の PdfWriter 版）
8. **context キャンセルで中断**
   - 渡された ctx がキャンセルされた時、ループが終了し PdfWriter.Close が呼ばれることを検証。

### フェイク実装

`internal/session/fakes_test.go` に Screener / Clicker / PdfWriter / Clock のフェイクを定義。各フェイクは「呼び出し履歴の記録」「特定回数で error を返す設定」「Clock の Sleep 即時 return」を提供。

### 既存のテストの先例

現状リポジトリには `_test.go` は1つも存在しない（Wails テンプレート直後の状態）。本 PRD で導入する `internal/session_test.go` がプロジェクトで初のテストとなる。以降のテストはこのスタイル（標準 testing パッケージ + 同 package 内 fake）を踏襲する。

### 自動テストしない領域（手動QA）

- 実スクリーンキャプチャの画質・色再現
- 実合成クリックがターゲットアプリで正しく反応すること
- 生成された Output Document が Preview.app で正常に開けること
- 範囲選択・クリック位置選択オーバーレイの UX
- フローティングバーの常時表示・ドラッグ移動・常に最前面挙動
- macOS 権限ダイアログ遷移と System Settings 連携

これらは README にスモークテスト手順として記載する。

## Out of Scope

以下は本 PRD（MVP）の対象外。将来必要が確認できた時に別 PRD として切り出す：

- **設定の永続化 / プリセット機能**（Q5 で deferred）。Repeat Count や Region 等を記憶して再起動後も復元する機能。
- **キー入力・ホイールスクロールによる送りアクション**（Q11 で deferred）。Advance Click Point の概念は「左クリック」固定のまま。汎用化時には `Advance Action` へのドメイン用語リネームが必要。
- **画像差分による自動停止**（Q3 で deferred）。ページが進まなくなったら止める機能。
- **Windows / Linux 対応**（Q6 で deferred）。
- **マルチモニタ環境の特殊対応**。MVP では主モニタ前提。
- **キャプチャ後の PDF 編集機能**（ページ削除、並び替え、再撮影差し替え等）。
- **OCR・テキスト抽出・PDF 内検索可能化**。Output Document はラスタ画像の集合体。
- **撮影後の PNG エクスポート / 個別ページ書き出し**。
- **クラウド連携・自動アップロード**。
- **フローティングバーの色テーマ・ダークモード対応**。OS のシステム設定に追従する標準的な見た目で良い。
- **キーボードショートカット**。Q3 で「停止は UI ボタンのみ」と確定済み。

## Further Notes

- 本 PRD は ADR-0001（PDF逐次追記）と ADR-0002（フローティングバー1つ集約）に従う。実装時にこの方針から逸脱する判断が必要になったら、新たな ADR を作成して理由を残すこと。
- ドメイン用語は `CONTEXT.md` を単一の正とする。新たな概念が必要になったら `CONTEXT.md` を更新してから実装に進む。
- フェイクを使ったテストの分離度を保つために、`internal/session` パッケージは `internal/screener` 等の具体実装パッケージを import してはならない。インターフェースは `internal/session` 側で定義し、実装側がそれを満たす形にする（依存性逆転）。
- 撮影時間は最大数十分に及び得るため、`Start` メソッドは長時間ブロッキングが前提。フロントとの通信は Wails のイベント機構を使い、進捗 / 完了 / エラーを非同期通知する。

## Comments
