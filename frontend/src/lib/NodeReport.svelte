<script lang="ts">
  import {
    Modal,
    GradientButton,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Spinner,
    Toggle,
    Button,
    Select,
  } from "flowbite-svelte";
  import { tick, createEventDispatcher } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNode,
    GetPollings,
    GetVPanelPorts,
    GetVPanelPowerInfo,
    GetEventLogs,
    GetHostResource,
    GetDefaultPolling,
    GetNodeMemo,
    SaveNodeMemo,
    GetNodeImage,
    SaveNodeImage,
    SelectFile,
    GetImage,
    GetMapConf,
    LLMExplainNodeReport,
  } from "../../wailsjs/go/main/App";
  import ReportAIDialog from "./ReportAIDialog.svelte";
  import {
    getIcon,
    getStateColor,
    getStateName,
    getTableLang,
    renderTime,
    renderState,
    renderBytes,
    renderCount,
    renderSpeed,
    renderHrSystemName,
    renderDuration,
    renderSLA,
  } from "./common";
  import {
    calcNodeDowntimeAndSLA,
    showNodeDowntimeChart,
  } from "./chart/eventlog";
  import { deleteVPanel, initVPanel, setVPanel } from "./vpanel";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { showHrBarChart, showHrSummary } from "./chart/hostResource";
  import { _ } from "svelte-i18n";
  import { copyText } from "svelte-copy";

  export let show: boolean = false;
  export let id = "";
  let node: any;
  let activeTab = "basic";
  let hasAI = false;
  let showAIReport = false;

  let selectedPortCount = 0;
  let selectedHrSystemCount = 0;
  let selectedhrStorageCount = 0;
  let selectedHrProcessCount = 0;
  let showPolling = false;
  let physicalPort = true;
  let showVPanelBtn = false;
  let chart: any = undefined;
  let chartMem: any = undefined;
  let logTable: any = undefined;
  let downtimeTable: any = undefined;
  let downtimeStats = {
    totalIncidents: 0,
    ongoingIncidents: 0,
    totalDowntimeSec: 0,
    maxDowntimeSec: 0,
    mttrSec: 0,
    overallSLA: 100,
    pollingStats: [] as any[],
    incidents: [] as any[],
  };


  const clearSelectedCount = () => {
    selectedPortCount = 0;
    selectedHrSystemCount = 0;
    selectedhrStorageCount = 0;
    selectedHrProcessCount = 0;
    showVPanelBtn = false;
    showSaveMemoBtn = false;
    deleteVPanel();
    if (chart) {
      try {
        chart.dispose();
      } catch (e) {
        // ignore
      }
      chart = undefined;
    }
    if (chartMem) {
      try {
        chartMem.dispose();
      } catch (e) {
        // ignore
      }
      chartMem = undefined;
    }
    if (logTable) {
      try {
        logTable.destroy();
      } catch (e) {
        // ignore
      }
      logTable = undefined;
    }
    if (downtimeTable) {
      try {
        downtimeTable.destroy();
      } catch (e) {
        // ignore
      }
      downtimeTable = undefined;
    }
    if (portTable) {
      try {
        portTable.destroy();
      } catch (e) {
        // ignore
      }
      portTable = undefined;
    }
    if (hrSystemTable) {
      try {
        hrSystemTable.destroy();
      } catch (e) {
        // ignore
      }
      hrSystemTable = undefined;
    }
    if (hrStorageTable) {
      try {
        hrStorageTable.destroy();
      } catch (e) {
        // ignore
      }
      hrStorageTable = undefined;
    }
    if (hrDeviceTable) {
      try {
        hrDeviceTable.destroy();
      } catch (e) {
        // ignore
      }
      hrDeviceTable = undefined;
    }
    if (hrFileSystemTable) {
      try {
        hrFileSystemTable.destroy();
      } catch (e) {
        // ignore
      }
      hrFileSystemTable = undefined;
    }
    if (hrProcessTable) {
      try {
        hrProcessTable.destroy();
      } catch (e) {
        // ignore
      }
      hrProcessTable = undefined;
    }
  };

  const safeInit = (selector: string, initFn: () => void) => {
    let retries = 0;
    const checkAndInit = () => {
      if (document.querySelector(selector)) {
        try {
          initFn();
        } catch (e) {
          console.error("Failed to initialize: " + selector, e);
        }
      } else if (retries < 10) {
        retries++;
        setTimeout(checkAndInit, 50);
      }
    };
    checkAndInit();
  };

  const showLog = async () => {
    clearSelectedCount();
    const logs = await GetEventLogs({
      NodeID:id,
      Start: "",
      End: "",
      Event:"",
      NodeName: "",
      EventType:"",
      Level:0,
    });
    if (activeTab !== "log") return;
    safeInit("#nodeReportLogTable", () => {
      logTable = new DataTable("#nodeReportLogTable", {
        destroy: true,
        data: logs,
        language: getTableLang(),
        order: [[1, "desc"]],
        columns: [
          {
            data: "Level",
            title: $_("NodeReport.Level"),
            width: "15%",
            render: renderState,
          },
          {
            data: "Time",
            title: $_("NodeReport.Time"),
            width: "20%",
            render: renderTime,
          },
          {
            data: "Type",
            title: $_("NodeReport.Type"),
            width: "15%",
          },
          {
            data: "Event",
            title: $_("NodeReport.Event"),
            width: "50%",
          },
        ],
      });
    });
  };

  const showDowntime = async () => {
    clearSelectedCount();
    const [logs, pollings] = await Promise.all([
      GetEventLogs({
        NodeID: id,
        Start: "",
        End: "",
        Event: "",
        NodeName: "",
        EventType: "",
        Level: 0,
      }),
      GetPollings(id),
    ]);
    if (activeTab !== "downtime") return;
    downtimeStats = calcNodeDowntimeAndSLA(logs, pollings);
    await tick();
    chart = showNodeDowntimeChart("nodeDowntimeChart", logs, pollings);
    await initDowntimeTable();
  };

  const initDowntimeTable = async () => {
    await tick();
    if (!document.getElementById("nodeDowntimeTable")) {
      return;
    }

    const sortedStats = [...downtimeStats.pollingStats].sort(
      (a, b) => b.totalDowntimeSec - a.totalDowntimeSec
    );

    const ongoingLabel = $_("EventLogReport.Ongoing") || "障害発生中";

    const tableData = sortedStats.map((p: any) => {
      const slaStr = renderSLA(p.sla, "");
      const totalDurationStr = renderDuration(p.totalDowntimeSec, "");
      const maxDurationStr = renderDuration(p.maxDowntimeSec, "");
      const statusHtml = p.ongoing
        ? `<span class="px-2 py-0.5 rounded text-xs font-semibold bg-red-900 text-red-200">${ongoingLabel}</span>`
        : renderState(p.currentLevel);

      return [
        p.event,
        slaStr,
        p.count,
        totalDurationStr,
        maxDurationStr,
        statusHtml,
      ];
    });

    downtimeTable = new DataTable("#nodeDowntimeTable", {
      destroy: true,
      paging: false,
      searching: true,
      info: false,
      scrollY: "300px",
      scrollCollapse: true,
      language: getTableLang(),
      data: tableData,
      order: [[3, "desc"]],
      columns: [
        { title: $_("NodeReport.PollingName") || "ポーリング名" },
        { title: $_("EventLogReport.OverallSLA") || "SLA (Availability)" },
        { title: $_("EventLogReport.TotalIncidents") || "Incidents" },
        { title: $_("EventLogReport.TotalDowntime") || "Total Downtime" },
        { title: $_("EventLogReport.MaxDowntime") || "Max Downtime" },
        { title: $_("EventLogReport.Status") || "Status" },
      ],
    });
  };

  let portTable: any = undefined;

  const showPortTable = (p: any) => {
    if (!document.getElementById("portTable")) {
      return;
    }
    portTable = new DataTable("#portTable", {
      destroy: true,
      stateSave: true,
      paging: false,
      searching: false,
      info: false,
      select: {
        style: "single",
      },
      scrollY: "20vh",
      data: p,
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
        {
          data: "Index",
          title: "No.",
          width: "5%",
          className: "dt-body-right",
        },
        {
          data: "State",
          title: $_("NodePolling.State"),
          width: "5%",
          render: renderState,
        },
        {
          data: "Name",
          title: $_("NodeReport.Name"),
          width: "15%",
        },
        {
          data: "Type",
          title: $_("NodeReport.Type"),
          width: "5%",
        },
        {
          data: "MAC",
          title: $_("NodeReport.MACAddress"),
          width: "10%",
        },
        {
          data: "Speed",
          title: $_("NodeReport.Speed"),
          width: "8%",
          render: renderSpeed,
          className: "dt-body-right",
        },
        {
          data: "OutPacktes",
          title: $_("NodeReport.OutPacktes"),
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "OutBytes",
          title: $_("NodeReport.OutBytes"),
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          data: "InPacktes",
          title: $_("NodeReport.InPacktes"),
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "InBytes",
          title: $_("NodeReport.InBytes"),
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          data: "LastChanged",
          title: $_('NodeReport.LastChanged'),
          width: "12%",
          render: renderTime,
        },
      ],
    });
    portTable.on("select", () => {
      selectedPortCount = portTable.rows({ selected: true }).count();
    });
    portTable.on("deselect", () => {
      selectedPortCount = portTable.rows({ selected: true }).count();
    });
  };

  let ports: any;
  let power: any;
  let waitVPanel = false;
  let rotateVPanel = false;
  let vpanelZoom = 1.0;
  let vpanelPortWrap = 16;
  let pVPanel: any = [];

  const showVPanel = async () => {
    clearSelectedCount();
    showVPanelBtn = true;
    if (!ports) {
      waitVPanel = true;
      const p = await GetVPanelPorts(id);
      const pow = await GetVPanelPowerInfo(id);
      if (activeTab !== "panel") return;
      ports = p;
      power = pow;
      waitVPanel = false;
    }
    if (activeTab !== "panel") return;
    const portsList = ports || [];
    pVPanel = physicalPort ? portsList.filter((e: any) => e.Type == 6) : portsList;

    if (ports && ports.length > 0) {
      safeInit("#vpanel", () => {
        initVPanel("vpanel");
      });

      safeInit("#portTable", () => {
        showPortTable(pVPanel);
      });
    }
  };

  $: setVPanel(pVPanel, power, rotateVPanel, vpanelZoom, vpanelPortWrap);

  const renderStatus = (s: any) => {
    switch (s) {
      case "Running":
        return (
          `<span class="text-blue-700 dark:text-blue-400">` +
          $_("NodeReport.Running") +
          `</span>`
        );
      case "Runnable":
        return (
          `<span class="text-blue-900 dark:text-blue-300">` +
          $_("NodeReport.Runnable") +
          `</span>`
        );
      case "Testing":
        return $_("NodeReport.Testing");
      case "NotRunnable":
        return $_("NodeReport.NotRunnable");
      case "Invalid":
      case "Down":
        return (
          `<span class="text-red-800 dark:text-red-400">` +
          $_("NodeReport.Down") +
          `</span>`
        );
    }
    return $_("NodeReport.Unknown");
  };

  const renderRate = (r: any) => {
    if (r < 80.0) {
      return `<span class="text-blue-700 dark:text-blue-400">${r.toFixed(2)}</span>`;
    } else if (r < 90.0) {
      return `<span class="text-yellow-700 dark:text-yellow-400">${r.toFixed(2)}</span>`;
    }
    return `<span class="text-red-700 dark:text-red-400">${r.toFixed(2)}</span>`;
  };

  const renderStorageType = (t: any) => {
    switch (t) {
      case "hrStorageCompactDisc":
        return $_("NodeReport.CDDrive");
      case "hrStorageRemovableDisk":
        return $_("NodeReport.RemovableDisk");
      case "hrStorageFloppyDisk":
        return $_("NodeReport.FloppyDIsk");
      case "hrStorageRamDisk":
        return $_("NodeReport.RamDisk");
      case "hrStorageFlashMemory":
        return $_("NodeReport.FlashMemory");
      case "hrStorageNetworkDisk":
        return $_("NodeReport.NetworkDIsk");
      case "hrStorageFixedDisk":
        return $_("NodeReport.FixedDisk");
      case "hrStorageVirtualMemory":
        return $_("NodeReport.VM");
      case "hrStorageRam":
        return $_("NodeReport.RAM");
    }
    return $_("NodeReport.Other");
  };

  const renderDeviceType = (t: any) => {
    return t.replace("hrDevice", "");
  };

  const renderFSType = (t: any) => {
    return t.replace("hrFS", "");
  };

  const renderTrueFalse = (v: any) => {
    if (v === 1) {
      return "Yes";
    }
    return "No";
  };

  const renderAccess = (v: any) => {
    if (v === 1) {
      return "R/W";
    }
    return "Read Only";
  };

  const renderCPU = (v: any) => {
    return (v / 100).toFixed(2);
  };

  const renderMem = (v: any, t: any) => {
    return renderBytes(v * 1024, t);
  };

  let hostResource: any = undefined;
  let hrSystemTable: any = undefined;
  let hrStorageTable: any = undefined;
  let hrDeviceTable:any = undefined;
  let hrFileSystemTable:any = undefined;
  let hrProcessTable: any = undefined;
  let waitHr = false;

  const showHrSystem = async () => {
    if (!hostResource) {
      waitHr = true;
      const res = await GetHostResource(id);
      if (activeTab !== "hostinfo") return;
      hostResource = res;
      waitHr = false;
    }
    if (!hostResource) {
      return;
    }
    if (activeTab !== "hostinfo") return;
    clearSelectedCount();
    safeInit("#hrSystemTable", () => {
      hrSystemTable = new DataTable("#hrSystemTable", {
        destroy: true,
        paging: false,
        searching: false,
        info: false,
        data: hostResource.System || [],
        language: getTableLang(),
        order: [[0, "asc"]],
        select: {
          style: "single",
        },
        columns: [
          {
            data: "Index",
            title: "No",
            width: "10%",
          },
          {
            data: "Key",
            title: $_("NodeReport.Name"),
            width: "30%",
            render: renderHrSystemName,
          },
          {
            data: "Value",
            title: $_("MIBBrowser.Value"),
            width: "60%",
          },
        ],
      });
      hrSystemTable.on("select", () => {
        selectedHrSystemCount = hrSystemTable.rows({ selected: true }).count();
      });
      hrSystemTable.on("deselect", () => {
        selectedHrSystemCount = hrSystemTable.rows({ selected: true }).count();
      });
      showHrSummaryChart();
    });
  };

  const showHrStorage = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    if (activeTab !== "storage") return;
    safeInit("#hrStorageTable", () => {
      hrStorageTable = new DataTable("#hrStorageTable", {
        destroy: true,
        paging: false,
        searching: false,
        info: false,
        scrollY: "25vh",
        data: hostResource.Storage || [],
        language: getTableLang(),
        order: [[4, "desc"]],
        select: {
          style: "single",
        },
        columns: [
          {
            title: $_("NodeReport.Type"),
            data: "Type",
            width: "20%",
            render: renderStorageType,
          },
          { title: $_("NodeReport.Descr"), data: "Descr", width: "40%" },
          {
            title: $_("NodeReport.Size"),
            data: "Size",
            width: "10%",
            render: renderBytes,
            className: "dt-body-right",
          },
          {
            title: $_("NodeReport.Used"),
            data: "Used",
            width: "10%",
            render: renderBytes,
            className: "dt-body-right",
          },
          {
            title: $_("NodeReport.Rate"),
            data: "Rate",
            width: "10%",
            render: renderRate,
            className: "dt-body-right",
          },
          {
            title: $_("NodeReport.Unit"),
            data: "Unit",
            width: "10%",
            className: "dt-body-right",
          },
        ],
      });
      hrStorageTable.on("select", () => {
        selectedhrStorageCount = hrStorageTable.rows({ selected: true }).count();
      });
      hrStorageTable.on("deselect", () => {
        selectedhrStorageCount = hrStorageTable.rows({ selected: true }).count();
      });
      showHrStorageChart();
    });
  };

  const showHrDevice = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    if (activeTab !== "device") return;
    safeInit("#hrDeviceTable", () => {
      hrDeviceTable = new DataTable("#hrDeviceTable", {
        destroy: true,
        paging: false,
        searching: false,
        info: false,
        scrollY: "55vh",
        data: hostResource.Device || [],
        language: getTableLang(),
        order: [[0, "asc"]],
        columns: [
          {
            title: $_("NodeReport.Status"),
            data: "Status",
            width: "10%",
            render: renderStatus,
          },
          { title: $_("NodeReport.Index"), data: "Index", width: "10%" },
          {
            title: $_("NodeReport.Type"),
            data: "Type",
            width: "30%",
            render: renderDeviceType,
          },
          { title: $_("NodeReport.Descr"), data: "Descr", width: "40%" },
          { title: $_("NodeReport.Errors"), data: "Errors", width: "10%" },
        ],
      });
    });
  };

  const showHrFileSystem = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    if (activeTab !== "filesystem") return;
    safeInit("#hrFileSystemTable", () => {
      hrFileSystemTable = new DataTable("#hrFileSystemTable", {
        destroy: true,
        paging: false,
        searching: false,
        info: false,
        scrollY: "55vh",
        data: hostResource.FileSystem || [],
        language: getTableLang(),
        order: [[0, "asc"]],
        columns: [
          { title: $_("NodeReport.Mount"), data: "Mount", width: "30%" },
          { title: $_("NodeReport.Remote"), data: "Remote", width: "30%" },
          {
            title: $_("NodePolling.Type"),
            data: "Type",
            width: "20%",
            render: renderFSType,
          },
          {
            title: $_("NodeReport.Access"),
            data: "Access",
            width: "10%",
            render: renderAccess,
          },
          {
            title: $_("NodeReport.Bootable"),
            data: "Bootable",
            width: "10%",
            render: renderTrueFalse,
          },
        ],
      });
    });
  };

  const showHrProcess = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    if (activeTab !== "process") return;
    safeInit("#hrProcessTable", () => {
      hrProcessTable = new DataTable("#hrProcessTable", {
        destroy: true,
        paging: false,
        searching: false,
        info: false,
        scrollY: "25vh",
        data: hostResource.Process || [],
        language: getTableLang(),
        order: [[1, "asc"]],
        select: {
          style: "single",
        },
        columns: [
          {
            title: $_("NodeReport.Status"),
            data: "Status",
            width: "10%",
            render: renderStatus,
          },
          { title: "PID", data: "PID", width: "10%" },
          { title: $_("NodeReport.Type"), data: "Type", width: "10%" },
          {
            title: $_("NodeReport.Name"),
            data: "Name",
            width: "15%",
          },
          { title: $_("NodeReport.Path"), data: "Path", width: "15%" },
          { title: $_("NodeReport.Param"), data: "Param", width: "20%" },
          {
            title: "CPU",
            data: "CPU",
            width: "10%",
            render: renderCPU,
            className: "dt-body-right",
          },
          {
            title: "Mem",
            data: "Mem",
            width: "10%",
            render: renderMem,
            className: "dt-body-right",
          },
        ],
      });
      hrProcessTable.on("select", () => {
        selectedHrProcessCount = hrProcessTable.rows({ selected: true }).count();
      });
      hrProcessTable.on("deselect", () => {
        selectedHrProcessCount = hrProcessTable.rows({ selected: true }).count();
      });
      showHrProcChart(true);
      showHrProcChart(false);
    });
  };

  const showHrSummaryChart = () => {
    if (!hostResource) {
      return;
    }
    safeInit("#hrSummaryChart", () => {
      const data = {
        CPU: 0,
        Mem: 0,
        VM: 0,
      };
      (hostResource.System || []).forEach((e: any) => {
        if (e.Key == "hrProcessorLoad") {
          data.CPU = Number(e.Value);
        }
      });
      (hostResource.Storage || []).forEach((e: any) => {
        if (e.Type.includes("hrStorageRam")) {
          data.Mem = e.Rate;
        }
        if (
          e.Type.includes("hrStorageVirtualMemory") &&
          !e.Descr.includes("wap")
        ) {
          data.VM = e.Rate;
        }
      });
      chart = showHrSummary("hrSummaryChart", data);
    });
  };

  const showHrStorageChart = () => {
    if (!hostResource) {
      return;
    }
    safeInit("#hrStorageChart", () => {
      const list: any = [];
      (hostResource.Storage || []).forEach((e: any) => {
        const t = renderStorageType(e.Type);
        if (!t.includes($_("NodeReport.Other"))) {
          list.unshift({
            Name: e.Descr + "(" + t + ")",
            Value: e.Rate,
          });
        }
      });
      chart = showHrBarChart(
        "hrStorageChart",
        $_("NodeReport.StorageUsgae"),
        "%",
        list
      );
    });
  };

  const showHrProcChart = (bCPU: boolean) => {
    if (!hostResource) {
      return;
    }
    const elementId = bCPU ? "hrProcessCPUChart" : "hrProcessMemChart";
    safeInit("#" + elementId, () => {
      let max = 0;
      const list: any = [];
      (hostResource.Process || []).forEach((e: any) => {
        const v = bCPU ? e.CPU / 100.0 : e.Mem * 1024;
        if (max < v) {
          max = v;
        }
        list.push({
          Name: e.Name + "(" + e.PID + ")",
          Value: v,
        });
      });
      list.sort((a: any, b: any) => {
        if (a.Value < b.Value) return -1;
        if (a.Value > b.Value) return 1;
        return 0;
      });
      while (list.length > 20) {
        list.shift();
      }
      const c = showHrBarChart(
        elementId,
        bCPU ? $_("NodeReport.CPUUsage") : $_("NodeReport.MemUsage"),
        bCPU ? $_("NodeReport.Sec") : "Bytes",
        list,
        max
      );
      if (bCPU) {
        chart = c;
      } else {
        chartMem = c;
      }
    });
  };

  const close = () => {
    deleteVPanel();
    show = false;
  };

  let pollingTmp: any = undefined;

  const watchPortState = async () => {
    const d = portTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(node.ID);
    pollingTmp.Name = d[0].Name + "I/F Status";
    pollingTmp.Type = "snmp";
    pollingTmp.Mode = "ifOperStatus";
    pollingTmp.Level = "low";
    pollingTmp.Params = d[0].Index;
    showPolling = true;
  };

  const watchPortTraffic = async () => {
    const d = portTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(node.ID);
    pollingTmp.Name = d[0].Name + "Traffic";
    pollingTmp.Type = "snmp";
    pollingTmp.Mode = "traffic";
    pollingTmp.Params = d[0].Index;
    pollingTmp.Level = "info";
    showPolling = true;
  };

  const canWacthHrSystem = () => {
    const d = hrSystemTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return false;
    }
    switch (d[0].Key) {
      case "hrSystemUptime":
      case "hrSystemDate":
      case "hrSystemProcesses":
      case "hrProcessorLoad":
        return true;
    }
    return false;
  };

  const watchHrSystem = async () => {
    const d = hrSystemTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(node.ID);
    switch (d[0].Key) {
      case "hrSystemUptime":
        pollingTmp = await GetDefaultPolling(node.ID);
        pollingTmp.Name = "SNMP restart";
        pollingTmp.Type = "snmp";
        pollingTmp.Mode = "sysUpTime";
        pollingTmp.Level = "low";
        showPolling = true;
        break;
      case "hrSystemDate":
        pollingTmp = await GetDefaultPolling(node.ID);
        pollingTmp.Name = "System date";
        pollingTmp.Type = "snmp";
        pollingTmp.Mode = "hrSystemDate";
        pollingTmp.Script = "diff < 1";
        pollingTmp.Level = "warn";
        showPolling = true;
        break;
      case "hrSystemProcesses":
        pollingTmp = await GetDefaultPolling(node.ID);
        pollingTmp.Name = "Process count";
        pollingTmp.Type = "snmp";
        pollingTmp.Mode = "get";
        pollingTmp.Params = "hrSystemProcesses.0";
        pollingTmp.Level = "info";
        showPolling = true;
        break;
      case "hrProcessorLoad":
        pollingTmp = await GetDefaultPolling(node.ID);
        pollingTmp.Name = "CPU Usage";
        pollingTmp.Type = "snmp";
        pollingTmp.Mode = "stats";
        pollingTmp.Params = "hrProcessorLoad";
        pollingTmp.Level = "low";
        pollingTmp.Script = "avg < 95.0";
        showPolling = true;
        break;
    }
  };

  const watchHrStorage = async () => {
    const d = hrStorageTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(node.ID);
    pollingTmp.Name = d[0].Descr + "Usage";
    pollingTmp.Type = "snmp";
    pollingTmp.Mode = "get";
    pollingTmp.Params =
      "hrStorageSize." + d[0].Index + ",hrStorageUsed." + d[0].Index;
    pollingTmp.Script = `
      s = hrStorageSize;
      u = hrStorageUsed;
      rate = s ? (100.0*u)/s : 0.0;
      setResult("rate",rate);
      rate < 95.0
    `;
    pollingTmp.Level = "low";
    showPolling = true;
  };

  const watchHrProcess = async () => {
    const d = hrProcessTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(node.ID);
    pollingTmp.Name = d[0].Name + " process count";
    pollingTmp.Type = "snmp";
    pollingTmp.Mode = "process";
    pollingTmp.Filter = d[0].Name;
    pollingTmp.Script = `count > 0`;
    pollingTmp.Level = "low";
    showPolling = true;
  };

  const onOpen = async () => {
    node = undefined;
    ports = undefined;
    power = undefined;
    hostResource = undefined;
    logTable = undefined;
    pollingTmp = undefined;
    memo = "";
    activeTab = "basic";
    clearSelectedCount();
    try {
      const conf = await GetMapConf();
      hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
    } catch (e) {
      hasAI = false;
    }
    try {
      node = await GetNode(id);
      memo = await GetNodeMemo(id);
      nodeImage = await GetNodeImage(id);
      rawImageSrc = nodeImage;
      if (nodeImage) {
        const img = new Image();
        img.onload = () => {
          imgNaturalWidth = img.naturalWidth || img.width;
          imgNaturalHeight = img.naturalHeight || img.height;
          imgResizedWidth = imgNaturalWidth;
          imgResizedHeight = imgNaturalHeight;
        };
        img.src = nodeImage;
      } else {
        imgNaturalWidth = 0;
        imgNaturalHeight = 0;
        imgResizedWidth = 0;
        imgResizedHeight = 0;
      }
      confirmDelete = false;
    } catch (e) {
      console.error("Error loading node/memo/image:", e);
      node = {
        ID: id,
        Name: "Unknown Node",
        IP: "",
        MAC: "",
        Vendor: "Unknown",
        Descr: "Failed to load node info: " + String(e)
      };
    }
  };

  const handleTabChange = (tab: string) => {
    switch (tab) {
      case "basic":
        clearSelectedCount();
        break;
      case "memo":
        clearSelectedCount();
        showMemo();
        break;
      case "downtime":
        clearSelectedCount();
        showDowntime();
        break;
      case "log":
        clearSelectedCount();
        showLog();
        break;
      case "panel":
        clearSelectedCount();
        showVPanel();
        break;
      case "hostinfo":
        clearSelectedCount();
        showHrSystem();
        break;
      case "storage":
        clearSelectedCount();
        showHrStorage();
        break;
      case "device":
        clearSelectedCount();
        showHrDevice();
        break;
      case "filesystem":
        clearSelectedCount();
        showHrFileSystem();
        break;
      case "process":
        clearSelectedCount();
        showHrProcess();
        break;
    }
  };

  $: if (activeTab) {
    handleTabChange(activeTab);
  }

  const resizeChart = () => {
    if (chart) {
      try {
        chart.resize();
      } catch (e) {
        // ignore
      }
    }
    if (chartMem) {
      try {
        chartMem.resize();
      } catch (e) {
        // ignore
      }
    }
  };

  let copiedIP = false;
  let copiedMAC = false;

  let memo = "";
  let savedMemo = false;
  let showSaveMemoBtn = false;
  $: showSaveMemoBtn = (activeTab === "memo");

  let nodeImage = "";
  let rawImageSrc = "";
  let imgNaturalWidth = 0;
  let imgNaturalHeight = 0;
  let imgResizedWidth = 0;
  let imgResizedHeight = 0;
  let maxDimension = 800;
  let quality = 0.85;
  let confirmDelete = false;
  let showImageEnlargeModal = false;
  let isDraggingOver = false;
  let fileInput: HTMLInputElement;

  $: imgDataSizeKB = nodeImage ? Math.max(1, Math.round((nodeImage.length * 3) / 4 / 1024)) : 0;

  const resizeImage = (
    src: string,
    maxDim: number,
    q: number
  ): Promise<{ dataUrl: string; width: number; height: number; naturalWidth: number; naturalHeight: number }> => {
    return new Promise((resolve) => {
      const img = new Image();
      img.onload = () => {
        const nw = img.naturalWidth || img.width;
        const nh = img.naturalHeight || img.height;
        let tw = nw;
        let th = nh;

        if (maxDim > 0 && (nw > maxDim || nh > maxDim)) {
          if (nw >= nh) {
            tw = maxDim;
            th = Math.round((nh * maxDim) / nw);
          } else {
            th = maxDim;
            tw = Math.round((nw * maxDim) / nh);
          }
        }

        const canvas = document.createElement("canvas");
        canvas.width = tw;
        canvas.height = th;
        const ctx = canvas.getContext("2d");
        if (!ctx) {
          resolve({ dataUrl: src, width: tw, height: th, naturalWidth: nw, naturalHeight: nh });
          return;
        }
        ctx.drawImage(img, 0, 0, tw, th);
        const dataUrl = canvas.toDataURL("image/jpeg", q);
        resolve({
          dataUrl,
          width: tw,
          height: th,
          naturalWidth: nw,
          naturalHeight: nh,
        });
      };
      img.onerror = () => {
        resolve({ dataUrl: src, width: 0, height: 0, naturalWidth: 0, naturalHeight: 0 });
      };
      img.src = src;
    });
  };

  const applyImage = async (dataUrl: string, maxDim = maxDimension) => {
    rawImageSrc = dataUrl;
    const res = await resizeImage(dataUrl, maxDim, quality);
    nodeImage = res.dataUrl;
    imgNaturalWidth = res.naturalWidth;
    imgNaturalHeight = res.naturalHeight;
    imgResizedWidth = res.width;
    imgResizedHeight = res.height;
    confirmDelete = false;
  };

  const handleMaxDimensionChange = async (dim: number) => {
    maxDimension = dim;
    if (rawImageSrc) {
      await applyImage(rawImageSrc, dim);
    }
  };

  const handlePaste = (e: ClipboardEvent) => {
    if (activeTab !== "memo") return;
    const items = e.clipboardData?.items;
    if (!items) return;
    for (let i = 0; i < items.length; i++) {
      if (items[i].type.indexOf("image") !== -1) {
        const file = items[i].getAsFile();
        if (file) {
          e.preventDefault();
          const reader = new FileReader();
          reader.onload = async (evt) => {
            if (evt.target?.result) {
              await applyImage(String(evt.target.result));
            }
          };
          reader.readAsDataURL(file);
          return;
        }
      }
    }
  };

  const handleDrop = (e: DragEvent) => {
    e.preventDefault();
    isDraggingOver = false;
    const files = e.dataTransfer?.files;
    if (files && files.length > 0 && files[0].type.startsWith("image/")) {
      const reader = new FileReader();
      reader.onload = async (evt) => {
        if (evt.target?.result) {
          await applyImage(String(evt.target.result));
        }
      };
      reader.readAsDataURL(files[0]);
    }
  };

  const selectImageFile = async () => {
    try {
      const p = await SelectFile($_("NodeReport.SelectImageFile") || "画像ファイル選択", true);
      if (p) {
        const data = await GetImage(p);
        if (data) {
          await applyImage(data);
          return;
        }
      }
    } catch (e) {
      console.warn("SelectFile failed:", e);
    }
    fileInput?.click();
  };

  const handleFileInputChange = (e: Event) => {
    const target = e.target as HTMLInputElement;
    if (target?.files && target.files.length > 0) {
      const file = target.files[0];
      const reader = new FileReader();
      reader.onload = async (evt) => {
        if (evt.target?.result) {
          await applyImage(String(evt.target.result));
        }
      };
      reader.readAsDataURL(file);
      target.value = "";
    }
  };

  const removeImage = () => {
    nodeImage = "";
    rawImageSrc = "";
    imgNaturalWidth = 0;
    imgNaturalHeight = 0;
    imgResizedWidth = 0;
    imgResizedHeight = 0;
    confirmDelete = false;
  };

  const dispatch = createEventDispatcher();
  const showMemo = () => {
    clearSelectedCount();
    showSaveMemoBtn = true;
  }
  const saveMemo = async () => {
    savedMemo = true
    try {
      await SaveNodeMemo(id, memo);
      await SaveNodeImage(id, nodeImage);
    } catch (e) {
      console.error("Failed to save node memo or image:", e);
    }
    setTimeout(()=> {
      savedMemo = false
    },1000)
  }


  let prevShow = false;
  $: {
    if (show && !prevShow) {
      onOpen();
    } else if (!show && prevShow) {
      deleteVPanel();
      dispatch('close');
    }
    prevShow = show;
  }
