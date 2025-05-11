<script lang="ts">
  import {
    Modal,
    GradientButton,
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
    Label,
    Spinner,
    Input,
  } from "flowbite-svelte";
  import { GetAddressInfo } from "../../wailsjs/go/main/App";
  import { createEventDispatcher } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { BrowserOpenURL } from "../../wailsjs/runtime";

  export let show: boolean = false;
  export let address: string = "";

  const dispatch = createEventDispatcher();

  const onOpen = () => {
    showAddressInfo();
  };

  let addressInfoList: any = [];
  let wait = false;
  let isGlobalIP = false;
  let latLong = "";

  const showAddressInfo = async () => {
    if (address == "") {
      return;
    }
    wait = true;
    addressInfoList = await GetAddressInfo(address);
    wait = false;
    latLong = "";
    isGlobalIP = false;
    if (address.match(/^\d{1,3}(\.\d{1,3}){3}$/) || (address.includes(":") && !address.match(/^[0-9a-fA-F]{2}(:[0-9a-fA-F]{2}){5}$/i))) {
      for (const i of addressInfoList) {
        if (i.Title == $_('Address.Location') || i.Title == "Location") {
          const a = i.Value.split(",");
          if (a.length > 2 && a[0] !== "LOCAL") {
            latLong = a[1] + "," + a[2];
          }
          isGlobalIP = !i.Value.includes("LOCAL");
          break;
        }
      }
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  {#if wait}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <form class="flex flex-col space-y-4" action="#">
      <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
        {$_('Address.AddressInfo')}
      </h3>
      <Label class="space-y-2 text-xs">
        <span> {$_('Address.Address')} </span>
        <Input class="h-8" bind:value={address} size="sm" />
      </Label>
    </form>
    <div class="flex flex-col space-y-4">
      <Table striped={true}>
        <TableHead>
          <TableHeadCell>{$_('Address.Name')}</TableHeadCell>
          <TableHeadCell>{$_('Address.Value')}</TableHeadCell>
        </TableHead>
        <TableBody tableBodyClass="divide-y">
          {#each addressInfoList as i}
            <TableBodyRow>
              <TableBodyCell tdClass="py-2 px-4 font-small"
                >{i.Title}</TableBodyCell
              >
              <TableBodyCell tdClass="py-2 px-4 font-small"
                >{i.Value}</TableBodyCell
              >
            </TableBodyRow>
          {/each}
        </TableBody>
      </Table>
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if !wait && address}
        <GradientButton
          type="button"
          color="blue"
          on:click={showAddressInfo}
          size="xs"
        >
          <Icon path={icons.mdiRecycle} size={1} />
          {$_('Address.Search')}
        </GradientButton>
        {#if isGlobalIP}
          {#if latLong}
            <GradientButton
              shadow
              type="button"
              color="lime"
              class="mr-2"
              size="xs"
              on:click={() => {
                BrowserOpenURL(
                  `https://www.google.com/maps/search/?api=1&query=` + latLong
                );
              }}
            >
              <Icon path={icons.mdiGoogleMaps} size={1} />
              Google MAP
            </GradientButton>
          {/if}
          <GradientButton
            shadow
            type="button"
            color="lime"
            class="mr-2"
            size="xs"
            on:click={() => {
              BrowserOpenURL(
                `https://www.virustotal.com/gui/ip-address/` + address
              );
            }}
          >
            <Icon path={icons.mdiVirus} size={1} />
            VirusTotal
          </GradientButton>
        {/if}
      {/if}
      <GradientButton type="button" color="teal" on:click={close} size="xs">
        <Icon path={icons.mdiCancel} size={1} />
        {$_("ArpReport.Close")}
      </GradientButton>
    </div>
  {/if}
</Modal>
