<script lang="ts">
  import Grid from "gridjs-svelte";
  import {h, html} from "gridjs";
  import { onMount} from "svelte";
  import jaJP from "./gridjsJaJP";
  import { GetNodes,DeleteNodes } from "../../wailsjs/go/main/App";
  import { cmpIP,cmpState, getIcon, getStateColor, getStateName } from "./common";
  import Node from "./Node.svelte";

  let data = [];
  let showEditNode = false;
  let selectedNode = "";

  const refreshNodes = async () => {
    data = [];
    const nodes = await GetNodes();
    for(const k in nodes) {
      data.push(nodes[k]);
    }
  };

  const formatState = (state,row) => {
    return html(`<span class="mdi `+ getIcon(row._cells[1].data) + ` text-xl" style="color:`+ getStateColor(state) 
    + `;" /><span class="ml-2 text-xs text-black dark:text-white">`+getStateName(state) +`</span>`);
  }

  const editNode = (id:string) => {
    console.log("editNode",id);
    if(!id) {
      return;
    }
    selectedNode = id;
    showEditNode = true;
  }

  const deleteNode = async (id:string) => {
    console.log("deleteNode",id);
    await DeleteNodes([id]);
    refreshNodes();
  }

  const columns = [
    {
      id: "State",
      name: "状態",
      width: "10%",
      formatter: formatState,
      sort: {
        compare: cmpState,
      },
    },
    {
      id: "Icon",
      name: "",
      hidden: true,
    },
    {
      id:"Name",
      name:"名前",
      width: "20%",
    },
    {
      id:"IP",
      name:"IPアドレス",
      width: "15%",
      sort: {
        compare: cmpIP,
      },
    },
    {
      id:"MAC",
      name:"MACアドレス",
      width: "15%",
    },
    {
      id:"Descr",
      name:"説明",
      width: "30%",
    },
    {
      id: "ID",
      name:"編集",
      sort: false,
      width: "5%",
      formatter: (id) => {
        return h("button",{
        className: "",
        onClick: () => {editNode(id)},
        },html(`<span class="mdi mdi-pencil text-lg" />`));
      },
    },
    {
      name:"削除",
      width: "5%",
      formatter: (_,row) => {
        const id = row._cells[row._cells.length -2].data;
        return h("button",{
        className: "",
        onClick: () => {deleteNode(id)},
        },html(`<span class="mdi mdi-delete text-red-600 text-lg" />`));
      },
    },
  ];  
  const pagination = {
    limit: 20,
  };

  onMount(()=>{
    refreshNodes();
  });

</script>

<div class="m-5 twsnmpfk">
  <Grid {data} {columns} {pagination} sort search language={jaJP}/>
</div>

{#if showEditNode}
  <Node
    nodeID={selectedNode}
    on:close={(e) => {
      showEditNode = false;
      refreshNodes();
    }}
  />
{/if}


<style>
  @import "../assets/css/gridjs.css";
</style>
