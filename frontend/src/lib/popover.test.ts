import { test } from 'node:test';
import assert from 'node:assert/strict';
import { MENU_W, placeInDetails, placeMenu } from './popover.ts';

const tile = (x: number, y: number) => ({ x, y, width: 166, height: 144 });

test('columns 1-2 open right of the tile, columns 3-4 left of it', () => {
  assert.deepEqual(placeMenu(tile(16, 100), 0, 280), { left: 190, top: 100 });
  assert.deepEqual(placeMenu(tile(538, 100), 3, 280), { left: 210, top: 100 });
  assert.equal(placeMenu(tile(190, 100), 1, 280).left, 364);
  assert.equal(placeMenu(tile(364, 100), 2, 280).left, 36);
});

test('top clamps so the menu stays 8px inside the window', () => {
  assert.equal(placeMenu(tile(16, 400), 0, 280).top, 192);
  assert.equal(placeMenu(tile(16, 0), 0, 280).top, 8);
  assert.equal(placeMenu(tile(16, -60), 0, 100).top, 8);
});

test('left never leaves the 8px inset', () => {
  for (const [x, col] of [[16, 0], [190, 1], [364, 2], [538, 3], [-100, 3], [700, 0]] as const) {
    const { left } = placeMenu(tile(x, 100), col, 280);
    assert.ok(left >= 8, `left ${left} for x ${x}`);
    assert.ok(left + MENU_W <= 712, `right ${left + MENU_W} for x ${x}`);
  }
});

test('window size can be passed', () => {
  assert.equal(placeMenu(tile(16, 700), 0, 280, { w: 1000, h: 800 }).top, 512);
});

test('placeInDetails sits 16px inside the content area top-right', () => {
  assert.deepEqual(placeInDetails(280), { left: 720 - 16 - 320, top: 84 + 16 });
});
