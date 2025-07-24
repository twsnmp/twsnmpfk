# MCP Server Tool Specifications

This document describes the list and specifications of tools provided by the TWSNMP-FK MCP (Mark3Labs Control Plane) server, as defined in `mcp.go`.

## Common Items

- All tools return their results as a JSON formatted string.
- Filter parameters (`..._filter`) can be specified using regular expressions. If an empty string is provided, no filtering is performed.
- Parameters for specifying a time range (`start_time`, `end_time`) can accept absolute times like `2025/07/25 10:00:00` or relative durations from the present, such as `-1h` or `30m`.

## Tool List

### 1. `get_node_list`

Retrieves a list of nodes registered in TWSNMP.

- **Description**: Gets the node list from TWSNMP.
- **Parameters**:
    - `name_filter` (string): Regular expression to filter by node name.
    - `ip_filter` (string): Regular expression to filter by IP address.
    - `state_filter` (string): Regular expression to filter by state (`normal`, `warn`, `low`, `high`, `repair`, `unknown`).
- **Output**: Array of node information (JSON)
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "IP": "...",
        "MAC": "...",
        "State": "normal",
        "X": 100,
        "Y": 200,
        "Icon": "desktop",
        "Descrption": "..."
      }
    ]
    ```

### 2. `get_network_list`

Retrieves a list of networks registered in TWSNMP.

- **Description**: Gets the network list from TWSNMP.
- **Parameters**:
    - `name_filter` (string): Regular expression to filter by network name.
    - `ip_filter` (string): Regular expression to filter by IP address.
- **Output**: Array of network information (JSON)
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "IP": "...",
        "Ports": ["port1=up", "port2=down"],
        "X": 100,
        "Y": 200,
        "Descrption": "..."
      }
    ]
    ```

### 3. `get_polling_list`

Retrieves a list of pollings registered in TWSNMP.

- **Description**: Gets the polling list from TWSNMP.
- **Parameters**:
    - `type_filter` (string): Regular expression to filter by polling type (e.g., `ping`, `tcp`, `http`).
    - `state_filter` (string): Regular expression to filter by state (e.g., `normal`, `warn`).
    - `name_filter` (string): Regular expression to filter by polling name.
    - `node_name_filter` (string): Regular expression to filter by the node name of the polling target.
- **Output**: Array of polling information (JSON)
    ```json
    [
      {
        "ID": "...",
        "Name": "...",
        "NodeID": "...",
        "NodeName": "...",
        "Type": "ping",
        "Level": "normal",
        "State": "normal",
        "LastTime": "2025-07-25T10:00:00Z",
        "Result": { "...": "..." }
      }
    ]
    ```

### 4. `do_ping`

Executes a Ping to the specified target.

- **Description**: Executes a ping.
- **Parameters**:
    - `target` (string, **Required**): Ping target (IP address or hostname).
    - `size` (number): Packet size (default: 64, min: 64, max: 1500).
    - `ttl` (number): TTL (default: 254, min: 1, max: 254).
    - `timeout` (number): Timeout (seconds) (default: 2, min: 1, max: 10).
- **Output**: Ping execution result (JSON)
    ```json
    {
      "Result": "Success",
      "Time": "2025-07-25T10:00:00Z",
      "RTT": "1.234ms",
      "RTTNano": 1234000,
      "Size": 64,
      "TTL": 64,
      "ResponceFrom": "192.168.1.1",
      "Location": "..."
    }
    ```

### 5. `get_MIB_tree`

Retrieves the MIB tree information held by TWSNMP.

- **Description**: Gets the MIB tree from TWSNMP.
- **Parameters**: None
- **Output**: MIB tree information (JSON)

### 6. `snmpwalk`

Executes an SNMP Walk on the specified target.

- **Description**: SNMP Walk tool.
- **Parameters**:
    - `target` (string, **Required**): SNMP Walk target (IP, hostname, or node name).
    - `mib_object_name` (string, **Required**): MIB object name.
    - `community` (string): SNMPv2c community name.
    - `user` (string): SNMPv3 username.
    - `password` (string): SNMPv3 password.
    - `snmpmode` (string): SNMP mode (`v2c`, `v3auth`, `v3authpriv`, `v3authprivex`).
- **Output**: Array of SNMP Walk results (JSON)
    ```json
    [
      {
        "Name": "sysDescr.0",
        "Value": "..."
      }
    ]
    ```

### 7. `add_node`

Adds a new node to TWSNMP.

- **Description**: Adds a node to TWSNMP.
- **Parameters**:
    - `name` (string, **Required**): Node name.
    - `ip` (string, **Required**): IP address.
    - `icon` (string): Icon name (default: `desktop`).
    - `description` (string): Description.
    - `x` (number): X coordinate (min: 64, max: 1000).
    - `y` (number): Y coordinate (min: 64, max: 1000).
- **Output**: Added node information (JSON)

### 8. `update_node`

Updates an existing node's information.

