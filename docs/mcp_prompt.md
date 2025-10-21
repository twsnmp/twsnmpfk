# MCP Server Prompt Specifications

This document outlines the specifications for the prompts available on the MCP server, as defined in `backend/mcp_prompts.go`.

## Prompts

### 1. `get_node_list`

- **Title:** Get node list with filters
- **Description:** Get a list of nodes registered in TWSNMP with filters.
- **Arguments:**
  - `state_filter` (Optional): Node state filter. Valid states are `normal`, `repair`, `warn`, `low`, `high`, `unknown`.
  - `name_filter` (Optional): Node name filter.
  - `ip_filter` (Optional): Node IP address filter.

### 2. `add_node`

- **Title:** Add a new node to TWSNMP.
- **Description:** Add a new node to TWSNMP.
- **Arguments:**
  - `name` (Required): Name of the new node.
  - `ip` (Required): IP address of the new node.
  - `icon` (Optional): Icon for the node. Valid icons are `desktop`, `laptop`, `server`, `cloud`, `router`, `ip`.
  - `description` (Optional): Description of the new node.
  - `position` (Optional): Position of the new node (e.g., `x=100,y=200`).

### 3. `update_node`

- **Title:** Update node on TWSNMP.
- **Description:** Update an existing node on TWSNMP.
- **Arguments:**
  - `id` (Required): The ID, name, or IP address of the node to update.
  - `name` (Optional): New name for the node.
  - `ip` (Optional): New IP address for the node.
  - `icon` (Optional): New icon for the node. Valid icons are `desktop`, `laptop`, `server`, `cloud`, `router`, `ip`.
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
  - `type_filter` (Optional): Polling type filter. Valid types are `ping`, `snmp`, `syslog`, `http`, `tcp`.
  - `name_filter` (Optional): Polling name filter.
  - `node_name_filter` (Optional): Node name filter.
  - `state_filter` (Optional): Polling state filter. Valid states are `normal`, `repair`, `warn`, `low`, `high`, `unknown`.

### 6. `get_polling_log`

- **Title:** Get polling log.
- **Description:** Get polling log from TWSNMP.
- **Arguments:**
  - `id` (Required): The ID of the polling.
  - `limit` (Optional): Maximum number of logs to retrieve (default is 100).

### 7. `do_ping`

- **Title:** Do ping
- **Description:** Perform a ping to a target from TWSNMP.
- **Arguments:**
  - `target` (Required): The ping target (IP, node name, or hostname).
  - `size` (Optional): Ping packet size (default is 64).
  - `ttl` (Optional): Ping packet TTL (default is 254).
  - `timeout` (Optional): Ping timeout in seconds (default is 3).

### 8. `snmpwalk`

- **Title:** Do snmpwalk
- **Description:** Perform an snmpwalk to a target from TWSNMP.
- **Arguments:**
  - `target` (Required): The SNMP walk target (IP, node name, or hostname).
  - `mib_object_name` (Required): The MIB object name to get.
  - `snmp_mode` (Optional): SNMP mode. Valid modes are `v2c`, `v3auth`, `v3authpriv`, `v3authprivex`.
  - `community` (Optional): Community name for v2c mode.
  - `user` (Optional): User name for v3 mode.
  - `password` (Optional): Password for v3 mode.

### 9. `search_event_log`

- **Title:** Search event log with filters
- **Description:** Search event logs with filters.
- **Arguments:**
  - `type_filter` (Optional): Event type filter. Valid types are `system`, `polling`, `arpwatch`, `mcp`.
  - `node_filter` (Optional): Node filter.
  - `level_name_filter` (Optional): Level filter. Valid levels are `info`, `normal`, `warn`, `low`, `high`.
  - `event_filter` (Optional): Event filter.
  - `start_time` (Optional): Start time for the search (default is -1h).
  - `end_time` (Optional): End date and time for the search (default is now).
  - `limit` (Optional): Maximum number of logs to search.

### 10. `search_syslog`

- **Title:** Search syslog with filters
- **Description:** Search syslogs with filters.
- **Arguments:**
  - `level_filter` (Optional): Level filter. Valid levels are `warn`, `low`, `high`, `debug`, `info`.
  - `host_filter` (Optional): Sender host filter.
  - `tag_filter` (Optional): Syslog tag filter.
  - `message_filter` (Optional): Syslog message filter.
  - `start_time` (Optional): Start time for the search (default is -1h).
  - `end_time` (Optional): End date and time for the search (default is now).
  - `limit` (Optional): Maximum number of syslogs to search.

### 11. `get_syslog_summary`

- **Title:** Get syslog summary with filters
- **Description:** Get a summary of syslogs with filters.
- **Arguments:**
  - `level_filter` (Optional): Level filter. Valid levels are `warn`, `low`, `high`, `debug`, `info`.
  - `host_filter` (Optional): Sender host filter.
  - `tag_filter` (Optional): Syslog tag filter.
  - `message_filter` (Optional): Syslog message filter.
  - `start_time` (Optional): Start time for the search (default is -1h).
  - `end_time` (Optional): End date and time for the search (default is now).
  - `top_n` (Optional): Number of top syslog summary entries to return.

### 12. `search_snmp_trap_log`

- **Title:** Search snmp trap log with filters
- **Description:** Search SNMP trap logs of TWSNMP with filters.
- **Arguments:**
  - `from_filter` (Optional): Trap sender filter.
  - `trap_type_filter` (Optional): Trap type filter.
  - `variable_filter` (Optional): Trap variable filter.
  - `start_time` (Optional): Start time for the search (default is -1h).
  - `end_time` (Optional): End date and time for the search (default is now).
  - `limit` (Optional): Number of SNMP trap logs to search.

### 13. `get_mib_tree`

- **Title:** Get MIB tree of TWSNMP.
- **Description:** Get the MIB tree of TWSNMP using the `get_mib_tree` tool.
- **Arguments:** None.

### 14. `get_ip_address_list`

- **Title:** Get the list of IP address managed by TWSNMP.
- **Description:** Get the list of IP addresses managed by TWSNMP using the `get_ip_address_list` tool.
- **Arguments:** None.

### 15. `get_resource_monitor_list`

- **Title:** Get resource monitor info of TWSNMP
- **Description:** Get resource monitor information of TWSNMP using the `get_resource_monitor_list` tool.
- **Arguments:** None.

### 16. `get_server_certificate_list`

- **Title:** Get the list of server certificates managed by TWSNMP
- **Description:** Get the list of server certificates managed by TWSNMP using the `get_server_certificate_list` tool.
- **Arguments:** None.

### 17. `add_event_log`

- **Title:** Add Event log to TWSNMP
- **Description:** Add an event log to TWSNMP.
- **Arguments:**
  - `level` (Required): Level of the event log. Valid levels are `info`, `normal`, `warn`, `low`, `high`.
  - `node` (Optional): Node name for the event log.
  - `event` (Required): The event log content.

### 18. `get_ip_address_info`

- **Title:** Get IP address information
- **Description:** Get information about an IP address.
- **Arguments:**
  - `ip` (Required): The IP address to get information for.

### 19. `get_mac_address_info`

- **Title:** Get MAC address information
- **Description:** Get information about a MAC address.
- **Arguments:**
  - `mac` (Required): The MAC address to get information for.
