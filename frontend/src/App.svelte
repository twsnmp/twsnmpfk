<script lang="ts">
  import "prismjs/themes/prism.css";
  import { onMount } from "svelte";
  import { HasDatastore } from "../wailsjs/go/main/App";
  import Top from "./lib/Top.svelte";
  import { _ } from 'svelte-i18n';
  import Wellcome from "./lib/Wellcome.svelte";
  let top: boolean = false;

  onMount(async () => {
    top = await HasDatastore();
  });

  const handleDone = (e:any) => {
    if (e && e.detail  && e.detail) {
      top = e.detail;
    }
  }

</script>


{#if top}
  <Top/>
{:else}
  <Wellcome on:done={handleDone} />
{/if}

<style global>
  div[role=dialog] .max-w-7xl {
    max-width: 90vw;
  }
  div[active="true"] svg {
    color: rgba(23, 146, 227, 0.6);
  }
</style>
