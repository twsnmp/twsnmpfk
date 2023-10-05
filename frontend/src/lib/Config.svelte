<script lang="ts">
  import { Modal,Tabs,TabItem,Checkbox,Label,Input,Select,GradientButton,Alert } from "flowbite-svelte";
  import { onMount, createEventDispatcher } from "svelte";
  import Icon from "mdi-svelte";
  import * as icons from "@mdi/js";
  import type { datastore } from "wailsjs/go/models";
  import { snmpModeList } from "./common";
  import { GetMapConf, UpdateMapConf,GetNotifyConf,UpdateNotifyConf,TestNotifyConf,GetAIConf,UpdateAIConf } from "../../wailsjs/go/main/App";
  import { _ } from 'svelte-i18n';

  let show: boolean = false;
  let mapConf: datastore.MapConfEnt | undefined = undefined;

  let notifyConf: datastore.NotifyConfEnt | undefined = undefined;
  let showTestError: boolean = false;
  let showTestOk: boolean = false;


  const dispatch = createEventDispatcher();

  onMount(async () => {
    mapConf = await GetMapConf();
    notifyConf = await GetNotifyConf();
    aiConf = await GetAIConf();
    show = true;
  });

  const close = () => {
    show = false;
    dispatch("close", {});
  };

  const saveMapConf = async () => {
    mapConf.PollInt *= 1;
    mapConf.Timeout *= 1;
    mapConf.Retry *= 1;
    mapConf.LogDays *= 1;
    const r = await UpdateMapConf(mapConf);
    close();
  };

  const notifyLevelList = [
    { name: $_('Config.Node'), value: "none" },
    { name: $_('Config.Warn'), value: "warn" },
    { name: $_('Config.Low'), value: "low" },
    { name: $_('Config.High'), value: "high" },
  ];

  const saveNotifyConf = async () => {
    notifyConf.Interval *= 1;
    await UpdateNotifyConf(notifyConf);
    close();
  };

  const testNotifyConf = async () => {
    showTestError = false;
    notifyConf.Interval *= 1;
    const ok = await TestNotifyConf(notifyConf);
    showTestError = !ok;
    showTestOk = ok;
  };

  let aiConf: datastore.AIConfEnt | undefined = undefined;

  const aiLevelList = [
    { name: $_('Config.AILevel0'), value: 0 },
    { name: $_('Config.AILivel110'), value: 110.8 },
    { name: $_('Config.AILevel106'), value: 106.1 },
    { name: $_('Config.AILevel101'), value: 101.9 },
    { name: $_('Config.AILevel97'), value: 97.5 },
    { name: $_('Config.AILevel92'), value: 92.6 },
    { name: $_('Config.AILevel86'), value: 86.8 },
    { name: $_('Config.AILevel80'), value: 80.8 },
    { name: $_('Config.AILevel73'), value: 73.2 },
    { name: $_('Config.AILevel62'), value: 62.8 },
  ];

  const saveAIConf = async () => {
    aiConf.HighThreshold *= 1;
    aiConf.LowThreshold *= 1;
    aiConf.WarnThreshold *= 1;
    await UpdateAIConf(aiConf);
    close();
  };

</script>

<Modal
  bind:open={show}
  size="xl"
  permanent
  class="w-full min-h-[90vh]"
  on:on:close={close}
