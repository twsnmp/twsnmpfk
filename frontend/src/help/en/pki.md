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

* **ACME/EST Server**
  Toggle to enable/disable the built-in ACME and EST (RFC 7030 / Enrollment over Secure Transport) servers.
* **CRL/OCSP/SCEP Server**
  Toggle to enable/disable the built-in CRL/OCSP/SCEP server.
* **ACME Server Basic URL**
  Response base URL of the ACME and EST servers.
* **CRL Update Interval**
  Interval in hours to update the CRL.
* **Certificate Period**
  Validity period in hours for newly issued certificates.

## ACME Server Supported Challenges and Device Attestation

The built-in ACME server supports the following identifiers and challenge validation methods. To issue a certificate, the target host or device must be registered as a node on the TWSNMP FK map.

* **DNS Identifier (`dns`)**
  * Supports `http-01`, `tls-alpn-01`, and `dns-01`.
  * Supports wildcard domain certificates (`*.example.com`) via `dns-01` validation and base-domain node matching.
* **IP Address Identifier (`ip` / RFC 8738)**
  * Supports IPv4 and IPv6 address validation via `http-01` and `tls-alpn-01` (`.in-addr.arpa.` / `.ip6.arpa.`).
* **Device Attestation (`permanent-identifier` / `hardware-module`, `device-attest-01`)**
  * Supports Apple Managed Device Attestation (`fmt: "apple"`, iOS 16 / iPadOS 16 / macOS 14+) and YubiKey PIV / Smallstep (`fmt: "step"`) attestation formats.
  * Issued certificates include the RFC 4043 `PermanentIdentifier` (OID `1.3.6.1.5.5.7.8.3`) in the Subject Alternative Name (SAN) `otherName` extension.
  * **Node Registration** : Register the device's Serial Number or UDID (`PermanentIdentifier`) in the node's **Name**, **Description**, **Public Key**, or **Memo** so TWSNMP FK can match and authorize the CSR.

## EST (RFC 7030) Server Features

When the ACME server is enabled, RFC 7030 compliant EST endpoints are automatically served over the same HTTPS port.

* **Supported Endpoints (supports standard and optional path label)**:
  * `GET /.well-known/est/cacerts` or `GET /.well-known/est/:label/cacerts`
    * Returns the Root CA certificate in a Base64-encoded PKCS#7 (`certs-only`) structure.
  * `POST /.well-known/est/simpleenroll` or `POST /.well-known/est/:label/simpleenroll`
    * Accepts a PKCS#10 CSR, authenticates the client and matches the target node, then returns the signed certificate in Base64-encoded PKCS#7 format.
  * `POST /.well-known/est/simplereenroll` or `POST /.well-known/est/:label/simplereenroll`
    * Renews an existing certificate using client certificate authentication or Basic auth.
* **Authentication Requirements (mTLS or HTTP Basic Auth required)**:
  * **mTLS (Client Certificate Auth)**: If the client connects using a valid, unrevoked client certificate issued by this CA, certificate issuance or renewal is granted.
  * **HTTP Basic Auth**: If no client certificate is presented, HTTP Basic Authentication using the matching node's password (`node.Password`) and username (if configured on `node.User`) is required (returns `401 Unauthorized` if unauthenticated).
* **Node Matching**:
  * The SANs (DNS names, IP addresses, RFC 4043 PermanentIdentifier), Subject CN, or remote IP address in the CSR must correspond to an active node registered on the map.

