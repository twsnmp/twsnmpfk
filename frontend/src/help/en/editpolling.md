#### Polling editing

<div class="text-xl">
Polling edit can be displayed from the button by selecting a polling list on the polling list.
</div>

<div class="text-sm">

| Items | Contents |
| ---- | ---- |
| Name | Polling name.|
| Level | Pauling disability level.|
| Type | Polling type.<br> Ping, SNMP, TCP, etc. |
| Mode | Operation mode depends on the type of polling.|
| Log mode | How to save the polling result log.|
| Parameter | Polling type and mode -dependent parameters.<br>For email polling, set the following:<br><ul><li>**Mail Server**: Hostname or IP address of the IMAP/POP3 server</li><li>**Port**: Port number of the IMAP/POP3 server (993 for IMAP, 995 for POP3 are common)</li><li>**User Name**: Username for the mailbox</li><li>**Password**: Password for the mailbox</li><li>**Protocol**: IMAP or POP3</li><li>**Secure Connection**: Whether to use SSL/TLS</li><li>**Keyword**: Check if the mail subject or body contains specific keywords (optional)</li></ul>|
| Filter | Polling type and filter condition that depends on mode.<br> Used for log search.|
| Extract pattern | This is a GROK pattern that depends on the type of polling and the mode.<br> Use when extracting data from logs.|
| Script | Java Script that determines disability and calculates variables.|
| Polling interval | Polling interval.|
| Timeout | Timeout at the time of polling.|
| Retry | This is the number of retry times when polling.|
|Failure Action|Sets the action when a failure occurs.|
|Return action|Sets the action when returning from an error.|

</div>


