<script lang="ts">
  import { onMount } from 'svelte';
  import { Events } from '@wailsio/runtime';
  import { ActionService } from '../bindings/gifsmith/internal/services';

  let input: HTMLInputElement;
  let status = $state('');
  let needsAccessibility = $state(false);
  // Paste hides the panel before ⌘V; a copied-only fallback message must survive to the next show.
  let keepStatus = false;

  onMount(() => {
    input.focus();
    Events.On('panel:show', () => {
      if (!keepStatus) status = '';
      keepStatus = false;
      needsAccessibility = false;
      input.focus();
      input.select();
    });
    Events.On('panel:status', (e) => {
      status = String(e.data);
    });
  });

  async function onKeydown(e: KeyboardEvent) {
    // Enter that commits an IME composition (WebKit: keyCode 229) is not a submit.
    if (e.key !== 'Enter' || e.isComposing || e.keyCode === 229 || e.repeat) return;
    e.preventDefault();
    try {
      if (e.metaKey) {
        const r = await ActionService.Paste();
        status = r.message;
        needsAccessibility = r.needsAccessibility;
        keepStatus = r.copied && !r.pasted && !r.needsAccessibility;
      } else {
        status = (await ActionService.Copy()).message;
      }
    } catch (err) {
      status = String(err);
    }
  }

  async function enableAutoPaste() {
    try {
      await ActionService.RequestAccessibility();
      status =
        'After turning Gifsmith on, press ⌘Enter again. If it still says it is not allowed, quit Gifsmith from the menu bar and open it again.';
      needsAccessibility = false;
    } catch (err) {
      status = String(err);
    }
    input.focus();
  }
</script>

<input bind:this={input} onkeydown={onKeydown} placeholder="Search GIFs" autocomplete="off" />
<img src="/sample.gif" width="120" height="120" alt="sample GIF" />
<p>{status}</p>
{#if needsAccessibility}
  <button onclick={enableAutoPaste}>Enable automatic paste…</button>
{/if}
