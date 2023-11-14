<script lang="ts">
  import "maplibre-gl/dist/maplibre-gl.css";
  import { Map, NavigationControl, Marker } from "maplibre-gl";
  import { GradientButton, Modal, Label, Select,Toast } from "flowbite-svelte";
  import { getIcon, getStateColor } from "./common";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import { onMount, onDestroy } from "svelte";
  import {
    GetNodes,
    GetLocConf,
    UpdateLocConf,
    UpdateNodeLoc,
    GetSettings,
  } from "../../wailsjs/go/main/App";
  import Node from "./Node.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import { _ } from "svelte-i18n";
  import { time } from "echarts";

  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let selectedNode = "";
  let nodes = undefined;
  let map = undefined;
  let markers = undefined;
  let locConf = undefined;
  let nodeList = undefined;
  let showAddNode = false;
  let addNodeID = "";
  let lastLoc = "";
  let timer = undefined;
  let lock = false;

  const refresh = async () => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    if (!map) {
      await makeMap();
    }
    if (markers) {
      for (const m of markers) {
        m.remove();
      }
    }
    markers = [];
    nodeList = [];
    nodes = await GetNodes();
    for (const k in nodes) {
      if (nodes[k].Loc) {
        addNodeMarker(nodes[k]);
      } else {
        nodeList.push({
          name: nodes[k].Name,
          value: k,
        });
      }
    }
    timer = setTimeout(refresh, 1000 * 30);
  };

  const getLngLat = (loc: string): [number, number] => {
    const a = loc.split(",");
    if (a.length < 2) {
      return [0, 0];
    }
    return [Number(a[0]), Number(a[1])];
  };

  const addNodeMarker = (n) => {
    const icon = getIcon(n.Icon);
    const color = getStateColor(n.State);
    const divSize = locConf.IconSize + 8;
    const nodeDiv = document.createElement("div");
    nodeDiv.classList.add("node");
    nodeDiv.innerHTML = `
    <div class="icon" style="height: ${divSize}px;width: ${divSize}px;background-color: ${color}; color: white;font-size: ${
      locConf.IconSize
    }px;text-align: center;line-height: ${divSize}px;">
			<span class="mdi ${icon}"></span>
		</div>
		<div style="font-size: ${locConf.IconSize / 2}px;text-align: center;">${
      n.Name
    }</div>`;
    nodeDiv.onclick = () => {
      if(lock) {
        return;
      }
      if (nodeDiv.classList.contains("selected")) {
        nodeDiv.classList.remove("selected");
        selectedNode = "";
      } else {
        for (const e of document.getElementsByClassName("node")) {
          e.classList.remove("selected");
        }
        nodeDiv.classList.add("selected");
        selectedNode = n.ID;
      }
    };

    const marker = new Marker({ draggable: true, element: nodeDiv })
      .setLngLat(getLngLat(n.Loc))
      .addTo(map)
      .on("dragend", (e) => {
        if (lock) {
          return;
        }
        const loc = e.target.getLngLat();
        UpdateNodeLoc(n.ID, loc.lng + "," + loc.lat);
      });
    markers.push(marker);
  };

  const makeMap = async () => {
    // style: "https://tile.openstreetmap.jp/styles/osm-bright-ja/style.json",
    locConf = await GetLocConf();
    const s = locConf.Style.startsWith("{") ? JSON.parse(locConf.Style) : locConf.Style;
    map = new Map({
      container: "map",
      style: s,
      center: getLngLat(locConf.Center),
      zoom: locConf.Zoom,
    });
    map.on("contextmenu", (e: any) => {
      if(lock) {
        return;
      }
      lastLoc = e.lngLat.lng + "," + e.lngLat.lat;
      if (lastLoc != "") {
        showAddNode = true;
      }
    });
    map.addControl(
      new NavigationControl({
        visualizePitch: true,
      })
    );
    const setting = await GetSettings();
    lock = setting.Lock != "";
  };

  const edit = () => {
    if (!selectedNode || lock) {
      return;
    }
    showEditNode = true;
  };

  const report = () => {
    if (!selectedNode) {
      return;
    }
    showNodeReport = true;
  };

  const polling = () => {
    if (!selectedNode) {
      return;
    }
    showPolling = true;
  };

  const add = async () => {
    if (!addNodeID || !lastLoc) {
      return;
    }
    await UpdateNodeLoc(addNodeID, lastLoc);
    showAddNode = false;
    refresh();
  };

  const del = async () => {
    if (!selectedNode || lock) {
      return;
    }
    await UpdateNodeLoc(selectedNode, "");
    refresh();
  };

  let inSaveDef = false;
  const saveDef = async() => {
    if (!map || lock) {
      return;
    }
    inSaveDef = true;
    const c = map.getCenter();
    locConf.Zoom = map.getZoom();
    locConf.Center = c.lng + "," + c.lat;
    await UpdateLocConf(locConf);
    inSaveDef = false;
  };

  onMount(() => {
    refresh();
  });

  onDestroy(() => {
    if(markers) {
      for(const m of markers) {
        m.remove();
      }
    }
    if (map) {
      map.remove();
      map = undefined;
    }
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
  });
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <div id="map" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedNode != ""}
      {#if !lock}
        <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={edit}
          size="xs"
        >
          <Icon path={icons.mdiPencil} size={1} />
          {$_('Location.Edit')}
        </GradientButton>
        <GradientButton
          shadow
          color="blue"
          type="button"
          on:click={polling}
          size="xs"
        >
          <Icon path={icons.mdiLanCheck} size={1} />
          {$_('Location.Polling')}
        </GradientButton>
        <GradientButton
          shadow
          color="red"
          type="button"
          on:click={del}
          size="xs"
        >
          <Icon path={icons.mdiTrashCan} size={1} />
          {$_('Location.Del')}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={report}
        size="xs"
      >
        <Icon path={icons.mdiChartBar} size={1} />
        {$_('Location.Report')}
      </GradientButton>
    {/if}
    {#if !lock}
      <GradientButton
        shadow
        type="button"
        color="red"
        disabled={inSaveDef}
        on:click={saveDef}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_('Location.SaveDef')}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_('Location.Reload')}
    </GradientButton>
  </div>
</div>

{#if showEditNode}
  <Node
    nodeID={selectedNode}
    on:close={(e) => {
      showEditNode = false;
      refresh();
    }}
  />
{/if}

{#if showNodeReport}
  <NodeReport
    id={selectedNode}
    on:close={(e) => {
      showNodeReport = false;
    }}
  />
{/if}

{#if showPolling}
  <NodePolling
    nodeID={selectedNode}
    on:close={(e) => {
      showPolling = false;
    }}
  />
{/if}

<Modal bind:open={showAddNode} size="sm" permanent>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">{$_('Location.SelectNode')}</h3>
    <Label class="space-y-2">
      <span> {$_('Location.Node')} </span>
      <Select
        items={nodeList}
        bind:value={addNodeID}
        placeholder="{$_('Location.SelectNode')}"
        size="sm"
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={add}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_('Location.Add')}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showAddNode = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_('Location.Cancel')}
      </GradientButton>
    </div>
  </form>
</Modal>

<style>
  #map {
    height: 80vh;
    width: 98vw;
    margin: 0 auto;
  }
  :global(div.node.selected div.icon) {
    border: 2px solid #00f;
  }
</style>
