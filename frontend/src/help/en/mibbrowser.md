#### MIB browser

<div class="text-xl mb-2">
This is a screen to get MIB information of SNMP from the node.
It is necessary to set SNMP access information in the node setting.
If you want to use MIB other than built -in, save the MIB file to the extmibs of the data folder.
</div>

<div class="text-lg">

| Items | Contents |
| ---- | ---- |
| Object name | Specify the object name of the MIB you want to get.<br> You can choose from the MIB tree.Example: System |
|<MIB Tree> button|Displays the MIB tree for object selection. |
| History | It is the history of the object name obtained so far.You can select and get it again.|
| Results | Acquired MIB information.In the case of MIB in a table format, it is automatically displayed in a table format.|
|Copy|Click to copy the selected MIB to the board. |
|Polling|Create a poll from the selected MIB. |
|SET|Displays a dialog to perform SET on the selected MIB. |
|Scalar only|Display only objects with index 0. |
| Raw data | Displays the acquired MIB information without converting it.<BR> In the case of off, convert the time data to an easy -to -understand display.|
| Get | Get MIB information.|
|MIB Tree|Displays the acquired MIB information in a tree view. |
| CSV | Export the obtained MIB information of the CSV file.|
| Excel | Export the acquired MIB information of the Excel file.|

</div>

---
#### MIB tree
<div class="text-xl mb-2">
This is a screen for selecting the obtained MIB object name.
Open the tree and click the object name to see the explanation.
Double click to select.
</div>

---
#### Set dialog

<div class="text-xl">
This is the screen for executing SNMP Set. Specify the object name, type, and value and press the <SET> button.
Click to send a Set request.

</div>