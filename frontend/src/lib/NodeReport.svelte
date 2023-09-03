<script lang="ts">
  import { Modal, Button,Tabs,TabItem,Table, TableBody, TableBodyCell, TableBodyRow, TableHead, TableHeadCell } from "flowbite-svelte";
  import { onMount, createEventDispatcher,tick } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { GetNode } from "../../wailsjs/go/main/App";
  import {getIcon,getStateColor,getStateName} from "./common";

  export let id = "";
  let node : datastore.NodeEnt | undefined = undefined;
  let show: boolean = false;
  const dispatch = createEventDispatcher();

  onMount(async () => {
    node = await GetNode(id);
    show = true;
  });

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full"
  on:on:close={close}
>
  <div class="flex flex-col space-y-4">
    <Tabs style="underline">
      <TabItem open on:click={()=>{}}>
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
                <span class="mdi {getIcon(node.Icon)} text-xl" style="color:{getStateColor(node.State)};"></span>
                <span class="ml-2 text-xs text-black dark:text-white">{getStateName(node.State)}</span>
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
      <TabItem on:click={()=>{}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiCheck} size={1} />
          ポーリング
        </div>
      </TabItem>
      <TabItem on:click={()=>{}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          ログ
        </div>
      </TabItem>
      <TabItem on:click={()=>{}}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartScatterPlot} size={1} />
          パネル
        </div>
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
