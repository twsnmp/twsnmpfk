<script lang="ts">
  import {
    Modal,
    GradientButton,
    Search,
    Select,
    Toggle,
    Progressbar,
    Alert,
    Input,
    Label,
  } from "flowbite-svelte";
  import { createEventDispatcher, onMount, tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetMIBTree,
    SnmpWalk,
    ExportAny,
    GetDefaultPolling,
    SnmpSet,
    LLMMIBSearch,
    LLMAskMIB,
    GetMapConf,
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import MibTree from "./MIBTree.svelte";
  import Polling from "./Polling.svelte";
  import AskLLMDialog from "./AskLLMDialog.svelte";
  import Neko from "./Neko.svelte";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { copyText } from "svelte-copy";

  export let show: boolean = false;
  export let nodeID = "";

  let hasAI = false;
  const checkAI = async () => {
    const conf = await GetMapConf();
    hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
  };

  const shortcutMibs = [
    "system",
    "interfaces",
    "ifTable",
    "ifXTable",
    "ip",
    "ipAddrTable",
    "hrStorageTable",
    "hrDeviceTable",
  ];

  let name = "";
  let raw = false;
  let history: any = [];
  let selected = "";
  let showNeko = false;
  let nekoStatus: "" | "waiting" | "ok" | "ng" = "";
  let timer: any = undefined;
  let table: any = undefined;
  let columns: any = [];
  let data: any = [];
  let selectedCount = 0;
  let showMIBTree = false;
  let mibTreeFilter = "";
  let mibTree: any = {
    oid: ".1",
    name: ".iso",
    MIBInfo: null,
    children: undefined,
  };

  // Hover Tooltip state
  let hoveredInfo: {
    name: string;
    oid: string;
    type?: string;
    tooltip: string;
    x: number;
    y: number;
  } | null = null;

  // Cache mapping MIB object name or OID to MIB info
  const mibInfoCache = new Map<string, { oid: string; type?: string; description: string }>();

  const buildMibInfoCache = (nodes: any[]) => {
    if (!nodes || !Array.isArray(nodes)) return;
    for (const n of nodes) {
      const nodeName = n.name || n.Name || "";
      const oid = n.oid || n.OID || "";
      const mibInfo = n.MIBInfo || n.mibInfo;
      if (nodeName && mibInfo) {
        const desc = mibInfo.Description || mibInfo.description || "";
        const type = mibInfo.Type || mibInfo.type || "";
        if (desc) {
          const entry = { oid, type, description: desc };
          mibInfoCache.set(nodeName, entry);
          if (oid) {
            mibInfoCache.set(oid, entry);
          }
        }
      }
      const children = n.children || n.Children;
      if (children && children.length > 0) {
        buildMibInfoCache(children);
      }
    }
  };

  const getMibInfo = (rawName: string) => {
    if (!rawName) return null;
    let name = rawName.trim();

    // 1. Direct match
    if (mibInfoCache.has(name)) {
      const entry = mibInfoCache.get(name)!;
      if (entry.description) return entry;
    }

    // 2. Strip trailing instance index (e.g. "sysDescr.0" -> "sysDescr", "ifDescr.1" -> "ifDescr")
    const dot = name.indexOf(".");
    if (dot > 0) {
      const base = name.substring(0, dot);
      if (mibInfoCache.has(base)) {
        const entry = mibInfoCache.get(base)!;
        if (entry.description) return entry;
      }
    }

    // 3. Strip "Instance" suffix (e.g. "sysUpTimeInstance" -> "sysUpTime")
    if (name.endsWith("Instance")) {
      const base = name.replace(/Instance$/, "");
      if (mibInfoCache.has(base)) {
        const entry = mibInfoCache.get(base)!;
        if (entry.description) return entry;
      }
    }

    // 4. Match with leading dot for numeric OID (e.g. "1.3.6..." -> ".1.3.6...")
    const withDot = name.startsWith(".") ? name : "." + name;
    if (mibInfoCache.has(withDot)) {
      const entry = mibInfoCache.get(withDot)!;
      if (entry.description) return entry;
    }

    // 5. Numeric OID with instance suffix (e.g. .1.3.6.1.2.1.1.1.0 -> .1.3.6.1.2.1.1.1)
    const lastDot = withDot.lastIndexOf(".");
    if (lastDot > 0) {
      const baseOid = withDot.substring(0, lastDot);
      if (mibInfoCache.has(baseOid)) {
        const entry = mibInfoCache.get(baseOid)!;
        if (entry.description) return entry;
      }
    }

    return null;
  };

  const handleTableMouseOver = (e: MouseEvent) => {
    const target = e.target as HTMLElement | null;
    if (!target) return;

    // 1. Check if hovering over element with [data-name]
    const namedEl = target.closest("[data-name]") as HTMLElement | null;
    if (namedEl) {
      const rawName = namedEl.dataset.name || namedEl.innerText.trim();
      const info = getMibInfo(rawName);
      if (info && info.description) {
        hoveredInfo = {
          name: rawName,
          oid: info.oid,
          type: info.type,
          tooltip: info.description,
          x: e.clientX,
          y: e.clientY,
        };
        return;
      }
    }

    // 2. Check if hovering over th header (in Table mode or list mode)
    const th = target.closest("th");
    if (th) {
      const colName = th.innerText.trim();
      if (colName && colName !== "Index" && colName !== "#") {
        const info = getMibInfo(colName);
        if (info && info.description) {
          hoveredInfo = {
            name: colName,
            oid: info.oid,
            type: info.type,
            tooltip: info.description,
            x: e.clientX,
            y: e.clientY,
          };
          return;
        }
      }
    }

    // 3. Check if hovering over td cell
    const td = target.closest("td") as HTMLTableCellElement | null;
    if (td) {
      const cellText = td.innerText.trim();
      if (cellText) {
        const info = getMibInfo(cellText);
        if (info && info.description) {
          hoveredInfo = {
            name: cellText,
            oid: info.oid,
            type: info.type,
            tooltip: info.description,
            x: e.clientX,
            y: e.clientY,
          };
          return;
        }
      }

      // If in Table mode (e.g. ifTable), cellIndex > 0 maps to columns[cellIndex]
      const cellIndex = td.cellIndex;
      if (isTable && cellIndex !== undefined && cellIndex > 0 && columns && columns[cellIndex]) {
        const colData = columns[cellIndex].data || columns[cellIndex].title;
        if (colData && typeof colData === "string") {
          const info = getMibInfo(colData);
          if (info && info.description) {
            hoveredInfo = {
              name: colData,
              oid: info.oid,
              type: info.type,
              tooltip: info.description,
              x: e.clientX,
              y: e.clientY,
            };
            return;
          }
        }
      }
    }

    hoveredInfo = null;
  };

  const handleTableMouseMove = (e: MouseEvent) => {
    if (hoveredInfo) {
      hoveredInfo = {
        ...hoveredInfo,
        x: e.clientX,
        y: e.clientY,
      };
    }
  };

  const handleTableMouseLeave = () => {
    hoveredInfo = null;
  };

  const loadMibTree = async () => {
    if (mibInfoCache.size === 0) {
      const treeData = await GetMIBTree();
      if (treeData && treeData.length > 0) {
        mibTree.children = treeData;
        mibInfoCache.clear();
        buildMibInfoCache(treeData);
      }
    }
  };

  onMount(async () => {
    await checkAI();
    await loadMibTree();
  });

  const filterMIBTree = (node: any, filter: string): any => {
    if (!node) return null;
    const nameMatch = node.name.toLowerCase().includes(filter);
    const oidMatch = node.oid.includes(filter);
    const match = nameMatch || oidMatch;

    let filteredChildren: any[] = [];
    if (node.children) {
      filteredChildren = node.children
        .map((c: any) => filterMIBTree(c, filter))
        .filter((c: any) => c !== null);
    }

    if (match || filteredChildren.length > 0) {
      return {
        ...node,
        children: filteredChildren,
        forceExpand: filteredChildren.length > 0,
      };
    }
    return null;
  };

  $: filteredMibTree = mibTreeFilter
    ? filterMIBTree(mibTree, mibTreeFilter.toLowerCase())
    : mibTree;

  let showHelp = false;
  let isTable = false;

  let showResultMIBTree = false;
  let resultMibTree: any = {};
  let resultMibTreeFilter = "";
  let resultMibTreeWait = false;
  let resultMibTreeProgress: string = "0";
  let stopResultMibTree = false;
  let showMissing = false;
  let missingList: any = [];
  let selectedMissingCount = 0;

  $: filteredResultMibTree = resultMibTreeFilter
    ? filterMIBTree(resultMibTree, resultMibTreeFilter.toLowerCase())
    : resultMibTree;

  let showSet = false;
  let setError = "";
  let setName = "";
  let setType = "integer";
  let setValue = "";
  const setTypeList = [
    { name: "INTEGER", value: "integer" },
    { name: "String", value: "string" },
  ];
  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    selectedCount = 0;
    await loadMibTree();
    data = [];
    resultMibTree = {};
    hoveredInfo = null;
  };

  const showTable = () => {
    if (table && DataTable.isDataTable("#mibTable")) {
      table.clear();
      table.destroy(true);
      table = undefined;
      const e = document.getElementById("mibbrTable");
      if (e) {
        e.innerHTML = `<table id="mibTable" class="display compact" style="width:99%"></table>`;
      }
    }
    selectedCount = 0;
    table = new DataTable("#mibTable", {
      columns: columns,
      data: data,
      paging: false,
      searching: true,
      info: false,
      scrollY: "65vh",
      language: getTableLang(),
      select: {
        style: "multi+shift",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });

    const wrapper = document.getElementById("mibbrTable");
    if (wrapper) {
      wrapper.onmouseover = handleTableMouseOver;
      wrapper.onmousemove = handleTableMouseMove;
      wrapper.onmouseleave = handleTableMouseLeave;
    }
  };

  const basicColumns = [
    {
      data: "Index",
      title: "Index",
      width: "10%",
    },
    {
      data: "Name",
      title: $_("MIBBrowser.ObjectName"),
      width: "30%",
      render: (data: string) => {
        return `<span class="cursor-pointer hover:text-cyan-500 hover:underline font-medium select-text" data-name="${data}">${data}</span>`;
      },
    },
    {
      data: "Value",
      title: $_("MIBBrowser.Value"),
      width: "60%",
    },
  ];

  let mibs: any = undefined;
  let scalar = false;

  const get = async () => {
    hoveredInfo = null;
    if (mibInfoCache.size === 0) {
      await loadMibTree();
    }
    waitAnimation("waiting");
    mibs = await SnmpWalk(nodeID, name, raw);
    if (!mibs) {
      waitAnimation("ng");
    } else {
      updateHistory();
      resultMibTree = [];
      // Cache MIB info directly from walk results if available
      for (const item of mibs) {
        if (item.Name && item.Description) {
          const entry = {
            oid: item.OID || "",
            type: item.Type || "",
            description: item.Description,
          };
          mibInfoCache.set(item.Name, entry);
          const dot = item.Name.indexOf(".");
          if (dot > 0) {
            mibInfoCache.set(item.Name.substring(0, dot), entry);
          }
          if (item.OID) {
            mibInfoCache.set(item.OID, entry);
          }
        }
      }
      waitAnimation("ok");
      refreshTable();
    }
  };

  const refreshTable = () => {
    if (!mibs) {
      return;
    }
    isTable = name.endsWith("Table");
    if (isTable) {
      tableMIBData();
    } else {
      columns = basicColumns;
      data = [];
      let i = 1;
      mibs.forEach((e: any) => {
        if (
          scalar &&
          (!e.Name.endsWith(".0") || e.Name.split(".").length != 2)
        ) {
          return;
        }
        data.push({
          Index: i,
          Name: e.Name,
          Value: e.Value,
        });
        i++;
      });
    }
    showTable();
  };

  const updateHistory = () => {
    const tmp = [];
    for (const h of history) {
      if (h.value != name) {
        tmp.push(h);
      }
    }
    tmp.unshift({
      name: name,
      value: name,
    });
    history = tmp;
  };

  const tableMIBData = () => {
    const names: any = [];
    const indexes: any = [];
    const rows: any = [];
    mibs.forEach((e: any) => {
      const name = e.Name;
      const val = e.Value;
      const i = name.indexOf(".");
      if (i > 0) {
        const base = name.substring(0, i);
        const index = name.substring(i + 1);
        if (index == "0") {
          return;
        }
        if (!names.includes(base)) {
          names.push(base);
        }
        if (!indexes.includes(index)) {
          indexes.push(index);
          rows.push([index]);
        }
        const r = indexes.indexOf(index);
        if (r >= 0) {
          rows[r].push(val);
        }
      }
    });
    columns = [
      {
        title: "Index",
        data: "Index",
      },
    ];
    names.forEach((e: any) => {
      columns.push({
        title: `<span class="cursor-pointer hover:text-cyan-500 font-semibold" data-name="${e}">${e}</span>`,
        data: e,
      });
    });
    data = [];
    let i = 1;
    rows.forEach((e: any) => {
      const d: any = { Index: i };
      for (let i = 1; i < e.length; i++) {
        d[names[i - 1]] = e[i];
      }
      data.push(d);
      i++;
    });
  };

  const updateResultMibTree = async () => {
    stopResultMibTree = false;
    resultMibTreeWait = true;
    resultMibTreeProgress = "0";
    let i = 0;
    const nameMap = new Map();
    const missingMap = new Map();
    for (const e of mibs) {
      if (stopResultMibTree) {
        break;
      }
      if (i % 500 === 0) {
        await new Promise((resolve) => {
          setTimeout(resolve, 0);
        });
        resultMibTreeProgress = ((100.0 * i) / mibs.length).toFixed(2);
      }
      i++;
      const a = e.Name.split(".");
      if (a.length < 2) {
        continue;
      }
      const t = getTreePath(a[0], mibTree.children);
      if (t) {
        if (hasChildren(t[0], mibTree.children)) {
          const mn = t[0] + "." + a[1];
          if (missingMap.has(mn)) {
            missingMap.set(mn, missingMap.get(mn) + 1);
          } else {
            missingMap.set(mn, 1);
          }
        }
        for (const n of t) {
          if (nameMap.has(n)) {
            nameMap.set(n, nameMap.get(n) + 1);
          } else {
            nameMap.set(n, 1);
          }
        }
      }
    }
    const c = makeTreeData(nameMap, mibTree.children);
    if (c && c.length > 0) {
      resultMibTree = {
        oid: ".1",
        name: ".iso",
        MIBInfo: null,
        children: c,
      };
    }
    resultMibTreeWait = false;
    missingList.length = 0;
    missingMap.forEach((v, k) => {
      missingList.push({
        name: k,
        oid: getOID(k, mibTree.children),
        count: v,
      });
    });
  };

  const hasChildren = (name: string, list: any): boolean => {
    for (let i = 0; i < list.length; i++) {
      if (list[i].name === name) {
        return list[i].children.length > 0;
      }
      if (list[i].children) {
        if (hasChildren(name, list[i].children)) {
          return true;
        }
      }
    }
    return false;
  };

  const getOID = (name: string, list: any): string => {
    const a = name.split(".");
    let idx = "";
    if (a.length > 1) {
      name = a[0];
      idx = "." + a[1];
    }
    for (let i = 0; i < list.length; i++) {
      if (list[i].name === name) {
        return list[i].oid + idx;
      }
      if (list[i].children) {
        const oid = getOID(name, list[i].children);
        if (oid) {
          return oid + idx;
        }
      }
    }
    return "";
  };

  let missingTable: any = undefined;

  const searchExtMIB = () => {
    if (!missingTable) {
      return;
    }
    const d = missingTable.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    let oid = d[0].oid;
    oid = oid.startsWith(".") ? oid.replace(".", "") : oid;
    const url =
      "https://mibbrowser.online/mibdb_search.php?search=" +
      oid +
      "&userdropdown=anymatch";
    BrowserOpenURL(url);
  };

  const missingColumns = [
    {
      data: "name",
      title: $_("MIBBrowser.Name"),
      width: "40%",
    },
    {
      data: "oid",
      title: "OID",
      width: "50%",
    },
    {
      data: "count",
      title: $_("MIBBrowser.Count"),
      width: "10%",
    },
  ];

  const showMissingDialog = async () => {
    showMissing = true;
    await tick();
    missingTable = new DataTable("#missingTable", {
      destroy: true,
      pageLength: 10,
      columns: missingColumns,
      stateSave: true,
      data: missingList,
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    missingTable.on("select", () => {
      selectedMissingCount = missingTable.rows({ selected: true }).count();
    });
    missingTable.on("deselect", () => {
      selectedMissingCount = missingTable.rows({ selected: true }).count();
    });
  };

  const getTreePath = (name: string, list: any): string[] | undefined => {
    for (let i = 0; i < list.length; i++) {
      if (list[i].name === name) {
        return [name];
      }
      if (list[i].children) {
        const n = getTreePath(name, list[i].children);
        if (n) {
          n.push(list[i].name);
          return n;
        }
      }
    }
    return undefined;
  };

  const makeTreeData = (nameMap: any, list: any): any => {
    const r: any = [];
    for (let i = 0; i < list.length; i++) {
      if (nameMap.has(list[i].name)) {
        const e: any = {
          name: list[i].name,
          oid: list[i].oid,
          children: [],
          MIBInfo: list[i].MIBInfo,
          count: nameMap.get(list[i].name),
        };
        if (list[i].children) {
          const cl = makeTreeData(nameMap, list[i].children);
          if (cl) {
            for (const c of cl) {
              e.children.push(c);
            }
          }
        }
        r.push(e);
      }
    }
    return r;
  };

  const waitAnimation = (status: "waiting" | "ok" | "ng") => {
    nekoStatus = status;
    if (status == "waiting") {
      showNeko = true;
    } else {
      setTimeout(() => {
        showNeko = false;
      }, 1000);
    }
  };

  const close = () => {
    show = false;
  };

  const exportMIB = (t: string) => {
    const ed: any = {
      Title: "TWSNMP MIB(" + name + ")",
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

  const copyMIB = () => {
    const selected = table.rows({ selected: true }).data();
    let s: string[] = [];
    const h = columns.map((e: any) => e.title);
    s.push(h.join("\t"));
    for (let i = 0; i < selected.length; i++) {
      const row: any = [];
      for (const c of columns) {
        row.push(selected[i][c.data] || "");
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
    let name = d[0].Name;
    const a = d[0].Name.split(".");
    if (a.length > 0) {
      name = a[0];
    }
    pollingTmp = await GetDefaultPolling(nodeID);
    pollingTmp.Name = "SNMP get " + d[0].Name;
    pollingTmp.Type = "snmp";
    pollingTmp.Mode = "get";
    pollingTmp.Level = "low";
    pollingTmp.Params = d[0].Name;
    pollingTmp.Script = name + "==" + d[0].Value;
    showPolling = true;
  };

  const showSetDialog = () => {
    const d = table.rows({ selected: true }).data();
    if (d.length != 1) {
      return;
    }
    setName = d[0].Name;
    setValue = d[0].Value;
    showSet = true;
  };

  const doSet = async () => {
    setError = await SnmpSet(nodeID, setName, setType, setValue);
    if (setError == "") {
      showSet = false;
      get();
    }
  };

  let showLLMMIBSearch = false;
  let llmMIBsearchError = "";
  let mibSearchPrompt = "";

  const llmMIBSearch = async () => {
    llmMIBsearchError = "";
    waitAnimation("waiting");
    const r = await LLMMIBSearch(mibSearchPrompt);
    if (r.Error != "") {
      waitAnimation("ng");
      llmMIBsearchError = r.Error;
      return;
    }
    name = r.ObjectName;
    showLLMMIBSearch = false;
    showMIBTree = false;
    waitAnimation("ok");
  };
  let askLLMError = "";
  let askLLMResult = "";
  let askLLMDialog = false;

  const askLLM = async () => {
    askLLMError = "";
    askLLMResult = "";
    if (!mibs || mibs.length < 1) {
      return;
    }
    const a: any = [];
    mibs.forEach((e: any) => {
      a.push(e.Name + "=" + e.Value);
    });
    waitAnimation("waiting");
    const r = await LLMAskMIB(a.join("\n"));
    askLLMDialog = true;
    if (r.Error != "") {
      waitAnimation("ng");
      askLLMError = r.Error;
      return;
    }
    askLLMResult = r.Results;
    waitAnimation("ok");
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
      hoveredInfo = null;
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
  <div class="flex flex-col space-y-3 min-h-[70vh] max-h-[85vh]">
    <div class="flex flex-row items-center gap-2">
      <div class="flex-auto">
        <Search
          size="sm"
          bind:value={name}
          placeholder={$_("MIBBrowser.ObjectName")}
          onkeydown={(e) => {
            if (e.key === "Enter") {
              e.preventDefault();
              get();
            }
          }}
        />
      </div>
      <GradientButton
        shadow
        color="blue"
        size="xs"
        class="shrink-0"
        onclick={() => {
          showMIBTree = true;
        }}
      >
        <Icon path={icons.mdiFileTree} size={1} />
      </GradientButton>
      <Select
        size="sm"
        class="w-64 shrink-0"
        items={history}
        bind:value={selected}
        placeholder={$_("MIBBrowser.History")}
        onchange={() => {
          name = selected;
        }}
      />
    </div>

    <!-- Shortcuts row for quick MIB selection (matches GNMITool styling) -->
    <div class="flex items-center gap-1.5 overflow-x-auto text-xs py-1 scrollbar-thin">
      <div class="flex items-center gap-1 text-[11px] font-semibold text-gray-700 dark:text-gray-300 shrink-0">
        <Icon path={icons.mdiLightningBolt} size={0.75} class="text-amber-500" />
        <span>{$_("MIBBrowser.Shortcuts") || "Shortcuts"}:</span>
      </div>
      {#each shortcutMibs as m}
        <button
          type="button"
          onclick={() => {
            name = m;
          }}
          ondblclick={() => {
            name = m;
            get();
          }}
          class="px-2.5 py-0.5 rounded-md text-[11px] font-mono font-semibold bg-cyan-50 dark:bg-cyan-950/60 border border-cyan-200 dark:border-cyan-700/80 text-cyan-600 dark:text-cyan-300 hover:bg-cyan-100 dark:hover:bg-cyan-900 hover:text-cyan-800 dark:hover:text-cyan-100 hover:border-cyan-400 dark:hover:border-cyan-500 transition-colors shadow-xs shrink-0 cursor-pointer"
          title={`${m} (Click to set, Double click to Get)`}
        >
          {m}
        </button>
      {/each}
    </div>

    <!-- DataTables Area wrapped in div to follow AGENTS.md rule with hover listeners -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      id="mibbrTable"
      class="flex-1 overflow-auto"
      onmouseover={handleTableMouseOver}
      onmousemove={handleTableMouseMove}
      onmouseleave={handleTableMouseLeave}
      role="region"
      aria-label="MIB Table"
    >
      <table id="mibTable" class="display compact" style="width:99%"></table>
    </div>

    <div class="flex justify-end space-x-2 mr-2">
      {#if nekoStatus != "waiting"}
        {#if selectedCount > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            onclick={copyMIB}
            size="xs"
          >
            {#if copied}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            Copy
          </GradientButton>
        {/if}
        {#if selectedCount == 1 && !isTable}
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
          <GradientButton
            shadow
            color="red"
            type="button"
            onclick={showSetDialog}
            size="xs"
          >
            <Icon path={icons.mdiSend} size={1} />
            SET
          </GradientButton>
        {/if}
        <Toggle bind:checked={scalar} onchange={refreshTable}
          >{$_("MIBBrowser.ScalarOnly")}</Toggle
        >
        <Toggle bind:checked={raw}>{$_("MIBBrowser.RawData")}</Toggle>
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
        {#if data.length > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            onclick={() => {
              showResultMIBTree = true;
              if (
                !resultMibTree ||
                resultMibTree.length == 0 ||
                stopResultMibTree
              ) {
                updateResultMibTree();
              }
            }}
            size="xs"
          >
            <Icon path={icons.mdiTree} size={1} />
            MIB Tree
          </GradientButton>
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
          {#if hasAI}
            <GradientButton
              shadow
              color="pink"
              type="button"
              onclick={askLLM}
              size="xs"
            >
              <Icon path={icons.mdiBrain} size={1} />
              {$_('MIBBrowser.AIExplain')}
            </GradientButton>
          {/if}
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

  <!-- Hover Tooltip for MIB Object Info rendered inside Modal context -->
  {#if hoveredInfo && hoveredInfo.tooltip}
    <div
      class="pointer-events-none fixed z-[99999] max-w-lg rounded-xl border border-slate-700 bg-slate-900/95 p-3 text-xs text-slate-100 shadow-2xl backdrop-blur-md"
      style="left: {Math.max(10, Math.min(hoveredInfo.x + 16, (typeof window !== 'undefined' ? window.innerWidth - 420 : 500)))}px; top: {Math.max(10, Math.min(hoveredInfo.y + 16, (typeof window !== 'undefined' ? window.innerHeight - 260 : 500)))}px;"
    >
      <div class="flex items-center gap-1.5 font-mono font-bold text-teal-400 pb-1 mb-1 border-b border-slate-700/80">
        <Icon path={icons.mdiFolderTree} size={0.75} class="text-teal-400" />
        <span>{hoveredInfo.name}</span>
        <span class="text-slate-400 font-normal">({hoveredInfo.oid}{hoveredInfo.type ? `:${hoveredInfo.type}` : ""})</span>
      </div>
      <pre class="font-sans whitespace-pre-wrap leading-relaxed text-slate-300 max-h-56 overflow-y-auto text-[11px] select-text">{hoveredInfo.tooltip}</pre>
    </div>
  {/if}
</Modal>

<Modal
  bind:open={showMIBTree}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <Search
      size="sm"
      bind:value={mibTreeFilter}
      placeholder={$_("MIBBrowser.ObjectName")}
    />
    <div id="mibtree">
      {#if filteredMibTree}
        <MibTree
          tree={filteredMibTree}
          on:select={(e) => {
            name = e.detail;
            showMIBTree = false;
          }}
          onselect={(selectedName) => {
            name = selectedName;
            showMIBTree = false;
          }}
        />
      {/if}
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if hasAI}
        <GradientButton
          shadow
          type="button"
          color="pink"
          onclick={() => {
            showLLMMIBSearch = true;
          }}
          size="xs"
        >
          <Icon path={icons.mdiBrain} size={1} />
          {$_('MIBBrowser.AskAI')}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => {
          showMIBTree = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showLLMMIBSearch}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    {#if llmMIBsearchError}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {llmMIBsearchError}
        </div>
      </Alert>
    {/if}
    <Search
      size="sm"
      bind:value={mibSearchPrompt}
      placeholder={$_('MIBBrowser.AskAIQ')}
    />
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="lime"
        onclick={llmMIBSearch}
        size="xs"
      >
        <Icon path={icons.mdiPlay} size={1} />
        {$_('MIBBrowser.Search')}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => {
          showLLMMIBSearch = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_('Map.Cancel')}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showResultMIBTree}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    {#if resultMibTreeWait}
      <Progressbar
        progress={resultMibTreeProgress}
        labelOutside={$_("MIBBrowser.An")}
      />
      <div class="flex justify-end space-x-2 mr-2">
        <GradientButton
          shadow
          type="button"
          color="red"
          onclick={() => {
            stopResultMibTree = true;
          }}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("MIBBrowser.Stop")}
        </GradientButton>
      </div>
    {:else}
      <div id="mibtree">
        <Search
          size="sm"
          bind:value={resultMibTreeFilter}
          placeholder={$_("MIBBrowser.ObjectName")}
        />
        {#if filteredResultMibTree}
          <MibTree
            tree={filteredResultMibTree}
            on:select={(e) => {
              name = e.detail;
              showResultMIBTree = false;
              get();
            }}
            onselect={(selectedName) => {
              name = selectedName;
              showResultMIBTree = false;
              get();
            }}
          />
        {/if}
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if missingList}
          <GradientButton
            shadow
            color="red"
            type="button"
            onclick={showMissingDialog}
            size="xs"
          >
            <Icon path={icons.mdiCheck} size={1} />
            {$_("MIBBrowser.Missing")}
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={() => {
            showResultMIBTree = false;
          }}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("MIBBrowser.Close")}
        </GradientButton>
      </div>
    {/if}
  </div>
</Modal>

<Modal
  bind:open={showMissing}
  size="lg"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <table id="missingTable" class="display compact" style="width:99%"></table>
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedMissingCount == 1}
        <GradientButton
          shadow
          color="lime"
          type="button"
          onclick={searchExtMIB}
          size="xs"
        >
          <Icon path={icons.mdiSearchWeb} size={1} />
          {$_("MIBBrowser.Search")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => (showMissing = false)}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showSet}
  size="md"
  dismissable={false}
  outsideclose={false}
  class="w-full min-h-[20vh]"
>
  <div class="flex flex-col space-y-4">
    {#if setError}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {setError}
        </div>
      </Alert>
    {/if}
    <div class="grid gap-2 grid-cols-4">
      <Label class="col-span-3 space-y-2 text-xs">
        <span>{$_("MIBBrowser.ObjectName")}</span>
        <Input
          class="h-8"
          bind:value={setName}
          placeholder={$_("MIBBrowser.ObjectName")}
          required
          size="sm"
        />
      </Label>
      <Label>
        {$_("MIBBrowser.Type")}
        <Select
          items={setTypeList}
          bind:value={setType}
          placeholder={$_("MIBBrowser.Type")}
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("MIBBrowser.Value")}</span>
      <Input
        class="h-8"
        bind:value={setValue}
        placeholder={$_("MIBBrowser.Value")}
        required
        size="sm"
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="red"
        type="button"
        onclick={doSet}
        size="xs"
      >
        <Icon path={icons.mdiSend} size={1} />
        SET
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => {
          showSet = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<AskLLMDialog bind:show={askLLMDialog} content={askLLMResult} error={askLLMError}/>

<Polling bind:show={showPolling} {pollingTmp} />

<Help bind:show={showHelp} page="mibbrowser" />

<Neko bind:show={showNeko} status={nekoStatus} />

<style>
  #mibtree {
    height: 70vh;
    overflow: scroll;
  }
</style>
