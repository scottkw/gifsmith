import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  actionErrorText,
  downloadedText,
  folderErrorText,
  libraryErrorText,
  errorState,
  pagingMessage,
  providerName,
  keyMessage,
  revokedMessage,
  testKeyMessage,
  keySavedText,
  keyRemovedText,
  keyRemoveFailedText,
  hotkeyErrorText,
} from './errors.ts';

const NOW = Date.UTC(2026, 9, 17, 19, 0, 0);
const SWITCH = 'Other providers still work: ⌘1–⌘4.';

test('provider names', () => {
  assert.equal(providerName('giphy'), 'GIPHY');
  assert.equal(providerName('clips'), 'GIPHY Clips');
  assert.equal(providerName('klipy'), 'Klipy');
  assert.equal(providerName('finergifs'), 'The Finer Gifs Club');
});

test('network, timeout and http rows', () => {
  const net = errorState({ kind: 'network', provider: 'giphy' }, NOW);
  assert.equal(net?.id, 'E3');
  assert.equal(net?.heading, "Can't reach GIPHY");
  assert.equal(net?.body, 'Check your connection, VPN, proxy, or firewall, then try again.');
  assert.equal(net?.tryAgain, true);

  const slow = errorState({ kind: 'timeout', provider: 'klipy' }, NOW);
  assert.equal(slow?.id, 'E4');
  assert.equal(slow?.heading, 'Klipy is taking too long');
  assert.equal(slow?.body, 'The server may be slow right now. Try again, or switch providers with ⌘1–⌘4.');

  const http = errorState({ kind: 'http', provider: 'giphy', status: 503 }, NOW);
  assert.equal(http?.id, 'E5');
  assert.equal(http?.heading, "GIPHY isn't responding");
  assert.equal(http?.body, `GIPHY returned an error (HTTP 503). ${SWITCH}`);
  assert.equal(http?.tryAgain, true);
});

test('2xx and missing status drop the HTTP parenthetical', () => {
  assert.equal(errorState({ kind: 'http', provider: 'giphy', status: 200 }, NOW)?.body, `GIPHY returned an error. ${SWITCH}`);
  assert.equal(errorState({ kind: 'http', provider: 'klipy' }, NOW)?.body, `Klipy returned an error. ${SWITCH}`);
});

test('Finer Gifs down rows', () => {
  for (const e of [
    { kind: 'network', provider: 'finergifs' },
    { kind: 'http', provider: 'finergifs', status: 500 },
    { kind: 'http', provider: 'finergifs', status: 530 },
    { kind: 'http', provider: 'finergifs', status: 522 },
  ]) {
    const s = errorState(e, NOW);
    assert.equal(s?.id, 'E5f', JSON.stringify(e));
    assert.equal(s?.heading, 'The Finer Gifs Club seems to be down');
    assert.equal(
      s?.body,
      "Its server isn't responding right now. This happens from time to time. Try again later, or switch providers with ⌘1–⌘4.",
    );
    assert.equal(s?.tryAgain, true);
  }
  assert.equal(errorState({ kind: 'http', provider: 'finergifs', status: 200 }, NOW)?.id, 'E5');
  assert.equal(errorState({ kind: 'timeout', provider: 'finergifs' }, NOW)?.id, 'E4');
});

test('rate limit with a known retry time', () => {
  const s = errorState({ kind: 'rate_limited', provider: 'giphy', retryAfterSec: 120 }, NOW);
  assert.equal(s?.id, 'E1');
  assert.equal(s?.heading, 'GIPHY request limit reached');
  assert.ok(s?.body.startsWith('Free keys allow 100 requests per hour. Try again in 2 minutes (at '), s?.body);
  assert.ok(s?.body.endsWith(`). ${SWITCH}`), s?.body);
  assert.match(s?.body ?? '', /\(at \d{1,2}:\d{2}\s(AM|PM)\)/);
  assert.equal(s?.tryAgain, false);
  assert.equal(s?.retryAt, NOW + 120_000);

  const body = (sec: number) => errorState({ kind: 'rate_limited', provider: 'giphy', retryAfterSec: sec }, NOW)?.body ?? '';
  assert.ok(body(30).includes('Try again in less than a minute.'), body(30));
  assert.ok(!body(30).includes('(at '), body(30));
  assert.ok(body(61).includes('in 2 minutes'), body(61));
  assert.ok(body(60).includes('in 1 minute (at '), body(60));
});

