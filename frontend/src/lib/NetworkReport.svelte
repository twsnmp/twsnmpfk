<script lang="ts">
  import {
    Modal,
    GradientButton,
    Tabs,
    TabItem,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Spinner,
    Toggle,
    Button,
  } from "flowbite-svelte";
  import { tick } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import {
    GetNetwork,
    GetVPanelPorts,
    GetVPanelPowerInfo,
    GetFDBTable,
  } from "../../wailsjs/go/main/App";
  import {
    getTableLang,
    renderBytes,
    renderCount,
    renderSpeed,
    renderTime,
    renderState,
  } from "./common";
  import { showFDBTableGraph } from "./chart/network";
  import { deleteVPanel, initVPanel, setVPanel } from "./vpanel";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import { copyText } from "svelte-copy";
  import Help from "./Help.svelte";

  export let show: boolean = false;
  export let id = "";
  let network: any;

  let physicalPort = true;
  let showVPanelBtn = false;
  let chart: any = undefined;
  let showHelp = false;

  let portTable: any = undefined;

  const clear = () => {
    showVPanelBtn = false;
    chart = undefined;
  };

  const showPortTable = (p: any) => {
    portTable = new DataTable("#portTable", {
      destroy: true,
      stateSave: true,
      paging: false,
      searching: false,
      info: false,
      scrollY: "20vh",
      data: p,
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
          title: $_("NodePolling.State"),
          width: "5%",
          render: renderState,
        },
        {
          data: "Name",
          title: $_("NodeReport.Name"),
          width: "15%",
        },
        {
          data: "Type",
          title: $_("NodeReport.Type"),
          width: "5%",
        },
        {
          data: "MAC",
          title: $_("NodeReport.MACAddress"),
          width: "10%",
        },
        {
          data: "Speed",
          title: $_("NodeReport.Speed"),
          width: "8%",
          render: renderSpeed,
          className: "dt-body-right",
        },
        {
          data: "OutPacktes",
          title: $_("NodeReport.OutPacktes"),
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "OutBytes",
          title: $_("NodeReport.OutBytes"),
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          data: "InPacktes",
          title: $_("NodeReport.InPacktes"),
          width: "10%",
          render: renderCount,
          className: "dt-body-right",
        },
        {
          data: "InBytes",
          title: $_("NodeReport.InBytes"),
          width: "10%",
          render: renderBytes,
          className: "dt-body-right",
        },
        {
          data: "LastChanged",
          title: $_('NodeReport.LastChanged'),
          width: "12%",
          render: renderTime,
        },
      ],
    });
  };

  let ports: any;
  let power: any;
  let waitVPanel = false;
  let rotateVPanel = false;

  const showVPanel = async () => {
    clear();
    showVPanelBtn = true;
    initVPanel("vpanel");
    if (!ports) {
      waitVPanel = true;
      ports = await GetVPanelPorts("NET:" + id);
      power = await GetVPanelPowerInfo("NET:" + id);
      waitVPanel = false;
    }
    const p = physicalPort ? ports.filter((e: any) => e.Type == 6) : ports;
    setVPanel(p, power, rotateVPanel);
    showPortTable(p);
  };

  let fdbTable: any = undefined;
  let fdbTableTable: any = undefined;
  let waitFDBTable = false;

  const showFDBTable = async () => {
    clear();
    if (!fdbTable) {
      waitFDBTable = true;
      fdbTable = await GetFDBTable(id);
      waitFDBTable = false;
      await tick();
    }
    if (!fdbTable) {
      return;
    }
    fdbTableTable = new DataTable("#fdbTable", {
      data: fdbTable,
      language: getTableLang(),
      order: [[0, "asc"]],
      columns: [
        {
          data: "IfIndex",
          title: "Index",
          width: "5%",
        },
        {
          data: "Port",
          title: $_('NetworkReport.Port'),
          width: "5%",
        },
        {
          data: "VLanID",
          title: "VLAN ID",
          width: "5%",
        },
        {
          data: "Node",
          title: $_('NetworkReport.Node'),
          width: "30%",
        },
        {
          data: "MAC",
          title: "MAC",
          width: "20%",
        },
        {
          data: "Vendor",
          title: $_('NetworkReport.Vendor'),
          width: "35%",
        },
      ],
    });
    showFDBTableChart();
  };

  const showFDBTableChart = async () => {
    await tick();
    chart = showFDBTableGraph("fdbTableChart", fdbTable);
  };
    
  const close = () => {
    deleteVPanel();
    show = false;
  };

  const onOpen = async () => {
    network = undefined;
    ports = undefined;
    power = undefined;
    fdbTable = undefined;
    network = await GetNetwork(id);
  };

  const resizeChart = () => {
    if (chart) {
      chart.resize();
    }
  };

  let copied = false;

