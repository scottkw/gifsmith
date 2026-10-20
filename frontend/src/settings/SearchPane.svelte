<script lang="ts">
  // Search pane (UI-SPEC "Search"): each change saves at once; Go validates it and flushes the
  // affected providers' cached pages, and the panel refetches a tab the next time it is shown.
  import { Events } from '@wailsio/runtime';
  import { onMount, untrack } from 'svelte';
  import { choices, setPref, type SettingsSnapshot } from '../lib/api';
  import { countryOptions, languageOptions, ratingLabel, type Option } from '../lib/prefs';
  import SettingRow from './SettingRow.svelte';

  let { initial }: { initial: SettingsSnapshot } = $props();

  type Name = 'pageSize' | 'giphyRating' | 'giphyLang' | 'klipyCountry' | 'gridSearch' | 'gridTrending';
  const valuesOf = (s: SettingsSnapshot): Record<Name, string> => ({
    gridSearch: s.gridSearch,
    gridTrending: s.gridTrending,
    pageSize: String(s.pageSize),
    giphyRating: s.giphyRating,
    giphyLang: s.giphyLang,
    klipyCountry: s.klipyCountry,
  });
  let values = $state(untrack(() => valuesOf(initial)));
  let errors = $state<Partial<Record<Name, string>>>({});
  let sizes = $state<Option[]>([]);
  let ratings = $state<Option[]>([]);
  let langs = $state<Option[]>([]);
  let countries = $state<Option[]>([]);

  onMount(() => {
    choices().then((c) => {
      sizes = (c.pageSizes ?? []).map((n) => ({ value: String(n), label: String(n) }));
      ratings = (c.ratings ?? []).map((r) => ({ value: r, label: ratingLabel(r) }));
      langs = languageOptions(c.languages ?? [], c.systemLanguage);
      countries = countryOptions(c.countries ?? [], c.systemCountry);
    });
    return Events.On('settings:changed', (e) => (values = valuesOf(e.data as SettingsSnapshot)));
  });

  async function choose(name: Name, v: string) {
    const prev = values[name];
    values[name] = v;
    errors[name] = '';
    try {
      values = valuesOf(await setPref(name, v));
    } catch {
      values[name] = prev;
      errors[name] = "Couldn't save this setting. Try again.";
    }
  }
</script>

{#snippet sizes3(name: Name, labelId: string)}
  <div class="radios" role="radiogroup" aria-labelledby={labelId}>
    {#each ['small', 'medium', 'large'] as v (v)}
      <label>
        <input type="radio" name={name} value={v} checked={values[name] === v} onchange={() => choose(name, v)} />
        {v[0].toUpperCase() + v.slice(1)}
      </label>
    {/each}
  </div>
{/snippet}

{#snippet picker(name: Name, labelId: string, opts: Option[])}
  <!-- selected per option: the options arrive after mount, so a select-level value would not stick -->
  <select aria-labelledby={labelId} onchange={(e) => choose(name, e.currentTarget.value)}>
    {#each opts as o (o.value)}
      <option value={o.value} selected={o.value === values[name]}>{o.label}</option>
    {/each}
  </select>
{/snippet}

<div class="settings-section">
  <SettingRow label="Search results" id="gridsearch-label" error={errors.gridSearch}>
    {@render sizes3('gridSearch', 'gridsearch-label')}
  </SettingRow>
  <SettingRow label="Trending" id="gridtrending-label" error={errors.gridTrending} helper="Small fits 5 GIFs per row, Medium 4, Large 3.">
    {@render sizes3('gridTrending', 'gridtrending-label')}
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow
    label="GIFs per page"
    id="pagesize-label"
    error={errors.pageSize}
    helper="Bigger pages use fewer requests. Free GIPHY and Klipy keys allow 100 per hour."
  >
    {@render picker('pageSize', 'pagesize-label', sizes)}
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow label="Content rating" id="rating-label" error={errors.giphyRating}>
    {@render picker('giphyRating', 'rating-label', ratings)}
  </SettingRow>
  <SettingRow label="Language" id="lang-label" error={errors.giphyLang} helper="Rating and language apply to GIPHY and Clips.">
    {@render picker('giphyLang', 'lang-label', langs)}
  </SettingRow>
</div>

<div class="settings-section">
  <SettingRow label="Country" id="country-label" error={errors.klipyCountry}>
    {@render picker('klipyCountry', 'country-label', countries)}
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
  select {
    min-width: 160px;
    max-width: 272px;
    font: var(--body);
  }
</style>