test('rate limit re-evaluated against the time it was received', () => {
  const e = { kind: 'rate_limited', provider: 'giphy', retryAfterSec: 120 };
  assert.ok(errorState(e, NOW + 60_000, NOW)?.body.includes('in 1 minute (at '));
  assert.equal(errorState(e, NOW + 60_000, NOW)?.tryAgain, false);
  assert.equal(errorState(e, NOW + 60_000, NOW)?.retryAt, NOW + 120_000);
  assert.equal(errorState(e, NOW + 120_000, NOW)?.tryAgain, true);
});

test('rate limit with an unknown retry time', () => {
  const s = errorState({ kind: 'rate_limited', provider: 'klipy', retryAfterSec: 0 }, NOW);
  assert.equal(s?.id, 'E1');
  assert.equal(s?.body, `Free keys allow 100 requests per hour. Try again later this hour. ${SWITCH}`);
  assert.equal(s?.tryAgain, true);
  assert.equal(s?.retryAt, undefined);
  assert.equal(errorState({ kind: 'rate_limited', provider: 'klipy' }, NOW)?.body, s?.body);
});

test('key errors belong to the setup card', () => {
  assert.equal(errorState({ kind: 'key_missing', provider: 'giphy' }, NOW), null);
  assert.equal(errorState({ kind: 'key_rejected', provider: 'klipy', status: 401 }, NOW), null);
});

test('paging messages', () => {
  assert.equal(
    pagingMessage({ kind: 'rate_limited', provider: 'giphy', retryAfterSec: 120 }, NOW),
    'Request limit reached. More results in 2 min.',
  );
  assert.equal(pagingMessage({ kind: 'rate_limited', provider: 'giphy' }, NOW), 'Request limit reached. Try again later this hour.');
  assert.equal(pagingMessage({ kind: 'network', provider: 'giphy' }, NOW), "Couldn't load more.");
  assert.equal(pagingMessage({ kind: 'http', provider: 'finergifs', status: 500 }, NOW), "Couldn't load more.");
});

test('no copy leaks a URL or undefined', () => {
  const kinds = ['network', 'timeout', 'http', 'rate_limited', 'key_missing', 'key_rejected', 'weird'];
  for (const provider of ['giphy', 'clips', 'klipy', 'finergifs', 'nope']) {
    for (const kind of kinds) {
      for (const extra of [{}, { status: 503 }, { retryAfterSec: 90 }]) {
        const e = { kind, provider, ...extra };
        const s = errorState(e, NOW);
        for (const text of [s?.heading ?? '', s?.body ?? '', pagingMessage(e, NOW)]) {
          assert.ok(!/https?:\/\/|undefined/.test(text), `${JSON.stringify(e)} -> ${text}`);
        }
      }
    }
  }
});

test('keyMessage: rejected key (K2)', () => {
  assert.equal(
    keyMessage({ kind: 'key_rejected', provider: 'giphy', status: 401 }, 'GIPHY', NOW),
    'GIPHY rejected that key. Check that you copied the whole key, then try again.',
  );
});

test('keyMessage: network, timeout and other http (K3)', () => {
  const k3 = "Couldn't reach Klipy to check the key. Check your connection, then try again.";
  assert.equal(keyMessage({ kind: 'network', provider: 'klipy' }, 'Klipy', NOW), k3);
  assert.equal(keyMessage({ kind: 'timeout', provider: 'klipy' }, 'Klipy', NOW), k3);
  assert.equal(keyMessage({ kind: 'http', provider: 'klipy', status: 503 }, 'Klipy', NOW), k3);
});

test('keyMessage: rate limited with known and unknown retry (K4)', () => {
  assert.equal(
    keyMessage({ kind: 'rate_limited', provider: 'giphy', retryAfterSec: 180 }, 'GIPHY', NOW),
    "GIPHY's request limit was hit while checking the key. Try again in 3 minutes.",
  );
  assert.equal(
    keyMessage({ kind: 'rate_limited', provider: 'giphy', retryAfterSec: 0 }, 'GIPHY', NOW),
    "GIPHY's request limit was hit while checking the key. Try again later this hour.",
  );
});

test('keyMessage: Keychain failure', () => {
  assert.equal(keyMessage(null, 'GIPHY', NOW), "Couldn't save the key to the macOS Keychain. Try again.");
});

test('revokedMessage (K6)', () => {
  assert.equal(revokedMessage('Klipy'), 'Klipy rejected your saved key. It may have been revoked. Paste a new key to continue.');
});

