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
* **[Get]** : Run SNMP GET or WALK for the specified object name/OID (can also be executed with `Cmd/Ctrl + Enter`).
* **[MIB Tree]** : Display the search results structured in a tree format.
* **[CSV]** : Export search results to a CSV file.
* **[Excel]** : Export search results to an Excel file.
* **[AI Explain]** : Request an explanation of the retrieved MIB data from the AI (LLM).
* **[Help]** : Open this help document.
* **[Close]** : Close the MIB browser tool (can also be closed with `Esc`).

## Dialog Descriptions

### MIB Tree Dialog

Dialog to explore and select MIB objects. Single-clicking any node selects and immediately applies the object name to the input field. The tree expansion state is preserved, and hovering over a node displays a tooltip with the MIB object description.

* **Filtering the MIB Tree**
  Use the search box at the top to filter the tree. OIDs, MIB names, and descriptions are searchable.

### SET Dialog

Dialog for performing SNMP SET operations. Specify the target Object Name, Data Type, and Value.

* **[SET]** : Send the SNMP SET request.
