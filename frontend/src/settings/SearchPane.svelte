<script lang="ts">
  // Search pane (UI-SPEC "Search"): each change saves at once; Go validates and flushes the cache.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { setPref, type SettingsSnapshot } from '../lib/api';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  let rating = $state(untrack(() => initial.giphyRating));
  let ratingError = $state('');

  const RATINGS = [
    { value: 'g', label: 'G' },
    { value: 'pg', label: 'PG' },
    { value: 'pg-13', label: 'PG-13' },
    { value: 'r', label: 'R' },
  ];

  onMount(() =>
    Events.On('settings:changed', (e) => {
      rating = (e.data as SettingsSnapshot).giphyRating;
    }),
  );

  async function chooseRating(v: string) {
    const prev = rating;
    rating = v;
    ratingError = '';
    try {
      rating = (await setPref('giphyRating', v)).giphyRating;
    } catch {
      rating = prev;
      ratingError = "Couldn't save this setting. Try again.";
    }
  }
</script>

<div class="settings-section">
  <SettingRow label="Content rating" id="rating-label" error={ratingError} helper="Rating and language apply to GIPHY and Clips.">
    <select aria-labelledby="rating-label" value={rating} onchange={(e) => chooseRating(e.currentTarget.value)}>
      {#each RATINGS as r (r.value)}
        <option value={r.value}>{r.label}</option>
      {/each}
    </select>
  </SettingRow>
</div>

<style>
  select {
    min-width: 160px;
    max-width: 272px;
    font: var(--body);
  }
</style>