</script>

<svelte:window onresize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[90vh]"
>
  {#if !node}
    <div class="text-center mt-10"><Spinner size="16" /></div>
  {:else}
    <div class="flex flex-col space-y-4">
      <!-- Custom Tabs List -->
      <ul class="flex flex-wrap -mb-px text-xs font-medium text-center border-b border-gray-200 dark:border-gray-700">
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'basic' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'basic'}
          >
            <Icon path={icons.mdiChartPie} size={0.75} />
            {$_("NodeReport.BasicInfo")}
          </button>
        </li>
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'memo' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'memo'}
          >
            <Icon path={icons.mdiNote} size={0.75} />
            {$_('NodeReport.Memo')}
          </button>
        </li>
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'downtime' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'downtime'}
          >
            <Icon path={icons.mdiTimerAlertOutline} size={0.75} />
            {$_("EventLogReport.Downtime")}
          </button>
        </li>
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'log' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'log'}
          >
            <Icon path={icons.mdiCalendarCheck} size={0.75} />
            {$_("NodeReport.Log")}
          </button>
        </li>
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'panel' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'panel'}
          >
            {#if waitVPanel}
              <Spinner color="red" size="4" />
            {:else}
              <Icon path={icons.mdiAppsBox} size={0.75} />
            {/if}
            {$_("NodeReport.Panel")}
          </button>
        </li>
        <li class="mr-1">
          <button
            type="button"
            class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'hostinfo' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
            onclick={() => activeTab = 'hostinfo'}
          >
            {#if waitHr}
              <Spinner color="red" size="4" />
            {:else}
              <Icon path={icons.mdiInformation} size={0.75} />
            {/if}
            <span>{$_("NodeReport.HostInfo")}</span>
          </button>
        </li>
        {#if hostResource}
          <li class="mr-1">
            <button
              type="button"
              class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'storage' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
              onclick={() => activeTab = 'storage'}
            >
              <Icon path={icons.mdiDatabase} size={0.75} />
              {$_("NodeReport.Storage")}
            </button>
          </li>
          <li class="mr-1">
            <button
              type="button"
              class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'device' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
              onclick={() => activeTab = 'device'}
            >
              <Icon path={icons.mdiApplicationCog} size={0.75} />
              {$_("NodeReport.Device")}
            </button>
          </li>
          <li class="mr-1">
            <button
              type="button"
              class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'filesystem' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
              onclick={() => activeTab = 'filesystem'}
            >
              <Icon path={icons.mdiFileCabinet} size={0.75} />
              {$_('NodeReport.FileSystem')}
            </button>
          </li>
          <li class="mr-1">
            <button
              type="button"
              class="inline-flex items-center justify-center px-2 py-2 border-b-2 rounded-t-lg gap-1.5 group {activeTab === 'process' ? 'border-blue-600 text-blue-600 dark:text-blue-500 dark:border-blue-500' : 'border-transparent text-gray-500 hover:text-gray-600 hover:border-gray-300 dark:text-gray-400 dark:hover:text-gray-300'}"
              onclick={() => activeTab = 'process'}
            >
              <Icon path={icons.mdiViewList} size={0.75} />
              {$_("NodeReport.Process")}
            </button>
          </li>
        {/if}
      </ul>

      <!-- Custom Tabs Content Panel -->
      <div class="p-4 bg-transparent rounded-lg mt-2">
        {#if activeTab === 'basic'}
          <Table striped={true}>
            <TableHead>
              <TableHeadCell>{$_("NodeReport.Item")}</TableHeadCell>
              <TableHeadCell>{$_("NodeReport.Content")}</TableHeadCell>
            </TableHead>
            <TableBody class="divide-y">
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Name")}</TableBodyCell>
                <TableBodyCell>{node.Name}</TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Status")}</TableBodyCell>
                <TableBodyCell>
                  <span
                    class="mdi {getIcon(node.Icon)} text-xl"
                    style="color:{getStateColor(node.State)};"></span>
                  <span class="ml-2 text-xs text-black dark:text-white"
                    >{getStateName(node.State)}</span
                  >
                </TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.IPAddress")}</TableBodyCell>
                <TableBodyCell>
                  {node.IP}
                  <Button
                    color="alternative"
                    type="button"
                    class="ml-2 !p-2"
                    onclick={async () => {
                      copiedIP = true
                      copyText(node.IP)
                      setTimeout(()=> copiedIP = false,2000);
                    }}
                    size="xs"
                  >
                    {#if copiedIP}
                      <Icon path={icons.mdiCheck} size={1} />
                    {:else}
                      <Icon path={icons.mdiContentCopy} size={1} />
                    {/if}
                  </Button>
                </TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.MACAddress")}</TableBodyCell>
                <TableBodyCell>
                  {node.MAC}
                  {#if node.MAC}
                    <Button
                      color="alternative"
                      type="button"
                      class="ml-2 !p-2"
                      onclick={async () => {
                        copiedMAC = true
                        copyText(node.MAC)
                        setTimeout(()=> copiedMAC = false,2000);
                      }}
                      size="xs"
                    >
                      {#if copiedMAC}
                        <Icon path={icons.mdiCheck} size={1} />
                      {:else}
                        <Icon path={icons.mdiContentCopy} size={1} />
                      {/if}
                    </Button>
                  {/if}
                </TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_('NodeList.Vendor')}</TableBodyCell>
                <TableBodyCell>{node.Vendor}</TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Descr")}</TableBodyCell>
                <TableBodyCell>{node.Descr}</TableBodyCell>
              </TableBodyRow>
            </TableBody>
          </Table>
        {:else if activeTab === 'memo'}
          <div
            class="grid grid-cols-1 lg:grid-cols-2 gap-4 h-[60vh]"
            onpaste={handlePaste}
          >
            <!-- Left: Text Memo -->
            <div class="flex flex-col h-full min-h-0">
              <div class="text-xs font-semibold text-gray-300 mb-1.5 flex items-center justify-between h-5">
                <span class="flex items-center gap-1.5">
                  <Icon path={icons.mdiNoteTextOutline} size={0.8} />
                  {$_("NodeReport.MemoText")}
                </span>
              </div>
              <textarea
                placeholder={$_('NodeReport.MemoPlaceHolder')}
                class="w-full flex-1 min-h-0 resize-none font-mono text-sm p-3 rounded-lg bg-gray-50 dark:bg-gray-700 text-gray-900 dark:text-white border border-gray-300 dark:border-gray-600 focus:ring-blue-500 focus:border-blue-500 dark:focus:ring-blue-500 dark:focus:border-blue-500 focus:outline-none shadow-sm"
                bind:value={memo}
              ></textarea>
            </div>

            <!-- Right: Node Appearance Image -->
            <div class="flex flex-col h-full min-h-0">
              <div class="text-xs font-semibold text-gray-300 mb-1.5 flex items-center justify-between h-5">
                <span class="flex items-center gap-1.5">
                  <Icon path={icons.mdiCameraOutline} size={0.8} />
                  {$_("NodeReport.NodeImage")}
                </span>
                {#if nodeImage}
                  <span class="text-[11px] text-gray-400 font-mono">
                    {imgResizedWidth} × {imgResizedHeight} px ({imgDataSizeKB} KB)
                  </span>
                {/if}
              </div>

              <div class="flex-1 flex flex-col min-h-0 bg-gray-800/60 rounded-lg p-3 border border-gray-700">
                {#if nodeImage}
                  <!-- Image Preview and Controls -->
                  <div class="flex-1 flex flex-col min-h-0">
                    <div class="flex-1 relative flex items-center justify-center bg-gray-900/80 rounded border border-gray-700/80 overflow-hidden group">
                      <img
                        src={nodeImage}
                        alt="Node Appearance"
                        class="max-w-full max-h-[32vh] object-contain cursor-pointer transition-transform duration-200 group-hover:scale-[1.02]"
                        onclick={() => (showImageEnlargeModal = true)}
                        title={$_("NodeReport.EnlargeImage")}
                      />
                      <button
                        type="button"
                        class="absolute top-2 right-2 p-1.5 bg-gray-900/80 hover:bg-gray-800 text-gray-200 rounded-full border border-gray-600 shadow opacity-0 group-hover:opacity-100 transition-opacity"
                        onclick={() => (showImageEnlargeModal = true)}
                        title={$_("NodeReport.EnlargeImage")}
                      >
                        <Icon path={icons.mdiMagnifyPlusOutline} size={0.8} />
                      </button>
                    </div>

                    <!-- Controls below image -->
                    <div class="mt-2.5 pt-2 border-t border-gray-700 flex flex-col space-y-2">
                      <div class="flex items-center justify-between text-xs text-gray-400">
                        <span class="flex items-center gap-1">
                          <Icon path={icons.mdiArrowExpand} size={0.7} />
                          {$_("NodeReport.MaxDimension")}:
                        </span>
                        <div class="flex items-center gap-1">
                          {#each [400, 600, 800, 1200, 0] as dim}
                            <button
                              type="button"
                              class="px-2 py-0.5 text-[11px] rounded border transition-colors {maxDimension === dim ? 'bg-blue-600 text-white border-blue-500 font-semibold' : 'bg-gray-700 text-gray-300 border-gray-600 hover:bg-gray-600'}"
                              onclick={() => handleMaxDimensionChange(dim)}
                            >
                              {dim === 0 ? $_("NodeReport.Original") : `${dim}px`}
                            </button>
                          {/each}
                        </div>
                      </div>

                      <div class="flex items-center justify-between pt-1">
                        <div class="flex items-center gap-2">
                          <Button
                            size="xs"
                            color="alternative"
                            onclick={selectImageFile}
                          >
                            <Icon path={icons.mdiImageEditOutline} size={0.8} class="mr-1" />
                            {$_("NodeReport.ChangeImage")}
                          </Button>
                        </div>

                        <div class="flex items-center">
                          {#if !confirmDelete}
                            <Button
                              size="xs"
                              color="red"
                              onclick={() => (confirmDelete = true)}
                            >
                              <Icon path={icons.mdiDeleteOutline} size={0.8} class="mr-1" />
                              {$_("NodeReport.DeleteImage")}
                            </Button>
                          {:else}
                            <div class="flex items-center gap-1">
                              <Button
                                size="xs"
                                color="red"
                                onclick={removeImage}
                              >
                                {$_("NodeReport.ConfirmDelete")}
                              </Button>
                              <Button
                                size="xs"
                                color="alternative"
                                onclick={() => (confirmDelete = false)}
                              >
                                {$_("NodeReport.Cancel")}
                              </Button>
                            </div>
                          {/if}
                        </div>
                      </div>
                    </div>
                  </div>
                {:else}
                  <!-- Dropzone / Empty State -->
                  <div
                    class="flex-1 flex flex-col items-center justify-center border-2 border-dashed rounded-lg p-6 text-center transition-colors {isDraggingOver ? 'border-blue-500 bg-blue-950/20' : 'border-gray-600 hover:border-gray-500 bg-gray-900/40'}"
                    ondragover={(e) => { e.preventDefault(); isDraggingOver = true; }}
                    ondragleave={() => (isDraggingOver = false)}
                    ondrop={handleDrop}
                    role="region"
                    aria-label="Image dropzone"
                  >
                    <div class="p-3 bg-gray-800 rounded-full mb-3 text-gray-400">
                      <Icon path={icons.mdiCameraPlusOutline} size={1.8} />
                    </div>
                    <p class="text-xs text-gray-300 mb-1 max-w-xs leading-relaxed">
                      {$_("NodeReport.DropOrPasteImage")}
                    </p>
                    <p class="text-[11px] text-gray-500 mb-3">
                      PNG, JPG, WEBP, SVG
                    </p>
                    <Button
                      size="xs"
                      color="blue"
                      onclick={selectImageFile}
                    >
                      <Icon path={icons.mdiFolderOpenOutline} size={0.8} class="mr-1" />
                      {$_("NodeReport.SelectImageFile")}
                    </Button>
                  </div>
                {/if}

                <!-- Hidden input for file selection fallback -->
                <input
                  type="file"
                  accept="image/*"
                  class="hidden"
                  bind:this={fileInput}
                  onchange={handleFileInputChange}
                />
              </div>
            </div>
          </div>
        {:else if activeTab === 'downtime'}
          <div class="flex flex-col space-y-3 w-full p-1">
            <!-- Summary Cards (Compact) -->
            <div class="grid grid-cols-2 md:grid-cols-4 gap-3 mb-1">
              <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                <div class="text-xs text-gray-400 font-semibold mb-0.5">
                  {$_("EventLogReport.OverallSLA")}
                </div>
                <div
                  class="text-lg font-bold {downtimeStats.overallSLA >= 99.9
                    ? 'text-green-400'
                    : downtimeStats.overallSLA >= 99.0
                    ? 'text-yellow-400'
                    : 'text-red-400'}"
                >
                  {downtimeStats.overallSLA.toFixed(3)}%
                </div>
              </div>
              <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                <div class="text-xs text-gray-400 font-semibold mb-0.5">
                  {$_("EventLogReport.TotalIncidents")}
                </div>
                <div class="text-lg font-bold text-blue-400">
                  {downtimeStats.totalIncidents}
                  <span class="text-xs font-normal text-gray-400">
                    ({$_("EventLogReport.OngoingIncidents")}: {downtimeStats.ongoingIncidents})
                  </span>
                </div>
              </div>
              <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                <div class="text-xs text-gray-400 font-semibold mb-0.5">
                  {$_("EventLogReport.MTTR")}
                </div>
                <div class="text-lg font-bold text-teal-400">
                  {renderDuration(downtimeStats.mttrSec)}
                </div>
              </div>
              <div class="p-2.5 bg-gray-800 rounded-lg border border-gray-700 text-center shadow-md">
                <div class="text-xs text-gray-400 font-semibold mb-0.5">
                  {$_("EventLogReport.MaxDowntime")} / {$_("EventLogReport.TotalDowntime")}
                </div>
                <div class="text-xs font-bold text-red-400 mt-1">
                  {renderDuration(downtimeStats.maxDowntimeSec)}
                  <span class="text-xs font-normal text-gray-400">
                    / {renderDuration(downtimeStats.totalDowntimeSec)}
                  </span>
                </div>
              </div>
            </div>

            <!-- 2 Columns Grid: Graph & Table -->
            <div class="grid grid-cols-1 lg:grid-cols-12 gap-3 w-full items-start">
              <!-- Left Column: ECharts Graph (5 cols) -->
              <div class="lg:col-span-5 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 shadow-inner">
                <div id="nodeDowntimeChart" class="w-full h-[360px]"></div>
              </div>

              <!-- Right Column: Detailed Table (7 cols) wrapped in div (AGENTS.md Rule) -->
              <div class="lg:col-span-7 w-full bg-gray-900 rounded-lg p-2 border border-gray-800 overflow-x-auto shadow-inner">
                <div>
                  <table
                    id="nodeDowntimeTable"
                    class="display compact w-full text-xs text-left text-gray-300"
                  ></table>
                </div>
              </div>
            </div>
          </div>
        {:else if activeTab === 'log'}
          <table id="nodeReportLogTable" class="display compact" style="width:99%"></table>
        {:else if activeTab === 'panel'}
          {#if ports && ports.length > 0}
            <div id="vpanel"></div>
            <div>
              <table
                id="portTable"
                class="display compact mt-5"
                style="width:99%"></table>
            </div>
          {:else if !waitVPanel}
            <div>{$_("NodeReport.NoVPanel")}</div>
          {/if}
        {:else if activeTab === 'hostinfo'}
          {#if hostResource}
            <div class="grid grid-cols-2 gap-1">
              <div id="hrSummaryChart"></div>
              <div>
                <table
                  id="hrSystemTable"
                  class="display compact"
                  style="width:100%"></table>
              </div>
            </div>
          {:else if !waitHr}
            <div>{$_("NodeReport.NoHRMIB")}</div>
          {/if}
        {:else if activeTab === 'storage' && hostResource}
          <div id="hrStorageChart" class="mb-2"></div>
          <table
            id="hrStorageTable"
            class="display compact mt-2"
            style="width:99%"></table>
        {:else if activeTab === 'device' && hostResource}
          <table
            id="hrDeviceTable"
            class="display compact mt-2"
            style="width:99%"></table>
        {:else if activeTab === 'filesystem' && hostResource}
          <table
            id="hrFileSystemTable"
            class="display compact mt-2"
            style="width:99%"></table>
        {:else if activeTab === 'process' && hostResource}
          <div class="grid grid-cols-2 gap-1 mb-2">
            <div id="hrProcessCPUChart"></div>
            <div id="hrProcessMemChart"></div>
          </div>
          <table
            id="hrProcessTable"
            class="display compact mt-2"
            style="width:99%"></table>
        {/if}
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if selectedPortCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={watchPortState}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.AddPollingIFState")}
          </GradientButton>
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={watchPortTraffic}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.AddPollingTraffic")}
          </GradientButton>
        {/if}
        {#if selectedHrSystemCount > 0 && canWacthHrSystem()}
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={watchHrSystem}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
        {/if}
        {#if selectedhrStorageCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={watchHrStorage}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
        {/if}
        {#if selectedHrProcessCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={watchHrProcess}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
        {/if}
        {#if showVPanelBtn && ports && ports.length > 0}
          <Toggle bind:checked={physicalPort} onchange={showVPanel}>
            {$_("NodeReport.PhysicalPort")}
          </Toggle>
          <Select
            size="sm"
            placeholder=""
            class="w-20"
            classes={{ select: "!py-1 !px-2 text-xs h-8" }}
            items={[
              { value: 16, name: "16" },
              { value: 24, name: "24" },
              { value: 32, name: "32" },
              { value: 48, name: "48" },
              { value: 64, name: "64" },
            ]}
            bind:value={vpanelPortWrap}
          />
          <div class="flex items-center gap-1">
            <Icon path={icons.mdiMagnifyMinus} size={0.8} />
            <input
              type="range"
              bind:value={vpanelZoom}
              min="0.1"
              max="5"
              step="0.1"
              class="w-24 h-2 bg-gray-200 rounded-lg appearance-none cursor-pointer dark:bg-gray-700"
            />
            <Icon path={icons.mdiMagnifyPlus} size={0.8} />
          </div>
          <Toggle bind:checked={rotateVPanel}>
            {$_("NodeReport.RotateVPanel")}
          </Toggle>
        {/if}
        {#if showSaveMemoBtn}
          <GradientButton
          shadow
          color="blue"
          type="button"
          onclick={saveMemo}
          size="xs"
        >
          {#if savedMemo}
            <Icon path={icons.mdiCheck} size={1} />
          {:else}
            <Icon path={icons.mdiContentSave} size={1} />
          {/if}
          <span>
            {$_('NodeReport.Save')}
          </span>
        </GradientButton>
        {/if}
        {#if hasAI && activeTab !== 'basic' && activeTab !== 'memo'}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={() => (showAIReport = true)}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_("ReportAI.AIExplain")}
        </GradientButton>
        {/if}
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={close}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("NodeReport.Close")}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<Polling bind:show={showPolling} {pollingTmp} />

<ReportAIDialog
  bind:show={showAIReport}
  title={$_("ReportAI.Title")}
  exportFilename={`node_${id}_${activeTab}_ai_explanation`}
  analyzeFunc={() => LLMExplainNodeReport(id, activeTab)}
/>

<Modal
  bind:open={showImageEnlargeModal}
  size="xl"
  dismissable={true}
  outsideclose={true}
  title={$_("NodeReport.NodeImage")}
>
  <div class="flex items-center justify-center p-2 max-h-[75vh] overflow-auto">
    {#if nodeImage}
      <img
        src={nodeImage}
        alt="Node Appearance Full"
        class="max-w-full max-h-[70vh] object-contain rounded shadow"
      />
    {/if}
  </div>
</Modal>

<style>
  #vpanel {
    width: 98%;
    min-height: 400px;
    height: 40vh;
    overflow: hidden;
    margin: 0 auto;
  }

  #hrSummaryChart {
    min-width: 350px;
    min-height: 350px;
    margin: 0 auto;
    width: 35vw;
    height: 35vw;
  }
  #hrStorageChart {
    min-height: 300px;
    height: 35vh;
    width: 98%;
  }
  #hrProcessCPUChart,
  #hrProcessMemChart {
    min-height: 300px;
    height: 35vh;
    width: 98%;
  }
</style>
