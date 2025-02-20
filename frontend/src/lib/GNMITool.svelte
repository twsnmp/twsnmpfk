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
  import { createEventDispatcher } from "svelte";
  import {
    Modal,
    GradientButton,
    Search,
    Select,
    Input,
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
  import { copyText } from "svelte-copy";
  import { tick } from "svelte";

  export let show: boolean = false;
  export let nodeID = "";

  let path = "";
  let target = "";
  let encoding = "";
  let history: any = [];
  let selected = "";
  let wait = false;
  let neko = neko1;
  let showNeko = false;
  let nekoNo = 0;
  const nekos: any = [];
  let timer: any = undefined;
  let table: any = undefined;
  let data: any = [];
  let selectedCount = 0;
  
  let showCap = false;
  let capTable: any = undefined;
  let capData : any = [];
  
  let showHelp = false;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    const node = await GetNode(nodeID);
    if (node) {
      if (node.GNMIPort) { 
        target = node.IP + ":" + node.GNMIPort;
      } else {
        target = node.IP + ":57400";
      }
      encoding = node.GNMIEncoding ? node.GNMIEncoding : "json_ietf"; 
    }
    data = [];
    nekos.push(neko1);
    nekos.push(neko2);
    nekos.push(neko3);
    nekos.push(neko4);
    nekos.push(neko5);
    nekos.push(neko6);
    nekos.push(neko7);
  };

  const columns = [
    {
      data: "Path",
      title: "Path",
      width: "60%",
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
      width: "30%",
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
      order:[[0,"desc"]],
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

  const get = async () => {
    wait = true;
    waitAnimation();
    data = await GNMIGet(nodeID,target, path, encoding);
    if (!data) {
      neko = neko_ng;
    } else {
      updateHistory();
      neko = neko_ok;
      showTable();
    }
    wait = false;
  };

  const capColumns = [
    {
      data: "name",
      title: $_('GNMITool.Name'),
      width: "60%",
    },
    {
      data: "organization",
      title: $_('GNMITool.Org'),
      width: "30%",
    },
    {
      data: "version",
      title: $_('GNMITool.Version'),
      width: "10%",
    },
  ];


  const showCapabilities = async () => {
    console.log(capData.Models);
    capTable = new DataTable("#gnmiCapTable", {
      destroy: true,
      columns: capColumns,
      data: capData.Models,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      order:[[0,"desc"]],
      language: getTableLang(),
    });
  }

  const cap = async () => {
    wait = true;
    waitAnimation();
    capData = await GNMICapabilities(nodeID, target);
    if (!capData) {
      neko = neko_ng;
    } else {
      neko = neko_ok;
      showCap = true;
      await tick();
      showCapabilities();
    }
    wait = false;
  };

  const updateHistory = () => {
    const tmp = [];
    for (const h of history) {
      if (h.value != path) {
        tmp.push(h);
      }
    }
    tmp.unshift({
      path: path,
      value: path,
    });
    history = tmp;
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
      Title: "TWSNMP gMMI(" + path + ")",
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
      <Input
        size="sm"
        class="mr-2 w-96 h-8"
        type="text"
        bind:value={target}
        placeholder={$_('GNMITool.Target')}
      />
      <Input
        size="sm"
        class="mr-2 w-96 h-8"
        type="text"
        bind:value={encoding}
        placeholder="Encoding"
      />
      <Search
        size="sm"
        bind:value={path}
        placeholder="Path"
      />
      <Select
        size="sm"
        class="ml-2 w-64"
        items={history}
        bind:value={selected}
        placeholder={$_("MIBBrowser.History")}
        on:change={() => {
          path = selected;
        }}
      />
    </div>
    <table id="gnmiGetTable" class="display compact" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      {#if !wait}
        {#if selectedCount > 0}
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
        {/if}
        {#if selectedCount == 1}
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
        <GradientButton
          shadow
          color="cyan"
          type="button"
          on:click={cap}
          size="xs"
        >
          <Icon path={icons.mdiTree} size={1} />
          Capabilities
        </GradientButton>
        <GradientButton
          shadow
          color="cyan"
          type="button"
          on:click={()=>{
            BrowserOpenURL("https://github.com/YangModels/yang");
          }}
          size="xs"
        >
          <Icon path={icons.mdiTree} size={1} />
          {$_('GNMITool.YangInfo')}
        </GradientButton>
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
        {#if selectedCount > 0}
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
  bind:open={showCap}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <div>
      <p>gNMI Version:{ capData.Version }</p>
      <p>Encodings:{ capData.Encodings }</p>
    </div>
    <table id="gnmiCapTable" class="display compact" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
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

