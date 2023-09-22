<script lang="ts">
  import {
    Modal,
    Button,
    Tabs,
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Spinner,
  } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { backend, datastore } from "wailsjs/go/models";
  import {
    GetNode,
    GetVPanelPorts,
    GetVPanelPowerInfo,
    GetEventLogs,
    GetHostResource,
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
  } from "./common";
  import { deleteVPanel, initVPanel, setVPanel } from "./vpanel";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";

  import { showHrBarChart, showHrSummary } from "./chart/hostResource";

  export let id = "";
  let node: datastore.NodeEnt | undefined = undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  let logTable = undefined;
  const showLog = async () => {
    if (logTable) {
      logTable.destroy();
      logTable = undefined;
    }
    logTable = new DataTable("#logTable", {
      data: await GetEventLogs(id),
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
        {
          data: "Level",
          title: "レベル",
          width: "15%",
          render: renderState,
        },
        {
          data: "Time",
          title: "発生日時",
          width: "20%",
          render: renderTime,
        },
        {
          data: "Type",
          title: "種別",
          width: "15%",
        },
        {
          data: "Event",
          title: "イベント",
          width: "50%",
        },
      ],
    });
  };


  let portTable = undefined;
  const showPortTable = (ports) => {
    if (portTable) {
      portTable.destroy();
      portTable = undefined;
    }
    portTable = new DataTable("#portTable", {
      paging: false,
      searching: false,
      info: false,
      scrollY: "180px",
      data: ports,
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
          title: "状態",
          width: "10%",
        },
        {
          data: "Name",
          title: "名前",
          width: "15%",
        },
        {
          data: "Type",
          title: "種別",
          width: "5%",
        },
        {
          data: "MAC",
          title: "MACアドレス",
          width: "15%",
        },
        {
          data: "Speed",
          title: "スピード",
          width: "10%",
          render: renderSpeed,
          className: "dt-body-right",
        },
        {
          data: "OutPacktes",
          title: "送信パケット",
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "OutBytes",
          title: "送信バイト",
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          data: "InPacktes",
          title: "受信パケット",
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "InBytes",
          title: "受信バイト",
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
      ],
    });
  };

  const showVPanel = async () => {
    initVPanel("vpanel");
    const ports = await GetVPanelPorts(id);
    const power = await GetVPanelPowerInfo(id);
    setVPanel(ports, power, 0);
    showPortTable(ports);
  };

  const renderStatus = (s) => {
    switch (s) {
      case "Running":
        return `<span class="text-blue-700">動作中</span>`;
      case "Runnable":
        return `<span class="text-blue-900">動作待ち</span>`;
      case "Testing":
        return "テスト中";
      case "NotRunnable":
        return "起動待";
      case "Invalid":
      case "Down":
        return `<span class="text-red-800">停止</span>`;
    }
    return "不明";
  };

  const renderRate = (r) => {
    if (r < 80.0) {
      return `<span class="text-blue-700">${r.toFixed(2)}</span>`;
    } else if (r < 90.0) {
      return `<span class="text-yellow-700">${r.toFixed(2)}</span>`;
    }
    return `<span class="text-red-700">${r.toFixed(2)}</span>`;
  };

  const renderStorageType = (t) => {
    switch (t) {
      case "hrStorageCompactDisc":
        return "CDドライブ";
      case "hrStorageRemovableDisk":
        return "リムーバブル";
      case "hrStorageFloppyDisk":
        return "フロッピー";
      case "hrStorageRamDisk":
        return "RAMディスク";
      case "hrStorageFlashMemory":
        return "フラッシュメモリ";
      case "hrStorageNetworkDisk":
        return "ネットワーク";
      case "hrStorageFixedDisk":
        return "固定ディスク";
      case "hrStorageVirtualMemory":
        return "仮想メモリ";
      case "hrStorageRam":
        return "実メモリ";
    }
    return "その他";
  };

  const renderDeviceType = (t) => {
    return t.replace("hrDevice", "");
  };

  const renderFSType = (t) => {
    return t.replace("hrFS", "");
  };

  const renderTrueFalse = (v) => {
    if (v === 1) {
      return "Yes";
    }
    return "No";
  };

  const renderAccess = (v) => {
    if (v === 1) {
      return "R/W";
    }
    return "Read Only";
  };

  const renderCPU = (v) => {
    return (v / 100).toFixed(2);
  };

  const renderMem = (v, t) => {
    return renderBytes(v * 1024, t);
  };

  let hostResource: backend.HostResourceEnt | undefined = undefined;
  let hrSystemTable = undefined;
  let hrStorageTable = undefined;
  let hrDeviceTable = undefined;
  let hrFileSystemTable = undefined;
  let hrProcessTable = undefined;
  let waitHr = false;

  const showHrSystem = async () => {
    if (!hostResource) {
      waitHr = true;
      hostResource = await GetHostResource(id);
      waitHr = false;
      await tick();
    }
    if ( !hostResource) {
      return;
    }
    hrSystemTable = new DataTable("#hrSystemTable", {
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
          data: "Name",
          title: "名前",
          width: "30%",
        },
        {
          data: "Value",
          title: "値",
          width: "60%",
        },
      ],
    });
    showHrSummaryChart();
    showHrCPUChart();
  };

  let selectedhrStorageCount = 0;

  const showHrStorage = () => {
    selectedhrStorageCount = 0;
    hrStorageTable = new DataTable("#hrStorageTable", {
      data: hostResource.Storage,
      language: getTableLang(),
      order: [[4, "desc"]],
      select: {
        style: "single",
      },
      columns: [
        {
          title: "種別",
          data: "Type",
          width: "20%",
          render: renderStorageType,
        },
        { title: "説明", data: "Descr", width: "40%" },
        {
          title: "サイズ",
          data: "Size",
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          title: "使用量",
          data: "Used",
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          title: "使用率",
          data: "Rate",
          width: "10%",
          render: renderRate,
          className: "dt-body-right",
        },
        {
          title: "単位",
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
    hrDeviceTable = new DataTable("#hrDeviceTable", {
      data: hostResource.Device,
      language: getTableLang(),
      order: [[0, "asc"]],
      columns: [
        { title: "状態", data: "Status", width: "10%", render: renderStatus },
        { title: "インデックス", data: "Index", width: "10%" },
        { title: "種別", data: "Type", width: "30%", render: renderDeviceType },
        { title: "説明", data: "Descr", width: "40%" },
        { title: "エラー", data: "Errors", width: "10%" },
      ],
    });
  };

  const showHrFileSystem = () => {
    hrFileSystemTable = new DataTable("#hrFileSystemTable", {
      data: hostResource.FileSystem,
      language: getTableLang(),
      order: [[0, "asc"]],
      columns: [
        { title: "マウント", data: "Mount", width: "30%" },
        { title: "リモート", data: "Remote", width: "30%" },
        { title: "種別", data: "Type", width: "20%", render: renderFSType },
        {
          title: "アクセス",
          data: "Access",
          width: "10%",
          render: renderAccess,
        },
        {
          title: "ブート",
          data: "Bootable",
          width: "10%",
          render: renderTrueFalse,
        },
      ],
    });
  };

  let selectedHrProcessCount = 0;

  const showHrProcess = () => {
    selectedHrProcessCount = 0;
    hrProcessTable = new DataTable("#hrProcessTable", {
      data: hostResource.Process,
      language: getTableLang(),
      order: [[1, "asc"]],
      select: {
        style: "single",
      },
      columns: [
        {
          title: "状態",
          data: "Status",
          width: "10%",
          render: renderStatus,
        },
        { title: "PID", data: "PID", width: "10%" },
        { title: "種別", data: "Type", width: "10%" },
        {
          title: "名前",
          data: "Name",
          width: "15%",
        },
        { title: "パス", data: "Path", width: "15%" },
        { title: "パラメータ", data: "Param", width: "20%" },
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
  };

  const showHrSummaryChart = async () => {
    await tick();
    const data = {
      CPU: 0,
      Mem: 0,
      VM: 0,
    };
    let cpu = 0;
    hostResource.System.forEach((e) => {
      if (e.Name.includes("CPU")) {
        cpu++;
        data.CPU = Number(e.Value);
      }
    });
    hostResource.Storage.forEach((e) => {
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
    data.CPU /= cpu > 0 ? cpu : 1;
    showHrSummary("hrSummaryChart", data);
  };

  const showHrCPUChart = async () => {
    await tick();
    const list = [];
    hostResource.System.forEach((e) => {
      if (e.Name.includes("CPU")) {
        list.unshift({
          Name: e.Name,
          Value: Number(e.Value),
        });
      }
    });
    showHrBarChart("hrSystemCPUChart", "CPU使用率", "%", list);
  };

  const showHrStorageChart = async () => {
    await tick();
    const list = [];
    hostResource.Storage.forEach((e) => {
      const t = renderStorageType(e.Type);
      if (!t.includes("その他")) {
        list.unshift({
          Name: e.Descr + "(" + t + ")",
          Value: e.Rate,
        });
      }
    });
    showHrBarChart("hrStorageChart", "ストレージ使用率", "%", list);
  };

  const showHrProcChart = async (bCPU) => {
    await tick();
    let max = 0;
    const list = [];
    hostResource.Process.forEach((e) => {
      const v = bCPU ? e.CPU / 100.0 : e.Mem * 1024;
      if (max < v) {
        max = v;
      }
      list.push({
        Name: e.Name + "(" + e.PID + ")",
        Value: v,
      });
    });
    list.sort((a, b) => {
      if (a.Value < b.Value) return -1;
      if (a.Value > b.Value) return 1;
      return 0;
    });
    while (list.length > 20) {
      list.shift();
    }
    showHrBarChart(
      bCPU ? "hrProcessCPUChart" : "hrProcessMemChart",
      bCPU ? "CPU使用量" : "Mem使用量",
      bCPU ? "秒" : "Bytes",
      list,
      max
    );
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  onMount(async () => {
    node = await GetNode(id);
    show = true;
  });

  onDestroy(() => {
    deleteVPanel();
  });
</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full min-h-[90vh]"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          基本情報
        </div>
        <Table striped={true}>
          <TableHead>
            <TableHeadCell>項目</TableHeadCell>
            <TableHeadCell>内容</TableHeadCell>
          </TableHead>
          <TableBody tableBodyClass="divide-y">
            <TableBodyRow>
              <TableBodyCell>名前</TableBodyCell>
              <TableBodyCell>{node.Name}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>状態</TableBodyCell>
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
              <TableBodyCell>IPアドレス</TableBodyCell>
              <TableBodyCell>{node.IP}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>MACアドレス</TableBodyCell>
              <TableBodyCell>{node.MAC}</TableBodyCell>
            </TableBodyRow>
            <TableBodyRow>
              <TableBodyCell>説明</TableBodyCell>
              <TableBodyCell>{node.Descr}</TableBodyCell>
            </TableBodyRow>
          </TableBody>
        </Table>
      </TabItem>
      <TabItem on:click={showLog}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiCalendarCheck} size={1} />
          ログ
        </div>
        <table id="logTable" class="display compact" style="width:99%" />
      </TabItem>
      <TabItem on:click={showVPanel}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiAppsBox} size={1} />
          パネル
        </div>
        <div id="vpanel" style="width: 98%; height: 500px" />
        <table id="portTable" class="display compact mt-2" style="width:99%" />
      </TabItem>
      <TabItem on:click={()=>{
        showHrSystem();
      }}>
        <div slot="title" class="flex items-center gap-2">
          {#if waitHr}
            <Spinner color="red" size="6" />
          {:else}
            <Icon path={icons.mdiInformation} size={1} />
          {/if}
          ホスト情報
        </div>
        {#if hostResource}
          <div class="flex w-full">
            <div id="hrSummaryChart" style="width: 35%; height: 300px" />
            <div id="hrSystemCPUChart" style="width: 63%; height: 300px" />
          </div>
          <table
            id="hrSystemTable"
            class="display compact mt-2"
            style="width:99%"
          />
        {:else if !waitHr}
          <div>ホストリソースMIBに対応していません。</div>
        {/if}
      </TabItem>
      {#if hostResource}
        <TabItem on:click={showHrStorage}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiDatabase} size={1} />
            ストレージ
          </div>
          <div id="hrStorageChart" style="width: 98%; height: 300px" />
          <table
            id="hrStorageTable"
            class="display compact mt-2"
            style="width:99%"
          />
        </TabItem>
        <TabItem on:click={showHrDevice}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiApplicationCog} size={1} />
            デバイス
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
            File System
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
            プロセス
          </div>
          <div class="flex w-full mx-auto">
            <div id="hrProcessCPUChart" style="width: 49%; height: 300px" />
            <div id="hrProcessMemChart" style="width: 49%; height: 300px" />
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
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        閉じる
      </Button>
    </div>
  </div>
</Modal>

<style global>
  #vpanel canvas {
    margin: 0 auto;
  }
</style>
