# Polling Editing

Screen for creating a new polling monitor or editing the settings of an existing one.

## Settings Parameters

* **Name**
  Name of the polling monitor.
* **Level**
  Severity level when a failure is detected (e.g., Severe, Mild, Warn).
* **Type**
  Monitoring method (PING, SNMP, TCP, HTTP, gRPC, TLS, DNS, SSH/SFTP, Command, etc.).
* **Mode**
  Operation mode based on the selected type.
* **Log mode**
  Method to save and process the results log ("None", "Always", "On change", "AI analysis").
* **MQTT server URL**
  Broker URL for sending polling results via MQTT (e.g., `tcp://localhost:1883`).
* **Topic**
  MQTT topic to publish to (default: `twsnmpfk/polling`).
* **Sent data columns**
  Comma-separated list of variable names to publish via MQTT.
* **AI mode**
  (Visible only when Log mode is "AI analysis") AI algorithm type (e.g., "Isolation Forest").
* **Variables to vectorize**
  (Visible only when Log mode is "AI analysis") Comma-separated variable names of numerical data to analyze.
* **Parameter**
  Configuration parameters depending on type and mode.
  * **Example: Mail Monitoring (IMAP/POP3)**
    * **Mail Server**: Hostname or IP address of the IMAP/POP3 server.
    * **Port**: Port number (commonly 993 for IMAP, 995 for POP3).
    * **User Name**: User account name for the mailbox.
    * **Password**: Password.
    * **Protocol**: IMAP or POP3.
    * **Secure Connection**: Enables SSL/TLS.
    * **Keyword**: Search keyword filter for subjects or bodies (optional).
* **Filter**
  Regex or search filter condition depending on type and mode.
* **Extract pattern**
  Grok pattern to extract structured variable values from raw text/logs.
* **Script**
  JavaScript code to determine failure conditions or calculate custom variables.
* **Polling interval**
  Time interval in seconds between execution checks.
* **Timeout**
  Response timeout limit in seconds.
* **Retry**
  Number of retry attempts if a timeout occurs.
* **Failure Action**
  Actions to execute upon failure detection (WOL, Mail, Webhook, command execution, etc.).
* **Return action**
  Actions to execute when the state recovers from a failure.

## Button Descriptions

* **[Save]** : Saves the polling configuration.
* **[Help]** : Displays this help.
* **[Cancel]** : Closes the window without saving.
