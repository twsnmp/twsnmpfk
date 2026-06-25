# 地図設定

地図画面の表示方法やソースを設定する画面。

## 設定項目

* **スタイル**
  地図を表示するためのスタイル（MapLibre GL JSのスタイル定義URLまたはJSONオブジェクト）。
* **中央座標**
  地図を開いた際の基準となる中心点の座標（「経度,緯度」の順でカンマ区切りで指定。例: `135.338,39.614`）。
* **ズーム**
  地図の初期表示ズームレベル（拡大率）。
* **アイコンサイズ**
  地図上に表示するノードアイコンのピクセルサイズ（スライダーで16〜64pxの範囲を指定可能）。

## ボタンの説明

* **[保存]** ： 設定内容を保存。
* **[キャンセル]** ： 変更を保存せずに画面を閉じる。

## 地図スタイルについて

* **概要**
  MapLibre GL JSを使用しており、表示する地図はスタイルデータで決定。MapLibre GL形式のスタイルURL、またはローカルタイルソース等を定義したJSONが指定可能。
* **スタイルURLの例**
  `https://tile.openstreetmap.jp/styles/osm-bright-ja/style.json`
* **スタイルJSONの例（ラスタタイルソース定義）**
  ```json
  {
    "version": 8,
    "sources": {
      "MIERUNEMAP": {
        "type": "raster",
        "tiles": ["https://tile.mierune.co.jp/mierune_mono/{z}/{x}/{y}.png"],
        "tileSize": 256,
        "attribution": "Maptiles by MIERUNE, under CC BY. Data by OpenStreetMap contributors, under ODbL."
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
