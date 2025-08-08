# MCP Server Tools

This document outlines the specifications of the tools available on the TWSNMP FK MCP Server.

## Tools

### 1. `get_node_list`

**Description:** Retrieves a list of nodes from TWSNMP.

**Parameters:**

- `name_filter` (string, optional): A regular expression to filter nodes by name. If omitted, all nodes are returned.
- `ip_filter` (string, optional): A regular expression to filter nodes by IP address. If omitted, all nodes are returned.
- `state_filter` (string, optional): A regular expression to filter nodes by state. If omitted, all nodes are returned. Valid states are "normal", "warn", "low", "high", "repair", and "unknown".

**Output:** A JSON array of node objects with the following properties:
- `ID` (string): The node ID.
- `Name` (string): The node name.
- `IP` (string): The node's IP address.
- `MAC` (string): The node's MAC address.
- `State` (string): The node's state.
- `X` (int): The X coordinate of the node on the map.
- `Y` (int): The Y coordinate of the node on the map.
- `Icon` (string): The node's icon.
- `Descrption` (string): The node's description.

### 2. `get_network_list`

**Description:** Retrieves a list of networks from TWSNMP.

**Parameters:**

- `name_filter` (string, optional): A regular expression to filter networks by name. If omitted, all networks are returned.
- `ip_filter` (string, optional): A regular expression to filter networks by IP address. If omitted, all networks are returned.

**Output:** A JSON array of network objects with the following properties:
- `ID` (string): The network ID.
- `Name` (string): The network name.
- `IP` (string): The network's IP address.
- `Ports` (array of strings): A list of the network's ports and their states.
- `X` (int): The X coordinate of the network on the map.
- `Y` (int): The Y coordinate of the network on the map.
- `Descrption` (string): The network's description.

### 3. `get_polling_list`

**Description:** Retrieves a list of pollings from TWSNMP.

**Parameters:**

- `type_filter` (string, optional): A regular expression to filter pollings by type. Valid types are "ping", "tcp", "http", "dns", "twsnmp", and "syslog".
- `state_filter` (string, optional): A regular expression to filter pollings by state. Valid states are "normal", "warn", "low", "high", "repair", and "unknown".
- `name_filter` (string, optional): A regular expression to filter pollings by name.
- `node_name_filter` (string, optional): A regular expression to filter pollings by node name.

**Output:** A JSON array of polling objects with the following properties:
- `ID` (string): The polling ID.
- `Name` (string): The polling name.
- `NodeID` (string): The ID of the node being polled.
- `NodeName` (string): The name of the node being polled.
- `Type` (string): The polling type.
- `Level` (string): The polling level.
- `State` (string): The polling state.
- `Logging` (boolean): Whether logging is enabled for the polling.
- `LastTime` (string): The last time the polling was executed.
- `Result` (object): The result of the last polling.

### 4. `get_polling_log`

**Description:** Retrieves the polling log for a specific polling.

**Parameters:**

- `id` (string, required): The ID of the polling.
- `limit` (number, optional, default: 100): The maximum number of log entries to retrieve (1-2000).

**Output:** A JSON array of polling log objects with the following properties:
- `Time` (string): The timestamp of the log entry.
- `State` (string): The state at the time of the log entry.
- `Result` (object): The result of the polling at the time of the log entry.

### 5. `do_ping`

**Description:** Executes a ping to a target.

**Parameters:**

- `target` (string, required): The IP address or hostname to ping.
- `size` (number, optional, default: 64): The packet size (64-1500).
- `ttl` (number, optional, default: 254): The TTL of the IP packet (1-254).
- `timeout` (number, optional, default: 2): The ping timeout in seconds (1-10).

**Output:** A JSON object with the following properties:
- `Result` (string): The ping result.
- `Time` (string): The timestamp of the ping.
- `RTT` (string): The round-trip time.
- `RTTNano` (int64): The round-trip time in nanoseconds.
- `Size` (int): The packet size.
- `TTL` (int): The TTL of the response packet.
- `ResponceFrom` (string): The IP address of the responder.
- `Location` (string): The location of the responder.

### 6. `get_MIB_tree`

**Description:** Retrieves the MIB tree from TWSNMP.

**Parameters:** None.

**Output:** A JSON object representing the MIB tree.

### 7. `snmpwalk`

**Description:** Performs an SNMP walk.

**Parameters:**

- `target` (string, required): The IP address, hostname, or node name to walk.
- `mib_object_name` (string, required): The MIB object name to walk.
- `community` (string, optional): The SNMPv2c community string.
- `user` (string, optional): The SNMPv3 username.
- `password` (string, optional): The SNMPv3 password.
- `snmpmode` (string, optional): The SNMP mode. Valid values are "v2c", "v3auth", "v3authpriv", and "v3authprivex".

**Output:** A JSON array of MIB objects with the following properties:
- `Name` (string): The MIB object name.
- `Value` (string): The MIB object value.

### 8. `add_node`

**Description:** Adds a new node to TWSNMP.

**Parameters:**

- `name` (string, required): The node name.
- `ip` (string, required): The node's IP address.
- `icon` (string, optional): The node's icon.
- `description` (string, optional): The node's description.
- `x` (number, optional): The X coordinate of the node on the map (64-1000).
- `y` (number, optional): The Y coordinate of the node on the map (64-1000).

**Output:** A JSON object representing the newly added node.

### 9. `update_node`

**Description:** Updates an existing node in TWSNMP.

**Parameters:**

- `id` (string, required): The ID of the node to update.
- `name` (string, optional): The new node name.
- `ip` (string, optional): The new IP address.
- `icon` (string, optional): The new icon.
- `description` (string, optional): The new description.
- `x` (number, optional): The new X coordinate.
- `y` (number, optional): The new Y coordinate.

