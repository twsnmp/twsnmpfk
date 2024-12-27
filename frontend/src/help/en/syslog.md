#### Syslog

<div class="text-xl mb-2">
Syslog screen.
At the top, there is a graph showing the number of logs in chronological order.
</div>

<div class="text-lg">

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
#### syslog<button>

<div class="text-lg">

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
#### syslog filter

<div class="text-xl mb-2">
This is a dialog that specifies the search conditions for syslog.
</div>

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Level | Syslog level.<br> All, more than information, more than caution, mild or higher, more severe.|
| Host | It is the source host.|
| Tags | The value of the syslog tag.|
| Message | Syslog message.|

<Span style = "color: red"> Character strings can be searched in regular expressions.</span>

</div>

---
#### By state
<div class="text-xl mb-4">
This is a report of the number of syslogs by state.
</div>

#### Heat map
<div class="text-xl mb-4">
This is a report of the number of cases of syslog on the heat map.
</div>

#### By host
<div class="text-xl mb-4">
This is a report of the number of syslogs by the source host.
</div>

---
#### By host (3D)
<div class="text-xl mb-4">
This is a report displayed in three -dimensional graphs of Syslog, source host, priority, and time.
</div>

#### Catalysis by FFT
<div class="text-xl mb-4">
This is a report that analyzes Syslog for each host and analyzes the number of receiving cases.
</div>