test('action error rows F1 and F3 come from the kind, never the error text', () => {
  const F1 = "Couldn't copy this GIF. The download failed. Try again or pick another GIF.";
  const g = { provider: 'finergifs' } as Parameters<typeof actionErrorText>[1];
  for (const k of ['download_failed', 'not_allowed', 'too_large', 'copy_failed', '', 'something_new']) {
    assert.equal(actionErrorText(k, g), F1, k);
  }
  assert.equal(actionErrorText('not_media', g), "The Finer Gifs Club sent something that isn't a GIF. Nothing was saved. Try another GIF.");
  assert.equal(actionErrorText('not_media', { ...g, provider: 'klipy' }), "Klipy sent something that isn't a GIF. Nothing was saved. Try another GIF.");
});

test('page errors show F8 and a missing link shows F1', () => {
  const g = { provider: 'giphy' } as Parameters<typeof actionErrorText>[1];
  assert.equal(actionErrorText('no_page', g), "Couldn't open the page in your browser.");
  assert.equal(actionErrorText('open_failed', g), "Couldn't open the page in your browser.");
  assert.equal(actionErrorText('no_link', g), "Couldn't copy this GIF. The download failed. Try again or pick another GIF.");
});

test('Clips copy errors name the MP4 (F1 and F3 Clips variants)', () => {
  const clip = { provider: 'clips', isClip: true } as Parameters<typeof actionErrorText>[1];
  assert.equal(actionErrorText('download_failed', clip), "Couldn't copy this MP4. The download failed. Try again or pick another clip.");
  assert.equal(actionErrorText('not_media', clip), "GIPHY Clips sent something that isn't an MP4. Nothing was saved. Try another clip.");
});

test('text copy/paste failures show F11 and invalid_item shows F12 (P25)', () => {
  const g = { provider: 'giphy' } as Parameters<typeof actionErrorText>[1];
  const F11 = "Couldn't copy to the clipboard. Try again.";
  for (const k of ['copy_failed', 'no_link', 'not_allowed', '']) assert.equal(actionErrorText(k, g, 'text'), F11, k);
  const F12 = "This GIF's details are incomplete. Pick another GIF.";
  assert.equal(actionErrorText('invalid_item', g, 'text'), F12);
  assert.equal(actionErrorText('invalid_item', g), F12);
  assert.equal(actionErrorText('invalid_item', g, 'download'), F12);
});

test('download errors show F2, F3 and F4', () => {
  const g = { provider: 'finergifs' } as Parameters<typeof actionErrorText>[1];
  const clip = { provider: 'clips', isClip: true } as Parameters<typeof actionErrorText>[1];
  assert.equal(actionErrorText('download_failed', g, 'download'), "Couldn't download this GIF. Check your connection, then try again.");
  assert.equal(actionErrorText('', g, 'download'), "Couldn't download this GIF. Check your connection, then try again.");
  assert.equal(actionErrorText('download_failed', clip, 'download'), "Couldn't download this MP4. Check your connection, then try again.");
  const F4 = "Couldn't save to Downloads. Check that the folder exists and that Gifsmith can write to it.";
  assert.equal(actionErrorText('folder_unwritable', g, 'download'), F4);
  assert.equal(actionErrorText('folder_unwritable', g, 'download', 'uat-downloads'), F4.replace('Downloads', 'uat-downloads'));
  const F3 = "The Finer Gifs Club sent something that isn't a GIF. Nothing was saved. Try another GIF.";
  assert.equal(actionErrorText('not_media', g, 'download'), F3);
  assert.equal(actionErrorText('not_media', g, 'copy'), F3);
});

test('square crop errors show F6 and F7', () => {
  const g = { provider: 'giphy' } as Parameters<typeof actionErrorText>[1];
  assert.equal(actionErrorText('crop_failed', g), "Couldn't crop this GIF. Press ↵ to copy it as is, or pick another GIF.");
  assert.equal(actionErrorText('crop_too_large', g), 'This GIF is too large to crop (over 15 MB). Press ↵ to copy it as is.');
});

test('library write failures show F5 for the list that changed', () => {
  assert.equal(libraryErrorText('favorites'), "Couldn't save your Favorites. Try again.");
  assert.equal(libraryErrorText('recents'), "Couldn't save your Recents. Try again.");
});

test('testKeyMessage: passed (K-T1) and rejected (K-T2)', () => {
  assert.equal(testKeyMessage(null, 'GIPHY', NOW), 'GIPHY accepted the key.');
  assert.equal(
    testKeyMessage({ kind: 'key_rejected', provider: 'klipy', status: 404 }, 'Klipy', NOW),
    'Klipy rejected the saved key. It may have been revoked. Change it to continue.',
  );
});

