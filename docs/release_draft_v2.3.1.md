# 次期リリース（v2.3.1想定）リリースノート・対応内容まとめ

別セッションでのリリース作業時に本ファイルを参照してください。

---

## 1. 概要
Go 1.27環境およびmacOS環境下において、ARP WatchおよびPINGバックエンドの処理競合やOSの省電力・画面ロック制御に起因して、ローカルノードのPING・TCP・UDP監視が一斉にタイムアウト障害となる問題を根本改修しました。また、macOSで常時監視を行うユーザー向けのシステム設定手順をドキュメントに明文化しました。

---

## 2. 変更内容（Changelog / Release Notes）

### 改善・修正（Improvements & Bug Fixes）
- **ARP WatchのPING送信平滑化と軽量化 (`logger/arpwatch.go`)**:
  - 5秒ごとのバースト送信（最大50個）を廃止し、**200ms間隔（1秒に5個）の平滑送信**に変更。
  - 1秒のタイムアウト完了待ち（ブロッキング）を排除し、ARP解決のみを目的とした**応答待ちなし送信（Fire-and-Forget / `SendPing`）**に変更。
  - macOSのARP解決キューやスイッチへの急激な負荷、およびARPテーブル飽和（`UHRLWI` 大量蓄積による `no route to host` エラー）を根本解消。
- **PINGバックエンドの送受信Goroutine分離 (`ping/ping.go`)**:
  - ソケットからの受信（`ReadFrom`）を**受信専用Goroutine**に分離。
  - イベントループから `default`（100ms受信ブロッキング）を完全撤廃し、送信要求（`pingSendCh`）、再送・タイムアウト判定（`timer.C`）、受信処理（`recvCh`）を独立イベント駆動化。
  - 未使用IPへの探索や外部宛て遅延パケットが通常ノードのPING監視の送受信やタイマー発火を妨げないように改善。
  - ARP解決専用の軽量送信API `SendPing(ip, size, ttl)` を新設（`PingEnt.NoWait = true`）。
- **macOS App Nap 無効化設定の追加 (`Info.plist`, `Info.dev.plist`)**:
  - `<key>NSAppSleepDisabled</key><true/>` を追加し、無操作時やスクリーンセーバー時にOSがアプリのタイマーや通信をスロットリングするのを防止。

### ドキュメント（Documentation）
- **macOS環境での常時監視に関する推奨設定の追記**:
  - Webページ日本語版 (`docs/index_ja.md`)
  - Webページ英語版 (`docs/index.md`)
  - マニュアル日本語版スライド (`docs/twsnmpfk_ja.md`)
  - マニュアル英語版スライド (`docs/twsnmpfk_en.md`)
  - ※ 画面ロック、スクリーンセーバー、ディスプレイオフ、ハードディスクスリープ、自動ログアウトの無効化手順を明文化。

---

## 3. 変更されたファイル一覧
- `logger/arpwatch.go`
- `ping/ping.go`
- `build/darwin/Info.plist`
- `build/darwin/Info.dev.plist`
- `docs/index_ja.md`
- `docs/index.md`
- `docs/twsnmpfk_ja.md`
- `docs/twsnmpfk_en.md`
