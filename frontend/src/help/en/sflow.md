#### sFlow

##### Flow Sample
<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Date and time | It is the date and time when the SFLOW sample received.|
| Sending source | It is the source IP.|
| Port | It is the port number of the source.|
| Location | It is the source of the source.GEOIP DB is required.|
| Mac | This is the MAC address of the source.|
| Address | IP for the destination.|
| Port | The destination port number.|
| Location | It is the destination position.GEOIP DB is required.|
| Mac | This is the destination MAC address.|
| Protocol | Protocol such as TCP/UDP/ICMP.|
| TCP flag | TCP flag.|
| Byte | The number of sending bytes.|
</div>

>>>
#### sFlow(button)

<div class="text-lg">

| Items | Contents |
| ---- | ---- |
| Counter | Switch between flow samples and counter samples.|
| Filter | Specify the search conditions and display SFLOW.|
| <Span style = "color: red;"> Delete all logs </span> | Delete all SFLOW.|
| Copy | Copy the selected log.|
| Report | Displays SFLOW analysis reports.|
| CSV | Export SFLOW to CSV file.|
| Excel | Export SFLOW to Excel file.|
| Update | Update the list of SFLOW to the latest state.|

</div>

>>>

#### sFlow sample filter

<div class="text-lg">

| Items | Contents |
| ---- | ---- |
| Start date and time | Specify the date and time of the search start.|
| End date and time | Specify the date and time of the search termination.|
| Simple mode | Mode to apply IP, port, and position in both directions.|
| IP | In the case of simple mode, specify the source and destination IP.|
| Port | In the case of simple mode, specify the source and destination port.|
| Location | Specify the source and destination position in the case of simple mode.|
| Sending source IP | Specify the source IP.|
| Port | Specify the source port.|
| Location | Specify the source position.|
| Destination IP | Specify the destination IP.|
| Port | Specify the destination port.|
| Location | Specify the destination position.|
| Protocol | Specify the protocol name.|
| TCP flag | Specify the TCP flag.|

<Span style = "color: red"> Character strings can be searched in regular expressions.</span>
</div>


---
##### sFlow Counter Sample
<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Date and time | It is the date and time when the SFLOW sample received.|
| Sending source | It is the source IP.|
| Type | The type of counter sample.(I/F, CPU, Memory, Disk, Netowk) |
| Data | Data of counter sample.|
</div>

>>>

#### sFlow Counter sample filter

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Start date and time | Specify the date and time of the search start.|
| End date and time | Specify the date and time of the search termination.|
| Submit | Specify the source.|
| Type | Specify the type of counter sample.|

<Span style = "color: red"> Character strings can be searched in regular expressions.</span>
</div>

---
#### sFlow report

<div class="text-xl mb-2">

| Report name | Contents |
| ---- | ---- |
| Heat map | Heat map showing the number of log time.|
| Communications | This is a time -series graph of communication volume.|
| TOP List | This is a ranking report for each item.|
| TOP list (3D) | Ranking reports for each item are displayed in 3D graphs.|
| IP Pair Flow | The combination of communication is displayed in a graph.|
| FFT Analysis | Analyzing the communication cycle with FFT.|
| FFT Analysis (3D) | Analyze the communication cycle with FFT and display it on the 3D graph.|
| Map | Display the position of the IP address on the map.|
</div>

---
#### sFlow Counter sample report

<div class="text-xl mb-2">

| Report name | Contents |
| ---- | ---- |
| Heat map | Heat map showing the number of log time.|
| I/F BPS | This is a time -series graph of the communication volume (Bytes/Sec) obtained from I/F sample.|
| I/F PPS | This is a time -series graph of the communication volume (Packets/Sec) obtained from I/F sample.|
| CPU | CPU usage rate and load time series graph acquired from CPU sample.|
| Memory | Memory is a graph of memory usage and available capacity obtained from a sample.|
| DISK | This is a graph of disk usage and access amount acquired from Disk sample.|
| Network | Network is a graph of network usage obtained from sample.|
</div>

