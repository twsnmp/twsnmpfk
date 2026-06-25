# MIB Browser

Tool for retrieving, inspecting, and writing SNMP MIB information from a node. Correct SNMP settings must be configured on the target node beforehand.
To use external MIB files, place them in the `extmibs` directory in the application data folder.

## Settings Parameters

* **Object Name**
  The MIB object name or OID to retrieve.
* **History**
  History of previously searched MIB objects.
* **Scalar Only**
  Toggle to display only scalar objects (with index 0).
* **Raw Data**
  Toggle to display retrieved MIB values as raw data. If disabled, timestamp and other values are automatically formatted into readable text.

## Button Descriptions

* **[MIB Tree]** (Tree icon next to Object Name) : Open the MIB Tree selection dialog.
* **[Copy]** : Copy the selected rows of MIB data to the clipboard.
* **[Polling]** : Create a new polling setting using the selected MIB object.
* **[SET]** : Open the SNMP SET dialog to write a value to the selected MIB object.
* **[Get]** : Run SNMP GET or WALK for the specified object name/OID.
* **[MIB Tree]** : Display the search results structured in a tree format.
* **[CSV]** : Export search results to a CSV file.
* **[Excel]** : Export search results to an Excel file.
* **[AI Explain]** : Request an explanation of the retrieved MIB data from the AI (LLM).
* **[Help]** : Open this help document.
* **[Close]** : Close the MIB browser tool.

## Dialog Descriptions

### MIB Tree Dialog

Dialog to explore and select MIB objects. Select a node in the tree to read its description. Double-click to insert the selected object into the Object Name input.

* **Filtering the MIB Tree**
  Use the search box at the top to filter the tree. OIDs, MIB names, and descriptions are searchable.

### SET Dialog

Dialog for performing SNMP SET operations. Specify the target Object Name, Data Type, and Value.

* **[SET]** : Send the SNMP SET request.
