# MCPサーバー ツール仕様

このドキュメントは、`mcp.go`で定義されているTWSNMP-FKのMCP（Mark3Labs Control Plane）サーバーが提供するツールの一覧とその仕様を記述したものです。

## 共通事項

- すべてのツールは、結果をJSON形式の文字列で返します。
- フィルタ系のパラメータ（`..._filter`）は、正規表現による指定が可能です。空文字列を指定した場合は、フィルタリングを行いません。
- 時刻範囲を指定するパラメータ（`start_time`, `end_time`）では、`2025/07/25 10:00:00`のような絶対時刻、または`-1h`や`30m`のような現在からの相対的な期間を指定できます。

## ツール一覧

### 1. `get_node_list`

TWSNMPに登録されているノードのリストを取得します。

- **説明**: TWSNMPからノードリストを取得します。
- **パラメータ**:
    - `name_filter` (string): ノード名でフィルタリングするための正規表現。
    - `ip_filter` (string): IPアドレスでフィルタリングするための正規表現。
    - `state_filter` (string): 状態（`normal`, `warn`, `low`, `high`, `repair`, `unknown`）でフィルタリングするための正規表現。
- **出力**: ノード情報の配列（JSON）
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "IP": "...",
        "MAC": "...",
        "State": "normal",
        "X": 100,
        "Y": 200,
        "Icon": "desktop",
        "Descrption": "..."
      }
    ]
    ```

### 2. `get_network_list`

TWSNMPに登録されているネットワークのリストを取得します。

- **説明**: TWSNMPからネットワークリストを取得します。
- **パラメータ**:
    - `name_filter` (string): ネットワーク名でフィルタリングするための正規表現。
    - `ip_filter` (string): IPアドレスでフィルタリングするための正規表現。
- **出力**: ネットワーク情報の配列（JSON）
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "IP": "...",
        "Ports": ["port1=up", "port2=down"],
        "X": 100,
        "Y": 200,
        "Descrption": "..."
      }
    ]
    ```

### 3. `get_polling_list`

TWSNMPに登録されているポーリングのリストを取得します。

- **説明**: TWSNMPからポーリングリストを取得します。
- **パラメータ**:
    - `type_filter` (string): ポーリング種別（`ping`, `tcp`, `http`など）でフィルタリングするための正規表現。
    - `state_filter` (string): 状態（`normal`, `warn`など）でフィルタリングするための正規表現。
    - `name_filter` (string): ポーリング名でフィルタリングするための正規表現。
    - `node_name_filter` (string): ポーリング対象のノード名でフィルタリングするための正規表現。
- **出力**: ポーリング情報の配列（JSON）
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "NodeID": "...",
        "NodeName": "...",
        "Type": "ping",
        "Level": "normal",
        "State": "normal",
        "LastTime": "2025-07-25T10:00:00Z",
        "Result": { "...": "..." }
      }
    ]
    ```

### 4. `do_ping`

指定したターゲットに対してPingを実行します。

- **説明**: Pingを実行します。
- **パラメータ**:
    - `target` (string, **必須**): Pingのターゲット（IPアドレスまたはホスト名）。
    - `size` (number): パケットサイズ（デフォルト: 64, 最小: 64, 最大: 1500）。
    - `ttl` (number): TTL（デフォルト: 254, 最小: 1, 最大: 254）。
    - `timeout` (number): タイムアウト（秒）（デフォルト: 2, 最小: 1, 最大: 10）。
- **出力**: Ping実行結果（JSON）
    ```json
    {
      "Result": "Success",
      "Time": "2025-07-25T10:00:00Z",
      "RTT": "1.234ms",
      "RTTNano": 1234000,
      "Size": 64,
      "TTL": 64,
      "ResponceFrom": "192.168.1.1",
      "Location": "..."
    }
    ```

### 5. `get_MIB_tree`

TWSNMPが保持しているMIBツリーの情報を取得します。

- **説明**: TWSNMPからMIBツリーを取得します。
- **パラメータ**: なし
- **出力**: MIBツリー情報（JSON）

### 6. `snmpwalk`

指定したターゲットに対してSNMP Walkを実行します。

- **説明**: SNMP Walkツール。
- **パラメータ**:
    - `target` (string, **必須**): SNMP Walkのターゲット（IP、ホスト名、ノード名）。
    - `mib_object_name` (string, **必須**): MIBオブジェクト名。
    - `community` (string): SNMPv2cのコミュニティ名。
    - `user` (string): SNMPv3のユーザー名。
    - `password` (string): SNMPv3のパスワード。
    - `snmpmode` (string): SNMPモード (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`)。
- **出力**: SNMP Walkの結果の配列（JSON）
    ```json
    [
      {
        "Name": "sysDescr.0",
        "Value": "..."
      }
    ]
    ```

### 7. `add_node`

TWSNMPに新しいノードを追加します。

- **説明**: TWSNMPにノードを追加します。
- **パラメータ**:
    - `name` (string, **必須**): ノード名。
    - `ip` (string, **必須**): IPアドレス。
    - `icon` (string): アイコン名（デフォルト: `desktop`）。
    - `description` (string): 説明。
    - `x` (number): X座標（最小: 64, 最大: 1000）。
    - `y` (number): Y座標（最小: 64, 最大: 1000）。
- **出力**: 追加されたノード情報（JSON）

### 8. `update_node`

既存のノード情報を更新します。

