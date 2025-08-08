# MCPサーバーツール

このドキュメントは、TWSNMP FK MCPサーバーで利用可能なツールの仕様について概説します。

## ツール

### 1. `get_node_list`

**説明:** TWSNMPからノードのリストを取得します。

**パラメータ:**

- `name_filter` (string, optional): 名前でノードをフィルタリングするための正規表現。省略した場合、すべてのノードが返されます。
- `ip_filter` (string, optional): IPアドレスでノードをフィルタリングするための正規表現。省略した場合、すべてのノードが返されます。
- `state_filter` (string, optional): 状態でノードをフィルタリングするための正規表現。省略した場合、すべてのノードが返されます。有効な状態は "normal", "warn", "low", "high", "repair", "unknown" です。

**出力:** 以下のプロパティを持つノードオブジェクトのJSON配列:
- `ID` (string): ノードID。
- `Name` (string): ノード名。
- `IP` (string): ノードのIPアドレス。
- `MAC` (string): ノードのMACアドレス。
- `State` (string): ノードの状態。
- `X` (int): マップ上のノードのX座標。
- `Y` (int): マップ上のノードのY座標。
- `Icon` (string): ノードのアイコン。
- `Descrption` (string): ノードの説明。

### 2. `get_network_list`

**説明:** TWSNMPからネットワークのリストを取得します。

**パラメータ:**

- `name_filter` (string, optional): 名前でネットワークをフィルタリングするための正規表現。省略した場合、すべてのネットワークが返されます。
- `ip_filter` (string, optional): IPアドレスでネットワークをフィルタリングするための正規表現。省略した場合、すべてのネットワークが返されます。

**出力:** 以下のプロパティを持つネットワークオブジェクトのJSON配列:
- `ID` (string): ネットワークID。
- `Name` (string): ネットワーク名。
- `IP` (string): ネットワークのIPアドレス。
- `Ports` (array of strings): ネットワークのポートとその状態のリスト。
- `X` (int): マップ上のネットワークのX座標。
- `Y` (int): マップ上のネットワークのY座標。
- `Descrption` (string): ネットワークの説明。

### 3. `get_polling_list`

**説明:** TWSNMPからポーリングのリストを取得します。

**パラメータ:**

- `type_filter` (string, optional): タイプでポーリングをフィルタリングするための正規表現。有効なタイプは "ping", "tcp", "http", "dns", "twsnmp", "syslog" です。
- `state_filter` (string, optional): 状態でポーリングをフィルタリングするための正規表現。有効な状態は "normal", "warn", "low", "high", "repair", "unknown" です。
- `name_filter` (string, optional): 名前でポーリングをフィルタリングするための正規表現。
- `node_name_filter` (string, optional): ノード名でポーリングをフィルタリングするための正規表現。

**出力:** 以下のプロパティを持つポーリングオブジェクトのJSON配列:
- `ID` (string): ポーリングID。
- `Name` (string): ポーリング名。
- `NodeID` (string): ポーリング対象のノードのID。
- `NodeName` (string): ポーリング対象のノードの名前。
- `Type` (string): ポーリングタイプ。
- `Level` (string): ポーリングレベル。
- `State` (string): ポーリングの状態。
- `Logging` (boolean): ポーリングのロギングが有効かどうか。
- `LastTime` (string): ポーリングが最後に実行された時刻。
- `Result` (object): 最後のポーリングの結果。

### 4. `get_polling_log`

**説明:** 特定のポーリングのポーリングログを取得します。

**パラメータ:**

- `id` (string, required): ポーリングのID。
- `limit` (number, optional, default: 100): 取得するログエントリの最大数 (1-2000)。

**出力:** 以下のプロパティを持つポーリングログオブジェクトのJSON配列:
- `Time` (string): ログエントリのタイムスタンプ。
- `State` (string): ログエントリ時の状態。
- `Result` (object): ログエントリ時のポーリングの結果。

### 5. `do_ping`

**説明:** ターゲットにpingを実行します。

**パラメータ:**

- `target` (string, required): pingを実行するIPアドレスまたはホスト名。
- `size` (number, optional, default: 64): パケットサイズ (64-1500)。
- `ttl` (number, optional, default: 254): IPパケットのTTL (1-254)。
- `timeout` (number, optional, default: 2): pingのタイムアウト（秒）(1-10)。

