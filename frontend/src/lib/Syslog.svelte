<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Select,
    Spinner,
    Button,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount, tick} from "svelte";
  import {
    GetSyslogs,
    ExportSyslogs,
    GetDefaultPolling,
    AutoGrok,
    DeleteAllSyslog,
    ExportAny,
  } from "../../wailsjs/go/main/App";
  import { renderState, renderTime, getTableLang } from "./common";
  import { showLogLevelChart, resizeLogLevelChart } from "./chart/loglevel";
  import { 
    showLogCountChart,
    resizeLogCountChart,
    showMagicTimeChart,
    showMagicHourChart,
    showMagicSumChart,
    showMagicGraphChart,
  } from "./chart/logcount";
  import SyslogReport from "./SyslogReport.svelte";
  import Polling from "./Polling.svelte";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import type { datastore, main } from "wailsjs/go/models";
  import { _ } from "svelte-i18n";
  import { CodeJar } from "@novacbn/svelte-codejar";
  import Prism from "prismjs";
  import "prismjs/components/prism-regex";
  import { copyText } from "svelte-copy";

  let data: any = [];
  let logs: any = [];
  let showReport = false;
  let table: any = undefined;
  let selectedCount = 0;
  let showPolling = false;
  let showFilter = false;
  let showLoading = false;

  const filter: main.SyslogFilterEnt = {
    Start: "",
    End: "",
    Severity: 6,
    Host: "",
    Tag: "",
    Message: "",
  };

  const levelList = [
    { name: $_("Syslog.All"), value: 7 },
    { name: $_("Syslog.Info"), value: 6 },
    { name: $_("Syslog.Warn"), value: 4 },
    { name: $_("Syslog.Low"), value: 3 },
    { name: $_("Syslog.High"), value: 2 },
  ];

  const showTable = () => {
    if (table && DataTable.isDataTable("#syslogTable")) {
      table.clear();
      table.destroy();
      table = undefined;
    }
    selectedCount = 0;
    table = new DataTable("#syslogTable", {
      columns: columns,
      data: data,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      stateSave: true,
      order:[[1,"desc"]],
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
  };

  const refresh = async () => {
    filter.Severity *= 1;
    showLoading = true;
    logs = await GetSyslogs(filter);
    data = [];
    for (let i = 0; i < logs.length; i++) {
      data.push(logs[i]);
    }
    logs.reverse();
    showTable();
    showChart();
    showLoading = false;
  };

  let chart : any = undefined;
  const showChart = async () => {
    await tick();
   chart = showLogLevelChart("chart", logs, zoomCallBack);
  };

  const zoomCallBack = (st: number, et: number) => {
    data = [];
    for (let i = logs.length - 1; i >= 0; i--) {
      if (logs[i].Time >= st && logs[i].Time <= et) {
        data.push(logs[i]);
      }
    }
    showTable();
  };

  const columns = [
    {
      data: "Level",
      title: $_("Syslog.Level"),
      width: "10%",
      render: renderState,
    },
    {
      data: "Time",
      title: $_("Syslog.Time"),
      width: "15%",
      render: renderTime,
    },
    {
      data: "Host",
      title: $_("Syslog.Host"),
      width: "15%",
    },
    {
      data: "Type",
      title: $_("Syslog.Type"),
      width: "10%",
    },
    {
      data: "Tag",
      title: $_("Syslog.Tag"),
      width: "10%",
    },
    {
      data: "Message",
      title: $_("Syslog.Message"),
      width: "40%",
    },
  ];

  onMount(() => {
    refresh();
  });

  const saveCSV = () => {
    ExportSyslogs("csv", filter,"");
  };

  const saveExcel = () => {
    ExportSyslogs("excel", filter,chart ? chart.getDataURL() : "");
  };

  let copied = false;
  const copy = () => {
    const selected = table.rows({ selected: true }).data();
    let s :string[] = [];
    const h = columns.map((e:any)=> e.title);
    s.push(h.join("\t"))
    for(let i = 0 ;i < selected.length;i++ ) {
      const row :any = [];
      for (const c of columns) {
        if (c.data == "Time") {
          row.push(renderTime(selected[i][c.data] || "",""));
        } else {
          row.push(selected[i][c.data] || "");
        }
      }
      s.push(row.join("\t"))
    }
    copyText(s.join("\n"))
    copied = true;
    setTimeout(()=> copied = false,2000);
  };

  let polling: datastore.PollingEnt | undefined = undefined;

  const watch = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
      return;
    }
    polling = await GetDefaultPolling(d[0].Host);
    polling.Extractor = await AutoGrok(d[0].Message);
    if (polling.Extractor == "") {
      polling.Mode = "count";
      polling.Script = "count < 1";
    }
    polling.Name = `syslog`;
    polling.Type = "syslog";
    polling.Filter = d[0].Type + " " + d[0].Tag;
    polling.Params = d[0].Host;
    showPolling = true;
  };

  const deleteAll = async () => {
    if (await DeleteAllSyslog()) {
      refresh();
    }
  };

  const highlight = (code: string, syntax: string | undefined) => {
    if (!syntax) {
      return "";
    }
    return Prism.highlight(code, Prism.languages[syntax], syntax);
  };

  let showMagic = false;
  let magicData: any;
  let magicDataOrg: any;
  let magicCopied = false;
  let magicSelectedCount = 0;
  let magicTable: any;

  const magicColumnsDefault = [
    {
      data: "Time",
      title: $_("Syslog.Time"),
      render: renderTime,
    },
    {
      data: "Host",
      title: $_("Syslog.Host"),
    },
    {
      data: "Type",
      title: $_("Syslog.Type"),
    },
    {
      data: "Tag",
      title: $_("Syslog.Tag"),
    },
  ];

  let magicColumns :any = [];

  const regCut = /[\{\}"'[;,\]\[]+?/g;
  const regSplunk = /^([a-zA-Z0-9-+]+)=(\S+)$/;
  const regJson = /^\s*\{.*\}\s*$/;
  const regs = [
    {
      type: "timeHHMMSS",
      pattern: `[0-9]{2}:[0-9]{2}:[0-9]{2}`,
      regex: undefined as any,
      count: 0,
    },
    {
      type: "timeHHMM",
      pattern: `[0-9]{2}:[0-9]{2}`,
      regex: undefined,
      count: 0,
    },
    {
      type: "ipv4",
      pattern: `[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}\\.[0-9]{1,3}`,
      regex: undefined,
      count: 0,
    },
    {
      type: "mac",
      pattern: `[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}:[0-9a-fA-F]{2}`,
      regex: undefined,
      count: 0,
    },
    {
      type: "url",
      pattern: `https?:\\/\\/[a-z@A-Z_\\.-]+`,
      regex: undefined,
      count: 0,
    },
    {
      type: "mail",
      pattern: `[a-zA-Z]+[0-9a-zA-Z\\._-]+@[a-zA-Z_\\.-]+`,
      regex: undefined,
      count: 0,
    },
    {
      type: "number",
      pattern: `-?[0-9]+[0-9.]*`,
      regex: undefined as any,
      count: 0,
    },
  ];

  const getFilter = (m: string): any => {
    const j = m.match(regJson)
    if(j) {
      try {
        const o = JSON.parse(m);
        if( typeof o == "object") {
          const keys = Object.keys(o);
          if(keys.length > 0){
            keys.forEach((k:string)=> {
              const v = o[k];
              if (typeof v == "number" || v.match(/^-?[0-9]+[0-9.]*$/)) {
                magicChartNumEntList.push({
                  name: k,
                  value: k,
                })
              } else {
                magicChartCatEntList.push({
                  name: k,
                  value: k,
                })
              }
            })
            return {
              headers: keys,
              json: true,
            }
          }
        }
      } catch (e) {
        console.warn(e);
      }
    }
    m = m.replaceAll(regCut, " ");
    m = m.replaceAll("->", " ");
    const a = m.split(/\s+/);
    const r: string[] = [];
    const h: string[] = [];
    magicChartNumEntList = [];
    magicChartCatEntList = [];
    regs.forEach((p) => {
      if (!p.regex) {
        p.regex = new RegExp("^"+p.pattern +"$");
      }
      p.count = 1;
    });
    a.forEach((e: string) => {
      if (!e || e == ":") {
        return;
      }
      const f = e.match(regSplunk);
      if (f) {
        const n = f[1];
        h.push(n);
        r.push(n + "=([^ ,;]+)");
        if (f[2].match(/-?[0-9]+[0-9.]*/)) {
          magicChartNumEntList.push({
            name: n,
            value: n,
          })
        } else {
          magicChartCatEntList.push({
            name: n,
            value: n,
          })
        }
        return;
      }
      for (const p of regs) {
        if (e.match(p.regex)) {
          const n = p.type + "_" + p.count;
          h.push(n);
          if( p.type == "number" ) {
            magicChartNumEntList.push({
              name: n,
              value: n,
            });
          } else if (!p.type.startsWith("time")) {
            magicChartCatEntList.push({
              name: n,
              value: n,
            });
          }
          p.count++;
          r.push("(" + p.pattern + ")+");
          return;
        }
      }
      r.push(e);
    });
    return {
      headers: h,
      paterns: r,
      json:false,
    };
  };

  const showMagicTable = async () => {
    await tick();
    if (magicTable && DataTable.isDataTable("#magicTable")) {
      magicTable.clear();
      magicTable.destroy(true);
      magicTable = undefined;
      const e = document.getElementById("magicTableBase");
      if(e) {
        e.innerHTML = `<table id="magicTable" class="display compact" style="width:99%" />`;
      }
    }
    magicSelectedCount = 0;
    magicTable = new DataTable("#magicTable", {
      columns: magicColumns,
      data: magicData,
      language: getTableLang(),
      select: {
        style: "multi",
      },
    });
    magicTable.on("select", () => {
      magicSelectedCount = magicTable.rows({ selected: true }).count();
    });
    magicTable.on("deselect", () => {
      magicSelectedCount = magicTable.rows({ selected: true }).count();
    });
  };

  let magicChartType = "count";
  let magicChartNumEntList: any = [];
  let magicChartCatEntList: any = [];
  let magicNumEnt = "";
  let magicCatEnt = "";
  let magicCatEnt2 = "";

  let magicChartTypes :any = [];
  
  const showMagicChart = async () => {
    await tick();
    switch(magicChartType) {
    case "count":
        showLogCountChart("magicChart", magicData, magicZoomCallBack);
        break;
    case "time":
        if (!magicNumEnt) {
          return;
        }
        showMagicTimeChart("magicChart", magicData, magicNumEnt);
        break;
    case "hour":
        showMagicHourChart("magicChart", magicData, magicNumEnt);
        break;
    case "sum":
        showMagicSumChart("magicChart", magicData, magicCatEnt);
        break;
    case "graph":
        showMagicGraphChart("magicChart", magicData, magicCatEnt,magicCatEnt2);
        break;
    }
  };

  const magicZoomCallBack = (st: number, et: number) => {
    magicData = [];
    magicDataOrg.forEach((l:any) => {
      if (l.Time >= st && l.Time <= et) {
        magicData.push(l);
      }
    });
    showMagicTable();
  };
  

  const magic = async () => {
    const d = table.rows({ selected: true }).data();
    if (!d || d.length != 1) {
      return;
    }
    const f = getFilter(d[0].Message);
    magicColumns = [];
    magicColumnsDefault.forEach((c)=> {
      magicColumns.push(c);
    });
    f.headers.forEach((h:string)=> {
      magicColumns.push({
        data:h,
        title:h,
      });
    });
    magicDataOrg = [];
    let st = Infinity;
    let et = 0;
    if (f.json) {
      for (let i = 0; i < logs.length; i++) {
        const log = logs[i];
        try {
          const o = JSON.parse(log.Message);
          let hit = true;
          for(const k of f.headers) {
            if (!Object.hasOwn(o,k)) {
              hit = false;
              break
            }
          }
          if (!hit) {
            continue;
          }
          const r:any = {
            Time: log.Time,
            Host: log.Host,
            Type: log.Type,
            Tag:  log.Tag,
          };
          f.headers.forEach((k:string)=> {
            r[k] = o[k];
          });
          st = Math.min(st,log.Time);
          et = Math.max(et,log.Time);
          magicDataOrg.push(r);
        } catch {
         continue; 
        }
      }
    } else {
      const reg = new RegExp(f.paterns.join(`.+?`));
      for (let i = 0; i < logs.length; i++) {
        const log = logs[i];
        const m = log.Message.match(reg);
        if (!m || m.length < f.headers.length + 1) {
          continue;
        }
        const r:any = {
            Time: log.Time,
            Host: log.Host,
            Type: log.Type,
            Tag:  log.Tag,
        };
        f.headers.forEach((e:string,i:number)=> {
          r[e] = m[i+1];
        });
        st = Math.min(st,log.Time);
        et = Math.max(et,log.Time);
        magicDataOrg.push(r);
      }
    }
    magicData = [];
    magicDataOrg.forEach((l:any)=>{
      magicData.push(l);
    });
    showMagic = true;
    magicChartType = "count";
    magicChartTypes =  [
      { name: $_('Syslog.Count'), value: "count" }
    ];
    if (magicChartNumEntList.length > 0 ) {
      magicChartTypes.push({ name: $_('Syslog.TimeChart'), value: "time" });
      if( et-st > 3600 * 1000 * 1000 * 1000) {
        magicChartTypes.push({ name: $_('Syslog.PerHourSum'), value: "hour" });
      } 
    }
    if (magicChartCatEntList.length > 0) {
      magicChartTypes.push({ name: $_('Syslog.MagicBarChart'), value: "sum" });
      if(magicChartCatEntList.length> 1) {
        magicChartTypes.push({ name: $_('SyslogMagicGraph'), value: "graph" });
      }
    }
    showMagicTable();
    showMagicChart();
  };

  const exportMagic = (t: string) => {
    const ed :any = {
      Title: "TWSNMP_Syslog_Magic",
      Header: magicColumns.map((e:any) => e.title),
      Data: [],
      Image: "",
    };
    for (const d of magicData) {
      const row :any = [];
      for (const c of magicColumns) {
        if (c.data == "Time") {
          row.push(renderTime(d[c.data] || "",""));
        } else {
          row.push(d[c.data] || "");
        }
      }
      ed.Data.push(row);
    }
    ExportAny(t, ed);
  };

  const copyMagic = () => {
    const selected = magicTable.rows({ selected: true }).data();
    let s :string[] = [];
    const h = magicColumns.map((e:any)=> e.title);
    s.push(h.join("\t"))
    for(let i = 0 ;i < selected.length;i++ ) {
      const row :any = [];
      for (const c of magicColumns) {
        if (c.data == "Time") {
          row.push(renderTime(selected[i][c.data] || "",""));
        } else {
          row.push(selected[i][c.data] || "");
        }
      }
      s.push(row.join("\t"))
    }
    copyText(s.join("\n"))
    magicCopied = true;
    setTimeout(()=> magicCopied = false,2000);
  };

</script>

<svelte:window on:resize={()=> {
    resizeLogLevelChart();
    resizeLogCountChart();
}} />

<div class="flex flex-col">
  <div id="chart" />
  <div class="m-5 grow">
    <table id="syslogTable" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    {#if selectedCount == 1}
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={watch}
        size="xs"
      >
        <Icon path={icons.mdiEye} size={1} />
        {$_("Syslog.Polling")}
      </GradientButton>
      <GradientButton
        shadow
        color="cyan"
        type="button"
        on:click={magic}
        size="xs"
      >
        <Icon path={icons.mdiMagicStaff} size={1} />
        {$_('Syslog.MagicBtn')}
      </GradientButton>
    {/if}
    {#if selectedCount > 0}
      <GradientButton
      shadow
      color="cyan"
      type="button"
      on:click={copy}
      size="xs"
    >
      {#if copied}
        <Icon path={icons.mdiCheck} size={1} />
      {:else}
        <Icon path={icons.mdiContentCopy} size={1} />
      {/if}
      Copy
    </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="blue"
      type="button"
      on:click={() => (showFilter = true)}
      size="xs"
    >
      <Icon path={icons.mdiFilter} size={1} />
      {$_("Syslog.Filter")}
    </GradientButton>
    <GradientButton
      shadow
      color="red"
      type="button"
      on:click={deleteAll}
      size="xs"
    >
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_("Syslog.DeleteAllLogs")}
    </GradientButton>
    {#if logs.length > 0}
      <GradientButton
        shadow
        type="button"
        color="green"
        on:click={() => {
          showReport = true;
        }}
        size="xs"
      >
        <Icon path={icons.mdiChartPie} size={1} />
        {$_("Syslog.Report")}
      </GradientButton>
    {/if}
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveCSV}
      size="xs"
    >
      <Icon path={icons.mdiFileDelimited} size={1} />
      CSV
    </GradientButton>
    <GradientButton
      shadow
      color="lime"
      type="button"
      on:click={saveExcel}
      size="xs"
    >
      <Icon path={icons.mdiFileExcel} size={1} />
      Excel
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={refresh}
      size="xs"
    >
      <Icon path={icons.mdiRecycle} size={1} />
      {$_("Syslog.Reload")}
    </GradientButton>
  </div>
</div>

<SyslogReport bind:show={showReport} {logs} />

<Polling bind:show={showPolling} pollingTmp={polling} />

<Modal bind:open={showFilter} size="sm" dismissable={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Syslog.Filter")}
    </h3>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.Start")}</span>
        <Input type="datetime-local" bind:value={filter.Start} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("EventLog.End")}</span>
        <Input type="datetime-local" bind:value={filter.End} size="sm" />
      </Label>
      <div class="flex">
        <Button
          class="!p-2 w-8 h-8 mt-6 ml-4"
          color="red"
          on:click={() => {
            filter.Start = "";
            filter.End = "";
          }}
        >
          <Icon path={icons.mdiCancel} size={1} />
        </Button>
      </div>
    </div>
    <div class="grid gap-2 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Level")}</span>
        <Select
          items={levelList}
          bind:value={filter.Severity}
          placeholder={$_("Syslog.SelectLevel")}
          size="sm"
        />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Host")}</span>
        <CodeJar style="padding: 6px;" syntax="regex" {highlight} bind:value={filter.Host} />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>{$_("Syslog.Tag")}</span>
        <CodeJar style="padding: 6px;" syntax="regex" {highlight} bind:value={filter.Tag} />
      </Label>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("Syslog.Message")}</span>
      <CodeJar style="padding: 6px;" syntax="regex" {highlight} bind:value={filter.Message} />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={() => {
          showFilter = false;
          refresh();
        }}
        size="xs"
      >
        <Icon path={icons.mdiSearchWeb} size={1} />
        {$_("Syslog.Search")}
      </GradientButton>
      <GradientButton
        shadow
        color="teal"
        type="button"
        on:click={() => {
          showFilter = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Syslog.Calcel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal bind:open={showLoading} size="sm" dismissable={false} class="w-full">
  <div>
    <Spinner />
    <span class="ml-2"> {$_("Syslog.Loading")} </span>
  </div>
</Modal>

<Modal bind:open={showMagic} size="xl" dismissable={false} class="w-full">
  <div class="flex flex-col space-y-4">
    <div id="magicChart" />
    <div class="m-5 grow" id="magicTableBase">
      <table id="magicTable" class="display compact" style="width:99%" />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      {#if magicData.length > 0}
        <Select
          size="sm"
          items={magicChartTypes}
          bind:value={magicChartType}
          placeholder={$_('Syslog.ChartType')}
          class="w-96"
          on:change={showMagicChart}
        />
        {#if magicChartType == "hour" || magicChartType == "time"}
          <Select
            size="sm"
            items={magicChartNumEntList}
            bind:value={magicNumEnt}
            placeholder={$_('Syslog.NumData')}
            class="w-96"
            on:change={showMagicChart}
          />
        {:else if magicChartType == "sum" || magicChartType == "graph"}
          <Select
            size="sm"
            items={magicChartCatEntList}
            bind:value={magicCatEnt}
            placeholder={$_('Syslog.CatData')}
            class="w-96"
            on:change={showMagicChart}
          />
          {#if magicChartType == "graph"}
            <Select
              size="sm"
              items={magicChartCatEntList}
              bind:value={magicCatEnt2}
              placeholder={$_('Syslog.CatData')}
              class="w-96"
              on:change={showMagicChart}
            />
          {/if}
        {/if}
        {#if magicSelectedCount > 0}
          <GradientButton
            shadow
            color="cyan"
            type="button"
            on:click={copyMagic}
            size="xs"
          >
            {#if magicCopied}
              <Icon path={icons.mdiCheck} size={1} />
            {:else}
              <Icon path={icons.mdiContentCopy} size={1} />
            {/if}
            Copy
          </GradientButton>
        {/if}
        <GradientButton
          shadow
          color="lime"
          type="button"
          on:click={() => {
            exportMagic("csv");
          }}
          size="xs"
        >
          <Icon path={icons.mdiFileDelimited} size={1} />
          CSV
        </GradientButton>
        <GradientButton
          shadow
          color="lime"
          type="button"
          on:click={() => {
            exportMagic("excel");
          }}
          size="xs"
        >
          <Icon path={icons.mdiFileExcel} size={1} />
          Excel
        </GradientButton>
      {/if}
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => (showMagic = false)}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("MIBBrowser.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<style>
  #chart{
    min-height: 200px;
    height: 20vh;
    width: 95vw;
    margin: 0 auto;
  }
  #magicChart {
    min-height: 200px;
    height: 20vh;
    width: 95%;
    margin: 0 auto;
  }
</style>
