<script lang="ts">
  import {
    initMAP,
    updateMAP,
    resetMap,
    deleteMap,
    grid,
    setEditDrawItems,
    setShowNodeInfo,
    zoom,
    horizontal,
    vertical,
    circle,
    setMapReadOnly,
  } from "./map";
  import { onMount, onDestroy } from "svelte";
  import { Modal, GradientButton, Label, Input, Button } from "flowbite-svelte";
  import * as icons from "@mdi/js";
  import { Icon } from "mdi-svelte-ts";
  import Discover from "./Discover.svelte";
  import Node from "./Node.svelte";
  import Line from "./Line.svelte";
  import DrawItem from "./DrawItem.svelte";
  import NodeReport from "./NodeReport.svelte";
  import NodePolling from "./NodePolling.svelte";
  import Ping from "./Ping.svelte";
  import Network from "./Network.svelte";
  import NetworkReport from "./NetworkReport.svelte";
  import NetworkLines from "./NetworkLines.svelte";
  import NeighborNetworksAndLines from "./NeighborNetworksAndLines.svelte";
  import {
    CheckPolling,
    DeleteDrawItems,
    DeleteNodes,
    CopyNode,
    CopyDrawItem,
    WakeOnLan,
    GetNode,
    SelectFile,
    GetImage,
    GetBackImage,
    SetBackImage,
    ImportV4Map,
    DeleteNetwork,
    CheckNetwork,
    ExportMap,
    GetMapConf,
  } from "../../wailsjs/go/main/App";
  import { BrowserOpenURL,WindowReloadApp } from "../../wailsjs/runtime";
  import MIBBrowser from "./MIBBrowser.svelte";
  import GNMITool from "./GNMITool.svelte";
  import NodeDiagnoseDialog from "./NodeDiagnoseDialog.svelte";
  import { _ } from "svelte-i18n";
  import type { datastore } from "wailsjs/go/models";

  let map: any;
  let posX: number = 0;
  let posY: number = 0;
  let showMapMenu: boolean = false;
  let showNodeMenu: boolean = false;
  let showDrawItemMenu: boolean = false;
  let showNetworkMenu: boolean = false;
  let showFormatNodesMenu: boolean = false;
  let showEditNode: boolean = false;
  let selectedNode: string = "";
  let showAIDiagnose: boolean = false;
  let hasAI: boolean = false;
  let showEditLine: boolean = false;
  let selectedLineNode1: string = "";
  let selectedLineNode2: string = "";
  let selectedLineID: string = "";
  let showEditDrawItem: boolean = false;
  let selectedDrawItem: string = "";
  let showEditNetwork: boolean = false;
  let selectedNetwork: string = "";
  let networkTemplate: any = undefined;
  let showNetworkLines: boolean = false;
  let showNeighborNetworksAndLines: boolean = false;
  let showDiscover: boolean = false;
  let showGrid: boolean = false;
  let gridSize: number = 40;
  let showNodeReport: boolean = false;
  let showNetworkReport: boolean = false;
  let showPolling: boolean = false;
  let showPing: boolean = false;
  let showMibBr: boolean = false;
  let editDrawItems: boolean = false;
  let showNodeInfo: boolean = false;

  let showGNMITool: boolean = false;
  let showExportModal = false;
  let exporting = false;

  let timer: any = undefined;
  let urls: any = [];
  let refreshCount = 0;

  const checkAI = async () => {
    const conf = await GetMapConf();
    hasAI = !!(conf && conf.LLMProvider && conf.LLMProvider !== "none");
  };

  onMount(async () => {
    await checkAI();
    refreshCount = 0;
    await initMAP(map, callBack);
    refreshMap();
  });

  onDestroy(() => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    deleteMap();
  });

  const showNodeMenuFunc = async (id: string) => {
    await checkAI();
    selectedNode = id;
    urls = [];
    const n = await GetNode(id);
    urls = n.URL.split(",");
    showNodeMenu = true;
  };

  let selectedNodes :any = [];
  let mapPosX = 0;
  let mapPosY = 0;

  const callBack = (p: any) => {
    refreshCount = 0;
    switch (p.Cmd) {
      case "contextMenu":
        if (map) {
          const bcr = map.getBoundingClientRect();
          const menuWidth = 190;
          let menuHeight = 360;
          if (p.Node) {
            menuHeight = 400;
          } else if (p.DrawItem) {
            menuHeight = 110;
          } else if (p.Network) {
            menuHeight = 270;
          }
          const relX = p.x - bcr.left;
          const relY = p.y - bcr.top;
          posX = relX + menuWidth > bcr.width ? Math.max(10, relX - menuWidth) : relX;
          posY = relY + menuHeight > bcr.height ? Math.max(10, relY - menuHeight) : relY;
        } else {
          posX = p.x;
          posY = p.y;
        }
        if (p.Node) {
          showNodeMenuFunc(p.Node);
        } else if (p.DrawItem) {
          showDrawItemMenu = true;
          selectedDrawItem = p.DrawItem;
        } else if (p.Network) {
          selectedNetwork =  p.Network;
          networkTemplate = undefined;
          showNetworkMenu = true;
        } else {
          showMapMenu = true;
          const bcr = map ? map.getBoundingClientRect() : { left: 0, top: 0 };
          mapPosX = Math.trunc(p.x - bcr.left + (map ? map.scrollLeft : 0));
          mapPosY = Math.trunc(p.y - bcr.top + (map ? map.scrollTop : 0));
        }
        break;
      case "editLine":
        if (p.Param) {
          selectedLineNode1 = p.Param[0];
          selectedLineNode2 = p.Param[1];
          selectedLineID = "";
          showEditLine = true;
          setMapReadOnly(true);
        }
        break;
      case "nodeDoubleClicked":
        selectedNode = p.Param;
        showNodeReport = true;
        break;
      case "itemDoubleClicked":
        selectedDrawItem = p.Param;
        showEditDrawItem = true;
        setMapReadOnly(true);
        break;
      case "networkDoubleClicked":
        selectedNetwork = p.Param;
        showEditNetwork = true;
        setMapReadOnly(true);
        break;
      case "deleteNodes":
        deleteNodes(p.Param);
        break;
      case "formatNodes":
        if (map) {
          const bcr = map.getBoundingClientRect();
          const menuWidth = 180;
          const menuHeight = 140;
          const relX = p.x - bcr.left;
          const relY = p.y - bcr.top;
          posX = relX + menuWidth > bcr.width ? Math.max(10, relX - menuWidth) : relX;
          posY = relY + menuHeight > bcr.height ? Math.max(10, relY - menuHeight) : relY;
        } else {
          posX = p.x;
          posY = p.y;
        }
        selectedNodes = [];
        for(const id of p.Nodes) {
          selectedNodes.push(id);
        }
        showFormatNodesMenu = true;
        break;
      case "deleteDrawItems":
        deleteDrawItems(p.Param);
        break;
      case "deleteNetwork":
        deleteNetwork(p.Param);
        break;
    }
  };

  const refreshMap = async () => {
    if (timer) {
      clearTimeout(timer);
      timer = undefined;
    }
    refreshCount++;
    if(refreshCount > 6+59) {
      WindowReloadApp();
      refreshCount = 0;
      return;
    }
    let t = 10;
    if (refreshCount > 6) {
      t = 60;
    } else if (refreshCount == 1) {
      t = 2;
    }
    updateMAP();
    timer = setTimeout(refreshMap, 1000 * t);
  };

  const deleteNodes = async (ids: string[]) => {
    await DeleteNodes(ids);
    showNodeMenu = false;
    refreshMap();
  };

  const deleteDrawItems = async (ids: string[]) => {
    await DeleteDrawItems(ids);
    showDrawItemMenu = false;
    refreshMap();
  };

  const deleteNetwork = async (id:string) => {
    await DeleteNetwork(id);
    showNetworkMenu = false;
    refreshMap();
  };

  let showEditBackImage = false;
  let backImage: datastore.BackImageEnt;
  let image: any = undefined;
  const showEditBackImageDlg = async () => {
    backImage = await GetBackImage();
    if (backImage.Path) {
      image = await GetImage(backImage.Path);
    }
    if (backImage.Height < 1) {
      backImage.Height = 100;
    }
    if (backImage.Width < 1) {
      backImage.Width = 100;
    }
    showMapMenu = false;
    showEditBackImage = true;
  };

  const selectImage = async () => {
    const p = await SelectFile($_("Map.BackImage"), true);
    if (p) {
      backImage.Path = p;
      image = await GetImage(p);
    }
  };

  const saveBackImage = async () => {
    showEditBackImage = false;
    backImage.Height *=1;
    backImage.Width *=1;
    backImage.X *=1;
    backImage.Y *=1;
    await SetBackImage(backImage);
    refreshMap();
  };

  const clearBackImage = async () => {
    showEditBackImage = false;
    backImage.Path = "";
    image = undefined;
    backImage.Height *=1;
    backImage.Width *=1;
    backImage.X *=1;
    backImage.Y *=1;
    await SetBackImage(backImage);
    refreshMap();
  };
  const saveMap = () => {
    showExportModal = true;
  };
  const handleExport = async (format: string) => {
    showExportModal = false;
    exporting = true;
    try {
      let pngBase64 = "";
      if (format === "png" || format === "pdf" || format === "excel") {
        const canvas = document.getElementById("defaultCanvas0") as HTMLCanvasElement | undefined;
        if (canvas) {
          pngBase64 = canvas.toDataURL("image/png");
        }
      }
      await ExportMap(format, pngBase64);
    } catch (err) {
      console.error(err);
    } finally {
      exporting = false;
    }
  };
