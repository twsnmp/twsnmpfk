# Syslog

View, search, and analyze received Syslog messages. A chronological log volume chart is displayed at the top.

## Table Columns

* **Level**
  Severity level of the log (Severe, Mild, Warn, Info).
* **Time**
  Date and time the Syslog message was received.
* **Host**
  The source host of the Syslog.
* **Type**
  Categorized representation of the Syslog facility and priority.
* **Tag**
  Process name and PID tags.
* **Message**
  The Syslog message text body.

## Button Descriptions

* **[Filter]** : Open the search filter dialog.
* **[Delete All Logs]** : Delete all Syslogs from the database.
* **[Report]** : Open statistical and analytical charts for Syslog logs.
* **[Magic]** (or Magic Analysis) : Automatically generate Grok patterns to extract and analyze structured information from the logs.
* **[Polling] / [AI Assist]** : Register a new polling task for the selected Syslog. When AI integration is enabled, this changes to `[AI Assist]`, opening a dialog where AI automatically generates optimal monitoring parameters, filters, Grok patterns, and scripts based on the selected log text.
* **[Copy]** : Copy selected logs to the clipboard.
* **[AI Explain]** : Request log analysis explanation from the AI (LLM) (Displayed only when AI integration is enabled).
* **[IP/MAC Info]** : Dropdown button to view detailed information of detected IP or MAC addresses.
* **[CSV]** : Export Syslogs to a CSV file.
* **[Excel]** : Export Syslogs to an Excel file.
* **[Reload]** : Refresh the Syslog list.

## Filter Settings

Search options in the filter dialog (supports regular expressions).

* **Level**
  Minimum severity level to display (All, Info, Warn, Low, High).
* **Host**
  Filter by source host.
* **Tag**
  Filter by process tags.
* **Message**
  Filter by message text content.

## Report Types

* **By State**
  Distribution of Syslogs by severity levels.
* **Heatmap**
  Hourly density heatmap of received logs.
* **By Host**
  Log count ranking by sending hosts.
* **By Host (3D)**
  3D graph representing Host, Uptime/Time, and Priority distribution.
* **Periodicity Analysis by FFT**
  Fast Fourier Transform analysis of host log generation cycles.
* **[AI Explain]**
  Sends current Syslog entries, top sender hosts, and severity trends to AI (LLM) for automated anomaly explanation and risk assessment (only visible when AI integration is enabled).
