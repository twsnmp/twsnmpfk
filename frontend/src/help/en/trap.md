# SNMP TRAP

Inspect and analyze received SNMP TRAP messages. Chronological log volume chart is displayed at the top.

## Table Columns

* **Time**
  Date and time the SNMP TRAP was received.
* **From IP**
  The source host address of the SNMP TRAP.
* **Type**
  Object name or OID identifier representing the TRAP Type.
* **Variables**
  Key-value list of variable bindings contained in the SNMP TRAP.

## Button Descriptions

* **[Filter]** : Open the search filter dialog.
* **[Delete All Logs]** : Delete all SNMP TRAP logs from the database.
* **[Report]** : Open statistical and analytical charts for SNMP TRAPs.
* **[Polling]** : Register a new polling task to monitor for specific SNMP TRAP occurrences.
* **[Copy]** : Copy selected logs to the clipboard.
* **[AI Explain]** : Request TRAP message explanation from the AI (LLM).
* **[CSV]** : Export SNMP TRAPs to a CSV file.
* **[Excel]** : Export SNMP TRAPs to an Excel file.
* **[Reload]** : Refresh the SNMP TRAP list.

## Filter Settings

Search options in the filter dialog (supports regular expressions).

* **From**
  Filter by source host IP.
* **Type**
  Filter by TRAP type or OID.

## Report Types

* **By TRAP Type**
  Distribution chart of SNMP TRAPs by their type.
* **Heatmap**
  Hourly density heatmap of received TRAP logs.
* **By Host**
  TRAP log count ranking by sending hosts.
* **From and Type (3D)**
  3D graph representing Host, Uptime/Time, and TRAP Type distribution.
