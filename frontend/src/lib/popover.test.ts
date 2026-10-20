import { test } from 'node:test';
import assert from 'node:assert/strict';
import { MENU_W, placeInDetails, placeMenu } from './popover.ts';

const tile = (x: number, y: number, width = 166, height = 144) => ({ x, y, width, height });

test('opens right of the tile when tile.right + 8 + 320 <= 712, else left of it', () => {
  assert.deepEqual(placeMenu(tile(16, 100), 280), { left: 190, top: 100 });
  assert.deepEqual(placeMenu(tile(538, 100), 280), { left: 210, top: 100 });
  assert.equal(placeMenu(tile(190, 100), 280).left, 364);
  assert.equal(placeMenu(tile(364, 100), 280).left, 36);
  // The boundary: right edge 384 fits exactly (384 + 8 + 320 = 712); right edge 400 does not.
  assert.equal(placeMenu(tile(218, 100), 280).left, 392);
  assert.equal(placeMenu(tile(234, 100), 280).left, 8);
});

test('3 columns: a tile whose right edge is 600 opens left at tile.left − 328, clamped >= 8', () => {
  assert.equal(placeMenu(tile(376, 100, 224, 192), 280).left, 48);
  assert.equal(placeMenu(tile(16, 100, 224, 192), 280).left, 248);
  assert.equal(placeMenu(tile(240, 100, 224, 192), 280).left, 8); // 240 − 328 clamped
});

test('the bottom stays at or above 440 so the footer attribution shows (P19)', () => {
  assert.equal(placeMenu(tile(16, 400), 280).top, 160);
  assert.equal(placeMenu(tile(16, 300), 100).top, 300);
  assert.equal(placeMenu(tile(16, 360), 100).top, 340);
  assert.equal(placeMenu(tile(16, 0), 280).top, 8);
  assert.equal(placeMenu(tile(16, -60), 100).top, 8);
});

test('left never leaves the 8px inset', () => {
  for (const x of [16, 190, 364, 538, -100, 700]) {
    const { left } = placeMenu(tile(x, 100), 280);
    assert.ok(left >= 8, `left ${left} for x ${x}`);
    assert.ok(left + MENU_W <= 712, `right ${left + MENU_W} for x ${x}`);
  }
});

test('window size can be passed; the footer clamp follows the window height', () => {
  assert.equal(placeMenu(tile(16, 700), 280, { w: 1000, h: 800 }).top, 480);
});

test('placeInDetails sits 16px inside the content area top-right, bottom clamped to 440', () => {
  assert.deepEqual(placeInDetails(280), { left: 720 - 16 - 320, top: 84 + 16 });
  assert.deepEqual(placeInDetails(380), { left: 384, top: 60 });
});
