#### Event Log

<div class="text-xl mb-2 text-left">
This is the event log screen.<br>
At the top, there is a graph showing the number of logs in chronological order.
</div>

![Event Log](../../help/en/2023-12-03_09-32-12.png)

>>>
#### Event log item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Level | Log level.<br> There is severe, mild, attention, return, and information.|
| Date and time | The date and time of the log is recorded.|
| Type | Log type.<BR> Polling, System, Oprate, User, ArpWatch, |
| Related node | Name of node related to logs.<br> The blank means that there is no related node.|
| Event | This is an event that occurred.|

</div>

>>>
#### Description of button

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Filter | Specify the search conditions and display the log.|
| <Span style = "color: red;"> Delete all logs </span> | Delete all event logs.|
| Report | Displays the event log analysis report.|
| CSV | Export the event log to the CSV file.|
| Excel | Export the event log to the Excel file.|
| Update | Update the list of event logs to the latest state.|

</div>


---
#### Filter

<div class="text-xl mb-2 text-left">
This is a dialog that specifies the search conditions for the event log.
</div>

![Event log filter](../../help/en/2023-12-03_09-34-18.png)

>>>
#### Filter item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Level | Log level.<br> All, there are more attention, more than severe, mild.|
| Type | Log type.<br> Polling, System, Oprate, User, ArpWatch, |
| Related node | Search by node name related to the log.|
| Event | Search by the string of the event that occurred.|

<span style="color:red">The string can be searched by regular expression.</span>

</div>


---
#### By state

<div class="text-xl mb-2 text-left">
This is a report of the number of event logs by state (level).
</div>

![By state](../../help/en/2023-12-03_09-36-05.png)

---
#### Heat map

<div class="text-xl mb-2 text-left">
This is a report of the number of cases of each event log on the heat map.
</div>

![Heatmap](../../help/en/2023-12-03_09-37-42.png)

---
#### By node

<div class="text-xl mb-2 text-left">
This is a report of the number of event logs by node.
</div>

![By node](../../help/en/2023-12-03_09-39-12.png)

---
#### Operating rate

<div class="text-xl mb-2 text-left">
This is a report that uses a chronological graph of the value of the operating rate (OPRATE) in the event log.
</div>

![Operating rate](../../help/en/2023-12-03_09-41-25.png)

---
#### ARP watch

<div class="text-xl mb-2 text-left">
This is a report of the value of the address usage rate (ARPWATCH) in the event log as a chronological graph.
</div>

![ARP Watch](../../help/en/2023-12-03_09-43-24.png)
