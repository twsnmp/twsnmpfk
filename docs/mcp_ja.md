# TWSNMP FK MCPサーバー仕様書

このドキュメントは、`backend/mcp.go`、`backend/mcp_tools.go`、`backend/mcp_prompts.go`のソースコードに基づき、TWSNMP FK MCP（Model Context Protocol）サーバーの仕様を概説します。

## 1. 概要

MCPサーバーは、AIエージェントがTWSNMP FK監視システムと対話するためのインターフェースを提供します。エージェントは、定義されたツールとプロンプトを通じて、監視データの取得、アクションの実行、およびシステム情報へのアクセスが可能です。

## 2. トランスポートとエンドポイント

サーバーは`datastore.MapConf.MCPTransport`の設定に基づいて起動します。以下の2つのトランスポートメカニズムをサポートしています。

- **Server-Sent Events (SSE)**:
  - `MCPTransport`が`"sse"`に設定されている場合に有効になります。
  - エンドポイント: `/sse` および `/message`。
- **ストリーマブルHTTP**:
  - その他のトランスポート設定（例: `"http"`）の場合に有効になります。
  - エンドポイント: `/mcp`。

サーバーは`datastore.MapConf.MCPEndpoint`で指定されたアドレスで待機します。

### TLSセキュリティ

サーバー証明書（`datastore.MCPCert`）と秘密鍵（`datastore.MCPKey`）が提供されている場合、TLSが自動的に有効になります。TLS 1.3と、安全な暗号スイートの制限付きセットを使用します。

## 3. 認証

MCPサーバーへのアクセスは、以下の2つのメカニズムによって制御されます。両方が設定されている場合は、両方の条件を満たす必要があります。

- **トークン認証**: `datastore.MapConf.MCPToken`が設定されている場合、リクエストの`Authorization`ヘッダーにこのトークンが含まれている必要があります。
- **IPアドレスACL**: `datastore.MapConf.MCPFrom`が（IPアドレスのカンマ区切りリストとして）設定されている場合、リクエスト元のIPアドレスが許可リストに含まれている必要があります。

## 4. ツール（関数）

サーバーはエージェント向けに以下のツールを公開しています。

---

### `get_node_list`
- **説明**: TWSNMPからノードのリストを取得します。
- **パラメータ**:
  - `name_filter` (string, 任意): ノード名でフィルタリングするための正規表現。
  - `ip_filter` (string, 任意): ノードのIPアドレスでフィルタリングするための正規表現。
  - `state_filter` (string, 任意): ノードの状態でフィルタリングするための正規表現 (`normal`, `warn`, `low`, `high`, `repair`, `unknown`)。
- **戻り値**: ノードオブジェクトのJSON配列。

---

### `get_network_list`
- **説明**: TWSNMPからネットワークのリストを取得します。
- **パラメータ**:
  - `name_filter` (string, 任意): ネットワーク名でフィルタリングするための正規表現。
  - `ip_filter` (string, 任意): ネットワークのIPアドレスでフィルタリングするための正規表現。
- **戻り値**: ネットワークオブジェクトのJSON配列。

---

### `get_polling_list`
- **説明**: TWSNMPからポーリングのリストを取得します。
- **パラメータ**:
  - `type_filter` (string, 任意): ポーリング種別でフィルタリングするための正規表現 (`ping`, `tcp`, `http`, `dns`, `twsnmp`, `syslog` など)。
  - `name_filter` (string, 任意): ポーリング名でフィルタリングするための正規表現。
  - `node_name_filter` (string, 任意): ポーリングに関連付けられたノード名でフィルタリングするための正規表現。
  - `state_filter` (string, 任意): ポーリングの状態でフィルタリングするための正規表現 (`normal`, `warn`, `low`, `high`, `repair`, `unknown`)。
- **戻り値**: ポーリングオブジェクトのJSON配列。

---

