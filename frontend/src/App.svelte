<script lang="ts">
	import logo from "./assets/images/appicon.png"; 
  import {
    Navbar,
    NavBrand,
    NavLi,
    NavUl,
		Button,
    Dropdown,
    DropdownItem,
    Modal, Label, Input, Checkbox,
    Select,
    Alert,
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import { 
    mdiLan,
    mdiLaptop,
    mdiLanCheck,
    mdiCalendarCheck,
    mdiBrain,
    mdiCog,
    mdiContentSave,
    mdiCancel,
    mdiMoonWaxingCrescent,
    mdiWeatherSunny,
    mdiEmail,
    mdiExclamation,
  } from "@mdi/js";
  import { onMount, tick } from "svelte";
  import {
    GetMapConf, 
    GetSettings, 
    GetVersion,
    SetMapConf,
    GetNotifyConf,
    SetNotifyConf,
    TestNotifyConf,
    GetAIConf,
    SetAIConf,
  } from "../wailsjs/go/main/App"
  import { snmpModeList } from "./lib/common";
  import type { datastore} from "wailsjs/go/models";
  import Map from "./lib/Map.svelte";
  import Log from "./lib/Log.svelte";

  let version = "";
  let settings :any = undefined;
  let map: any;
	let dark: boolean = false;
  let showMapConf :boolean= false;
  let mapConf: datastore.MapConfEnt | undefined = undefined;
  let showNotifyConf :boolean= false;
  let notifyConf: datastore.NotifyConfEnt | undefined = undefined;
  let showTestNotifyError : boolean = false;
  let showTestNotifyOk : boolean = false;
  let showAIConf :boolean= false;
  let aiConf: datastore.AIConfEnt | undefined = undefined;

  let page = "map";

  const saveMapConf = async () => {
    mapConf.PollInt *=1;
    mapConf.Timeout *=1;
    mapConf.Retry *=1;
    mapConf.LogDays *=1;
    await SetMapConf(mapConf);
    showMapConf = false;
  }

  const notifyLevelList = [
    {name:"通知しない", value:"none"},
    {name:"注意", value:"warn"},
    {name:"軽度", value:"low"},
    {name:"重度", value:"high"},
  ];
 
  const saveNotifyConf = async () => {
    notifyConf.Interval *=1;
    await SetNotifyConf(notifyConf);
    showNotifyConf = false;
  }

  const testNotifyConf = async () => {
    showTestNotifyError = false;
    notifyConf.Interval *=1;
    const ok = await TestNotifyConf(notifyConf);
    showTestNotifyError = !ok;
    showTestNotifyOk = ok;
  }

  const aiLevelList = [
    {name:"判定しない", value:0},
    {name:"10億回に1回の確率", value:110.8},
    {name:"1億回に1回の確率", value:106.1},
    {name:"1000万回に1回の確率", value:101.9},
    {name:"100万回に1回の確率", value:97.5},
    {name:"10万回に1回の確率", value:92.6},
    {name:"1万回に1回の確率", value:86.8},
    {name:"1000回に1回の確率", value:80.8},
    {name:"100回に1回の確率", value:73.2},
    {name:"10回に1回の確率", value:62.8},
  ];
 
  const saveAIConf = async () => {
    aiConf.HighThreshold *=1;
    aiConf.LowThreshold *=1;
    aiConf.WarnThreshold *=1;
    await SetAIConf(aiConf);
    showAIConf = false;
  }

  onMount(async () => {
    version = await GetVersion();
    settings = await GetSettings();
    mapConf = await GetMapConf();
    notifyConf = await GetNotifyConf();
    aiConf = await GetAIConf();
  });

	const toggleDark = () => {
		const e = document.querySelector('html');
		e.classList.toggle('dark');
		dark = e.classList.contains('dark');
	}
</script>

<Navbar let:hidden let:toggle style="--wails-draggable:drag">
  <NavBrand href="/">
    <img
      src="{logo}"
      class="mr-3 h-12"
      alt="TWSNMP FK Logo"
    />
    <span class="self-center whitespace-nowrap text-xl font-semibold dark:text-white">
      TWSNMP FK - {mapConf?.MapName || ''}
    </span>
  </NavBrand>
  <NavUl>
    <NavLi active={page=="map"}>
			<Icon path={mdiLan} size={1} />
      マップ
    </NavLi>
    <NavLi >
			<Icon path={mdiLaptop} size={1} />
      ノード
    </NavLi>
    <NavLi >
			<Icon path={mdiLanCheck} size={1} />
      ポーリング
    </NavLi>
    <NavLi >
      <Icon path={mdiCalendarCheck} size={1} />
      ログ
    </NavLi>
    <NavLi>
			<Icon path={mdiBrain} size={1} />
      AI分析
    </NavLi>
    <NavLi id="nav-config">
			<Icon path={mdiCog} size={1} />
      設定
    </NavLi>
    <Dropdown triggeredBy="#nav-config" class="w-44 z-20">
      <DropdownItem on:click={()=> {showMapConf = true}}>
        マップ
      </DropdownItem>
      <DropdownItem on:click={()=> {
          showNotifyConf = true;
          showTestNotifyOk = false;
          showTestNotifyError = false;
        }}>
        通知
      </DropdownItem>
      <DropdownItem  on:click={()=> {showAIConf = true}} >
        AI分析
      </DropdownItem>
    </Dropdown>    
  </NavUl>
	<Button class="!p-2" color="alternative" on:click={toggleDark} >
		{#if dark}
			<Icon path={mdiWeatherSunny} size={1} />
		{:else}
			<Icon path={mdiMoonWaxingCrescent} size={1} />
		{/if}
	</Button>
</Navbar>


{#if page =="map"}
<div class="grid grid-rows-4 grid-cols-1 gap-0  w-full" style="height:{window.innerHeight - 96}px;">
  <div class="row-span-3">
    <Map {dark}></Map>
  </div>
  <div class="row-span-1">
    <Log></Log>
  </div>
</div>
{/if}

<Modal bind:open={showMapConf} size="lg" autoclose={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">マップ設定</h3>
    <Label class="space-y-2">
      <span>マップ名</span>
      <Input bind:value={mapConf.MapName} placeholder="マップ名" required  size="sm"/>
    </Label>
    <div class="grid gap-4 mb-4 md:grid-cols-4">
      <Label class="space-y-2">
        <span>
          ポーリング間隔(秒)
        </span>
        <Input type="number" min={5} max={3600*24} step={1} bind:value={mapConf.PollInt}  size="sm"/>
      </Label>
      <Label class="space-y-2">
        <span>
          タイムアウト(秒)
        </span>
        <Input type="number" min={1} max={120} step={1} bind:value={mapConf.Timeout}  size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>
          リトライ(回)
        </span>
        <Input type="number" min={0} max={100} step={1} bind:value={mapConf.Retry}  size="sm"/>
      </Label>
      <Label class="space-y-2">
        <span>
          ログ保存日数(日)
        </span>
        <Input type="number" min={1} max={365*5} step={1} bind:value={mapConf.LogDays}  size="sm"/>
      </Label>
    </div>
    <div class="grid gap-4  md:grid-cols-3">
      <Label class="space-y-2">
        <span>
          SNMPモード
        </span>
        <Select items={snmpModeList} bind:value={mapConf.SnmpMode} placeholder="SNMPのモードを選択" size="sm"/>
      </Label>
      {#if mapConf.SnmpMode == "v1" || mapConf.SnmpMode == "v2c" }
        <Label class="space-y-2">
          <span>SNMP Community</span>
          <Input bind:value={mapConf.Community} placeholder="public"  size="sm" />
        </Label>
      {:else}
        <Label class="space-y-2">
          <span>SNMPユーザー</span>
          <Input bind:value={mapConf.SnmpUser} placeholder="snmp user"  size="sm" />
        </Label>
        <Label class="space-y-2">
          <span>SNMPパスワード</span>
          <Input type="password" bind:value={mapConf.SnmpPassword}  placeholder="•••••"   size="sm" />
        </Label>
      {/if}
    </div>
    <div class="grid gap-4 mb-4 md:grid-cols-3">
        <Checkbox bind:checked={mapConf.EnableSyslogd}>Syslog受信</Checkbox>
        <Checkbox bind:checked={mapConf.EnableTrapd}>SNMP TRAP受信</Checkbox>
        <Checkbox bind:checked={mapConf.EnableArpWatch}>ARP監視</Checkbox>
    </div>
    <div class="flex space-x-2">
      <Button type="button" on:click={saveMapConf} size="sm" >
        <Icon path={mdiContentSave} size={1} />
        保存
      </Button>
      <Button type="button"color="alternative"  on:click={()=>{showMapConf = false}} size="sm" >
        <Icon path={mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>

<Modal bind:open={showNotifyConf} size="lg" autoclose={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">通知設定</h3>
    {#if showTestNotifyError }
      <Alert color="red" dismissable>
        <div class="flex">
          <Icon path={mdiExclamation} size={1} />
          通知メールの送信テストに失敗しました
        </div>
      </Alert>
    {/if}
    {#if showTestNotifyOk }
      <Alert class="flex" color="blue" dismissable>
        <div class="flex">
          <Icon path={mdiExclamation} size={1} />
          通知メールの送信テストに成功しました
        </div>
      </Alert>
    {/if}
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>メールサーバー</span>
        <Input bind:value={notifyConf.MailServer} placeholder="host|ip:port" required  size="sm"/>
      </Label>
      <Checkbox bind:checked={notifyConf.InsecureSkipVerify}>サーバー証明書をチェックしない</Checkbox>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>ユーザー</span>
        <Input bind:value={notifyConf.User} placeholder="smtp user"  size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>パスワード</span>
        <Input type="password" bind:value={notifyConf.Password}  placeholder="•••••"   size="sm" />
      </Label>
    </div>
    <div class="grid gap-4 md:grid-cols-2">
      <Label class="space-y-2">
        <span>送信元</span>
        <Input bind:value={notifyConf.MailFrom} placeholder="送信元メールアドレス"  size="sm" />
      </Label>
      <Label class="space-y-2">
        <span>宛先</span>
        <Input bind:value={notifyConf.MailTo}  placeholder="宛先メールアドレス"   size="sm" />
      </Label>
    </div>
    <Label class="space-y-2">
      <span>
        件名
      </span>
      <Input  bind:value={notifyConf.Subject}  size="sm"/>
    </Label>
    <div class="grid gap-4 md:grid-cols-4">
      <Label class="space-y-2">
        <span>
          通知レベル
        </span>
        <Select items={notifyLevelList} bind:value={notifyConf.Level} placeholder="通知レベルを選択" size="sm"/>
      </Label>
      <Label class="space-y-2">
        <span>
          通知間隔(秒)
        </span>
        <Input type="number" min={60} max={3600*24} step={10} bind:value={notifyConf.Interval}  size="sm" />
      </Label>
      <Checkbox bind:checked={notifyConf.Report}>定期レポート</Checkbox>
      <Checkbox bind:checked={notifyConf.NotifyRepair}>復帰通知</Checkbox>
    </div>
    <Label class="space-y-2">
      <span>
        コマンド実行
      </span>
      <Input  class="w-full" bind:value={notifyConf.ExecCmd}  size="sm"/>
    </Label>
    <div class="flex space-x-3">
      <Button type="button" on:click={saveNotifyConf} size="sm" >
        <Icon path={mdiContentSave} size={1} />
        保存
      </Button>
      <Button type="button" color="red" on:click={testNotifyConf} size="sm" >
        <Icon path={mdiEmail} size={1} />
        テスト
      </Button>
      <Button type="button" color="alternative"  on:click={()=>{showNotifyConf = false}} size="sm" >
        <Icon path={mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>

<Modal bind:open={showAIConf} size="lg" autoclose={false} class="w-full">
  <form class="flex flex-col space-y-4" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">AI分析設定</h3>
    <div class="grid gap-4  md:grid-cols-3">
      <Label class="space-y-2">
        <span>
          重度と判定するレベル
        </span>
        <Select items={aiLevelList} bind:value={aiConf.HighThreshold} placeholder="レベルを選択" size="sm"/>
      </Label>
      <Label class="space-y-2">
        <span>
          軽度と判定するレベル
        </span>
        <Select items={aiLevelList} bind:value={aiConf.LowThreshold} placeholder="レベルを選択" size="sm"/>
      </Label>
      <Label class="space-y-2">
        <span>
          注意と判定するレベル
        </span>
        <Select items={aiLevelList} bind:value={aiConf.WarnThreshold} placeholder="レベルを選択" size="sm"/>
      </Label>
    </div>
    <div class="flex space-x-2">
      <Button type="button" on:click={saveAIConf} size="sm" >
        <Icon path={mdiContentSave} size={1} />
        保存
      </Button>
      <Button type="button"color="alternative"  on:click={()=>{showAIConf = false}} size="sm" >
        <Icon path={mdiCancel} size={1} />
        キャンセル
      </Button>
    </div>
  </form>
</Modal>
