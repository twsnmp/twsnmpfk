<script context="module" lang="ts">
  // retain module scoped expansion state for each tree node
  const _expansionState = {
    /* treeNodeId: expanded <boolean> */
  };
</script>

<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Tooltip } from "flowbite-svelte";
  export let tree;
  const { oid, name, MIBInfo, children } = tree;
  const dispatch = createEventDispatcher();

  let mibInfoTooltip = "";
  let type = "";
  if (MIBInfo) {
    mibInfoTooltip = MIBInfo.Description;
    type = ":" + MIBInfo.Type;
  }

  let expanded = _expansionState[oid] || false;
  const toggleExpansion = () => {
    expanded = _expansionState[oid] = !expanded;
  };
  $: arrowDown = expanded;
</script>

<ul>
  <li>
    {#if children && children.length > 0}
      <!-- svelte-ignore a11y-click-events-have-key-events -->
      <span on:click={toggleExpansion} class="hover:bg-sky-400">
        <span class="arrow" class:arrowDown>&#x25b6</span>
        <span
          on:dblclick={() => {
            dispatch("select", name);
          }}
        >
          {name}({oid}{type})
        </span>
        {#if mibInfoTooltip}
          <Tooltip trigger="click">{mibInfoTooltip}</Tooltip>
        {/if}
      </span>
      {#if expanded}
        {#each children as child}
          <svelte:self tree={child} on:select />
        {/each}
      {/if}
    {:else}
      <span class="hover:bg-sky-400">
        {#if type == ":Notification"}
          <span class="text-red-500">*</span>
        {:else}
          <span class="no-arrow" />
        {/if}
        <span
          on:dblclick={() => {
            dispatch("select", name);
          }}
        >
          {name}({oid}{type})
        </span>
        {#if mibInfoTooltip}
          <Tooltip trigger="click">
            <pre>{mibInfoTooltip}</pre>
          </Tooltip>
        {/if}
      </span>
    {/if}
  </li>
</ul>

<style>
  ul {
    margin: 0;
    list-style: none;
    padding-left: 1.2rem;
    user-select: none;
  }
  .no-arrow {
    padding-left: 1rem;
  }
  .arrow {
    cursor: pointer;
    display: inline-block;
  }
  .arrowDown {
    transform: rotate(90deg);
  }
</style>
