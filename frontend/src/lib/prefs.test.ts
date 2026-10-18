import { test } from 'node:test';
import assert from 'node:assert/strict';
import { countryOptions, languageOptions, ratingLabel } from './prefs.ts';

test('languageOptions puts Match System first, then languages by English name', () => {
  assert.deepEqual(languageOptions(['de', 'en'], 'en'), [
    { value: '', label: 'Match System (English)' },
    { value: 'en', label: 'English' },
    { value: 'de', label: 'German' },
  ]);
});

test('languageOptions names GIPHY Chinese codes by script', () => {
  const labels = languageOptions(['zh-TW', 'zh-CN'], 'zh-TW').map((o) => o.label);
  assert.deepEqual(labels, ['Match System (Traditional Chinese)', 'Simplified Chinese', 'Traditional Chinese']);
});

test('equal names are ordered by code', () => {
  assert.deepEqual(
    languageOptions(['iw', 'he'], 'en').map((o) => o.value),
    ['', 'he', 'iw'],
  );
});

test('countryOptions sorts by English region name with Match System first', () => {
  assert.deepEqual(countryOptions(['us', 'de', 'at'], 'us'), [
    { value: '', label: 'Match System (United States)' },
    { value: 'at', label: 'Austria' },
    { value: 'de', label: 'Germany' },
    { value: 'us', label: 'United States' },
  ]);
});

test('ratingLabel upper-cases the GIPHY rating', () => {
  assert.deepEqual(['g', 'pg', 'pg-13', 'r'].map(ratingLabel), ['G', 'PG', 'PG-13', 'R']);
});