### `get_polling_log`
- **説明**: 特定のポーリングIDのポーリングログを取得します。
- **パラメータ**:
  - `id` (string, 必須): ポーリングのID。
  - `limit` (int, 任意): 取得するログエントリの最大数（1-2000、デフォルト100）。
- **戻り値**: ポーリングログエントリのJSON配列。

---

### `do_ping`
- **説明**: ターゲットに対してpingを実行します。
- **パラメータ**:
  - `target` (string, 必須): ターゲットのIPアドレスまたはホスト名。
  - `size` (int, 任意): パケットサイズ（1-1500、デフォルト64）。
  - `ttl` (int, 任意): IPパケットのTTL（1-255、デフォルト254）。
  - `timeout` (int, 任意): タイムアウト（秒）（1-10、デフォルト3）。
- **戻り値**: pingの結果を含むJSONオブジェクト。

---

### `get_mib_tree`
- **説明**: TWSNMPからMIBツリーを取得します。
- **パラメータ**: なし。
- **戻り値**: MIBツリー構造を表すJSONオブジェクト。

---

### `snmpwalk`
- **説明**: SNMPウォークを実行します。
- **パラメータ**:
  - `target` (string, 必須): ターゲットのIPアドレスまたはノード名。
  - `mib_object_name` (string, 必須): ウォークを開始するMIBオブジェクト名またはOID。
  - `community` (string, 任意): SNMPv2cのコミュニティ文字列。指定しない場合は、ノードに設定されたコミュニティが使用されます。
  - `user` (string, 任意): SNMPv3のユーザー名。
  - `password` (string, 任意): SNMPv3のパスワード。
  - `snmp_mode` (string, 任意): SNMPモード (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`)。
- **戻り値**: 名前と値を持つMIBオブジェクトのJSON配列。

---

### `add_node`
- **説明**: TWSNMPに新しいノードを追加します。
- **パラメータ**:
  - `name` (string, 必須): ノード名。
  - `ip` (string, 必須): ノードのIPアドレス。
  - `icon` (string, 任意): ノードのアイコン（デフォルト: `desktop`）。
  - `description` (string, 任意): ノードの説明。
  - `x` (int, 任意): マップ上のX座標（1-1000、デフォルト64）。
  - `y` (int, 任意): マップ上のY座標（1-1000、デフォルト64）。
- **戻り値**: 新しく作成されたノードのJSONオブジェクト。`PING`ポーリングが自動的に追加されます。

---

### `update_node`
- **説明**: ノードのプロパティ（名前、IP、位置、説明、アイコン）を更新します。
- **パラメータ**:
  - `id` (string, 必須): ノードID、現在の名前、または現在のIPアドレス。
  - `name` (string, 任意): 新しいノード名。
  - `ip` (string, 任意): 新しいIPアドレス。
  - `icon` (string, 任意): 新しいアイコン。
  - `description` (string, 任意): 新しい説明。
  - `x` (int, 任意): 新しいX座標。
  - `y` (int, 任意): 新しいY座標。
- **戻り値**: 更新されたノードのJSONオブジェクト。

---

### `get_ip_address_list`
- **説明**: ARPによって発見されたIPアドレスのリストを取得します。
- **パラメータ**: なし。
- **戻り値**: IP、MAC、関連ノード、ベンダー、タイムスタンプを含むオブジェクトのJSON配列。

---

### `get_resource_monitor_list`
- **説明**: TWSNMP自体のリソース監視データ（CPU、メモリなど）のリストを取得します。
- **パラメータ**: なし。
- **戻り値**: リソース使用状況スナップショットのJSON配列。

---

### `search_event_log`
- **説明**: イベントログを検索します。
- **パラメータ**:
  - `node_filter` (string, 任意): ノード名でフィルタリングするための正規表現。
  - `type_filter` (string, 任意): イベント種別でフィルタリングするための正規表現。
  - `level_filter` (string, 任意): イベントレベルでフィルタリングするための正規表現 (`info`, `warn`, `low`, `high`, `debug`)。
  - `event_filter` (string, 任意): イベントメッセージの内容でフィルタリングするための正規表現。
  - `start_time` (string, 任意): 検索の開始時刻（例: "2023-10-27 00:00:00" または "-1h"）。デフォルトは "-1h"。
  - `end_time` (string, 任意): 検索の終了時刻（例: "2023-10-27 23:59:59" または "now"）。デフォルトは "now"。
  - `limit_log_count` (int, 任意): 返すログの最大数（100-10000、デフォルト100）。
- **戻り値**: イベントログエントリのJSON配列。

---

### `search_syslog`
- **説明**: syslogを検索します。
- **パラメータ**:
  - `level_filter` (string, 任意): レベルでフィルタリングするための正規表現 (`info`, `warn`, `low`, `high`, `debug`)。
  - `host_filter` (string, 任意): ホスト名でフィルタリングするための正規表現。
  - `tag_filter` (string, 任意): syslogタグでフィルタリングするための正規表現。
  - `message_filter` (string, 任意): メッセージ内容でフィルタリングするための正規表現。
  - `start_time` (string, 任意): 開始時刻（デフォルト: "-1h"）。
  - `end_time` (string, 任意): 終了時刻（デフォルト: "now"）。
  - `limit_log_count` (int, 任意): 返すログの最大数（100-10000、デフォルト100）。
- **戻り値**: syslogエントリのJSON配列。

---

### `get_syslog_summary`
- **説明**: syslogパターンの要約を取得します。
- **パラメータ**:
  - `level_filter`, `host_filter`, `tag_filter`, `message_filter`, `start_time`, `end_time`: `search_syslog`と同様。
  - `top_n` (int, 任意): 返す上位パターンの数（5-500、デフォルト5）。
- **戻り値**: syslogパターンとそのカウントのJSON配列。

---

### `search_snmp_trap_log`
- **説明**: SNMPトラップログを検索します。
- **パラメータ**:
  - `from_filter` (string, 任意): 送信元アドレスでフィルタリングするための正規表現。
  - `trap_type_filter` (string, 任意): トラップタイプでフィルタリングするための正規表現。
  - `variable_filter` (string, 任意): トラップ変数の内容でフィルタリングするための正規表現。
  - `start_time` (string, 任意): 開始時刻（デフォルト: "-1h"）。
  - `end_time` (string, 任意): 終了時刻（デフォルト: "now"）。
  - `limit` (int, 任意): 返すログの最大数（100-10000、デフォルト100）。
- **戻り値**: SNMPトラップログエントリのJSON配列。

---

### `get_server_certificate_list`
- **説明**: 監視対象のサーバー証明書のリストを取得します。
- **パラメータ**: なし。
- **戻り値**: 証明書監視エントリのJSON配列。

---

### `add_event_log`
- **説明**: TWSNMPにイベントログを追加します。
- **パラメータ**:
  - `level` (string, 任意): イベントレベル（`info`, `normal`, `warn`, `low`, `high`）。デフォルトは`info`。
  - `node` (string, 任意): イベントに関連付けられたノードの名前。
  - `event` (string, 必須): イベントログの内容。
- **戻り値**: 成功時に文字列"ok"を返します。

---

### `get_ip_address_info`
- **説明**: IPアドレスに関する情報（DNS、管理対象ノード、位置情報、RDAP）を取得します。
- **パラメータ**:
  - `ip` (string, 必須): 問い合わせるIPアドレス。
- **戻り値**: IPアドレスに関する集約された情報を含むJSONオブジェクト。

---

### `get_mac_address_info`
- **説明**: MACアドレスに関する情報（IP、管理対象ノード、ベンダー）を取得します。
- **パラメータ**:
  - `mac` (string, 必須): 問い合わせるMACアドレス。
- **戻り値**: MACアドレスに関する集約された情報を含むJSONオブジェクト。

## 5. プロンプト

