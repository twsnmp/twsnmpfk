#### Node list

<div class="text-xl mb-2">
A list of nodes to be managed.
</div>

>>>
#### Node list item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| State | Node condition.<br> Severe, mild, precautions, return, normal, unknown.|
| Name | Node name.|
| IP address | Node IP address.|
| MAC address | Node MAC address.|
| Vendor | The name of the vendor corresponding to the MAC address.|
| Description | Supplementary information about nodes.|
</div>

>>>
#### Description of button

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Edit | Edit node settings.|
| Polling | Displays a list of polling related to the selected node.|
| Report | Displays the selected node analysis report.|
| <Span style = "color: red;"> Delete </span> | Delete the selected node.|
| Reconfirm | Reconfirm the polling of the selected node.|
| Remost confirmation | Reconfirm all nodes polling.|
| CSV | Export the node list to the CSV file.|
| Excel | Export the node list to the Excel file.|
| Reload | Update the node list to the latest state.|

</div>


---
#### Node polling list

<div class="text-xl mb-2">
A list of polling related to nodes.
</div>

>>>

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| State | Polling state.<br> Severe, mild, precautions, return, normal, unknown.|
| Name | Polling name.|
| Level | Pauling level.|
| Type | Polling type.<br> Ping, SNMP, TCP, etc. |
| Log | Log mode.|
| Last confirmation | This is the last date and time when polling was implemented.|
</div>

>>>
#### Description of button

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Add | Add polling to nodes.|
| Edit | Edit the selected polling.|
| Copy | Create a selected polling copy.|
| Report | Displays the selected polling analysis report.|
| <Span style = "color: red;"> Delete </span> | Delete the selected polling.|
| Reload | Update the polling list to the latest state.|
| Close | Close the list of polling.|

</div>


---
#### Basic information report

<div class="text-xl mb-2">
Basic information about nodes.
</div>

---
#### log

<div class="text-xl mb-2">
This is an event log related to the node.
</div>

---
#### Panel

<div class="text-xl mb-2">
Displays the appearance of the node.<br>
Displays the port from the acquisition of the interface mib by SNMP or the line connection information.<br>
The <physical port> switch can only be displayed on the physical port.<br>
Rotate the panel display with the <rotation> switch.
</div>

---
#### Host information

<div class="text-xl mb-2">
Displays the information of the host resource mib of SNMP.<br>
<Span style = "color: red;"> If it is not compatible with the host resource MIB, it cannot be displayed.</span>
</div>

---
#### Storage

<div class="text-xl mb-2">
Displays the storage information of SNMP host resource mib.<br>
When you select, the addition button of the polling will be displayed.<br>
<Span style = "color: red;"> If it is not compatible with the host resource MIB, it cannot be displayed.</span>
</div>

---
#### Device

<div class="text-xl mb-2">
Displays the device information of the SNMP host resource MIB.<br>
<Span style = "color: red;"> If it is not compatible with the host resource MIB, it cannot be displayed.</span>
</div>

---
#### File System

<div class="text-xl mb-2">
Displays File System, information on SNMP host sources MIB.<br>
<Span style = "color: red;"> If it is not compatible with the host resource MIB, it cannot be displayed.</span>
</div>


---
#### Process

<div class="text-xl mb-2">
Displays the process information of SNMP host resource mib.<br>
When you select, the addition button of the polling will be displayed.<br>
<Span style = "color: red;"> If it is not compatible with the host resource MIB, it cannot be displayed.</span>
</div>