</script>

<svelte:window on:resize={resizeChart} />

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  {#if !network}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <div class="flex flex-col space-y-4">
      <Tabs style="underline">
        <TabItem open on:click={clear}>
          <div slot="title" class="flex items-center gap-2">
            <Icon path={icons.mdiChartPie} size={1} />
            {$_("NodeReport.BasicInfo")}
          </div>
          <Table striped={true}>
            <TableHead>
              <TableHeadCell>{$_("NodeReport.Item")}</TableHeadCell>
              <TableHeadCell>{$_("NodeReport.Content")}</TableHeadCell>
            </TableHead>
            <TableBody tableBodyClass="divide-y">
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Name")}</TableBodyCell>
                <TableBodyCell>{network.Name}</TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.IPAddress")}</TableBodyCell>
                <TableBodyCell>
                  {network.IP}
                  <Button
                    color="alternative"
                    type="button"
                    class="ml-2 !p-2"
                    on:click={async () => {
                      copied = true
                      copyText(network.IP)
                      setTimeout(()=> copied = false,2000);
                    }}
                    size="xs"
                  >
                    {#if copied}
                      <Icon path={icons.mdiCheck} size={1} />
                    {:else}
                      <Icon path={icons.mdiContentCopy} size={1} />
                    {/if}
                  </Button>
                </TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_('NetworkReport.PortNum')}</TableBodyCell>
                <TableBodyCell>{network.Ports.length}</TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_('NetworkReport.Error')}</TableBodyCell>
                <TableBodyCell><span class="text-red-700 dark:text-red-400">{network.Error}</span></TableBodyCell>
              </TableBodyRow>
              <TableBodyRow>
                <TableBodyCell>{$_("NodeReport.Descr")}</TableBodyCell>
                <TableBodyCell>{network.Descr}</TableBodyCell>
              </TableBodyRow>
            </TableBody>
          </Table>
        </TabItem>
        <TabItem on:click={showVPanel}>
          <div slot="title" class="flex items-center gap-2">
            {#if waitVPanel}
              <Spinner color="red" size="6" />
            {:else}
              <Icon path={icons.mdiAppsBox} size={1} />
            {/if}
            {$_("NodeReport.Panel")}
          </div>
          <div id="vpanel" />
          <table
            id="portTable"
            class="display compact mt-5"
            style="width:99%"
          />
        </TabItem>
        <TabItem on:click={showFDBTable} >
          <div slot="title" class="flex items-center gap-2">
            {#if waitFDBTable}
              <Spinner color="red" size="6" />
            {:else}
              <Icon path={icons.mdiInformation} size={1} />
            {/if}
            <span>{$_('NetworkReport.FDBTable')}</span>
          </div>
          {#if fdbTable}
            <div class="grid grid-cols-2 gap-1">
              <div id="fdbTableChart" />
              <div>
                <table
                  id="fdbTable"
                  class="display compact"
                  style="width:100%"
                />
              </div>
            </div>
          {:else if !waitFDBTable}
            <div>{$_('NetworkReport.FDBTableNotSupported')}</div>
          {/if}
        </TabItem>
      </Tabs>
      <div class="flex justify-end space-x-2 mr-2">
        {#if showVPanelBtn}
          <Toggle bind:checked={physicalPort} on:change={showVPanel}>
            {$_("NodeReport.PhysicalPort")}
          </Toggle>
          <Toggle bind:checked={rotateVPanel} on:change={showVPanel}>
            {$_("NodeReport.RotateVPanel")}
          </Toggle>
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
            {$_("Line.Help")}
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
          {$_("NodeReport.Close")}
        </GradientButton>
      </div>
    </div>
  {/if}
</Modal>

<Help bind:show={showHelp} page="network_report" />

<style>
  #vpanel {
    width: 98%;
    min-height: 400px;
    height: 40vh;
    overflow: scroll;
    margin: 0 auto;
  }

  #fdbTableChart {
    min-width: 350px;
    min-height: 350px;
    margin: 0 auto;
    width: 35vw;
    height: 35vw;
  }
</style>
