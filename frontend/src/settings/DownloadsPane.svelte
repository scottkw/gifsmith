<script lang="ts">
  // Downloads pane (UI-SPEC "Downloads"): the folder comes only from the native picker, through Go,
  // which validates it before saving; the webview never sends a path.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { chooseDownloadDir, resetDownloadDir, type SettingsSnapshot } from '../lib/api';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  let snap = $state(untrack(() => initial));
  let folderError = $state('');

  onMount(() => Events.On('settings:changed', (e) => (snap = e.data as SettingsSnapshot)));

  // shortcut: assumes the home folder is /Users/<name>; a home elsewhere shows the full path.
  const tilde = (p: string) => p.replace(/^\/Users\/[^/]+(?=\/|$)/, '~');

  async function change(f: () => Promise<SettingsSnapshot>) {
    folderError = '';
    try {
      snap = await f();
    } catch {
      folderError = "Couldn't save this setting. Try again.";
    }
  }
</script>

<div class="settings-section">
  <SettingRow label="Save to" id="saveto-label" error={folderError}>
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
  .buttons {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }
</style>
