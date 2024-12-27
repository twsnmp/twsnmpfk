#### 地図設定
<div class="text-xl">
地図の設定をする画面です。
</div>

<div class="text-lg">

|項目|内容|
|----|----|
|スタイル|地図のスタイルを指定します。URLかJSONで指定します。|
|中央座標|地図の中央座標を経度,緯度の順にしています。<br>例:135.33885756281734,39.614306840830096|
|ズーム|地図の拡大レベルを指定します。|
|アイコンサイズ|表示するアイコンのサイズを指定します。|

</div>

---
#### 地図のスタイルについて

<div class="text-xl">
  MapLibre GL JSを使って地図を表示しています。表示する地図は、スタイルで指定します。
  URLやJSONで指定できます。MapLibre GL JSで検索して適したものを見つけてください。
</div>

#### URLの例

```
https://tile.openstreetmap.jp/styles/osm-bright-ja/style.json
```


>>>

#### JSONの例

```json
{
			 	"version": 8,
			 	"sources": {
			 		"MIERUNEMAP": {
						"type": "raster",
			 			"tiles": ["https://tile.mierune.co.jp/mierune_mono/{z}/{x}/{y}.png"],
						"tileSize": 256,
			 			"attribution":
			 				"Maptiles by <a href='https://mierune.co.jp/' target='_blank'>MIERUNE</a>, under CC BY. Data by <a href='https://osm.org/copyright' target='_blank'>OpenStreetMap</a> contributors, under ODbL."
			 		}
			 	},
			 	"layers": [
					{
						"id": "MIERUNEMAP",
		 				"type": "raster",
			 			"source": "MIERUNEMAP",
			 			"minzoom": 0,
			 			"maxzoom": 18
			 		}
			 	]
}
```
