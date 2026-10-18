import { test } from 'node:test';
import assert from 'node:assert/strict';
import { resolveTheme } from './theme.ts';

test('System follows the OS appearance', () => {
  assert.equal(resolveTheme('', true), 'dark');
  assert.equal(resolveTheme('', false), 'light');
});

test('Light and Dark override the OS', () => {
  assert.equal(resolveTheme('light', true), 'light');
  assert.equal(resolveTheme('dark', false), 'dark');
});

test('an unknown value follows the system', () => {
  assert.equal(resolveTheme('neon', true), 'dark');
});
