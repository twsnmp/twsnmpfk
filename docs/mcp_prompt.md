# MCP Server Prompt Specifications

This document outlines the specifications for the 59 prompts available on the TWSNMP FK MCP server (`backend/mcp_prompts.go` and `backend/mcp_prompts_extended.go`).

## Core Prompts (1–19)

### 1. `get_node_list`

- **Title:** Get node list with filters
- **Description:** Get a list of nodes registered in TWSNMP with filters.
- **Arguments:**
  - `state_filter` (Optional): Node state filter (`normal`, `repair`, `warn`, `low`, `high`, `unknown`).
  - `name_filter` (Optional): Node name filter.
  - `ip_filter` (Optional): Node IP address filter.

### 2. `add_node`

- **Title:** Add a new node to TWSNMP.
- **Description:** Add a new node to TWSNMP.
- **Arguments:**
  - `name` (Required): Name of the new node.
  - `ip` (Required): IP address of the new node.
  - `icon` (Optional): Icon for the node (`desktop`, `laptop`, `server`, `cloud`, `router`, `ip`, etc.).
  - `description` (Optional): Description of the new node.
  - `position` (Optional): Position of the new node (e.g., `x=100,y=200`).

### 3. `update_node`

- **Title:** Update node on TWSNMP.
- **Description:** Update an existing node on TWSNMP.
- **Arguments:**
  - `id` (Required): The ID, name, or IP address of the node to update.
  - `name` (Optional): New name for the node.
  - `ip` (Optional): New IP address for the node.
  - `icon` (Optional): New icon for the node.
  - `description` (Optional): New description for the node.
  - `position` (Optional): New position for the node (e.g., `x=100,y=200`).

### 4. `get_network_list`

- **Title:** Get network node list with filters
- **Description:** Get a list of network nodes registered in TWSNMP with filters.
- **Arguments:**
  - `name_filter` (Optional): Network node name filter.
  - `ip_filter` (Optional): Network node IP address filter.

### 5. `get_polling_list`

- **Title:** Get polling list with filters
- **Description:** Get a list of pollings registered in TWSNMP with filters.
- **Arguments:**
  - `type_filter` (Optional): Polling type filter (`ping`, `snmp`, `syslog`, `http`, `tcp`, etc.).
  - `name_filter` (Optional): Polling name filter.
  - `node_name_filter` (Optional): Node name filter.
  - `state_filter` (Optional): Polling state filter (`normal`, `repair`, `warn`, `low`, `high`, `unknown`).
  - `node_id` (Optional): Exact node ID or unique node name filter.

### 6. `get_polling_log`

- **Title:** Get polling log.
- **Description:** Get polling log from TWSNMP.
- **Arguments:**
  - `id` (Required): Polling ID.
  - `limit` (Optional): Max number of logs to retrieve (default `100`).
  - `start_time` (Optional): Start time filter (duration like `"-24h"` or RFC3339).
  - `end_time` (Optional): End time filter (default `now`).

### 7. `do_ping`

- **Title:** Do ping
- **Description:** Perform a ping to a target from TWSNMP.
- **Arguments:**
  - `target` (Required): Ping target (IP, node name, or hostname).
  - `size` (Optional): Ping packet size (default `64`).
  - `ttl` (Optional): Ping packet TTL (default `254`).
  - `timeout` (Optional): Ping timeout in seconds (default `3`).

### 8. `snmpwalk`