</script>

<div bind:this={map} class="h-full w-full overflow-scroll"></div>

<Button
  color="alternative"
  class="!p-2 absolute end-6 bottom-16"
  onclick={saveMap}
>
  <Icon path={icons.mdiContentSave}></Icon>
</Button>

<Button
  color="alternative"
  class="!p-2 absolute end-20 bottom-6"
  onclick={() => zoom(true)}
>
  <Icon path={icons.mdiMagnifyPlus}></Icon>
</Button>

<Button
  color="alternative"
  class="!p-2 absolute end-6 bottom-6"
  onclick={() => zoom(false)}
>
  <Icon path={icons.mdiMagnifyMinus}></Icon>
</Button>

{#if showMapMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block z-50" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-slate-800/95 text-slate-200 border border-slate-700/80 shadow-2xl backdrop-blur-md rounded-xl p-1.5 min-w-[180px] space-y-0.5 select-none text-xs"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          selectedNode = "";
          showEditNode = true;
          showMapMenu = false;
          setMapReadOnly(true);
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiPlus} size={0.7} /></span>
        <div>
          {$_("Map.AddNode")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          selectedDrawItem = "";
          showEditDrawItem = true;
          showMapMenu = false;
          setMapReadOnly(true);
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiDrawing} size={0.7} /></span>
        <div>
          {$_("Map.AddDrawItem")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          selectedNetwork = "";
          showEditNetwork = true;
          showMapMenu = false;
          setMapReadOnly(true);
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiDrawing} size={0.7} /></span>
        <div>
          {$_('Map.AddNetwork')}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showMapMenu = false;
          CheckPolling("all");
          refreshMap();
        }}
      >
        <span class="text-emerald-400"><Icon path={icons.mdiCached} size={0.7} /></span>
        <div>
          {$_("Map.CheckAll")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showMapMenu = false;
          showDiscover = true;
        }}
      >
        <span class="text-indigo-400"><Icon path={icons.mdiFileFind} size={0.7} /></span>
        <div>
          {$_("Map.Discover")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={async () => {
          showMapMenu = false;
          if (await ImportV4Map()) {
            refreshMap();
          }
        }}
      >
        <span class="text-amber-400"><Icon path={icons.mdiImport} size={0.7} /></span>
        <div>
          {$_('Map.Import')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showMapMenu = false;
          showGrid = true;
        }}
      >
        <span class="text-slate-400"><Icon path={icons.mdiGrid} size={0.7} /></span>
        <div>
          {$_("Map.Grid")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={showEditBackImageDlg}
      >
        <span class="text-slate-400"><Icon path={icons.mdiImage} size={0.7} /></span>
        <div>
          {$_("Map.BackImage")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showMapMenu = false;
          resetMap();
          refreshMap();
        }}
      >
        <span class="text-slate-400"><Icon path={icons.mdiRecycle} size={0.7} /></span>
        <div>
          {$_("Map.Reload")}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          editDrawItems = !editDrawItems;
          setEditDrawItems(editDrawItems);
          showMapMenu = false;
        }}
      >
        {#if editDrawItems}
          <span class="text-purple-400"><Icon path={icons.mdiEye} size={0.7} /></span>
          <div>
            {$_("Map.showDrawItemNomal")}
          </div>
        {:else}
          <span class="text-purple-400"><Icon path={icons.mdiDraw} size={0.7} /></span>
          <div>
            {$_("Map.showDrawItemEdit")}
          </div>
        {/if}
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeInfo = !showNodeInfo;
          setShowNodeInfo(showNodeInfo);
          showMapMenu = false;
        }}
      >
        {#if showNodeInfo}
          <span class="text-purple-400"><Icon path={icons.mdiEye} size={0.7} /></span>
          <div>
           {$_('Map.ShowNodeInfo')}
          </div>
        {:else}
          <span class="text-purple-400"><Icon path={icons.mdiEyeClosed} size={0.7} /></span>
          <div>
          {$_('Map.ShowNodeInfo')}
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}

