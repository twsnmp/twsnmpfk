#### SNMP TRAP
<div class="text-xl mb-4">
SNMP TRAPログの画面です。
上部にログの発生件数を時系列で示したグラフがあります。
</div>

<div class="text-lg">
|項目|内容|
|----|----|
|日時|SNMP TRAPを受信した日時です。|
|送信元|SNMP TRAPの送信元ホストです。|
|タイプ|SNMP TRAPのタイプです。|
|変数|SNMP TRAPに付帯した変数です。|

</div>

>>>
#### SNMP TRAP(ボタン)

<div class="text-lg">

|項目|内容|
|----|----|
|ポーリング|選択したSNMP TRAPからポーリングを登録します。|
|フィルター|検索条件を指定してSNMP TRAPを表示します。|
|<span style="color: red;">全ログ削除</span>|全てのSyslogを削除します。|
|コピー|選択したログをコピーします。|
|レポート|SNMP TRAPの分析レポートを表示します。|
|CSV|SNMP TRAPをCSVファイルにエクスポートします。|
|Excel|SNMP TRAPをExcelファイルにエクスポートします。|
|更新|SNMP TRAPのリストを最新の状態に更新します。|

</div>


---
#### SNMP TRAP フィルター

<div class="text-xl mb-2">
SNMP TRAPの検索条件を指定するダイアログです。
</div>

<div class="text-lg">

|項目|内容|
|----|----|
|送信元|送信元のホストです。|
|タイプ|SNMP TRAPのタイプです。|

<span style="color:red">文字列は、正規表現で検索できます。</span>

</div>


---
#### TRAP種類別
<div class="text-xl mb-4">
 SNMP TRAPの件数を種類別に集計したレポートです。
</div>

#### ヒートマップ
<div class="text-xl mb-4">
SNMP TRAPの時間毎の件数をヒートマップで集計したレポートです。
</div>

#### ホスト別
<div class="text-xl mb-4">
SNMP TRAPの受信件数を送信元ホスト別に集計したレポートです。
</div>

#### 送信元と種別(3D)
<div class="text-xl mb-4">
SNMP TRAPの受信ログを送信元ホスト、種別、時刻の３次元グラフで表示したレポートです。
</div>

