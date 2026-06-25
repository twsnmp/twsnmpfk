# System Resources

Monitor the system resources of the host machine and the performance of the TWSNMP FK process itself. Chronological resource utilization charts are displayed at the top.

## Table Columns

* **Time**
  Date and time when the system metrics were recorded.
* **CPU**
  Host CPU usage rate (%).
* **Memory**
  Host memory usage rate (%).
* **My CPU**
  CPU usage rate of the TWSNMP FK process (%).
* **My Memory**
  Memory usage rate of the TWSNMP FK process (%).
* **Swap**
  Host swap space utilization (%).
* **Disk**
  Disk space utilization where the data folder is located (%).
* **Load**
  Host CPU load average.
* **Net**
  Network transfer speed of the host network interface.
* **Conn**
  Total number of active TCP connections.
* **Proc**
  Total number of active processes on the host.
* **Go Goroutines**
  Total number of Go Goroutines running inside the application.
* **Heap**
  Heap memory allocation size of the application.
* **Sys**
  Virtual memory size obtained from the OS.
* **DB Size**
  Physical file size of the embedded database.

## Button Descriptions

* **[Size Prediction]** : Display a simulation chart predicting database size growth and disk usage over the next 12 months.
* **[Backup]** : Generate a backup archive file of the system database.
* **[Reload]** : Refresh system resource metrics.
