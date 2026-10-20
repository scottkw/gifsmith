<script lang="ts">
  // About (UI-SPEC "About (RELS-03)"): identity, links, credits to the original extension and the providers.
  import { onMount } from 'svelte';
  import { openExternal, type SettingsSnapshot } from '../lib/api';
  import { ABOUT_URLS } from '../lib/providers';
  import { onTheme, resolvedTheme } from '../lib/theme';

  let { initial }: { initial: SettingsSnapshot } = $props();

  // The GIPHY mark follows the resolved theme (a manual Light/Dark included), like the footer.
  let theme = $state(resolvedTheme());
  onMount(() => onTheme((t) => (theme = t)));

  const links = [
    { label: 'Website ↗', url: ABOUT_URLS.website },
    { label: 'Report an Issue ↗', url: ABOUT_URLS.issues },
    { label: 'View License ↗', url: ABOUT_URLS.license },
  ];

  function open(e: MouseEvent, url: string) {
    e.preventDefault();
    openExternal(url).catch(() => {});
  }
</script>

<div class="settings-section">
  <img class="icon" src="/appicon.png" alt="" width="64" height="64" />
  <h2>Gifsmith</h2>
  <p class="version">Version {initial.version}</p>
  <p class="tagline">Search GIFs from the menu bar and copy or paste them into any app.</p>
  <div class="links">
    {#each links as l (l.url)}
      <a href={l.url} onclick={(e) => open(e, l.url)}>{l.label}</a>
    {/each}
  </div>
</div>

<div class="settings-section">
  <h3>Credits</h3>
  <p class="body">
    Gifsmith is a port of the GIF Search extension for Raycast by Joseph Schmitt and contributors: tm.wrnr, j3lte,
    samuelkraft, ridemountainpig, spookywy, seanparkross, YourMCGeek, thomas, tim_gailey, clins1994, 0xdhrv, Olli0103
    and alexi.build. Used under the MIT License, © 2021 Raycast.
  </p>
  <a class="after" href={ABOUT_URLS.extension} onclick={(e) => open(e, ABOUT_URLS.extension)}>GIF Search extension ↗</a>
</div>

<div class="settings-section">
  <h3>GIF providers</h3>
  <!-- Official GIPHY mark, unmodified (no filter, opacity or crop), not a link. -->
  <img class="mark" src="/attribution/giphy-powered-{theme}.png" alt="Powered by GIPHY" height="16" />
  <p class="small after">Search by GIPHY and KLIPY. GIFs also from The Finer Gifs Club.</p>
  <p class="small copyright">© 2026 Ken Scott. MIT License.</p>
</div>

<style>
  p,
  h2,
  h3 {
    margin: 0;
  }
  .icon {
    display: block;
  }
  h2 {
    margin-top: 16px;
    font: var(--heading);
    color: var(--text);
  }
  .version {
    margin-top: 4px;
    font: var(--label);
    color: var(--text-secondary);
    user-select: text;
    -webkit-user-select: text;
  }
  .tagline {
    margin-top: 8px;
    font: var(--body);
    color: var(--text-secondary);
  }
  .links {
    display: flex;
    gap: 16px;
    margin-top: 16px;
  }
  a {
    font: var(--body);
    color: var(--accent);
    text-decoration: none;
    white-space: nowrap;
  }
  a:hover {
    text-decoration: underline;
  }
  h3 {
    margin-bottom: 8px;
    font: var(--label);
    color: var(--text-secondary);
  }
  .body {
    font: var(--body);
    color: var(--text);
  }
  .after {
    display: inline-block;
    margin-top: 8px;
  }
  .mark {
    display: block;
  }
  .small {
    font: var(--label);
    color: var(--text-secondary);
  }
  .copyright {
    margin-top: 16px;
  }
</style>
