#### Syslog

<div class="text-xl mb-2">
Syslogの画面です。<br>
上部にログの発生件数を時系列で示したグラフがあります。
</div>

![Syslog](../../help/ja/2023-12-02_09-15-05.png)

>>>
#### Syslogの項目

<div class="text-xl">

|項目|内容|
|----|----|
|レベル|Syslogのレベルです。<br>重度、軽度、注意、情報があります。|
|日時|Syslogを受信した日時です。|
|ホスト|Syslogの送信元ホストです。|
|タイプ|syslogのファシリティーと優先度の文字列です。|
|タグ|Syslogのタグです。プロセスとプロセスIDなどです。|
|メッセージ|Syslogのメッセージです。|

</div>

>>>
#### ボタンの説明

<div class="text-xl">

|項目|内容|
|----|----|
|ポーリング|選択したSyslogからポーリングを登録します。|
|フィルター|検索条件を指定してSyslogを表示します。|
|コピー|選択したログをコピーします。|
|<span style="color: red;">全ログ削除</span>|全てのSyslogを削除します。|
|レポート|Syslogの分析レポートを表示します。|
|マジック分析|syslogから情報を自動抽出して分析できます。|
|CSV|SyslogをCSVファイルにエクスポートします。|
|Excel|SyslogをExcelファイルにエクスポートします。|
|更新|Syslogのリストを最新の状態に更新します。|

</div>


---
#### フィルター

<div class="text-xl mb-2">
Syslogの検索条件を指定するダイアログです。
</div>

![Syslogフィルター](../../help/ja/2023-12-02_13-39-17.png)

>>>
#### フィルターの項目

<div class="text-xl">

|項目|内容|
|----|----|
|レベル|Syslogのレベルです。<br>全て、情報以上、注意以上、軽度以上、重度があります。|
|ホスト|送信元のホストです。|
|タグ|Syslogのタグの値です。|
|メッセージ|Syslogのメッセージです。|

<span style="color:red">文字列は、正規表現で検索できます。</span>

</div>


---
#### 状態別

<div class="text-xl mb-2">
 Syslogの件数を状態別に集計したレポートです。
</div>

![Syslogの状態別](../../help/ja/2023-12-02_13-43-23.png)

---
#### ヒートマップ

<div class="text-xl mb-2">
Syslogの時間毎の件数をヒートマップで集計したレポートです。
</div>

![ヒートマップ](../../help/ja/2023-12-02_13-43-39.png)

---
#### ホスト別

<div class="text-xl mb-2">
Syslogの件数を送信元ホスト別に集計したレポートです。
</div>

![ホスト別](../../help/ja/2023-12-02_13-43-51.png)

---
#### ホスト別(3D)

<div class="text-xl mb-2">
Syslogを送信元ホスト、プライオリティー、時刻の３次元グラフで表示したレポートです。
</div>

![ホスト別(3D)](../../help/ja/2023-12-02_13-44-07.png)

---
#### FFTによる周期分析

<div class="text-xl mb-2">
Syslogをホスト毎にFFT分析して受信件数の周期を分析したレポートです。
</div>

![Syslog FFT分析](../../help/ja/2023-12-02_13-44-43.png)