**出力:** 以下のプロパティを持つJSONオブジェクト:
- `Result` (string): pingの結果。
- `Time` (string): pingのタイムスタンプ。
- `RTT` (string): ラウンドトリップタイム。
- `RTTNano` (int64): ラウンドトリップタイム（ナノ秒）。
- `Size` (int): パケットサイズ。
- `TTL` (int): 応答パケットのTTL。
- `ResponceFrom` (string): 応答元のIPアドレス。
- `Location` (string): 応答元の場所。

### 6. `get_MIB_tree`

**説明:** TWSNMPからMIBツリーを取得します。

**パラメータ:** なし。

**出力:** MIBツリーを表すJSONオブジェクト。

### 7. `snmpwalk`

**説明:** SNMPウォークを実行します。

**パラメータ:**

- `target` (string, required): ウォークするIPアドレス、ホスト名、またはノード名。
- `mib_object_name` (string, required): ウォークするMIBオブジェクト名。
- `community` (string, optional): SNMPv2cコミュニティ文字列。
- `user` (string, optional): SNMPv3ユーザー名。
- `password` (string, optional): SNMPv3パスワード。
- `snmpmode` (string, optional): SNMPモード。有効な値は "v2c", "v3auth", "v3authpriv", "v3authprivex" です。

**出力:** 以下のプロパティを持つMIBオブジェクトのJSON配列:
- `Name` (string): MIBオブジェクト名。
- `Value` (string): MIBオブジェクトの値。

### 8. `add_node`

**説明:** TWSNMPに新しいノードを追加します。

**パラメータ:**

- `name` (string, required): ノード名。
- `ip` (string, required): ノードのIPアドレス。
- `icon` (string, optional): ノードのアイコン。
- `description` (string, optional): ノードの説明。
- `x` (number, optional): マップ上のノードのX座標 (64-1000)。
- `y` (number, optional): マップ上のノードのY座標 (64-1000)。

**出力:** 新しく追加されたノードを表すJSONオブジェクト。

### 9. `update_node`

**説明:** TWSNMPの既存のノードを更新します。

**パラメータ:**

- `id` (string, required): 更新するノードのID。
- `name` (string, optional): 新しいノード名。
- `ip` (string, optional): 新しいIPアドレス。
- `icon` (string, optional): 新しいアイコン。
- `description` (string, optional): 新しい説明。
- `x` (number, optional): 新しいX座標。
- `y` (number, optional): 新しいY座標。

**出力:** 更新されたノードを表すJSONオブジェクト。

### 10. `get_ip_address_list`

**説明:** TWSNMPからIPアドレスのリストを取得します。

**パラメータ:** なし。

**出力:** 以下のプロパティを持つIPアドレスオブジェクトのJSON配列:
- `IP` (string): IPアドレス。
- `MAC` (string): MACアドレス。
- `Node` (string): 関連付けられたノード名。
- `Vendor` (string): ネットワークインターフェースのベンダー。
- `FirstTime` (string): IPアドレスが最初に確認された時刻。
- `LastTime` (string): IPアドレスが最後に確認された時刻。

### 11. `get_resource_monitor_list`

**説明:** TWSNMPからリソースモニターデータのリストを取得します。

**パラメータ:** なし。

**出力:** 以下のプロパティを持つリソースモニターオブジェクトのJSON配列:
- `Time` (string): データのタイムスタンプ。
- `CPUUsage` (string): CPU使用率。
- `MemoryUsage` (string): メモリ使用率。
- `SwapUsage` (string): スワップ使用率。
- `DiskUsage` (string): ディスク使用率。
- `Load` (string): システム負荷。

### 12. `search_event_log`

**説明:** TWSNMPのイベントログを検索します。

**パラメータ:**

- `node_filter` (string, optional): ノード名でフィルタリングするための正規表現。
- `type_filter` (string, optional): タイプでフィルタリングするための正規表現。
- `level_filter` (string, optional): レベルでフィルタリングするための正規表現。有効なレベルは "warn", "low", "high", "debug", "info" です。
- `event_filter` (string, optional): イベントでフィルタリングするための正規表現。
- `limit_log_count` (number, optional, default: 100): 取得するログエントリの最大数 (1-10000)。
- `start_time` (string, optional, default: "-1h"): 検索の開始時刻。
- `end_time` (string, optional, default: "now"): 検索の終了時刻。

**出力:** 以下のプロパティを持つイベントログオブジェクトのJSON配列:
- `Time` (string): イベントのタイムスタンプ。
- `Type` (string): イベントタイプ。
- `Level` (string): イベントレベル。
- `Node` (string): 関連付けられたノード名。
- `Event` (string): イベントの説明。

