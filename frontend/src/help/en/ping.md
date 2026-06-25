# PING

Run ICMP PING diagnostics against a target IP address, measuring response times and route paths in real-time.

## Settings Parameters

* **IP Address**
  The target IP address to ping.
* **Count**
  The number of ping packets to send (Continuous, 1, 3, 5, 10, 20, 30, 50, 100).
* **Size**
  Packet payload size in bytes. Select "Increase" to run pings while gradually increasing the packet size.
* **TTL**
  Time-To-Live value. Select "Trace Route" to increase the TTL sequentially to trace the network hops.
* **BEEP**
  Toggle to play a sound notification for successful or failed ping responses.

## Button Descriptions

* **[Start]** : Begin sending PING packets.
* **[Stop]** : Stop sending PING packets.
* **[Help]** : Open this help document.
* **[Close]** : Close the PING tool window.

## Diagnostic Chart Tabs

* **Standard Chart**
  Time-series chart showing response times and TTL values.
* **Histogram**
  Frequency distribution of ping response times.
* **3D Analysis**
  3D scatter plot representing response times, packet sizes, and timestamps.
* **Line Prediction**
  Bandwidth and transmission rate estimator based on response times of variable packet sizes.
* **Route Analysis**
  Geographic path visualizer showing network hops on a map (requires GeoIP database).
