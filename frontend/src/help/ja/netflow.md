#### NetFlow

<div class="text-xl">

|項目|内容|
|----|----|
|日時|NetFlowを受信した日時です。|
|送信元|送信元のIPです。|
|ポート|送信元のポート番号です。|
|位置|送信元の位置です。GeoIP DBが必要です。|
|宛先|宛先のIPです。|
|ポート|宛先のポート番号です。|
|位置|宛先の位置です。GeoIP DBが必要です。|
|プロトコル|tcp/udp/icmpなどのプロトコルです。|
|TCPフラグ|TCPのフラグです。|
|パケット|送信パケット数です。|
|バイト|送信バイト数です。|
|期間|フローの通信時間です。|

</div>

>>>
#### ボタンの説明

<div class="text-xl">

|項目|内容|
|----|----|
|フィルター|検索条件を指定してNetFlowを表示します。|
|<span style="color: red;">全ログ削除</span>|全てのNetFlowを削除します。|
|Copy|選択したログをコピーします。|
|レポート|NetFlowの分析レポートを表示します。|
|CSV|NetFlowをCSVファイルにエクスポートします。|
|Excel|NetFlowをExcelファイルにエクスポートします。|
|更新|NetFlowのリストを最新の状態に更新します。|

</div>


---

#### フィルターの項目

<div class="text-xl">

|項目|内容|
|----|----|
|開始日時|検索開始の日時を指定します。|
|終了日時|検索終了の日時を指定します。|
|簡易モード|IP、ポート、位置を双方向に適用するモードです。|
|IP|簡易モードの場合の、送信元、宛先のIPを指定します。|
|ポート|簡易モードの場合の、送信元、宛先のポートを指定します。|
|位置|簡易モードの場合の、送信元、宛先の位置を指定します。|
|送信元IP|送信元のIPを指定します。|
|ポート|送信元のポートを指定します。|
|位置|送信元の位置を指定します。|
|宛先IP|宛先のIPを指定します。|
|ポート|宛先のポートを指定します。|
|位置|宛先の位置を指定します。|
|プロトコル|プロトコル名を指定します。|
|TCPフラグ|TCPフラグを指定します。|

<span style="color:red">文字列は、正規表現で検索できます。</span>

</div>


---
#### レポート

<div class="text-xl mb-2">

|レポート名|内容|
|----|----|
|ヒートマップ|NetFlowの受信数の時間帯別のヒートマップです。|
|ヒストグラム|数値データのヒストグラムです。|
|トラフィック|通信量の時系列グラフです。|
|TOPリスト|項目別のランキングレポートです。|
|TOPリスト(3D)|項目別のランキングレポートを3Dのグラフで表示したものです。|
|IPペアーフロー|通信の組み合わせをグラフで表示したものです。|
|FFT分析|FFTで通信の周期を分析するものです。|
|FFT分析(3D)|FFTで通信の周期を分析して３Dのグラフに表示します。|
|地図|IPアドレスの位置を地図に表示します。| 


</div>
