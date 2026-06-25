# sFlow

Analyze flow samples (traffic data) and counter samples (performance metrics) received from sFlow agents.

## Flow Sample Columns

* **Time**
  Date and time the sFlow sample was received.
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
  TCP flags detected in the sample.
* **Bytes**
  Sent bytes.
* **Reason**
  The reason code for forwarding/dropping the packet (as defined by sFlow standard).

## Counter Sample Columns

* **Time**
  Date and time the counter sample was received.
* **Src IP**
  Source IP address of the agent.
* **Type**
  Type of counter sample (GenericInterfaceCounter, HostCPUCounter, HostMemoryCounter, HostDiskCounter, HostNetCounter).
* **Data**
  The detailed key-value metrics returned in the counter sample.

## Button Descriptions

* **[Counter / Flow]** (Counter toggle) : Toggle between flow samples and counter samples views.
* **[Filter]** : Open the search filter dialog.
* **[Delete All Logs]** : Delete all sFlow logs.
* **[Copy]** : Copy the selected log text to the clipboard.
* **[Report]** : Open the sFlow traffic or counter metrics analysis reports.
* **[IP/MAC Info]** : Dropdown button to view detailed IP/MAC address lookup information.
* **[CSV]** : Export the logs list to a CSV file.
* **[Excel]** : Export the logs list to an Excel file.
* **[Reload]** : Refresh the logs list.

## Filter Settings

### Flow Filter
* **Start Time / End Time** : Date and time range.
* **Simple Mode** : Apply IP, Port, and Location filters bidirectionally.
* **Src IP / Dst IP** : Filter source/destination IP address (regex).
* **Port** : Filter port numbers.
* **Loc** : Filter geographic locations.
* **Protocol** : Filter protocols.
* **TCP Flags** : Filter TCP flags.

### Counter Filter
* **Start Time / End Time** : Date and time range.
* **Src IP** : Filter source agent IP.
* **Type** : Filter counter sample types.

## Report Types

### Flow Sample Reports
* **Heatmap** : Hourly density heatmap of received flows.
* **Traffic** : Time-series chart of traffic volume.
* **TOP List** : Top rankings of IPs, ports, and protocols.
* **TOP List (3D)** : Top rankings represented as a 3D bar chart.
* **IP Pair Flow** : Sankey diagram representing communication flow combinations.
* **FFT Analysis** : Periodicity analysis of traffic using FFT.
* **FFT Analysis (3D)** : 3D visualization of FFT periodicity analysis.
* **Map** : Map visualizing the location of IP addresses.

### Counter Sample Reports
* **Heatmap** : Hourly density heatmap of received samples.
* **I/F BPS** : Time-series chart of interface speed (Bytes/Sec).
* **I/F PPS** : Time-series chart of interface packet rates (Packets/Sec).
* **CPU** : Time-series chart of host CPU usage and load average.
* **Memory** : Chart of host memory utilization and available capacity.
* **Disk** : Chart of disk space utilization and disk I/O.
* **Network** : Chart of host network statistics.
