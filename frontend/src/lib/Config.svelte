<script lang="ts">
  import {
    Modal,
    Tabs,
    TabItem,
    Checkbox,
    Label,
    Input,
    Select,
    GradientButton,
    Alert,
    Range,
    Spinner,
    Textarea,
    Button,
  } from "flowbite-svelte";
  import { createEventDispatcher } from "svelte";
  import { Icon } from "mdi-svelte-ts";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import {
    snmpModeList,
    getTableLang,
    getStateIcon,
    getStateColor,
    setIconToList,
    deleteIconFromList,
  } from "./common";
  import {
    GetMapConf,
    UpdateMapConf,
    GetNotifyConf,
    UpdateNotifyConf,
    TestNotifyConf,
    GetAIConf,
    UpdateAIConf,
    GetMIBModules,
    GetMIBTree,
    GetLocConf,
    UpdateLocConf,
    GetIcons,
    UpdateIcon,
    DeleteIcon,
    SelectAudioFile,
    GetAudio,
    TestLine,
    GetSshdPublicKeys,
    SaveSshdPublicKeys,
    InitMySSHKey,
    GetMySSHPublicKey,
  } from "../../wailsjs/go/main/App";
  import { _ } from "svelte-i18n";
  import DataTable from "datatables.net-dt";
  import "datatables.net-select-dt";
  import MibTree from "./MIBTree.svelte";
  import { CodeJar } from "@novacbn/svelte-codejar";
  import Help from "./Help.svelte";
  import Prism from "prismjs";
  import { copyText } from "svelte-copy";

  const highlight = (code: string, syntax: string | undefined): string => {
    if (!syntax) {
      return "";
    }
    return Prism.highlight(code, Prism.languages[syntax], syntax);
  };

  export let show: boolean = false;
  let helpPage: any = undefined;
  let showHelp = false;
  let mapConf: any = undefined;

  let notifyConf: any = undefined;
  let showTestError: boolean = false;
  let showTestOk: boolean = false;
  let locConf: any = undefined;
  let showLocStyleError = false;

  let showLineTestError: boolean = false;
  let showLineTestOk: boolean = false;

  let showSSHPublicKey = false;
  let sshHostPublicKey = "";
  let sshMyPublicKey = "";

  const dispatch = createEventDispatcher();

  const onOpen = async () => {
    mapConf = await GetMapConf();
    notifyConf = await GetNotifyConf();
    aiConf = await GetAIConf();
    locConf = await GetLocConf();
    sshHostPublicKey = await GetSshdPublicKeys();
    sshMyPublicKey = await GetMySSHPublicKey();
  };

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const saveMapConf = async () => {
    mapConf.PollInt *= 1;
    mapConf.Timeout *= 1;
    mapConf.Retry *= 1;
    mapConf.LogDays *= 1;
    mapConf.IconSize *= 1;
    const r = await UpdateMapConf(mapConf);
    close();
  };

  const notifyLevelList = [
    { name: $_("Config.None"), value: "none" },
    { name: $_("Config.Warn"), value: "warn" },
    { name: $_("Config.Low"), value: "low" },
    { name: $_("Config.High"), value: "high" },
  ];

  const saveNotifyConf = async () => {
    notifyConf.Interval *= 1;
    await UpdateNotifyConf(notifyConf);
    close();
  };

  const testMail = async () => {
    showTestError = false;
    notifyConf.Interval *= 1;
    const ok = await TestNotifyConf(notifyConf);
    showTestError = !ok;
    showTestOk = ok;
  };

  const testLine = async () => {
    showLineTestError = false;
    notifyConf.Interval *= 1;
    const ok = await TestLine(notifyConf);
    showLineTestError = !ok;
    showLineTestOk = ok;
  };

  let showAudioError = false;
  const selectBeep = async (h: any) => {
    showAudioError = false;
    const p = await SelectAudioFile(
      h ? $_("Config.SelectAudioHigh") : $_("Config.SelectAudioLow")
    );
    if (p == "") {
      return;
    }
    const s = await GetAudio(p);
    if (s == "") {
      showAudioError = true;
      return;
    }
    if (h) {
      notifyConf.BeepHigh = s;
    } else {
      notifyConf.BeepLow = s;
    }
  };

  const deleteBeep = (h: any) => {
    if (h) {
      notifyConf.BeepHigh = "";
    } else {
      notifyConf.BeepLow = "";
    }
  };

  let aiConf: any = undefined;

  const aiLevelList = [
    { name: $_("Config.AILevel0"), value: 0 },
    { name: $_("Config.AILivel110"), value: 110.8 },
    { name: $_("Config.AILevel106"), value: 106.1 },
    { name: $_("Config.AILevel101"), value: 101.9 },
    { name: $_("Config.AILevel97"), value: 97.5 },
    { name: $_("Config.AILevel92"), value: 92.6 },
    { name: $_("Config.AILevel86"), value: 86.8 },
    { name: $_("Config.AILevel80"), value: 80.8 },
    { name: $_("Config.AILevel73"), value: 73.2 },
    { name: $_("Config.AILevel62"), value: 62.8 },
  ];

  const saveAIConf = async () => {
    aiConf.HighThreshold *= 1;
    aiConf.LowThreshold *= 1;
    aiConf.WarnThreshold *= 1;
    await UpdateAIConf(aiConf);
    close();
  };

  let showMIBTree = false;
  let mibTree: any = {
    oid: ".1",
    name: ".iso",
    MIBInfo: null,
    children: undefined,
  };

  const renderType = (d: any, t: string, r: any) => {
    if (t == "sort") {
      return t;
    }
    const state = r.Error ? "high" : "info";
    const name = d == "int" ? $_("Config.IntMIB") : $_("Config.ExeMIB");
    return (
      `<span class="mdi ` +
      getStateIcon(state) +
      ` text-xs" style="color:` +
      getStateColor(state) +
      `;"></span><span class="ml-2">` +
      name +
      `</span>`
    );
  };

  const showMIBModules = async () => {
    if (!mibTree.children) {
      mibTree.children = await GetMIBTree();
    }
    const mibModules = await GetMIBModules();
    new DataTable("#mibModuleTable", {
      data: mibModules,
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      language: getTableLang(),
      stateSave: true,
      order: [[3, "asc"]],
      columns: [
        {
          title: $_("Config.MIBType"),
          data: "Type",
          width: "10%",
          render: renderType,
        },
        { title: $_("Config.MIBName"), data: "Name", width: "30%" },
        { title: $_("Config.MIBFile"), data: "File", width: "30%" },
        { title: $_("Config.MIBError"), data: "Error", width: "30%" },
      ],
    });
  };

  const saveLocConf = async () => {
    showLocStyleError = false;
    locConf.Style.trim();
    if (locConf.Style.startsWith("{")) {
      try {
        const s = JSON.parse(locConf.Style);
      } catch (e) {
        showLocStyleError = true;
        return;
      }
    }
    locConf.Zoom *= 1;
    locConf.IconSize *= 1;
    await UpdateLocConf(locConf);
    close();
  };

  let icon: datastore.IconEnt = {
    Name: "",
    Icon: "",
    Code: 0,
  };
  let iconTable: any = undefined;
  let showEditIcon = false;
  let selectedIcon = 0;
  let iconList: any = [];
  let disableIconSelect = false;
  const iconCodeMap = new Map();

  const showIconList = async () => {
    if (iconList.length < 1) {
      makeIconList();
    }
    if (iconTable && DataTable.isDataTable("#iconTable")) {
      iconTable.clear();
      iconTable.destroy();
      iconTable = undefined;
    }
    selectedIcon = 0;
    iconTable = new DataTable("#iconTable", {
      stateSave: true,
      order: [[1, "asc"]],
      pageLength: window.innerHeight > 1000 ? 25 : 10,
      columns: [
        {
          title: $_("Config.Icon"),
          data: "Icon",
          width: "20%",
          render: (i) => `<span class="mdi ${i} text-2xl"></span>`,
        },
        { title: $_("Config.Name"), data: "Name", width: "50%" },
        { title: $_("Config.Code"), data: "Code", width: "30%" },
      ],
      data: await GetIcons(),
      language: getTableLang(),
      select: {
        style: "single",
      },
    });
    iconTable.on("select", () => {
      selectedIcon = iconTable.rows({ selected: true }).count();
    });
    iconTable.on("deselect", () => {
      selectedIcon = iconTable.rows({ selected: true }).count();
    });
  };

  const addIcon = () => {
    icon = {
      Name: "",
      Icon: "",
      Code: 0,
    };
    disableIconSelect = false;
    showEditIcon = true;
  };
  const editIcon = () => {
    const selected = iconTable.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    icon = selected[0];
    disableIconSelect = true;
    showEditIcon = true;
  };

  const delIcon = async () => {
    const selected = iconTable.rows({ selected: true }).data();
    if (selected.length != 1) {
      return;
    }
    await DeleteIcon(selected[0].Icon);
    deleteIconFromList(selected[0].Icon);
    showIconList();
  };

  let hasIconTextError = false;
  const saveIocn = async () => {
    hasIconTextError = !icon.Name;
    if (icon && icon.Icon && icon.Name) {
      icon.Code = iconCodeMap.get(icon.Icon);
      if (icon.Code) {
        await UpdateIcon(icon);
        setIconToList(icon);
        showEditIcon = false;
        showIconList();
        return;
      }
    }
  };

  const makeIconList = () => {
    iconList = [];
    iconCodeMap.clear();
    const re = /mdi-[^:]+/;
    for (const ss of document.styleSheets) {
      if (!ss || !ss.cssRules) {
        continue;
      }
      for (const cr of ss.cssRules) {
        const e = cr as CSSStyleRule;
        if (
          e &&
          e.selectorText &&
          e.selectorText.includes("::before") &&
          e.style &&
          e.style.content
        ) {
          const m = e.selectorText.match(re);
          if (m) {
            const code =
              e.style.content && e.style.content.length > 2
                ? e.style.content.codePointAt(1)
                : 0;
            if (code !== 0) {
              iconList.push({
                name: m[0],
                value: m[0],
              });
              iconCodeMap.set(m[0], code);
            }
          }
        }
      }
    }
  };
  const saveSSHPublicKey = async () => {
    await SaveSshdPublicKeys(sshHostPublicKey);
    showSSHPublicKey = false;
  };

  let copied = false;

  const copyMySSHPublicKey = async () => {
    await copyText(sshMyPublicKey);
    copied = true;
    setTimeout(() => {
      copied = false;
    }, 2000);
  };
  const refreshMySSHPublicKey = async () => {
    await InitMySSHKey();
    sshMyPublicKey = await GetMySSHPublicKey();
  }
