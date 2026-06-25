# PKI (Before CA Build)

Configuration screen for initializing and building a new Certificate Authority (CA).

## CA Build Parameters

* **Name**
  Name of the CA, used as the Subject for the self-signed CA certificate.
* **DNS name**
  Comma-separated list of hostnames or IP addresses used for CDP (CRL Distribution Point), OCSP address, and SANs (Subject Alternative Names) of the ACME server certificate.
* **ACME Server Basic URL**
  Base URL of the ACME server. If blank, it is automatically resolved from the hostname.
* **CRL/OCSP/SCEP Server Basic URL**
  Base URL of the CRL/OCSP/SCEP server. If blank, it is automatically resolved from the hostname.
* **CA key type**
  Algorithm and length of the CA private/public key pair (e.g., RSA, ECDSA).
* **CA certificate duration**
  Validity period of the CA certificate in years.
* **CRL Update Interval**
  Interval in hours to update the Certificate Revocation List (CRL).
* **Certificate Period**
  Validity period in hours for certificates issued by this CA.
* **CRL/OCSP/SCEP server port number**
  HTTP port number for the CRL/OCSP/SCEP server. Cannot be changed after building.
* **ACME Server Port Number**
  Port number for the ACME server. Cannot be changed after building.

---

# PKI (After CA Build)

Management screen for issued digital certificates and manual signing operations after the CA has been established.

## Certificate List Columns

* **Status**
  Current status of the certificate (Valid, Expired, Revoked).
* **Type**
  Type of certificate (e.g., Client, Server).
* **ID**
  Serial number of the certificate.
* **Subject**
  Subject details of the certificate owner.
* **Related Node**
  The node name associated with the certificate.
* **Start**
  Start date and time of the certificate's validity.
* **End**
  Expiration date and time of the certificate.
* **Revoked**
  Date and time when the certificate was revoked.

## Button Descriptions

* **[Create CSR]** : Open the Certificate Signing Request (CSR) creation dialog.
* **[Certificate creation]** : Read and sign a CSR to issue a new certificate.
* **[CA Initialization]** : Delete and destroy the current CA configuration and restore the PKI to the unbuilt state.
* **[Server Control]** : Open the server control dialog to toggle and adjust PKI services.
* **[Revokes]** : Revoke the selected certificate and add it to the revocation list.
* **[Export]** : Export and download the selected certificate file.
* **[Renew]** : Refresh the certificate list.

## CSR Generation Parameters

* **Key type**
  Key type and length to generate.
* **Name**
  Common Name (CN) for the certificate.
* **DNS Name**
  Comma-separated DNS names for Subject Alt Names (SAN).
* **Organization name**
  Organization (O) name (optional).
* **Organization Unit**
  Organizational Unit (OU) name (optional).
* **Country code**
  Country (C) code (optional).
* **State/Province name**
  State or province (ST) name (optional).
* **City name**
  Locality (L) name (optional).

## Server Control Parameters

* **ACME Server**
  Toggle to enable/disable the built-in ACME server.
* **CRL/OCSP/SCEP Server**
  Toggle to enable/disable the built-in CRL/OCSP/SCEP server.
* **ACME Server Basic URL**
  Response base URL of the ACME server.
* **CRL Update Interval**
  Interval in hours to update the CRL.
* **Certificate Period**
  Validity period in hours for newly issued certificates.
