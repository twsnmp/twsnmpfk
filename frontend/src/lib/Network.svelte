<script lang="ts">
  import {
    Alert,
    Select,
    Modal,
    Label,
    Input,
    Checkbox,
    GradientButton,
    Spinner,
    Badge,
  } from "flowbite-svelte";
  import { createEventDispatcher, tick } from "svelte";
  import { GetNetwork, UpdateNetwork, ImportPortDef, ExportPortDef } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { snmpModeList, getTableLang, renderNodeState } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";

  export let show: boolean = false;
  export let id: string = "";
  export let posX = 0;
  export let posY = 0;
  export let ip = "";
  export let template: any = undefined;

  let network: any = undefined;
  let showHelp = false;
  let data: any = [];
  let table: any = undefined;
  let selectedCount = 0;
  let toltalPorts = 8;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    if (template) {
      network = template;
    } else if (id) {
      network = await GetNetwork(id);
      toltalPorts = network.Ports.length;
    } else {
      network = {
        ID: "",
        Name: "",
        Descr: "",
        X: posX,
        Y: posY,
        IP: ip,
        SnmpMode: "v2c",
        Community: "public",
        User: "",
        Password: "",
        URL: "",
        HPorts: 24,
        LLDP: false,
        ArpWatch: false,
        Unmanaged: false,
        Ports: [],
      };
    }
    if (!network.SnmpMode) {
      network.SnmpMode = "v2c";
    }
    showTable();
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const updateUnmanagedNetworkPort = () => {
    if (toltalPorts != network.Ports.length) {
      let x = 0;
      let y = 0;
      network.Ports.length = 0;
      for(let i = 0 ; i < toltalPorts;i++ ) {
        network.Ports.push({
          ID: "#" + i,
          Name: "#" + i,
          X: x,
          Y: y,
        })
        x++
        if (x >= network.HPorts) {
          y++
          x++
        }
      }
    }
  }

  const save = async () => {
    if (network.Unmanaged) {
      updateUnmanagedNetworkPort()
    }
    const r = await UpdateNetwork(network);
    if (r) {
      close();
    }
  };

  const columns = [
    {
      data: "State",
      title: $_("NodeList.State"),
      width: "10%",
      render: renderNodeState,
    },
    {
      data: "Name",
      title: $_("NodeList.Name"),
      width: "15%",
    },
    {
      data: "Polling",
      title: $_("NodeList.Polling"),
      width: "10%",
    },
    {
      data: "X",
      title: "X",
      width: "10%",
    },
    {
      data: "Y",
      title: "Y",
      width: "10%",
    },
  ];

  const showTable = async () => {
    await tick();
    selectedCount = 0;
    data = [];
    for (const p of network.Ports) {
      data.push(p);
    }
    table = new DataTable("#portTable", {
      destroy: true,
      columns: columns,
      stateSave: true,
      data: data,
      paging: false,
      searching: false,
      ordering: false,
      info: false,
      scrollY: "100px",
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
  const portDelete = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    network.Ports.splice(i, 1);
    showTable();
  };
  const portTop = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    const r = network.Ports.splice(i, 1);
    network.Ports.unshift(r[0]);
    showTable();
  };
  const portUp = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i <= 0) {
      return;
    }
    const r = network.Ports.splice(i, 1);
    network.Ports.splice(i - 1, 0, r[0]);
    showTable();
  };
  const portDown = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i < 0 || i > network.Ports.length - 1) {
      return;
    }
    const r = network.Ports.splice(i, 1);
    network.Ports.splice(i + 1, 0, r[0]);
    showTable();
  };
  const portBottom = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    const r = network.Ports.splice(i, 1);
    network.Ports.push(r[0]);
    showTable();
  };
  const reNumber = () => {
    const ports: any = [];
    let x = 0;
    let y = 0;
    const HPorts = network.HPorts || 24;
    network.Ports.forEach((p: any) => {
      p.X = x;
      p.Y = y;
      x++;
      if (x >= HPorts) {
        x = 0;
        y++;
      }
      ports.push(p);
    });
    network.Ports = ports;
    showTable();
  };
  const reSearch = () => {
    network.Error = "";
    network.Ports = [];
    network.LLDP = false;
    network.X *= 1;
    network.Y *= 1;
    network.H *= 1;
    network.W *= 1;
    network.HPorts *= 1;
    UpdateNetwork(network);
    close();
  };
  let showEditPort = false;
  let editPort: any = {};
  let selectedPortIndex = -1;
  const portEdit = () => {
    if (selectedCount != 1) {
      return;
    }
    const sels = table.rows({ selected: true }).data();
    const i = network.Ports.indexOf(sels[0]);
    if (i < 0) {
      return;
    }
    editPort = network.Ports[i];
    selectedPortIndex = i;
    showEditPort = true;
  };
  const savePort = () => {
    showEditPort = false;
    if (selectedPortIndex < 0 || selectedPortIndex >= network.Ports.length) {
      return;
    }
    network.Ports[selectedPortIndex] = editPort;
    showTable();
  };
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if !network}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Network.EditNetwork")}
      </h3>
      {#if network.Error}
        <Alert color="red" dismissable>
          <div class="flex">
            <Icon path={icons.mdiExclamation} size={1} />
            {network.Error}
          </div>
        </Alert>
      {/if}
      <div class="grid gap-4 mb-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.Name")}</span>
          <Input bind:value={network.Name} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.IPAddress")}</span>
          <Input bind:value={network.IP} size="sm" />
        </Label>
      </div>
      {#if !id}
        <Checkbox bind:checked={network.Unmanaged}>
          {$_('Network.Unmanaged')}
        </Checkbox>
      {/if}
      {#if network.Unmanaged}
        <div class="grid gap-4 mb-4 md:grid-cols-2">
          <Label class="space-y-2 text-xs">
            <span>{$_('Network.AllPort')}</span>
            <Input
              type="number"
              min="5"
              max="100"
              bind:value={toltalPorts}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span>{$_("Network.HPorts")}</span>
            <Input
              type="number"
              min="5"
              max="100"
              bind:value={network.HPorts}
              size="sm"
            />
          </Label>
        </div>
      {:else}
        <div class="grid gap-4 mb-4 md:grid-cols-3">
          <Label class="space-y-2 text-xs">
            <span>{$_("Network.HPorts")}</span>
            <Input
              type="number"
              min="5"
              max="100"
              bind:value={network.HPorts}
              size="sm"
            />
          </Label>
          <div>
            {#if network.LLDP}
              <Badge rounded color="green" class="m-8">LLDP</Badge>
            {:else}
              <Badge rounded color="red" class="m-8">Not LLDP</Badge>
            {/if}
          </div>
          <Checkbox bind:checked={network.ArpWatch}
            >{$_("Network.ArpWatch")}</Checkbox
          >
        </div>
        <div class="grid gap-4 md:grid-cols-3">
          <Label class="space-y-2 text-xs">
            <span> {$_("Node.SNMPMode")} </span>
            <Select
              items={snmpModeList}
              bind:value={network.SnmpMode}
              placeholder={$_("Node.SelectSnmpMode")}
              size="sm"
            />
          </Label>
          {#if network.SnmpMode == "v1" || network.SnmpMode == "v2c"}
            <Label class="space-y-2 text-xs">
              <span>SNMP Community</span>
              <Input
                bind:value={network.Community}
                placeholder="public"
                size="sm"
              />
            </Label>
            <div></div>
          {:else}
            <Label class="space-y-2 text-xs">
              <span>SNMP{$_("Node.SnmpUser")}</span>
              <Input bind:value={network.User} size="sm" />
            </Label>
            <Label class="space-y-2 text-xs">
              <span>{$_("Node.SnmpPassword")}</span>
              <Input type="password" bind:value={network.Password} size="sm" />
            </Label>
          {/if}
        </div>
      {/if}
      <Label class="space-y-2 text-xs">
        <span>URL</span>
        <Input bind:value={network.URL} placeholder="URL" size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>{$_("Node.Descr")}</span>
        <Input bind:value={network.Descr} size="sm" />
      </Label>
      <div class="m-5 grow">
        <table id="portTable" class="display compact" style="width:99%" />
      </div>
      <div class="flex justify-end space-x-2 mr-2">
        {#if selectedCount > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={portTop}
            size="xs"
          >
            <Icon path={icons.mdiArrowCollapseUp} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={portUp}
            size="xs"
          >
            <Icon path={icons.mdiArrowUp} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={portDown}
            size="xs"
          >
            <Icon path={icons.mdiArrowDown} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={portBottom}
            size="xs"
          >
            <Icon path={icons.mdiArrowCollapseDown} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={portEdit}
            size="xs"
          >
            <Icon path={icons.mdiPencil} size={1} />
          </GradientButton>
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={portDelete}
            size="xs"
          >
            <Icon path={icons.mdiTrashCan} size={1} />
          </GradientButton>
        {/if}
        {#if id && !template}
          <GradientButton
            shadow
            color="lime"
            type="button"
            on:click={reNumber}
            size="xs"
          >
            <Icon path={icons.mdiOrderNumericAscending} size={1} />
            {$_("Network.ReNumberPort")}
          </GradientButton>
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={reSearch}
            size="xs"
          >
            <Icon path={icons.mdiRefresh} size={1} />
            {$_("Network.ReSearch")}
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          color="cyan"
          type="button"
          on:click={() => {
            const ports = [];
            for(const p of network.Ports) {
              ports.push({
                Name: p.Name,
                X: p.X,
                Y: p.Y,
                Polling: p.Polling,
              })
            }
            ExportPortDef(JSON.stringify(ports,null,"  "));
          }}
          size="xs"
        >
          <Icon path={icons.mdiDownload} size={1} />
        </GradientButton>
        <GradientButton
          shadow
          color="cyan"
          type="button"
          on:click={async () => {
            const d = await ImportPortDef();
            if (!d) {
              return;
            }
            const ports = JSON.parse(d);
            if(!ports) {
              return;
            }
            for(let i = 0; i < ports.length;i ++) {
              if (i < network.Ports.length) {
                network.Ports[i].Name = ports[i].Name || "";
                network.Ports[i].X = ports[i].X || 0;
                network.Ports[i].Y = ports[i].Y || 0;
                if (ports[i].Polling) {
                  network.Ports[i].Polling = ports[i].Polling;
                }
              }
            }
            showTable();
          }}
          size="xs"
        >
          <Icon path={icons.mdiUpload} size={1} />
        </GradientButton>
        <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={save}
          size="xs"
        >
          <Icon path={icons.mdiContentSave} size={1} />
          {$_("Node.Save")}
        </GradientButton>
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
          {$_("Node.Cancel")}
        </GradientButton>
      </div>
    </form>
  {/if}
</Modal>

<Modal bind:open={showEditPort} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Network.EditPort")}
    </h3>
    <Label class="space-y-2 text-xs">
      <span>{$_("Node.Name")}</span>
      <Input bind:value={editPort.Name} size="sm" />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>{$_("NodeList.Polling")}</span>
      <Input bind:value={editPort.Polling} size="sm" />
    </Label>

    <div class="grid gap-2 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>X</span>
        <Input
          type="number"
          min="0"
          max="100"
          bind:value={editPort.X}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>Y</span>
        <Input
          type="number"
          min="0"
          max="100"
          bind:value={editPort.Y}
          size="sm"
        />
      </Label>
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={savePort}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_("Node.Save")}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showEditPort = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Node.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Help bind:show={showHelp} page="editnetwork" />
