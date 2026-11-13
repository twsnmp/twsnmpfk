# TWSNMP FK MCP Server Specification

## Overview

This document defines the specification for the Model Context Protocol (MCP) server implemented in TWSNMP FK (`backend/mcp.go`, `backend/mcp_tools.go`, `backend/mcp_tools_crud.go`, `backend/mcp_tools_analytics.go`).

The MCP server exposes 58 tools and 59 prompts allowing AI agents to query monitoring data, inspect flows/telemetry/logs, manage nodes/pollings/lines/networks/certificates, run live diagnostics, and organize the map.

## Server Endpoint and Transports

The server starts based on `datastore.MapConf.MCPTransport` (`"off"`, `"sse"`, `"stream"`):
- `"stream"` (Streamable HTTP, default when enabled): serves `/mcp`
- `"sse"` (Server-Sent Events): serves `/sse` and `/message`
- `"off"`: MCP server is disabled

The HTTP server binds to `datastore.MapConf.MCPEndpoint` (for example `127.0.0.1:8089` or `:8089`). TLS 1.3 is enabled automatically when `-mcpCert` and `-mcpKey` are provided.

## Authentication and Access Control

All MCP HTTP endpoints are guarded by `checkMCPACL`. When both token authentication and an IP allowlist are configured, both checks must pass:

1. **Token Authentication (`datastore.MapConf.MCPToken`)**:
   - When non-empty, clients must send an `Authorization` header whose token matches `MCPToken` (constant-time comparison).
   - Both `Authorization: Bearer <token>` (case-insensitive `Bearer ` prefix) and `Authorization: <token>` are accepted.
2. **IP Access Control (`datastore.MapConf.MCPFrom`)**:
   - Comma-separated list of client IP addresses.
   - Reloaded atomically when MCP configuration changes (stale IPs are removed).
   - If empty, all client IPs are permitted.
3. **Credential Masking**:
   - Secret fields (`Community`, `Password`, `GNMIPassword`, `PublicKey`) returned by `get_node` and `get_network` are masked as `"********"` when set.

For prompt definitions, see [MCP Server Prompt Specifications](./mcp_prompt.md).

---

## Tool Specifications (58 Tools)

### Node Management

#### `get_node_list`
Get a list of nodes registered in TWSNMP.
- `name_filter` (string, optional): Regex filter on node name.
- `ip_filter` (string, optional): Regex filter on node IP address.
- `state_filter` (string, optional): Regex filter on node state (`normal`, `warn`, `low`, `high`, `repair`, `unknown`).

#### `get_node`
Get full non-secret configuration details and associated pollings for a single node (`Community`, `Password`, `GNMIPassword`, `PublicKey` are masked).
- `id` (string, required): Node ID, unique node name, or unique IP address.

#### `add_node`
Add a new node to TWSNMP.
- `name` (string, required): Node name.
- `ip` (string, required): Node IP address.
- `icon` (string, optional): Icon name (default `desktop`).
- `description` (string, optional): Node description.
- `x` (integer, optional): Map X position (`1`–`1000`, default `64`).
- `y` (integer, optional): Map Y position (`1`–`1000`, default `64`).

#### `update_node`
Update node metadata or map coordinates (`id` accepts node ID, unique name, or unique IP).
- `id` (string, required): Node ID, current name, or current IP.
- `name` (string, optional): New node name.
- `ip` (string, optional): New IP address.
- `icon` (string, optional): New icon name.
- `description` (string, optional): Description text.
- `x` (integer, optional): X position.
- `y` (integer, optional): Y position.

#### `delete_node`
Delete a node along with all of its pollings (stopping active gNMI/MQTT subscriptions) and connected map lines.
- `id` (string, required): Node ID, unique node name, or unique IP address.

---

### Polling Management & Execution

#### `get_polling_list`
Get a list of pollings registered in TWSNMP.
- `type_filter` (string, optional): Regex filter on polling type (`ping`, `tcp`, `http`, `dns`, `twsnmp`, `syslog`, `stun`, `snmp`, etc.).
- `name_filter` (string, optional): Regex filter on polling name.
- `node_name_filter` (string, optional): Regex filter on node name.
- `state_filter` (string, optional): Regex filter on polling state (`normal`, `warn`, `low`, `high`, `repair`, `unknown`).
- `node_id` (string, optional): Exact node ID or unique node name filter.

#### `get_polling`
Get full configuration and latest result for a single polling check.
- `id` (string, required): Polling ID.

