<script lang="ts">
  import "maplibre-gl/dist/maplibre-gl.css";
  import { Map, NavigationControl, Marker } from "maplibre-gl";
  import { GradientButton, Modal, Label, Select} from "flowbite-svelte";
  import { getIcon, getStateColor } from "./common";
  import {Icon} from "mdi-svelte-ts";
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

  let showEditNode = false;
  let showNodeReport = false;
  let showPolling = false;
  let selectedNode = "";
  let nodes = undefined;
  let map :any= undefined;
  let markers: any = [];
  let locConf: any = undefined;
  let nodeList: any = undefined;
  let showAddNode = false;
  let addNodeID = "";
  let lastLoc = "";
  let timer: any = undefined;
  let lock = false;

  const refresh = async () => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    if (!map) {
      await makeMap();
    }
    if (!map) {
      return;
    }
    if (markers && Array.isArray(markers)) {
      for (const m of markers) {
        if (m && typeof m.remove === "function") {
          m.remove();
        }
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

  const addNodeMarker = (n:any) => {
    if (!map) {
      return;
    }
    const icon = getIcon(n.Icon);
    const color = getStateColor(n.State);
    const divSize = (locConf?.IconSize || 24) + 8;
    const iconSize = locConf?.IconSize || 24;
    const nodeDiv = document.createElement("div");
    nodeDiv.classList.add("node");
    nodeDiv.innerHTML = `
    <div class="icon" style="height: ${divSize}px;width: ${divSize}px;background-color: ${color}; color: white;font-size: ${
      iconSize
    }px;text-align: center;line-height: ${divSize}px;">
			<span class="mdi ${icon}"></span>
		</div>
		<div style="font-size: ${iconSize / 2}px;text-align: center;">${
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
      .addTo(map);
    marker.on("dragend", (e: any) => {
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
    if (!locConf || !locConf.Style || !locConf.Style.trim()) {
      return;
    }
    let s = locConf.Style;
    if (locConf.Style && locConf.Style.trim().startsWith("{")) {
      try {
        s = JSON.parse(locConf.Style);
      } catch (e) {
        console.error("Failed to parse map style JSON", e);
        return;
      }
    }
    try {
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
    } catch (e) {
      console.error("Failed to initialize map", e);
      map = undefined;
      return;
    }
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
    if (markers && Array.isArray(markers)) {
      for (const m of markers) {
        if (m && typeof m.remove === "function") {
          m.remove();
        }
      }
    }
    markers = [];
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

<div class="flex flex-col max-h-[calc(100vh-50px)] overflow-y-auto pb-2">
  <div class="mx-2 mt-2">
    <div id="map"></div>
  </div>
  <div class="flex justify-end space-x-2 mr-2 py-2">
    {#if selectedNode != ""}
      {#if !lock}
        <GradientButton
          shadow
          color="blue"
          type="button"
          onclick={edit}
          size="xs"
        >
          <Icon path={icons.mdiPencil} size={1} />
          {$_('Location.Edit')}
        </GradientButton>
        <GradientButton
          shadow
          color="blue"
          type="button"
          onclick={polling}
          size="xs"
        >
          <Icon path={icons.mdiLanCheck} size={1} />
          {$_('Location.Polling')}
        </GradientButton>
        <GradientButton
          shadow
          color="red"
          type="button"
          onclick={del}
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
        onclick={report}
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
        onclick={saveDef}
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
      onclick={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_('Location.Reload')}
    </GradientButton>
  </div>
</div>

<Node bind:show={showEditNode}
  nodeID={selectedNode}
  on:close={(e) => {
    refresh();
  }}
/>

<NodeReport bind:show={showNodeReport} id={selectedNode} />
<NodePolling bind:show={showPolling} nodeID={selectedNode} />


<Modal bind:open={showAddNode} size="sm" dismissable={false} outsideclose={false}>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">{$_('Location.SelectNode')}</h3>
    <Label class="space-y-2 text-xs">
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
        onclick={add}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_('Location.Add')}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        onclick={() => {
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
    max-height: calc(100vh - 120px);
    width: 95vw;
    margin: 0 auto;
  }
  :global(div.node.selected div.icon) {
    border: 2px solid #00f;
  }
</style>
