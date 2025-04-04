<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Spinner,
    Toggle,
    Button,
    Textarea,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNode,
    GetVPanelPorts,
    GetVPanelPowerInfo,
    GetEventLogs,
    GetHostResource,
    GetDefaultPolling,
    GetNodeMemo,
    SaveNodeMemo,
  } from "../../wailsjs/go/main/App";
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
  } from "./common";
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

  let selectedPortCount = 0;
  let selectedHrSystemCount = 0;
  let selectedhrStorageCount = 0;
  let selectedHrProcessCount = 0;
  let showPolling = false;
  let physicalPort = true;
  let showVPanelBtn = false;
  let chart: any = undefined;
  let chartMem: any = undefined;

  const clearSelectedCount = () => {
    selectedPortCount = 0;
    selectedHrSystemCount = 0;
    selectedhrStorageCount = 0;
    selectedHrProcessCount = 0;
    showVPanelBtn = false;
    chart = undefined;
    chartMem = undefined;
  };

  let logTable = undefined;
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
    logTable = new DataTable("#logTable", {
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
  };

  let portTable: any = undefined;

  const showPortTable = (p: any) => {
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
          width: "10%",
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
          width: "15%",
        },
        {
          data: "Speed",
          title: $_("NodeReport.Speed"),
          width: "10%",
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

  const showVPanel = async () => {
    clearSelectedCount();
    showVPanelBtn = true;
    initVPanel("vpanel");
    if (!ports) {
      waitVPanel = true;
      ports = await GetVPanelPorts(id);
      power = await GetVPanelPowerInfo(id);
      waitVPanel = false;
    }
    const p = physicalPort ? ports.filter((e: any) => e.Type == 6) : ports;
    setVPanel(p, power, rotateVPanel);
    showPortTable(p);
  };

  const renderStatus = (s: any) => {
    switch (s) {
      case "Running":
        return (
          `<span class="text-blue-700">` + $_("NodeReport.Running") + `</span>`
        );
      case "Runnable":
        return (
          `<span class="text-blue-900">` + $_("NodeReport.Runnable") + `</span>`
        );
      case "Testing":
        return $_("NodeReport.Testing");
      case "NotRunnable":
        return $_("NodeReport.NotRunnable");
      case "Invalid":
      case "Down":
        return (
          `<span class="text-red-800">` + $_("NodeReport.Down") + `</span>`
        );
    }
    return $_("NodeReport.Unknown");
  };

  const renderRate = (r: any) => {
    if (r < 80.0) {
      return `<span class="text-blue-700">${r.toFixed(2)}</span>`;
    } else if (r < 90.0) {
      return `<span class="text-yellow-700">${r.toFixed(2)}</span>`;
    }
    return `<span class="text-red-700">${r.toFixed(2)}</span>`;
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
  let hrDeviceTable = undefined;
  let hrFileSystemTable = undefined;
  let hrProcessTable: any = undefined;
  let waitHr = false;

  const showHrSystem = async () => {
    if (!hostResource) {
      waitHr = true;
      hostResource = await GetHostResource(id);
      waitHr = false;
      await tick();
    }
    if (!hostResource) {
      return;
    }
    clearSelectedCount();
    hrSystemTable = new DataTable("#hrSystemTable", {
      paging: false,
      searching: false,
      info: false,
      data: hostResource.System,
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
  };

  const showHrStorage = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    hrStorageTable = new DataTable("#hrStorageTable", {
      paging: false,
      searching: false,
      info: false,
      scrollY: "25vh",
      data: hostResource.Storage,
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
  };

  const showHrDevice = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    hrDeviceTable = new DataTable("#hrDeviceTable", {
      paging: false,
      searching: false,
      info: false,
      scrollY: "55vh",
      data: hostResource.Device,
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
  };

  const showHrFileSystem = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    hrFileSystemTable = new DataTable("#hrFileSystemTable", {
      paging: false,
      searching: false,
      info: false,
      scrollY: "55vh",
      data: hostResource.FileSystem,
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
  };

  const showHrProcess = () => {
    clearSelectedCount();
    if (!hostResource) {
      return;
    }
    hrProcessTable = new DataTable("#hrProcessTable", {
      paging: false,
      searching: false,
      info: false,
      scrollY: "25vh",
      data: hostResource.Process,
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
    chart = showHrProcChart(true);
    chartMem = showHrProcChart(false);
  };

  const showHrSummaryChart = async () => {
    if (!hostResource) {
      return;
    }
    await tick();
    const data = {
      CPU: 0,
      Mem: 0,
      VM: 0,
    };
    hostResource.System.forEach((e: any) => {
      if (e.Key == "hrProcessorLoad") {
        data.CPU = Number(e.Value);
      }
    });
    hostResource.Storage.forEach((e: any) => {
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
  };

  const showHrStorageChart = async () => {
    if (!hostResource) {
      return;
    }
    await tick();
    const list: any = [];
    hostResource.Storage.forEach((e: any) => {
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
  };

  const showHrProcChart = async (bCPU: boolean) => {
    if (!hostResource) {
      return;
    }
    await tick();
    let max = 0;
    const list: any = [];
    hostResource.Process.forEach((e: any) => {
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
    return showHrBarChart(
      bCPU ? "hrProcessCPUChart" : "hrProcessMemChart",
      bCPU ? $_("NodeReport.CPUUsage") : $_("NodeReport.MemUsage"),
      bCPU ? $_("NodeReport.Sec") : "Bytes",
      list,
      max
    );
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
    node = await GetNode(id);
    memo = await GetNodeMemo(id)
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
    if (chartMem) {
      chartMem.resize();
    }
  };

  let copiedIP = false;
  let copiedMAC = false;

  let memo = "";
  let savedMemo = false;
  let showSaveMemoBtn = false
  const showMemo = () => {
    clearSelectedCount();
    showSaveMemoBtn = true;
  }
  const saveMemo = async () => {
    savedMemo = true
    await SaveNodeMemo(id,memo)
    setTimeout(()=> {
      savedMemo = false
    },1000)
  }

</script>

<svelte:window on:resize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  {#if !node}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <div class="flex flex-col space-y-4">
      <Tabs style="underline">
        <TabItem open on:click={clearSelectedCount}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("NodeReport.BasicInfo")}
          </div>
          <Table striped={true}>
            <TableHead>
              <TableHeadCell>{$_("NodeReport.Item")}</TableHeadCell>
              <TableHeadCell>{$_("NodeReport.Content")}</TableHeadCell>
            </TableHead>
            <TableBody tableBodyClass="divide-y">
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Name")}</TableBodyCell>
                <TableBodyCell>{node.Name}</TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Status")}</TableBodyCell>
                <TableBodyCell>
                  <span
                    class="mdi {getIcon(node.Icon)} text-xl"
                    style="color:{getStateColor(node.State)};"
                  />
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
                    on:click={async () => {
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
                      on:click={async () => {
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
        </TabItem>
        <TabItem on:click={showMemo}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiNote} size={1} />
            {$_('NodeReport.Memo')}
          </div>
          <Textarea placeholder={$_('NodeReport.MemoPlaceHolder')} rows="16" bind:value={memo} />
        </TabItem>
        <TabItem on:click={showLog}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiCalendarCheck} size={1} />
            {$_("NodeReport.Log")}
          </div>
          <table id="logTable" class="display compact" style="width:99%" />
        </TabItem>
        <TabItem on:click={showVPanel}>
          <div slot="title" class="flex items-center gap-2">
            {#if waitVPanel}
              <Spinner color="red" size="6" />
            {:else}
              <Icon path={icons.mdiAppsBox} size={1} />
            {/if}
            {$_("NodeReport.Panel")}
          </div>
          <div id="vpanel" />
          <table
            id="portTable"
            class="display compact mt-5"
            style="width:99%"
          />
        </TabItem>
        <TabItem on:click={showHrSystem} >
          <div slot="title" class="flex items-center gap-2">
            {#if waitHr}
              <Spinner color="red" size="6" />
            {:else}
              <Icon path={icons.mdiInformation} size={1} />
            {/if}
            <span>{$_("NodeReport.HostInfo")}</span>
          </div>
          {#if hostResource}
            <div class="grid grid-cols-2 gap-1">
              <div id="hrSummaryChart" />
              <div>
                <table
                  id="hrSystemTable"
                  class="display compact"
                  style="width:100%"
                />
              </div>
            </div>
          {:else if !waitHr}
            <div>{$_("NodeReport.NoHRMIB")}</div>
          {/if}
        </TabItem>
        {#if hostResource}
          <TabItem on:click={showHrStorage}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiDatabase} size={1} />
              {$_("NodeReport.Storage")}
            </div>
            <div id="hrStorageChart" class="mb-2" />
            <table
              id="hrStorageTable"
              class="display compact mt-2"
              style="width:99%"
            />
          </TabItem>
          <TabItem on:click={showHrDevice}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiApplicationCog} size={1} />
              {$_("NodeReport.Device")}
            </div>
            <table
              id="hrDeviceTable"
              class="display compact mt-2"
              style="width:99%"
            />
          </TabItem>
          <TabItem on:click={showHrFileSystem}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiFileCabinet} size={1} />
              {$_('NodeReport.FileSystem')}
            </div>
            <table
              id="hrFileSystemTable"
              class="display compact mt-2"
              style="width:99%"
            />
          </TabItem>
          <TabItem on:click={showHrProcess}>
            <div slot="title" class="flex items-center gap-2">
              <Icon path={icons.mdiViewList} size={1} />
              {$_("NodeReport.Process")}
            </div>
            <div class="grid grid-cols-2 gap-1 mb-2">
              <div id="hrProcessCPUChart" />
              <div id="hrProcessMemChart" />
            </div>
            <table
              id="hrProcessTable"
              class="display compact mt-2"
              style="width:99%"
            />
          </TabItem>
        {/if}
      </Tabs>
      <div class="flex justify-end space-x-2 mr-2">
        {#if selectedPortCount > 0}
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={watchPortState}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.AddPollingIFState")}
          </GradientButton>
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={watchPortTraffic}
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
            on:click={watchHrSystem}
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
            on:click={watchHrStorage}
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
            on:click={watchHrProcess}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
        {/if}
        {#if showVPanelBtn}
          <Toggle bind:checked={physicalPort} on:change={showVPanel}>
            {$_("NodeReport.PhysicalPort")}
          </Toggle>
          <Toggle bind:checked={rotateVPanel} on:change={showVPanel}>
            {$_("NodeReport.RotateVPanel")}
          </Toggle>
        {/if}
        {#if showSaveMemoBtn}
          <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={saveMemo}
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
        <GradientButton
          shadow
          type="button"
          color="teal"
          on:click={close}
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

<style>
  #vpanel {
    width: 98%;
    min-height: 400px;
    height: 40vh;
    overflow: scroll;
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
