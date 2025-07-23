<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import { GradientButton,Modal,Label,Input } from "flowbite-svelte";
  import {Icon} from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetCertMonitorList,
    DeleteCertMonitor,
    UpateCertMonitor,
  } from "../../wailsjs/go/main/App";
  import { renderTime, renderState, getTableLang,renderTimeUnix } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from 'svelte-i18n';

  let table :any = undefined;
  let data = [];
  let selectedCount = 0;
  let certMonitor: any = undefined;
  let showEditDialog = false;

  const columns = [
    {
      className: 'dt-control',
      orderable: false,
      data: null,
      defaultContent: '',
      width:'5%'
    },
    {
      data: "State",
      title: $_('CertMonitor.State'),
      width: "5%",
      render: renderState,
    },
    {
      data: "Target",
      title: $_('CertMonitor.Target'),
      width: "10%",
    },
    {
      data: "Port",
      title: $_('CertMonitor.Port'),
      width: "5%",
    },
    {
      data: "Subject",
      title: "Subject",
      width: "20%",
    },
    {
      data: "Issuer",
      title: "Issuer",
      width: "25%",
    },
    {
      data: "NotBefore",
      title: $_('CertMonitor.NotBefore'),
      width: "10%",
      render: renderTimeUnix,
    },
    {
      data: "NotAfter",
      title: $_('CertMonitor.NotAfter'),
      width: "10%",
      render: renderTimeUnix,
    },
    {
      data: "LastTime",
      title: $_('AIList.LastTime'),
      width: "10%",
      render: renderTime,
    },
  ];

  const formatCert = (d:any) => {
    return (
        '<dl>' +
        '<dt>State:</dt>' +
        '<dd>' +
          d.State +
        '</dd>' +
        '<dt>Target:Port:</dt>' +
        '<dd>' +
          d.Target + ':' + d.Port +
        '</dd>' +
        '<dt>Subject:</dt>' +
        '<dd>' +
          d.Subject +
        '</dd>' +
        '<dt>Issuer:</dt>' +
        '<dd>' +
          d.Issuer +
        '</dd>' +
        '<dt>Serial Number:</dt>' +
        '<dd>' +
          d.SerialNumber +
        '</dd>' +
        '<dt>Verify:</dt>' +
        '<dd>' +
          d.Verify +
        '</dd>' +
        '<dt>Error:</dt>' +
        '<dd>' +
          d.Error +
        '</dd>' +
        '<dt>Term:</dt>' +
        '<dd>' +
          renderTimeUnix(d.NotBefore,"") + ' - ' + renderTimeUnix(d.NotAfter,"") +
        '</dd>' +
        '<dt>Days:</dt>' +
        '<dd>' +
          ((d.NotAfter - d.NotBefore) /(24 * 3600)).toFixed(0) + '' + 
        '</dd>' +
        '</dl>'
    );   
  }

  const refresh = async () => {
    data = await GetCertMonitorList();
    selectedCount = 0;
    if (table) {
      table.clear().rows.add(data).draw();
      return
    }
    table = new DataTable("#certMonitorListTable", {
      destroy: true,
      columns: columns,
      pageLength: window.innerHeight > 800 ? 25 : 10,
      stateSave: true,
      data: data,
      order:[[0,"desc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
    });
    table.on('click', 'tbody td.dt-control', function (e:any) {
      let tr = e.target.closest('tr');
      let row = table.row(tr);
      if (row.child.isShown()) {
        row.child.hide();
      } else {
        row.child(formatCert(row.data())).show();
      }
    });
  };

  onMount(() => {
    refresh();
  });

  const add = () => {
    certMonitor = {
      ID: "",
      Target: "",
      Port: 443,
    }
    showEditDialog = true;
  };

  const edit = () => {
    const selected = table.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    certMonitor = {
      ID: selected[0].ID,
      Target: selected[0].Target,
      Port: selected[0].Port,
    }
    showEditDialog = true;
  };

  const del = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length != 1) {
      return;
    }
    await DeleteCertMonitor(selected[0]);
    refresh();
  };

  const update = async () => {
    certMonitor.Port *= 1
    await UpateCertMonitor(certMonitor);
    showEditDialog = false;
    refresh();
  }

</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    <table id="certMonitorListTable" class="display compact" style="width:99%" />
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton shadow color="blue" type="button" on:click={add} size="xs">
      <Icon path={icons.mdiPlus} size={1} />
      {$_('CertMonitor.Add')}
    </GradientButton>
  {#if selectedCount == 1}
    <GradientButton shadow color="blue" type="button" on:click={edit} size="xs">
      <Icon path={icons.mdiPencil} size={1} />
      {$_('CertMonitor.Edit')}
    </GradientButton>
    <GradientButton shadow color="red" type="button" on:click={del} size="xs">
      <Icon path={icons.mdiTrashCan} size={1} />
      {$_('CertMonitor.Del')}
    </GradientButton>
  {/if} 
    <GradientButton shadow type="button" color="teal" on:click={refresh} size="xs">
      <Icon path={icons.mdiRecycle} size={1} />
      { $_('AIList.Reload') }
    </GradientButton>
  </div>
</div>

<Modal
  bind:open={showEditDialog}
  size="sm"
  dismissable={false}
  class="w-full"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_('CertMonitor.EditTitle')}
    </h3>
    <Label class="space-y-2 text-xs">
      <span>{$_('CertMonitor.TargetIPOrHost')}</span>
      <Input class="h-8" bind:value={certMonitor.Target} size="sm" />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>{$_('CertMonitor.PortNumber')}</span>
      <Input
        class="h-8 w-24 text-right"
        type="number"
        min={1}
        max={65535}
        bind:value={certMonitor.Port}
        size="sm"
      />
    </Label>
  </form>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="blue"
      on:click={update}
      size="xs"
    >
      <Icon path={icons.mdiContentSave} size={1} />
      {$_('CertMonitor.Save')}
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={() => {
        showEditDialog = false;
      }}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      {$_("PKI.Cancel")}
    </GradientButton>
  </div>
</Modal>

<style global>
  #certMonitorListTable  dt {
    color: #fff;
    font-size: 14px;
  }
  #certMonitorListTable  dd {
    color: #eee;
    margin-left: 10px;
    font-size: 12px;
  }
</style>