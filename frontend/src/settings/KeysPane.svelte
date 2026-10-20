<script lang="ts">
  // Keys pane (UI-SPEC "Keys"): GIPHY then Klipy. Status comes from Go (last 4 characters at most)
  // and is re-read on open, on providers:changed (save, remove, a rejected key) and on settings:show.
  import { Events } from '@wailsio/runtime';
  import { onMount } from 'svelte';
  import { keyStatus, type KeyStatus, type SettingsSnapshot } from '../lib/api';
  import { GIPHY_KEY_URL, KLIPY_KEY_URL } from '../lib/providers';
  import KeyRow from './KeyRow.svelte';

  const CARDS = [
    {
      provider: 'giphy',
      label: 'GIPHY',
      description: 'Used for GIPHY and GIPHY Clips. Free keys allow 100 requests per hour.',
      consequence: 'GIPHY and Clips search stop working until you add a key again.',
      keyUrl: GIPHY_KEY_URL,
    },
    {
      provider: 'klipy',
      label: 'Klipy',
      description: 'Used for Klipy. Free keys allow 100 requests per hour.',
      consequence: 'Klipy search stops working until you add a key again.',
      keyUrl: KLIPY_KEY_URL,
    },
  ] as const;

  // Every pane gets the snapshot; key state is not in it (no key ever crosses in a snapshot).
  let {}: { initial: SettingsSnapshot } = $props();

  // null until the first read: a card never shows Not set / Add Key… for a key that is saved.
  let statuses = $state<Record<string, KeyStatus> | null>(null);
  let loadError = $state('');

  async function load() {
    try {
      statuses = Object.fromEntries((await keyStatus()).map((s) => [s.provider, s]));
      loadError = '';
    } catch {
      loadError = "Couldn't read your keys from the macOS Keychain. Open Keys again to retry.";
    }
  }

  onMount(() => {
    load();
    const offChanged = Events.On('providers:changed', load);
    const offShow = Events.On('settings:show', load);
    return () => {
      offChanged();
      offShow();
    };
  });
</script>

{#if loadError}
  <p class="load-error" role="alert">{loadError}</p>
{/if}
<div class="cards">
  {#each CARDS as c (c.provider)}
    <KeyRow {...c} status={statuses?.[c.provider] ?? null} />
  {/each}
</div>

<style>
  .load-error {
    margin: 0 0 16px;
    font: var(--label);
    color: var(--destructive);
  }
  .cards {
    display: flex;
    flex-direction: column;
    gap: 16px;
  }
</style>
