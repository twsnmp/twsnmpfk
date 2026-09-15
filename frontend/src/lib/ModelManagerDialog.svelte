<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import {
    Modal,
    Button,
    Alert,
    Input,
    Select,
    Progressbar,
    Badge,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import {
    GetAIHardwareStatus,
    GetLocalModels,
    GetModelPresets,
    DownloadModel,
    CancelModelDownload,
    DeleteLocalModel,
    DownloadGPULibrary,
    CancelGPUDownload,
  } from "../../wailsjs/go/main/App";
  import { EventsOn, EventsOff } from "../../wailsjs/runtime/runtime";
  import type { main, model } from "../../wailsjs/go/models";

  export let show = false;

  const dispatch = createEventDispatcher();

  let hardwareStatus: main.AIHardwareStatus = {
    acceleration: "",
    detail: "",
    model_dir: "",
    lib_dir: "",
    wgpu_lib_path: "",
    has_gpu_lib: false,
  };

  let localModels: model.ModelInfo[] = [];
  let presets: model.PresetModelInfo[] = [];

  let selectedPreset = "";
  let customTarget = "";

  let isDownloadingModel = false;
  let modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };

  let isSettingUpGPU = false;
  let gpuProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };

  let errorMsg = "";
  let infoMsg = "";

  const refreshData = async () => {
    try {
      hardwareStatus = await GetAIHardwareStatus();
      localModels = (await GetLocalModels()) || [];
      presets = (await GetModelPresets()) || [];
      if (presets.length > 0 && !selectedPreset) {
        selectedPreset = presets[0].name;
      }
    } catch (e: any) {
      errorMsg = e?.message || String(e);
    }
  };

  $: if (show) {
    refreshData();
    errorMsg = "";
    infoMsg = "";
  }

  const handleModelProgress = (data: any) => {
    isDownloadingModel = true;
    modelProgress = {
      percent: data.percent || 0,
      downloaded_human: data.downloaded_human || "0 B",
      total_human: data.total_human || "0 B",
    };
  };

  const handleGPUProgress = (data: any) => {
    isSettingUpGPU = true;
    gpuProgress = {
      percent: data.percent || 0,
      downloaded_human: data.downloaded_human || "0 B",
      total_human: data.total_human || "0 B",
    };
  };

  onMount(() => {
    EventsOn("model_download_progress", handleModelProgress);
    EventsOn("gpu_download_progress", handleGPUProgress);
  });

  onDestroy(() => {
    EventsOff("model_download_progress");
    EventsOff("gpu_download_progress");
  });

  const downloadPreset = async () => {
    if (!selectedPreset) return;
    errorMsg = "";
    infoMsg = "";
    isDownloadingModel = true;
    modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadModel(selectedPreset);
      isDownloadingModel = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = `${selectedPreset} downloaded successfully.`;
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e: any) {
      isDownloadingModel = false;
      errorMsg = e?.message || String(e);
    }
  };

  const downloadCustom = async () => {
    const target = customTarget.trim();
    if (!target) return;
    errorMsg = "";
    infoMsg = "";
    isDownloadingModel = true;
    modelProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadModel(target);
      isDownloadingModel = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = `Model ${target} downloaded successfully.`;
        customTarget = "";
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e: any) {
      isDownloadingModel = false;
      errorMsg = e?.message || String(e);
    }
  };

  const cancelModel = async () => {
    try {
      await CancelModelDownload();
      isDownloadingModel = false;
    } catch (e: any) {
      errorMsg = e?.message || String(e);
    }
  };

  const deleteModel = async (name: string) => {
    if (!confirm($_("ModelManager.DeleteConfirm", { values: { name } }))) {
      return;
    }
    errorMsg = "";
    infoMsg = "";
    try {
      const err = await DeleteLocalModel(name);
      if (err && err !== "cancel") {
        errorMsg = err;
      } else if (!err) {
        infoMsg = `Model ${name} deleted successfully.`;
        await refreshData();
        dispatch("modelsChanged");
      }
    } catch (e: any) {
      errorMsg = e?.message || String(e);
    }
  };

  const setupGPU = async () => {
    errorMsg = "";
    infoMsg = "";
    isSettingUpGPU = true;
    gpuProgress = { percent: 0, downloaded_human: "0 B", total_human: "0 B" };
    try {
      const err = await DownloadGPULibrary();
      isSettingUpGPU = false;
      if (err) {
        errorMsg = err;
      } else {
        infoMsg = "WebGPU library installed successfully.";
        await refreshData();
      }
    } catch (e: any) {
      isSettingUpGPU = false;
      errorMsg = e?.message || String(e);
    }
  };

  const cancelGPU = async () => {
    try {
      await CancelGPUDownload();
      isSettingUpGPU = false;
    } catch (e: any) {
      errorMsg = e?.message || String(e);
    }
  };

  const close = () => {
    show = false;
    dispatch("close");
  };

  $: presetOptions = presets.map((p) => ({
    value: p.name,
    name: `${p.name} (${p.size}, ${p.description})`,
  }));
