# Node

List of managed nodes, settings parameters, and detailed performance reports.

## Table Columns

* **State**
  Status of the node (Severe, Mild, Warn, Recovery, Normal, Unknown).
* **Name**
  Registered name of the node.
* **IP Address**
  IP address or hostname of the node.
* **MAC Address**
  MAC address of the node.
* **Vendor**
  Vendor name resolved from the MAC address OUI.
* **Description**
  Supplementary description or memo.

## Button Descriptions

* **[Edit]** : Open the node configuration editor (to edit IP, icon, SNMP/SSH authentication, etc.).
* **[Polling]** : View the list of pollings (monitoring tasks) defined for the selected node.
* **[Report]** : Open the node's detailed diagnostic and performance reports.
* **[Action]** : Dropdown button for executing network tools (PING, MIB Browser, gNMI Tool, Wake on LAN).
* **[Delete]** : Delete the selected nodes from the map and monitoring database.
* **[Recheck]** : Run all pollings for the selected nodes immediately.
* **[Map Items]** : Switch view to map items list (lines, networks, drawing items).
* **[Check All]** : Run all pollings for all nodes immediately.
* **[CSV]** : Export the node list to a CSV file.
* **[Excel]** : Export the node list to an Excel file.
* **[Reload]** : Refresh the node list.

## Node Polling List

Polling settings list screen for the selected node.

* **[Add]** : Add a new polling to the node.
* **[Edit]** : Edit the selected polling.
* **[Copy]** : Duplicate the selected polling.
* **[Report]** : Display the performance history reports for the selected polling.
* **[Delete]** : Delete the selected polling.
* **[Reload]** : Refresh the polling list.
* **[Close]** : Close the polling list screen.

## Node Report Tabs

* **Basic Info**
  Displays basic node profiles, overall statuses, and quick action buttons.
* **Log**
  Event logs related to the selected node.
* **Memo**
  Free-text area for writing local notes or memos about the node.
* **Panel**
  Visual representation of the node's port layout (VPanel). Ports can be filtered to physical only, and the panel orientation can be rotated.
* **Host Info**
  Information retrieved from SNMP Host Resources MIB (CPU usage, system uptime, host name, etc.). Only available if the node supports Host Resources MIB.
* **Storage**
  Storage capacity details (RAM, Disk partitions). Select a row to create a storage space monitoring polling. Only available if the node supports Host Resources MIB.
* **Device**
  Details of recognized hardware components (CPUs, drives, ports, etc.). Only available if the node supports Host Resources MIB.
* **File System**
  Details of mounted file systems. Only available if the node supports Host Resources MIB.
* **Process**
  List of running processes. Select a row to create a process monitoring polling. Only available if the node supports Host Resources MIB.
