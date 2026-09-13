<script lang="ts">
  import {
    Modal,
    Checkbox,
    GradientButton,
    Spinner,
    Badge,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import {
    DetectNodeType,
    ApplyNodeDetection,
  } from "../../wailsjs/go/main/App";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { getIcon } from "./common";
  import { _ } from "svelte-i18n";

  export let show: boolean = false;
  export let nodeID: string = "";

  let detecting = false;
  let applying = false;
  let result: any = undefined;
  let applyIcon = true;
  let applyPolling = true;

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    if (!nodeID) {
      return;
    }
    detecting = true;
    result = undefined;
    try {
      result = await DetectNodeType(nodeID);
      applyIcon = true;
      applyPolling = !!(result && result.SensorPollings && result.SensorPollings.length > 0);
    } catch (e) {
      console.error(e);
    } finally {
      detecting = false;
    }
  };

  const apply = async () => {
    if (!nodeID) {
      return;
    }
    applying = true;
    try {
      const ok = await ApplyNodeDetection(nodeID, applyIcon, applyPolling);
      if (ok) {
        dispatch("applied", result);
        close();
      }
    } catch (e) {
      console.error(e);
    } finally {
      applying = false;
    }
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  $: if (show) {
    onOpen();
  }
</script>

<Modal
  bind:open={show}
  size="lg"
  dismissable={false}
  class="w-full"
>
  <h3 class="mb-2 text-base font-semibold text-gray-900 dark:text-white flex items-center gap-2">
    <Icon path={icons.mdiAutoFix} size={1} />
    {$_("Node.AutoDetect") || "ノード種別自動判定"}
  </h3>

  {#if detecting}
    <div class="flex flex-col items-center justify-center p-8 space-y-3">
      <Spinner size="12" />
      <span class="text-sm text-gray-600 dark:text-gray-300">
        {$_("Node.Detecting") || "ノードのSNMP/HTTP/MAC情報を収集・判定中..."}
      </span>
    </div>
  {:else if result}
    <div class="space-y-4">
      <div class="flex items-center space-x-4 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
        <div class="flex-shrink-0 text-blue-600 dark:text-blue-400">
          <span class="mdi {getIcon(result.Icon)} text-5xl"></span>
        </div>
        <div class="flex-grow">
          <div class="flex items-center gap-2">
            <span class="text-lg font-bold text-gray-900 dark:text-white">{result.Name}</span>
            <Badge color="blue">{result.Category}</Badge>
            {#if result.SubType}
              <Badge color="indigo">{result.SubType}</Badge>
            {/if}
          </div>
          <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
            {$_("Node.RuleID") || "判定ルール"}: <code class="font-mono">{result.RuleID}</code>
            {#if result.Confidence}
              <span class="ml-3">スコア: {result.Confidence}</span>
            {/if}
          </div>
        </div>
      </div>

      <div class="space-y-2">
        <h4 class="text-xs font-semibold text-gray-700 dark:text-gray-300">
          {$_("Node.DetectedSensors") || "検出された推奨センサーポーリング"}
        </h4>
        {#if result.SensorPollings && result.SensorPollings.length > 0}
          <div class="overflow-x-auto border border-gray-200 dark:border-gray-700 rounded-lg">
            <table class="w-full text-xs text-left text-gray-500 dark:text-gray-400">
              <thead class="text-xs text-gray-700 uppercase bg-gray-100 dark:bg-gray-700 dark:text-gray-400">
                <tr>
                  <th scope="col" class="py-2 px-3">名前</th>
                  <th scope="col" class="py-2 px-3">パラメータ(MIBシンボル)</th>
                  <th scope="col" class="py-2 px-3">判定式</th>
                  <th scope="col" class="py-2 px-3">重要度</th>
                </tr>
              </thead>
              <tbody>
                {#each result.SensorPollings as sp}
                  <tr class="bg-white border-b dark:bg-gray-800 dark:border-gray-700">
                    <td class="py-2 px-3 font-medium text-gray-900 dark:text-white whitespace-nowrap">{sp.Name}</td>
                    <td class="py-2 px-3 font-mono text-xs">{sp.Params}</td>
                    <td class="py-2 px-3 font-mono text-xs">{sp.Script || '-'}</td>
                    <td class="py-2 px-3"><Badge color={sp.Level === 'high' ? 'red' : 'yellow'}>{sp.Level || 'low'}</Badge></td>
                  </tr>
                {/each}
              </tbody>
            </table>
          </div>
        {:else}
          <div class="text-xs text-gray-500 dark:text-gray-400 p-3 bg-gray-50 dark:bg-gray-800 rounded border border-gray-200 dark:border-gray-700">
            {$_("Node.NoSensorsDetected") || "この機種向けの推奨センサーポーリング（CPU/メモリ/温度等）はありません。"}
          </div>
        {/if}
      </div>

      <div class="flex flex-col sm:flex-row gap-4 pt-2 border-t border-gray-200 dark:border-gray-700">
        <Checkbox bind:checked={applyIcon}>
          {$_("Node.ApplyDetectIcon") || "アイコンと説明文を反映する"}
        </Checkbox>
        <Checkbox
          bind:checked={applyPolling}
          disabled={!result.SensorPollings || result.SensorPollings.length === 0}
        >
          {$_("Node.ApplyDetectPolling") || "センサーポーリングを追加する"}
        </Checkbox>
      </div>
    </div>
  {:else}
    <div class="text-sm text-red-500 p-4">
      {$_("Node.DetectFailed") || "ノード種別の検出に失敗しました。"}
    </div>
  {/if}

  <div class="flex justify-end space-x-2 mt-4 pt-2">
    {#if result}
      <GradientButton
        shadow
        color="blue"
        type="button"
        onclick={apply}
        disabled={applying || (!applyIcon && !applyPolling)}
        size="xs"
      >
        <Icon path={icons.mdiCheck} size={1} />
        {$_("Node.Apply") || "反映する"}
      </GradientButton>
      <GradientButton
        shadow
        color="green"
        type="button"
        onclick={onOpen}
        disabled={detecting || applying}
        size="xs"
      >
        <Icon path={icons.mdiRefresh} size={1} />
        {$_("Node.ReDetect") || "再判定"}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      type="button"
      color="teal"
      onclick={close}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("Node.Cancel") || "閉じる"}
    </GradientButton>
  </div>
</Modal>
