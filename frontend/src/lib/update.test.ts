import { test } from 'node:test';
import assert from 'node:assert/strict';
import { updateErrorText, updateText, updateToastText } from './update.ts';

const st = (state: string, extra: Partial<{ done: number; total: number }> = {}) => ({
  state,
  current: '1.0.0',
  latest: '9.9.0',
  done: 0,
  total: 0,
  ...extra,
});

test('update status lines (U2-U7)', () => {
  assert.equal(updateText({ ...st('idle'), latest: '' }), '');
  assert.equal(updateText(st('checking')), 'Checking for updates…');
  assert.equal(updateText({ ...st('current'), latest: '' }), 'Gifsmith 1.0.0 is up to date.');
  assert.equal(updateText(st('available')), 'Gifsmith 9.9.0 is available. You have 1.0.0.');
  assert.equal(updateText(st('downloading', { done: 524288, total: 1048576 })), 'Downloading Gifsmith 9.9.0… 50%');
  assert.equal(updateText(st('downloading')), 'Downloading Gifsmith 9.9.0…');
  assert.equal(updateText(st('verifying')), "Checking the download's signature…");
  assert.equal(
    updateText(st('opened')),
    'The Gifsmith 9.9.0 disk image is open. Quit Gifsmith, then drag the new Gifsmith to Applications and choose Replace.',
  );
});

test('update errors (U-E1-U-E4)', () => {
  const U_E1 = "Couldn't check for updates. Check your connection, then try again.";
  assert.equal(updateErrorText('update_check_failed', st('idle')), U_E1);
  assert.equal(updateErrorText('update_download_failed', st('available')), "Couldn't download Gifsmith 9.9.0. Check your connection, then try again.");
  assert.equal(updateErrorText('update_verify_failed', st('available')), "The download didn't pass the signature check, so Gifsmith deleted it. Nothing was opened.");
  assert.equal(updateErrorText('update_open_failed', st('available')), "Gifsmith couldn't open the downloaded disk image. Try again.");
  assert.equal(updateErrorText('something_else', st('idle')), U_E1);
});

test('update toast (T12)', () => {
  assert.equal(updateToastText('9.9.0'), 'Gifsmith 9.9.0 is available.');
});
