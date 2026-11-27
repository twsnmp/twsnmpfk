# TWSNMP FK MCPサーバー仕様書

## 概要

このドキュメントは、TWSNMP FKに実装されているModel Context Protocol（MCP）サーバー（`backend/mcp.go`、`backend/mcp_tools.go`、`backend/mcp_tools_crud.go`、`backend/mcp_tools_analytics.go`）の仕様を定義します。

AIエージェントは、定義された **58個のツール** と **59個のプロンプト** を通じて、監視データの取得、フロー/テレメトリー/ログの検索、ノード/ポーリング/ライン/ネットワーク/証明書の管理、リアルタイム診断、およびマップレイアウトを実行できます。

## エンドポイントとトランスポート

MCPサーバーは `datastore.MapConf.MCPTransport`（`"off"`、`"sse"`、`"stream"`）の設定に従って動作します：
- `"stream"`（Streamable HTTP）: `/mcp` を提供
- `"sse"`（Server-Sent Events）: `/sse` と `/message` を提供
- `"off"`: MCPサーバーは無効

待受アドレスは `datastore.MapConf.MCPEndpoint`（例: `127.0.0.1:8089` または `:8089`）を使用します。`-mcpCert` と `-mcpKey` を指定するとTLS 1.3で暗号化されます。

## 認証とアクセス制御

すべてのエンドポイントは `checkMCPACL` によって保護されます。トークン認証とIPアクセス制御の両方が設定されている場合、**両方の条件を満たす必要があります**：

1. **トークン認証 (`datastore.MapConf.MCPToken`)**:
   - 設定されている場合、HTTP `Authorization` ヘッダーと定数時間（constant-time）での完全一致検証を行います。
   - `Authorization: Bearer <token>`（`Bearer ` は大文字・小文字を区別せず）および `Authorization: <token>` の両形式を受け付けます。
2. **IPアクセス制御 (`datastore.MapConf.MCPFrom`)**:
   - カンマ区切りで許可IPアドレスを指定します。
   - 設定更新時に古いIP許可エントリはクリアされ、最新の許可リストに原子的に切り替わります。
   - 空の場合は全IPからのアクセスを許可します。
3. **認証情報のマスク処理**:
   - `get_node` および `get_network` が返す機密項目（`Community`、`Password`、`GNMIPassword`、`PublicKey`）は、設定済みの場合 `"********"` にマスクされます。

プロンプト仕様の詳細は [MCPサーバープロンプト仕様](./mcp_prompt_ja.md) を参照してください。

---

## ツール仕様（全58ツール）

### ノード管理

#### `get_node_list`
TWSNMPに登録されているノードの一覧を取得します。
- `name_filter` (string, オプション): ノード名の正規表現フィルタ。
- `ip_filter` (string, オプション): IPアドレスの正規表現フィルタ。
- `state_filter` (string, オプション): 状態（`normal`, `warn`, `low`, `high`, `repair`, `unknown`）の正規表現フィルタ。

#### `get_node`
単一ノードの詳細設定（機密フィールドはマスク）および関連ポーリング一覧を取得します。
- `id` (string, 必須): ノードID、一意なノード名、または一意なIPアドレス。

#### `add_node`
新しいノードをTWSNMPに追加します。
- `name` (string, 必須): ノード名。
- `ip` (string, 必須): IPアドレス。
- `icon` (string, オプション): アイコン名（デフォルト: `desktop`）。
- `description` (string, オプション): 説明文。
- `x` (integer, オプション): X座標（`1`〜`1000`、デフォルト `64`）。
- `y` (integer, オプション): Y座標（`1`〜`1000`、デフォルト `64`）。

#### `update_node`
既存ノードの情報・座標を更新します（`id` はID、一意名、一意IPに対応）。
- `id` (string, 必須): 対象ノードのID、現在の名前、またはIPアドレス。
- `name`, `ip`, `icon`, `description` (string, オプション), `x`, `y` (integer, オプション)。

