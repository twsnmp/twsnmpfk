#### PKI Build CA
<div class="text-xl mb-2">
This is the settings screen for building a CA.
</div>

<div class="text-lg">

|Item|Content|
|----|---|
|Name|This is the name of the CA.I'll try to use the Subject of the CA certificate.|
|DNS name|Specify the CDP of the certificate to be issued, the OCSP address, the host name and IP address to be used for SANs for the certificate of the ACME server, separated by commas.|
|ACME Server Basic URL|This is the basic URL for the ACME server.Blanks will be automatically set from the host name.|
|CRL/OCSP/SCEP Server Basic URL|This is the basic URL for the CRL/OCSP/SCEP Server.Blanks will be automatically set from the host name.|
|CA key type|Specify the CA key type.|
|CA certificate duration|Specify the number of years the certificate is valid.|
|CRL Update Interval|Specify the CRL update interval in hours.|
|Certificate Period|Specify the period of the certificate to be issued in hours.|
|CRL/OCSP/SCEP server port number|Specify the HTTP server port number.Cannot be changed later.|
|ACME Server Port Number|Specify the ACME Server Port Number.Cannot be changed later.|
</div>


---

#### PKI Cert List
<div class="text-xl mb-2">
This is a list screen of certificates issued by the CA.
</div>

<div class="text-lg">

|Item|Content|
|----|---|
|Status|Certificate status.|
|Type|Certificate type.|
|ID|Certificate serial number.|
|Subject|A Subject for the certificate.|
|Related Node|The node where the certificate was obtained.|
|Start|The start date and time of the certificate period.|
|End|The end date and time of the certificate period.|
|Revoked|The date and time the certificate was revoked.|
</div>

>>>
<div class="text-lg">

|Item|Content|
|----|---|
|Create CSR|Displays the screen for creating a certificate request (CSR).|
|Certificate creation|Read the CSR and issue the certificate.|
|CA Initialization|Destroy CA.|
|Server Control|Displays the server control screen.|
|Renew|Update the certificate list.|
|Revokes|Revokes the selected certificate.|
|Export|Saves the selected certificate to a file.|
</div>

---

#### Create CSR
<div class="text-xl mb-2">
This is the screen for creating a CSR.
</div>

<div class="text-lg">

|Item|Content|
|----|---|
|Key type|Specify the type of key for CSR.|
|Name|Specifies a value for CN.|
|DNS Name|Specify the DNS names for the Subject Alt Name, separated by commas.|
|Organization name|Specify the organization name.It's OK to leave blank.|
|Organization Unit|Specify an organizational unit.It's OK to leave blank.|
|Country code|Specify the country code.It's OK to leave blank.|
|State/Province name|Specify the state or prefecture name.It's OK to leave blank.|
|City name|Specify the city name.It's OK to leave blank.|
</div>

---
### Server Control

<div class="text-xl mb-2">
This is a screen that controls the operation of the PKI server.
</div>

<div class="text-lg">

|Item|Content|
|----|---|
|ACME Server|Start the ACME server.|
|CRL/OCSP/SCEP Server|Start the CRL/OCSP/SCEP server.|
|ACME Server Basic URL|Specifies the basic URL that the ACME server responds to.|
|CRL Update Interval|Specify the CRL update interval in hours.|
|Certificate Period|Specify the period of the certificate to be issued in hours.|

</div>