- **説明**: ノードの名前、IP、位置、説明、アイコンを更新します。
- **パラメータ**:
    - `id` (string, **必須**): 更新対象のノードID。
    - `name` (string): 新しいノード名。
    - `ip` (string): 新しいIPアドレス。
    - `icon` (string): 新しいアイコン名。
    - `description` (string): 新しい説明。
    - `x` (number): 新しいX座標。
    - `y` (number): 新しいY座標。
- **出力**: 更新されたノード情報（JSON）

### 9. `get_ip_address_list`

TWSNMPがARPログなどから収集したIPアドレスのリストを取得します。

- **説明**: TWSNMPからIPアドレスリストを取得します。
- **パラメータ**: なし
- **出力**: IPアドレス情報の配列（JSON）
    ```json
    [
      {
        "IP": "...",
        "MAC": "...",
        "Node": "...",
        "Vendor": "...",
        "FirstTime": "2025-07-25T10:00:00Z",
        "LastTime": "2025-07-25T11:00:00Z"
      }
    ]
    ```

### 10. `get_resource_monitor_list`

TWSNMPサーバー自体のリソース使用状況のリストを取得します。

- **説明**: TWSNMPからリソースモニターリストを取得します。
- **パラメータ**: なし
- **出力**: リソース情報の配列（JSON）
    ```json
    [
      {
        "Time": "2025-07-25T10:00:00Z",
        "CPUUsage": "10.50%",
        "MemoryUsage": "25.20%",
        "SwapUsage": "0.00%",
        "DiskUsage": "45.80%",
        "Load": "0.75"
      }
    ]
    ```

### 11. `search_event_log`

イベントログを検索します。

- **説明**: TWSNMPからイベントログを検索します。
- **パラメータ**:
    - `node_filter` (string): ノード名でのフィルタ。
    - `type_filter` (string): 種別でのフィルタ。
    - `level_filter` (string): レベル (`warn`, `info`など)でのフィルタ。
    - `event_filter` (string): イベント内容でのフィルタ。
    - `limit_log_count` (number): 取得するログの最大件数（デフォルト: 100, 最大: 10000）。
    - `start_time` (string): 検索開始時刻（デフォルト: `-1h`）。
    - `end_time` (string): 検索終了時刻（デフォルト: 現在時刻）。
- **出力**: イベントログの配列（JSON）
    ```json
    [
      {
        "Time": "2025-07-25T10:00:00Z",
        "Type": "...",
        "Level": "warn",
        "Node": "...",
        "Event": "..."
      }
    ]
    ```

### 12. `search_syslog`

Syslogを検索します。

- **説明**: TWSNMPからsyslogを検索します。
- **パラメータ**:
    - `host_filter` (string): ホスト名でのフィルタ。
    - `tag_filter` (string): タグでのフィルタ。
    - `level_filter` (string): レベルでのフィルタ。
    - `message_filter` (string): メッセージ内容でのフィルタ。
    - `limit_log_count` (number): 取得するログの最大件数（デフォルト: 100, 最大: 10000）。
    - `start_time` (string): 検索開始時刻（デフォルト: `-1h`）。
    - `end_time` (string): 検索終了時刻（デフォルト: 現在時刻）。
- **出力**: Syslogエントリの配列（JSON）
    ```json
    [
      {
        "Time": "...",
        "Level": "info",
        "Host": "...",
        "Type": "...",
        "Tag": "...",
        "Message": "...",
        "Severity": 6,
        "Facility": 1
      }
    ]
    ```

### 13. `get_syslog_summary`

Syslogをパターンで集計し、サマリーを取得します。

- **説明**: TWSNMPからsyslogサマリーを取得します。
- **パラメータ**:
    - `host_filter`, `tag_filter`, `level_filter`, `message_filter`: `search_syslog`と同様。
    - `top_n` (number): 上位何件のパターンを表示するか（デフォルト: 10, 最大: 100）。
    - `start_time`, `end_time`: `search_syslog`と同様。
- **出力**: Syslogサマリーの配列（JSON）
    ```json
    [
      {
        "Pattern": "...",
        "Count": 123
      }
    ]
    ```

### 14. `search_snmp_trap_log`

SNMP Trapログを検索します。

- **説明**: TWSNMPからSNMPトラップログを検索します。
- **パラメータ**:
    - `from_filter` (string): 送信元アドレスでのフィルタ。
    - `trap_type_filter` (string): トラップ種別でのフィルタ。
    - `variable_filter` (string): 変数（Variables）の内容でのフィルタ。
    - `limit_log_count`, `start_time`, `end_time`: `search_event_log`と同様。
- **出力**: SNMP Trapログの配列（JSON）
    ```json
    [
      {
        "Time": "...",
        "FromAddress": "...",
        "TrapType": "linkDown",
        "Variables": "..."
      }
    ]
    ```

### 15. `get_server_certificate_list`

監視対象のサーバー証明書のリストを取得します。

- **説明**: TWSNMPからサーバー証明書リストを取得します。
- **パラメータ**: なし
- **出力**: サーバー証明書情報の配列（JSON）
    ```json
    [
      {
        "State": "valid",
        "Server": "example.com",
        "Port": 443,
        "Subject": "CN=example.com,...",
        "Issuer": "...",
        "SerialNumber": "...",
        "Verify": true,
        "NotAfter": "2026-07-25T10:00:00Z",
        "NotBefore": "2024-07-25T10:00:00Z",
        "Error": "",
        "FirstTime": "...",
        "LastTime": "..."
      }
    ]
    ```
