# Network Report

Inspect switch connection details, VPanel, and FDB (Forwarding Database) table information for the selected network map item.

## Tab Structure

* **[Basic Info]** : Profile of the switch including Name, IP, MAC address, Description, and overall status.
* **[Port]** : Port statuses visualized via a virtual hardware panel (VPanel), alongside detailed port metrics (In/Out packets and bytes, link speed, state change time).
* **[FDB]** : FDB table records listing connected nodes, MAC addresses, and vendor details, with a graph visualization.

## Port Table Columns (under Port tab)

* **No.**
  Port index.
* **State**
  Operational status of the port.
* **Name**
  Port name (ifDescr).
* **Type**
  Interface type (e.g., Ethernet).
* **MAC Address**
  Physical MAC address of the interface.
* **Speed**
  Link speed of the port.
* **Out Packets**
  Cumulative sent packets.
* **Out Bytes**
  Cumulative sent bytes.
* **In Packets**
  Cumulative received packets.
* **In Bytes**
  Cumulative received bytes.
* **Last Changed**
  Date and time when the link state last changed.

## FDB Table Columns (under FDB tab)

* **Index**
  Interface index.
* **Port**
  Associated physical port number.
* **VLAN ID**
  VLAN ID assigned to the port.
* **Node**
  Name and IP address of the connected node (mapped via ARP cache).
* **MAC**
  MAC address of the connected device.
* **Vendor**
  MAC address vendor name.

## Button Descriptions

* **[Copy]** : Copy target item fields (e.g., IP address) to the clipboard.
* **[Help]** : Open this help document.
* **[Close]** : Close the network report window.
