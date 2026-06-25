# gNMI Tool

Screen for retrieving management information from nodes using the gNMI (gRPC Network Management Interface) protocol, viewing results, and creating polling monitors.

## Settings Parameters

* **Target**
  IP address and port number to access via gNMI (e.g., `192.168.1.1:57400`).
* **Encoding**
  Encoding format for gNMI messages ("json_ietf", "json", "bytes", "proto", "ascii").
* **Path**
  gNMI sensor path to retrieve.
* **History**
  History list of previously retrieved paths. Select to reload.
* **Result**
  Table list of retrieved gNMI values.

## Button Descriptions

* **[Copy]** : Copies the retrieval results to the clipboard.
* **[Polling]** : Creates a new polling monitor based on the selected result value.
* **[Capabilities]** : Retrieves and displays the gNMI version, supported encodings, and supported YANG models of the target node.
* **[YANG Info]** : Opens the YangModels/yang repository on GitHub in a web browser.
* **[Get]** : Executes the gNMI Get request with the specified path.
* **[CSV]** : Exports the retrieval results to a CSV file.
* **[Excel]** : Exports the retrieval results to an Excel file.
* **[Help]** : Displays this help.
* **[Close]** : Closes the window.
