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
  } from "flowbite-svelte";
  import Icon from "mdi-svelte";
  import { 
    mdiLan,
    mdiLaptop,
    mdiLanCheck,
    mdiCalendarCheck,
    mdiBrain,
    mdiCog,
    mdiEmail,
    mdiContentSave,
    mdiCancel,
    mdiMoonWaxingCrescent,
    mdiWeatherSunny,
  } from "@mdi/js";
  import { setMAP, showMAP, setMapContextMenu } from "./lib/map";
  import { onMount, tick } from "svelte";
  import {GetMapConf, GetSettings, GetVersion,SetMapConf} from "../wailsjs/go/main/App"
  import { snmpModeList } from "./lib/common";
  import type { datastore} from "wailsjs/go/models";
  let version = "";
  let settings :any = undefined;
  let map: any;
	let dark: boolean = false;
  let showMapConf :boolean= false;
  let mapConf: datastore.MapConfEnt | undefined = undefined;

  let page = "map";

  const saveMapConf = async () => {
    mapConf.PollInt *=1;
    mapConf.Timeout *=1;
    mapConf.Retry *=1;
    mapConf.LogDays *=1;
    console.log(mapConf);
    await SetMapConf(mapConf);
    showMapConf = false;
  }
  onMount(async () => {
    version = await GetVersion();
    settings = await GetSettings();
    mapConf = await GetMapConf();
    await tick();
    showMAP(map);
		maptest();
    setMapContextMenu(true);
  });
	const maptest = async() => {
    await tick();
    setMAP(
      {
        Nodes: {
          node1: {
            ID: "node1",
            X: 100,
            Y: 200,
            Icon: "mdi-microsoft-windows",
            State: "normal",
            Name: "Node1",
          },
          node2: {
            ID: "node2",
            X: 160,
            Y: 200,
            Icon: "mdi-linux",
            State: "low",
            Name: "Node2",
          },
        },
        Lines: {
          line1: {
            ID: "line1",
            NodeID1: "node1",
            NodeID2: "node2",
            State1: "normal",
            State2: "low",
          },
        },
        Items: {
          item1: {
            ID: "item1",
            Type: 2,
            Size: 24,
            X: 50,
            Y: 100,
            Text: "test",
            Color: "red",
          },
        },
        MapConf: {
          BackImage: {
             Color: dark ? "black" : "white",
          },
        },
      },
			dark,
      settings.Lock,
    );
	}
	const toggleDark = () => {
		const e = document.querySelector('html');
		e.classList.toggle('dark');
		dark = e.classList.contains('dark');
		maptest();
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
      <DropdownItem>
        通知
      </DropdownItem>
      <DropdownItem>
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

<div bind:this={map} class="w-full h-screen" />

<Modal bind:open={showMapConf} size="lg" autoclose={false} class="w-full">
  <form class="flex flex-col space-y-6" action="#">
    <h3 class="mb-1 font-medium text-gray-900 dark:text-white">マップ設定</h3>
    <Label class="space-y-2">
      <span>マップ名</span>
      <Input bind:value={mapConf.MapName} placeholder="マップ名" required  size="sm"/>
    </Label>
    <div class="flex justify-between">
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
    <div class="flex justify-between">
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
    <div class="flex justify-between">
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