{#if showNodeMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block z-50" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-slate-800/95 text-slate-200 border border-slate-700/80 shadow-2xl backdrop-blur-md rounded-xl p-1.5 min-w-[180px] space-y-0.5 select-none text-xs"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showNodeReport = true;
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiChartBarStacked} size={0.7} /></span>
        <div>
          {$_("Map.Report")}
        </div>
      </div>
      {#if hasAI}
        <!-- svelte-ignore a11y-no-static-element-interactions -->
        <div
          class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
          onclick={() => {
            showNodeMenu = false;
            showAIDiagnose = true;
          }}
        >
          <span class="text-violet-400"><Icon path={icons.mdiBrain} size={0.7} /></span>
          <div>
            {$_("Map.AIDiagnose")}
          </div>
        </div>
      {/if}
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showPing = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-emerald-400"><Icon path={icons.mdiCheckNetwork} size={0.7} /></span>
        <div>PING</div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showMibBr = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-teal-400"><Icon path={icons.mdiEye} size={0.7} /></span>
        <div>
          {$_("Map.MIBBrowser")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showGNMITool = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-cyan-400"><Icon path={icons.mdiEye} size={0.7} /></span>
        <div>
          {$_('GNMITool.gNMITool')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          WakeOnLan(selectedNode);
        }}
      >
        <span class="text-amber-400"><Icon path={icons.mdiAlarm} size={0.7} /></span>
        <div>Wake On Lan</div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showEditNode = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiPencil} size={0.7} /></span>
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          showPolling = true;
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiLanCheck} size={0.7} /></span>
        <div>
          {$_("Map.Polling")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNodeMenu = false;
          CheckPolling(selectedNode);
          refreshMap();
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiCached} size={0.7} /></span>
        <div>
          {$_("Map.ReCheck")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={async () => {
          showNodeMenu = false;
          await CopyNode(selectedNode);
          refreshMap();
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiContentCopy} size={0.7} /></span>
        <div>
          {$_("Map.Copy")}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 text-rose-400 hover:bg-rose-500/20 hover:text-rose-300"
        onclick={() => {
          deleteNodes([selectedNode]);
          refreshMap();
        }}
      >
        <span class="text-rose-400"><Icon path={icons.mdiDelete} size={0.7} /></span>
        <div>
          {$_("Map.Delete")}
        </div>
      </div>

      {#if urls.some((u: string) => u)}
        <div class="border-t border-slate-700/60 my-1"></div>
        {#each urls as url}
          {#if url}
            <!-- svelte-ignore a11y-no-static-element-interactions -->
            <div
              class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300 overflow-hidden"
              onclick={() => {
                showNodeMenu = false;
                BrowserOpenURL(url);
              }}
            >
              <div class="flex-none text-sky-400">
                <Icon path={icons.mdiLink} size={0.7} />
              </div>
              <div class="truncate">
                {url}
              </div>
            </div>
          {/if}
        {/each}
      {/if}
    </div>
  </div>
{/if}

{#if showDrawItemMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block z-50" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-slate-800/95 text-slate-200 border border-slate-700/80 shadow-2xl backdrop-blur-md rounded-xl p-1.5 min-w-[180px] space-y-0.5 select-none text-xs"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showDrawItemMenu = false;
          showEditDrawItem = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiPencil} size={0.7} /></span>
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={async () => {
          showDrawItemMenu = false;
          await CopyDrawItem(selectedDrawItem);
          refreshMap();
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiContentCopy} size={0.7} /></span>
        <div>
          {$_("Map.Copy")}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 text-rose-400 hover:bg-rose-500/20 hover:text-rose-300"
        onclick={() => {
          deleteDrawItems([selectedDrawItem]);
          refreshMap();
        }}
      >
        <span class="text-rose-400"><Icon path={icons.mdiDelete} size={0.7} /></span>
        <div>
          {$_("Map.Delete")}
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showNetworkMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block z-50" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-slate-800/95 text-slate-200 border border-slate-700/80 shadow-2xl backdrop-blur-md rounded-xl p-1.5 min-w-[180px] space-y-0.5 select-none text-xs"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          showNetworkReport = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiChartBarStacked} size={0.7} /></span>
        <div>
          {$_("Map.Report")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          CheckNetwork(selectedNetwork);
          refreshMap();
        }}
      >
        <span class="text-emerald-400"><Icon path={icons.mdiCached} size={0.7} /></span>
        <div>
          {$_("Map.ReCheck")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          selectedNode = "NET:" + selectedNetwork;
          showPing = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-emerald-400"><Icon path={icons.mdiCheckNetwork} size={0.7} /></span>
        <div>PING</div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          selectedNode = "NET:" + selectedNetwork;
          showMibBr = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-teal-400"><Icon path={icons.mdiEye} size={0.7} /></span>
        <div>
          {$_("Map.MIBBrowser")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          showNeighborNetworksAndLines = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-indigo-400"><Icon path={icons.mdiLanConnect} size={0.7} /></span>
        <div>
          {$_('Map.FindNeighbor')}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          showEditNetwork = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiPencil} size={0.7} /></span>
        <div>
          {$_("Map.Edit")}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showNetworkMenu = false;
          showNetworkLines = true;
          setMapReadOnly(true);
        }}
      >
        <span class="text-blue-400"><Icon path={icons.mdiPlaylistEdit} size={0.7} /></span>
        <div>
          {$_('Map.EditLine')}
        </div>
      </div>

      <div class="border-t border-slate-700/60 my-1"></div>

      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 text-rose-400 hover:bg-rose-500/20 hover:text-rose-300"
        onclick={() => {
          deleteNetwork(selectedNetwork);
          refreshMap();
        }}
      >
        <span class="text-rose-400"><Icon path={icons.mdiDelete} size={0.7} /></span>
        <div>
          {$_("Map.Delete")}
        </div>
      </div>
    </div>
  </div>
{/if}

{#if showFormatNodesMenu}
  <!-- svelte-ignore a11y-click-events-have-key-events -->
  <div class="block z-50" style="position: absolute; left:{posX}px;top: {posY}px">
    <div
      class="bg-slate-800/95 text-slate-200 border border-slate-700/80 shadow-2xl backdrop-blur-md rounded-xl p-1.5 min-w-[180px] space-y-0.5 select-none text-xs"
    >
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showFormatNodesMenu = false;
          horizontal(selectedNodes);
          selectedNodes = [];
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiFormatVerticalAlignCenter} size={0.7} /></span>
        <div>
          {$_('Map.Horizontal')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showFormatNodesMenu = false;
          vertical(selectedNodes);
          selectedNodes = [];
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiFormatHorizontalAlignCenter} size={0.7} /></span>
        <div>
          {$_('Map.Vertical')}
        </div>
      </div>
      <!-- svelte-ignore a11y-no-static-element-interactions -->
      <div
        class="flex items-center space-x-2.5 px-2.5 py-1.5 rounded-lg cursor-pointer transition-colors duration-150 hover:bg-slate-700/80 hover:text-white text-slate-300"
        onclick={() => {
          showFormatNodesMenu = false;
          circle(selectedNodes);
          selectedNodes = [];
        }}
      >
        <span class="text-sky-400"><Icon path={icons.mdiCircleOutline} size={0.7} /></span>
        <div>
          {$_('Map.Circle')}
        </div>
      </div>
    </div>
  </div>
{/if}

<Discover
  bind:show={showDiscover}
  posX={mapPosX}
  posY={mapPosY}
  on:close={() => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<Node
  bind:show={showEditNode}
  nodeID={selectedNode}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<Line
  bind:show={showEditLine}
  nodeID1={selectedLineNode1}
  nodeID2={selectedLineNode2}
  id={selectedLineID}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<DrawItem
  bind:show={showEditDrawItem}
  id={selectedDrawItem}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<Network
  bind:show={showEditNetwork}
  id={selectedNetwork}
  template={networkTemplate}
  posX={mapPosX}
  posY={mapPosY}
  on:close={(e) => {
    networkTemplate = undefined;
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<NetworkReport
  bind:show={showNetworkReport}
  id={selectedNetwork}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<NetworkLines
  bind:show={showNetworkLines}
  id={selectedNetwork}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
  on:editLine={(e) => {
    selectedLineID = e.detail;
    selectedLineNode1= "";
    selectedLineNode2= "";
    showEditLine = true;
    setMapReadOnly(true);
  }}
/>

<NeighborNetworksAndLines
  bind:show={showNeighborNetworksAndLines}
  id={selectedNetwork}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
  on:addNetwork={(e) => {
    networkTemplate = e.detail;
    showEditNetwork = true;
    setMapReadOnly(true);
  }}
/>

<NodeReport bind:show={showNodeReport} id={selectedNode}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<NodeDiagnoseDialog bind:show={showAIDiagnose} nodeID={selectedNode} />

<NodePolling
  bind:show={showPolling}
  nodeID={selectedNode}
  on:close={(e) => {
    setMapReadOnly(false);
    refreshMap();
  }}
/>

<Ping 
  bind:show={showPing} 
  nodeID={selectedNode}
  on:close={(e) => {
    setMapReadOnly(false);
  }}
/>

<MIBBrowser 
  bind:show={showMibBr}
  nodeID={selectedNode}
  on:close={(e) => {
    setMapReadOnly(false);
  }}
/>

<GNMITool
  bind:show={showGNMITool}
  nodeID={selectedNode}
  on:close={(e) => {
    setMapReadOnly(false);
  }}
/>

<Modal bind:open={showGrid} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Map.Grid")}
    </h3>
    <Label class="space-y-2 text-xs">
      <span>{$_("Map.GridSize")} </span>
      <Input
        class="h-8 w-24 text-right"
        type="number"
        min={20}
        max={120}
        step={1}
        bind:value={gridSize}
        size="sm"
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="red"
        type="button"
        onclick={() => {
          showGrid = false;
          grid(gridSize, false);
        }}
        size="xs"
      >
        <Icon path={icons.mdiRun} size={1} />
        {$_("Map.Exec")}
      </GradientButton>
      <GradientButton
        shadow
        color="lime"
        type="button"
        onclick={() => {
          showGrid = false;
          grid(gridSize, true);
        }}
        size="xs"
      >
        <Icon path={icons.mdiTestTube} size={1} />
        {$_("Map.Test")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        onclick={() => {
          showGrid = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal
  bind:open={showEditBackImage}
  size="sm"
  dismissable={false}
  class="w-full"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Map.BackImage")}
    </h3>
    <div class="grid gap-4 mb-4 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>X</span>
        <Input
          class="h-8 w-24 text-right"
          type="number"
          min={0}
          max={2000}
          bind:value={backImage.X}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>Y</span>
        <Input
          class="h-8 w-24 text-right"
          type="number"
          min={0}
          max={2000}
          bind:value={backImage.Y}
          size="sm"
        />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>{$_("DrawItem.Width")}</span>
        <Input
        class="h-8 w-24 text-right"
        type="number"
          min={10}
          max={1000}
          bind:value={backImage.Width}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("DrawItem.Height")}</span>
        <Input
          class="h-8 w-24 text-right"
          type="number"
          min={10}
          max={1000}
          bind:value={backImage.Height}
          size="sm"
        />
      </Label>
    </div>
    <GradientButton
      shadow
      class="h-8 mt-6 w-20"
      type="button"
      size="xs"
      color="blue"
      onclick={selectImage}
    >
      <Icon path={icons.mdiImage} size={1} />
      {$_("DrawItem.Select")}
    </GradientButton>
    <Label class="space-y-2 text-xs">
      <span>{$_("DrawItem.Image")}</span>
      {#if image}
        <img src={image} alt="" class="h-32" />
      {:else}
        <div></div>
      {/if}
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        color="blue"
        type="button"
        onclick={saveBackImage}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_("Map.Save")}
      </GradientButton>
      {#if backImage.Path}
        <GradientButton
          color="red"
          type="button"
          onclick={clearBackImage}
          size="xs"
        >
          <Icon path={icons.mdiDelete} size={1} />
          {$_("Map.Clear")}
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        color="teal"
        type="button"
        onclick={() => {
          showEditBackImage = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal bind:open={showExportModal} size="md" dismissable={false} class="w-full">
  <div class="p-6">
    <h3 class="text-lg font-bold text-sky-400 mb-2">{$_("Map.ExportTitle")}</h3>
    <p class="text-slate-300 text-sm mb-6">{$_("Map.ExportDesc")}</p>
    
    <div class="grid grid-cols-2 gap-3 mb-6">
      <button onclick={() => handleExport('png')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportPng")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportPngDesc")}</span>
      </button>
      <button onclick={() => handleExport('svg')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportSvg")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportSvgDesc")}</span>
      </button>
      <button onclick={() => handleExport('pdf')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportPdf")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportPdfDesc")}</span>
      </button>
      <button onclick={() => handleExport('drawio')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportDrawio")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportDrawioDesc")}</span>
      </button>
      <button onclick={() => handleExport('json_map')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportJsonMap")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportJsonMapDesc")}</span>
      </button>
      <button onclick={() => handleExport('csv')} class="flex flex-col items-start p-3 bg-slate-700 hover:bg-slate-650 rounded-xl transition duration-150 text-left border border-slate-600 w-full">
        <span class="text-sm font-semibold text-slate-100">{$_("Map.ExportCsv")}</span>
        <span class="text-xs text-slate-400">{$_("Map.ExportCsvDesc")}</span>
      </button>
      <button onclick={() => handleExport('excel')} class="flex flex-col items-start p-3 bg-indigo-950/30 hover:bg-indigo-900/50 border border-indigo-800/50 rounded-xl transition duration-150 text-left col-span-2 w-full">
        <span class="text-sm font-semibold text-indigo-200">{$_("Map.ExportExcel")}</span>
        <span class="text-xs text-indigo-400">{$_("Map.ExportExcelDesc")}</span>
      </button>
    </div>
    
    <div class="flex justify-end">
      <GradientButton
        shadow
        color="teal"
        type="button"
        onclick={() => {
          showExportModal = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Map.Cancel")}
      </GradientButton>
    </div>
  </div>
</Modal>

<svelte:window
  onclick={() => {
    showMapMenu = false;
    showNetworkMenu = false;
    showNodeMenu = false;
    showDrawItemMenu = false;
    showFormatNodesMenu = false;
    refreshCount = 0;
  }}
/>