#### `delete_node`
ノードを削除し、配下のポーリング（稼働中のgNMI/MQTTサブスクリプションも停止）および接続されているマップラインも連動して削除します。
- `id` (string, 必須): ノードID、一意なノード名、または一意なIPアドレス。

---

### ポーリング管理・実行

#### `get_polling_list`
登録されているポーリング設定の一覧を取得します。
- `type_filter` (string, オプション): ポーリング種別の正規表現（`ping`, `tcp`, `http`, `dns`, `twsnmp`, `syslog`, `stun`, `snmp` など）。
- `name_filter` (string, オプション): ポーリング名の正規表現。
- `node_name_filter` (string, オプション): ノード名の正規表現。
- `state_filter` (string, オプション): 状態の正規表現。
- `node_id` (string, オプション): 指定ノードIDまたは一意名での完全一致フィルタ。

#### `get_polling`
単一ポーリングの全設定と最新結果を取得します。
- `id` (string, 必須): ポーリングID。

#### `add_polling`
既存ノードに新しいポーリングを追加します。
- `node_id` (string, 必須): 対象ノードID、一意なノード名、またはIPアドレス。
- `name` (string, 必須): ポーリング名。
- `type` (string, 必須): ポーリング種別（`ping`, `tcp`, `http`, `dns`, `tls`, `snmp`, `gnmi`, `mqtt`, `syslog`, `trap`, `netflow`, `sflow`, `twsnmp`, `ntp`, `ssh`, `stun`, `cmd`, `report` など）。
- `mode`, `params`, `filter`, `extractor`, `script` (string, オプション)。
- `level` (string, オプション): `off`, `low`（デフォルト）, `warn`, `high`。
- `poll_int` (integer, オプション): 間隔秒（デフォルト: マップ設定または `60`）。
- `timeout` (integer, オプション): タイムアウト秒（デフォルト: マップ設定または `3`）。
- `retry` (integer, オプション): リトライ回数（デフォルト: マップ設定または `0`）。
- `log_mode` (string, オプション): `none`（デフォルト）, `always`, `on_change`, `ai`。
- `fail_action`, `repair_action`, `ai_mode`, `vector_cols`, `mqtt_url`, `mqtt_topic`, `mqtt_cols` (string, オプション)。

#### `update_polling`
既存ポーリングを更新します。対象ノード・モード・種別・パラメータ・MQTT設定の変更時は稼働中サブスクリプションを安全に停止します。
- `id` (string, 必須): ポーリングID。
- その他 `add_polling` と同様の任意フィールド（`retry` に負数を指定すると `0` にリセット）。

#### `delete_polling`
ポーリングを削除し、関連するgNMI/MQTTサブスクリプションを停止します。
- `id` (string, 必須): ポーリングID。

#### `check_polling`
単一ポーリング、指定ノードの全ポーリング、または全ポーリング/ネットワークの即時実行を指示します。
- `polling_id` (string, オプション): ポーリングID。
- `node_id` (string, オプション): ノードID、一意名、またはIP。
- `all` (boolean, オプション): `true` で全ポーリングとネットワークチェックを実行。

#### `get_polling_log`
ポーリング履歴ログをJSONで取得します。
- `id` (string, 必須): ポーリングID。
- `limit` (integer, オプション): 取得件数（`1`〜`2000`、デフォルト `100`）。

#### `get_polling_log_data`
ポーリングの数値ログを決定論的にソートされたCSVヘッダー（`time,state,<昇順メトリクス名...>`）で出力します。
- `id` (string, 必須): ポーリングID。
- `limit` (integer, オプション): 取得行数（`1`〜`2000`、デフォルト `100`）。

---

### ライン・ネットワーク管理

#### `get_line_list`
マップ上のライン一覧を、接続端の名前・状態付きで取得します。
- `node_id` (string, オプション): 指定ノードまたはネットワークに接続するラインのみに絞り込み。
- `state_filter` (string, オプション): ライン全体状態の正規表現フィルタ。

#### `add_line`
2つのノードまたはネットワーク（`NET:<id>` または一意名/IP）を接続ラインで結び、即座にライン状態を更新します。
- `node_id1`, `node_id2` (string, 必須): 両端のノードまたはネットワーク。
- `polling_id1`, `polling_id2`, `polling_id`, `port`, `info` (string, オプション), `width` (integer, オプション: `1`〜`10`, デフォルト `2`)。

