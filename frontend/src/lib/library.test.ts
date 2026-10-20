import { test } from 'node:test';
import assert from 'node:assert/strict';
import { flatItems, moveSectioned, rows, sections } from './library.ts';
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
  const r = rows(sections([...many, b], ''), 4, 152);
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

// Sections A (5 items: rows [0 1 2 3] [4]) and B (3 items: row [5 6 7]).
test('moveSectioned crosses sections with left/right and by column with up/down', () => {
  const m = (i: number, d: 'left' | 'right' | 'up' | 'down') => moveSectioned(i, d, [5, 3], 4);
  assert.equal(m(4, 'right'), 5);
  assert.equal(m(5, 'left'), 4);
  assert.equal(m(1, 'down'), 4); // nothing directly below inside A: A's later row's last tile
  assert.equal(m(3, 'down'), 4);
  assert.equal(m(4, 'down'), 5); // A's last row -> B's first row, same column
  assert.equal(m(5, 'down'), 5); // last section, last row
  assert.equal(m(7, 'up'), 4); // column 2 clamped to A's last row's last tile
  assert.equal(m(5, 'up'), 4);
  assert.equal(m(2, 'up'), 2); // first section, first row
  assert.equal(m(0, 'left'), 0);
  assert.equal(m(7, 'right'), 7);
});

test('moveSectioned within one section is the Phase 2 grid move; down into a short next row clamps', () => {
  assert.equal(moveSectioned(1, 'down', [10], 4), 5);
  assert.equal(moveSectioned(6, 'up', [10], 4), 2);
  assert.equal(moveSectioned(3, 'down', [4, 2], 4), 5); // column 3 clamped to B's last tile
  assert.equal(moveSectioned(0, 'down', [1], 4), 0);
  assert.equal(moveSectioned(0, 'right', [], 4), 0);
});

test('rows and moveSectioned take the column count and tile stride', () => {
  const seven = Array.from({ length: 7 }, (_, i) => gif('giphy', `g${i}`));
  const r = rows(sections(seven, ''), 3, 200);
  assert.deepEqual(
    r.map((x) => [x.kind, x.size, x.kind === 'tiles' ? x.items.length : 0]),
    [
      ['header', 32, 0],
      ['tiles', 200, 3],
      ['tiles', 200, 3],
      ['tiles', 200, 1],
    ],
  );
  // 5 columns: ↓ from column 3 of A's last row lands on column 3 of B's first row.
  assert.equal(moveSectioned(8, 'down', [9, 6], 5), 12);
  assert.equal(moveSectioned(12, 'up', [9, 6], 5), 8);
  assert.equal(moveSectioned(2, 'down', [9, 6], 5), 7);
});
