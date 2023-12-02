#### SNMP TRAP

<div class="text-xl mb-2">
SNMP TRAPログの画面です。<br>
上部にログの発生件数を時系列で示したグラフがあります。
</div>

![SNMP TRAP](../../help/ja/2023-12-02_14-35-34.png)

>>>
#### SNMP TRAPログの項目

<div class="text-xl">

|項目|内容|
|----|----|
|日時|SNMP TRAPを受信した日時です。|
|送信元|SNMP TRAPの送信元ホストです。|
|タイプ|SNMP TRAPのタイプです。|
|変数|SNMP TRAPに付帯した変数です。|

</div>

>>>
#### ボタンの説明

<div class="text-xl">

|項目|内容|
|----|----|
|ポーリング|選択したSNMP TRAPからポーリングを登録します。|
|フィルター|検索条件を指定してSNMP TRAPを表示します。|
|<span style="color: red;">全ログ削除</span>|全てのSyslogを削除します。|
|レポート|SNMP TRAPの分析レポートを表示します。|
|CSV|SNMP TRAPをCSVファイルにエクスポートします。|
|Excel|SNMP TRAPをExcelファイルにエクスポートします。|
|更新|SNMP TRAPのリストを最新の状態に更新します。|

</div>


---
#### フィルター

<div class="text-xl mb-2">
SNMP TRAPの検索条件を指定するダイアログです。
</div>

![SNMP TRAPフィルター](../../help/ja/2023-12-02_14-38-44.png)

>>>
#### フィルターの項目

<div class="text-xl">

|項目|内容|
|----|----|
|送信元|送信元のホストです。|
|タイプ|SNMP TRAPのタイプです。|

<span style="color:red">文字列は、正規表現で検索できます。</span>

</div>


---
#### TRAP種類別

<div class="text-xl mb-2">
 SNMP TRAPの件数を種類別に集計したレポートです。
</div>

![TRAP種類別](../../help/ja/2023-12-02_14-41-15.png)

---
#### ヒートマップ

<div class="text-xl mb-2">
SNMP TRAPの時間毎の件数をヒートマップで集計したレポートです。
</div>

![ヒートマップ](../../help/ja/2023-12-02_14-42-06.png)

---
#### ホスト別

<div class="text-xl mb-2">
SNMP TRAPの受信件数を送信元ホスト別に集計したレポートです。
</div>

![ホスト別](../../help/ja/2023-12-02_14-42-19.png)

---
#### 送信元と種別(3D)

<div class="text-xl mb-2">
SNMP TRAPの受信ログを送信元ホスト、種別、時刻の３次元グラフで表示したレポートです。
</div>

![ホストと種別(3D)](../../help/ja/2023-12-02_14-42-30.png)
