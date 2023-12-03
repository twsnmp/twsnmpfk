#### Address list

<div class="text-lg mb-2">
This is a list of IP address found by TWSNMP.<br>
Only the IP address in the same segment found in the ARP monitoring function is displayed.<br>
You can detect duplicate and the change in the address.
</div>

![Address list](../../help/en/2023-12-03_05-44-21.png)

>>>
#### Address list item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| State | It is the state of the address.(Normal, duplicate, IP change, Mac change.) |
| Address | IP address.|
| MAC address | MAC address.|
| Node name | The name of the node registered on the map as a management target.|
| Vendor | The name of the vendor corresponding to the MAC address.|
| Final change | This is the last change date and time.|
</div>

>>>
#### Description of button

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Add node | Add the selected IP address to the map.<br> It is displayed only when it is not registered.|
| <Span style = "color: red;"> Delete </span> | Delete the selected IP address.|
| Report | Display the address list report.|
| <Span style = "color: red;"> clear </span> | Clear all address lists.|
| CSV | Export the address list to the CSV file.|
| Excel | Export the address list to the Excel file.|
| Update | Update the address list to the latest state.|
</div>

---
#### Relationship between IP and MAC address (force model)

<div class="text-lg mb-2">
This is a report that shows the relationship between IP address and MAC address with an force model.<br>
The normal address is one -on -one for the IP address and the MAC address.<br>
You can detect MAC addresses using the same IP address on multiple Macs or having multiple IP addresses.</div>

![Address -related force model](../../help/en/2023-12-03_05-49-52.png)

---
#### Relationship between IP and MAC address (circular model)

<div class="text-lg mb-2">
This is a report that shows the relationship between IP address and MAC address with a circular model.<br>
The normal address is one -on -one for the IP address and the MAC address.<br>
You can detect MAC addresses with the same IP address on multiple Macs or have multiple IP addresses.
</div>

![IP address -related circular model](../../help/en/2023-12-03_05-52-16.png)