</script>

<Modal bind:open={show} size="lg" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <!-- Header -->
    <div class="flex items-center justify-between border-b pb-2 dark:border-gray-700">
      <div class="flex items-center space-x-2">
        <Icon path={icons.mdiChip} size={1.2} />
        <h3 class="text-lg font-semibold text-gray-900 dark:text-white">
          {$_("ModelManager.Title")}
        </h3>
      </div>
      <button
        type="button"
        class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200"
        onclick={close}
      >
        <Icon path={icons.mdiClose} size={1} />
      </button>
    </div>

    {#if errorMsg}
      <Alert color="red">
        <div class="flex items-center justify-between w-full">
          <div class="flex items-center space-x-2">
            <Icon path={icons.mdiAlertCircle} size={1} />
            <span>{errorMsg}</span>
          </div>
          <button type="button" onclick={() => (errorMsg = "")}>
            <Icon path={icons.mdiClose} size={0.8} />
          </button>
        </div>
      </Alert>
    {/if}

    {#if infoMsg}
      <Alert color="green">
        <div class="flex items-center justify-between w-full">
          <div class="flex items-center space-x-2">
            <Icon path={icons.mdiCheckCircle} size={1} />
            <span>{infoMsg}</span>
          </div>
          <button type="button" onclick={() => (infoMsg = "")}>
            <Icon path={icons.mdiClose} size={0.8} />
          </button>
        </div>
      </Alert>
    {/if}

    <div class="space-y-4 max-h-[70vh] overflow-y-auto pr-1">
      <!-- Hardware Acceleration Status -->
      <div class="p-3 bg-gray-50 dark:bg-gray-800 rounded-lg border dark:border-gray-700 space-y-2">
        <div class="flex items-center space-x-2">
          <Icon path={icons.mdiSpeedometer} size={1} />
          <h4 class="font-semibold text-sm text-gray-900 dark:text-white">
            {$_("ModelManager.GPUAcceleration")}
          </h4>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs">
          <span class="font-medium text-gray-700 dark:text-gray-300">
            {$_("ModelManager.ActiveBackend")}:
          </span>
          {#if hardwareStatus.acceleration === "GPU"}
            <Badge color="green">GPU</Badge>
          {:else if hardwareStatus.acceleration && hardwareStatus.acceleration.includes("SIMD")}
            <Badge color="blue">SIMD</Badge>
          {:else}
            <Badge color="gray">CPU</Badge>
          {/if}
          <span class="text-gray-500 dark:text-gray-400">{hardwareStatus.detail}</span>
        </div>

        <div class="flex flex-wrap items-center gap-2 text-xs pt-1">
          <span class="font-medium text-gray-700 dark:text-gray-300">
            {$_("ModelManager.WGPULibrary")}:
          </span>
          {#if hardwareStatus.has_gpu_lib}
            <span class="text-green-600 dark:text-green-400 flex items-center gap-1 font-semibold">
              <Icon path={icons.mdiCheckCircle} size={0.7} />
              {$_("ModelManager.GPULibInstalled")}
            </span>
            <span class="text-gray-400 truncate max-w-xs" title={hardwareStatus.wgpu_lib_path}>
              ({hardwareStatus.wgpu_lib_path})
            </span>
          {:else}
            <span class="text-yellow-600 dark:text-yellow-400 font-semibold">
              {$_("ModelManager.GPULibNotInstalled")}
            </span>
            {#if isSettingUpGPU}
              <div class="flex items-center gap-2 flex-1 min-w-[200px]">
                <div class="flex-1">
                  <Progressbar progress={gpuProgress.percent} size="h-2" />
                </div>
                <span class="text-xs text-gray-500 whitespace-nowrap">
                  {gpuProgress.percent}% ({gpuProgress.downloaded_human} / {gpuProgress.total_human})
                </span>
                <Button size="xs" color="red" onclick={cancelGPU}>
                  {$_("ModelManager.Cancel")}
                </Button>
              </div>
            {:else}
              <Button size="xs" color="blue" onclick={setupGPU}>
                <span class="mr-1 inline-flex items-center"><Icon path={icons.mdiDownload} size={0.8} /></span>
                {$_("ModelManager.SetupGPU")}
              </Button>
            {/if}
          {/if}
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 pt-1">
          {$_("ModelManager.GPUHelp")}
        </p>
      </div>

      <!-- Downloaded Models List -->
      <div class="border dark:border-gray-700 rounded-lg overflow-hidden">
        <div class="p-3 bg-gray-100 dark:bg-gray-700 flex items-center justify-between">
          <h4 class="font-semibold text-sm text-gray-900 dark:text-white flex items-center gap-2">
            <Icon path={icons.mdiDatabase} size={0.9} />
            {$_("ModelManager.LocalModels")}
          </h4>
          <Badge color="gray">{localModels.length}</Badge>
        </div>
        {#if localModels.length === 0}
          <div class="p-6 text-center text-xs text-gray-500 dark:text-gray-400">
            {$_("ModelManager.NoModels")}
          </div>
        {:else}
          <div class="overflow-x-auto max-h-48 overflow-y-auto">
            <table class="w-full text-xs text-left text-gray-700 dark:text-gray-300">
              <thead class="bg-gray-50 dark:bg-gray-800 text-gray-500 border-b dark:border-gray-700 sticky top-0">
                <tr>
                  <th class="p-2">{$_("ModelManager.Name")}</th>
                  <th class="p-2">{$_("ModelManager.Size")}</th>
                  <th class="p-2">{$_("ModelManager.ModTime")}</th>
                  <th class="p-2 text-right">{$_("ModelManager.Action")}</th>
                </tr>
              </thead>
              <tbody class="divide-y dark:divide-gray-700">
                {#each localModels as m}
                  <tr class="hover:bg-gray-50 dark:hover:bg-gray-800">
                    <td class="p-2 font-medium text-gray-900 dark:text-white truncate max-w-[200px]" title={m.name}>
                      {m.name}
                    </td>
                    <td class="p-2 whitespace-nowrap">{m.size_human}</td>
                    <td class="p-2 whitespace-nowrap text-gray-500">
                      {new Date(m.mod_time).toLocaleString()}
                    </td>
                    <td class="p-2 text-right whitespace-nowrap">
                      <Button size="xs" color="red" outline onclick={() => deleteModel(m.name)}>
                        <span class="mr-1 inline-flex items-center"><Icon path={icons.mdiTrashCan} size={0.8} /></span>
                        {$_("ModelManager.Delete")}
                      </Button>
                    </td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {/if}
      </div>

      <!-- Model Download Progress Banner -->
      {#if isDownloadingModel}
        <div class="p-3 bg-blue-50 dark:bg-blue-950 border border-blue-300 dark:border-blue-800 rounded-lg space-y-2">
          <div class="flex items-center justify-between text-xs text-blue-900 dark:text-blue-100 font-semibold">
            <span>
              {$_("ModelManager.Downloading")} {modelProgress.percent}% ({modelProgress.downloaded_human} / {modelProgress.total_human})
            </span>
            <Button size="xs" color="red" onclick={cancelModel}>
              {$_("ModelManager.Cancel")}
            </Button>
          </div>
          <Progressbar progress={modelProgress.percent} size="h-2" color="blue" />
        </div>
      {/if}

      <!-- Preset Model Download -->
      <div class="p-3 border dark:border-gray-700 rounded-lg space-y-2">
        <h4 class="font-semibold text-sm text-gray-900 dark:text-white">
          {$_("ModelManager.DownloadPreset")}
        </h4>
        <p class="text-xs text-gray-500 dark:text-gray-400">
          {$_("ModelManager.PresetHelp")}
        </p>
        <div class="flex gap-2 items-center">
          <div class="flex-1">
            <Select
              items={presetOptions}
              bind:value={selectedPreset}
              size="sm"
              disabled={isDownloadingModel}
            />
          </div>
          <Button
            size="sm"
            color="blue"
            disabled={isDownloadingModel || !selectedPreset}
            onclick={downloadPreset}
          >
            <span class="mr-1 inline-flex items-center"><Icon path={icons.mdiDownload} size={0.9} /></span>
            {$_("ModelManager.Download")}
          </Button>
        </div>
      </div>

      <!-- Custom Model Download -->
      <div class="p-3 border dark:border-gray-700 rounded-lg space-y-2">
        <h4 class="font-semibold text-sm text-gray-900 dark:text-white">
          {$_("ModelManager.DownloadCustom")}
        </h4>
        <div class="flex gap-2 items-center">
          <div class="flex-1">
            <Input
              type="text"
              size="sm"
              placeholder={$_("ModelManager.CustomPlaceholder")}
              bind:value={customTarget}
              disabled={isDownloadingModel}
            />
          </div>
          <Button
            size="sm"
            color="alternative"
            disabled={isDownloadingModel || !customTarget.trim()}
            onclick={downloadCustom}
          >
            <span class="mr-1 inline-flex items-center"><Icon path={icons.mdiDownload} size={0.9} /></span>
            {$_("ModelManager.Download")}
          </Button>
        </div>
      </div>
    </div>

    <!-- Footer -->
    <div class="flex justify-end pt-2 border-t dark:border-gray-700">
      <Button color="alternative" size="sm" onclick={close}>
        {$_("ModelManager.Close")}
      </Button>
    </div>
  </div>
</Modal>
