import { test } from 'node:test';
import assert from 'node:assert/strict';
import { chipsVisible, shouldSave, CHIP_MAX } from './searches.ts';

const base = { tab: 'provider' as const, query: '', card: false, searches: ['cats'] };

test('chipsVisible: provider tab, empty query, no card, at least one saved search', () => {
  assert.equal(chipsVisible(base), true);
  assert.equal(chipsVisible({ ...base, query: '   ' }), true);
});

test('chipsVisible is false on library tabs, with a query, behind a card or with no searches', () => {
  assert.equal(chipsVisible({ ...base, tab: 'recents' }), false);
  assert.equal(chipsVisible({ ...base, tab: 'favorites' }), false);
  assert.equal(chipsVisible({ ...base, query: 'c' }), false);
  assert.equal(chipsVisible({ ...base, card: true }), false);
  assert.equal(chipsVisible({ ...base, searches: [] }), false);
  assert.equal(chipsVisible({ ...base, searches: undefined }), false);
});

test('shouldSave needs 1.5 s idle, a non-empty first page and a 2+ code point query', () => {
  assert.equal(shouldSave({ sinceKeyMs: 1500, firstPageItems: 1, query: 'cats' }), true);
  assert.equal(shouldSave({ sinceKeyMs: 1499, firstPageItems: 5, query: 'cats' }), false);
  assert.equal(shouldSave({ sinceKeyMs: 2000, firstPageItems: 0, query: 'cats' }), false);
  assert.equal(shouldSave({ sinceKeyMs: 2000, firstPageItems: 3, query: 'c' }), false);
  assert.equal(shouldSave({ sinceKeyMs: 2000, firstPageItems: 3, query: ' 😀😀 ' }), true);
  assert.equal(CHIP_MAX, 8);
});
