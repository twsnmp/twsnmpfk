# Drawing Items

Edit screen for items used to draw backgrounds or additional information on the map.

## Common Parameters

* **Type**
  Type of drawing item (Rectangle, Ellipse, Label, Image, Polling Result, etc. Changeable only during creation).
* **Display condition**
  Visibility condition based on the map state ("Always", "When status is Low anomaly or higher", "When status is High anomaly").
* **Magnification**
  Display scale ratio.

## Parameters by Type

* **Rectangle, Ellipse, Group (Frame), Group (Fill)**
  * **Width**
    Width of the item.
  * **Height**
    Height of the item.
  * **Color**
    Fill color and opacity.
* **Label**
  * **Font Size**
    Font size of the label text.
  * **Color**
    Text color.
  * **Text**
    String to be displayed.
* **Image**
  * **Width**
    Width of the displayed image.
  * **Height**
    Height of the displayed image.
  * **Image**
    Image file path (specified using the **[Select]** button).
* **Polling Results (Text, Gauge, New Gauge, Bar, Line)**
  * **Size**
    Font size or display size of the gauge, bar, or line.
  * **Node**
    Node to select the polling target.
  * **Polling**
    Target polling to display results from.
  * **Variable name**
    Name of the variable to extract and display from the polling results (blank for automatic setting).
  * **Display format**
    (Text only) Format specifier for displaying data (blank for automatic setting).
  * **Gauge label**
    (Gauge and New Gauge only) Label string displayed below the gauge.

## Button Descriptions

* **[Save]** : Saves the drawing item settings.
* **[Select]** : (For images) Selects an image file from the local file system.
* **[Help]** : Displays this help.
* **[Cancel]** : Closes the window without saving.