### 13. `search_syslog`

**説明:** TWSNMPのsyslogを検索します。

**パラメータ:**

- `host_filter` (string, optional): ホスト名でフィルタリングするための正規表現。
- `tag_filter` (string, optional): タグでフィルタリングするための正規表現。
- `level_filter` (string, optional): レベルでフィルタリングするための正規表現。有効なレベルは "warn", "low", "high", "debug", "info" です。
- `message_filter` (string, optional): メッセージでフィルタリングするための正規表現。
- `limit_log_count` (number, optional, default: 100): 取得するログエントリの最大数 (1-10000)。
- `start_time` (string, optional, default: "-1h"): 検索の開始時刻。
- `end_time` (string, optional, default: "now"): 検索の終了時刻。

**出力:** 以下のプロパティを持つsyslogオブジェクトのJSON配列:
- `Time` (string): ログエントリのタイムスタンプ。
- `Level` (string): ログレベル。
- `Host` (string): ホスト名。
- `Type` (string): ログタイプ。
- `Tag` (string): ログタグ。
- `Message` (string): ログメッセージ。
- `Severity` (int): 重大度レベル。
- `Facility` (int): ファシリティコード。

### 14. `get_syslog_summary`

**説明:** TWSNMPからsyslogのサマリーを取得します。

**パラメータ:**

- `host_filter` (string, optional): ホスト名でフィルタリングするための正規表現。
- `tag_filter` (string, optional): タグでフィルタリングするための正規表現。
- `level_filter` (string, optional): レベルでフィルタリングするための正規表現。
- `message_filter` (string, optional): メッセージでフィルタリングするための正規表現。
- `top_n` (number, optional, default: 10): 取得する上位のsyslogパターンの数 (5-100)。
- `start_time` (string, optional, default: "-1h"): 検索の開始時刻。
- `end_time` (string, optional, default: "now"): 検索の終了時刻。

**出力:** 以下のプロパティを持つsyslogサマリーオブジェクトのJSON配列:
- `Pattern` (string): ログパターン。
- `Count` (int): パターンの出現回数。

### 15. `search_snmp_trap_log`

**説明:** TWSNMPのSNMPトラップログを検索します。

**パラメータ:**

- `from_filter` (string, optional): 送信元アドレスでフィルタリングするための正規表現。
- `trap_type_filter` (string, optional): トラップタイプでフィルタリングするための正規表現。
- `variable_filter` (string, optional): トラップ変数でフィルタリングするための正規表現。
- `limit_log_count` (number, optional, default: 100): 取得するログエントリの最大数 (1-10000)。
- `start_time` (string, optional, default: "-1h"): 検索の開始時刻。
- `end_time` (string, optional, default: "now"): 検索の終了時刻。

**出力:** 以下のプロパティを持つSNMPトラップログオブジェクトのJSON配列:
- `Time` (string): トラップのタイムスタンプ。
- `FromAddress` (string): 送信元のアドレス。
- `TrapType` (string): トラップタイプ。
- `Variables` (string): トラップ変数。

### 16. `get_server_certificate_list`

**説明:** TWSNMPからサーバー証明書のリストを取得します。

**パラメータ:** なし。

**出力:** 以下のプロパティを持つサーバー証明書オブジェクトのJSON配列:
- `State` (string): 証明書の状態。
- `Server` (string): サーバーアドレス。
- `Port` (uint16): サーバーポート。
- `Subject` (string): 証明書のサブジェクト。
- `Issuer` (string): 証明書の発行者。
- `SerialNumber` (string): 証明書のシリアル番号。
- `Verify` (boolean): 証明書が検証されているかどうか。
- `NotAfter` (string): 証明書の有効期限。
- `NotBefore` (string): 証明書の発行日。
- `Error` (string): 証明書に関連するエラー。
- `FirstTime` (string): 証明書が最初に確認された時刻。
- `LastTime` (string): 証明書が最後に確認された時刻。

### 17. `add_event_log`

**説明:** TWSNMPにイベントログを追加します。

**パラメータ:**

- `level` (string, optional, default: "info"): イベントレベル ("info", "normal", "warn", "low", "high")。
- `node` (string, optional): イベントに関連付けられたノードの名前。
- `event` (string, optional): イベントログの内容。

**出力:** 操作の結果を示す文字列 ("ok")。
