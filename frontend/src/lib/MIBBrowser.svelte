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
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetMIBTree,
    SnmpWalk,
    ExportAny,
    GetDefaultPolling,
  } from "../../wailsjs/go/main/App";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import MibTree from "./MIBTree.svelte";
  import Polling from "./Polling.svelte";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";
  import { copyText } from "svelte-copy";

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
  let mibTree: any = {
    oid: ".1.3.6.1",
    name: ".iso.org.dod.internet",
    MIBInfo: null,
    children: undefined,
  };
  let showHelp = false;
  let isTable = false;

  let showResultMIBTree = false;
  let resultMibTree: any = {};
  let resultMibTreeWait = false;
  let resultMibTreeProgress :string = "0";
  let stopResultMibTree = false;

  const onOpen = async () => {
    mibTree.children = await GetMIBTree();
    data = [];
    resultMibTree = {};
    nekos.push(neko1);
    nekos.push(neko2);
    nekos.push(neko3);
    nekos.push(neko4);
    nekos.push(neko5);
    nekos.push(neko6);
    nekos.push(neko7);
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
    stopResultMibTree = false
    resultMibTreeWait = true
    resultMibTreeProgress = "0"
    let i = 0
    const nameMap = new Map()
    for (const e of mibs) {
      if (stopResultMibTree) {
        break
      }
      if (i % 500 === 0) {
        await new Promise((resolve) => {
          setTimeout(resolve, 0)
        })
        resultMibTreeProgress = ((100.0 * i) / mibs.length).toFixed(2)
      }
        i++
      const a = e.Name.split(".", 2);
      const t = getTreePath(a[0], mibTree.children);
      if (t) {
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
        oid: ".1.3.6.1",
        name: ".iso.org.dod.internet",
        MIBInfo: null,
        children: c,
      };
    }
    resultMibTreeWait = false
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
              if (!resultMibTree || resultMibTree.length == 0 || stopResultMibTree) {
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
  bind:open={showNeko}
  size="sm"
  dismissable={false}
  class="w-full bg-white bg-opacity-75 dark:bg-white"
>
  <div class="flex justify-center items-center">
    <img src={neko} alt="neko" />
  </div>
</Modal>

<Modal
  bind:open={showMIBTree}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <div id="mibtree">
      <MibTree
        tree={mibTree}
        on:select={(e) => {
          name = e.detail;
          showMIBTree = false;
        }}
      />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
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
  bind:open={showResultMIBTree}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
  {#if resultMibTreeWait}
    <Progressbar progress={resultMibTreeProgress} labelOutside={$_('MIBBrowser.An')} />
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
        {$_('MIBBrowser.Stop')}
      </GradientButton>
    </div>
  {:else}
    <div id="mibtree">
      <MibTree
        tree={resultMibTree}
        on:select={(e) => {
          name = e.detail;
          showResultMIBTree = false;
          get();
        }}
      />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
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

<Polling bind:show={showPolling} {pollingTmp} />

<Help bind:show={showHelp} page="mibbrowser" />

<style>
  #mibTable {
    height: 70vh;
  }
  #mibtree {
    height: 70vh;
    overflow: scroll;
  }
</style>
