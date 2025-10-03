#### Notification settings
<div class="text-xl">
This is the screen to set the notification.
</div>

<div class="text-sm">

| Items | Contents |
| ---- | ---- |
|Provider|Select from smtp/Google/Microsoft.|
|Client ID|The OAuth2 (Google/Microsoft) client ID.|
|Client Secret|The OAuth2 (Google/Microsoft) client secret.|
|Tenant Name|The OAuth2 (Microsoft) tenant name.|
| Mail server | Specify a mail server to send notification emails.<br> Host name or IP address: port number |
| Do not check the server certificate | Check when the specified mail server is Oleore certificate.|
| User | Set a user ID when authentication is required when sending an email.|
| Password | Set the password when authentication is required when sending email.|

</div>

>>>

<div class="text-sm">

| Items | Contents |
| ---- | ---- |
| From  | From email address.|
| Address | Notification email destination email address.<BR> You can specify multiple by separation of comma.|
Subject | Notification email subject.|
| Notification level | Specify the monitoring level to send notifications.|
| Notification interval | Specify the interval to check the notification.|
| Notify repair | We will also send an email when you reapir.|
| Comment execution | Run the command specified in the state parameter when the map changes.<br> $ Level is in the map.0: Severe, 1: Mild, 2: Note, 3: Normal, -1: Unknown |
| Sounds played during severe disorders | Specify the audio file to play when the state of the map is severe.|
| Sounds played during mild disability | Specify the audio file to be played when the state of the map is mild.|

</div>

---
#### Email send test

<div class="text-xl">
Click the [Test] button to send the test email with the configured content.
Click the [Webhook Test] button to test the webhook with the configured settings.
</div>

