# Polling Editing

Screen for creating a new polling monitor or editing the settings of an existing one.

## Settings Parameters

* **Node**
  Target node associated with the polling monitor (mandatory when creating a new monitor, changeable when editing).
* **Name**
  Name of the polling monitor.
* **Level**
  Severity level when a failure is detected (e.g., Severe, Mild, Warn).
* **Type**
  Monitoring method (PING, SNMP, TCP, HTTP, gRPC, TLS, DNS, SSH/SFTP, Command, etc.).
* **Mode**
  Operation mode based on the selected type.
  * **PING Smokeping Mode (`smoke`)**: Measures statistical metrics (min, max, mean/avg, median response time, packet loss rate %, jitter) by sending continuous PING packets in short intervals.
* **Log mode**
  Method to save and process the results log ("None", "Always", "On change", "AI analysis").
* **MQTT server URL**
  Broker URL for sending polling results via MQTT (e.g., `tcp://localhost:1883`).
* **Topic**
  MQTT topic to publish to (default: `twsnmpfk/polling`).
* **Sent data columns**
  Comma-separated list of variable names to publish via MQTT.
* **AI mode**
  (Visible only when Log mode is "AI analysis") AI algorithm type (e.g., "Isolation Forest", "Hotelling's Theory", "k-NN").
* **Variables to vectorize**
  (Visible only when Log mode is "AI analysis") Comma-separated variable names of numerical data to analyze.
* **Parameter**
  Configuration parameters depending on type and mode.
  * **PING Smokeping Mode Example**: `count=10,size=64,ttl=64` (`count`: continuous ping count, `size`: packet size, `ttl`: TTL)
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
  JavaScript code to determine failure conditions or calculate custom variables (Evaluates to `true` for normal state, `false` for failure).
  * **Variables available in PING Smokeping Mode**:
    * `rtt` / `avg` / `mean`: Mean response time in nanoseconds
    * `min`: Minimum response time in nanoseconds
    * `max`: Maximum response time in nanoseconds
    * `median`: Median response time in nanoseconds
    * `jitter`: Jitter (`max - min`) in nanoseconds
    * `loss`: Packet loss rate (`0.0` to `100.0` %)
    * `fail` / `lossCount`: Number of lost packets
    * `count`: Total count of pings sent
    * `ttl`: RecvTTL
    * Example: `loss < 10.0 && avg < 100 * 1000 * 1000` (Normal if packet loss < 10% and avg response time < 100ms)
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
* **[AI Assist]** : AI (LLM) assistant that generates and suggests monitoring types, parameters, and JavaScript scripts based on prompt requests (Displayed only when AI integration is enabled).
* **[Help]** : Displays this help.
* **[Cancel]** : Closes the window without saving.