#### `delete_line`
指定IDのマップラインを削除します。
- `id` (string, 必須): ラインID。

#### `get_network_list`
ネットワークコンテナの一覧を取得します。
- `name_filter`, `ip_filter` (string, オプション): 正規表現フィルタ。

#### `get_network`
単一ネットワークコンテナの詳細およびポートテーブルを取得します（SNMP資格情報はマスク）。
- `id` (string, 必須): ネットワークID（`NET:` 接頭辞の有無は不問）、一意名、またはIP。

#### `add_network`
ネットワーク/スイッチコンテナを追加します。
- `name` (string, 必須), `ip` (string, 必須)。
- `description`, `snmp_mode`, `community`, `user`, `password`, `url` (string, オプション)。
- `snmp_port` (`161`), `h_ports` (`24`), `x` (`64`), `y` (`64`) (integer, オプション)。
- `arp_watch`, `unmanaged`, `lldp` (boolean, オプション)。

#### `update_network`
既存ネットワークコンテナを更新します。
- `id` (string, 必須): ネットワークID、一意名、またはIP。
- `name`, `ip`, `description`, `snmp_mode`, `community`, `user`, `password`, `url` (string, オプション)。
- `snmp_port` (`1`〜`65535`), `h_ports`, `x`, `y` (integer, オプション)。
- `arp_watch`, `unmanaged`, `lldp` (string, オプション: `"true"`, `"false"`, または空で維持)。

#### `delete_network`
ネットワークコンテナと接続されているラインを削除します。
- `id` (string, 必須): ネットワークID、一意名、またはIP。

#### `check_network`
ネットワークコンテナのSNMP/LLDPステータス確認を即時実行します。
- `id` (string, オプション): 対象ネットワーク。
- `all` (boolean, オプション): `true` で全ネットワークを対象。

#### `auto_connect_lines`
未接続のマップトポロジラインを自動推論して接続します。
- `mode` (string, オプション): `"strict"`（デフォルト）または `"speculative"`。

#### `auto_layout`
ノード座標を自動配置、または直前の自動配置を取り消します。
- `mode` (string, オプション): `"hierarchical"`（デフォルト）、`"cluster"`、`"categorized"`。
- `undo` (boolean, オプション): `true` で直前の自動配置を元に戻す。

---

### リアルタイム診断・SNMP・MIB・gNMI

#### `do_ping`
TWSNMPからICMP Pingを実行します。
- `target` (string, 必須): IPアドレス、ホスト名、ノードID、またはノード名。
- `size` (`1`〜`1500`, デフォルト `64`), `ttl` (`1`〜`255`, デフォルト `254`), `timeout` (`1`〜`10` 秒, デフォルト `3`)。

#### `diagnose_node`
統合ライブ診断（ノード解決、異常中ポーリング一覧、Ping、SNMP `sysDescr`/`sysObjectID`/`sysUpTime`/`sysName`/`sysLocation`、TCPポートスキャン＆バナー取得、HTTP/HTTPSレスポンス検査）を一括実行します。
- `target` (string, 必須): ノードID、ノード名、IP、またはホスト名。
- `ports` (string, オプション): カンマ区切りTCPポート（デフォルト `"22,80,443,8080,8443"`、最大32個）。
- `timeout` (integer, オプション): タイムアウト秒（`1`〜`10`, デフォルト `3`）。

#### `get_mib_tree`
読み込まれているSNMP MIBツリーを取得します。
- `root_oid` (string, オプション): サブツリー絞込用のOIDまたはオブジェクト名。
- `max_depth` (integer, オプション): 最大走査深さ（`0` は無制限）。

#### `search_mib`
MIB名、OID、または説明文に対して正規表現/部分一致検索を行います。
- `query` (string, 必須): 大文字小文字を区別しない検索パターン。
- `limit` (integer, オプション): 最大返却件数（`1`〜`500`, デフォルト `50`）。

