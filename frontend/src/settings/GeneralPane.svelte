<script lang="ts">
  // General pane (UI-SPEC "General"): every change applies at once, no Save button.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import {
    accessibilityStatus,
    requestAccessibility,
    setLaunchAtLogin,
    setPref,
    settingsError,
    type SettingsSnapshot,
  } from '../lib/api';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  // settings:changed keeps both current (a change from the other window, or this one).
  let theme = $state(untrack(() => initial.theme));
  let launchAtLogin = $state(untrack(() => initial.launchAtLogin));
  let themeError = $state('');
  let loginError = $state('');
  let axOn = $state<boolean | null>(null); // null until the first read
  let axError = $state('');
  let axTimer: ReturnType<typeof setInterval> | undefined;

  const THEMES = [
    { value: '', label: 'System' },
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
  ];

  // The grant can change outside the app: re-read on open and every 2 s while it is Off.
  async function readAX() {
    axOn = await accessibilityStatus();
    if (axOn) {
      clearInterval(axTimer);
      axTimer = undefined;
    } else axTimer ??= setInterval(readAX, 2000);
  }

  onMount(() => {
    readAX();
    const offShow = Events.On('settings:show', readAX);
    const offChanged = Events.On('settings:changed', (e) => {
      const s = e.data as SettingsSnapshot;
      theme = s.theme;
      launchAtLogin = s.launchAtLogin;
    });
    return () => {
      offShow();
      offChanged();
      clearInterval(axTimer);
    };
  });

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

  async function toggleLogin() {
    const on = launchAtLogin; // bind:checked already flipped it
    loginError = '';
    try {
      launchAtLogin = (await setLaunchAtLogin(on)).launchAtLogin;
    } catch (e) {
      launchAtLogin = !on;
      loginError =
        settingsError(e)?.kind === 'not_in_applications'
          ? 'Move Gifsmith to the Applications folder, then try again.'
          : "Couldn't change Open at login. Try again.";
    }
  }

  async function grant() {
    axError = '';
    try {
      if (await requestAccessibility()) axOn = true;
    } catch {
      axError = "Couldn't ask macOS for Accessibility access. Try again.";
    }
  }
</script>

<div class="settings-section">
  <SettingRow label="Theme" id="theme-label" error={themeError}>
    <div class="radios" role="radiogroup" aria-labelledby="theme-label">
      {#each THEMES as t (t.value)}
        <label>
          <input type="radio" name="theme" value={t.value} checked={theme === t.value} onchange={() => chooseTheme(t.value)} />
          {t.label}
        </label>
      {/each}
    </div>
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow label="Startup" id="startup-label" error={loginError}>
    <label>
      <input type="checkbox" bind:checked={launchAtLogin} onchange={toggleLogin} />
      Open Gifsmith at login
    </label>
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow
    label="Automatic paste"
    id="ax-label"
    error={axError}
    helper={axOn === null
      ? ''
      : axOn
        ? 'Paste actions put the GIF straight into the app you were using.'
        : 'Paste actions need Accessibility access. Copying always works without it.'}
  >
    {#if axOn}
      <span class="on">
        <svg viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>
        On
      </span>
    {:else if axOn === false}
      <span class="off">Off</span>
      <button class="btn-primary" aria-label="Grant Accessibility access" onclick={grant}>Grant Access…</button>
    {/if}
  </SettingRow>
</div>

<style>
  .radios {
    display: flex;
    gap: 16px;
  }
  label {
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--body);
    color: var(--text);
  }
  input {
    margin: 0;
  }
  .on,
  .off {
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--body);
  }
  .on {
    color: var(--success);
  }
  .off {
    color: var(--text-secondary);
  }
  .on svg {
    width: 12px;
    height: 12px;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
</style>
