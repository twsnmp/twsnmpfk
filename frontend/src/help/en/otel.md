#### OpenTelemetry
<div class="text-xl mb-2">
This is the OpenTelemetry Collector screen.
Toggle in the Metrics, Traces, and Logs tabs.
</div>

---
### metric

<div class="text-lg">
A list of received metrics.

|Item|Content|
|----|---|
|Source Host|The source host for the metric.|
|Service|Name of the service associated with the metric.|
|Scope|Name of the scope associated with the metric.|
|Name|Name of the metric.|
|Type|Metric type.|
|Number of times|The number of times the metric is received.|
|First time|The date and time when the metric was first received.|
|Last|The date and time when the metric was last received.|
</div>

>>>
#### Metric (button)
<div class="text-xl">

|Item|Content|
|----|---|
|Report|View a graph for the selected metric.|
|<span style="color: red;">Delete all logs</span>|Delete all data from OpenTelemetry.|
|Update|Update information.|
</div>

---
### trace

<div class="text-lg">

This is the screen for the received trace.At the top there is a graph showing the start time, processing time, and number of spans of traces.

|Item|Content|
|----|---|
|Start date and time|The start date and time for the trace.|
|End Date and Time|The end date and time of the trace.|
|Time|Tracing processing time.|
|Trace ID|The ID that identifies the trace.|
|Source Host|The source host for the trace.|
|Service|The service name associated with the trace.|
|Span|The number of Spans in the trace.|
|Scope|Related scope for traces.|

</div>

>>>
#### Trace (button)
<div class="text-xl">

|Item|Content|
|----|---|
|Report|View the graph for the selected trace.|
|DAG|Views relationships between services from traces for the selected time range.|
|Time Range|Specifies the time range for the trace.|
|<span style="color: red;">Delete all logs</span>|Delete all data from OpenTelemetry.|
|Update|Update information.|
</div>

---
### log

<div class="text-lg">

This is a screen to search for the received OpenTelemetry logs from syslog.
At the top, you will see a graph by log level.

|Item|Content|
|----|---|
|Level|Syslog level.<br>Severe, mild, warnings and information.|
|Date and Time|The date and time when Syslog was received.|
|Host|The source host for Syslog.|
|Type|Stand for syslog facility and priority.|
|Tag|Syslog tag.Process and process ID etc.|
|Message|Syslog message.|

</div>

>>>
#### Log (button)
<div class="text-xl">

|Item|Content|
|----|---|
|<span style="color: red;">Delete all logs</span>|Delete all data from OpenTelemetry.I won't delete syslog.|
|Update|Update information.|

</div>

