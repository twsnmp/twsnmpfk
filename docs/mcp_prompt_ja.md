# MCPサーバープロンプト仕様

このドキュメントは、TWSNMP FK MCPサーバー（`backend/mcp_prompts.go` および `backend/mcp_prompts_extended.go`）で利用可能な **59個のプロンプト** の仕様を定義します。

## 基本プロンプト（1–19）

### 1. `get_node_list`

- **タイトル:** フィルタ付きでノードリストを取得
- **説明:** TWSNMPに登録されているノードのリストをフィルタ付きで取得します。
- **引数:**
  - `state_filter` (オプション): ノードの状態フィルタ (`normal`, `repair`, `warn`, `low`, `high`, `unknown`)。
  - `name_filter` (オプション): ノード名フィルタ。
  - `ip_filter` (オプション): ノードIPアドレスフィルタ。

### 2. `add_node`

- **タイトル:** TWSNMPに新しいノードを追加
- **説明:** TWSNMPに新しいノードを追加します。
- **引数:**
  - `name` (必須): 新しいノードの名前。
  - `ip` (必須): 新しいノードのIPアドレス。
  - `icon` (オプション): ノードのアイコン (`desktop`, `laptop`, `server`, `cloud`, `router`, `ip` 等)。
  - `description` (オプション): 説明文。
  - `position` (オプション): 表示座標 (例: `x=100,y=200`)。

### 3. `update_node`

- **タイトル:** TWSNMPのノードを更新
- **説明:** TWSNMP上の既存ノードを更新します。
- **引数:**
  - `id` (必須): 更新するノードのID、名前、またはIPアドレス。
  - `name`, `ip`, `icon`, `description`, `position` (すべてオプション)。

### 4. `get_network_list`

- **タイトル:** フィルタ付きでネットワークリストを取得
- **説明:** TWSNMPに登録されているネットワークコンテナのリストをフィルタ付きで取得します。
- **引数:**
  - `name_filter` (オプション), `ip_filter` (オプション)。

### 5. `get_polling_list`

- **タイトル:** フィルタ付きでポーリングリストを取得
- **説明:** TWSNMPに登録されているポーリングのリストをフィルタ付きで取得します。
- **引数:**
  - `type_filter`, `name_filter`, `node_name_filter`, `state_filter` (オプション)。
  - `node_id` (オプション): 対象ノードIDまたは一意なノード名。

### 6. `get_polling_log`

- **タイトル:** ポーリングログを取得
- **説明:** TWSNMPからポーリングログを取得します。
- **引数:**
  - `id` (必須): ポーリングID。
  - `limit` (オプション): 最大件数 (デフォルト `100`)。
  - `start_time` (オプション): 開始時刻 (`"-24h"` や RFC3339 等)。
  - `end_time` (オプション): 終了時刻 (デフォルト `now`)。

### 7. `do_ping`

- **タイトル:** pingを実行
- **説明:** TWSNMPからターゲットにpingを実行します。
- **引数:**
  - `target` (必須), `size` (オプション), `ttl` (オプション), `timeout` (オプション)。

### 8. `snmpwalk`

