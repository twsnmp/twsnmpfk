#### PING

<div class="text-xl mb-2 text-left">
PINGを実行する画面です。<br>
<span style="color:red;">位置情報を取得するには、GeoIPのデータベースファイルが必要です。</span>

</div>

![PING](../../help/ja/2023-12-01_06-44-23.png)

>>>
#### 説明

<div class="text-lg">

|項目|内容|
|----|----|
|IPアドレス|PINGを実行する対象のIPアドレスです。|
|回数|PINGの実行回数です。|
|サイズ|PINGパケットのサイズです。<br>変化モードは、サイズを増やしながら実行します。|
|TTL|PINGパケットのTTL値です。<br>トレースルートは、TTL値を増やしながら実行します。|
|結果グラフ|PINGの実行結果の応答時間、TTL値のグラフです。|
|結果|PINGの実行結果です。<br>結果、実施日時、応答時間、サイズ、送信受信のTTL、応答元IP、位置|
|BEEP|PINGの実行結果を音で知らせます。|
|開始|PINGを開始します。|
|停止|PINGを停止します。|
|閉じる|PINGを終了します。|

</div>


---
#### ヒストグラム

<div class="text-xl mb-2 text-left">
応答時間のヒストグラムです。
</div>

![ヒストグラム](../../help/ja/2023-12-01_06-58-41.png)

---
#### ３D分析

<div class="text-xl mb-2 text-left">
応答時間、サイズ、実施日時を３Dのグラフで表示したものです。
</div>

![PING 3D](../../help/ja/2023-12-01_06-36-33.png)

---
#### 回線予測

<div class="text-xl mb-2 text-left">

サイズを変化させならが実施した場合に応答時間の変化から
回線速度を予測するレポートです。
</div>

![回線速度予測](../../help/ja/2023-12-01_06-36-19.png)

---
#### 経路分析

<div class="text-xl mb-2 text-left">
位置情報を表示します。GeoIPのデータベースがないと表示できません。
</div>

![経路分析](../../help/ja/2023-12-01_07-17-17.png)