- **Title:** Do snmpwalk
- **Description:** Perform an snmpwalk to a target from TWSNMP.
- **Arguments:**
  - `target` (Required): SNMP walk target (IP, node name, or hostname).
  - `mib_object_name` (Required): MIB object name or OID to walk.
  - `snmp_mode` (Optional): SNMP mode (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`, `v3sha256aes128`, `v3sha512aes256`).
  - `community` (Optional): Community name for v2c mode.
  - `user` (Optional): User name for v3 mode.
  - `password` (Optional): Password for v3 mode.
  - `port` (Optional): SNMP UDP port (default node `SnmpPort` or `161`).
  - `limit` (Optional): Maximum number of variables to return.

### 9. `search_event_log`

- **Title:** Search event log with filters
- **Description:** Search event logs with filters.
- **Arguments:**
  - `type_filter` (Optional): Event type filter (`system`, `polling`, `arpwatch`, `mcp`).
  - `node_filter` (Optional): Node filter.
  - `level_filter` (Optional): Level filter (`info`, `normal`, `warn`, `low`, `high`).
  - `event_filter` (Optional): Event filter.
  - `start_time` (Optional): Start time (default `-1h`).
  - `end_time` (Optional): End time (default `now`).
  - `limit` (Optional): Max logs to search.

### 10. `search_syslog`

- **Title:** Search syslog with filters
- **Description:** Search syslogs with filters.
- **Arguments:**
  - `level_filter` (Optional): Level filter (`warn`, `low`, `high`, `debug`, `info`).
  - `host_filter` (Optional): Sender host filter.
  - `tag_filter` (Optional): Syslog tag filter.
  - `message_filter` (Optional): Syslog message filter.
  - `start_time` (Optional): Start time (default `-1h`).
  - `end_time` (Optional): End time (default `now`).
  - `limit` (Optional): Max syslogs to search.

### 11. `get_syslog_summary`

- **Title:** Get syslog summary with filters
- **Description:** Get a summary of syslogs with filters.
- **Arguments:**
  - `level_filter` (Optional): Level filter (`warn`, `low`, `high`, `debug`, `info`).
  - `host_filter` (Optional): Sender host filter.
  - `tag_filter` (Optional): Syslog tag filter.
  - `message_filter` (Optional): Syslog message filter.
  - `start_time` (Optional): Start time (default `-1h`).
  - `end_time` (Optional): End time (default `now`).
  - `top_n` (Optional): Top syslog pattern count.

### 12. `search_snmp_trap_log`

- **Title:** Search snmp trap log with filters
- **Description:** Search SNMP trap logs of TWSNMP with filters.
- **Arguments:**
  - `from_filter` (Optional): Trap sender filter.
  - `trap_type_filter` (Optional): Trap type filter.
  - `variable_filter` (Optional): Trap variable filter.
  - `start_time` (Optional): Start time (default `-1h`).
  - `end_time` (Optional): End time (default `now`).
  - `limit` (Optional): Max trap logs to search.

### 13. `get_mib_tree`

- **Title:** Get MIB tree of TWSNMP.
- **Description:** Get the MIB tree of TWSNMP using `get_mib_tree` (prefer `search_mib` for keyword search).
- **Arguments:**
  - `oid` (Optional): Subtree root OID.
  - `name` (Optional): Subtree object name.
  - `depth` (Optional): Maximum traversal depth.

### 14. `get_ip_address_list`

- **Title:** Get the list of IP addresses managed by TWSNMP.
- **Description:** Get the list of IP addresses managed by TWSNMP using `get_ip_address_list`.
- **Arguments:**
  - `ip_filter` (Optional): IP address regex filter.
  - `mac_filter` (Optional): MAC address regex filter.
  - `vendor_filter` (Optional): Vendor name regex filter.
  - `limit` (Optional): Max entries to return.

### 15. `get_resource_monitor_list`

- **Title:** Get resource monitor info of TWSNMP
- **Description:** Get resource monitor information of TWSNMP using `get_resource_monitor_list`.
- **Arguments:** None.

### 16. `get_server_certificate_list`

- **Title:** Get the list of server certificates managed by TWSNMP
- **Description:** Get the list of server certificates managed by TWSNMP using `get_server_certificate_list`.
- **Arguments:** None.

### 17. `add_event_log`

- **Title:** Add Event log to TWSNMP
- **Description:** Add an event log to TWSNMP.
- **Arguments:**
  - `level` (Required): Event log level (`info`, `normal`, `warn`, `low`, `high`).
  - `node` (Optional): Node name or ID for the event log.
  - `event` (Required): Event log content.

### 18. `get_ip_address_info`

- **Title:** Get IP address information
- **Description:** Get information about an IP address.
- **Arguments:**
  - `ip` (Required): Target IP address.

### 19. `get_mac_address_info`

- **Title:** Get MAC address information
- **Description:** Get information about a MAC address.
- **Arguments:**
  - `mac` (Required): Target MAC address.

---

## Workflow & Extended Prompts (20–59)

### 20. `troubleshoot_network`

- **Title:** Investigate network or node issues
- **Description:** Guided multi-step troubleshooting workflow using TWSNMP tools (`get_system_status`, abnormal nodes, failing pollings, lines, event logs, syslogs, and `diagnose_node`).
- **Arguments:**
  - `target` (Optional): Node ID, name, or IP to focus on.
  - `start_time` (Optional): Lookback window (default `-1h`).

### 21. `diagnose_node`

- **Title:** Run live diagnostics on a node or target
- **Description:** Execute multi-layer Ping, SNMP, TCP port, and HTTP diagnostics via `diagnose_node`.
- **Arguments:**
  - `target` (Required): Node ID, name, IP, or hostname.
  - `ports` (Optional): Comma-separated TCP ports.
  - `timeout` (Optional): Timeout in seconds (`1`–`10`).

### 22. `get_system_status`

- **Title:** Get TWSNMP system status
- **Description:** Inspect TWSNMP runtime, receivers, DB size, and node/polling summaries.
- **Arguments:** None.

### 23. `get_node`

- **Title:** Get single node details
- **Description:** Retrieve full node details and pollings via `get_node`.
- **Arguments:**
  - `id` (Required): Node ID, unique name, or IP.

### 24. `delete_node`

- **Title:** Delete a node
- **Description:** Verify and delete a node, its pollings, and its lines via `delete_node`.
- **Arguments:**
  - `id` (Required): Node ID, unique name, or IP.

### 25. `get_polling`

- **Title:** Get polling configuration
- **Description:** Get full polling configuration and result via `get_polling`.
- **Arguments:**
  - `id` (Required): Polling ID.

### 26. `add_polling`

- **Title:** Add a polling check
- **Description:** Create a new polling on a node via `add_polling`.
- **Arguments:**
  - `node_id` (Required): Target node ID, name, or IP.
  - `name` (Required): Polling name.
  - `type` (Required): Polling type (`ping`, `tcp`, `http`, `dns`, `snmp`, etc.).
  - `params` (Optional): Parameter (port, URL, OID).
  - `level` (Optional): Severity level (`off`, `low`, `warn`, `high`).

### 27. `update_polling`

- **Title:** Update a polling configuration
- **Description:** Modify an existing polling via `update_polling`.
- **Arguments:**
  - `id` (Required): Polling ID.
  - `level` (Optional): Severity level (`off`, `low`, `warn`, `high`).
  - `poll_int` (Optional): Interval in seconds.
  - `params` (Optional): Parameter override.

### 28. `delete_polling`

- **Title:** Delete a polling
- **Description:** Delete a polling check via `delete_polling`.
- **Arguments:**
  - `id` (Required): Polling ID.

### 29. `check_polling`

- **Title:** Run polling checks immediately
- **Description:** Trigger immediate execution of a polling, node pollings, or all pollings via `check_polling`.
- **Arguments:**
  - `polling_id` (Optional): Specific polling ID.
  - `node_id` (Optional): Check all pollings for a node.
  - `all` (Optional): Set `true` to check all pollings.

### 30. `get_polling_log_data`

- **Title:** Export polling log data as CSV
- **Description:** Fetch polling log metrics in CSV format via `get_polling_log_data`.
- **Arguments:**
  - `id` (Required): Polling ID.
  - `limit` (Optional): Max rows.
  - `start_time` (Optional), `end_time` (Optional).

### 31. `get_line_list`

- **Title:** Get map line list
- **Description:** List map lines between nodes and networks via `get_line_list`.
- **Arguments:**
  - `node_id` (Optional): Node or network filter.
  - `state_filter` (Optional): State regex.

### 32. `add_line`

- **Title:** Add map line
- **Description:** Connect two nodes or networks with a line via `add_line`.
- **Arguments:**
  - `node_id1` (Required), `node_id2` (Required).
  - `polling_id1` (Optional), `polling_id2` (Optional).

### 33. `delete_line`

- **Title:** Delete map line
- **Description:** Delete a map line by ID via `delete_line`.
- **Arguments:**
  - `id` (Required): Line ID.

### 34. `get_network`

- **Title:** Get network container details
- **Description:** Get network container details and port table via `get_network`.
- **Arguments:**
  - `id` (Required): Network ID, name, or IP.

### 35. `add_network`

- **Title:** Add network container
- **Description:** Add a network/switch container via `add_network`.
- **Arguments:**
  - `name` (Required), `ip` (Required), `unmanaged` (Optional).

### 36. `update_network`

- **Title:** Update network container
- **Description:** Update network container fields via `update_network`.
- **Arguments:**
  - `id` (Required), `name` (Optional), `ip` (Optional).

### 37. `delete_network`

- **Title:** Delete network container
- **Description:** Delete a network container and connected lines via `delete_network`.
- **Arguments:**
  - `id` (Required): Network ID, name, or IP.

### 38. `check_network`

- **Title:** Check network status
- **Description:** Trigger SNMP/LLDP check for one or all network containers via `check_network`.
- **Arguments:**
  - `id` (Optional), `all` (Optional).

### 39. `update_server_certificate`

- **Title:** Add or update TLS server certificate monitor
- **Description:** Add or update a TLS certificate monitor entry via `update_server_certificate`.
- **Arguments:**
  - `id` (Optional), `target` (Optional), `port` (Optional).

### 40. `delete_server_certificate`

- **Title:** Delete TLS server certificate monitor
- **Description:** Remove a server certificate monitor entry via `delete_server_certificate`.
- **Arguments:**
  - `id` (Required): Entry ID.

### 41. `snmpset`

- **Title:** Set SNMP MIB object value
- **Description:** Modify a writable SNMP MIB object via `snmpset`.
- **Arguments:**
  - `target` (Required), `mib_object_name` (Required), `type` (Required: `integer`/`string`), `value` (Required).

### 42. `search_netflow`

- **Title:** Search NetFlow records
- **Description:** Query NetFlow traffic flows via `search_netflow`.
- **Arguments:**
  - `src_filter` (Optional), `dst_filter` (Optional), `start_time` (Optional), `limit` (Optional).

### 43. `search_sflow`

- **Title:** Search sFlow records
- **Description:** Query sFlow flow samples via `search_sflow`.
- **Arguments:**
  - `src_filter` (Optional), `dst_filter` (Optional), `start_time` (Optional), `limit` (Optional).

### 44. `search_sflow_counter`

- **Title:** Search sFlow counter samples
- **Description:** Query sFlow interface counter samples via `search_sflow_counter`.
- **Arguments:**
  - `remote_filter` (Optional), `type_filter` (Optional), `start_time` (Optional).

### 45. `get_otel_metrics`

- **Title:** Get OpenTelemetry metrics
- **Description:** Query OpenTelemetry metrics via `get_otel_metrics`.
- **Arguments:**
  - `host_filter` (Optional), `service_filter` (Optional), `name_filter` (Optional).

### 46. `get_otel_traces`

- **Title:** Get OpenTelemetry traces
- **Description:** List recent or failing OpenTelemetry traces via `get_otel_traces`.
- **Arguments:**
  - `service_filter` (Optional), `error_only` (Optional).

### 47. `get_otel_trace_detail`

- **Title:** Get OpenTelemetry trace spans
- **Description:** Inspect all spans of an OTel trace via `get_otel_trace_detail`.
- **Arguments:**
  - `trace_id` (Required), `bucket` (Optional).

### 48. `get_ai_anomaly_list`

- **Title:** Get AI anomaly detection scores
- **Description:** List latest AI anomaly scores across pollings via `get_ai_anomaly_list`.
- **Arguments:**
  - `node_id` (Optional), `min_score` (Optional).

### 49. `search_mib`

- **Title:** Search MIB definitions
- **Description:** Search loaded SNMP MIB objects via `search_mib`.
- **Arguments:**
  - `query` (Required), `limit` (Optional).

### 50. `get_sensor_list`

- **Title:** Get sensor health list
- **Description:** List registered telemetry sensors via `get_sensor_list`.
- **Arguments:**
  - `host_filter` (Optional), `type_filter` (Optional).

### 51. `get_report_summary`

- **Title:** Get log analytics report summary
- **Description:** Query aggregated log reports via `get_report_summary`.
- **Arguments:**
  - `report_type` (Required), `filter` (Optional).

### 52. `get_mac_address_list`

- **Title:** Get MAC address list
- **Description:** Get aggregated MAC addresses via `get_mac_address_list`.
- **Arguments:**
  - `mac_filter` (Optional), `vendor_filter` (Optional).

### 53. `get_wifi_ap_list`

- **Title:** Get detected Wi-Fi AP list
- **Description:** List detected Wi-Fi access points via `get_wifi_ap_list`.
- **Arguments:**
  - `ssid_filter` (Optional), `bssid_filter` (Optional).

### 54. `get_bluetooth_device_list`

- **Title:** Get detected Bluetooth device list
- **Description:** List detected Bluetooth/BLE devices via `get_bluetooth_device_list`.
- **Arguments:**
  - `name_filter` (Optional), `public_address_only` (Optional).

### 55. `get_fdb_table`

- **Title:** Get SNMP switch FDB table
- **Description:** Fetch live switch FDB entries via `get_fdb_table`.
- **Arguments:**
  - `id` (Required): Network or switch reference.

### 56. `auto_connect_lines`

- **Title:** Auto-connect map lines
- **Description:** Infer and add map topology lines via `auto_connect_lines`.
- **Arguments:**
  - `mode` (Optional): `strict` (default) or `speculative`.

### 57. `auto_layout`

- **Title:** Auto-arrange map layout
- **Description:** Arrange node coordinates or undo previous layout via `auto_layout`.
- **Arguments:**
  - `mode` (Optional): `hierarchical`, `cluster`, or `categorized`.
  - `undo` (Optional): `true` to revert last layout.

### 58. `gnmi_capabilities`

- **Title:** Get gNMI capabilities
- **Description:** Query gNMI capabilities and models via `gnmi_capabilities`.
- **Arguments:**
  - `node_id` (Required), `target` (Optional).

### 59. `gnmi_get`

- **Title:** Query gNMI telemetry path
- **Description:** Execute gNMI Get on a telemetry path via `gnmi_get`.
- **Arguments:**
  - `node_id` (Required), `path` (Required), `encoding` (Optional).
