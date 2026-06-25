# Map

Visual representation of network topology, node statuses, and link connections.

## Screen Structure

* **Toolbar**
  Area for switching screens and other general operations.
* **Map**
  Area for displaying nodes, drawing items, and network configurations.
* **Event Log**
  Display of the latest 100 event logs.

## Map Menu

Menu displayed when right-clicking on an empty space on the map.

* **[Add Node]** : Manually add a node to the map.
* **[Draw Item]** : Add a drawing item (such as shapes or text labels) to the map.
* **[Add Network]** : Add a network subnet to the map.
* **[Check All]** : Immediately check the statuses of all nodes currently in error state.
* **[Discover]** : Open the auto-discovery settings screen.
* **[Import]** : Import a map file from TWSNMP v4.x.
* **[Grid]** : Align icons to the specified grid interval.
* **[Back Image]** : Open the background image configuration.
* **[Reload]** : Reload the map to the latest state.
* **[Normal View / Edit View]** : Toggle between normal view and draw item editing mode.
* **[Show Node Info / Hide Node Info]** : Toggle the display of detailed node information on the map.

## Background Image Settings

Dialog for setting the map background image.

* **X**
  X-coordinate of the image's top-left corner.
* **Y**
  Y-coordinate of the image's top-left corner.
* **Width**
  Display width of the background image.
* **Height**
  Display height of the background image.
* **[Select File]** : Select the image file to use for the background.
* **[Clear]** : Clear the background image.
* **[Save]** : Save and apply the background image settings.

## Grid Alignment

Dialog for aligning node icons to a specified grid interval.

* **Size**
  Grid interval in pixels.
* **[Test]** : Preview the node arrangement before applying the grid alignment.
* **[Save]** : Apply and save the grid alignment.

## Node Menu

Menu displayed when right-clicking a node on the map.

* **[Report]** : Display the node report screen.
* **[PING]** : Display the PING tool screen.
* **[MIB Browser]** : Open the MIB browser tool.
* **[gNMI Tool]** : Open the gNMI tool.
* **[Wake on LAN]** : Send a Wake-on-LAN magic packet to the node.
* **[Edit]** : Open the node configuration editor.
* **[Polling]** : Display the polling settings list for the node.
* **[Reconfirm]** : Run the polling immediately to reconfirm the node's status.
* **[Copy]** : Duplicate the node.
* **[Delete]** : Delete the node from the map.

## Drawing Item Menu

Menu displayed when right-clicking a drawing item.

* **[Edit]** : Open the drawing item editor.
* **[Copy]** : Duplicate the drawing item.
* **[Delete]** : Delete the drawing item.

## Network Menu

Menu displayed when right-clicking a network subnet icon.

* **[Reconfirm]** : Reconfirm the status of the network.
* **[Edit]** : Open the network configuration editor.
* **[Line Edit]** : Edit the lines connected to this network.
* **[Ping]** : Run PING check for all addresses in this network.
* **[MIB Browser]** : Open the MIB browser for the network.
* **[Find Connection]** : Search for connected ports of this network.
* **[Delete]** : Delete the network.

## Alignment Menu (Multiple Nodes Selected)

Menu displayed when multiple nodes are selected and right-clicked.

* **[Horizontal]** : Align selected nodes horizontally.
* **[Vertical]** : Align selected nodes vertically.
* **[Circle]** : Arrange selected nodes in a circle.
