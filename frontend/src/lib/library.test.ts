import { test } from 'node:test';
import assert from 'node:assert/strict';
import { flatItems, rows, sections } from './library.ts';
import type { Gif } from './api';

const gif = (provider: string, id: string, title = `Moment ${id}`): Gif =>
  ({ provider, id, title, previewUrl: 'p', downloadUrl: 'd', isClip: false }) as Gif;
const a = gif('giphy', 'a', 'Office party');
const b = gif('finergifs', 'b', 'The OFFICE');
const c = gif('giphy', 'c', 'Cats');

test('sections group by provider in slot order, keeping input order within a section', () => {
  const s = sections([a, b, c], '');
  assert.deepEqual(
    s.map((x) => [x.provider, x.label, x.items.map((g) => g.id), x.start]),
    [
      ['giphy', 'GIPHY', ['a', 'c'], 0],
      ['finergifs', 'Finer Gifs', ['b'], 2],
    ],
  );
  assert.deepEqual(
    flatItems(s).map((g) => g.id),
    ['a', 'c', 'b'],
  );
});

test('sections order is GIPHY, Clips, Klipy, Finer Gifs whatever the input order', () => {
  const s = sections([gif('finergifs', '1'), gif('klipy', '2'), gif('clips', '3'), gif('giphy', '4')], '');
  assert.deepEqual(
    s.map((x) => x.provider),
    ['giphy', 'clips', 'klipy', 'finergifs'],
  );
});

test('filter matches titles case-insensitively and drops empty sections', () => {
  assert.deepEqual(
    sections([a, b, c], '  OFFICE ').map((x) => [x.provider, x.items.map((g) => g.id)]),
    [
      ['giphy', ['a']],
      ['finergifs', ['b']],
    ],
  );
  assert.deepEqual(
    sections([a, b, c], 'cat').map((x) => x.items.map((g) => g.id)),
    [['c']],
  );
  assert.deepEqual(sections([a, b, c], 'zzz'), []);
  assert.deepEqual(sections([], ''), []);
});

test('rows: a 32px header then ceil(n/4) 152px tile rows per section', () => {
  const many = Array.from({ length: 5 }, (_, i) => gif('giphy', `g${i}`));
  const r = rows(sections([...many, b], ''));
  assert.deepEqual(
    r.map((x) => [x.kind, x.size, x.section, x.kind === 'tiles' ? x.start : -1]),
    [
      ['header', 32, 0, -1],
      ['tiles', 152, 0, 0],
      ['tiles', 152, 0, 4],
      ['header', 32, 1, -1],
      ['tiles', 152, 1, 5],
    ],
  );
});