#### `add_polling`
Add a new polling check to an existing node.
- `node_id` (string, required): Target node ID, unique node name, or unique IP address.
- `name` (string, required): Polling name.
- `type` (string, required): Polling type (`ping`, `tcp`, `http`, `dns`, `tls`, `snmp`, `gnmi`, `mqtt`, `syslog`, `trap`, `netflow`, `sflow`, `twsnmp`, `ntp`, `ssh`, `stun`, `cmd`, `report`, etc.).
- `mode` (string, optional): Polling sub-mode (`get`, `walk`, `count`, `subscribe`, etc.).
- `params` (string, optional): Target parameter (port, URL, OID, hostname, etc.).
- `filter` (string, optional): Filter pattern.
- `extractor` (string, optional): Grok / value extractor pattern.
- `script` (string, optional): JavaScript evaluation script.
- `level` (string, optional): Severity level: `off`, `low` (default), `warn`, `high`.
- `poll_int` (integer, optional): Interval in seconds (default map `PollInt` or `60`).
- `timeout` (integer, optional): Timeout in seconds (default map `Timeout` or `3`).
- `retry` (integer, optional): Retry count (default map `Retry` or `0`).
- `log_mode` (string, optional): Logging mode: `none` (default), `always`, `on_change`, `ai`.
- `fail_action` (string, optional): Action on failure.
- `repair_action` (string, optional): Action on recovery.
- `ai_mode` (string, optional): AI anomaly detection mode.
- `vector_cols` (string, optional): Comma-separated vector columns for AI analysis.
- `mqtt_url` (string, optional): MQTT broker URL.
- `mqtt_topic` (string, optional): MQTT subscription topic.
- `mqtt_cols` (string, optional): MQTT JSON field extraction columns.

#### `update_polling`
Update an existing polling configuration. If target node, mode, type, params, or MQTT settings change while running gNMI/MQTT subscribe, existing subscriptions are stopped cleanly.
- `id` (string, required): Polling ID.
- `node_id`, `name`, `type`, `mode`, `params`, `filter`, `extractor`, `script`, `level`, `poll_int` (`>0`), `timeout` (`>0`), `retry` (`>0` to set, `<0` to reset to `0`), `log_mode`, `fail_action`, `repair_action`, `ai_mode`, `vector_cols`, `mqtt_url`, `mqtt_topic`, `mqtt_cols` (all optional).

#### `delete_polling`
Delete a polling entry and stop any active gNMI/MQTT subscription for it.
- `id` (string, required): Polling ID.

#### `check_polling`
Trigger immediate execution of a specific polling check, all pollings on a node, or all pollings across TWSNMP.
- `polling_id` (string, optional): Specific polling ID.
- `node_id` (string, optional): Node ID, unique name, or IP.
- `all` (boolean, optional): Set `true` to check all pollings and networks.

#### `get_polling_log`
Get historical polling execution logs as JSON.
- `id` (string, required): Polling ID.
- `limit` (integer, optional): Max records (`1`–`2000`, default `100`).

#### `get_polling_log_data`
Get historical polling numeric metrics formatted as CSV with deterministic column order (`time,state,<sorted metric keys>`).
- `id` (string, required): Polling ID.
- `limit` (integer, optional): Max rows (`1`–`2000`, default `100`).

---

### Map Lines & Network Containers

#### `get_line_list`
List map topology lines with resolved endpoint names and line state.
- `node_id` (string, optional): Filter lines touching a specific node or network ID/name/IP.
- `state_filter` (string, optional): Regex filter on overall line state.

#### `add_line`
Connect two nodes or networks on the map (`NET:<id>` or unique name/IP is supported for networks). Refreshes line state immediately after creation.
- `node_id1` (string, required): First endpoint ID, unique name, or unique IP.
- `node_id2` (string, required): Second endpoint ID, unique name, or unique IP.
- `polling_id1` (string, optional): Polling ID on endpoint 1 (or port ID if endpoint 1 is a network).
- `polling_id2` (string, optional): Polling ID on endpoint 2 (or port ID if endpoint 2 is a network).
- `polling_id` (string, optional): Traffic polling ID for the line.
- `width` (integer, optional): Line width (`1`–`10`, default `2`).
- `port` (string, optional): Port label text.
- `info` (string, optional): Additional info text.

#### `delete_line`
Delete a map line by ID.
- `id` (string, required): Line ID.

#### `get_network_list`
List network containers registered in TWSNMP.
- `name_filter` (string, optional): Regex filter on network name.
- `ip_filter` (string, optional): Regex filter on IP address.

#### `get_network`
Get detailed configuration and port table for a network container (SNMP credentials masked).
- `id` (string, required): Network ID (with or without `NET:` prefix), unique name, or unique IP.

