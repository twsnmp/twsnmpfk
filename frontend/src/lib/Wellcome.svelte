<script>
  import logo from "../assets/images/appicon.png";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { Card, GradientButton, Spinner } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import { _ } from "svelte-i18n";
  import { GetVersion, SelectDatastore } from "../../wailsjs/go/main/App";
  import { Quit } from "../../wailsjs/runtime/runtime";
  import Help from "./Help.svelte";

  const dispatch = createEventDispatcher();
  let version = "1.0.0(xxxxx)";
  let started = false;
  let showHelp = false;

  onMount(async () => {
    version = await GetVersion();
    const e = document.querySelector("html");
    if (e) {
      e.classList.add("dark");
    }
  });

  const select = async () => {
    started = true;
    const r = await SelectDatastore();
    if (r) {
      dispatch("done", true);
    } else {
      started = false;
    }
  };
</script>

<div class="grid h-screen place-items-center">
  <Card padding="xl" size="xl">
    <div class="flex justify-center">
      <img id="logo" class="margin" src={logo} alt="logo" />
    </div>
    <div class="flex justify-center mt-5">
      <span
        class="text-xl font-semibold tracking-tight text-gray-900 dark:text-white"
      >
        TWSNMP FK
      </span>
      <span>
        <small>{version}</small>
      </span>
    </div>
    <div class="flex justify-center mt-4">
      <GradientButton
        shadow
        type="button"
        size="xl"
        color="teal"
        on:click={select}
        disabled={started}
      >
        {#if started}
          <Spinner class="mr-3" size="4" />
          <span>
            {$_("Wellcom.Starting")}
          </span>
        {:else}
          <Icon path={icons.mdiRun} size={1} />
          <span>
            {$_("Wellcome.Start")}
          </span>
        {/if}
      </GradientButton>
      {#if !started}
        <GradientButton
          shadow
          type="button"
          size="xl"
          color="red"
          class="ml-2"
          on:click={Quit}
        >
          <Icon path={icons.mdiStop} size={1} />
          <span>
            {$_("Wellcome.Stop")}
          </span>
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        size="xl"
        color="lime"
        class="ml-2"
        on:click={() => {
          showHelp = true;
        }}
      >
        <Icon path={icons.mdiHelp} size={1} />
        <span>
          {$_("Wellcom.Help")}
        </span>
      </GradientButton>
    </div>
  </Card>
</div>

<Help bind:show={showHelp} page="wellcome" />

<style>
  #logo {
    height: 512px;
    width: 512px;
    transform: rotateY(560deg);
    animation: turn 3.5s ease-out forwards 1s;
  }

  @keyframes turn {
    100% {
      transform: rotateY(0deg);
    }
  }
</style>
