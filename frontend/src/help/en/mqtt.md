# MQTT

List of received MQTT topics and their statistical statuses.

## Table Columns

* **Status**
  Reception status. Warnings are displayed if no messages are received for over 24 hours, and minor error states for over 7 days.
* **Client ID**
  The client identifier of the sender.
* **Remote**
  The source IP address of the MQTT client.
* **Topic**
  Name of the received MQTT topic.
* **Count**
  Total number of times the topic has been received.
* **Bytes**
  Total bytes of data received under this topic.
* **First**
  Date and time when the topic was first received.
* **Last Checked**
  Date and time when the topic was last received.

## Button Descriptions

* **[Delete]** : Delete statistical data for the selected topics.
* **[Delete All]** : Delete all MQTT statistics.
* **[Copy]** : Copy selected topic names to the clipboard.
* **[Make Polling]** : Open the polling configuration screen to monitor the selected MQTT topic.
* **[Reload]** : Refresh the MQTT stats list.