#### `add_network`
Add a network/switch container node to the map.
- `name` (string, required): Network name.
- `ip` (string, required): Management IP address.
- `description`, `snmp_mode`, `community`, `user`, `password`, `url` (string, optional).
- `snmp_port` (integer, optional, default `161`), `h_ports` (integer, optional, default `24`), `x` (optional, default `64`), `y` (optional, default `64`).
- `arp_watch`, `unmanaged`, `lldp` (boolean, optional).

#### `update_network`
Update an existing network container.
- `id` (string, required): Network ID (with or without `NET:` prefix), unique name, or unique IP.
- `name`, `ip`, `description`, `snmp_mode`, `community`, `user`, `password`, `url` (string, optional).
- `snmp_port` (`1`–`65535`), `h_ports` (`>0`), `x` (`>0`), `y` (`>0`) (integer, optional).
- `arp_watch`, `unmanaged`, `lldp` (string, optional: `"true"`, `"false"`, or empty to preserve).

#### `delete_network`
Delete a network container and any lines connected to it.
- `id` (string, required): Network ID (with or without `NET:` prefix), unique name, or unique IP.

#### `check_network`
Trigger SNMP/LLDP status checks for one network container or all network containers.
- `id` (string, optional): Network ID (with or without `NET:` prefix), unique name, or unique IP.
- `all` (boolean, optional): Set `true` to check all networks.

#### `auto_connect_lines`
Automatically infer and connect missing map topology lines.
- `mode` (string, optional): `"strict"` (default) or `"speculative"`.

#### `auto_layout`
Automatically arrange map node coordinates or undo the previous auto-layout change.
- `mode` (string, optional): `"hierarchical"` (default), `"cluster"`, or `"categorized"`.
- `undo` (boolean, optional): Set `true` to restore node positions from the previous auto-layout.

---

### Live Diagnostics, SNMP, MIB & gNMI

#### `do_ping`
Execute an ICMP ping to a target from TWSNMP.
- `target` (string, required): Target IP, hostname, node ID, or node name.
- `size` (integer, optional): Packet size (`1`–`1500`, default `64`).
- `ttl` (integer, optional): IP TTL (`1`–`255`, default `254`).
- `timeout` (integer, optional): Timeout seconds (`1`–`10`, default `3`).

#### `diagnose_node`
Run a unified multi-layer diagnostic probe against a target (managed node resolution, currently failing pollings, ICMP ping, SNMP `sysDescr`/`sysObjectID`/`sysUpTime`/`sysName`/`sysLocation`, TCP port scans with banners, and HTTP/HTTPS web signature inspection).
- `target` (string, required): Node ID, node name, IP address, or hostname.
- `ports` (string, optional): Comma-separated TCP ports (default `"22,80,443,8080,8443"`, max 32 ports).
- `timeout` (integer, optional): Probe timeout in seconds (`1`–`10`, default `3`).

#### `get_mib_tree`
Retrieve the SNMP MIB object tree loaded in TWSNMP.
- `root_oid` (string, optional): Optional subtree OID or object name (e.g. `.1.3.6.1.2.1.1` or `system`).
- `max_depth` (integer, optional): Optional max traversal depth (`0` = unlimited).

#### `search_mib`
Search loaded SNMP MIB definitions by name, OID, or description regex/substring.
- `query` (string, required): Case-insensitive regex or substring.
- `limit` (integer, optional): Max results (`1`–`500`, default `50`).

#### `snmpwalk`
Execute an SNMP walk on a target device. When `target` resolves to a managed node and credentials are omitted, the node's saved SNMP settings and `SnmpPort` are used automatically.
- `target` (string, required): Target IP, hostname, node ID, or node name.
- `mib_object_name` (string, required): MIB object name or numeric OID.
- `snmp_mode` (string, optional): `v2c`, `v3auth`, `v3authpriv`, `v3authprivex`, `v3sha256aes128`, `v3sha512aes256`.
- `community`, `user`, `password` (string, optional): SNMP credentials override.
- `port` (integer, optional): SNMP UDP port override (default node `SnmpPort` or `161`).
- `limit` (integer, optional): Max variables (`1`–`50000`, default unlimited).

#### `snmpset`
Write a writable SNMP MIB variable (`integer` or `string`) on a target device.
- `target` (string, required): Target IP, hostname, node ID, or node name.
- `mib_object_name` (string, required): MIB object name or numeric OID.
- `type` (string, required): Value type (`"integer"` or `"string"`).
- `value` (string, required): Value to set.
- `snmp_mode`, `community`, `user`, `password`, `port` (optional overrides).

