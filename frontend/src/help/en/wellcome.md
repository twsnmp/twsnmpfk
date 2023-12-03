## Welcome to TWSNMP FK

Let's create a network management map.

first,<button class="bg-green-600"><Start></button>Click the button.


---
#### Select a data folder
<span class="text-xl">
A folder for storing a database file or an extended MIB that records the map and logs.
</span>


>>>
#### File in the datastore
<span class="text-xl">
You can customize it by saving the following files in the data folder.

| File | Contents |
| --- | --- |
| TWSNMPFK.db | Database file.If it does not exist, it will be created automatically.|
| Services.txt | This is a file to use the service name conversion.(Optional) |
| Mac-vendors-export.csv | Mac A database that indicates the relationship between the MAC address and the vendor name.(Optional) |
| Polling.json | Polling settings (optional) |
| mail_test.html | Notification test mail template (optional) |
| mail_notify.html | Notification mail template (optional) |
| Mail_repot.html | Template of report mail (optional) |
| EXTMIBS/*| Additional reading extended MIB (optional) |
</span>

---
#### Display map
<span class="text-xl">
Select a folder to display a blank map.</span>

---
#### Starting parameter
<span class="text-xl">
You can specify the following parameters at startup.
</span>

<pre class="text-sm font-mono">
Usage of twsnmpfk:
  -datastore string
    	Path to data dtore directory
  -kiosk
    	Kisok mode(frameless and full screen)
  -lang string
    	Language(en|jp)
  -lock string
    	Disable edit map and lock page(map or loc)
  -maxDispLog int
    	Max log size to diplay (default 10000)
  -ping string
    	ping mode icmp or udp
  -syslogPort int
    	Syslog port (default 514)
  -trapPort int
      SNMP TRAP port (default 162)
</pre>

>>>
#### Explanation of startup parameter

<div class="text-xl">

| Parameters | Description |
| --- | --- |
| dataStore | Datstore Pass |
| kiosk | Kiosk mode (frameless, full screen) |
| lock <page> | Fix the page that prohibits the editing of the map and displays it <BR> (specified Map or LOC for Page) |
| Maxdisplog <number> | Maximum number of logs (default 10000) |
| ping <Mode> | Ping operation mode (ICMP or UDP) |
| syslogPort <PORT> | Syslog receiving port (default 514) |
| trapPort <Port> | SNMP TRAP Reception port (Default 162) |

</div>

