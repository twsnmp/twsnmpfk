# ARP Watch Log

ARP (Address Resolution Protocol) monitoring log for changes and new discoveries. Displays a time-series graph of log counts at the top.

## Table Columns

* **State**
  Log status ("New" or "Change").
* **Date and time**
  Date and time of the log.
* **IP address**
  Target IP address (addresses starting with 169.254. are shown in red).
* **Node**
  Name of the node registered on the map (blank if not registered).
* **New MAC address**
  Newly discovered MAC address or MAC address after change.
* **New vendor**
  Vendor name corresponding to the new MAC address.
* **Old MAC address**
  MAC address before change.
* **Old vendor**
  Vendor name corresponding to the old MAC address.

## Button Descriptions

* **[Report]** : Displays the ARP watch log analysis report.
* **[CSV]** : Exports the ARP watch log to a CSV file.
* **[Excel]** : Exports the ARP watch log to an Excel file with the chart.
* **[Reload]** : Reloads the ARP watch log list.

## Report Descriptions

* **By IP address**
  Aggregated log counts by IP address (useful for identifying IP addresses with frequent changes).
* **By IP address (3D)**
  3D aggregated report of logs by IP address and time series (useful for identifying when new discoveries or changes occurred).