test('testKeyMessage: rate limited (K-T3)', () => {
  const rl = (retryAfterSec?: number) => testKeyMessage({ kind: 'rate_limited', provider: 'giphy', retryAfterSec }, 'GIPHY', NOW);
  const base = "GIPHY's request limit was hit while testing the key. Try again ";
  assert.equal(rl(30), base + 'in less than a minute.');
  assert.equal(rl(600), base + 'in 10 minutes.');
  assert.equal(rl(61), base + 'in 2 minutes.');
  assert.equal(rl(), base + 'later this hour.');
});

test('testKeyMessage: network, timeout and http (K-T4)', () => {
  for (const kind of ['network', 'timeout', 'http']) {
    assert.equal(
      testKeyMessage({ kind, provider: 'giphy', status: 503 }, 'GIPHY', NOW),
      "Couldn't reach GIPHY to test the key. Check your connection, then try again.",
    );
  }
});

test('Keys pane save, remove and remove-failed texts (K-A, K-D, K-F)', () => {
  assert.equal(keySavedText('Klipy'), 'Saved. Klipy is ready.');
  assert.equal(keyRemovedText('GIPHY'), 'Removed. Add a key to use GIPHY again.');
  assert.equal(keyRemoveFailedText(), "Couldn't remove the key from the macOS Keychain. Try again.");
});

test('hotkeyErrorText: H1, H2, H3, H4, H4b and H7 with the combo as glyphs', () => {
  assert.equal(hotkeyErrorText('hotkey_no_modifier', 'H', ''), 'Include ⌘, ⌥ or ⌃ in the shortcut.');
  assert.equal(hotkeyErrorText('hotkey_duplicate', 'Ctrl+Option+Cmd+F', 'Open Favorites'), '⌃⌥⌘F is already used for Open Favorites.');
  assert.equal(hotkeyErrorText('hotkey_duplicate', 'Cmd+Ctrl+Option+G', 'Open Gifsmith'), '⌃⌥⌘G is already used for Open Gifsmith.');
  assert.equal(hotkeyErrorText('hotkey_duplicate', 'Cmd+Ctrl+Option+R', 'Open Recents'), '⌃⌥⌘R is already used for Open Recents.');
  assert.equal(hotkeyErrorText('hotkey_system', 'Cmd+Q', ''), '⌘Q is a macOS shortcut. Choose another one.');
  assert.equal(hotkeyErrorText('hotkey_register_failed', 'Cmd+Ctrl+Option+J', ''), "macOS couldn't register ⌃⌥⌘J. Your previous shortcut still works.");
  assert.equal(hotkeyErrorText('hotkey_register_failed_new', 'Cmd+Ctrl+Option+J', ''), "macOS couldn't register ⌃⌥⌘J. Try a different shortcut.");
  assert.equal(hotkeyErrorText('test_failed', 'Cmd+Ctrl+Option+F', ''), 'Nothing happened. Another app may be using ⌃⌥⌘F. Try a different shortcut.');
  assert.equal(hotkeyErrorText('prefs_write_failed', 'Cmd+Ctrl+Option+F', ''), "Couldn't save this setting. Try again.");
});

test('Downloads pane D5–D8 name the folder by display name', () => {
  assert.equal(
    folderErrorText('folder_fat', 'Stick'),
    "Gifsmith can't save safely to Stick (exFAT or FAT drive). Choose a folder on a Mac-formatted drive.",
  );
  assert.equal(folderErrorText('folder_network', 'Share'), 'Share is on a network drive. Choose a folder on this Mac.');
  assert.equal(folderErrorText('folder_unwritable', 'Locked'), "Gifsmith can't write to Locked. Choose another folder.");
  assert.equal(folderErrorText('folder_missing', 'Gone'), 'This folder no longer exists. Choose another folder or reset to Downloads.');
  assert.equal(folderErrorText('picker_failed', ''), "Couldn't save this setting. Try again.");
});

test('T6 and F4 use the folder display name, never a path', () => {
  assert.equal(downloadedText('GIFs'), 'Downloaded to GIFs');
  const g = { provider: 'finergifs' } as Parameters<typeof actionErrorText>[1];
  const f4 = actionErrorText('folder_unwritable', g, 'download', 'GIFs');
  assert.equal(f4, "Couldn't save to GIFs. Check that the folder exists and that Gifsmith can write to it.");
  assert.ok(!f4.includes('/'));
});
