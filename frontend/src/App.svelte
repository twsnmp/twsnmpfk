<script lang="ts">
  import { onMount, tick } from "svelte";
  import { HasDatastore } from "../wailsjs/go/main/App";
  import Top from "./lib/Top.svelte";
  import { _ } from 'svelte-i18n';
  import Wellcome from "./lib/Wellcome.svelte";
  let top: boolean = false;

  onMount(async () => {
    top = await HasDatastore();
  });

  const handleDone = (e) => {
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
    max-width: 98vw;
  }
</style>