#### `gnmi_capabilities`
Query gNMI capabilities (supported encodings, version, YANG models) for a managed node.
- `node_id` (string, required): Node ID, unique node name, or unique IP.
- `target` (string, optional): Optional `host:port` override (default node IP + `GNMIPort` or `57400`).

#### `gnmi_get`
Fetch gNMI telemetry data at a path from a managed node.
- `node_id` (string, required): Node ID, unique node name, or unique IP.
- `path` (string, required): gNMI path (e.g. `/interfaces/interface` or `/system`).
- `target` (string, optional): Optional `host:port` override.
- `encoding` (string, optional): `json_ietf` (default unless set on node), `json`, `proto`, `bytes`, `ascii`.
- `limit` (integer, optional): Max leaf entries (`1`–`5000`, default `500`).

---

### IP, MAC, Switch FDB & Wireless Discovery

#### `get_ip_address_list`
Get IP-to-MAC entries from TWSNMP's ARP/IP table.
- `ip_filter`, `mac_filter`, `vendor_filter` (string, optional): Regex filters.
- `limit` (integer, optional): Max entries (`1`–`2000`, default `500`).

#### `get_ip_address_info`
Inspect a single IP address (MAC, vendor, location, DNS names, node link).
- `ip` (string, required): IPv4/IPv6 address.

#### `get_mac_address_list`
List MAC address inventory aggregated from ARP/IP reports.
- `mac_filter`, `ip_filter`, `vendor_filter` (string, optional): Regex filters.
- `limit` (integer, optional): Max entries (`1`–`2000`, default `500`).

#### `get_mac_address_info`
Inspect a single MAC address (associated IPs, node/port links, vendor).
- `mac` (string, required): MAC address.

#### `get_wifi_ap_list`
List detected Wi-Fi access points from wireless reports (compact by default).
- `ssid_filter`, `bssid_filter`, `host_filter`, `vendor_filter` (string, optional): Regex filters.
- `include_history` (boolean, optional): Include `RSSI` time-series array (default `false`).
- `limit` (integer, optional): Max entries (`1`–`500`, default `100`).

#### `get_bluetooth_device_list`
List detected Bluetooth / BLE devices from sensor reports (compact by default).
- `name_filter`, `address_filter`, `host_filter`, `vendor_filter` (string, optional): Regex filters.
- `public_address_only` (boolean, optional): Exclude random/private addresses when `true`.
- `include_history` (boolean, optional): Include `RSSI` time-series array (default `false`).
- `limit` (integer, optional): Max entries (`1`–`500`, default `100`).

#### `get_fdb_table`
Fetch the live SNMP bridge Forwarding Database (MAC-to-switch-port mapping) from a network/switch container.
- `id` (string, required): Network ID (with or without `NET:` prefix), unique name, or unique IP.

---

### System Status & Certificates

#### `get_system_status`
Get TWSNMP FK runtime health, uptime, CPU/goroutines, DB size, datastore path, ping mode, MCP config/status (`MCPTransport`, `MCPEndpoint`, `MCPStatus`, `MCPError`), node/polling/line/network/certificate summary counts, latest resource monitor snapshot, and receiver status map (`syslogd`, `trapd`, `netflowd`, `sflowd`, `otel`, `mqtt`, `arpwatch`, `sshd`, `tcpd`).
- Parameters: None.

#### `get_resource_monitor_list`
Get CPU, memory, disk, swap, network, connection, DB size, and process snapshots collected by TWSNMP.
- `limit` (integer, optional): Max snapshots (`1`–`500`, default `100`).

#### `get_server_certificate_list`
List all TLS server certificate monitor entries.
- Parameters: None.

#### `update_server_certificate`
Create a new TLS server certificate monitor (when `id` is empty) or update an existing entry (`target` / `port`), resetting state and triggering re-verification.
- `id` (string, optional): Existing entry ID (omit to add new).
- `target` (string, optional): Hostname or IP (required when adding).
- `port` (integer, optional): TLS port (`1`–`65535`, default `443` when adding).

#### `delete_server_certificate`
Delete a TLS server certificate monitor entry by ID.
- `id` (string, required): Entry ID.

---

### Logs, Flows, OpenTelemetry, AI Anomaly & Reports

Time-range fields (`start_time`, `end_time`) support relative durations (e.g. `"-1h"`, `"-24h"`), `"now"`, and RFC3339 / common timestamps.

#### `search_event_log`
Search TWSNMP event logs.
- `node_filter`, `type_filter`, `level_filter` (`warn`, `low`, `high`, `debug`, `info`), `event_filter` (string, optional): Regex filters.
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `limit` (integer, optional): Max logs (`1`–`10000`, default `100`).

