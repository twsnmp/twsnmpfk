<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
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
    renderState,
  } from "./common";
  import AddressReport from "./AddressReport.svelte";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { copyText } from "svelte-copy";

  let arp: any = [];
  let nodes :any= undefined;
  let showReport = false;
  let arpTable :any = undefined;
  let newIP = new Map();
  let changeIP:any = new Map();
  let changeMAC:any  = new Map();
  let selectedIP = "";
  let selectedNodeID = "";
  let selectedCount =0;

  let showEditNode = false;
  let showAddNode = false;
  let showNodeReport = false;

  const showArpTable = () => {
    selectedIP = selectedNodeID = "";
    arpTable = new DataTable("#arpTable", {
      destroy: true,
      columns: arpColumns,
      pageLength: window.innerHeight > 800 ? 25 : 10,
      stateSave: true,
      data: arp,
      order: [[0,"asc"]],
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    arpTable.on("select", () => {
      selectedIP = selectedNodeID = "";
      const d = arpTable.rows({ selected: true }).data();
      selectedCount = d.length;
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
      selectedCount = c;
    });
  };

  const renderNode = (id:any) => {
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
          ` text-xs" style="color:` +
          getStateColor("high") +
          `;"></span><span class="ml-2">` +
          $_('Address.Dup') +
          `</span>`
        );
      case 1:
        return (
          `<span class="mdi ` +
          getStateIcon("low") +
          ` text-xs" style="color:` +
          getStateColor("low") +
          `;"></span><span class="ml-2">` +
          $_('Address.IPChange') +
          `</span>`
        );
      case 2:
        return (
          `<span class="mdi ` +
          getStateIcon("warn") +
          ` text-xs" style="color:` +
          getStateColor("warn") +
          `;"></span><span class="ml-2">` +
          $_('Address.MACChange') +
          `</span>`
        );
    }
    return (
      `<span class="mdi ` +
      getStateIcon("normal") +
      ` text-xs" style="color:` +
      getStateColor("normal") +
      `;"></span><span class="ml-2">` +
      $_('Address.Normal') +
      `</span>`
    );
  };

  const renderArpStateString = (s: string) => {
    switch (Number(s)) {
      case 0:
        return $_('Address.Dup');
      case 1:
        return $_('Address.IPChange');
      case 2:
        return $_('Address.MACChange');
    }
    return $_('Address.Normal');
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
    arp.forEach((e:any) => {
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

  let copied = false;
  const copy = () => {
    const selected = arpTable.rows({ selected: true }).data();
    let s: string[] = [];
    const h = arpColumns.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of arpColumns) {
        switch (c.data){
        case "Last":
          row.push(renderTime(selected[i][c.data] || "", ""));
          break;
        case "State":
          row.push(renderArpStateString(selected[i][c.data] || ""));
          break;
        case  "NodeID":
          row.push(renderNode(selected[i][c.data] || ""));
          break;
        default:
          row.push(selected[i][c.data] || "");
        }
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };
 
  const reset = async () => {
    await ResetArpTable();
    refresh();
  };

  const deleteAddress = async () => {
    const selected = arpTable.rows({ selected: true }).data().pluck("IP");
    if (selected.length < 1) {
      return;
    }
    await DeleteArpEnt(selected.toArray());
    arpTable.rows({ selected: true }).remove().draw();
    selectedNodeID = selectedIP = "";
    selectedCount  = 0;
  }
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

    {#if selectedCount > 0 }
      <GradientButton
        shadow
        color="cyan"
        type="button"
        on:click={copy}
        size="xs"
      >
        {#if copied}
          <Icon path={icons.mdiCheck} size={1} />
        {:else}
          <Icon path={icons.mdiContentCopy} size={1} />
        {/if}
        Copy
      </GradientButton>
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={deleteAddress}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        {$_('Address.Delete')}
      </GradientButton>
    {/if}
    {#if arp.length > 0}
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
    {/if}
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

<AddressReport
  bind:show ={showReport}
  {arp}
  {changeIP}
  {changeMAC}
/>

<Node
  bind:show={showAddNode}
  ip={selectedIP}
  posX={100}
  posY={120}
  on:close={(e) => {
    refresh();
  }}
/>

<Node
  bind:show={showEditNode}
  nodeID={selectedNodeID}
  on:close={(e) => {
    refresh();
  }}
/>

<NodeReport
  bind:show={showNodeReport}
  id={selectedNodeID}
/>