#### `snmpwalk`
対象機器に対してSNMP Walkを実行します。管理対象ノードのID/名前/IPを指定し資格情報を省略した場合はノード設定（`SnmpPort` 含む）を自動適用します。
- `target` (string, 必須): IP、ホスト名、ノードID、またはノード名。
- `mib_object_name` (string, 必須): MIBオブジェクト名または数値OID。
- `snmp_mode` (string, オプション): `v2c`, `v3auth`, `v3authpriv`, `v3authprivex`, `v3sha256aes128`, `v3sha512aes256`。
- `community`, `user`, `password` (string, オプション), `port` (integer, オプション), `limit` (integer, オプション: `1`〜`50000`)。

#### `snmpset`
書き込み可能なSNMP MIB変数（`integer` または `string`）の値を設定します。
- `target` (string, 必須), `mib_object_name` (string, 必須), `type` (`"integer"` / `"string"`, 必須), `value` (string, 必須)。
- `snmp_mode`, `community`, `user`, `password`, `port` (オプション)。

#### `gnmi_capabilities`
ノードのgNMI Capabilities（対応エンコーディング、バージョン、YANGモデル）を取得します。
- `node_id` (string, 必須): ノードID、一意名、またはIP。
- `target` (string, オプション): `host:port` 上書き指定（既定はノードIP + `GNMIPort` または `57400`）。

#### `gnmi_get`
gNMIテレメトリーパスの値を取得します。
- `node_id` (string, 必須), `path` (string, 必須: 例 `/interfaces/interface`)。
- `target` (string, オプション), `encoding` (string, オプション: `json_ietf`, `json`, `proto`, `bytes`, `ascii`), `limit` (integer, オプション: `1`〜`5000`, デフォルト `500`)。

---

### IP・MAC・FDB・無線探索

#### `get_ip_address_list`
ARP/IP管理テーブルからIPアドレス一覧を取得します。
- `ip_filter`, `mac_filter`, `vendor_filter` (string, オプション): 正規表現フィルタ。
- `limit` (integer, オプション): `1`〜`2000`（デフォルト `500`）。

#### `get_ip_address_info`
指定IPアドレスの詳細情報（MAC、ベンダー、ロケーション、DNS名、関連ノード）を取得します。
- `ip` (string, 必須)。

#### `get_mac_address_list`
集約されたMACアドレス一覧（関連IP・ベンダー情報付き）を取得します。
- `mac_filter`, `ip_filter`, `vendor_filter` (string, オプション): 正規表現フィルタ。
- `limit` (integer, オプション): `1`〜`2000`（デフォルト `500`）。

#### `get_mac_address_info`
単一MACアドレスの詳細を取得します。
- `mac` (string, 必須)。

#### `get_wifi_ap_list`
検出されたWi-Fiアクセスポイント一覧を取得します（既定では履歴省略で軽量返却）。
- `ssid_filter`, `bssid_filter`, `host_filter`, `vendor_filter` (string, オプション)。
- `include_history` (boolean, オプション): `RSSI` 時系列履歴を含める場合は `true`（デフォルト `false`）。
- `limit` (integer, オプション): `1`〜`500`（デフォルト `100`）。

#### `get_bluetooth_device_list`
検出されたBluetooth/BLEデバイス一覧を取得します（既定では履歴省略で軽量返却）。
- `name_filter`, `address_filter`, `host_filter`, `vendor_filter` (string, オプション)。
- `public_address_only` (boolean, オプション): Publicアドレスのみ抽出する場合は `true`。
- `include_history` (boolean, オプション): `RSSI` 時系列履歴を含める場合は `true`（デフォルト `false`）。
- `limit` (integer, オプション): `1`〜`500`（デフォルト `100`）。

#### `get_fdb_table`
ネットワーク/スイッチコンテナからSNMP経由でライブFDB（MAC-ポート対応表）を取得します。
- `id` (string, 必須): ネットワークID（`NET:` 有無不問）、一意名、またはIP。

---

### システム状態・証明書

