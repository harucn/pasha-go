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

**Capture Session Plan**:
Capture Session を開始するために確定している入力一式：Repeat Count・Step Interval・Capture Region・Advance Click Point・Output Document の保存先とファイル名。ユーザーが「開始」を押した瞬間にスナップショットされ、以降のセッション内では変更されない。
_Avoid_: 設定、パラメータ、リクエスト

**Screen Space**:
pasha-go 内で画面上の座標を扱うときの唯一の共通座標系。プライマリディスプレイの左上を原点 (0, 0) とし、x を右、y を下向き、単位は logical points。マルチディスプレイの負値やプライマリ幅超えも正当な Screen Space 値として扱う。詳細は `docs/adr/0003-canonical-screen-coordinate-space.md`。
_Avoid_: 画面座標（曖昧なので使わない）、pixel 座標

**Selection Window**:
Capture Region と Advance Click Point を選ぶあいだ、フローティングバーが一時的に姿を変えた、リサイズ可能な枠つきウィンドウ。その矩形がそのまま Capture Region になる。開くときにバーの geometry（位置・サイズ・サイズ固定）を退避してサイズ固定を解除し、確定・取消のいずれで閉じても必ず退避した状態へ復元する。この「解除と復元は対である」が Selection Window の唯一の不変条件。
_Avoid_: 範囲選択ダイアログ、オーバーレイ、選択モード
