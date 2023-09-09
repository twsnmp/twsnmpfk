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
  import { Modal, Button, Search, Select,Toggle } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { GetMIBTree, GetNode, SnmpWalk } from "../../wailsjs/go/main/App";
  import { getTableLang } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  export let nodeID = "";

  let show: boolean = false;
  let name = "";
  let raw = false;
  let history = [];
  let selected = "";
  let wait = false;
  let neko = neko1;
  let showNeko = false;
  let nekoNo = 0;
  const nekos = [];
  let timer = undefined;
  let table = undefined;
  let columns = [];
  let data = [];
  let selectedCount = 0;

  const dispatch = createEventDispatcher();

  onMount(async () => {
    const mibTree = await GetMIBTree();
    console.log(mibTree);
    const node = await GetNode(nodeID);
    show = true;
    nekos.push(neko1);
    nekos.push(neko2);
    nekos.push(neko3);
    nekos.push(neko4);
    nekos.push(neko5);
    nekos.push(neko6);
    nekos.push(neko7);
  });

  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  });

  const showTable = () => {
    if (table) {
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#mibTable", {
      columns: columns,
      data: data,
      language: getTableLang(),
      select: {
        style: "single",
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
      title: "オブジェクト名",
      width: "30%",
    },
    {
      data: "Value",
      title: "値",
      width: "60%",
    },
  ];

  const get = async () => {
    wait = true;
    waitAnimation();
    const mibs = await SnmpWalk(nodeID, name, raw);
    if (!mibs) {
      neko = neko_ng;
    } else {
      updateHistory();
      neko = neko_ok;
      if (name.endsWith("Table")) {
        tableMIBData(mibs);
        console.log(mibs);
        console.log(columns);
        console.log(data);
      } else {
        columns = basicColumns;
        data = [];
        let i = 1;
        mibs.forEach((e) => {
          data.push({
            Index: i,
            Name: e.Name,
            Value: e.Value,
          });
          i++;
        });
      }
      showTable();
    }
    wait = false;
  };

  const updateHistory = () => {
    const tmp = [];
    for(const h of history) {
      if (h.value != name) {
        tmp.push(h);
      }
    }
    tmp.unshift({
      name:name,
      value:name,
    });
    history = tmp;
  }

  const tableMIBData = (mibs) => {
    const names = [];
    const indexes = [];
    const rows = [];
    mibs.forEach((e) => {
      const name = e.Name;
      const val = e.Value;
      const i = name.indexOf(".");
      if (i > 0) {
        const base = name.substring(0, i);
        const index = name.substring(i + 1);
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
    names.forEach((e) => {
      columns.push({
        title: e,
        data: e,
      });
    });
    data = [];
    let i = 1;
    rows.forEach((e) => {
      const d = { Index: i };
      for (let i = 1; i < e.length; i++) {
        d[names[i - 1]] = e[i];
      }
      data.push(d);
      i++;
    });
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
  };
</script>

<Modal bind:open={show} size="xl" permanent class="w-full" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <div class="flex flex-row mb-2">
      <div class="flex-auto">
        <Search size="sm" bind:value={name} placeholder="オブジェクト名">
        </Search>
      </div>
      <Button size="sm" class="ml-2">
        <Icon path={icons.mdiFileTree} size={1} />
      </Button>
      <Select
        size="sm"
        class="ml-2 w-64"
        items={history}
        bind:value={selected}
        on:change={()=>{
          name = selected;
        }}
        placeholder="履歴"
      />
    </div>
    <table id="mibTable" class="display compact" style="width:99%" />
    <div class="flex justify-end space-x-2 mr-2">
      {#if !wait}
        <Toggle bind:checked={raw}>生データ</Toggle>
        <Button type="button" color="blue" on:click={get} size="sm">
          <Icon path={icons.mdiPlay} size={1} />
          取得
        </Button>
      {/if}
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showNeko}
  size="sm"
  permanent
  class="w-full bg-white bg-opacity-75 dark:bg-white"
>
  <div class="flex justify-center items-center">
    <img src={neko} alt="neko" />
  </div>
</Modal>
