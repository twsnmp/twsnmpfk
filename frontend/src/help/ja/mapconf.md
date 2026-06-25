# マップ設定

管理マップおよび各種サーバー機能、収集機能の設定を行う画面。

## パラメータの説明

* **マップ名**
  マップの名前。画面左上に表示される。
* **マップサイズ**
  マップのキャンバスサイズ（自動、A4縦、A4横など）。
* **アイコンサイズ**
  マップに表示するアイコンのサイズ。
* **ポーリング間隔**
  デフォルトのポーリング実行間隔（秒）。
* **タイムアウト**
  デフォルトの通信タイムアウト（秒）。
* **リトライ**
  デフォルトの通信リトライ回数。
* **ログ保存日数**
  イベントログや統計データの保存日数。期限を過ぎたデータは自動的に削除される。
* **OpenTelemetry保存時間**
  OpenTelemetryデータの保存時間（時間）。
* **OpenTelemetry送信元**
  OpenTelemetryサーバーへのデータ送信元IPアドレスの制限設定。
* **MCPサーバートランスポート**
  MCPサーバーのトランスポート設定（OFF / SSE / Streamable）。
* **MCPサーバーエンドポイント**
  MCPサーバーの待受IPアドレスおよびポート。
* **MCPサーバー送信元**
  MCPサーバーへの接続を許可する送信元IPアドレスの制限設定。
* **MCPサーバーアクセスキー**
  MCPサーバーの認証用アクセスキー。
* **AIプロバイダー**
  使用するLLM（大規模言語モデル）のプロバイダー。
* **AI APIベースURL**
  LLM APIのベースURL。
* **AI APIキー**
  LLM APIのアクセスキー。
* **AIモデル**
  使用するLLMのモデル名。
* **SNMPモード**
  デフォルトのSNMPバージョンとセキュリティモード（v1 / v2c / v3）。
* **SNMPコミュニティ**
  SNMPv1/v2c時のコミュニティ名。
* **SNMPユーザー**
  SNMPv3時のユーザー名。
* **SNMPパスワード**
  SNMPv3時の認証・暗号化パスワード。
* **Syslog**
  Syslogの受信を行う機能の有効化。
* **NetFlow**
  NetFlowの受信を行う機能の有効化。
* **sFlow**
  sFlowの受信を行う機能の有効化。
* **SNMP TRAP**
  SNMP TRAPの受信を行う機能の有効化。
* **ARP Watch**
  ARP監視機能の有効化。
* **ARP監視範囲**
  ARP監視を行うIPアドレスの範囲。
* **ARP消去時間**
  ARPキャッシュの保存期限。
* **SSHサーバー**
  SSHサーバー機能の起動。
* **TCPサーバー**
  TCP経由でのログ受信を行うTCPサーバー機能の起動。
* **OpenTelemetry**
  OpenTelemetryコレクター機能の起動。
* **MQTT**
  MQTTブローカー/クライアント機能の起動。
* **MQTT → Syslog**
  MQTTメッセージを受信しSyslogとして記録する機能の有効化。

## 受信ポートの変更方法

SyslogやSNMP TRAP、NetFlowなどの受信ポートは、プログラムの起動パラメータで指定する。

* **-netflowPort** : Netflowポート（デフォルト 2055）
* **-otelGRPCPort** : OpenTelemetry gRPCポート（デフォルト 4317）
* **-otelHTTPPort** : OpenTelemetry HTTPポート（デフォルト 4318）
* **-sFlowPort** : sFlowポート（デフォルト 6343）
* **-sshdPort** : SSHサーバーポート（デフォルト 2022）
* **-syslogPort** : Syslogポート（デフォルト 514）
* **-tcpdPort** : TCPサーバーポート（デフォルト 8086）
* **-trapPort** : SNMP TRAPポート（デフォルト 162）

※ SyslogやSNMP TRAPなどのパケットを受信できない場合は、OSやセキュリティソフトのファイアウォール設定を確認すること。
