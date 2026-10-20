<script lang="ts" module>
  // Set by SettingsApp when settings:reveal arrives before this pane is mounted.
  export const pendingReveal = { updates: false };
</script>

<script lang="ts">
  // General pane (UI-SPEC "General"): every change applies at once, no Save button.
  import { Events } from '@wailsio/runtime';
  import { onMount } from 'svelte';
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
  import { updateErrorText, updateText } from '../lib/update';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  // These follow the snapshot SettingsApp keeps current (settings:changed, and a re-read on every
  // show); a change below overrides one until the next snapshot.
  let theme = $derived(initial.theme);
  let launchAtLogin = $derived(initial.launchAtLogin);
  let actions = $derived({ primaryAction: initial.primaryAction, secondaryAction: initial.secondaryAction });
  let actionError = $state('');
  const gated = $derived(!initial.klipyFileActions); // a build constant: it never changes at run time
  let themeError = $state('');
  let loginError = $state('');
  let axOn = $state<boolean | null>(null); // null until the first read
  let axError = $state('');
  let axTimer: ReturnType<typeof setInterval> | undefined;
  let upd = $state<UpdateStatus>({ state: 'idle', current: '', latest: '', done: 0, total: 0 });
  let updError = $state('');
  let autoCheck = $derived(initial.updateCheck);
  let autoError = $state('');

  function revealUpdates() {
    pendingReveal.updates = false;
    document.getElementById('updates')?.scrollIntoView({ block: 'nearest' });
  }
  const updBusy = $derived(upd.state === 'checking' || upd.state === 'downloading' || upd.state === 'verifying');

  const THEMES = [
    { value: '', label: 'System' },
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
  ];

  // The grant can change outside the app: re-read on open, and every 2 s while it is Off and
  // Settings is visible (the page loads hidden at launch and lives until quit).
  let visible = !document.hidden;
  let alive = true;
  async function readAX() {
    let on: boolean;
    try {
      on = await accessibilityStatus();
    } catch {
      return; // keep the last status; a running poll tries again
    }
    if (!alive) return;
    axOn = on;
    if (on || !visible) stopAX();
    else axTimer ??= setInterval(readAX, 2000);
  }
  function stopAX() {
    clearInterval(axTimer);
    axTimer = undefined;
  }

  onMount(() => {
    readAX();
    updateStatus()
      .then((st) => (upd = st))
      .catch(() => {}); // stays idle; update:status still arrives
    const offUpdate = Events.On('update:status', (e) => (upd = e.data as UpdateStatus));
    const offReveal = Events.On('settings:reveal', revealUpdates);
    if (pendingReveal.updates) revealUpdates();
    const offShow = Events.On('settings:show', () => {
      visible = true;
      readAX();
    });
    const offHide = Events.On('settings:hide', () => {
      visible = false;
      stopAX();
    });
    return () => {
      alive = false;
      offUpdate();
      offReveal();
      offShow();
      offHide();
      stopAX();
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
    } catch (e) {
      updError = updateErrorText(updateError(e)?.kind ?? '', upd);
    }
  }

  async function download() {
    updError = '';
    try {
      upd = await downloadUpdate();
    } catch (e) {
      updError = updateErrorText(updateError(e)?.kind ?? 'update_download_failed', upd);
    }
  }

  async function toggleAutoCheck() {
    const on = autoCheck; // bind:checked already flipped it
    autoError = '';
    try {
      autoCheck = (await setPref('updateCheck', String(on))).updateCheck;
    } catch {
      autoCheck = !on;
      autoError = "Couldn't save this setting. Try again.";
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
    {#if updateText(upd)}
      <p class="update-status {upd.state}" role="status">
        {#if upd.state === 'checking'}<span class="spinner" aria-hidden="true"></span>{/if}
        {#if upd.state === 'current'}<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M3 8.5 6.5 12 13 4.5" /></svg>{/if}
        {updateText(upd)}
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
  <SettingRow
    label=""
    id="update-auto-label"
    error={autoError}
    helper="Gifsmith asks GitHub for the latest release at launch and once a day. Nothing about you or your Mac is sent."
  >
    <label>
      <input type="checkbox" bind:checked={autoCheck} onchange={toggleAutoCheck} />
      Check for updates automatically
    </label>
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
