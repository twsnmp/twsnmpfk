<script lang="ts">
  import {
    Modal,
    GradientButton,
    Textarea,
    Alert,
    Spinner,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { _ } from "svelte-i18n";
  import { lang } from "../i18n/i18n";
  import { BrowserOpenURL } from "../../wailsjs/runtime";
  import { marked } from "marked";
  import DOMPurify from "dompurify";
  import { SendFeedback } from "../../wailsjs/go/main/App";

  export let page = "";
  export let show: boolean = false;
  let markdownContent = "";
  let renderedContent = "";
  let showFeedback = false;
  let feedback = "";
  let feedbackError = false;
  let sending = false;

  const dispatch = createEventDispatcher();

  const fetchHelp = async () => {
    if (!page) return;
    try {
      const response = await fetch(`help/${lang}/${page}.md`);
      if (response.ok) {
        const text = await response.text();
        markdownContent = text
          .replace(/^\n>>>\n/gm, "\n\n---\n\n")
          .replace(/^>>>$/gm, "---");
      } else {
        markdownContent = `Help not found for page: ${page}`;
      }
    } catch (err) {
      markdownContent = `Failed to load help: ${err}`;
    }
  };

  $: if (show && page) {
    fetchHelp();
  }

  $: renderedContent = DOMPurify.sanitize(marked.parse(markdownContent || "") as string);

  const close = () => {
    show = false;
    dispatch("close", {});
  };
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full help"
>
  <div class="flex flex-col h-[70vh]">
    <div
      class="flex-1 overflow-y-auto p-6 bg-gray-50 dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 help-content"
    >
      <article class="prose prose-sm md:prose-base dark:prose-invert max-w-none">
        {@html renderedContent}
      </article>
    </div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 mt-4">
    <GradientButton
      shadow
      type="button"
      color="lime"
      class="mr-2"
      size="xs"
      onclick={() => {
        BrowserOpenURL(
          `https://twsnmp.github.io/twsnmpfk/`
        );
      }}
    >
      <Icon path={icons.mdiFilePdfBox} size={1} />
      {$_("Help.Manual")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      class="mr-2"
      size="xs"
      onclick={() => {
        showFeedback = true;
      }}
    >
      <Icon path={icons.mdiChat} size={1} />
      {$_("Help.Feedback")}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      size="xs"
      onclick={close}
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("Help.Close")}
    </GradientButton>
  </div>
</Modal>

<Modal bind:open={showFeedback} size="md" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Help.Feedback")}
    </h3>
    {#if feedbackError}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {$_("Help.feedbackError")}
        </div>
      </Alert>
    {/if}
    <Textarea rows={10} bind:value={feedback}></Textarea>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="red"
        type="button"
        size="xs"
        onclick={async () => {
          if (feedback) {
            sending = true;
            feedbackError = false;
            if (await SendFeedback(feedback)) {
              showFeedback = false;
              sending = false;
              return;
            }
          }
          sending = false;
          feedbackError = true;
        }}
      >
        {#if sending}
          <Spinner class="me-3" size="4" />
        {:else}
          <Icon path={icons.mdiSend} size={1} />
        {/if}
        {$_("Help.Send")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        size="xs"
        onclick={() => {
          showFeedback = false;
        }}
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<style global>
  .help-content img {
    margin: 0 auto;
  }
</style>

