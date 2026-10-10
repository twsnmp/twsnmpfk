<script lang="ts">
  import { createEventDispatcher, tick } from "svelte";
  import {
    Modal,
    GradientButton,
    Search,
    Select,
    Input,
    Badge,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNode,
    GNMIGet,
    GNMICapabilities,
    ExportAny,
    GetDefaultPolling,
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import Polling from "./Polling.svelte";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import Neko from "./Neko.svelte";
  import { copyText } from "svelte-copy";

  export let show: boolean = false;
  export let nodeID = "";

  let nodeName = "";
  let path = "/interfaces";
  let target = "";
  let encoding = "json_ietf";
  const encodingOptions = [
    { value: "json_ietf", name: "json_ietf" },
    { value: "json", name: "json" },
    { value: "proto", name: "proto" },
    { value: "bytes", name: "bytes" },
    { value: "ascii", name: "ascii" },
  ];
  let history: any = [
    { value: "/interfaces", name: "/interfaces" },
    { value: "/system", name: "/system" },
    { value: "/components", name: "/components" },
  ];
  const shortcutPaths = [
    "/interfaces",
    "/system",
    "/components",
    "/network-instances",
    "/platform",
    "/acl",
    "/qos",
    "/lldp",
  ];
  let selected = "";
  let showNeko = false;
  let nekoStatus: "" | "waiting" | "ok" | "ng" = "";
  let timer: any = undefined;
  let table: any = undefined;
  let data: any = [];
  let selectedCount = 0;

  let showCap = false;
  let capTable: any = undefined;
  let capData: any = {};
  let capModelFilter = "";

  let showHelp = false;

  const dispatch = createEventDispatcher();

  const waitAnimation = (status: "" | "waiting" | "ok" | "ng") => {
    nekoStatus = status;
    if (status === "waiting") {
      showNeko = true;
    } else {
      setTimeout(() => {
        showNeko = false;
      }, 1500);
    }
  };

  const onOpen = async () => {
    const node = await GetNode(nodeID);
    if (node) {
      nodeName = node.Name || "";
      if (node.GNMIPort) {
        target = node.IP + ":" + node.GNMIPort;
      } else {
        target = node.IP + ":57400";
      }
      encoding = node.GNMIEncoding ? node.GNMIEncoding : "json_ietf";
    }
    data = [];
    selectedCount = 0;
  };

  const columns = [
    {
      data: "Path",
      title: $_("GNMITool.Path") || "Path",
      width: "55%",
      className: "dt-nowrap",
    },
    {
      data: "Index",
      title: "Index",
      width: "10%",
    },
    {
      data: "Value",
      title: $_("MIBBrowser.Value"),
      width: "35%",
    },
  ];

  const showTable = () => {
    selectedCount = 0;
    table = new DataTable("#gnmiGetTable", {
      destroy: true,
      columns: columns,
      stateSave: true,
      data: data,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order: [[0, "asc"]],
      language: getTableLang(),
      select: {
        style: "multi+shift",
      },
    });
    table.on("select deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("dblclick", "tr", function (this: HTMLElement) {
      const rowData = table.row(this).data();
      if (rowData && rowData.Path) {
        path = rowData.Path;
      }
    });
  };

  const get = async () => {
    if (!path) return;
    waitAnimation("waiting");
    data = await GNMIGet(nodeID, target, path, encoding);
    if (!data) {
      waitAnimation("ng");
    } else {
      updateHistory();
      waitAnimation("ok");
      await tick();
      showTable();
    }
  };

  const capColumns = [
    {
      data: "name",
      title: $_("GNMITool.Name"),
      width: "55%",
    },
    {
      data: "organization",
      title: $_("GNMITool.Org"),
      width: "30%",
    },
    {
      data: "version",
      title: $_("GNMITool.Version"),
      width: "15%",
    },
  ];

  const showCapabilities = () => {
    const models = capData?.Models || [];
    const q = capModelFilter.trim().toLowerCase();
    const filtered = q
      ? models.filter(
          (m: any) =>
            (m.name && m.name.toLowerCase().includes(q)) ||
            (m.organization && m.organization.toLowerCase().includes(q)) ||
            (m.version && m.version.toLowerCase().includes(q))
        )
      : models;

    capTable = new DataTable("#gnmiCapTable", {
      destroy: true,
      columns: capColumns,
      data: filtered,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order: [[0, "asc"]],
      language: getTableLang(),
    });
  };

  const cap = async () => {
    waitAnimation("waiting");
    capData = await GNMICapabilities(nodeID, target);
    if (!capData) {
      waitAnimation("ng");
    } else {
      waitAnimation("ok");
      capModelFilter = "";
      showCap = true;
      await tick();
      showCapabilities();
    }
  };

  const onFilterCapChange = () => {
    showCapabilities();
  };

  const updateHistory = () => {
    if (!path) return;
    const tmp = history.filter((h: any) => h.value !== path);
    tmp.unshift({
      name: path,
      value: path,
    });
    history = tmp.slice(0, 10);
  };

  const close = () => {
    show = false;
  };

  const exportMIB = (t: string) => {
    const ed: any = {
      Title: "TWSNMP gNMI(" + path + ")",
      Header: columns.map((e: any) => e.title),
      Data: [],
      Image: "",
    };
    for (const d of data) {
      const row: any = [];
      for (const c of columns) {
        row.push(d[c.data] || "");
      }
      ed.Data.push(row);
    }
    ExportAny(t, ed);
  };

  let copied = false;

  const copy = () => {
    const selectedRows = table?.rows({ selected: true }).data();
    const rowsToCopy =
      selectedRows && selectedRows.length > 0 ? selectedRows : data;
    if (!rowsToCopy || rowsToCopy.length === 0) return;

    let s: string[] = [];
    const h = columns.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < rowsToCopy.length; i++) {
      const row: any = [];
      for (const c of columns) {
        row.push(rowsToCopy[i][c.data] || "");
      }
      s.push(row.join("\t"));
    }
    copyText(s.join("\n"));
    copied = true;
    setTimeout(() => (copied = false), 2000);
  };

  let showPolling = false;
  let pollingTmp: any = undefined;

  const addPolling = async () => {
    const d = table.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    pollingTmp = await GetDefaultPolling(nodeID);
    pollingTmp.Name = "gNMI get " + d[0].Path;
    pollingTmp.Type = "gnmi";
    pollingTmp.Mode = "get";
    pollingTmp.Level = "low";
    pollingTmp.Params = target;
    pollingTmp.Filter = d[0].Path;
    pollingTmp.Script = `
var value = JSON.parse(data);
value == "${d[0].Value}";`;
    showPolling = true;
  };

  let prevShow = false;
  $: {
    if (show && !prevShow) {
      onOpen();
    } else if (!show && prevShow) {
      if (timer) {
        clearTimeout(timer);
        timer = undefined;
      }
      dispatch("close", {});
    }
    prevShow = show;
  }
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  outsideclose={false}
  class="w-full"
