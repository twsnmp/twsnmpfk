# Map Settings

Configuration of the management map, server functions, and data collection services.

## Settings Parameters

* **Map Name**
  Name of the map, displayed in the upper-left corner of the screen.
* **Map Size**
  Canvas size for the map (Auto, A4 Portrait, A4 Landscape).
* **Icon Size**
  Size of the node icons displayed on the map.
* **Polling Interval**
  Default execution interval for pollings (seconds).
* **Timeout**
  Default communication timeout (seconds).
* **Retry**
  Default number of communication retries.
* **Log Saving Days**
  Retention period for event logs and statistical data. Expired data is deleted automatically.
* **OpenTelemetry Retention**
  Retention time for OpenTelemetry data (hours).
* **OpenTelemetry Source**
  Limit incoming OpenTelemetry traffic to the specified IP source.
* **MCP Server Transport**
  Transport setting for the MCP (Model Context Protocol) server (OFF / SSE / Streamable).
* **MCP Server Endpoint**
  IP address and port for the MCP server.
* **MCP Server From**
  Limit incoming MCP connections to the specified IP source.
* **MCP Server Token**
  Access token for MCP server authentication.
* **AI Provider**
  LLM (Large Language Model) provider to use.
* **AI API Base URL**
  Base URL of the LLM API.
* **AI API Key**
  Access key for the LLM API.
* **AI Model**
  LLM model name.
* **SNMP Mode**
  Default SNMP version and security mode (v1 / v2c / v3).
* **SNMP Community**
  Community string for SNMPv1/v2c.
* **SNMP User**
  Username for SNMPv3.
* **SNMP Password**
  Authentication and encryption password for SNMPv3.
* **Syslog**
  Enable Syslog receiving server.
* **NetFlow**
  Enable NetFlow collection server.
* **sFlow**
  Enable sFlow collection server.
* **SNMP TRAP**
  Enable SNMP TRAP receiving server.
* **ARP Watch**
  Enable ARP monitoring.
* **ARP Watch Range**
  IP address range for ARP monitoring.
* **ARP Timeout**
  Retention timeout for ARP cache entries.
* **SSH Server**
  Enable built-in SSH server.
* **TCP Server**
  Enable TCP server for receiving logs via TCP.
* **OpenTelemetry**
  Enable OpenTelemetry receiving server.
* **MQTT**
  Enable MQTT broker/client functionality.
* **MQTT -> Syslog**
  Enable recording of received MQTT messages to Syslog.

## Changing Listening Ports

Listening ports for servers such as Syslog, SNMP TRAP, or NetFlow are specified as program startup command-line parameters.

* **-netflowPort** : Netflow port (default: 2055)
* **-otelGRPCPort** : OpenTelemetry server gRPC port (default: 4317)
* **-otelHTTPPort** : OpenTelemetry server HTTP port (default: 4318)
* **-sFlowPort** : sFlow port (default: 6343)
* **-sshdPort** : SSH server port (default: 2022)
* **-syslogPort** : Syslog port (default: 514)
* **-tcpdPort** : TCP server port (default: 8086)
* **-trapPort** : SNMP TRAP port (default: 162)

If Syslog or SNMP TRAP packets are not received, check the firewall settings of the OS or security software.
