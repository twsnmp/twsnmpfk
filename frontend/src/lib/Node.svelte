<script lang="ts">
  import {
    Select,
    Modal,
    Label,
    Input,
    Checkbox,
    GradientButton,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import {
    GetNode,
    UpdateNode,
    GetImageIconList,
    GetImageIcon,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { addrModeList, getIcon, iconList, snmpModeList } from "./common";
  import { _ } from "svelte-i18n";
  import Help from "./Help.svelte";

  export let show: boolean = false;
  export let nodeID: string = "";
  export let posX = 0;
  export let posY = 0;
  export let ip = "";

  let node: any = undefined;
  let showHelp = false;
  let imageIcon : any  = undefined;
  const imageIconList: any = [];

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    if (nodeID) {
      node = await GetNode(nodeID);
    } else {
      node = {
        ID: "",
        Name: $_("Node.NewNode"),
        Descr: "",
        Icon: "",
        Image: "",
        State: "",
        X: posX,
        Y: posY,
        IP: ip,
        MAC: "",
        Vendor: "",
        SnmpMode: "v2c",
        Community: "public",
        User: "",
        SSHUser: "",
        Password: "",
        PublicKey: "",
        URL: "",
        AddrMode: "ip",
        AutoAck: false,
        Loc: "",
      };
    }
    if (!node.AddrMode) {
      node.AddrMode = "ip";
    }
    if (!node.SnmpMode) {
      node.SnmpMode = "v2c";
    }
    const icons = await GetImageIconList();
    imageIconList.length = 0;
    imageIconList.push({
      name: $_('Node.None'),
      value: "",
    });
    if (icons) {
      for (const icon of icons) {
        imageIconList.push({
          name: icon,
          value: icon,
        });
      }
    }
    if (node.Image) {
      imageIcon = await GetImageIcon(node.Image);
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const save = async () => {
    const r = await UpdateNode(node);
    if (r) {
      close();
    }
  };

  const selectImageIcon = async () => {
    if (node.Image) {
      imageIcon = await GetImageIcon(node.Image);
    } else {
      imageIcon = undefined;
    }
  }

</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full"
  on:open={onOpen}
>
  {#if !node}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_("Node.EditNode")}
      </h3>
      <div class="grid gap-4 mb-4 md:grid-cols-4">
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.Name")}</span>
          <Input class="h-8" bind:value={node.Name} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.IPAddress")}</span>
          <Input class="h-8" bind:value={node.IP} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_('Node.MAC')}</span>
          <Input class="h-8" bind:value={node.MAC} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span> {$_("Node.AddressMode")} </span>
          <Select
            items={addrModeList}
            bind:value={node.AddrMode}
            placeholder={$_("Node.SelectAddressMode")}
            size="sm"
          />
        </Label>
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-3">
        <Label class="space-y-2 text-xs">
          <span> {$_("Node.Icon")} </span>
          <Select
            items={iconList}
            bind:value={node.Icon}
            placeholder={$_("Node.SelectIcon")}
            size="sm"
          />
        </Label>
        <div class="mt-5 ml-5">
          <span class="mdi {getIcon(node.Icon)} text-4xl" />
        </div>
        <Checkbox bind:checked={node.AutoAck}>{$_("Node.AutoCheck")}</Checkbox>
      </div>
      <div class="grid gap-4 mb-4 md:grid-cols-3">
        <Label class="space-y-2 text-xs col-span-2">
          <span> {$_('Node.Image')} </span>
          <Select
            items={imageIconList}
            bind:value={node.Image}
            on:change={selectImageIcon}
            placeholder={$_('Node.SelectImageIcon')}
            size="sm"
          />
        </Label>
        <div class="mt-5 ml-5">
          {#if imageIcon}
            <img src={imageIcon} alt="" class="h-[48px]" />
          {/if}
        </div>
      </div>
      <div class="grid gap-4 md:grid-cols-4">
        <Label class="space-y-2 text-xs">
          <span> {$_("Node.SNMPMode")} </span>
          <Select
            items={snmpModeList}
            bind:value={node.SnmpMode}
            placeholder={$_("Node.SelectSnmpMode")}
            size="sm"
          />
        </Label>
        {#if node.SnmpMode == "v1" || node.SnmpMode == "v2c"}
          <Label class="space-y-2 text-xs">
            <span>SNMP Community</span>
            <Input class="h-8" bind:value={node.Community} placeholder="public" size="sm" />
          </Label>
        {:else}
          <div></div>
        {/if}
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.User")}</span>
          <Input class="h-8" bind:value={node.User} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.Password")}</span>
          <Input class="h-8" type="password" bind:value={node.Password} size="sm" />
        </Label>
    </div>
      <div class="grid gap-4 md:grid-cols-4">
        <Label class="space-y-2 text-xs">
          <span>gNMI Port</span>
          <Input class="h-8" bind:value={node.GNMIPort} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>gNMI Encoding</span>
          <Input class="h-8" bind:value={node.GNMIEncoding} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_('Node.GNMIUser')}</span>
          <Input class="h-8" bind:value={node.GNMIUser} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_('Node.GNMIPassword')}</span>
          <Input class="h-8" type="password" bind:value={node.GNMIPassword} size="sm" />
        </Label>
      </div>
      <Label class="space-y-2 text-xs">
        <span>URL</span>
        <Input class="h-8" bind:value={node.URL} placeholder="URL" size="sm" />
      </Label>
      <div class="grid gap-4 md:grid-cols-2">
        <Label class="space-y-2 text-xs">
          <span>SSH{$_("Node.User")}</span>
          <Input class="h-8" bind:value={node.SSHUser} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>{$_("Node.PublicKey")}</span>
          <Input class="h-8" bind:value={node.PublicKey} size="sm" />
        </Label>
      </div>
      <Label class="space-y-2">
        <span>{$_("Node.Descr")}</span>
        <Input class="h-8" bind:value={node.Descr} size="sm" />
      </Label>
      <div class="flex justify-end space-x-2 mr-2">
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

<Help bind:show={showHelp} page="editnode" />