</script>

<Modal
  bind:open={show}
  size="xl"
  dismissable={false}
  class="w-full min-h-[90vh]"
  on:open={onOpen}
>
  {#if !locConf}
    <div class="text-center mt-10"><Spinner size={16} /></div>
  {:else}
    <Tabs style="underline">
      <TabItem open>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiCog} size={1} />
          {$_("Config.Map")}
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <div class="grid gap-2 grid-cols-4">
            <Label class="col-span-3 space-y-2 text-xs">
              <span>{$_("Config.MapName")}</span>
              <Input
                bind:value={mapConf.MapName}
                placeholder={$_("Config.MapName")}
                required
                size="sm"
              />
            </Label>
            <Label>
              {$_("Config.IconSize")}
              <Range size="sm" min="1" max="5" bind:value={mapConf.IconSize} />
            </Label>
          </div>
          <div class="grid gap-4 mb-4 md:grid-cols-4">
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.PollingIntSec")} </span>
              <Input
                type="number"
                min={5}
                max={3600 * 24}
                step={1}
                bind:value={mapConf.PollInt}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.TimeoutSec")} </span>
              <Input
                type="number"
                min={1}
                max={120}
                step={1}
                bind:value={mapConf.Timeout}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.Retry")} </span>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                bind:value={mapConf.Retry}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.LogDays")} </span>
              <Input
                type="number"
                min={1}
                max={365 * 5}
                step={1}
                bind:value={mapConf.LogDays}
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 md:grid-cols-3">
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.SNMPMode")} </span>
              <Select
                items={snmpModeList}
                bind:value={mapConf.SnmpMode}
                placeholder={$_("Config.SelectSnmpMode")}
                size="sm"
              />
            </Label>
            {#if mapConf.SnmpMode == "v1" || mapConf.SnmpMode == "v2c"}
              <Label class="space-y-2 text-xs">
                <span>SNMP Community</span>
                <Input
                  bind:value={mapConf.Community}
                  placeholder="public"
                  size="sm"
                />
              </Label>
            {:else}
              <Label class="space-y-2 text-xs">
                <span>{$_("Config.SnmpUser")}</span>
                <Input
                  bind:value={mapConf.SnmpUser}
                  placeholder="snmp user"
                  size="sm"
                />
              </Label>
              <Label class="space-y-2 text-xs">
                <span>{$_("Config.SnmpPassword")}</span>
                <Input
                  type="password"
                  bind:value={mapConf.SnmpPassword}
                  placeholder="•••••"
                  size="sm"
                />
              </Label>
            {/if}
          </div>
          <div class="grid gap-4 mb-4 md:grid-cols-6">
            <Checkbox bind:checked={mapConf.EnableSyslogd}>Syslog</Checkbox>
            <Checkbox bind:checked={mapConf.EnableNetflowd}>NetFlow</Checkbox>
            <Checkbox bind:checked={mapConf.EnableSFlowd}>sFlow</Checkbox>
            <Checkbox bind:checked={mapConf.EnableTrapd}>SNMP TRAP</Checkbox>
            <Checkbox bind:checked={mapConf.EnableArpWatch}>ARP Watch</Checkbox>
            <Checkbox bind:checked={mapConf.EnableSshd}>SSH Sever</Checkbox>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={() => (showSSHPublicKey = true)}
              size="xs"
            >
              <Icon path={icons.mdiKeyChain} size={1} />
              {$_("Config.SSHPublicKey")}
            </GradientButton>
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={saveMapConf}
              size="xs"
            >
              <Icon path={icons.mdiContentSave} size={1} />
              {$_("Config.Save")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              size="xs"
              color="lime"
              class="ml-2"
              on:click={() => {
                helpPage = "mapconf";
                showHelp = true;
              }}
            >
              <Icon path={icons.mdiHelp} size={1} />
              <span>
                {$_("Config.Help")}
              </span>
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="teal"
              on:click={close}
              size="xs"
            >
              <Icon path={icons.mdiCancel} size={1} />
              {$_("Config.Cancel")}
            </GradientButton>
          </div>
        </form>
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiSend} size={1} />
          {$_("Config.Notify")}
        </div>
        <form class="flex flex-col space-y-4" action="#">
          {#if showTestError}
            <Alert color="red" dismissable>
              <div class="flex">
                <Icon path={icons.mdiExclamation} size={1} />
                {$_("Config.FailedSendMail")}
              </div>
            </Alert>
          {/if}
          {#if showLineTestError}
            <Alert color="red" dismissable>
              <div class="flex">
                <Icon path={icons.mdiExclamation} size={1} />
                {$_("Config.LineTestNG")}
              </div>
            </Alert>
          {/if}
          {#if showAudioError}
            <Alert color="red" dismissable>
              <div class="flex">
                <Icon path={icons.mdiExclamation} size={1} />
                {$_("Config.SelectAudioError")}
              </div>
            </Alert>
          {/if}
          {#if showTestOk}
            <Alert class="flex" color="blue" dismissable>
              <div class="flex">
                <Icon path={icons.mdiCheck} size={1} />
                {$_("Config.SentTestMail")}
              </div>
            </Alert>
          {/if}
          {#if showLineTestOk}
            <Alert class="flex" color="blue" dismissable>
              <div class="flex">
                <Icon path={icons.mdiCheck} size={1} />
                {$_("Config.LineTestOK")}
              </div>
            </Alert>
          {/if}
          <div class="grid gap-4 grid-cols-4">
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.MailServer")}</span>
              <Input
                bind:value={notifyConf.MailServer}
                placeholder="host|ip:port"
                required
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.InsecureSkipVerify}>
              {$_("Config.NoCheckCert")}
            </Checkbox>
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.SmtpUser")}</span>
              <Input
                bind:value={notifyConf.User}
                placeholder="smtp user"
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.SmtpPassword")}</span>
              <Input
                type="password"
                bind:value={notifyConf.Password}
                placeholder="•••••"
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 grid-cols-4">
            <Label class="space-y-2 text-xs col-span-2">
              <span> {$_("Config.Subject")} </span>
              <Input bind:value={notifyConf.Subject} size="sm" />
            </Label>
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.MailFrom")}</span>
              <Input
                bind:value={notifyConf.MailFrom}
                placeholder={$_("Config.MailFromAddress")}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.MailTo")}</span>
              <Input
                bind:value={notifyConf.MailTo}
                placeholder={$_("Config.MailToAddress")}
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 grid-cols-4">
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.NotifyLevel")} </span>
              <Select
                items={notifyLevelList}
                bind:value={notifyConf.Level}
                placeholder={$_("Config.SelectNotifyLevel")}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              <span> {$_("Config.NotifyIntSec")} </span>
              <Input
                type="number"
                min={60}
                max={3600 * 24}
                step={10}
                bind:value={notifyConf.Interval}
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.Report}
              >{$_("Config.MailReport")}</Checkbox
            >
            <Checkbox bind:checked={notifyConf.NotifyRepair}
              >{$_("Config.NotifyRepair")}</Checkbox
            >
          </div>
          <div class="grid gap-4 grid-cols-4">
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.LineLevel")}</span>
              <Select
                items={notifyLevelList}
                bind:value={notifyConf.LineLevel}
                placeholder={$_("Config.SelectNotifyLevel")}
                size="sm"
              />
            </Label>
            <Checkbox class="mt-6" bind:checked={notifyConf.LineNotifyRepair}>
              {$_("Config.NotifyRepair")}
            </Checkbox>
            <Label class="space-y-2 text-xs col-span-2">
              <span> {$_("Config.LineToken")} </span>
              <Input
                class="w-full"
                type="password"
                bind:value={notifyConf.LineToken}
                size="sm"
              />
            </Label>
          </div>
          <Label class="space-y-2 text-xs">
            <span> {$_("Config.ExecCommand")} </span>
            <Input class="w-full" bind:value={notifyConf.ExecCmd} size="sm" />
          </Label>
          <div class="grid gap-4 grid-cols-4">
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.AudioHigh")}</span>
              {#if notifyConf.BeepHigh}
                <audio src={notifyConf.BeepHigh} controls />
              {/if}
            </Label>
            {#if notifyConf.BeepHigh}
              <GradientButton
                shadow
                class="h-8 mt-6 w-28"
                color="red"
                type="button"
                on:click={() => deleteBeep(true)}
                size="xs"
              >
                <Icon path={icons.mdiTrashCan} size={1} />
                {$_("Config.Delete")}
              </GradientButton>
            {:else}
              <GradientButton
                shadow
                class="h-8 mt-6 w-28"
                color="blue"
                type="button"
                on:click={() => selectBeep(true)}
                size="xs"
              >
                <Icon path={icons.mdiSoundbar} size={1} />
                {$_("Config.SelectAodio")}
              </GradientButton>
            {/if}
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.AodioLow")}</span>
              {#if notifyConf.BeepLow}
                <audio src={notifyConf.BeepLow} controls />
              {/if}
            </Label>
            {#if notifyConf.BeepLow}
              <GradientButton
                shadow
                class="h-8 mt-6 w-28"
                color="red"
                type="button"
                on:click={() => deleteBeep(false)}
                size="xs"
              >
                <Icon path={icons.mdiTrashCan} size={1} />
                {$_("Config.Delete")}
              </GradientButton>
            {:else}
              <GradientButton
                shadow
                class="h-8 mt-6 w-28"
                color="blue"
                type="button"
                on:click={() => selectBeep(false)}
                size="xs"
              >
                <Icon path={icons.mdiSoundbar} size={1} />
                {$_("Config.SelectAodio")}
              </GradientButton>
            {/if}
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={saveNotifyConf}
              size="xs"
            >
              <Icon path={icons.mdiContentSave} size={1} />
              {$_("Config.Save")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="red"
              on:click={testMail}
              size="xs"
            >
              <Icon path={icons.mdiEmail} size={1} />
              {$_("Config.Test")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="red"
              on:click={testLine}
              size="xs"
            >
              <Icon path={icons.mdiChat} size={1} />
              LINE{$_("Config.Test")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              size="xs"
              color="lime"
              class="ml-2"
              on:click={() => {
                helpPage = "notifyconf";
                showHelp = true;
              }}
            >
              <Icon path={icons.mdiHelp} size={1} />
              <span>
                {$_("Config.Help")}
              </span>
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="teal"
              on:click={close}
              size="xs"
            >
              <Icon path={icons.mdiCancel} size={1} />
              {$_("Config.Cancel")}
            </GradientButton>
          </div>
        </form>
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiBrain} size={1} />
          {$_("Config.AI")}
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <Label class="space-y-2 text-xs">
            <span> {$_("Config.AIHighLevel")} </span>
            <Select
              items={aiLevelList}
              bind:value={aiConf.HighThreshold}
              placeholder={$_("Config.SelectAILevel")}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span> {$_("Config.AILevelLow")} </span>
            <Select
              items={aiLevelList}
              bind:value={aiConf.LowThreshold}
              placeholder={$_("Config.SelectAILevel")}
              size="sm"
            />
          </Label>
          <Label class="space-y-2 text-xs">
            <span> {$_("Config.AIlevelWarn")} </span>
            <Select
              items={aiLevelList}
              bind:value={aiConf.WarnThreshold}
              placeholder={$_("Config.SelectAILevel")}
              size="sm"
            />
          </Label>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={saveAIConf}
              size="xs"
            >
              <Icon path={icons.mdiContentSave} size={1} />
              {$_("Config.Save")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              size="xs"
              color="lime"
              class="ml-2"
              on:click={() => {
                helpPage = "aiconf";
                showHelp = true;
              }}
            >
              <Icon path={icons.mdiHelp} size={1} />
              <span>
                {$_("Config.Help")}
              </span>
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="teal"
              on:click={close}
              size="xs"
            >
              <Icon path={icons.mdiCancel} size={1} />
              {$_("Config.Cancel")}
            </GradientButton>
          </div>
        </form>
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiMap} size={1} />
          {$_("Config.LocConf")}
        </div>
        <form class="flex flex-col space-y-4" action="#">
          {#if showLocStyleError}
            <Alert color="red" dismissable>
              <div class="flex">
                <Icon path={icons.mdiAlert} size={1} />
                {$_("Config.LocStyleError")}
              </div>
            </Alert>
          {/if}
          <Label class="space-y-2 text-xs">
            <span>{$_("Config.LocStyle")}</span>
            <CodeJar
              syntax="javascript"
              {highlight}
              bind:value={locConf.Style}
            />
          </Label>
          <div class="grid gap-4 md:grid-cols-3">
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.LocCenter")}</span>
              <Input type="text" bind:value={locConf.Center} size="sm" />
            </Label>
            <Label class="space-y-2 text-xs">
              <span>{$_("Config.LocZoom")}</span>
              <Input
                type="number"
                min="2"
                max="12"
                bind:value={locConf.Zoom}
                size="sm"
              />
            </Label>
            <Label class="space-y-2 text-xs">
              {$_("Config.IconSize")}
              <Range
                size="sm"
                min="16"
                max="64"
                bind:value={locConf.IconSize}
              />
            </Label>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={saveLocConf}
              size="xs"
            >
              <Icon path={icons.mdiContentSave} size={1} />
              {$_("Config.Save")}
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              size="xs"
              color="lime"
              class="ml-2"
              on:click={() => {
                helpPage = "locconf";
                showHelp = true;
              }}
            >
              <Icon path={icons.mdiHelp} size={1} />
              <span>
                {$_("Config.Help")}
              </span>
            </GradientButton>
            <GradientButton
              shadow
              type="button"
              color="teal"
              on:click={close}
              size="xs"
            >
              <Icon path={icons.mdiCancel} size={1} />
              {$_("Config.Cancel")}
            </GradientButton>
          </div>
        </form>
      </TabItem>
      <TabItem on:click={showIconList}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiDotsGrid} size={1} />
          {$_("Config.IconMan")}
        </div>
        <table id="iconTable" class="display compact mt-2" style="width:99%" />
        <div class="flex justify-end space-x-2 mr-2 mt-3">
          <GradientButton
            shadow
            color="blue"
            type="button"
            on:click={addIcon}
            size="xs"
          >
            <Icon path={icons.mdiPlus} size={1} />
            {$_("Config.Add")}
          </GradientButton>
          {#if selectedIcon}
            <GradientButton
              shadow
              color="blue"
              type="button"
              on:click={editIcon}
              size="xs"
            >
              <Icon path={icons.mdiPencil} size={1} />
              {$_("Config.Edit")}
            </GradientButton>
            <GradientButton
              shadow
              color="red"
              type="button"
              on:click={delIcon}
              size="xs"
            >
              <Icon path={icons.mdiTrashCan} size={1} />
              {$_("Config.Delete")}
            </GradientButton>
          {/if}
          <GradientButton
            shadow
            type="button"
            size="xs"
            color="lime"
            class="ml-2"
            on:click={() => {
              helpPage = "iconconf";
              showHelp = true;
            }}
          >
            <Icon path={icons.mdiHelp} size={1} />
            <span>
              {$_("Config.Help")}
            </span>
          </GradientButton>
          <GradientButton
            shadow
            type="button"
            color="teal"
            on:click={close}
            size="xs"
          >
            <Icon path={icons.mdiCancel} size={1} />
            {$_("Config.Close")}
          </GradientButton>
        </div>
      </TabItem>
      <TabItem on:click={showMIBModules}>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiFileTree} size={1} />
          {$_("Config.MIB")}
        </div>
        <table
          id="mibModuleTable"
          class="display compact mt-2"
          style="width:99%"
        />
        <div class="flex justify-end space-x-2 mr-2">
          <GradientButton
            shadow
            color="lime"
            type="button"
            on:click={() => (showMIBTree = true)}
            size="xs"
          >
            <Icon path={icons.mdiFileTree} size={1} />
            {$_("Config.MIBTree")}
          </GradientButton>
          <GradientButton
            shadow
            type="button"
            size="xs"
            color="lime"
            class="ml-2"
            on:click={() => {
              helpPage = "mibconf";
              showHelp = true;
            }}
          >
            <Icon path={icons.mdiHelp} size={1} />
            <span>
              {$_("Config.Help")}
            </span>
          </GradientButton>
          <GradientButton
            shadow
            type="button"
            color="teal"
            on:click={close}
            size="xs"
          >
            <Icon path={icons.mdiCancel} size={1} />
            {$_("Config.Close")}
          </GradientButton>
        </div>
      </TabItem>
    </Tabs>
  {/if}
</Modal>

<Modal
  bind:open={showMIBTree}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <div class="flex flex-col space-y-4">
    <div id="mibtree">
      <MibTree tree={mibTree} on:select={(e) => {}} />
    </div>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showMIBTree = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Config.Close")}
      </GradientButton>
    </div>
  </div>
</Modal>

<Modal
  bind:open={showEditIcon}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Config.EditIcon")}
    </h3>
    <div class="grid gap-4 mb-4 md:grid-cols-2">
      <Label class="space-y-2 text-xs">
        <span> {$_("Node.Icon")} </span>
        <Select
          items={iconList}
          bind:value={icon.Icon}
          placeholder={$_("Config.SelectIcon")}
          size="sm"
          disabled={disableIconSelect}
        />
      </Label>
      <div class="mt-5 ml-5">
        <span class="mdi {icon.Icon} text-4xl" />
      </div>
    </div>
    <Label class="space-y-2 text-xs">
      <span>{$_("Config.Name")}</span>
      <Input
        bind:value={icon.Name}
        required
        size="sm"
        color={hasIconTextError ? "red" : "base"}
      />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={saveIocn}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_("Config.Save")}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showEditIcon = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Config.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Modal
  bind:open={showSSHPublicKey}
  size="lg"
  dismissable={false}
  class="w-full min-h-[80vh]"
>
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">
      {$_("Config.SSHPublicKey")}
    </h3>
    <Label class="space-y-2 text-xs">
      <span>
        {$_("Config.MySSHPublicKey")}
        <Button
          color="alternative"
          type="button"
          class="ml-2 !p-2"
          on:click={copyMySSHPublicKey}
          size="xs"
        >
          {#if copied}
            <Icon path={icons.mdiCheck} size={1} />
          {:else}
            <Icon path={icons.mdiContentCopy} size={1} />
          {/if}
        </Button>
        <Button
          color="red"
          type="button"
          class="ml-2 !p-2"
          on:click={refreshMySSHPublicKey}
          size="xs"
        >
          <Icon path={icons.mdiRefresh} size={1} />
        </Button>
      </span>
      <Textarea rows="8" bind:value={sshMyPublicKey} readonly />
    </Label>
    <Label class="space-y-2 text-xs">
      <span>{$_("Config.HostSSHPublicKey")}</span>
      <Textarea rows="8" bind:value={sshHostPublicKey} />
    </Label>
    <div class="flex justify-end space-x-2 mr-2">
      <GradientButton
        shadow
        color="blue"
        type="button"
        on:click={saveSSHPublicKey}
        size="xs"
      >
        <Icon path={icons.mdiContentSave} size={1} />
        {$_("Config.Save")}
      </GradientButton>
      <GradientButton
        shadow
        type="button"
        color="teal"
        on:click={() => {
          showSSHPublicKey = false;
        }}
        size="xs"
      >
        <Icon path={icons.mdiCancel} size={1} />
        {$_("Config.Cancel")}
      </GradientButton>
    </div>
  </form>
</Modal>

<Help bind:show={showHelp} page={helpPage} />

<style>
  #mibtree {
    height: 70vh;
    overflow: scroll;
  }
</style>
