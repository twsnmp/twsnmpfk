#### PING

<div class="text-xl mb-2 text-left">
This is the screen to execute ping.<br>
<Span style = "color: red;"> To get a location information, you need a Geoip database file.</span>
</div>

![PING](../../help/en/2023-12-03_11-20-46.png)

>>>
#### Explanation

<div class="text-lg">

| Items | Contents |
| ---- | ---- |
| IP address | This is the IP address to run ping.|
| Number of times | Ping is the number of execution times.|
| Size | Ping packet size.<br> The change mode is executed while increasing the size.|
| TTL | TTL value of ping packet.<br> The trace route runs while increasing the TTL value.|
Result Graph | Ping's execution result is a graph of the response time, TTL value.|
| Results | Ping execution results.<br> As a result, the date and time of implementation, the response time, the size, the transmission reception TTL, the source IP, the location |
| Beep | Ping will be informed by sound.|
| Start | Start ping.|
| Stop | Ping stops.|
| Close | Ends ping.|
</div>


---
#### histogram

<div class="text-xl mb-2 text-left">
It is a histogram of response time.
</div>

![Histogram](../../help/en/2023-12-03_11-22-17.png)

---
#### 3D analysis

<div class="text-xl mb-2 text-left">
The response time, size, and implementation date and time are displayed in 3D graphs.
</div>

![PING 3D](../../help/en/2023-12-03_11-22-27.png)

---
#### Line prediction

<div class="text-xl mb-2 text-left">

From the change in response time if the size is changed
This is a report that predicts the line speed.
</div>

![Line speed prediction](../../help/en/2023-12-03_11-23-21.png)

---
#### Route analysis

<div class="text-xl mb-2 text-left">
Display location information.It cannot be displayed without a GEOIP database.
</div>

![Route analysis](../../help/en/2023-12-03_11-24-05.png)
