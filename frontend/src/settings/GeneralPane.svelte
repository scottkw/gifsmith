<script lang="ts">
  // General pane (UI-SPEC "General"): every change applies at once, no Save button.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import {
    accessibilityStatus,
    checkForUpdates,
    downloadUpdate,
    requestAccessibility,
    setLaunchAtLogin,
    setPref,
    settingsError,
    updateError,
    updateStatus,
    type SettingsSnapshot,
    type UpdateStatus,
  } from '../lib/api';
  import { DEFAULT_ACTIONS } from '../lib/prefs';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  // settings:changed keeps both current (a change from the other window, or this one).
  let theme = $state(untrack(() => initial.theme));
  let launchAtLogin = $state(untrack(() => initial.launchAtLogin));
  let actions = $state(untrack(() => ({ primaryAction: initial.primaryAction, secondaryAction: initial.secondaryAction })));
  let actionError = $state('');
  const gated = $derived(!initial.klipyFileActions); // a build constant: it never changes at run time
  let themeError = $state('');
  let loginError = $state('');
  let axOn = $state<boolean | null>(null); // null until the first read
  let axError = $state('');
  let axTimer: ReturnType<typeof setInterval> | undefined;
  let upd = $state<UpdateStatus>({ state: 'idle', current: '', latest: '', done: 0, total: 0 });
  let updError = $state('');
  const updBusy = $derived(upd.state === 'checking' || upd.state === 'downloading' || upd.state === 'verifying');

  function updText(st: UpdateStatus): string {
    switch (st.state) {
      case 'checking':
        return 'Checking for updates…';
      case 'current':
        return `Gifsmith ${st.current} is up to date.`;
      case 'available':
        return `Gifsmith ${st.latest} is available. You have ${st.current}.`;
      case 'downloading':
        return `Downloading Gifsmith ${st.latest}…` + (st.total > 0 ? ` ${Math.floor((st.done * 100) / st.total)}%` : '');
      case 'verifying':
        return "Checking the download's signature…";
      case 'opened':
        return `The Gifsmith ${st.latest} disk image is open. Quit Gifsmith, then drag the new Gifsmith to Applications and choose Replace.`;
    }
    return '';
  }

  const UPDATE_ERRORS: Record<string, (st: UpdateStatus) => string> = {
    update_download_failed: (st) => `Couldn't download Gifsmith ${st.latest}. Check your connection, then try again.`,
    update_verify_failed: () => "The download didn't pass the signature check, so Gifsmith deleted it. Nothing was opened.",
    update_open_failed: () => "Gifsmith couldn't open the downloaded disk image. Try again.",
  };

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
    updateStatus().then((st) => (upd = st));
    const offUpdate = Events.On('update:status', (e) => (upd = e.data as UpdateStatus));
    const offShow = Events.On('settings:show', readAX);
    const offChanged = Events.On('settings:changed', (e) => {
      const s = e.data as SettingsSnapshot;
      theme = s.theme;
      launchAtLogin = s.launchAtLogin;
      actions = { primaryAction: s.primaryAction, secondaryAction: s.secondaryAction };
    });
    return () => {
      offUpdate();
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

  // Go swaps the two when one key takes the other's action, so read both back.
  async function chooseAction(name: 'primaryAction' | 'secondaryAction', box: HTMLSelectElement) {
    actionError = '';
    try {
      const s = await setPref(name, box.value);
      actions = { primaryAction: s.primaryAction, secondaryAction: s.secondaryAction };
    } catch {
      box.value = actions[name];
      actionError = "Couldn't save this setting. Try again.";
    }
  }

  async function check() {
    updError = '';
    try {
      upd = await checkForUpdates();
    } catch {
      updError = "Couldn't check for updates. Check your connection, then try again.";
    }
  }

  async function download() {
    updError = '';
    try {
      upd = await downloadUpdate();
    } catch (e) {
      const text = UPDATE_ERRORS[updateError(e)?.kind ?? ''] ?? UPDATE_ERRORS.update_download_failed;
      updError = text(upd);
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

<div class="settings-section">
  {#snippet actionPicker(name: 'primaryAction' | 'secondaryAction', labelId: string)}
    <select aria-labelledby={labelId} onchange={(e) => chooseAction(name, e.currentTarget)}>
      {#each DEFAULT_ACTIONS as a (a.id)}
        <option value={a.id} selected={a.id === actions[name]}>{a.label}</option>
      {/each}
    </select>
  {/snippet}
  <SettingRow label="Primary action (↵)" id="primary-label">
    {@render actionPicker('primaryAction', 'primary-label')}
  </SettingRow>
  <SettingRow
    label="Secondary action (⌘↵)"
    id="secondary-label"
    error={actionError}
    helper={gated ? 'Klipy GIFs always copy and paste as links.' : ''}
  >
    {@render actionPicker('secondaryAction', 'secondary-label')}
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow label="Updates" id="updates" error={updError}>
    {#if updText(upd)}
      <p class="update-status {upd.state}" role="status">
        {#if upd.state === 'checking'}<span class="spinner" aria-hidden="true"></span>{/if}
        {#if upd.state === 'current'}<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>{/if}
        {updText(upd)}
      </p>
    {/if}
    {#if upd.state === 'downloading'}
      <progress class="update-progress" max={upd.total || undefined} value={upd.done}></progress>
    {/if}
    {#if upd.state === 'available'}
      <button class="btn-primary" disabled={updBusy} onclick={download}>Download and Open…</button>
    {/if}
    <button class="btn-secondary" disabled={updBusy} onclick={check}>Check for Updates…</button>
  </SettingRow>
</div>

<style>
  .update-status {
    flex-basis: 100%;
    display: flex;
    align-items: center;
    gap: 4px;
    margin: 0;
    font: var(--body);
    color: var(--text);
  }
  .update-progress {
    flex-basis: 100%;
    width: 272px;
    margin: -4px 0 0;
  }
  .update-status.checking,
  .update-status.downloading,
  .update-status.verifying {
    color: var(--text-secondary);
  }
  .update-status.current {
    color: var(--success);
  }
  .update-status .spinner {
    width: 12px;
    height: 12px;
  }
  .update-status svg {
    width: 12px;
    height: 12px;
    flex: none;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linecap: round;
    stroke-linejoin: round;
  }
  select {
    min-width: 160px;
    max-width: 272px;
    font: var(--body);
  }
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
