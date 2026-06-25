# OpenTelemetry

Inspect and analyze metrics, traces, and logs collected via OpenTelemetry. Switch between the Metrics, Traces, and Logs tabs.

## Metrics Columns

* **Host**
  The source host that sent the metric.
* **Service**
  The name of the service that emitted the metric.
* **Scope**
  The measurement scope (e.g., instrumentation library name).
* **Name**
  The name of the metric.
* **Type**
  The data type of the metric (e.g., Gauge, Sum, Histogram).
* **Count**
  Total count of received metric data points.
* **First**
  Date and time when the metric was first received.
* **Last**
  Date and time when the metric was last received.

## Metrics Buttons

* **[Report]** : View a performance graph of the selected metric.
* **[Delete All]** : Delete all OpenTelemetry data from the database.
* **[Reload]** : Refresh the metrics list.

## Traces Columns

Distributed tracing data.

* **Start**
  Start date and time of the trace.
* **End**
  End date and time of the trace.
* **Duration**
  Total duration of the trace in milliseconds.
* **Trace ID**
  Unique identifier for the trace.
* **Host**
  List of hosts involved in this trace.
* **Service**
  List of services involved in this trace.
* **Span**
  Number of spans contained in the trace.
* **Scope**
  Scope associated with the trace.

## Traces Buttons

* **[Report]** : Open the waterfall trace/span visualizer for the selected trace.
* **[DAG]** : Generate and display a Service Dependency Graph (Directed Acyclic Graph) for the selected trace data.
* **[Delete All]** : Delete all OpenTelemetry data.
* **[Reload]** : Refresh the trace data and chart.

## Logs Columns

Logs received via OpenTelemetry.

* **Level**
  Severity level of the log (Severe, Mild, Warn, Info, etc.).
* **Time**
  Date and time when the log occurred.
* **Host**
  The source host of the log.
* **Type**
  Attributes representing facility and priority.
* **Tag**
  Process tags or categories.
* **Message**
  The log message text.

## Logs Buttons

* **[Delete All]** : Delete all OpenTelemetry logs (does not delete regular Syslog logs).
* **[Reload]** : Refresh the logs table and level distribution chart.
