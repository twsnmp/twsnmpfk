# AI Analysis Settings

Settings for AI analysis anomaly detection thresholds.

## Settings Parameters

* **Level to be high**
  Deviation score threshold to determine a severe anomaly.
* **Level to be low**
  Deviation score threshold to determine a mild anomaly.
* **Level to be warn**
  Deviation score threshold to determine a warning anomaly.

## Button Descriptions

* **[Save]** : Saves the settings.
* **[Cancel]** : Closes the settings window without saving.

## About AI Analysis

* **Execution**
  Triggered when the log mode is set to "AI Analysis" in the polling settings.
* **Methodology**
  Detects anomalies in numeric polling results using the Isolation Forest algorithm and calculates a deviation score.
* **Interpretation of Score**
  Indicates how rare the occurrence is. Thresholds are expressed in statistical probabilities (e.g., once in 10,000 times).
