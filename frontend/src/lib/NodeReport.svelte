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
  } from "flowbite-svelte";
  import { onMount, createEventDispatcher, tick, onDestroy } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {
    GetNode,
    GetVPanelPorts,
    GetVPanelPowerInfo,
    GetEventLogs,
    GetPollings,
  } from "../../wailsjs/go/main/App";
  import {
    getIcon,
    getStateColor,
    getStateName,
    getTableLang,
    renderTime,
    renderState,
    getLogModeName,
    renderBytes,
    renderCount,
    renderSpeed,
  } from "./common";
  import { deleteVPanel, initVPanel, setVPanel } from "./vpanel";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { globals } from "svelte/internal";

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

  let pollingTable = undefined;

  const showPolling = async () => {
    if (pollingTable) {
      pollingTable.destroy();
      pollingTable = undefined;
    }
    pollingTable = new DataTable("#pollingTable", {
      data: await GetPollings(id),
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
        {
          data: "State",
          title: "状態",
          width: "15%",
          render: renderState,
        },
        {
          data: "Name",
          title: "名前",
          width: "35%",
        },
        {
          data: "Level",
          title: "レベル",
          width: "15%",
          render: renderState,
        },
        {
          data: "Type",
          title: "種別",
          width: "10%",
        },
        {
          data: "LogMode",
          title: "ログ",
          width: "10%",
          render: getLogModeName,
        },
        {
          data: "LastTime",
          title: "最終確認",
          width: "15%",
          render: renderTime,
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
      searching:false,
      info:false,
      scrollY: "180px",
      data: ports,
      language: getTableLang(),
      order: [[1, "desc"]],
      columns: [
        {
          data: "Index",
          title: "No.",
          width: "5%",
          className: 'dt-body-right',
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
          className: 'dt-body-right',
        },
        {
          data: "OutPacktes",
          title: "送信パケット",
          width: "10%",
          render: renderCount,
          className: 'dt-body-right',
        },
        {
          data: "OutBytes",
          title: "送信バイト",
          width: "10%",
          render: renderBytes,
          className: 'dt-body-right',
        },
        {
          data: "InPacktes",
          title: "受信パケット",
          width: "10%",
          render: renderCount,
          className: 'dt-body-right',
        },
        {
          data: "InBytes",
          title: "受信バイト",
          width: "10%",
          render: renderBytes,
          className: 'dt-body-right',
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

<Modal bind:open={show} size="xl" permanent class="w-full" on:on:close={close}>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={() => {}}>
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
      <TabItem on:click={showPolling}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiLanCheck} size={1} />
          ポーリング
        </div>
        <table id="pollingTable" class="display compact" style="width:99%" />
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
    margin:  0 auto;
  } 
</style>