# Auto Discovery

Screen for automatically discovering network devices and hosts not yet registered on the map.

## Settings Parameters

* **Start IP**
  Start address of the IP address range to search.
* **End IP**
  End address of the IP address range to search.
* **Timeout**
  PING timeout in seconds for verifying host availability.
* **Retry**
  Number of PING retries.
* **Port scan**
  Performs a port scan on discovered nodes to identify running services (caution: may trigger security software alerts).
* **Polling automatic setting**
  Automatically creates and assigns polling monitors to discovered nodes.
* **Recheck**
  Rechecks already registered nodes.
* **Add Network**
  Automatically generates network elements on the map based on discovered subnets.

## Button Descriptions

* **[Start]** : Starts the discovery process and opens the status view.
* **[Auto IP Range]** : Automatically sets the IP address range (Class C) based on the local machine's IP address.
* **[Help]** : Displays this help.
* **[Close]** : Closes the window.
* **[Stop]** : (On status view) Stops the active discovery process.

## Status Descriptions

* **Status view during execution**
  Displays execution time, sent/waiting counts, discovered nodes, and SNMP-responsive nodes in real-time.
* **Additional status with port scanning**
  If port scanning is enabled, displays the count of nodes running services such as Web, Mail, SSH, File, RDP/VNC, and LDAP/AD.
