<script lang="ts">
  import { _ } from "svelte-i18n";
  import { Input, Label, Button } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";

  interface Props {
    start?: string;
    end?: string;
  }

  let {
    start = $bindable(""),
    end = $bindable("")
  }: Props = $props();

  function formatLocalIso(date: Date): string {
    const pad = (n: number) => String(n).padStart(2, "0");
    const year = date.getFullYear();
    const month = pad(date.getMonth() + 1);
    const day = pad(date.getDate());
    const hours = pad(date.getHours());
    const minutes = pad(date.getMinutes());
    return `${year}-${month}-${day}T${hours}:${minutes}`;
  }

  type ShortcutType = "1h" | "6h" | "24h" | "today" | "yesterday" | "7d" | "30d" | "clear";

  const applyShortcut = (type: ShortcutType) => {
    const now = new Date();
    if (type === "clear") {
      start = "";
      end = "";
      return;
    }
    if (type === "1h") {
      start = formatLocalIso(new Date(now.getTime() - 60 * 60 * 1000));
      end = formatLocalIso(now);
    } else if (type === "6h") {
      start = formatLocalIso(new Date(now.getTime() - 6 * 60 * 60 * 1000));
      end = formatLocalIso(now);
    } else if (type === "24h") {
      start = formatLocalIso(new Date(now.getTime() - 24 * 60 * 60 * 1000));
      end = formatLocalIso(now);
    } else if (type === "today") {
      const s = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 0, 0, 0);
      start = formatLocalIso(s);
      end = formatLocalIso(now);
    } else if (type === "yesterday") {
      const s = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1, 0, 0, 0);
      const e = new Date(now.getFullYear(), now.getMonth(), now.getDate() - 1, 23, 59, 59);
      start = formatLocalIso(s);
      end = formatLocalIso(e);
    } else if (type === "7d") {
      start = formatLocalIso(new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000));
      end = formatLocalIso(now);
    } else if (type === "30d") {
      start = formatLocalIso(new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000));
      end = formatLocalIso(now);
    }
  };
</script>

<div class="rounded-lg border border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50 p-2.5 space-y-2">
  <div class="flex items-center justify-between text-xs font-semibold text-gray-700 dark:text-gray-300">
    <div class="flex items-center gap-1.5">
      <Icon path={icons.mdiClockOutline} size={0.8} />
      <span>{$_("TimeRange.TimeRange")}</span>
    </div>
    <span class="text-[11px] font-normal text-gray-500 dark:text-gray-400">{$_("TimeRange.Shortcuts")}</span>
  </div>

  <!-- Shortcuts Pill Bar -->
  <div class="flex flex-wrap items-center gap-1.5">
    <button
      type="button"
      onclick={() => applyShortcut("1h")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Last1h")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("6h")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Last6h")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("24h")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Last24h")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("today")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Today")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("yesterday")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Yesterday")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("7d")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Last7d")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("30d")}
      class="rounded border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-700 px-2 py-0.5 text-[11px] font-medium text-gray-700 dark:text-gray-300 hover:border-blue-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Last30d")}
    </button>
    <button
      type="button"
      onclick={() => applyShortcut("clear")}
      class="rounded border border-red-300 dark:border-red-800 bg-red-50 dark:bg-red-950/40 px-2 py-0.5 text-[11px] font-semibold text-red-600 dark:text-red-400 hover:bg-red-100 dark:hover:bg-red-900/60 transition-colors cursor-pointer"
    >
      {$_("TimeRange.Clear")}
    </button>
  </div>

  <!-- Datetime Inputs -->
  <div class="flex items-end gap-2 pt-1">
    <Label class="space-y-1 text-xs flex-1">
      <span>{$_("TimeRange.Start")}</span>
      <Input
        class="h-9 text-xs"
        type="datetime-local"
        bind:value={start}
        size="sm"
      />
    </Label>
    <Label class="space-y-1 text-xs flex-1">
      <span>{$_("TimeRange.End")}</span>
      <Input
        class="h-9 text-xs"
        type="datetime-local"
        bind:value={end}
        size="sm"
      />
    </Label>
    {#if start || end}
      <Button
        class="!p-2 w-9 h-9 shrink-0"
        color="red"
        onclick={() => applyShortcut("clear")}
        title={$_("TimeRange.Clear")}
      >
        <Icon path={icons.mdiCancel} size={1} />
      </Button>
    {/if}
  </div>
</div>