>
  <div class="flex flex-col space-y-2.5 min-h-[70vh] max-h-[85vh]">
    <div class="flex items-center justify-between pb-1 border-b border-gray-200 dark:border-gray-700">
      <div class="flex items-center gap-2">
        <Icon path={icons.mdiRadioTower} size={1} class="text-cyan-500" />
        <span class="font-bold text-base text-gray-800 dark:text-gray-200">
          {$_("GNMITool.gNMITool")}{nodeName ? ` - ${nodeName}` : ""}
        </span>
      </div>
      <div class="flex items-center gap-2">
        <span class="text-[11px] text-gray-400">
          {$_("GNMITool.DblClickTip")}
        </span>
        {#if data && data.length > 0}
          <Badge color="indigo">
            {$_("GNMITool.ResultCount", { values: { count: data.length } })}
          </Badge>
        {/if}
      </div>
    </div>

    <!-- Controls Row with unified height -->
    <div class="flex flex-row items-center gap-2">
      <div class="w-52 shrink-0">
        <Input
          size="sm"
          class="!h-9 text-xs"
          type="text"
          bind:value={target}
          placeholder={$_("GNMITool.Target")}
        />
      </div>
      <div class="w-36 shrink-0">
        <Select
          size="sm"
          class="!h-9 text-xs"
          items={encodingOptions}
          bind:value={encoding}
          placeholder={$_("GNMITool.Encoding")}
        />
      </div>
      <div class="flex-1 min-w-[200px]">
        <Search
          size="sm"
          class="!h-9 text-xs"
          bind:value={path}
          placeholder={$_("GNMITool.Path")}
          onkeydown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              get();
            }
          }}
        />
      </div>
      <div class="w-44 shrink-0">
        <Select
          size="sm"
          class="!h-9 text-xs"
          items={history}
          bind:value={selected}
          placeholder={$_("GNMITool.History")}
          onchange={() => {
            if (selected) path = selected;
          }}
        />
      </div>
    </div>

    <!-- Shortcuts row for quick Path selection -->
    <div class="flex items-center gap-1.5 overflow-x-auto text-xs py-1 scrollbar-thin">
      <div class="flex items-center gap-1 text-[11px] font-semibold text-gray-700 dark:text-gray-300 shrink-0">
        <Icon path={icons.mdiLightningBolt} size={0.75} class="text-amber-500" />
        <span>{$_("GNMITool.Shortcuts")}:</span>
      </div>
      {#each shortcutPaths as p}
        <button
          type="button"
          onclick={() => {
            path = p;
          }}
          ondblclick={() => {
            path = p;
            get();
          }}
          class="px-2.5 py-0.5 rounded-md text-[11px] font-mono font-semibold bg-cyan-50 dark:bg-cyan-950/60 border border-cyan-200 dark:border-cyan-700/80 text-cyan-600 dark:text-cyan-300 hover:bg-cyan-100 dark:hover:bg-cyan-900 hover:text-cyan-800 dark:hover:text-cyan-100 hover:border-cyan-400 dark:hover:border-cyan-500 transition-colors shadow-xs shrink-0 cursor-pointer"
          title={`${p} (Click to set, Double click to Get)`}
        >
          {p}
        </button>
      {/each}
    </div>

    <!-- DataTables Area wrapped in div to follow AGENTS.md rule -->
    <div id="gnmiGetTableWrapper" class="flex-1 overflow-auto">
      <table id="gnmiGetTable" class="display compact" style="width:99%"></table>
    </div>

    <!-- Bottom Buttons Row -->
    <div class="flex justify-end items-center space-x-2 mr-2 pt-1">
      {#if nekoStatus != "waiting"}
        {#if data && data.length > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            onclick={copy}
            size="xs"
          >
            {#if copied}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            {selectedCount > 0
              ? $_("GNMITool.CopySelected", { values: { count: selectedCount } })
              : $_("GNMITool.CopyAll")}
          </GradientButton>
        {/if}
        {#if selectedCount == 1}
          <GradientButton
            shadow
            color="blue"
            type="button"
            onclick={addPolling}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          color="cyan"
          type="button"
          onclick={cap}
          size="xs"
        >
          <Icon path={icons.mdiInformationOutline} size={1} />
          {$_("GNMITool.Capabilities") || "Capabilities"}
        </GradientButton>
        <GradientButton
          shadow
          color="cyan"
          type="button"
          onclick={() => {
            BrowserOpenURL("https://github.com/YangModels/yang");
          }}
          size="xs"
        >
          <Icon path={icons.mdiOpenInNew} size={1} />
          {$_("GNMITool.YangInfo")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="blue"
          onclick={get}
          size="xs"
        >
          <Icon path={icons.mdiPlay} size={1} />
          {$_("MIBBrowser.Get")}
        </GradientButton>
        {#if data && data.length > 0}
          <GradientButton
            shadow
            color="lime"
            type="button"
            onclick={() => {
              exportMIB("csv");
            }}
            size="xs"
          >
            <Icon path={icons.mdiFileDelimited} size={1} />
            CSV
          </GradientButton>
          <GradientButton
            shadow
            color="lime"
            type="button"
            onclick={() => {
              exportMIB("excel");
            }}
            size="xs"
          >
            <Icon path={icons.mdiFileExcel} size={1} />
            Excel
          </GradientButton>
        {/if}
      {/if}
      <GradientButton
        shadow
        type="button"
        size="xs"
        color="lime"
        class="ml-2"
        onclick={() => {
          showHelp = true;
        }}
      >
        <Icon path={icons.mdiHelp} size={1} />
        <span>
          {$_("DrawItem.Help")}
        </span>
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Neko bind:show={showNeko} status={nekoStatus} />

<Modal
  bind:open={showCap}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <div class="flex items-center justify-between border-b pb-2 dark:border-gray-700">
      <div class="flex items-center gap-2">
        <Icon path={icons.mdiInformationOutline} size={1} class="text-cyan-500" />
        <span class="font-bold text-lg text-gray-800 dark:text-gray-200">
          {$_("GNMITool.CapabilitiesTitle") || "Capabilities"}
        </span>
      </div>
      <Badge color="blue">
        Version: {capData?.Version || "N/A"}
      </Badge>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2 text-xs">
      <div class="flex items-center gap-2">
        <span class="font-semibold text-gray-700 dark:text-gray-300">
          {$_("GNMITool.Encodings")}:
        </span>
        <span class="font-mono text-cyan-600 dark:text-cyan-400">
          {capData?.Encodings || "N/A"}
        </span>
      </div>
      <div class="w-64">
        <Search
          size="sm"
          class="!h-9 text-xs"
          bind:value={capModelFilter}
          oninput={onFilterCapChange}
          placeholder={$_("GNMITool.SearchModels")}
        />
      </div>
    </div>

    <!-- Capabilities DataTables Area wrapped in div to follow AGENTS.md rule -->
    <div id="gnmiCapTableWrapper" class="flex-1 overflow-auto max-h-[55vh]">
      <table id="gnmiCapTable" class="display compact" style="width:99%"></table>
    </div>

    <div class="flex justify-end space-x-2 mr-2 pt-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => {
          showCap = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Polling bind:show={showPolling} {pollingTmp} />

<Help bind:show={showHelp} page="gnmitool" />
