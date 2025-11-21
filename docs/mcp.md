# TWSNMP FK MCP Server Specification

This document outlines the specifications for the TWSNMP FK MCP (Model Context Protocol) Server, based on the source code in `backend/mcp.go`, `backend/mcp_tools.go`, and `backend/mcp_prompts.go`.

## 1. Overview

The MCP server provides an interface for AI agents to interact with the TWSNMP FK monitoring system. It allows agents to retrieve monitoring data, perform actions, and access system information through a set of defined tools and prompts.

## 2. Transport and Endpoints

The server starts based on the configuration `datastore.MapConf.MCPTransport`. It supports two transport mechanisms:

- **Server-Sent Events (SSE)**:
  - Enabled when `MCPTransport` is set to `"sse"`.
  - Endpoints: `/sse` and `/message`.
- **Streamable HTTP**:
  - Enabled for other transport settings (e.g., `"http"`).
  - Endpoint: `/mcp`.

The server listens on the address specified in `datastore.MapConf.MCPEndpoint`.

### TLS Security

TLS is automatically enabled if a server certificate (`datastore.MCPCert`) and private key (`datastore.MCPKey`) are provided. It uses TLS 1.3 with a restricted set of secure cipher suites.

## 3. Authentication

Access to the MCP server is controlled by two mechanisms, both of which must be satisfied if configured:

- **Token Authentication**: If `datastore.MapConf.MCPToken` is set, the `Authorization` header of the incoming request must contain this token.
- **IP Address ACL**: If `datastore.MapConf.MCPFrom` is set (as a comma-separated list of IP addresses), the request's source IP address must be in the allow list.

## 4. Tools (Functions)

The server exposes the following tools for agents.

---

### `get_node_list`
- **Description**: Get a list of nodes from TWSNMP.
- **Parameters**:
  - `name_filter` (string, optional): Regex to filter by node name.
  - `ip_filter` (string, optional): Regex to filter by node IP address.
  - `state_filter` (string, optional): Regex to filter by node state (`normal`, `warn`, `low`, `high`, `repair`, `unknown`).
- **Returns**: A JSON array of node objects.

---

### `get_network_list`
- **Description**: Get a list of networks from TWSNMP.
- **Parameters**:
  - `name_filter` (string, optional): Regex to filter by network name.
  - `ip_filter` (string, optional): Regex to filter by network IP address.
- **Returns**: A JSON array of network objects.

---

### `get_polling_list`
- **Description**: Get a list of pollings from TWSNMP.
- **Parameters**:
  - `type_filter` (string, optional): Regex to filter by polling type (`ping`, `tcp`, `http`, `dns`, `twsnmp`, `syslog`, etc.).
  - `name_filter` (string, optional): Regex to filter by polling name.
  - `node_name_filter` (string, optional): Regex to filter by the node name associated with the polling.
  - `state_filter` (string, optional): Regex to filter by polling state (`normal`, `warn`, `low`, `high`, `repair`, `unknown`).
- **Returns**: A JSON array of polling objects.

---

### `get_polling_log`
- **Description**: Get the polling log for a specific polling ID.
- **Parameters**:
  - `id` (string, required): The ID of the polling.
  - `limit` (int, optional): The maximum number of log entries to retrieve (1-2000, default 100).
- **Returns**: A JSON array of polling log entries.

---

### `get_polling_log_data`
- **Description**: Get the polling log data for a specific polling ID.
- **Parameters**:
  - `id` (string, required): The ID of the polling.
  - `limit` (int, optional): The maximum number of log entries to retrieve (1-2000, default 100).
- **Returns**: A CSV of polling log  data entries.

---

### `do_ping`
- **Description**: Perform a ping to a target.
- **Parameters**:
  - `target` (string, required): The target IP address or hostname.
  - `size` (int, optional): Packet size (1-1500, default 64).
  - `ttl` (int, optional): IP packet TTL (1-255, default 254).
  - `timeout` (int, optional): Timeout in seconds (1-10, default 3).
- **Returns**: A JSON object with the ping result.

---

### `get_mib_tree`
- **Description**: Get the MIB tree from TWSNMP.
- **Parameters**: None.
- **Returns**: A JSON object representing the MIB tree structure.

---