- **タイトル:** snmpwalkを実行
- **説明:** TWSNMPからターゲットにsnmpwalkを実行します。
- **引数:**
  - `target` (必須), `mib_object_name` (必須)。
  - `snmp_mode` (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`, `v3sha256aes128`, `v3sha512aes256`), `community`, `user`, `password`, `port`, `limit` (すべてオプション)。

### 9. `search_event_log`

- **タイトル:** フィルタ付きでイベントログを検索
- **説明:** フィルタ付きでイベントログを検索します。
- **引数:**
  - `type_filter`, `node_filter`, `level_filter` (`info`, `normal`, `warn`, `low`, `high`), `event_filter`, `start_time`, `end_time`, `limit` (すべてオプション)。

### 10. `search_syslog`

- **タイトル:** フィルタ付きでsyslogを検索
- **説明:** フィルタ付きでsyslogを検索します。
- **引数:**
  - `level_filter`, `host_filter`, `tag_filter`, `message_filter`, `start_time`, `end_time`, `limit` (すべてオプション)。

### 11. `get_syslog_summary`

- **タイトル:** フィルタ付きでsyslogサマリーを取得
- **説明:** フィルタ付きでsyslogのサマリーを取得します。
- **引数:**
  - `level_filter`, `host_filter`, `tag_filter`, `message_filter`, `start_time`, `end_time`, `top_n` (すべてオプション)。

### 12. `search_snmp_trap_log`

- **タイトル:** フィルタ付きでSNMPトラップログを検索
- **説明:** TWSNMPのSNMPトラップログをフィルタ付きで検索します。
- **引数:**
  - `from_filter`, `trap_type_filter`, `variable_filter`, `start_time`, `end_time`, `limit` (すべてオプション)。

### 13. `get_mib_tree`

- **タイトル:** TWSNMPのMIBツリーを取得
- **説明:** `get_mib_tree` ツールでMIBツリーを取得します（キーワード検索には `search_mib` を推奨）。
- **引数:**
  - `oid` (オプション): サブツリーOID。
  - `name` (オプション): オブジェクト名。
  - `depth` (オプション): 最大深さ。

### 14. `get_ip_address_list`

- **タイトル:** 管理対象IPアドレス一覧を取得
- **説明:** `get_ip_address_list` ツールを使用して管理されているIPアドレス一覧を取得します。
- **引数:**
  - `ip_filter`, `mac_filter`, `vendor_filter`, `limit` (すべてオプション)。

### 15. `get_resource_monitor_list`

- **タイトル:** リソースモニター情報を取得
- **説明:** `get_resource_monitor_list` ツールでリソースモニター情報を取得します。
- **引数:** なし。

### 16. `get_server_certificate_list`

- **タイトル:** サーバー証明書リストを取得
- **説明:** `get_server_certificate_list` ツールで監視中のサーバー証明書一覧を取得します。
- **引数:** なし。

### 17. `add_event_log`

- **タイトル:** イベントログを追加
- **説明:** TWSNMPにイベントログを追加します。
- **引数:**
  - `level` (必須): ログレベル (`info`, `normal`, `warn`, `low`, `high`)。
  - `node` (オプション): ノード名またはID。
  - `event` (必須): イベントログの内容。

### 18. `get_ip_address_info`

- **タイトル:** IPアドレス情報を取得
- **説明:** 指定IPアドレスの詳細情報を取得します。
- **引数:** `ip` (必須)。

### 19. `get_mac_address_info`

- **タイトル:** MACアドレス情報を取得
- **説明:** 指定MACアドレスの詳細情報を取得します。
- **引数:** `mac` (必須)。

---

## 拡張プロンプト・診断ワークフロー（20–59）

### 20. `troubleshoot_network`

- **タイトル:** ネットワーク／ノード障害の総合調査
- **説明:** `get_system_status`、異常ノード (`get_node_list`)、異常ポーリング (`get_polling_list`)、ライン (`get_line_list`)、イベントログ、Syslog、対象指定時は `diagnose_node` を組み合わせた段階的調査ガイド。
- **引数:** `target` (オプション), `start_time` (オプション)。

### 21. `diagnose_node`

- **タイトル:** 対象ノード／IPのライブ総合診断
- **説明:** `diagnose_node` で Ping / SNMP / TCPポート / HTTP を一括診断します。
- **引数:** `target` (必須), `ports` (オプション), `timeout` (オプション)。

### 22. `get_system_status`

- **タイトル:** TWSNMPシステム状態を取得
- **説明:** TWSNMPランタイム、レシーバー稼働状況、DBサイズ、監視サマリーを確認します。
- **引数:** なし。

### 23. `get_node`

- **タイトル:** 単一ノードの詳細を取得
- **説明:** `get_node` でノード詳細設定とポーリング一覧を取得します。
- **引数:** `id` (必須)。

### 24. `delete_node`

- **タイトル:** ノードを削除
- **説明:** `get_node` で確認後、`delete_node` でノード・関連ポーリング・関連ラインを削除します。
- **引数:** `id` (必須)。

### 25. `get_polling`

- **タイトル:** ポーリング設定を取得
- **説明:** `get_polling` で設定と最新結果を取得します。
- **引数:** `id` (必須)。

### 26. `add_polling`

- **タイトル:** ポーリングを追加
- **説明:** `add_polling` で新しいポーリングを作成します。
- **引数:** `node_id` (必須), `name` (必須), `type` (必須), `params` (オプション), `level` (オプション)。

### 27. `update_polling`

- **タイトル:** ポーリング設定を更新
- **説明:** `update_polling` で既存ポーリングを変更します。
- **引数:** `id` (必須), `level` (オプション), `poll_int` (オプション), `params` (オプション)。

### 28. `delete_polling`

- **タイトル:** ポーリングを削除
- **説明:** `delete_polling` でポーリングを削除します。
- **引数:** `id` (必須)。

### 29. `check_polling`

- **タイトル:** ポーリングを即時実行
- **説明:** `check_polling` で単一・ノード単位・全体のポーリング即時実行をトリガーします。
- **引数:** `polling_id` (オプション), `node_id` (オプション), `all` (オプション)。

### 30. `get_polling_log_data`

- **タイトル:** ポーリングログをCSV出力
- **説明:** `get_polling_log_data` で時系列CSVデータを取得します。
- **引数:** `id` (必須), `limit` (オプション), `start_time` (オプション), `end_time` (オプション)。

### 31. `get_line_list`

- **タイトル:** マップライン一覧を取得
- **説明:** `get_line_list` でライン一覧を取得します。
- **引数:** `node_id` (オプション), `state_filter` (オプション)。

### 32. `add_line`

- **タイトル:** マップラインを追加
- **説明:** `add_line` で2つのノードまたはネットワークを接続します。
- **引数:** `node_id1` (必須), `node_id2` (必須), `polling_id1` (オプション), `polling_id2` (オプション)。

### 33. `delete_line`

- **タイトル:** マップラインを削除
- **説明:** `delete_line` で指定IDのラインを削除します。
- **引数:** `id` (必須)。

### 34. `get_network`

- **タイトル:** ネットワークコンテナ詳細を取得
- **説明:** `get_network` で詳細情報とポートテーブルを取得します。
- **引数:** `id` (必須)。

### 35. `add_network`

- **タイトル:** ネットワークコンテナを追加
- **説明:** `add_network` でネットワーク/スイッチを追加します。
- **引数:** `name` (必須), `ip` (必須), `unmanaged` (オプション)。

### 36. `update_network`

- **タイトル:** ネットワークコンテナを更新
- **説明:** `update_network` で設定を変更します。
- **引数:** `id` (必須), `name` (オプション), `ip` (オプション)。

### 37. `delete_network`

- **タイトル:** ネットワークコンテナを削除
- **説明:** `delete_network` でネットワークと接続ラインを削除します。
- **引数:** `id` (必須)。

### 38. `check_network`

- **タイトル:** ネットワーク状態を確認
- **説明:** `check_network` でSNMP/LLDPチェックを即時実行します。
- **引数:** `id` (オプション), `all` (オプション)。

### 39. `update_server_certificate`

- **タイトル:** TLSサーバー証明書監視を追加・更新
- **説明:** `update_server_certificate` で証明書監視エントリを追加または更新します。
- **引数:** `id` (オプション), `target` (オプション), `port` (オプション)。

### 40. `delete_server_certificate`

- **タイトル:** TLSサーバー証明書監視を削除
- **説明:** `delete_server_certificate` で証明書監視エントリを削除します。
- **引数:** `id` (必須)。

### 41. `snmpset`

- **タイトル:** SNMP MIB値を設定
- **説明:** `snmpset` で書き込み可能なMIBオブジェクトを変更します。
- **引数:** `target` (必須), `mib_object_name` (必須), `type` (必須), `value` (必須)。

### 42. `search_netflow`

- **タイトル:** NetFlowを検索
- **説明:** `search_netflow` でフローデータを検索します。
- **引数:** `src_filter` (オプション), `dst_filter` (オプション), `start_time` (オプション), `limit` (オプション)。

### 43. `search_sflow`

- **タイトル:** sFlowを検索
- **説明:** `search_sflow` でsFlowサンプルを検索します。
- **引数:** `src_filter` (オプション), `dst_filter` (オプション), `start_time` (オプション), `limit` (オプション)。

### 44. `search_sflow_counter`

- **タイトル:** sFlowカウンターを検索
- **説明:** `search_sflow_counter` でインターフェースカウンターを検索します。
- **引数:** `remote_filter` (オプション), `type_filter` (オプション), `start_time` (オプション)。

### 45. `get_otel_metrics`

- **タイトル:** OpenTelemetryメトリクスを取得
- **説明:** `get_otel_metrics` でOTelメトリクスを検索します。
- **引数:** `host_filter` (オプション), `service_filter` (オプション), `name_filter` (オプション)。

### 46. `get_otel_traces`

- **タイトル:** OpenTelemetryトレース一覧を取得
- **説明:** `get_otel_traces` で最近のトレースまたはエラートレースを取得します。
- **引数:** `service_filter` (オプション), `error_only` (オプション)。

### 47. `get_otel_trace_detail`

- **タイトル:** OpenTelemetryトレース詳細を取得
- **説明:** `get_otel_trace_detail` でトレースの全スパンを確認します。
- **引数:** `trace_id` (必須), `bucket` (オプション)。

### 48. `get_ai_anomaly_list`

- **タイトル:** AI異常検知スコア一覧を取得
- **説明:** `get_ai_anomaly_list` で異常検知結果を取得します。
- **引数:** `node_id` (オプション), `min_score` (オプション)。

### 49. `search_mib`

- **タイトル:** MIB定義を検索
- **説明:** `search_mib` でロード済みMIBオブジェクトを検索します。
- **引数:** `query` (必須), `limit` (オプション)。

### 50. `get_sensor_list`

- **タイトル:** センサー稼働状況を取得
- **説明:** `get_sensor_list` で登録センサーのヘルス一覧を取得します。
- **引数:** `host_filter` (オプション), `type_filter` (オプション)。

### 51. `get_report_summary`

- **タイトル:** ログ分析レポートを取得
- **説明:** `get_report_summary` で集約済み各種レポートを取得します。
- **引数:** `report_type` (必須), `filter` (オプション)。

### 52. `get_mac_address_list`

- **タイトル:** MACアドレス一覧を取得
- **説明:** `get_mac_address_list` で集約MACテーブルを取得します。
- **引数:** `mac_filter` (オプション), `vendor_filter` (オプション)。

### 53. `get_wifi_ap_list`

- **タイトル:** 検出Wi-Fi AP一覧を取得
- **説明:** `get_wifi_ap_list` でWi-Fi AP一覧を取得します。
- **引数:** `ssid_filter` (オプション), `bssid_filter` (オプション)。

### 54. `get_bluetooth_device_list`

- **タイトル:** 検出Bluetoothデバイス一覧を取得
- **説明:** `get_bluetooth_device_list` でBluetooth/BLEデバイス一覧を取得します。
- **引数:** `name_filter` (オプション), `public_address_only` (オプション)。

### 55. `get_fdb_table`

- **タイトル:** スイッチFDBテーブルを取得
- **説明:** `get_fdb_table` でSNMPスイッチのFDBテーブルを取得します。
- **引数:** `id` (必須)。

### 56. `auto_connect_lines`

- **タイトル:** マップライン自動接続
- **説明:** `auto_connect_lines` でトポロジラインを自動接続します。
- **引数:** `mode` (オプション: `strict` または `speculative`)。

### 57. `auto_layout`

- **タイトル:** マップ自動レイアウト
- **説明:** `auto_layout` でノード配置を最適化、または直前の配置を元に戻します。
- **引数:** `mode` (オプション: `hierarchical`, `cluster`, `categorized`), `undo` (オプション)。

### 58. `gnmi_capabilities`

- **タイトル:** gNMI Capabilitiesを取得
- **説明:** `gnmi_capabilities` でgNMI対応モデル・エンコーディングを取得します。
- **引数:** `node_id` (必須), `target` (オプション)。

### 59. `gnmi_get`

- **タイトル:** gNMIテレメトリーパスを取得
- **説明:** `gnmi_get` で指定パスのgNMIデータを取得します。
- **引数:** `node_id` (必須), `path` (必須), `encoding` (オプション)。