- **Description**: Updates a node's name, IP, position, description, or icon.
- **Parameters**:
    - `id` (string, **Required**): ID of the node to update.
    - `name` (string): New node name.
    - `ip` (string): New IP address.
    - `icon` (string): New icon name.
    - `description` (string): New description.
    - `x` (number): New X coordinate.
    - `y` (number): New Y coordinate.
- **Output**: Updated node information (JSON)

### 9. `get_ip_address_list`

Retrieves a list of IP addresses collected by TWSNMP from ARP logs, etc.

- **Description**: Gets the IP address list from TWSNMP.
- **Parameters**: None
- **Output**: Array of IP address information (JSON)
    ```json
    [
      {
        "IP": "...",
        "MAC": "...",
        "Node": "...",
        "Vendor": "...",
        "FirstTime": "2025-07-25T10:00:00Z",
        "LastTime": "2025-07-25T11:00:00Z"
      }
    ]
    ```

### 10. `get_resource_monitor_list`

Retrieves a list of the resource usage of the TWSNMP server itself.

- **Description**: Gets the resource monitor list from TWSNMP.
- **Parameters**: None
- **Output**: Array of resource information (JSON)
    ```json
    [
      {
        "Time": "2025-07-25T10:00:00Z",
        "CPUUsage": "10.50%",
        "MemoryUsage": "25.20%",
        "SwapUsage": "0.00%",
        "DiskUsage": "45.80%",
        "Load": "0.75"
      }
    ]
    ```

### 11. `search_event_log`

Searches the event logs.

- **Description**: Searches event logs from TWSNMP.
- **Parameters**:
    - `node_filter` (string): Filter by node name.
    - `type_filter` (string): Filter by type.
    - `level_filter` (string): Filter by level (e.g., `warn`, `info`).
    - `event_filter` (string): Filter by event content.
    - `limit_log_count` (number): Maximum number of logs to retrieve (default: 100, max: 10000).
    - `start_time` (string): Search start time (default: `-1h`).
    - `end_time` (string): Search end time (default: current time).
- **Output**: Array of event logs (JSON)
    ```json
    [
      {
        "Time": "2025-07-25T10:00:00Z",
        "Type": "...",
        "Level": "warn",
        "Node": "...",
        "Event": "..."
      }
    ]
    ```

### 12. `search_syslog`

Searches Syslog.

- **Description**: Searches syslog from TWSNMP.
- **Parameters**:
    - `host_filter` (string): Filter by hostname.
    - `tag_filter` (string): Filter by tag.
    - `level_filter` (string): Filter by level.
    - `message_filter` (string): Filter by message content.
    - `limit_log_count` (number): Maximum number of logs to retrieve (default: 100, max: 10000).
    - `start_time` (string): Search start time (default: `-1h`).
    - `end_time` (string): Search end time (default: current time).
- **Output**: Array of Syslog entries (JSON)
    ```json
    [
      {
        "Time": "...",
        "Level": "info",
        "Host": "...",
        "Type": "...",
        "Tag": "...",
        "Message": "...",
        "Severity": 6,
        "Facility": 1
      }
    ]
    ```

### 13. `get_syslog_summary`

Aggregates Syslog by pattern and retrieves a summary.

- **Description**: Gets a syslog summary from TWSNMP.
- **Parameters**:
    - `host_filter`, `tag_filter`, `level_filter`, `message_filter`: Same as `search_syslog`.
    - `top_n` (number): How many top patterns to display (default: 10, max: 100).
    - `start_time`, `end_time`: Same as `search_syslog`.
- **Output**: Array of Syslog summary (JSON)
    ```json
    [
      {
        "Pattern": "...",
        "Count": 123
      }
    ]
    ```

### 14. `search_snmp_trap_log`

Searches SNMP Trap logs.

- **Description**: Searches SNMP trap logs from TWSNMP.
- **Parameters**:
    - `from_filter` (string): Filter by sender address.
    - `trap_type_filter` (string): Filter by trap type.
    - `variable_filter` (string): Filter by the content of Variables.
    - `limit_log_count`, `start_time`, `end_time`: Same as `search_event_log`.
- **Output**: Array of SNMP Trap logs (JSON)
    ```json
    [
      {
        "Time": "...",
        "FromAddress": "...",
        "TrapType": "linkDown",
        "Variables": "..."
      }
    ]
    ```

### 15. `get_server_certificate_list`

Retrieves a list of monitored server certificates.

- **Description**: Gets the server certificate list from TWSNMP.
- **Parameters**: None
- **Output**: Array of server certificate information (JSON)
    ```json
    [
      {
        "State": "valid",
        "Server": "example.com",
        "Port": 443,
        "Subject": "CN=example.com,...",
        "Issuer": "...",
        "SerialNumber": "...",
        "Verify": true,
        "NotAfter": "2026-07-25T10:00:00Z",
        "NotBefore": "2024-07-25T10:00:00Z",
        "Error": "",
        "FirstTime": "...",
        "LastTime": "..."
      }
    ]
    ```
