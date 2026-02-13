# TWSNMP FK

[日本語版はこちら](README_ja.md)

[![Go Report Card](https://goreportcard.com/badge/github.com/twsnmp/twsnmpfk)](https://goreportcard.com/report/github.com/twsnmp/twsnmpfk)
![GitHub Go version](https://img.shields.io/github/go-mod/go-version/twsnmp/twsnmpfk)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/twsnmp/twsnmpfk)
![GitHub License](https://img.shields.io/github/license/twsnmp/twsnmpfk)
![GitHub Repo stars](https://img.shields.io/github/stars/twsnmp/twsnmpfk?style=social)
【Built with】
![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Wails](https://img.shields.io/badge/wails-%23E24329.svg?style=for-the-badge&logo=wails&logoColor=white)
![Svelte](https://img.shields.io/badge/svelte-%23f1413d.svg?style=for-the-badge&logo=svelte&logoColor=white)

TWSNMP FK is a next-generation Network Management System. It combines the performance of Go, the simplicity of Svelte, and the seamless desktop experience of Wails to provide a lightweight yet powerful observability tool.

![TWSNMP FK](docs/images/en/2026-02-13_04-41-52.png)

---

Ultra lightweight SNMP manager.
To keep maps and event logs, etc. always visible.
It is designed to be used in Windows kiosk mode.
Of course, it can also be used as a normal application.

![](doc/images/en/2025-03-11_06-29-01.png)

## Document

[English](https://twsnmp.github.io/twsnmpfk/index.html)

## Status

The following functions will work

- Network map
- Node list
- Polling (PING/TCP/HTTP/NTP/DNS/SNMP/gNMI)
- Event log
- Syslog reception
- SNMP TRAP reception
- ARP monitoring
- MIB browser
- PING Confirmation
- Panel Display
- Host resource MIB display
- Wake On LAN support
- HTML e-mail notification, periodic report
- AI Analysis
- NetFlow/IPFIX
- sFlow
- gNMI
- PKI (CA and CRL/OCSP/ACME/SCEP server)
- SSH Server
- TCP Log server
- OpenTelemetry collector
- MCP Server
- MQTT Server and Polling

## Build 

The following environment is used for development

 - go 1.24 or higher
 - wails 2.9.3 or higher
 - nsis
 - go-task

You can build it with the following command.

 ````
 task
 ````

 ## Run

 Double-click from the built executable file to drive it as a normal application.
It can also be started from the command line by specifying the following parameters

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
  -mcpCert string
    	MCP server cert path
  -mcpKey string
    	MCP server key path
  -mqttCert string
    	MQTT server cert path
  -mqttFrom string
    	MQTT client IP
  -mqttKey string
    	MQTT server key path
  -mqttTCPPort int
    	MQTT server TCP port (default 1883)
  -mqttUsers string
    	MQTT user and password
  -mqttWSPort int
    	MQTT server WebSock port (default 1884)
  -netflowPort int
    	Netflow port (default 2055)
  -notifyOAuth2Port int
    	OAuth2 redirect port (default 8180)
  -otelCA string
    	OpenTelementry CA cert path
  -otelCert string
    	OpenTelemetry server cert path
  -otelGRPCPort int
    	OpenTelemetry server gRPC port (default 4317)
  -otelHTTPPort int
    	OpenTelemetry server HTTP port (default 4318)
  -otelKey string
    	OpenTelemetry server key path
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

---

| Parameters | Description |
| --- | --- |
| dataStore | Datstore Pass |
| kiosk | Kiosk mode (frameless, full screen) |
| lock <page> | disable edit map and show fixed page |
| Maxdisplog <number> | Maximum number of logs (default 10000) |
| ping <Mode> | Ping operation mode (ICMP or UDP) |
| syslogPort <PORT> | Syslog receiving port (default 514) |
| trapPort <Port> | SNMP TRAP Reception port (Default 162) |
| sshdPort <Port> | SSH server port (Default 162) |
| sshdPort <port> | SSH Server Receive Port (Default 2022)|
| netflowPort <port> | NetFlow/IPFIX receive port (default 2055)|
| sFlowPort <port> | sFlow receiving port (default 6343)|
| tcpdPort <port> | TCP log receiving port (default 8086)|
| otelCA |OpenTelementry CA cert path|
| otelCert |OpenTelemetry server cert path|
| otelGRPCPort |OpenTelemetry server gRPC port (default 4317)|
| otelHTTPPort |OpenTelemetry server HTTP port (default 4318)|
| otelKey |OpenTelemetry server key path|
| mqttTCPPort |MQTT server TCP port (default 1883)|
| mqttWSPort |MQTT server Websock port (default 1884)|
| mqttCert |MQTT server cert path|
| mqttKey |MQTT server key path|
| mqttFrom |MQTT server Client|
| mqttUsers |MQTT server User ID and password list|
| mcpCert |MCP server cert path|
| mcpKey |MCP server key path|
| notifyOAuth2Port |OAuth2 redirect port (default 8180)|
