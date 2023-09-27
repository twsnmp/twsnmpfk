<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Checkbox,
    Button,
  } from "flowbite-svelte";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import { GetNode, UpdateNode } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { addrModeList, getIcon, iconList, snmpModeList } from "./common";
  import { _ } from "svelte-i18n";

  export let nodeID: string = "";
  export let posX = 0;
  export let posY = 0;
  export let ip = "";
  let node: datastore.NodeEnt | undefined = undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    if (nodeID) {
      node = await GetNode(nodeID);
    } else {
      node = {
        ID: "",
        Name: $_('Node.NewNode'),
        Descr: "",
        Icon: "",
        State: "",
        X: posX,
        Y: posY,
        IP: ip,
        IPv6: "",
        MAC: "",
        SnmpMode: "v2c",
        Community: "public",
        User: "",
        Password: "",
        PublicKey: "",
        URL: "",
        AddrMode: "ip",
        AutoAck: false,
      };
    }
    show = true;
  });

  onDestroy(() => {});

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    const r = await UpdateNode(node);
    if (r) {
      close();
    } else {
    }
  };
</script>

<Modal bind:open={show} size="lg" permanent class="w-full" on:on:close={close}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">{ $_('Node.EditNode') }</h3>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span>{ $_('Node.Name') }</span>
        <Input
          bind:value={node.Name}
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>{ $_('Node.IPAddress') }</span>
        <Input
          bind:value={node.IP}
          required
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span> { $_('Node.AddressMode') } </span>
        <Select
          items={addrModeList}
          bind:value={node.AddrMode}
          placeholder={ $_('Node.SelectAddressMode') }
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
      <Label class="space-y-2">
        <span> { $_('Node.Icon') } </span>
        <Select
          items={iconList}
          bind:value={node.Icon}
          placeholder={ $_('Node.SelectIcon') }
          size="sm"
        />
      </Label>
      <div class="mt-5 ml-5">
        <span class="mdi {getIcon(node.Icon)} text-4xl" />
      </div>
      <Checkbox bind:checked={node.AutoAck}>{ $_('Node.AutoCheck') }</Checkbox>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span> { $_('Node.SNMPMode') } </span>
        <Select
          items={snmpModeList}
          bind:value={node.SnmpMode}
          placeholder="{ $_('Node.SelectSnmpMode') }"
          size="sm"
        />
      </Label>
      <Label class="space-y-2">
        <span>SNMP Community</span>
        <Input bind:value={node.Community} placeholder="public" size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>{ $_('Node.SnmpUser') }</span>
        <Input bind:value={node.User} size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>{ $_('Node.SnmpPassword') }</span>
        <Input
          type="password"
          bind:value={node.Password}
          size="sm"
        />
      </Label>
    </div>
    <Label class="space-y-2">
      <span>{ $_('Node.PublicKey') }</span>
      <Input bind:value={node.PublicKey} size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>URL</span>
      <Input bind:value={node.URL} placeholder="URL" size="sm" />
    </Label>
    <Label class="space-y-2">
      <span>{ $_('Node.Descr') }</span>
      <Input bind:value={node.Descr} size="sm" />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <Button color="blue" type="button" on:click={save} size="sm">
        <Icon path={icons.mdiContentSave} size={1} />
        { $_('Node.Save') }
      </Button>
      <Button type="button" color="alternative" on:click={close} size="sm">
        <Icon path={icons.mdiCancel} size={1} />
        { $_('Node.Cancel') }
      </Button>
    </div>
  </form>
</Modal>