>
    <Tabs style="underline">
      <TabItem open>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartPie} size={1} />
          { $_('Config.Map') }
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <Label class="space-y-2">
            <span>{ $_('Config.MapName') }</span>
            <Input
              bind:value={mapConf.MapName}
              placeholder="{ $_('Config.MapName') }"
              required
              size="sm"
            />
          </Label>
          <div class="grid gap-4 mb-4 md:grid-cols-4">
            <Label class="space-y-2">
              <span> { $_('Config.PollingIntSec') } </span>
              <Input
                type="number"
                min={5}
                max={3600 * 24}
                step={1}
                bind:value={mapConf.PollInt}
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.TimeoutSec') } </span>
              <Input
                type="number"
                min={1}
                max={120}
                step={1}
                bind:value={mapConf.Timeout}
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.Retry') } </span>
              <Input
                type="number"
                min={0}
                max={100}
                step={1}
                bind:value={mapConf.Retry}
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.LogDays') } </span>
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
            <Label class="space-y-2">
              <span> { $_('Config.SNMPMode') } </span>
              <Select
                items={snmpModeList}
                bind:value={mapConf.SnmpMode}
                placeholder={ $_('Config.SelectSnmpMode') }
                size="sm"
              />
            </Label>
            {#if mapConf.SnmpMode == "v1" || mapConf.SnmpMode == "v2c"}
              <Label class="space-y-2">
                <span>SNMP Community</span>
                <Input bind:value={mapConf.Community} placeholder="public" size="sm" />
              </Label>
            {:else}
              <Label class="space-y-2">
                <span>{ $_('Config.SnmpUser') }</span>
                <Input bind:value={mapConf.SnmpUser} placeholder="snmp user" size="sm" />
              </Label>
              <Label class="space-y-2">
                <span>{ $_('Config.SnmpPassword') }</span>
                <Input
                  type="password"
                  bind:value={mapConf.SnmpPassword}
                  placeholder="•••••"
                  size="sm"
                />
              </Label>
            {/if}
          </div>
          <div class="grid gap-4 mb-4 md:grid-cols-3">
            <Checkbox bind:checked={mapConf.EnableSyslogd}>Syslog</Checkbox>
            <Checkbox bind:checked={mapConf.EnableTrapd}>SNMP TRAP</Checkbox>
            <Checkbox bind:checked={mapConf.EnableArpWatch}>ARP Watch</Checkbox>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton shadow color="blue" type="button" on:click={saveMapConf} size="xs">
              <Icon path={icons.mdiContentSave} size={1} />
              { $_('Config.Save') }
            </GradientButton>
            <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
              <Icon path={icons.mdiCancel} size={1} />
              { $_('Config.Cancel') }
            </GradientButton>
          </div>
                </form>
        <!-- <MapConf on:close={close}></MapConf> -->
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBox} size={1} />
          { $_('Config.Notify') }
        </div>
        <form class="flex flex-col space-y-4" action="#">
          {#if showTestError}
            <Alert color="red" dismissable>
              <div class="flex">
                <Icon path={icons.mdiExclamation} size={1} />
                { $_('Config.FailedSendMail') }
              </div>
            </Alert>
          {/if}
          {#if showTestOk}
            <Alert class="flex" color="blue" dismissable>
              <div class="flex">
                <Icon path={icons.mdiCheck} size={1} />
                { $_('Config.SentTestMail') }
              </div>
            </Alert>
          {/if}
          <div class="grid gap-4 md:grid-cols-2">
            <Label class="space-y-2">
              <span>{ $_('Config.MailServer') }</span>
              <Input
                bind:value={notifyConf.MailServer}
                placeholder="host|ip:port"
                required
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.InsecureSkipVerify}>
              { $_('Config.NoCheckCert') }
            </Checkbox>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <Label class="space-y-2">
              <span>{ $_('Config.SmtpUser') }</span>
              <Input bind:value={notifyConf.User} placeholder="smtp user" size="sm" />
            </Label>
            <Label class="space-y-2">
              <span>{ $_('Config.SmtpPassword') }</span>
              <Input
                type="password"
                bind:value={notifyConf.Password}
                placeholder="•••••"
                size="sm"
              />
            </Label>
          </div>
          <div class="grid gap-4 md:grid-cols-2">
            <Label class="space-y-2">
              <span>{ $_('Config.MailFrom') }</span>
              <Input
                bind:value={notifyConf.MailFrom}
                placeholder={ $_('Config.MailFromAddress') }
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span>{ $_('Config.MailTo') }</span>
              <Input
                bind:value={notifyConf.MailTo}
                placeholder={ $_('Config.MailToAddress') }
                size="sm"
              />
            </Label>
          </div>
          <Label class="space-y-2">
            <span> { $_('Config.Subject') } </span>
            <Input bind:value={notifyConf.Subject} size="sm" />
          </Label>
          <div class="grid gap-4 md:grid-cols-4">
            <Label class="space-y-2">
              <span> { $_('Config.NotifyLevel') } </span>
              <Select
                items={notifyLevelList}
                bind:value={notifyConf.Level}
                placeholder={ $_('Config.SelectNotifyLevel') }
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.NotifyIntSec') } </span>
              <Input
                type="number"
                min={60}
                max={3600 * 24}
                step={10}
                bind:value={notifyConf.Interval}
                size="sm"
              />
            </Label>
            <Checkbox bind:checked={notifyConf.Report}>{ $_('Config.MailReport') }</Checkbox>
            <Checkbox bind:checked={notifyConf.NotifyRepair}>{ $_('Config.NotifyRepair') }</Checkbox>
          </div>
          <Label class="space-y-2">
            <span> { $_('Config.ExecCommand') } </span>
            <Input class="w-full" bind:value={notifyConf.ExecCmd} size="sm" />
          </Label>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton shadow color="blue" type="button" on:click={saveNotifyConf} size="xs">
              <Icon path={icons.mdiContentSave} size={1} />
              { $_('Config.Save') }
            </GradientButton>
            <GradientButton shadow type="button" color="red" on:click={testNotifyConf} size="xs">
              <Icon path={icons.mdiEmail} size={1} />
              { $_('Config.Test') }
            </GradientButton>
            <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
              <Icon path={icons.mdiCancel} size={1} />
              { $_('Config.Cancel') }
            </GradientButton>
          </div>
        </form>
      </TabItem>
      <TabItem>
        <div slot="title" class="flex items-center gap-2">
          <Icon path={icons.mdiChartBarStacked} size={1} />
          { $_('Config.AI') }
        </div>
        <form class="flex flex-col space-y-4" action="#">
          <div class="grid gap-4 md:grid-cols-3">
            <Label class="space-y-2">
              <span> { $_('Config.AIHighLevel') } </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.HighThreshold}
                placeholder={ $_('Config.SelectAILevel') }
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.AILevelLow') } </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.LowThreshold}
                placeholder={ $_('Config.SelectAILevel') }
                size="sm"
              />
            </Label>
            <Label class="space-y-2">
              <span> { $_('Config.AIlevelWarn') } </span>
              <Select
                items={aiLevelList}
                bind:value={aiConf.WarnThreshold}
                placeholder={ $_('Config.SelectAILevel') }
                size="sm"
              />
            </Label>
          </div>
          <div class="flex justify-end space-x-2 mr-2">
            <GradientButton shadow color="blue" type="button" on:click={saveAIConf} size="xs">
              <Icon path={icons.mdiContentSave} size={1} />
              { $_('Config.Save') }
            </GradientButton>
            <GradientButton shadow type="button" color="teal" on:click={close} size="xs">
              <Icon path={icons.mdiCancel} size={1} />
              { $_('Config.Cancel') }
            </GradientButton>
          </div>
        </form>
      </TabItem>
    </Tabs>
</Modal>
