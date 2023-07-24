<script lang="ts">
  import { Alert } from 'flowbite-svelte';
	import Icon from 'mdi-svelte';
  import { mdiAccount } from '@mdi/js';
 </script>

<div class="p-8">
	<Alert>
		<span class="mdi mdi-access-point-check text-3xl"></span>
		<span class="font-medium">Info alert!</span> Change a few things up and try submitting again.
	</Alert>
</div>

<Icon path={mdiAccount} color="black" />
