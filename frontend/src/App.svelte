<script lang="ts">
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { ActionService } from '../bindings/gifsmith/internal/services';

  let input: HTMLInputElement;
  let status = $state('');

  onMount(() => {
    input.focus();
    Events.On('panel:show', () => {
      status = '';
      input.focus();
      input.select();
    });
    Events.On('panel:status', (e) => {
      status = String(e.data);
    });
  });

  async function onKeydown(e: KeyboardEvent) {
    if (e.key !== 'Enter' || e.metaKey) return;
    e.preventDefault();
    try {
      status = (await ActionService.Copy()).message;
    } catch (err) {
      status = String(err);
    }
  }
</script>

<input bind:this={input} onkeydown={onKeydown} placeholder="Search GIFs" autocomplete="off" />
<img src="/sample.gif" width="120" height="120" alt="sample GIF" />
<p>{status}</p>
