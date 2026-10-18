import { test } from 'node:test';
import assert from 'node:assert/strict';
import { move, nearEnd, routeKey } from './grid.ts';

test('move left/right wraps across rows and clamps at the ends', () => {
  assert.equal(move(0, 'left', 10), 0);
  assert.equal(move(9, 'right', 10), 9);
  assert.equal(move(3, 'right', 10), 4);
  assert.equal(move(4, 'left', 10), 3);
});

test('move up by a row, first row does nothing', () => {
  assert.equal(move(2, 'up', 10), 2);
  assert.equal(move(6, 'up', 10), 2);
});

test('move down by a row, last full row falls to the last tile', () => {
  assert.equal(move(1, 'down', 10), 5);
  assert.equal(move(5, 'down', 10), 9);
  assert.equal(move(2, 'down', 6), 5);
  assert.equal(move(3, 'down', 6), 5);
  assert.equal(move(4, 'down', 6), 4);
  assert.equal(move(9, 'down', 10), 9);
});

test('move on one or zero tiles stays at 0', () => {
  assert.equal(move(0, 'down', 1), 0);
  assert.equal(move(0, 'right', 0), 0);
});

test('nearEnd is true within the last 8 tiles', () => {
  assert.equal(nearEnd(15, 24), false);
  assert.equal(nearEnd(16, 24), true);
  assert.equal(nearEnd(0, 3), true);
});

const key = (k: string, o: Partial<{ metaKey: boolean; shiftKey: boolean; isComposing: boolean; keyCode: number; repeat: boolean }> = {}) =>
  routeKey({ key: k, metaKey: false, shiftKey: false, isComposing: false, keyCode: 0, repeat: false, ...o });

test('routeKey maps Enter, Escape, Cmd+K, Tab and arrows', () => {
  assert.equal(key('Enter'), 'copy');
  assert.equal(key('Enter', { metaKey: true }), 'paste');
  assert.equal(key('Escape'), 'escape');
  assert.equal(key('k', { metaKey: true }), 'actions');
  assert.equal(key('Tab'), 'tab');
  assert.equal(key('ArrowLeft'), 'left');
  assert.equal(key('ArrowRight'), 'right');
  assert.equal(key('ArrowUp'), 'up');
  assert.equal(key('ArrowDown'), 'down');
  assert.equal(key('a'), 'none');
});

test('routeKey leaves IME composition keys and repeated Enter alone', () => {
  assert.equal(key('Enter', { isComposing: true }), 'none');
  assert.equal(key('Enter', { keyCode: 229 }), 'none');
  assert.equal(key('Enter', { repeat: true }), 'none');
  assert.equal(key('ArrowDown', { isComposing: true }), 'none');
  assert.equal(key('Escape', { keyCode: 229 }), 'none');
});

test('routeKey maps Cmd+1..Cmd+6 to provider and library slots', () => {
  assert.equal(key('1', { metaKey: true }), 'slot1');
  assert.equal(key('2', { metaKey: true }), 'slot2');
  assert.equal(key('3', { metaKey: true }), 'slot3');
  assert.equal(key('4', { metaKey: true }), 'slot4');
  assert.equal(key('5', { metaKey: true }), 'slot5');
  assert.equal(key('6', { metaKey: true }), 'slot6');
  assert.equal(key('7', { metaKey: true }), 'none');
  assert.equal(key('6'), 'none');
  assert.equal(key('1'), 'none');
});

test('routeKey maps Cmd+Shift+C to copyLink and leaves Cmd+C to the input', () => {
  assert.equal(key('c', { metaKey: true, shiftKey: true }), 'copyLink');
  assert.equal(key('C', { metaKey: true, shiftKey: true }), 'copyLink');
  assert.equal(key('c', { metaKey: true }), 'none');
  assert.equal(key('C', { shiftKey: true }), 'none');
});

test('routeKey maps Cmd+Shift+M to copyMarkdown and Cmd+O (no Shift) to open', () => {
  assert.equal(key('m', { metaKey: true, shiftKey: true }), 'copyMarkdown');
  assert.equal(key('M', { metaKey: true, shiftKey: true }), 'copyMarkdown');
  assert.equal(key('m', { metaKey: true }), 'none');
  assert.equal(key('o', { metaKey: true }), 'open');
  assert.equal(key('O', { metaKey: true, shiftKey: true }), 'none');
  assert.equal(key('o'), 'none');
});

test('routeKey maps Cmd+S to download and Cmd+R to reveal (no Shift)', () => {
  assert.equal(key('s', { metaKey: true }), 'download');
  assert.equal(key('S', { metaKey: true, shiftKey: true }), 'none');
  assert.equal(key('s'), 'none');
  assert.equal(key('r', { metaKey: true }), 'reveal');
  assert.equal(key('R', { metaKey: true, shiftKey: true }), 'none');
});
