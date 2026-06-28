# pasha-go

デスクトップの一部を繰り返しキャプチャし、所定の位置をクリックして次の画面に進ませる自動化ツール。紙資料・PDFビューア・プレゼン資料などを順次PDF化するユースケース。

## Language

**Capture Region**:
キャプチャ対象として指定された画面上の矩形領域。1セッション中は固定。
_Avoid_: スクショ範囲、エリア

**Advance Click Point**:
1ステップごとにクリックする画面上の1点。次の画面（次ページ等）に送るための座標。
_Avoid_: 次へボタン、クリック先

**Capture Step**:
1回分の「キャプチャ → クリック → 待機」の一連の動作。
_Avoid_: 1ループ、1イテレーション、1ページ

**Capture Session**:
ユーザーが「開始」してから停止・完了までの、N回のCapture Stepの実行全体。
_Avoid_: ジョブ、ラン、撮影

**Repeat Count**:
Capture Sessionで実行するCapture Stepの予定回数。
_Avoid_: ページ数、ループ回数

**Step Interval**:
1つのCapture Stepの「クリック」から次のStepの「キャプチャ」までの待ち時間。
_Avoid_: ディレイ、ウェイト

**Output Document**:
Capture Sessionの成果物として逐次追記されるマルチページPDF。
_Avoid_: 出力ファイル、結果PDF
