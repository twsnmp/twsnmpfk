#### Syslog

<div class="text-xl mb-2">
Syslog screen.<br>
At the top, there is a graph showing the number of logs in chronological order.
</div>

![Syslog](../../help/en/2023-12-03_11-43-37.png)

>>>
#### Syslog item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Level | Syslog level.<br> There is severe, mild, precautions, and information.|
| Date and time | It is the date and time when I received syslog.|
| Host | SYSLOG source host.|
| Type | Syslog Facility and priority string.|
| Tags | Syslog tag.Process and process ID.|
| Message | Syslog message.|

</div>

>>>
#### Description of button

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Polling | Register the polling from the selected syslog.|
| Filter | Specify the search conditions and display syslog.|
| <Span style = "color: red;"> Delete all logs </span> | Delete all syslogs.|
| Report | Displays Syslog analysis reports.|
| Export CSV | syslog to CSV file.|
| Excel | EXCEL file is exported to syslog.|
| Reload | Update the list of syslog to the latest state.|

</div>


---
#### Filter

<div class="text-xl mb-2">
This is a dialog that specifies the search conditions for syslog.
</div>

![Syslog filter](../../help/en/2023-12-03_11-45-43.png)

>>>
#### Filter item

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Level | Syslog level.<BR> All, more than information, more than caution, mild or higher, more severe.|
| Host | It is the source host.|
| Tags | The value of the syslog tag.|
| Message | Syslog message.|

<Span style = "color: red"> Character strings can be searched in regular expressions.</span>

</div>


---
#### By state

<div class="text-xl mb-2">
This is a report of the number of syslogs by state.
</div>

![By state of syslog](../../help/en/2023-12-03_11-47-06.png)

---
#### Heat map

<div class="text-xl mb-2">
This is a report of the number of cases of syslog on the heat map.
</div>

![Heat map](../../help/en/2023-12-03_11-47-15.png)

---
#### By host

<div class="text-xl mb-2">
This is a report of the number of syslogs by the source host.
</div>

![By host](../../help/en/2023-12-03_11-47-27.png)

---
#### By host (3D)

<div class="text-xl mb-2">
This is a report displayed in three -dimensional graphs of Syslog, source host, priority, and time.
</div>

![By host (3D)](../../help/en/2023-12-03_11-47-39.png)

---
#### Catalysis by FFT

<div class="text-xl mb-2">
This is a report that analyzes Syslog for each host and analyzes the number of receiving cases.
</div>

![Syslog FFT analysis](../../help/en/2023-12-03_11-47-51.png)
