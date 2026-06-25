# Node Editing

Screen for creating a new node or editing the settings of an existing node.

## Settings Parameters

* **Name**
  Display name of the node.
* **IP address**
  IP address or host name of the node.
* **Address mode**
  Resolution and tracking mode for the IP address ("Fixed IP Address", "Fixed MAC Address", "Fixed Host Name").
* **Icon**
  Icon to display on the map (standard icons or custom registered image icons).
* **Automatic confirmation when returning**
  Enables automatic acknowledgment (status reset to normal) when the node recovers from a failure state.
* **SNMP mode**
  SNMP protocol version and security level (SNMPv2c, SNMPv3 auth/priv, etc.).
* **SNMP Community**
  Community name for SNMPv2c.
* **User**
  User ID for SNMPv3.
* **SSH User**
  Username for SSH/SFTP polling connections.
* **Password**
  Password for SNMPv3 or SSH access.
* **Public key**
  Public key of the node for SSH polling. If left blank, it is automatically retrieved and set during the first connection.
* **URL**
  Management URL(s) of the node (comma-separated list. Used in right-click menus and for TWSNMP federation links).
* **Description**
  Supplementary comments or notes.

## Button Descriptions

* **[Save]** : Saves node configuration changes.
* **[Help]** : Displays this help.
* **[Cancel]** : Closes the window without saving.
