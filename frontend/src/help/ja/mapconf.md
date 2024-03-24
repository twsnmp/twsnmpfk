#### マップ設定
<div class="text-xl">
管理マップの設定をする画面です。
</div>

![マップ設定](../../help/ja/2023-11-29_04-48-42.png)

>>>

<div class="text-xl">

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
|SSH Server|SSHサーバーを起動します。|
|ARP Watch|ARP監視機能を有効にします。|

</div>

---
#### Syslog,SNMP TRAPの受信ポートを変えたい時

<div class="text-xl">

ポート番号は、プログラムの起動パラメータで指定します。

</div>

```
  -syslogPort int
    	Syslog port (default 514)
  -trapPort int
      SNMP TRAP port (default 162)
  -sshdPort int
      SSH server port (default 2022)
```

<p style="color:red;font-size: 16px;">
  ※ syslogやSNMP TRAPが受信できない時は、OSやセキュリティーソフトのファイヤーウオールの設定を確認ください。
</p>
