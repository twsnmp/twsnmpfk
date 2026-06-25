# Event Log

Screen displaying system and monitoring event logs. A time-series graph showing log counts is presented at the top.

## Table Columns

* **Level**
  Severity level of the log ("Severe", "Mild", "Warn", "Recovery", "Info").
* **Date and time**
  Date and time the event was recorded.
* **Type**
  Type of event source ("polling", "system", "oprate", "user", "arpwatch").
* **Node**
  Name of the node related to the log (blank if not related to a specific node).
* **Event**
  Details of the recorded event.

## Button Descriptions

* **[Filter]** : Opens the filter settings dialog to search logs.
* **[Delete All Logs]** : Deletes all event logs from the database.
* **[Report]** : Displays statistical analysis reports.
* **[CSV]** : Exports the filtered event log to a CSV file.
* **[Excel]** : Exports the filtered event log to an Excel file with the chart.
* **[Reload]** : Reloads the event log list.

## Filter Settings

* **Start**
  Start date and time of the search range.
* **End**
  End date and time of the search range.
* **Level**
  Log levels to include ("All", "Warn", "Low", "High").
* **Type**
  Filter by event type (regular expressions supported).
* **Node**
  Filter by node name (regular expressions supported).
* **Event**
  Filter by event description (regular expressions supported).
* **[Search]** : Applies the filters and performs the search.
* **[Cancel]** : Closes the filter window without applying changes.

## Report Descriptions

* **By State**
  Distribution chart of event logs by severity level.
* **Heatmap**
  Heat map aggregating event occurrences by hour and day.
* **By Node**
  Ranking chart of event logs by node.
* **Availability**
  Time-series chart of device availability metrics from "oprate" events.
* **ARP Watch**
  Time-series chart of address usage rates from "arpwatch" events.
