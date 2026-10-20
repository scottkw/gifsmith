import { test } from 'node:test';
import assert from 'node:assert/strict';
import { colsFor, countryOptions, DEFAULT_ACTIONS, skeletonCount, tileHeight, effectiveAction, languageOptions, ratingLabel, shortLabel } from './prefs.ts';
import type { Gif } from './api';

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

const item = (o: Partial<Gif> = {}): Gif => ({ id: '1', provider: 'giphy', title: 't', previewUrl: 'p', downloadUrl: 'd', isClip: false, ...o }) as Gif;
const prefs = (primaryAction: string, secondaryAction = 'paste', klipyFileActions = false) => ({ primaryAction, secondaryAction, klipyFileActions });

test('effectiveAction maps file actions on gated Klipy items to Copy Link / Paste Link', () => {
  const klipy = item({ provider: 'klipy' });
  assert.equal(effectiveAction(item(), 'primary', prefs('copy')), 'copy');
  assert.equal(effectiveAction(klipy, 'primary', prefs('copy')), 'copyLink');
  assert.equal(effectiveAction(klipy, 'secondary', prefs('copy')), 'pasteLink');
  assert.equal(effectiveAction(klipy, 'primary', prefs('download')), 'copyLink');
  assert.equal(effectiveAction(klipy, 'primary', prefs('copyMarkdown')), 'copyMarkdown');
  assert.equal(effectiveAction(klipy, 'primary', prefs('copy', 'paste', true)), 'copy');
  assert.equal(effectiveAction(item({ provider: 'clips', isClip: true }), 'primary', prefs('copy')), 'copy');
  assert.equal(effectiveAction(undefined, 'secondary', prefs('copy', 'pasteMarkdown')), 'pasteMarkdown');
});

test('DEFAULT_ACTIONS lists the seven choices in Settings order; shortLabel names them for the footer', () => {
  assert.deepEqual(
    DEFAULT_ACTIONS.map((a) => a.label),
    ['Copy GIF', 'Paste GIF', 'Copy Link', 'Paste Link', 'Copy Markdown', 'Paste Markdown', 'Download GIF'],
  );
  assert.deepEqual(
    DEFAULT_ACTIONS.map((a) => shortLabel(a.id)),
    ['Copy', 'Paste', 'Copy Link', 'Paste Link', 'Copy Markdown', 'Paste Markdown', 'Download'],
  );
});

test('colsFor, tileHeight and skeletonCount per grid size; unknown is Medium', () => {
  assert.deepEqual(['small', 'medium', 'large', ''].map(colsFor), [5, 4, 3, 4]);
  assert.deepEqual(['small', 'medium', 'large', ''].map(tileHeight), [112, 144, 192, 144]);
  assert.deepEqual(['small', 'medium', 'large', ''].map(skeletonCount), [10, 8, 6, 8]);
});
