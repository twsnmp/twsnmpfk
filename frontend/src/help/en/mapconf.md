#### Map settings
<div class="text-xl">
This is the screen to set the management map.
</div>

![Map settings](../../help/en/2023-12-03_10-29-27.png)

>>>

<div class="text-xl">

| Items | Contents |
| ---- | ---- |
| Map name | Map name.It will be displayed in the upper left of the screen.<br> Please give your favorite name.|
| Icon size | It is the size of the icon to be displayed on the map.|
| Polling interval | Default polling interval.|
| Timeout | Default timeout.|
| Retry | Default number of retry times.|
| Log saving days | It is the number of days to save the log.The log will be deleted automatically after passing.|
| SNMP mode | SNMP version and type of encryption.(SNMPV1, SNMPv2C, SNMPv3) |
| SNMP Community | Community name for SNMPV1, V2C.|
| SNMP user | User name at SNMPv3.|
| SNMP password | Password name for SNMPv3.|
| Syslog | Receive syslog.|
| SNMP Trap | Receive SNMP Trap.|
| SSH Server | start SSH server.|
| ARP Watch | Enable ARP monitoring function.|
</div>

---
#### When you want to change the receiving port of syslog, SNMP Trap

<div class="text-xl">

The port number is specified by the startup parameter of the program.

</div>

```
  -syslogPort int
    	Syslog port (default 514)
  -trapPort int
      SNMP TRAP port (default 162)
  -sshdPort int
      SSH Server port (default 2022)
```

<p style="color:red;font-size: 16px;">
* If SYSLOG or SNMP Trap cannot be received, check the OS and security software firewall settings.
</p>
