# Anomaly List

Machine learning anomaly detection list. Displays items for which analysis is enabled in the polling log settings.

## Table Columns

* **Anomaly score**
  Deviation value indicating the degree of anomaly (average is 50, higher values indicate higher anomaly).
* **Node Name**
  Name of the target node.
* **Polling**
  Name of the target polling.
* **Count**
  Number of data points analyzed (lower counts may reduce accuracy).
* **Last time**
  Date and time of the last analysis.

## Button Descriptions

* **[Report]** : Displays the report for the selected anomaly detection results.
* **[CSV(Data)]** : Saves the selected analysis data to a CSV file.
* **[Clear]** : Clears the selected anomaly detection results.
* **[Reload]** : Reloads the anomaly list.

## Report Descriptions

* **AI anomaly score heat map**
  Heat map showing the anomaly score on a daily basis (red indicates higher anomaly).
* **AI anomaly score percentage**
  Pie chart showing the distribution of anomaly scores over the entire period.
* **AI anomaly score time chart**
  Time-series chart showing anomaly scores.
