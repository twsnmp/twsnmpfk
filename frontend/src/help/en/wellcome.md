# Welcome to TWSNMP FK

Welcome screen for starting TWSNMP FK, selecting the datastore directory, and configuring initial settings.

## Data Folder

The folder where TWSNMP FK stores its database, logs, and custom files. You will be prompted to select this folder when starting the application.

## Files in the Datastore

You can customize application behavior by placing the following files in the datastore folder.

* **twsnmpfk.db**
  The main database file. Created automatically if it does not exist.
* **services.txt**
  Text file mapping port numbers to service names (optional).
* **mac-vendors-export.csv**
  Database mapping MAC address OUIs to manufacturer/vendor names (optional).
* **polling.json**
  Custom polling templates definition file (optional).
* **mail_test.html**
  HTML template for test alert emails (optional).
* **mail_notify.html**
  HTML template for alert notification emails (optional).
* **mail_report.html**
  HTML template for daily report emails (optional).
* **extmibs/***
  Directory to store custom external MIB files (optional).

## Button Descriptions

* **[Start]** : Open the folder selection dialog to select a datastore and launch the map.
* **[Stop]** : Terminate the application and exit.
* **[Help]** : Open this help document.

## Startup Parameters

The following parameters can be specified via the command line at startup.

* **-datastore <path>** : Path to the datastore directory.
* **-kiosk** : Launch in kiosk mode (frameless and full screen).
* **-lang <lang>** : Interface language selection (`en` or `ja`).
* **-lock <page>** : Disable map edits and lock view to the specified page (`map` or `loc`).
* **-maxDispLog <number>** : Maximum number of logs displayed in lists (default: 10000).
* **-ping <mode>** : PING operation mode (`icmp` or `udp`).
* **-syslogPort <port>** : Syslog server receiving port (default: 514).
* **-trapPort <port>** : SNMP TRAP server receiving port (default: 162).
* **-sshdPort <port>** : SSH server listening port (default: 2022).
* **-netflowPort <port>** : NetFlow/IPFIX collector port (default: 2055).
* **-sFlowPort <port>** : sFlow collector port (default: 6343).
* **-tcpdPort <port>** : TCP log receiving port (default: 8086).
* **-caCert <file>** : CA certificate path for TLS communication with TWLogEye.
* **-clientCert <file>** : Client certificate path for mTLS communication with TWLogEye.
* **-clientKey <file>** : Client private key path for mTLS communication with TWLogEye.
