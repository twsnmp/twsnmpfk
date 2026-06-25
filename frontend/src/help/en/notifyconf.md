# Notification Settings

Configure alert notifications, periodic report schedules, external integrations, and local audio/command alerts.

## Settings Parameters

* **Provider**
  The mail service to use (smtp / google / microsoft).
* **Client ID**
  OAuth2 client ID for Google or Microsoft mail services.
* **Client Secret**
  OAuth2 client secret for Google or Microsoft mail services.
* **Tenant Name**
  Active Directory Tenant ID or Name for Microsoft Mail service.
* **Mail Server**
  Hostname (or IP) and port of the SMTP server (e.g., `smtp.example.com:587`).
* **Do not check the server certificate**
  Toggle to disable verification of SSL/TLS certificates (useful for self-signed certificates).
* **User**
  Username for SMTP authentication, or the sender's email address for OAuth2.
* **Password**
  Password for SMTP authentication.
* **Subject**
  Prefix subject line for alert emails.
* **From**
  Sender's email address.
* **To**
  Destination email addresses (comma-separated for multiple addresses).
* **Notification Level**
  Minimum severity level to trigger an email alert (Severe, Mild, Warn, etc.).
* **Notification Interval**
  Interval (minutes) to evaluate statuses and compile notification emails.
* **Periodic Report**
  Enable daily status summary report emails.
* **AI Summary**
  Enable LLM-based system log summarization inside periodic report emails.
* **Notify Repair**
  Enable email notifications when a node status returns to normal.
* **Notification Webhook**
  Webhook URL (HTTP POST) to send JSON payloads upon status alerts.
* **Report Webhook**
  Webhook URL (HTTP POST) to send JSON payloads for periodic reports.
* **Command Execution**
  Local command string to execute upon status changes. The `$level` variable is replaced by the status level (0: Severe, 1: Mild, 2: Warn, 3: Normal, -1: Unknown).
* **Severe Alarm Sound**
  Local audio file to play in the browser when the map status turns to Severe.
* **Mild Alarm Sound**
  Local audio file to play in the browser when the map status turns to Mild.

## Button Descriptions

* **[Test Mail]** (or Test) : Send a test email immediately using the current configuration.
* **[Webhook Test]** : Send a test payload to the configured Webhook URL.
