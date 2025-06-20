#### Map settings
<div class="text-xl">
This is the screen to set the management map.
</div>

<div class="text-lg">

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

>>>

#### Map Settings (continued)

<div class="text-lg">

|Item|Content|
|----|---|
|TCP Server|Received on the TCP server.|
|OpenTelemetry|Start the OpenTelemetry server.|
|OpenTelemetry Retention Time|Specify the retention time for OpenTelemetry data.|
|OpenTelemetry Source|Limits the source of data to the OpenTelemetry server.|
|MCP Server Transport|Specify the transport of the MCP server (OFF/SSE/Steamable).|
|MCP Server Endpoint|Specify the incoming IP and port of the MCP server.|

</div>

---
#### When you want to change the receiving port of syslog, SNMP Trap

<div class="text-xl">

The port number is specified by the startup parameter of the program.

```
  -netflowPort int
    	Netflow port (default 2055)
  -otelGRPCPort int
    	OpenTelemetry server gRPC port (default 4317)
  -otelHTTPPort int
    	OpenTelemetry server HTTP port (default 4318)
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

</div>

<p style="color:red;font-size: 16px;">
* If SYSLOG or SNMP Trap cannot be received, check the OS and security software firewall settings.
</p>
