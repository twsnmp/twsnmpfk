# Polling

Manage active polling tasks (status monitoring and performance data collection) executed against target nodes.

## Table Columns

* **State**
  Current status of the polling task (Severe, Mild, Warn, Recovery, Normal, Unknown).
* **Node**
  The name of the node associated with this polling.
* **Name**
  The name of the polling task.
* **Level**
  Alarm level triggered on failure (Severe, Mild, Warn).
* **Type**
  The monitoring protocol/type (e.g., ping, tcp, http, dns, gNMI, syslog, mail).
* **Log**
  Log recording mode (None, normal only, error only, always, AI analysis, etc.).
* **Last Checked**
  The date and time when the polling was last executed.

## Button Descriptions

* **[Add]** : Open the template selection screen to add a new polling.
* **[Edit]** : Edit the settings of the selected polling.
* **[Copy]** : Duplicate the selected polling.
* **[Export]** : Export the selected polling configuration as a template JSON file.
* **[Report]** : Open performance reports, history graphs, and AI analysis for the selected polling.
* **[Delete Logs]** : Delete execution logs of the selected polling.
* **[Delete]** : Delete the selected polling configuration and its logs.
* **[CSV]** : Export the polling list to a CSV file.
* **[Excel]** : Export the polling list to an Excel file.
* **[Reload]** : Refresh the polling list.

## Dialog Descriptions

### Polling Template Selection (Adding Polling)

Dialog for selecting a pre-defined monitoring template to add a new polling.

* **[Add]** : Advance to the configuration editor using the selected template.
* **[Template file]** : Import and restore polling configurations from a local template JSON file.
* **[Cancel]** : Close the template selection dialog.

### Polling Report Tabs

* **Basic Info**
  Configuration details and summary of recent executions.
* **Polling Log**
  Historical list of polling logs (only available if log mode is enabled).
* **Time Chart**
  Time-series line chart of measured numerical metrics (only available if logs are recorded).
* **Histogram**
  Frequency distribution of measured metrics (only available if logs are recorded).
* **AI Analysis**
  Anomalies and predictions analyzed by AI (only available when log mode is set to AI analysis and sufficient data has been collected).
