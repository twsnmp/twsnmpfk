# Network Editing

Edit screen for network equipment (such as switches). Supports managed switches (SNMP enabled) and unmanaged switches.

## Settings Parameters (Managed Switch)

* **Name**
  Name of the network equipment.
* **IP address**
  IP address of the network equipment.
* **Maximum number of ports on the side**
  The number of ports before wrapping in the layout display.
* **LLDP**
  Enables LLDP (Link Layer Discovery Protocol) to automatically discover topology connections.
* **ARP monitoring**
  Enables monitoring of the device's ARP table to detect new or changed MAC addresses.
* **SNMP mode**
  SNMP protocol version and security level (SNMPv2c, SNMPv3 auth/priv, etc.).
* **SNMP Community**
  Community name for SNMPv2c.
* **User**
  User ID for SNMPv3.
* **Password**
  Password for SNMPv3 authentication.
* **URL**
  Management URL of the device (if set, double-clicking opens this URL in a browser).
* **Description**
  Supplementary comments or notes.
* **Port**
  List of LAN ports on the device (ports can be edited).

## Settings Parameters (Unmanaged Switch)

* **Name**
  Name of the network equipment.
* **IP address**
  IP address of the network equipment.
* **All ports**
  Total number of physical ports on the device.
* **Maximum number of ports on the side**
  The number of ports displayed on a single horizontal row.
* **URL**
  Management URL.
* **Description**
  Supplementary comments or notes.
* **Port**
  List of LAN ports.

## Button Descriptions

* **[Rediscover]** : (Managed switches only) Refreshes port information from the device using SNMP.
* **[Export]** (Download icon): Exports port definitions to a JSON file.
* **[Import]** (Upload icon): Imports port definitions from a JSON file.
* **[Save]** : Saves network configuration changes.
* **[Help]** : Displays this help.
* **[Cancel]** : Closes the window without saving.
