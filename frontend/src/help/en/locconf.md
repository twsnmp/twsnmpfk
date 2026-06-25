# Location Map Settings

Settings tab to configure the location map display style and default view.

## Settings Parameters

* **Style**
  Style configuration for the map (MapLibre GL style URL or JSON object).
* **Central coordinates**
  Default center coordinates of the map specified as "longitude,latitude" (e.g., `135.338,39.614`).
* **Zoom**
  Initial zoom level of the map.
* **Icon size**
  Display size of the node icons in pixels (configurable between 16 and 64 pixels via slider).

## Button Descriptions

* **[Save]** : Saves the settings.
* **[Cancel]** : Closes the settings window without saving.

## About Map Styles

* **Overview**
  The map is displayed using MapLibre GL JS. The visual design and source tiles of the map are defined by style data, which can be configured via a remote style JSON URL or a custom style JSON object.
* **Style URL Example**
  `https://tile.openstreetmap.jp/styles/osm-bright-ja/style.json`
* **Style JSON Example (Raster Tile Source Definition)**
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
