<script lang="ts">
  import "../assets/css/jquery.dataTables.css";
  import {
    GradientButton,
    Modal,
    Label,
    Input,
    Select,
    Alert,
    Checkbox,
    Spinner,
  } from "flowbite-svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import { onMount } from "svelte";
  import {
    GetCerts,
    IsCAValid,
    GetDefaultCreateCAReq,
    CreateCA,
    DestroyCA,
    RevokeCert,
    CreateCertificateRequest,
    ExportCert,
    GetPKIControl,
    SetPKIControl,
    CreateCertificate,
  } from "../../wailsjs/go/main/App";
  import { renderTime, getTableLang, keyTypeList } from "./common";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import { _ } from "svelte-i18n";

  let table: any = undefined;
  let data = [];
  let selectedCount = 0;
  let selectedID = "";
  let selectedForRevoke = "";
  let createCAErr = "";
  let createCSRErr = "";
  let createCRTErr = "";
  let showCreateCA = false;
  let showCreateCSRDialog = false;
  let showPKIControlDialog = false;
  let pkiControl: any = {};
  let caReq: any = undefined;
  let wait = false;
  const csrReq: any = {
    KeyType: "rsa-4096",
    CommonName: "",
    OrganizationalUnit: "",
    Organization: "",
    Locality: "",
    Province: "",
    Country: "JP",
    Sans: "",
  };

  const renderStatus = (status: string, type: string) => {
    if (type == "sort") {
      return status;
    }
    switch (status) {
      case "expired":
        return (
          `<span class="mdi mdi-clock-remove text-xs" style="color: #dfdf22;"></span><span class="ml-2">` +
          "期限切れ" +
          `</span>`
        );
      case "revoked":
        return (
          `<span class="mdi mdi-book-remove text-xs" style="color: #e31a1c;"></span><span class="ml-2">` +
          "失効" +
          `</span>`
        );
      default:
        return (
          `<span class="mdi mdi-certificate text-xs" style="color: #1f78b4;"></span><span class="ml-2">` +
          "有効" +
          `</span>`
        );
    }
  };

  const revokeCert = async () => {
    if (selectedForRevoke != "") {
      await RevokeCert(selectedForRevoke);
      setTimeout(refresh, 500);
    }
  };

  const exportCert = async () => {
    const selected = table.rows({ selected: true }).data().pluck("ID");
    if (selected.length == 1) {
      await ExportCert(selected[0]);
    }
  };

  const doCreateCA = async () => {
    wait = true;
    caReq.AcmePort *= 1;
    caReq.HttpPort *= 1;
    createCAErr = await CreateCA(caReq);
    wait = false;
    if (createCAErr == "") {
      refresh();
    }
  };

  const doCreateCSR = async () => {
    createCSRErr = await CreateCertificateRequest(csrReq);
    if (createCSRErr == "") {
      refresh();
    }
    showCreateCSRDialog = false;
  };

  const doCreateCRT = async () => {
    createCRTErr = await CreateCertificate();
    if (createCRTErr == "") {
      refresh();
    }
  };

  const destroyCA = async () => {
    await DestroyCA();
    setTimeout(refresh, 1200);
  };

  const doPKIControl = async () => {
    await SetPKIControl(pkiControl);
    showPKIControlDialog = false;
    setTimeout(refresh, 1200);
  };

  const columns = [
    {
      data: "Status",
      title: "状態",
      width: "10%",
      render: renderStatus,
    },
    {
      data: "Type",
      title: "Type",
      width: "10%",
    },
    {
      data: "ID",
      title: "ID",
      width: "10%",
    },
    {
      data: "Subject",
      title: "Subject",
      width: "25%",
    },
    {
      data: "Node",
      title: "関連ノード",
      width: "15%",
    },
    {
      data: "Created",
      title: "開始",
      width: "10%",
      render: (data: any, type: any) => renderTime(data, type),
    },
    {
      data: "Expire",
      title: "終了",
      width: "10%",
      render: (data: any, type: any) => renderTime(data, type),
    },
    {
      data: "Revoked",
      title: "失効",
      width: "10%",
      render: (data: any, type: any) => renderTime(data, type),
    },
  ];

  const refresh = async () => {
    if (!caReq) {
      caReq = await GetDefaultCreateCAReq();
      console.log(caReq);
    }
    showCreateCA = !(await IsCAValid());
    if (showCreateCA) {
      return;
    }
    pkiControl = await GetPKIControl();
    data = await GetCerts();
    selectedCount = 0;
    table = new DataTable("#certListTable", {
      destroy: true,
      columns: columns,
      pageLength: window.innerHeight > 800 ? 25 : 10,
      stateSave: true,
      data: data,
      order: [[0, "desc"]],
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    table.on("select", () => {
      selectedCount = table.rows({ selected: true }).count();
      const data = table.rows({ selected: true }).data();
      if (
        data.length == 1 &&
        data[0].Type != "system" &&
        data[0].Statue != "revoked"
      ) {
        selectedForRevoke = data[0].ID;
      } else {
        selectedForRevoke = "";
      }
    });
    table.on("deselect", () => {
      selectedCount = table.rows({ selected: true }).count();
      if (selectedCount != 1) {
        selectedForRevoke = "";
        selectedID = "";
      }
    });
  };

  onMount(() => {
    refresh();
  });
</script>

<div class="flex flex-col">
  <div class="m-5 grow">
    {#if createCSRErr}
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={icons.mdiExclamation} size={1} />
          {createCSRErr}
        </div>
      </Alert>
    {/if}
    {#if showCreateCA}
      {#if createCAErr}
        <Alert color="red" dismissable>
          <div class="flex">
            <Icon path={icons.mdiExclamation} size={1} />
            {createCAErr}
          </div>
        </Alert>
      {/if}
      <form class="flex flex-col space-y-4" action="#">
        <h3 class="mb-1 font-medium text-gray-900 dark:text-white">CAの構築</h3>
        {#if wait}
          <Alert color="blue">
            <div class="flex">
              <Spinner size={6}></Spinner>
              <span class="ml-2">
                CAを構築しています...
              </span>
            </div>
          </Alert>
        {/if}
        <Label class="space-y-2 text-xs">
          <span>名前</span>
          <Input class="h-8" bind:value={caReq.Name} size="sm" />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>DNS名</span>
          <Input
            class="h-8"
            bind:value={caReq.SANs}
            placeholder="SANs"
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>ACMEサーバーの基本URL</span>
          <Input
            class="h-8"
            bind:value={caReq.AcmeBaseURL}
            placeholder="https://<host|ip>:port"
            size="sm"
          />
        </Label>
        <Label class="space-y-2 text-xs">
          <span>OCSP/SCEP/CRLの基本URL</span>
          <Input
            class="h-8"
            bind:value={caReq.HttpBaseURL}
            placeholder="http://<host|ip>:port"
            size="sm"
          />
        </Label>
        <div class="grid gap-4 grid-cols-5">
          <Label class="space-y-2 text-xs">
            <span>CAの鍵タイプ</span>
            <Select
              items={keyTypeList}
              bind:value={caReq.RootCAKeyType}
              placeholder=""
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span>CA証明証の有効期間(年)</span>
            <Input
              class="h-8 w-24 text-right"
              type="number"
              min="1"
              max="100"
              bind:value={caReq.RootCATerm}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span>CRLの更新間隔(時間)</span>
            <Input
              class="h-8 w-24 text-right"
              type="number"
              min="1"
              max="96"
              bind:value={caReq.CrlInterval}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span>OCSP/SCEPサーバーのポート番号</span>
            <Input
              class="h-8 w-24 text-right"
              type="number"
              min="1"
              max="65534"
              bind:value={caReq.HttpPort}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span>ACMEサーバーのポート番号</span>
            <Input
              class="h-8 w-24 text-right"
              type="number"
              min="1"
              max="65534"
              bind:value={caReq.AcmePort}
              size="sm"
            />
          </Label>
        </div>
      </form>
    {:else}
      {#if createCRTErr}
        <Alert color="red" dismissable>
          <div class="flex">
            <Icon path={icons.mdiExclamation} size={1} />
            {createCRTErr}
          </div>
        </Alert>
      {/if}
      <table id="certListTable" class="display compact" style="width:99%" />
    {/if}
  </div>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      color="green"
      type="button"
      on:click={() => (showCreateCSRDialog = true)}
      size="xs"
    >
      <Icon path={icons.mdiKey} size={1} />
      CSR作成
    </GradientButton>
    {#if showCreateCA}
      <GradientButton
        shadow
        color="blue"
        type="button"
        disabled={wait}
        on:click={doCreateCA}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        CA構築
      </GradientButton>
    {:else}
      {#if selectedCount > 0}
        <GradientButton
          shadow
          color="lime"
          type="button"
          on:click={exportCert}
          size="xs"
        >
          <Icon path={icons.mdiContentSave} size={1} />
          エクスポート
        </GradientButton>
        {#if selectedForRevoke != ""}
          <GradientButton
            shadow
            color="red"
            type="button"
            on:click={revokeCert}
            size="xs"
          >
            <Icon path={icons.mdiBookRemove} size={1} />
            失効
          </GradientButton>
        {/if}
      {/if}
      <GradientButton
        shadow
        color="green"
        type="button"
        on:click={doCreateCRT}
        size="xs"
      >
        <Icon path={icons.mdiCertificate} size={1} />
        証明書作成
      </GradientButton>
      <GradientButton
        shadow
        color="red"
        type="button"
        on:click={destroyCA}
        size="xs"
      >
        <Icon path={icons.mdiTrashCan} size={1} />
        CA初期化
      </GradientButton>
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={() => (showPKIControlDialog = true)}
        size="xs"
      >
        <Icon path={icons.mdiCog} size={1} />
        サーバー制御
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={refresh}
        size="xs"
      >
        <Icon path={icons.mdiRecycle} size={1} />
        {$_("AIList.Reload")}
      </GradientButton>
    {/if}
  </div>
</div>

<Modal
  bind:open={showPKIControlDialog}
  size="sm"
  dismissable={false}
  class="w-full"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      サーバーの制御
    </h3>
    <Alert
      color={pkiControl.HttpStatus.indexOf("error") != -1 ? "red" : "blue"}
      class="m-1 text-xs p-0"
    >
      <div class="flex">
        <Icon
          path={pkiControl.HttpStatus.indexOf("error") != -1
            ? icons.mdiExclamation
            : icons.mdiInformation}
          size={1}
        />
        <span class="ml-2">
          CRL/OCSP/SCEPサーバー:
        </span>
        {pkiControl.HttpStatus}
      </div>
    </Alert>
    <Alert
      color={pkiControl.AcmeStatus.indexOf("error") != -1 ? "red" : "blue"}
      class="m-1 text-xs p-0"
    >
      <div class="flex">
        <Icon
          path={pkiControl.AcmeStatus.indexOf("error") != -1
            ? icons.mdiExclamation
            : icons.mdiInformation}
          size={1}
        />
        <span class="ml-2">
          ACMEサーバー:
        </span>
        {pkiControl.AcmeStatus}
      </div>
    </Alert>
    <div class="grid gap-4 mb-4 grid-cols-2">
      <Checkbox bind:checked={pkiControl.EnableAcme}>ACMEサーバー</Checkbox>
      <Checkbox bind:checked={pkiControl.EnableHttp}>OCSP/SCEPサーバー</Checkbox
      >
    </div>
    <Label class="space-y-2 text-xs">
      <span>ACMEサーバーの基本URL</span>
      <Input
        class="h-8"
        bind:value={pkiControl.AcmeBaseURL}
        placeholder="https://<host|ip>:port"
        size="sm"
      />
    </Label>
  </form>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="blue"
      on:click={doPKIControl}
      size="xs"
    >
      <Icon path={icons.mdiController} size={1} />
      設定
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={() => {
        showPKIControlDialog = false;
      }}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      キャンセル
    </GradientButton>
  </div>
</Modal>

<Modal
  bind:open={showCreateCSRDialog}
  size="sm"
  dismissable={false}
  class="w-full"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      証明書要求(CSR)の作成
    </h3>
    <Label class="space-y-2 text-xs">
      <span>鍵タイプ</span>
      <Select
        items={keyTypeList}
        bind:value={csrReq.KeyType}
        placeholder=""
        size="sm"
      />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>名前(Common Name)</span>
      <Input class="h-8" bind:value={csrReq.CommonName} size="sm" />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>DNS名(Subject Alt Names)</span>
      <Input class="h-8" bind:value={csrReq.SANs} size="sm" />
    </Label>
    <div class="grid gap-4 mb-4 grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span>組織名(Organization)</span>
        <Input class="h-8" bind:value={csrReq.Organization} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>組織単位(Organizational Unit)</span>
        <Input class="h-8" bind:value={csrReq.OrganizationalUnit} size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 mb-4 grid-cols-3">
      <Label class="space-y-2 text-xs">
        <span>国コード(Country)</span>
        <Input class="h-8" bind:value={csrReq.Country} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>州/都道府県名(Province)</span>
        <Input class="h-8" bind:value={csrReq.Province} size="sm" />
      </Label>
      <Label class="space-y-2 text-xs">
        <span>市町村名(Locality)</span>
        <Input class="h-8" bind:value={csrReq.Locality} size="sm" />
      </Label>
    </div>
  </form>
  <div class="flex justify-end space-x-2 mr-2">
    <GradientButton
      shadow
      type="button"
      color="blue"
      on:click={doCreateCSR}
      size="xs"
    >
      <Icon path={icons.mdiContentSave} size={1} />
      作成
    </GradientButton>
    <GradientButton
      shadow
      type="button"
      color="teal"
      on:click={() => {
        showCreateCSRDialog = false;
      }}
      size="xs"
    >
      <Icon path={icons.mdiCancel} size={1} />
      キャンセル
    </GradientButton>
  </div>
</Modal>
