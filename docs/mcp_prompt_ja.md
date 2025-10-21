# MCPサーバープロンプト仕様

このドキュメントは、`backend/mcp_prompts.go`で定義されているMCPサーバーで利用可能なプロンプトの仕様を概説します。

## プロンプト

### 1. `get_node_list`

- **タイトル:** フィルタ付きでノードリストを取得
- **説明:** TWSNMPに登録されているノードのリストをフィルタ付きで取得します。
- **引数:**
  - `state_filter` (オプション): ノードの状態フィルタ。有効な状態は `normal`, `repair`, `warn`, `low`, `high`, `unknown` です。
  - `name_filter` (オプション): ノード名フィルタ。
  - `ip_filter` (オプション): ノードIPアドレスフィルタ。

### 2. `add_node`

- **タイトル:** TWSNMPに新しいノードを追加
- **説明:** TWSNMPに新しいノードを追加します。
- **引数:**
  - `name` (必須): 新しいノードの名前。
  - `ip` (必須): 新しいノードのIPアドレス。
  - `icon` (オプション): ノードのアイコン。有効なアイコンは `desktop`, `laptop`, `server`, `cloud`, `router`, `ip` です。
  - `description` (オプション): 新しいノードの説明。
  - `position` (オプション): 新しいノードの位置 (例: `x=100,y=200`)。

### 3. `update_node`

- **タイトル:** TWSNMPのノードを更新
- **説明:** TWSNMP上の既存のノードを更新します。
- **引数:**
  - `id` (必須): 更新するノードのID、名前、またはIPアドレス。
  - `name` (オプション): ノードの新しい名前。
  - `ip` (オプション): ノードの新しいIPアドレス。
  - `icon` (オプション): ノードの新しいアイコン。有効なアイコンは `desktop`, `laptop`, `server`, `cloud`, `router`, `ip` です。
  - `description` (オプション): ノードの新しい説明。
  - `position` (オプション): ノードの新しい位置 (例: `x=100,y=200`)。

### 4. `get_network_list`

- **タイトル:** フィルタ付きでネットワークノードリストを取得
- **説明:** TWSNMPに登録されているネットワークノードのリストをフィルタ付きで取得します。
- **引数:**
  - `name_filter` (オプション): ネットワークノード名フィルタ。
  - `ip_filter` (オプション): ネットワークノードIPアドレスフィルタ。

### 5. `get_polling_list`

- **タイトル:** フィルタ付きでポーリングリストを取得
- **説明:** TWSNMPに登録されているポーリングのリストをフィルタ付きで取得します。
- **引数:**
  - `type_filter` (オプション): ポーリングタイプフィルタ。有効なタイプは `ping`, `snmp`, `syslog`, `http`, `tcp` です。
  - `name_filter` (オプション): ポーリング名フィルタ。
  - `node_name_filter` (オプション): ノード名フィルタ。
  - `state_filter` (オプション): ポーリング状態フィルタ。有効な状態は `normal`, `repair`, `warn`, `low`, `high`, `unknown` です。

### 6. `get_polling_log`

- **タイトル:** ポーリングログを取得
- **説明:** TWSNMPからポーリングログを取得します。
- **引数:**
  - `id` (必須): ポーリングのID。
  - `limit` (オプション): 取得するログの最大数 (デフォルトは100)。

### 7. `do_ping`

- **タイトル:** pingを実行
- **説明:** TWSNMPからターゲットにpingを実行します。
- **引数:**
  - `target` (必須): pingのターゲット (IP、ノード名、またはホスト名)。
  - `size` (オプション): pingパケットサイズ (デフォルトは64)。
  - `ttl` (オプション): pingパケットTTL (デフォルトは254)。
  - `timeout` (オプション): pingのタイムアウト (秒) (デフォルトは3)。

### 8. `snmpwalk`

- **タイトル:** snmpwalkを実行
- **説明:** TWSNMPからターゲットにsnmpwalkを実行します。
- **引数:**
  - `target` (必須): SNMPウォークのターゲット (IP、ノード名、またはホスト名)。
  - `mib_object_name` (必須): 取得するMIBオブジェクト名。
  - `snmp_mode` (オプション): SNMPモード。有効なモードは `v2c`, `v3auth`, `v3authpriv`, `v3authprivex` です。
  - `community` (オプション): v2cモードのコミュニティ名。
  - `user` (オプション): v3モードのユーザー名。
  - `password` (オプション): v3モードのパスワード。

### 9. `search_event_log`