#### `get_system_status`
TWSNMP FKの稼働状態（バージョン、起動時刻、稼働時間、CPU数、Goroutine数、DBサイズ、データストアパス、Pingモード、`MCPTransport`/`MCPEndpoint`/`MCPStatus`/`MCPError`、ノード/ポーリング/ライン/ネットワーク/証明書件数サマリー、最新リソースモニター値、各レシーバー `syslogd`/`trapd`/`netflowd`/`sflowd`/`otel`/`mqtt`/`arpwatch`/`sshd`/`tcpd` の稼働状況）を取得します。
- 引数: なし。

#### `get_resource_monitor_list`
システムリソースモニターの履歴スナップショットを取得します。
- `limit` (integer, オプション): `1`〜`500`（デフォルト `100`）。

#### `get_server_certificate_list`
登録済みTLSサーバー証明書監視エントリの一覧を取得します。
- 引数: なし。

#### `update_server_certificate`
サーバー証明書監視エントリを新規登録（`id` 省略時）または更新し、即時再チェックをトリガーします。
- `id` (string, オプション), `target` (string, 新規時は必須), `port` (integer, オプション: `1`〜`65535`, 新規既定 `443`)。

#### `delete_server_certificate`
指定IDのサーバー証明書監視エントリを削除します。
- `id` (string, 必須)。

---

### ログ・フロー・OpenTelemetry・AI異常検知・レポート

日時範囲引数（`start_time`, `end_time`）は相対時間（`"-1h"`, `"-24h"`）、`"now"`、またはRFC3339等のタイムスタンプに対応しています。

#### `search_event_log`
イベントログを検索します。
- `node_filter`, `type_filter`, `level_filter` (`warn`, `low`, `high`, `debug`, `info`), `event_filter` (string, オプション)。
- `start_time`, `end_time` (string, オプション: デフォルト `-1h` 〜 `now`)。
- `limit` (integer, オプション: `1`〜`10000`, デフォルト `100`)。

#### `add_event_log`
TWSNMPにイベントログを追加します（種別は `"mcp"` として記録）。
- `event` (string, 必須): イベント内容。
- `level` (string, オプション): `info`（デフォルト）, `normal`, `warn`, `low`, `high`。
- `node` (string, オプション): 関連ノードのIDまたは名前。

#### `search_syslog`
Syslogを検索します。
- `level_filter`, `host_filter`, `tag_filter`, `message_filter` (string, オプション), `start_time`, `end_time`, `limit` (`1`〜`10000`, デフォルト `100`)。

#### `get_syslog_summary`
正規化されたSyslogパターンを出現回数上位順で集計します。
- `level_filter`, `host_filter`, `tag_filter`, `message_filter` (string, オプション), `start_time`, `end_time`。
- `top_n` (`5`〜`500`, デフォルト `5`), `max_scan` (`1000`〜`2000000`, デフォルト `200000`)。

#### `search_snmp_trap_log`
SNMPトラップログを検索します。
- `from_filter`, `trap_type_filter`, `variable_filter` (string, オプション), `start_time`, `end_time`, `limit` (`1`〜`10000`, デフォルト `100`)。

#### `search_netflow`
NetFlowレコードを検索します。
- `src_filter`, `dst_filter`, `protocol_filter` (string, オプション), `src_port`, `dst_port` (integer, オプション), `start_time`, `end_time`, `limit` (`1`〜`2000`, デフォルト `200`)。

#### `search_sflow`
sFlowフローサンプルを検索します（引数は `search_netflow` と共通）。

#### `search_sflow_counter`
sFlowカウンターサンプルを検索します。
- `remote_filter`, `type_filter` (string, オプション), `start_time`, `end_time`, `limit` (`1`〜`2000`, デフォルト `200`)。

#### `get_otel_metrics`
OpenTelemetryメトリクスを検索します。
- `host_filter`, `service_filter`, `name_filter` (string, オプション), `limit` (`1`〜`500`, デフォルト `100`)。

#### `get_otel_traces`
OpenTelemetryトレースサマリーを取得します。
- `service_filter` (string, オプション), `error_only` (boolean, オプション), `limit` (`1`〜`200`, デフォルト `50`)。

