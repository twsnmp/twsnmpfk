# PING

Run ICMP PING diagnostics against a target IP address, measuring response times, jitter, packet loss, and route paths (Traceroute / MTR) in real-time.

## Operation Modes

You can intuitively switch between four operation modes using the mode buttons at the top of the window:

* **Normal Ping**
  Sends standard ICMP Echo Requests. You can specify the ping count, payload size, and TTL.
* **Smoke**
  Sends high-frequency ICMP pings (at 200ms intervals) for short periods or continuously to measure response times, jitter (latency variation), and packet loss with high precision.
* **Traceroute**
  Sequentially increases the TTL starting from 1 to discover the network router hops along the path to the target.
* **MTR (My Traceroute)**
  Automatically discovers all hop routers along the path, then performs continuous sampling (pinging) to each hop to track real-time response times, jitter, and packet loss rates per hop.

## Settings Parameters

* **IP Address**
  The target IP address or hostname to ping.
* **Count / Duration**
  - **Normal Ping**: Number of packets to send (1, 3, 5, 10, 20, 30, 50, 100, or Continuous).
  - **Smoke / MTR**: Measurement duration (1 min, 3 min, 5 min, 10 min, or Continuous). Default for MTR is 1 minute.
* **Size**
  Packet payload size in bytes (64, 128, 256, 512, 1024, 1500). Select "Inc Size" in Normal Ping mode to ping while increasing packet sizes.
* **TTL**
  Time-To-Live value. Available in Normal Ping mode (1 to 254).
* **BEEP**
  Toggle sound notifications for ping responses (success/fail).

## Button Descriptions

* **[Start]** : Begin sending PING packets in the selected mode.
* **[Stop]** : Stop the active ping session.
* **[AI Explain]** : Send the displayed charts, MTR, Smoke, and diagnostic data to AI (LLM) for automated analysis on network performance, loss, and route issues (available when AI integration is enabled).
* **[Help]** : Open this help document.
* **[Close]** : Close the PING window.

## Diagnostic Chart Tabs

* **Do ping (Standard)**
  Time-series chart showing response times and TTL values, alongside a detailed logs table.
* **MTR**
  Available in MTR mode. Displays the hop topology diagram (Hop Flow), a statistics table (Sent, Avg, Best, Worst, StDev, Loss %), and a hop latency profile chart.
* **Smokeping**
  Available in Smoke mode. Displays a Smokeping-style chart visualizing median RTT, latency spread, jitter, and packet loss intensity with gradient color coding.
* **Histogram**
  Frequency distribution of ping response times.
* **3D Analysis**
  3D scatter plot representing response times, packet sizes, and timestamps.
* **Line Prediction**
  Bandwidth and line speed predictor based on response times with variable packet sizes.
* **World Map**
  Geographic path visualizer showing network hops on a map (requires GeoIP database).
