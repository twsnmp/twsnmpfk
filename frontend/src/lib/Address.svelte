<script lang="ts">
  import { GradientButton } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    ExportArpTable,
    GetArpLogs,
    GetArpTable,
    GetNodes,
    ResetArpTable,
    DeleteArpEnt,
  } from "../../wailsjs/go/main/App";
  import {
    getTableLang,
    renderTime,
    getStateColor,
    getStateIcon,
  } from "./common";
  import AddressReport from "./AddressReport.svelte";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let arp: any = [];
  let nodes = undefined;
  let showReport = false;
  let arpTable = undefined;
  let newIP = new Map();
  let changeIP = new Map();
  let changeMAC = new Map();
  let selectedIP = "";
  let selectedNodeID = "";

  let showEditNode = false;
  let showAddNode = false;
  let showNodeReport = false;

  const showArpTable = () => {
    if (arpTable) {
      arpTable.destroy();
      arpTable = undefined;
    }
    selectedIP = selectedNodeID = "";
    arpTable = new DataTable("#arpTable", {
      columns: arpColumns,
      data: arp,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    arpTable.on("select", () => {
      selectedIP = selectedNodeID = "";
      const d = arpTable.rows({ selected: true }).data();
      if (!d || d.length != 1) {
        return;
      }
      if (d[0].NodeID) {
        selectedNodeID = d[0].NodeID;
      } 
      if (d[0].IP) {
        selectedIP = d[0].IP;
      }
    });
    arpTable.on("deselect", () => {
      const c = arpTable.rows({ selected: true }).count();
      if (c != 1) {
        selectedIP = selectedNodeID = "";
      }
    });
  };

  const renderNode = (id) => {
    return nodes[id] ? nodes[id].Name : id;
  };

  const renderArpIP = (ip: string, type: string) => {
    if (type == "sort") {
      return ip
        .split(".")
        .reduce((int, v) => Number(int) * 256 + Number(v) + "");
    }
    if (changeIP.has(ip)) {
      return `<span class="text-yellow-500">${ip}</span>`;
    }
    if (ip.startsWith("169.254.")) {
      return `<span class="text-red-500">${ip}</span>`;
    }
    return ip;
  };

  const renderArpMAC = (mac: string, type: string) => {
    if (type == "sort") {
      return mac;
    }
    if (changeMAC.has(mac)) {
      return `<span class="text-red-600">${mac}</span>`;
    }
    return mac;
  };

  const renderArpState = (s: string, type: string) => {
    if (type == "sort") {
      return Number(s);
    }
    switch (Number(s)) {
      case 0:
        return (
          `<span class="mdi ` +
          getStateIcon("high") +
          ` text-xl" style="color:` +
          getStateColor("high") +
          `;"></span><span class="ml-2">` +
          $_('Address.Dup') +
          `</span>`
        );
      case 1:
        return (
          `<span class="mdi ` +
          getStateIcon("low") +
          ` text-xl" style="color:` +
          getStateColor("low") +
          `;"></span><span class="ml-2">` +
          $_('Address.IPChange') +
          `</span>`
        );
      case 2:
        return (
          `<span class="mdi ` +
          getStateIcon("warn") +
          ` text-xl" style="color:` +
          getStateColor("warn") +
          `;"></span><span class="ml-2">` +
          $_('Address.MACChange') +
          `</span>`
        );
    }
    return (
      `<span class="mdi ` +
      getStateIcon("normal") +
      ` text-xl" style="color:` +
      getStateColor("normal") +
      `;"></span><span class="ml-2">` +
      $_('Address.Normal') +
      `</span>`
    );
  };

  const arpColumns = [
    {
      data: "State",
      title: $_('Address.State'),
      width: "10%",
      render: renderArpState,
    },
    {
      data: "IP",
      title: $_("Address.IPAddress"),
      width: "15%",
      render: renderArpIP,
    },
    {
      data: "MAC",
      title: $_("Address.MACAddress"),
      width: "15%",
      render: renderArpMAC,
    },
    {
      data: "NodeID",
      title: $_("Address.NodeName"),
      width: "20%",
      render: renderNode,
    },
    {
      data: "Vendor",
      title: $_("Address.Vendor"),
      width: "25%",
      render: renderNode,
    },
    {
      data: "Last",
      title: $_('Address.Last'),
      width: "25%",
      render: renderTime,
    },
  ];

  const refresh = async () => {
    nodes = await GetNodes();
    const arpLogs = await GetArpLogs();
    arp = await GetArpTable();
    changeIP.clear();
    newIP.clear();
    changeMAC.clear();
    arpLogs.reverse();
    for (let i = 0; i < arpLogs.length; i++) {
      if (arpLogs[i].State == "Change") {
        changeIP.set(arpLogs[i].IP, arpLogs[i].Time);
        changeMAC.set(arpLogs[i].NewMAC, true);
        changeMAC.set(arpLogs[i].OldMAC, true);
      } else {
        newIP.set(arpLogs[i].IP, arpLogs[i].Time);
      }
    }
    arp.forEach((e) => {
      e.State = changeIP.has(e.IP)
        ? 1
        : e.IP.startsWith("169.254.")
        ? 0
        : changeMAC.has(e.MAC)
        ? 2
        : 3;
      e.Last = changeIP.has(e.IP)
        ? changeIP.get(e.IP)
        : newIP.has(e.IP)
        ? newIP.get(e.IP)
        : 0;
    });
    showArpTable();
  };

  onMount(() => {
    refresh();
  });

  const saveCSV = () => {
    ExportArpTable("csv");
  };

  const saveExcel = () => {
    ExportArpTable("excel");
  };

  const reset = async () => {
    await ResetArpTable();
    refresh();
  };
</script>

<div class="flex flex-col">
  <table id="arpTable" class="display compact" style="width:99%" />
  <div class="flex justify-end space-x-2 mr-2 mt-2">
    {#if selectedNodeID}
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={() => (showNodeReport = true)}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_("Address.NodeInfo")}
      </GradientButton>
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={() => (showEditNode = true)}
        size="xs"
      >
        <Icon path={icons.mdiPencil} size={1} />
        {$_("Address.EditNode")}
      </GradientButton>
    {/if}

    {#if selectedIP && !selectedNodeID}
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={() => {
          showAddNode = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiPlus} size={1} />
        {$_("Address.AddNode")}
      </GradientButton>
    {/if}

    {#if selectedIP }
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={()=> {
          DeleteArpEnt(selectedIP);
          refresh();
        }}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_('Address.Delete')}
      </GradientButton>
    {/if}

    <GradientButton
      type="button"
      color="green"
      on:click={() => {
        showReport = true;
      }}
      size="xs"
    >
      <Icon path={icons.mdiChartPie} size={1} />
      {$_("Address.Report")}
    </GradientButton>
    <GradientButton shadow color="red" type="button" on:click={reset} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Address.Clear")}
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveCSV}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveExcel}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("Address.Reload")}
    </GradientButton>
  </div>
</div>

{#if showReport}
  <AddressReport
    {arp}
    {changeIP}
    {changeMAC}
    on:close={() => {
      showReport = false;
    }}
  />
{/if}

{#if showAddNode}
  <Node
    ip={selectedIP}
    posX={100}
    posY={120}
    on:close={(e) => {
      refresh();
      showAddNode = false;
    }}
  />
{/if}

{#if showEditNode}
  <Node
    nodeID={selectedNodeID}
    on:close={(e) => {
      refresh();
      showEditNode = false;
    }}
  />
{/if}

{#if showNodeReport}
  <NodeReport
    id={selectedNodeID}
    on:close={(e) => {
      showNodeReport = false;
    }}
  />
{/if}

<style>
  @import "../assets/css/jquery.dataTables.css";
</style>
