<script lang="ts">
  // Downloads pane (UI-SPEC "Downloads"): the folder comes only from the native picker, through Go,
  // which validates it before saving; the webview never sends a path.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { chooseDownloadDir, resetDownloadDir, setPref, settingsError, type SettingsSnapshot } from '../lib/api';
  import { folderErrorText } from '../lib/errors';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  let snap = $state(untrack(() => initial));
  let folderError = $state(''); // D5–D7 from the last Choose…; the old folder stays saved
  let hideError = $state('');

  onMount(() => Events.On('settings:changed', (e) => (snap = e.data as SettingsSnapshot)));

  // shortcut: assumes the home folder is /Users/<name>; a home elsewhere shows the full path.
  const tilde = (p: string) => p.replace(/^\/Users\/[^/]+(?=\/|$)/, '~');

  async function change(f: () => Promise<SettingsSnapshot>) {
    folderError = '';
    try {
      snap = await f();
    } catch (e) {
      const c = settingsError(e);
      folderError = folderErrorText(c?.kind ?? '', c?.folder ?? '');
    }
  }

  async function toggleHide(box: HTMLInputElement) {
    hideError = '';
    try {
      snap = await setPref('hideFilenames', String(box.checked));
    } catch {
      box.checked = snap.hideFilenames;
      hideError = "Couldn't save this setting. Try again.";
    }
  }
</script>

<div class="settings-section">
  <SettingRow
    label="Save to"
    id="saveto-label"
    error={folderError || (snap.downloadDirMissing ? folderErrorText('folder_missing', snap.downloadDirName) : '')}
  >
    <div class="folder">
      <div class="display" role="group" aria-label="Download folder: {snap.downloadDirName}">
        <span class="name">{snap.downloadDirName}</span>
        <span class="path" title={snap.downloadDir}>{tilde(snap.downloadDir)}</span>
      </div>
      <div class="buttons">
        <button class="btn-secondary" aria-label="Choose download folder" onclick={() => change(chooseDownloadDir)}>Choose…</button>
        <button
          class="btn-secondary"
          aria-label="Reset download folder to Downloads"
          disabled={snap.downloadDirDefault}
          onclick={() => change(resetDownloadDir)}>Reset to Downloads</button
        >
      </div>
    </div>
  </SettingRow>
  <SettingRow label="File names" id="filenames-label" error={hideError} helper={"Name downloads gifsmith-{id}.gif instead of using the GIF's title."}>
    <label class="check">
      <input type="checkbox" checked={snap.hideFilenames} onchange={(e) => toggleHide(e.currentTarget)} />
      Hide original file names
    </label>
  </SettingRow>
</div>

<style>
  .folder {
    width: 100%;
    min-width: 0;
  }
  .display {
    display: flex;
    flex-direction: column;
    min-height: 32px;
    justify-content: center;
  }
  .name {
    font: var(--body);
    color: var(--text);
  }
  .path {
    font: var(--label);
    color: var(--text-secondary);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .check {
    display: flex;
    align-items: center;
    gap: 4px;
    font: var(--body);
    color: var(--text);
  }
  .check input {
    margin: 0;
  }
  .buttons {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }
</style>
