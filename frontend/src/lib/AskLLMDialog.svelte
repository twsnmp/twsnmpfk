<script lang="ts">
  import { marked } from "marked";
  import { Modal, GradientButton, Alert } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import DOMPurify from "dompurify";
  import { _ } from "svelte-i18n";

  export let show = false;
  export let content = "";
  export let error = "";

  $: renderedContent = DOMPurify.sanitize(marked.parse(content || "") as string);
</script>

<Modal bind:open={show} size="lg" dismissable={false} class="w-full">
  <div class="flex flex-col max-h-[70vh]">
    {#if error}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {error}
        </div>
      </Alert>
    {/if}
    <div
      class="overflow-y-auto p-4 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
    >
      <article
        class="prose prose-sm md:prose-base dark:prose-invert max-w-none"
      >
        {@html renderedContent}
      </article>
    </div>

    <div class="flex justify-end space-x-2 mt-4">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => (show = false)}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>
