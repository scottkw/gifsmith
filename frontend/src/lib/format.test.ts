import { test } from 'node:test';
import assert from 'node:assert/strict';
import { dimensions, fileSize, linkText, providerLine, tagsSummary } from './format.ts';
import type { Gif } from './api';

const gif = (o: Partial<Gif> = {}): Gif =>
  ({ id: '1', provider: 'giphy', title: 't', previewUrl: 'p', downloadUrl: 'd', isClip: false, ...o }) as Gif;

test('fileSize rounds to KB (at least 1) below 1 MB and to one decimal MB above', () => {
  assert.equal(fileSize(5678), '6 KB');
  assert.equal(fileSize(1234), '1 KB');
  assert.equal(fileSize(400), '1 KB');
  assert.equal(fileSize(2_500_000), '2.5 MB');
  assert.equal(fileSize(1_000_000), '1.0 MB');
  assert.equal(fileSize(0), '');
  assert.equal(fileSize(undefined), '');
});

test('dimensions needs both sides', () => {
  assert.equal(dimensions(480, 480), '480 × 480');
  assert.equal(dimensions(0, 480), '');
  assert.equal(dimensions(480, undefined), '');
});

test('linkText drops the https:// prefix and keeps host, path and query', () => {
  assert.equal(linkText('https://giphy.com/gifs/moment-giphy-1'), 'giphy.com/gifs/moment-giphy-1');
  assert.equal(linkText('https://example.org/a?b=c'), 'example.org/a?b=c');
  assert.equal(linkText(''), '');
});

test('tagsSummary lists the first 12 and counts the rest', () => {
  const tags = Array.from({ length: 14 }, (_, i) => `t${i + 1}`);
  assert.equal(tagsSummary(tags), 't1, t2, t3, t4, t5, t6, t7, t8, t9, t10, t11, t12 +2 more');
  assert.equal(tagsSummary(['funny', 'cat']), 'funny, cat');
  assert.equal(tagsSummary([]), '');
  assert.equal(tagsSummary(null), '');
});

test('providerLine names the provider, then MP4 for Clips, then In Favorites', () => {
  assert.equal(providerLine(gif({ provider: 'clips' as Gif['provider'], isClip: true }), true), 'GIPHY Clips · MP4 · In Favorites');
  assert.equal(providerLine(gif({ provider: 'finergifs' as Gif['provider'] }), false), 'The Finer Gifs Club');
  assert.equal(providerLine(gif(), true), 'GIPHY · In Favorites');
});
