#### Address list

<div class="text-lg">
This is a list of IP address found by TWSNMP.<br>
Only the IP address in the same segment found in the ARP monitoring function is displayed.<br>
You can detect duplicate and the change in the address.

| Items | Contents |
| ---- | ---- |
| State | It is the state of the address.(Normal, duplicate, IP change, Mac change.) |
| Address | IP address.|
| Domain | Domain information obtained from reverse IP lookup, etc.|
| MAC address | MAC address.|
| Node name | The name of the node registered on the map as a management target.|
| Vendor | The name of the vendor corresponding to the MAC address.|
| Risk | Risk level determined from the IP address.|
| Final change | This is the last change date and time.|
</div>

>>>
#### Address list(button)

<div class="text-lg">

| Items | Contents |
| ---- | ---- |
| Add node | Add the selected IP address to the map.<br> It is displayed only when it is not registered.|
| <Span style = "color: red;"> Delete </span> | Delete the selected IP address.|
| Report | Display the address list report.|
| <Span style = "color: red;"> clear </span> | Clear all address lists.|
| CSV | Export the address list to the CSV file.|
| Excel | Export the address list to the Excel file.|
| Reload | Update the address list to the latest state.|
</div>

---
#### IP address usage status

<div class = "text-lg mb-4">
The heat map and list shows the usage status of the specified IP address.
</div>

#### Relationship between IP and MAC address (force model)

<div class="text-lg mb-4">
This is a report that shows the relationship between IP address and MAC address with an force model.
The normal address is one -on -one for the IP address and the MAC address.
You can detect MAC addresses using the same IP address on multiple Macs or having multiple IP addresses.
</div>



#### Relationship between IP and MAC address (circular model)

<div class="text-lg mb-4">
This is a report that shows the relationship between IP address and MAC address with a circular model.
The normal address is one -on -one for the IP address and the MAC address.
You can detect MAC addresses with the same IP address on multiple Macs or have multiple IP addresses.
</div>
