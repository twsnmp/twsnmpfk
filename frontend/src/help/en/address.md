# Address List

List of IP and MAC addresses detected on the local segment via the ARP watch function. Useful for monitoring address assignments, detecting duplicate IPs, and tracking MAC/IP changes.

## Table Columns

* **State**
  Status of the address (Normal, Duplicate, IP Change, MAC Change).
* **Address**
  IP address.
* **Domain**
  Domain name resolved via reverse IP lookup.
* **MAC Address**
  MAC address.
* **Node Name**
  The registered name of the node on the map, if it is monitored.
* **Vendor**
  Vendor name resolved from the MAC address OUI.
* **Risk**
  Risk score or threat level associated with the IP address.
* **Final Change**
  Date and time when the address entry was last updated or changed.

## Button Descriptions

* **[Node Info]** : Open the detailed diagnostic report for the selected registered node.
* **[Edit Node]** : Open the configuration editor for the selected registered node.
* **[Add Node]** : Add the selected unregistered IP address to the map as a new node.
* **[Address Info]** : Open the IP/MAC details window (displaying GeoIP, MAC vendor, history).
* **[Delete]** : Delete the selected address entries.
* **[Copy]** : Copy the selected address rows to the clipboard.
* **[Report]** : Open the Address List statistical reports (IP usage heatmap and IP-MAC relationship charts).
* **[AI Explain]** : Request an AI (LLM) explanation of the address management status, sending normal address count and detailed abnormal/changed address entries for diagnosis and recommendations (Displayed only when AI integration is enabled).
* **[STUN]** : Open the STUN information dialog to inspect the external public IP address, mapped port, reverse DNS, local address, RTT, and GeoIP via STUN servers (supports IPv4/IPv6, copying, and links to Google Maps/VirusTotal).
* **[Clear]** : Clear all addresses from the ARP monitoring list.
* **[CSV]** : Export the address list to a CSV file.
* **[Excel]** : Export the address list to an Excel file.
* **[Reload]** : Refresh the address list.

## Report Descriptions

* **IP Address Usage Status (IPAM)**
  Visualizes IP address allocation and pool utilization across subnets.
  * **Subnet Relative Density Heatmap**: Aggregates all address ranges normalized into 100 buckets, providing a lightweight, high-speed overview of utilization across large-scale networks.
  * **Subnet Range List Table**: Lists configured address ranges with Start/End IP, pool size, active count, and usage rate. Clicking a row selects and switches the detailed view.
  * **Hierarchical Drill-Down**: For wide address spaces, displays /24 subnet blocks with aggregated usage. Clicking a block drills down into individual host allocation (1–254).
  * **Host Search**: Real-time search by hostname, IP address, or MAC address with inline visual highlighting.
* **[AI Explain]** (Inside Report Dialog)
  Sends current tab data (IPAM / IP-MAC relation) to AI (LLM) for subnet allocation efficiency, IP conflict risk analysis, and recommendations (only visible when AI integration is enabled).
* **IP-MAC Relationship (Force-directed Model)**
  Visualizes IP and MAC relationships in a dynamic graph topology. Helps quickly identify abnormal connections (e.g., multiple MACs sharing one IP).
* **IP-MAC Relationship (Circular Model)**
  Arranges IP and MAC relationships in a circular layout. Provides a clean overview of 1-to-1 mappings and duplicates even in dense networks.
