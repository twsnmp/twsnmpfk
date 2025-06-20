#### マップ設定
<div class="text-xl">
管理マップの設定をする画面です。
</div>

<div class="text-lg">

|項目|内容|
|----|----|
|マップ名|マップの名前です。画面の左上に表示されます。<br>好きな名前をつけてください。|
|アイコンサイズ|マップに表示するアイコンのサイズです。|
|ポーリング間隔|デフォルトのポーリング間隔です。|
|タイムアウト|デフォルトのタイムアウトです。|
|リトライ|デフォルトのリトライ回数です。|
|ログ保存日数|ログを保存する日数です。過ぎた場合にログは自動で削除します。|
|SNMPモード|SNMPのバージョンと暗号化の種類です。(SNMPv1,SNMPv2c,SNMPv3)|
|SNMP Community|SNMPv1,v2cの時のCommunity名です。|
|SNMP ユーザー|SNMPv3の時のユーザー名です。|
|SNMP パスワード|SNMPv3の時のパスワード名です。|
|Syslog|Syslogを受信します。|
|SNMP TRAP|SNMP TRAPを受信します。|
|ARP Watch|ARP監視機能を有効にします。|
|SSH Server|SSHサーバーを起動します。|

</div>

>>>

#### マップ設定（続き）
<div class="text-xl">
管理マップの設定をする画面です。
</div>

<div class="text-lg">

|項目|内容|
|----|----|
|TCP Server|TCPサーバーで受信します。|
|OpenTelemetry|OpenTelemetryサーバーを起動します。|
|OpenTelemetry保存時間|OpenTelemetryデータの保存時間を指定します。|
|OpenTelemetry送信元|OpenTelemetryサーバーへのデータ送信元を限定します。|
|MCPサーバートランスポート|MCPサーバーのトランスポート(OFF/SSE/Steamable)を指定します。|
|MCPサーバーエンドポイント|MCPサーバーの受信IPとポートを指定します。|

</div>

---
#### Syslog,SNMP TRAPなどの受信ポートを変えたい時

<div class="text-lg">
ポート番号は、プログラムの起動パラメータで指定します。

```
  -netflowPort int
    	Netflow port (default 2055)
  -otelGRPCPort int
    	OpenTelemetry server gRPC port (default 4317)
  -otelHTTPPort int
    	OpenTelemetry server HTTP port (default 4318)
  -sFlowPort int
    	sFlow port (default 6343)
  -sshdPort int
    	SSH server port (default 2022)
  -syslogPort int
    	Syslog port (default 514)
  -tcpdPort int
    	tcp server port (default 8086)
  -trapPort int
    	SNMP TRAP port (default 162)
```
</div>


<p style="color:red;font-size: 16px;">
  ※ syslogやSNMP TRAPが受信できない時は、OSやセキュリティーソフトのファイヤーウオールの設定を確認ください。
</p>
