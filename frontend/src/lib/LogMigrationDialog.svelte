<script lang="ts">
  import { Modal, GradientButton, Progressbar, Alert, Spinner } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
  import { GetLogStoreInfo, MigrateLogStoreToParquet } from "../../wailsjs/go/main/App";

  export let show: boolean = false;

  const dispatch = createEventDispatcher();

  let progress = {
    current: 0,
    total: 0,
    percent: 0,
    bucket: "",
    status: "idle", // idle, migrating, done, error
    message: "",
  };

  let stats: any = null;
  let error = "";

  const loadStats = async () => {
    try {
      const info = await GetLogStoreInfo();
      if (info && info.stats) {
        stats = info.stats;
      }
    } catch (e: any) {
      console.error(e);
    }
  };

  let autoClosing = false;

  const handleDone = () => {
    if (autoClosing) return;
    autoClosing = true;
    dispatch("done");
    setTimeout(() => {
      closeDialog();
      autoClosing = false;
    }, 1000);
  };

  const startMigration = async () => {
    progress = {
      current: 0,
      total: stats?.totalCount || 0,
      percent: 0,
      bucket: "init",
      status: "migrating",
      message: $_("Migration.Starting"),
    };
    error = "";
    autoClosing = false;

    try {
      await MigrateLogStoreToParquet();
      progress.status = "done";
      progress.percent = 100;
      handleDone();
    } catch (e: any) {
      progress.status = "error";
      error = e?.message || String(e);
    }
  };

  const handleProgress = (data: any) => {
    if (data) {
      progress = {
        current: data.current || 0,
        total: data.total || 0,
        percent: data.percent || 0,
        bucket: data.bucket || "",
        status: data.status || progress.status,
        message: data.message || "",
      };
      if (progress.status === "error") {
        error = data.message;
      } else if (progress.status === "done") {
        handleDone();
      }
    }
  };

  const closeDialog = () => {
    const wasDone = progress.status === "done";
    show = false;
    dispatch("close", { done: wasDone });
  };

  $: if (show && progress.status !== "migrating") {
    if (progress.status === "done" || progress.status === "error") {
      progress = {
        current: 0,
        total: 0,
        percent: 0,
        bucket: "",
        status: "idle",
        message: "",
      };
      error = "";
    }
    loadStats();
  }

  onMount(() => {
    EventsOn("log-migration-progress", handleProgress);
  });

  onDestroy(() => {
    EventsOff("log-migration-progress");
  });
</script>

<Modal bind:open={show} size="md" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <div class="flex items-center space-x-2 border-b pb-2 dark:border-gray-700">
      <Icon path={icons.mdiDatabaseArrowRight} size={1.2} class="text-blue-500" />
      <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
        {$_("Migration.Title")}
      </h3>
    </div>

    {#if error}
      <Alert color="red">
        <div class="flex items-center space-x-2">
          <Icon path={icons.mdiAlertCircle} size={1} />
          <span>{error}</span>
        </div>
      </Alert>
    {/if}

    {#if progress.status === "idle"}
      <div class="space-y-3 py-1 text-sm text-gray-700 dark:text-gray-300">
        <p class="leading-relaxed">
          {$_("Migration.ConfirmDescription")}
        </p>

        {#if stats}
          <div class="bg-gray-50 dark:bg-gray-800 p-3 rounded-lg border dark:border-gray-700 text-xs space-y-1">
            <div class="flex justify-between font-semibold text-sm border-b pb-1 dark:border-gray-700">
              <span>{$_("Migration.TotalRecords")}</span>
              <span class="text-blue-600 dark:text-blue-400 font-mono">{stats.totalCount.toLocaleString()}</span>
            </div>
            {#if stats.syslogCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>Syslog:</span>
                <span class="font-mono">{stats.syslogCount.toLocaleString()}</span>
              </div>
            {/if}
            {#if stats.trapCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>SNMP Trap:</span>
                <span class="font-mono">{stats.trapCount.toLocaleString()}</span>
              </div>
            {/if}
            {#if stats.pollingLogCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>Polling Logs:</span>
                <span class="font-mono">{stats.pollingLogCount.toLocaleString()}</span>
              </div>
            {/if}
            {#if stats.netflowCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>NetFlow / IPFIX:</span>
                <span class="font-mono">{stats.netflowCount.toLocaleString()}</span>
              </div>
            {/if}
            {#if stats.sFlowCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>sFlow:</span>
                <span class="font-mono">{stats.sFlowCount.toLocaleString()}</span>
              </div>
            {/if}
            {#if stats.arpLogCount > 0}
              <div class="flex justify-between text-gray-600 dark:text-gray-400">
                <span>ARP Log:</span>
                <span class="font-mono">{stats.arpLogCount.toLocaleString()}</span>
              </div>
            {/if}
          </div>
        {/if}
      </div>

      <div class="flex justify-end space-x-2 pt-2 border-t dark:border-gray-700">
        <GradientButton
          shadow
          type="button"
          color="red"
          onclick={startMigration}
          size="xs"
        >
          <Icon path={icons.mdiPlay} size={1} />
          {$_("Migration.Start")}
        </GradientButton>
        <GradientButton
          shadow
          type="button"
          color="teal"
          onclick={closeDialog}
          size="xs"
        >
          <Icon path={icons.mdiCancel} size={1} />
          {$_("Migration.Cancel")}
        </GradientButton>
      </div>
    {:else}
      <div class="flex flex-col space-y-2 py-2">
        <div class="flex justify-between text-sm font-medium text-gray-700 dark:text-gray-300">
          <span class="flex items-center gap-2">
            {#if progress.status === "migrating"}
              <Spinner size="4" />
            {:else if progress.status === "done"}
              <Icon path={icons.mdiCheckCircle} size={0.9} class="text-green-500" />
            {/if}
            {progress.message || $_("Migration.Processing")}
          </span>
          {#if progress.total > 0}
            <span>{progress.current.toLocaleString()} / {progress.total.toLocaleString()}</span>
          {/if}
        </div>

        <Progressbar
          progress={progress.percent}
          size="h-5"
          color={progress.status === "error" ? "red" : progress.status === "done" ? "green" : "blue"}
          labelInside
        />

        {#if progress.bucket && progress.bucket !== "done" && progress.bucket !== "init"}
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {$_("Migration.CurrentBucket")}: <span class="font-mono font-semibold">{progress.bucket}</span>
          </div>
        {/if}
      </div>

      {#if progress.status === "done"}
        <div class="p-3 bg-green-50 dark:bg-green-950/40 rounded-lg text-sm text-green-700 dark:text-green-300">
          {$_("Migration.DoneDescription")}
        </div>
      {/if}

      <div class="flex justify-end space-x-2 pt-2 border-t dark:border-gray-700">
        <GradientButton
          shadow
          type="button"
          color={progress.status === "done" ? "green" : "teal"}
          disabled={progress.status === "migrating"}
          onclick={closeDialog}
          size="xs"
        >
          <Icon path={icons.mdiClose} size={1} />
          {$_("Migration.Close")}
        </GradientButton>
      </div>
    {/if}
  </div>
</Modal>