- **タイトル:** フィルタ付きでイベントログを検索
- **説明:** フィルタ付きでイベントログを検索します。
- **引数:**
  - `type_filter` (オプション): イベントタイプフィルタ。有効なタイプは `system`, `polling`, `arpwatch`, `mcp` です。
  - `node_filter` (オプション): ノードフィルタ。
  - `level_name_filter` (オプション): レベルフィルタ。有効なレベルは `info`, `normal`, `warn`, `low`, `high` です。
  - `event_filter` (オプション): イベントフィルタ。
  - `start_time` (オプション): 検索の開始時刻 (デフォルトは-1h)。
  - `end_time` (オプション): 検索の終了日時 (デフォルトはnow)。
  - `limit` (オプション): 検索するログの最大数。

### 10. `search_syslog`

- **タイトル:** フィルタ付きでsyslogを検索
- **説明:** フィルタ付きでsyslogを検索します。
- **引数:**
  - `level_filter` (オプション): レベルフィルタ。有効なレベルは `warn`, `low`, `high`, `debug`, `info` です。
  - `host_filter` (オプション): 送信元ホストフィルタ。
  - `tag_filter` (オプション): Syslogタグフィルタ。
  - `message_filter` (オプション): Syslogメッセージフィルタ。
  - `start_time` (オプション): 検索の開始時刻 (デフォルトは-1h)。
  - `end_time` (オプション): 検索の終了日時 (デフォルトはnow)。
  - `limit` (オプション): 検索するsyslogの最大数。

### 11. `get_syslog_summary`

- **タイトル:** フィルタ付きでsyslogサマリーを取得
- **説明:** フィルタ付きでsyslogのサマリーを取得します。
- **引数:**
  - `level_filter` (オプション): レベルフィルタ。有効なレベルは `warn`, `low`, `high`, `debug`, `info` です。
  - `host_filter` (オプション): 送信元ホストフィルタ。
  - `tag_filter` (オプション): Syslogタグフィルタ。
  - `message_filter` (オプション): Syslogメッセージフィルタ。
  - `start_time` (オプション): 検索の開始時刻 (デフォルトは-1h)。
  - `end_time` (オプション): 検索の終了日時 (デフォルトはnow)。
  - `top_n` (オプション): 返すトップsyslogサマリーエントリの数。

### 12. `search_snmp_trap_log`

- **タイトル:** フィルタ付きでSNMPトラップログを検索
- **説明:** TWSNMPのSNMPトラップログをフィルタ付きで検索します。
- **引数:**
  - `from_filter` (オプション): トラップ送信元フィルタ。
  - `trap_type_filter` (オプション): トラップタイプフィルタ。
  - `variable_filter` (オプション): トラップ変数フィルタ。
  - `start_time` (オプション): 検索の開始時刻 (デフォルトは-1h)。
  - `end_time` (オプション): 検索の終了日時 (デフォルトはnow)。
  - `limit` (オプション): 検索するSNMPトラップログの数。

### 13. `get_mib_tree`

- **タイトル:** TWSNMPのMIBツリーを取得
- **説明:** `get_mib_tree`ツールを使用してTWSNMPのMIBツリーを取得します。
- **引数:** なし。

### 14. `get_ip_address_list`

- **タイトル:** TWSNMPで管理されているIPアドレスのリストを取得
- **説明:** `get_ip_address_list`ツールを使用してTWSNMPで管理されているIPアドレスのリストを取得します。
- **引数:** なし。

### 15. `get_resource_monitor_list`

- **タイトル:** TWSNMPのリソースモニター情報を取得
- **説明:** `get_resource_monitor_list`ツールを使用してTWSNMPのリソースモニター情報を取得します。
- **引数:** なし。

### 16. `get_server_certificate_list`

- **タイトル:** TWSNMPで管理されているサーバー証明書のリストを取得
- **説明:** `get_server_certificate_list`ツールを使用してTWSNMPで管理されているサーバー証明書のリストを取得します。
- **引数:** なし。

### 17. `add_event_log`

- **タイトル:** TWSNMPにイベントログを追加
- **説明:** TWSNMPにイベントログを追加します。
- **引数:**
  - `level` (必須): イベントログのレベル。有効なレベルは `info`, `normal`, `warn`, `low`, `high` です。
  - `node` (オプション): イベントログのノード名。
  - `event` (必須): イベントログの内容。

### 18. `get_ip_address_info`

- **タイトル:** IPアドレス情報を取得
- **説明:** IPアドレスに関する情報を取得します。
- **引数:**
  - `ip` (必須): 情報を取得するIPアドレス。

### 19. `get_mac_address_info`

- **タイトル:** MACアドレス情報を取得
- **説明:** MACアドレスに関する情報を取得します。
- **引数:**
  - `mac` (必須): 情報を取得するMACアドレス。
