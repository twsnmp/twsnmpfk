#### NetFlow

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Date and time | The date and time of receiving Netflow.|
| Sending source | It is the source IP.|
| Port | It is the port number of the source.|
| Location | It is the source of the source.GEOIP DB is required.|
| Address | IP for the destination.|
| Port | The destination port number.|
| Location | It is the destination position.GEOIP DB is required.|
| Protocol | Protocol such as TCP/UDP/ICMP.|
| TCP flag | TCP flag.|
| Packet | The number of sending packets.|
| Byte | The number of sending bytes.|
| Period | Flow communication time.|

</div>

>>>
#### NetFlow(button)

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Filter | Specify the search conditions and display Netflow.|
| <Span style = "color: red;"> Delete all logs </span> | Delete all Netflow.|
| Copy | Copy the selected log.|
| Report | Displays Netflow analysis reports.|
| CSV | Export Netflow to CSV file.|
| Excel | Export Netflow to Excel file.|
| Update | Update the Netflow list to the latest state.|

</div>


---

#### NetFlwo Filter

<div class="text-sm">

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
#### NetFlow Report

<div class="text-xl mb-2">

| Report name | Contents |
| ---- | ---- |
| Heat Map | Heat map by time zone of Netflow's receiving number.|
| Histogram | Histgram of numerical data.|
| Traffic | This is a time -series graph of traffic.|
| TOP List | This is a ranking report for each item.|
| TOP list (3D) | Ranking reports for each item are displayed in 3D graphs.|
| IP Pair Flow | The combination of communication is displayed in a graph.|
| FFT Analysis | Analyzing the communication cycle with FFT.|
| FFT Analysis (3D) | Analyze the communication cycle with FFT and display it on the 3D graph.|
| Map | Display the position of the IP address on the map.|

</div>

