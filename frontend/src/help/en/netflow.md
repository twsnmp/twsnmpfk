# NetFlow

Inspect received NetFlow logs and analyze traffic.

## Table Columns

* **Time**
  Date and time the NetFlow packet was received.
* **Src IP**
  Source IP address.
* **Port**
  Source port number.
* **Loc**
  Source geographic location (requires GeoIP database).
* **Src MAC**
  Source MAC address.
* **Dst IP**
  Destination IP address.
* **Port**
  Destination port number.
* **Loc**
  Destination geographic location (requires GeoIP database).
* **Dst MAC**
  Destination MAC address.
* **Protocol**
  Protocol name (e.g., TCP, UDP, ICMP).
* **TCP Flags**
  TCP flags detected in the flow.
* **Packets**
  Total packets sent in this flow.
* **Bytes**
  Total bytes sent in this flow.
* **Duration**
  Duration of the flow in milliseconds.

## Button Descriptions

* **[Filter]** : Open the search filter dialog.
* **[Delete All Logs]** : Delete all NetFlow logs from the database.
* **[Report]** : Open the NetFlow traffic analysis reports.
* **[Copy]** : Copy the selected log text to the clipboard.
* **[AI Explain]** : Request an explanation of the selected logs from the AI (LLM).
* **[IP/MAC Info]** : Dropdown button to view detailed IP/MAC address lookup information.
* **[CSV]** : Export current logs to a CSV file.
* **[Excel]** : Export current logs to an Excel file.
* **[Reload]** : Refresh the logs list.

## Filter Settings

Search parameters in the filter dialog (supports regular expressions).

* **Start Time**
  Start date and time of the search.
* **End Time**
  End date and time of the search.
* **Simple Mode**
  Toggle to apply IP, Port, and Location filters to both source and destination bidirectionally.
* **IP (Simple Mode)**
  Filter IP addresses.
* **Port (Simple Mode)**
  Filter port numbers.
* **Loc (Simple Mode)**
  Filter locations.
* **Src IP**
  Filter source IP address.
* **Port**
  Filter source port.
* **Loc**
  Filter source location.
* **Dst IP**
  Filter destination IP address.
* **Port**
  Filter destination port.
* **Loc**
  Filter destination location.
* **Protocol**
  Filter protocol.
* **TCP Flags**
  Filter TCP flags.

## Report Types

Available analysis report types.

* **Heatmap**
  Hourly density heatmap of received flows.
* **Histogram**
  Statistical distribution charts.
* **Traffic**
  Time-series charts of traffic volume.
* **TOP List**
  Top rankings for IP addresses, ports, and protocols.
* **TOP List (3D)**
  Ranking report rendered as a 3D bar chart.
* **IP Pair Flow**
  Sankey diagram representing traffic combinations between IP addresses.
* **FFT Analysis**
  FFT-based periodicity analysis of traffic flow.
* **FFT Analysis (3D)**
  3D periodicity chart of FFT analysis.
* **Map**
  Map visualizing the locations of IP addresses.
