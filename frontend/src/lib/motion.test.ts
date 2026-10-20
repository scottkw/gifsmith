import { test } from 'node:test';
import assert from 'node:assert/strict';
import { animates } from './motion.ts';

test('animates: every tile without Reduce Motion', () => {
  for (const i of [0, 1, 7, 40]) assert.equal(animates(i, 3, false), true);
  assert.equal(animates(0, -1, false), true);
});

test('animates: only the selected tile under Reduce Motion', () => {
  assert.equal(animates(3, 3, true), true);
  assert.equal(animates(2, 3, true), false);
  assert.equal(animates(0, -1, true), false); // nothing selected: all still
});
