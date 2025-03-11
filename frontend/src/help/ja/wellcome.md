## TWSNMP FKへようこそ

ネットワーク管理マップを作成してみましょう。

まず、<button class="bg-green-600">＜はじめる＞</button>ボタンをクリックします。


---
#### データフォルダを選択
<span class="text-xl">
マップの設定やログを記録するデータベースファイルや拡張MIBなどを保存するためのフォルダです。
</span>


>>>
#### データストア内のファイル
<span class="text-xl">
データフォルダ内に以下ファイルを保存すればカスタマイズできます。

|ファイル|内容|
|---|---|
|twsnmpfk.db|データベースファイルです。存在しない場合は自動作成されます。|
|services.txt|サービス名の変換の使用するファイルです。(オプション)|
|mac-vendors-export.csv|MACアドレスとベンダー名の関係を示すデータベースです。(オプション)|
|polling.json|ポーリング設定(オプション)|
|mail_test.html|通知テストメールのテンプレート(オプション)|
|mail_notify.html|通知メールのテンプレート(オプション)|
|mail_repot.html|レポートメールのテンプレート(オプション)|
|extmibs/*|追加読み込みの拡張MIB(オプション)|

</span>

---
#### マップを表示
<span class="text-xl">
フォルダーを選択すると空白のマップが表示されます。
</span>

---
#### 起動パラメータ
<span class="text-xl">
起動時に以下のパラメータを指定できます。
</span>

```
Usage of twsnmpfk:
  -caCert string
    	CA Cert path
  -clientCert string
    	Client cert path
  -clientKey string
    	Client key path
  -datastore string
    	Path to data store directory
  -kiosk
    	Kisok mode(frameless and full screen)
  -lang string
    	Language(en|jp)
  -lock string
    	Disable edit map and lock page(map or loc)
  -maxDispLog int
    	Max log size to diplay (default 10000)
```
>>>
```
  -netflowPort int
    	Netflow port (default 2055)
  -ping string
    	ping mode icmp or udp
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

>>>
#### 起動パラメータの説明

<div class="text-xl">

|パラメータ|説明|
|---|---|
|datastore|データストアのパス|
|kiosk|キオスクモード（フレームレス、フルスクリーン）|
|lock <page>| マップの編集を禁止して表示するページを固定<br>(pageにはmap又はlocを指定)|
|maxDispLog <number>|ログの最大表示数(デフォルト 10000)| 
|ping <mode>|pingの動作モード(icmp又はudp)|
|syslogPort <port>|syslogの受信ポート(デフォルト514)|
|trapPort <port>| SNMP TRAP受信ポート(デフォルト162)|
>>>
|パラメータ|説明|
|---|---|
|sshdPort <port>| SSH Server受信ポート(デフォルト2022)|
|netflowPort <port>| NetFlow/IPFIX受信ポート(デフォルト2055)|
|sFlowPort <port>| sFlow受信ポート(デフォルト6343)|
|tcpdPort <port>| TCPログ受信ポート(デフォルト8086)|
|caCert <file>| TWLogEyeとTLS通信するためのCA証明書|
|clientCert <file>| TWLogEyeとmTLS通信するためのクライアント証明書|
|clientKey <file>| TWLogEyeとmTLS通信するためのクライアント鍵|

</div>