### `snmpwalk`
- **Description**: Perform an SNMP walk.
- **Parameters**:
  - `target` (string, required): The target IP address or node name.
  - `mib_object_name` (string, required): The MIB object name or OID to start the walk from.
  - `community` (string, optional): Community string for SNMPv2c. If not provided, the node's configured community is used.
  - `user` (string, optional): Username for SNMPv3.
  - `password` (string, optional): Password for SNMPv3.
  - `snmp_mode` (string, optional): SNMP mode (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`).
- **Returns**: A JSON array of MIB objects with their names and values.

---

### `add_node`
- **Description**: Add a new node to TWSNMP.
- **Parameters**:
  - `name` (string, required): Node name.
  - `ip` (string, required): Node IP address.
  - `icon` (string, optional): Icon for the node (default: `desktop`).
  - `description` (string, optional): Description of the node.
  - `x` (int, optional): X position on the map (1-1000, default 64).
  - `y` (int, optional): Y position on the map (1-1000, default 64).
- **Returns**: A JSON object of the newly created node. A `PING` polling is automatically added.

---

### `update_node`
- **Description**: Update a node's properties (name, IP, position, description, or icon).
- **Parameters**:
  - `id` (string, required): The node ID, current name, or current IP address.
  - `name` (string, optional): New node name.
  - `ip` (string, optional): New IP address.
  - `icon` (string, optional): New icon.
  - `description` (string, optional): New description.
  - `x` (int, optional): New X position.
  - `y` (int, optional): New Y position.
- **Returns**: A JSON object of the updated node.

---

### `get_ip_address_list`
- **Description**: Get a list of IP addresses discovered via ARP.
- **Parameters**: None.
- **Returns**: A JSON array of objects, each containing IP, MAC, associated node, vendor, and timestamps.

---

### `get_resource_monitor_list`
- **Description**: Get a list of TWSNMP's own resource monitoring data (CPU, memory, etc.).
- **Parameters**: None.
- **Returns**: A JSON array of resource usage snapshots.

---

### `search_event_log`
- **Description**: Search the event log.
- **Parameters**:
  - `node_filter` (string, optional): Regex to filter by node name.
  - `type_filter` (string, optional): Regex to filter by event type.
  - `level_filter` (string, optional): Regex to filter by event level (`info`, `warn`, `low`, `high`, `debug`).
  - `event_filter` (string, optional): Regex to filter by event message content.
  - `start_time` (string, optional): Start time for the search (e.g., "2023-10-27 00:00:00" or "-1h"). Default is "-1h".
  - `end_time` (string, optional): End time for the search (e.g., "2023-10-27 23:59:59" or "now"). Default is "now".
  - `limit_log_count` (int, optional): Max number of logs to return (100-10000, default 100).
- **Returns**: A JSON array of event log entries.

---

### `search_syslog`
- **Description**: Search the syslog.
- **Parameters**:
  - `level_filter` (string, optional): Regex to filter by level (`info`, `warn`, `low`, `high`, `debug`).
  - `host_filter` (string, optional): Regex to filter by hostname.
  - `tag_filter` (string, optional): Regex to filter by syslog tag.
  - `message_filter` (string, optional): Regex to filter by message content.
  - `start_time` (string, optional): Start time (default: "-1h").
  - `end_time` (string, optional): End time (default: "now").
  - `limit_log_count` (int, optional): Max number of logs to return (100-10000, default 100).
- **Returns**: A JSON array of syslog entries.

---

### `get_syslog_summary`
- **Description**: Get a summary of syslog patterns.
- **Parameters**:
  - `level_filter`, `host_filter`, `tag_filter`, `message_filter`, `start_time`, `end_time`: Same as `search_syslog`.
  - `top_n` (int, optional): The number of top patterns to return (5-500, default 5).
- **Returns**: A JSON array of syslog patterns and their counts.

---

### `search_snmp_trap_log`
- **Description**: Search the SNMP trap log.
- **Parameters**:
  - `from_filter` (string, optional): Regex to filter by sender address.
  - `trap_type_filter` (string, optional): Regex to filter by trap type.
  - `variable_filter` (string, optional): Regex to filter by trap variable content.
  - `start_time` (string, optional): Start time (default: "-1h").
  - `end_time` (string, optional): End time (default: "now").
  - `limit` (int, optional): Max number of logs to return (100-10000, default 100).
- **Returns**: A JSON array of SNMP trap log entries.

---

### `get_server_certificate_list`
- **Description**: Get a list of monitored server certificates.
- **Parameters**: None.
- **Returns**: A JSON array of certificate monitoring entries.

---

### `add_event_log`
- **Description**: Add an event log to TWSNMP.
- **Parameters**:
  - `level` (string, optional): Event level (`info`, `normal`, `warn`, `low`, `high`). Default is `info`.
  - `node` (string, optional): The name of the node associated with the event.
  - `event` (string, required): The content of the event log.
- **Returns**: A string "ok" on success.

---

### `get_ip_address_info`
- **Description**: Get information about an IP address (DNS, managed node, Geo location, RDAP).
- **Parameters**:
  - `ip` (string, required): The IP address to query.
- **Returns**: A JSON object containing aggregated information about the IP address.

---

### `get_mac_address_info`
- **Description**: Get information about a MAC address (IP, managed node, vendor).
- **Parameters**:
  - `mac` (string, required): The MAC address to query.
- **Returns**: A JSON object containing aggregated information about the MAC address.


