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

```
Usage of twsnmpfk:
  -caCert string
    	CA Cert path
  -clientCert string
    	Client cert path
  -clientKey string
    	Client key path
  -datastore string
    	Path to data store directory
  -kiosk
    	Kisok mode(frameless and full screen)
  -lang string
    	Language(en|jp)
  -lock string
    	Disable edit map and lock page(map or loc)
  -maxDispLog int
    	Max log size to diplay (default 10000)
```
>>>
```
  -netflowPort int
    	Netflow port (default 2055)
  -ping string
    	ping mode icmp or udp
  -sFlowPort int
    	sFlow port (default 6343)
  -sshdPort int
    	SSH server port (default 2022)
  -syslogPort int
    	Syslog port (default 514)
  -tcpdPort int
    	tcp server port (default 8086)
  -trapPort int
    	SNMP TRAP port (default 162)
```

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
>>>
|Parameters|Description|
|---|---|
|sshdPort <port>| SSH Server Receive Port (Default 2022)|
|netflowPort <port>| NetFlow/IPFIX receive port (default 2055)|
|sFlowPort <port>| sFlow receiving port (default 6343)|
|tcpdPort <port>| TCP log receiving port (default 8086)|
|caCert <file>| CA certificate for TLS communication with TWLogEye |
|clientCert <file>| Client certificate for mTLS communication with TWLogEye |
|clientKey <file>| Client key for mTLS communication with TWLogEye |

</div>

