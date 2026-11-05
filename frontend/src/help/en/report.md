# Comprehensive Report (Report)

Aggregates, visualizes, and analyzes telemetry, events, traffic, and environmental metrics collected from various sensors and endpoint agents.

## Tabs

* **[Sensors]** : Displays a list of all registered sensors (twWifiScan, twBlueScan, twpcap, twwinlog, twagentneo, Syslog, Trap, MQTT, etc.), including health state, sensor type, message counters, last seen timestamp, and telemetry/resource monitoring modals.
* **[Wi-Fi AP]** : Shows nearby Wi-Fi access points detected via twWifiScan (SSID, BSSID, RSSI, channel, vendor, encryption), along with channel distribution charts.
* **[Bluetooth (BLE)]** : Proximity detection report for Bluetooth / BLE beacons and devices (smartwatches, tags, beacons) discovered via twBlueScan, including vendor share pie charts, RSSI distribution, and detected device tables.
* **[Environmental Sensors]** : Collects and charts readings from environmental sensors (temperature, humidity, atmospheric pressure, CO2), power monitoring plugs, and motion/light sensors. Supports sub-tabs for "Environmental Sensors", "Power Sensors", and "Motion Sensors".
* **[Pcap Packet Analysis]** : Analyzes packet data captured via twpcap, presenting breakdown statistics for EtherTypes, DNS queries and responses (with cached reverse lookup), RADIUS authentication flows, and TLS encrypted session flows.
* **[Windows Events]** : Aggregates and audits Windows security event logs collected via twwinlog (Event ID, logon/logoff, account management, Kerberos authentication, privilege usage, process creation, scheduled tasks), featuring KPI cards and detailed event modals.
* **[twagentneo]** : Monitors endpoint operational status, AI threat analysis (via tensai), File Integrity Monitoring (FIM), Security Configuration Assessment (SCA), and active process TCP flows from twagentneo agents.

## Action Buttons

* **[Clear All]** : Clears all report data in the currently active tab from the database.
* **[CSV]** : Exports the displayed report data to a CSV file.
* **[Excel]** : Exports the displayed report data to an Excel (.xlsx) file.
* **[Reload]** : Reloads report data with the latest state.