#### `get_otel_trace_detail`
指定 `trace_id` に属する全スパン詳細（親子関係、サービス、属性、イベント、エラー有無）を取得します。
- `trace_id` (string, 必須), `bucket` (string, オプション)。

#### `get_ai_anomaly_list`
AI異常検知スコア一覧と `AIConf` しきい値に基づく判定レベル（`high`, `low`, `warn`, `normal`）を取得します。
- `node_id` (string, オプション), `min_score` (number, オプション), `include_history` (boolean, デフォルト `false`), `limit` (`1`〜`1000`, デフォルト `100`)。

#### `get_sensor_list`
センサー一覧（`twWifiScan`, `twBlueScan`, `twpcap`, `twwinlog`, `twagentneo` など）を取得します。
- `type_filter`, `state_filter` (string, オプション), `include_history` (boolean, デフォルト `false` で `Stats`/`Monitors` 省略), `limit` (`1`〜`500`, デフォルト `100`)。

#### `get_report_summary`
ログ分析レポート（`wifiAP`, `blueDevice`, `envMonitor`, `powerMonitor`, `motionSensor`, `etherType`, `dnsq`, `radiusFlow`, `tlsFlow`, `winEventID`, `winLogon`, `winAccount`, `winKerberos`, `winPrivilege`, `winProcess`, `winTask`, `flow`, `server`, `fumble`, `syslogStats`, `trapStats`, `twAgentInfo`, `twAgentAlert`, `twAgentFIM`, `twAgentSCA`, `twAgentTCP`）を取得します。
- `report_type` (string, 必須), `filter` (string, オプション), `include_history` (boolean, デフォルト `false` で `RSSI`/`EnvData`/`Data` 省略), `limit` (`1`〜`200`, デフォルト `50`)。

#### `analyze_polling_correlation`
同一または異なるノード間の2つのポーリングログ時系列の相関関係を、統計および機械学習アルゴリズムで詳細分析します。
- `polling_id_1`, `polling_id_2` (string, 必須): 分析対象のポーリングID。
- `vars_1`, `vars_2` (string, オプション): カンマ区切りの数値変数名（空で自動選択）。
- `algorithm` (string, オプション): `all`（デフォルト）, `pearson`, `spearman`, `kendall`, `crosscorr`, `dcor`, `mlp`, `lstm`, `knn`。
- `bucket_sec` (integer, オプション): 集計時間バケット幅（秒、`0` で自動）。
- `period_hours` (integer, オプション): 集計対象期間（時間、デフォルト `168`、`-1` で全期間）。
- `max_lag` (integer, オプション): 相互相関の最大ラグ数（デフォルト `12`、最大 `60`）。
- `include_state` (boolean, オプション): ポーリング状態値（`1.0`=正常, `0.5`=不明, `0.0`=障害）を含めるか。

#### `discover_polling_correlations`
ノードを横断して相関関係のあるポーリングペアを自動探索・ランキングします。
- `base_polling_id` (string, オプション): 基準となるポーリングID（空で全体 N対N 探索）。
- `algorithm` (string, オプション): `pearson`（デフォルト）, `spearman`, `kendall`, `crosscorr`, `dcor`, `mlp`, `knn`。
- `bucket_sec` (integer, オプション): 集計時間バケット幅（秒、`0` で自動）。
- `period_hours` (integer, オプション): 集計対象期間（時間、デフォルト `168`、`-1` で全期間）。
- `min_abs_score` (number, オプション): 最小絶対相関係数しきい値（デフォルト `0.7`）。
- `min_samples` (integer, オプション): 最小重複サンプル数（デフォルト `10`）。
- `cross_node_only` (boolean, オプション): 異なるノード間のペアのみ抽出する場合は `true`。
- `include_state` (boolean, オプション): ポーリング状態値を含めるか。
- `max_lag` (integer, オプション): `crosscorr` 使用時の最大ラグ数（デフォルト `6`）。
- `limit` (integer, オプション): 返却ペア数上限（`1`〜`500`, デフォルト `30`）。
