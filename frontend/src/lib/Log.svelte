<script lang="ts">
  import {
    Table,
    TableBody,
    TableBodyCell,
    TableBodyRow,
    TableHead,
    TableHeadCell,
  } from "flowbite-svelte";
  import { onMount, onDestroy } from "svelte";
  import { GetLastEventLogs } from "../../wailsjs/go/main/App";
  import Icon from "mdi-svelte";
  import {
    getStateColor,
    getStateIcon,
    getStateName,
    timeFormat,
  } from "./common";
  let logs = [];
  let timer: number | undefined = undefined;
  const updateLogs = async () => {
    logs = await GetLastEventLogs(100);
    timer = setTimeout(() => {
      updateLogs();
    }, 60 * 1000);
  };
  onMount(() => {
    updateLogs();
  });
  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  });
</script>

<Table
  divClass="relative overflow-x-auto overflow-y-auto h-full text-xs"
  hoverable
>
  <TableHead theadClass="p-1 text-xs">
    <TableHeadCell padding="p-2">状態</TableHeadCell>
    <TableHeadCell padding="p-2">発生日時</TableHeadCell>
    <TableHeadCell padding="p-2">種別</TableHeadCell>
    <TableHeadCell padding="p-2">関連ノード</TableHeadCell>
    <TableHeadCell padding="p-2">イベント</TableHeadCell>
  </TableHead>
  <TableBody>
    {#each logs as l}
      <TableBodyRow>
        <TableBodyCell tdClass="text-xs p-1">
          <div class="flex">
            <Icon
              path={getStateIcon(l.Level)}
              color={getStateColor(l.Level)}
              size={1}
            />
            <span class="text-xs mt-1 ml-1">
              {getStateName(l.Level)}
            </span>
          </div>
        </TableBodyCell>
        <TableBodyCell tdClass="text-xs p-1"
          >{timeFormat(new Date(l.Time / (1000 * 1000)), "")}</TableBodyCell
        >
        <TableBodyCell tdClass="text-xs p-1">{l.Type}</TableBodyCell>
        <TableBodyCell tdClass="text-xs p-1">{l.NodeName}</TableBodyCell>
        <TableBodyCell tdClass="text-xs p-1">{l.Event}</TableBodyCell>
      </TableBodyRow>
    {/each}
  </TableBody>
</Table>