#### `add_event_log`
Append an event log entry (event type is recorded as `"mcp"`).
- `event` (string, required): Event message.
- `level` (string, optional): `info` (default), `normal`, `warn`, `low`, `high`.
- `node` (string, optional): Associated node ID or node name.

#### `search_syslog`
Search stored syslog messages.
- `level_filter`, `host_filter`, `tag_filter`, `message_filter` (string, optional): Regex filters.
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `limit` (integer, optional): Max logs (`1`–`10000`, default `100`).

#### `get_syslog_summary`
Aggregate normalized syslog patterns ordered by frequency.
- `level_filter`, `host_filter`, `tag_filter`, `message_filter` (string, optional): Regex filters.
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `top_n` (integer, optional): Top N patterns (`5`–`500`, default `5`).
- `max_scan` (integer, optional): Max syslog records to scan (`1000`–`2000000`, default `200000`).

#### `search_snmp_trap_log`
Search received SNMP traps.
- `from_filter`, `trap_type_filter`, `variable_filter` (string, optional): Regex filters.
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `limit` (integer, optional): Max traps (`1`–`10000`, default `100`).

#### `search_netflow`
Search NetFlow records.
- `src_filter`, `dst_filter`, `protocol_filter` (string, optional): Regex filters.
- `src_port`, `dst_port` (integer, optional): Exact port filter (`>0`).
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `limit` (integer, optional): Max records (`1`–`2000`, default `200`).

#### `search_sflow`
Search sFlow flow samples.
- Same parameters as `search_netflow`.

#### `search_sflow_counter`
Search sFlow interface/system counter records.
- `remote_filter`, `type_filter` (string, optional): Regex filters on exporter IP and counter type.
- `start_time`, `end_time` (string, optional): Default `-1h` to `now`.
- `limit` (integer, optional): Max records (`1`–`2000`, default `200`).

#### `get_otel_metrics`
List OpenTelemetry metrics received by TWSNMP.
- `host_filter`, `service_filter`, `name_filter` (string, optional): Regex filters.
- `limit` (integer, optional): Max metrics (`1`–`500`, default `100`).

#### `get_otel_traces`
List OpenTelemetry trace summaries.
- `service_filter` (string, optional): Regex filter on service name.
- `error_only` (boolean, optional): Only return traces containing errors when `true`.
- `limit` (integer, optional): Max traces (`1`–`200`, default `50`).

#### `get_otel_trace_detail`
Inspect all spans (parent/child IDs, service, operation, attributes, events, error status) for an OpenTelemetry trace.
- `trace_id` (string, required): OpenTelemetry `TraceID`.
- `bucket` (string, optional): Specific OTel bucket timestamp (searches all recent buckets if omitted).

#### `get_ai_anomaly_list`
List AI anomaly detection results and severity classification (`high`, `low`, `warn`, `normal`) based on configured `AIConf` thresholds.
- `node_id` (string, optional): Node ID, unique name, or IP filter.
- `min_score` (number, optional): Minimum latest anomaly score threshold.
- `include_history` (boolean, optional): Include full score time-series (`ScoreData`) when `true` (default `false`).
- `limit` (integer, optional): Max entries (`1`–`1000`, default `100`).

#### `get_sensor_list`
List connected telemetry sensors (`twWifiScan`, `twBlueScan`, `twpcap`, `twwinlog`, `twagentneo`, etc.).
- `type_filter`, `state_filter` (string, optional): Regex filters.
- `include_history` (boolean, optional): Include `Stats` and `Monitors` time-series arrays (default `false` for compact summaries).
- `limit` (integer, optional): Max sensors (`1`–`500`, default `100`).

#### `get_report_summary`
Query analytics report entries (`wifiAP`, `blueDevice`, `envMonitor`, `powerMonitor`, `motionSensor`, `etherType`, `dnsq`, `radiusFlow`, `tlsFlow`, `winEventID`, `winLogon`, `winAccount`, `winKerberos`, `winPrivilege`, `winProcess`, `winTask`, `flow`, `server`, `fumble`, `syslogStats`, `trapStats`, `twAgentInfo`, `twAgentAlert`, `twAgentFIM`, `twAgentSCA`, `twAgentTCP`).
- `report_type` (string, required): Report kind name.
- `filter` (string, optional): Regex filter applied to the serialized JSON of each record.
- `include_history` (boolean, optional): Include historical arrays such as `RSSI`, `EnvData`, `Data` (default `false` for compact output).
- `limit` (integer, optional): Max records (`1`–`200`, default `50`).
