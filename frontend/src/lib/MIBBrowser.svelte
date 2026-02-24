<script lang="ts">
  import neko_ng from "../assets/images/neko_ng.png";
  import neko_ok from "../assets/images/neko_ok.png";
  import neko1 from "../assets/images/neko_anm1.png";
  import neko2 from "../assets/images/neko_anm2.png";
  import neko3 from "../assets/images/neko_anm3.png";
  import neko4 from "../assets/images/neko_anm4.png";
  import neko5 from "../assets/images/neko_anm5.png";
  import neko6 from "../assets/images/neko_anm6.png";
  import neko7 from "../assets/images/neko_anm7.png";
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
  import { createEventDispatcher } from "svelte";
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
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import MibTree from "./MIBTree.svelte";
  import Polling from "./Polling.svelte";
  import AskLLMDailog from "./AskLLMDialog.svelte";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { copyText } from "svelte-copy";
  import { tick } from "svelte";

  export let show: boolean = false;
  export let nodeID = "";

  let name = "";
  let raw = false;
  let history: any = [];
  let selected = "";
  let wait = false;
  let neko = neko1;
  let showNeko = false;
  let nekoNo = 0;
  const nekos: any = [];
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
    mibTree.children = await GetMIBTree();
    data = [];
    resultMibTree = {};
    if (nekos.length === 0) {
      nekos.push(neko1);
      nekos.push(neko2);
      nekos.push(neko3);
      nekos.push(neko4);
      nekos.push(neko5);
      nekos.push(neko6);
      nekos.push(neko7);
    }
  };

  const showTable = () => {
    if (table && DataTable.isDataTable("#mibTable")) {
      table.clear();
      table.destroy(true);
      table = undefined;
      const e = document.getElementById("mibbrTable");
      if (e) {
        e.innerHTML = `<table id="mibTable" class="display compact" style="width:99%" />`;
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
        style: "multi",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
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
    wait = true;
    waitAnimation();
    mibs = await SnmpWalk(nodeID, name, raw);
    if (!mibs) {
      neko = neko_ng;
    } else {
      updateHistory();
      resultMibTree = [];
      neko = neko_ok;
      refreshTable();
    }
    wait = false;
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
        title: e,
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

  const waitAnimation = () => {
    if (!wait) {
      setTimeout(() => {
        showNeko = false;
      }, 2000);
      return;
    }
    showNeko = true;
    neko = nekos[nekoNo];
    nekoNo++;
    if (nekoNo >= nekos.length) {
      nekoNo = 0;
    }
    timer = setTimeout(waitAnimation, 200);
  };

  const close = () => {
    show = false;
    dispatch("close", {});
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
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
    wait = true;
    waitAnimation();
    const r = await LLMMIBSearch(mibSearchPrompt);
    wait = false;
    if (r.Error != "") {
      llmMIBsearchError = r.Error;
      return;
    }
    name = r.ObjectName;
    showLLMMIBSearch = false;
    showMIBTree = false;
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
    wait = true;
    waitAnimation();
    const r = await LLMAskMIB(a.join("\n"));
    wait = false;
    askLLMDialog = true;
    if (r.Error != "") {
      askLLMError = r.Error;
      return;
    }
    askLLMResult = r.Results;
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  <div class="flex flex-col space-y-4">
    <div class="flex flex-row mb-2">
      <div class="flex-auto">
        <Search
          size="sm"
          bind:value={name}
          placeholder={$_("MIBBrowser.ObjectName")}
        />
      </div>
      <GradientButton
        shadow
        color="blue"
        size="xs"
        class="ml-2"
        on:click={() => {
          showMIBTree = true;
        }}
      >
        <Icon path={icons.mdiFileTree} size={1} />
      </GradientButton>
      <Select
        size="sm"
        class="ml-2 w-64"
        items={history}
        bind:value={selected}
        placeholder={$_("MIBBrowser.History")}
        on:change={() => {
          name = selected;
        }}
      />
    </div>
    <div id="mibbrTable">
      <table id="mibTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if !wait}
        {#if selectedCount > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={copyMIB}
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
            on:click={addPolling}
            size="xs"
          >
            <Icon path={icons.mdiEye} size={1} />
            {$_("NodeReport.Polling")}
          </GradientButton>
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={showSetDialog}
            size="xs"
          >
            <Icon path={icons.mdiSend} size={1} />
            SET
          </GradientButton>
        {/if}
        <Toggle bind:checked={scalar} on:change={refreshTable}
          >{$_("MIBBrowser.ScalarOnly")}</Toggle
        >
        <Toggle bind:checked={raw}>{$_("MIBBrowser.RawData")}</Toggle>
        <GradientButton
          shadow
          type="button"
          color="blue"
          on:click={get}
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
            on:click={() => {
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
            on:click={() => {
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
            on:click={() => {
              exportMIB("excel");
            }}
            size="xs"
          >
            <Icon path={icons.mdiFileExcel} size={1} />
            Excel
          </GradientButton>
          <GradientButton
            shadow
            color="pink"
            type="button"
            on:click={askLLM}
            size="xs"
          >
            <Icon path={icons.mdiBrain} size={1} />
            {$_('MIBBrowser.AIExprain')}
          </GradientButton>
        {/if}
      {/if}
      <GradientButton
        shadow
        type="button"
        size="xs"
        color="lime"
        class="ml-2"
        on:click={() => {
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
        on:click={close}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showMIBTree}
  size="lg"
  dismissable={false}
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
        />
      {/if}
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="pink"
        on:click={() => {
          showLLMMIBSearch = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiBrain} size={1} />
        {$_('MIBBrowser.AskAI')}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
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
        on:click={llmMIBSearch}
        size="xs"
      >
        <Icon path={icons.mdiPlay} size={1} />
        {$_('MIBBrowser.Search')}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
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
          on:click={() => {
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
          />
        {/if}
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if missingList}
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={showMissingDialog}
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
          on:click={() => {
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
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <table id="missingTable" class="display compact" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      {#if selectedMissingCount == 1}
        <GradientButton
          shadow
          color="lime"
          type="button"
          on:click={searchExtMIB}
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
        on:click={() => (showMissing = false)}
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
        on:click={doSet}
        size="xs"
      >
        <Icon path={icons.mdiSend} size={1} />
        SET
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
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

<AskLLMDailog bind:show={askLLMDialog} content={askLLMResult}  error={askLLMError}/>

<Polling bind:show={showPolling} {pollingTmp} />

<Help bind:show={showHelp} page="mibbrowser" />

<Modal
  bind:open={showNeko}
  size="sm"
  dismissable={false}
  class="w-full bg-white bg-opacity-75 dark:bg-white"
>
  <div class="flex justify-center items-center">
    <img src={neko} alt="neko" />
  </div>
</Modal>

<style>
  #mibtree {
    height: 70vh;
    overflow: scroll;
  }
</style>