**Output:** A JSON object representing the updated node.

### 10. `get_ip_address_list`

**Description:** Retrieves a list of IP addresses from TWSNMP.

**Parameters:** None.

**Output:** A JSON array of IP address objects with the following properties:
- `IP` (string): The IP address.
- `MAC` (string): The MAC address.
- `Node` (string): The associated node name.
- `Vendor` (string): The vendor of the network interface.
- `FirstTime` (string): The first time the IP address was seen.
- `LastTime` (string): The last time the IP address was seen.

### 11. `get_resource_monitor_list`

**Description:** Retrieves a list of resource monitor data from TWSNMP.

**Parameters:** None.

**Output:** A JSON array of resource monitor objects with the following properties:
- `Time` (string): The timestamp of the data.
- `CPUUsage` (string): The CPU usage.
- `MemoryUsage` (string): The memory usage.
- `SwapUsage` (string): The swap usage.
- `DiskUsage` (string): The disk usage.
- `Load` (string): The system load.

### 12. `search_event_log`

**Description:** Searches the event log in TWSNMP.

**Parameters:**

- `node_filter` (string, optional): A regular expression to filter by node name.
- `type_filter` (string, optional): A regular expression to filter by type.
- `level_filter` (string, optional): A regular expression to filter by level. Valid levels are "warn", "low", "high", "debug", and "info".
- `event_filter` (string, optional): A regular expression to filter by event.
- `limit_log_count` (number, optional, default: 100): The maximum number of log entries to retrieve (1-10000).
- `start_time` (string, optional, default: "-1h"): The start time of the search.
- `end_time` (string, optional, default: "now"): The end time of the search.

**Output:** A JSON array of event log objects with the following properties:
- `Time` (string): The timestamp of the event.
- `Type` (string): The event type.
- `Level` (string): The event level.
- `Node` (string): The associated node name.
- `Event` (string): The event description.

### 13. `search_syslog`

**Description:** Searches the syslog in TWSNMP.

**Parameters:**

- `host_filter` (string, optional): A regular expression to filter by hostname.
- `tag_filter` (string, optional): A regular expression to filter by tag.
- `level_filter` (string, optional): A regular expression to filter by level. Valid levels are "warn", "low", "high", "debug", and "info".
- `message_filter` (string, optional): A regular expression to filter by message.
- `limit_log_count` (number, optional, default: 100): The maximum number of log entries to retrieve (1-10000).
- `start_time` (string, optional, default: "-1h"): The start time of the search.
- `end_time` (string, optional, default: "now"): The end time of the search.

**Output:** A JSON array of syslog objects with the following properties:
- `Time` (string): The timestamp of the log entry.
- `Level` (string): The log level.
- `Host` (string): The hostname.
- `Type` (string): The log type.
- `Tag` (string): The log tag.
- `Message` (string): The log message.
- `Severity` (int): The severity level.
- `Facility` (int): The facility code.

### 14. `get_syslog_summary`

**Description:** Retrieves a summary of the syslog from TWSNMP.

**Parameters:**

- `host_filter` (string, optional): A regular expression to filter by hostname.
- `tag_filter` (string, optional): A regular expression to filter by tag.
- `level_filter` (string, optional): A regular expression to filter by level.
- `message_filter` (string, optional): A regular expression to filter by message.
- `top_n` (number, optional, default: 10): The number of top syslog patterns to retrieve (5-100).
- `start_time` (string, optional, default: "-1h"): The start time of the search.
- `end_time` (string, optional, default: "now"): The end time of the search.

**Output:** A JSON array of syslog summary objects with the following properties:
- `Pattern` (string): The log pattern.
- `Count` (int): The number of occurrences of the pattern.

### 15. `search_snmp_trap_log`

**Description:** Searches the SNMP trap log in TWSNMP.

**Parameters:**

- `from_filter` (string, optional): A regular expression to filter by sender address.
- `trap_type_filter` (string, optional): A regular expression to filter by trap type.
- `variable_filter` (string, optional): A regular expression to filter by trap variables.
- `limit_log_count` (number, optional, default: 100): The maximum number of log entries to retrieve (1-10000).
- `start_time` (string, optional, default: "-1h"): The start time of the search.
- `end_time` (string, optional, default: "now"): The end time of the search.

**Output:** A JSON array of SNMP trap log objects with the following properties:
- `Time` (string): The timestamp of the trap.
- `FromAddress` (string): The sender's address.
- `TrapType` (string): The trap type.
- `Variables` (string): The trap variables.

### 16. `get_server_certificate_list`

**Description:** Retrieves a list of server certificates from TWSNMP.

**Parameters:** None.

**Output:** A JSON array of server certificate objects with the following properties:
- `State` (string): The certificate state.
- `Server` (string): The server address.
- `Port` (uint16): The server port.
- `Subject` (string): The certificate subject.
- `Issuer` (string): The certificate issuer.
- `SerialNumber` (string): The certificate serial number.
- `Verify` (boolean): Whether the certificate is verified.
- `NotAfter` (string): The certificate's expiration date.
- `NotBefore` (string): The certificate's issuance date.
- `Error` (string): Any error associated with the certificate.
- `FirstTime` (string): The first time the certificate was seen.
- `LastTime` (string): The last time the certificate was seen.

### 17. `add_event_log`

**Description:** Adds an event log to TWSNMP.

**Parameters:**

- `level` (string, optional, default: "info"): The event level ("info", "normal", "warn", "low", "high").
- `node` (string, optional): The name of the node associated with the event.
- `event` (string, optional): The event log content.

**Output:** A string indicating the result of the operation ("ok").
