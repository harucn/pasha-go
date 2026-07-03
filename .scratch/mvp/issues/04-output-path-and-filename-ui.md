Status: done

# 04: 出力先・ファイル名 UI

## Parent

`.scratch/mvp/PRD.md`

## What to build

Output Document の保存先フォルダとファイル名を UI で指定できるようにする。

- フロント：
  - 「保存先フォルダ」ボタン押下 → ネイティブのフォルダ選択ダイアログ（Wails の `runtime.OpenDirectoryDialog`）→ 選んだパスを表示。
  - ファイル名入力欄：デフォルト値は `pasha-YYYY-MM-DD_HH-MM` 形式のタイムスタンプ（フロント側 or Go 側で生成）。ユーザーが編集可能。
- `app.go`：
  - 保存先フォルダ + ファイル名から完全パスを組み立てる。
  - 同名ファイルが既に存在する場合、`-2`, `-3`, ... の連番を拡張子の前に付与（衝突解消は CaptureSession 開始前に行い、確定したパスを PdfWriter に渡す）。
- 開始ボタンは「フォルダ未選択」または「ファイル名空」のとき押せない。

## Acceptance criteria

- [x] UI からフォルダを選び、ファイル名を編集して開始すると、その場所にその名前で PDF が生成される
- [x] ファイル名を空にしたまま開始ボタンが押せない
- [x] 同名 PDF が既存の状態で開始すると、`pasha-...-2.pdf` のような連番ファイルが生成される（既存ファイルは上書きされない）
- [x] ファイル名入力欄のデフォルト値がタイムスタンプ形式で初期表示される

## Blocked by

- 01: Tracer bullet — 最小e2e Capture Session
