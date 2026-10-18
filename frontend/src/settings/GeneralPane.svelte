<script lang="ts">
  // General pane (UI-SPEC "General"): every change applies at once, no Save button.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { setPref, type SettingsSnapshot } from '../lib/api';

  let { initial }: { initial: SettingsSnapshot } = $props();

  let theme = $state(untrack(() => initial.theme)); // settings:changed keeps it current
  let themeError = $state('');

  const THEMES = [
    { value: '', label: 'System' },
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
  ];

  onMount(() => Events.On('settings:changed', (e) => (theme = (e.data as SettingsSnapshot).theme)));

  async function chooseTheme(v: string) {
    const prev = theme;
    theme = v;
    themeError = '';
    try {
      theme = (await setPref('theme', v)).theme;
    } catch {
      theme = prev;
      themeError = "Couldn't save this setting. Try again.";
    }
  }
</script>

<h1>General</h1>
<div class="row">
  <span id="theme-label">Theme</span>
  <div role="radiogroup" aria-labelledby="theme-label">
    {#each THEMES as t (t.value)}
      <label>
        <input type="radio" name="theme" value={t.value} checked={theme === t.value} onchange={() => chooseTheme(t.value)} />
        {t.label}
      </label>
    {/each}
  </div>
</div>
{#if themeError}<p role="alert">{themeError}</p>{/if}
