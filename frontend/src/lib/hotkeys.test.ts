import { test } from 'node:test';
import assert from 'node:assert/strict';
import { accelFromEvent, canonical, DENYLIST, glyphs, hasModifier, spoken } from './hotkeys.ts';

const ev = (o: Partial<{ key: string; code: string; ctrlKey: boolean; altKey: boolean; metaKey: boolean; shiftKey: boolean }>) =>
  accelFromEvent({ key: '', code: '', ctrlKey: false, altKey: false, metaKey: false, shiftKey: false, ...o });

test('accelFromEvent builds Wails accelerators from letters, digits, F-keys, Space and arrows', () => {
  assert.deepEqual(ev({ key: 'g', code: 'KeyG', ctrlKey: true, altKey: true, metaKey: true }), { kind: 'accel', accel: 'Ctrl+Option+Cmd+G' });
  assert.deepEqual(ev({ key: '5', code: 'Digit5', metaKey: true }), { kind: 'accel', accel: 'Cmd+5' });
  assert.deepEqual(ev({ key: 'F13', code: 'F13', metaKey: true }), { kind: 'accel', accel: 'Cmd+F13' });
  assert.deepEqual(ev({ key: ' ', code: 'Space', ctrlKey: true }), { kind: 'accel', accel: 'Ctrl+Space' });
  assert.deepEqual(ev({ key: 'ArrowUp', code: 'ArrowUp', metaKey: true }), { kind: 'accel', accel: 'Cmd+Up' });
  assert.deepEqual(ev({ key: 'K', code: 'KeyK', shiftKey: true, metaKey: true }), { kind: 'accel', accel: 'Shift+Cmd+K' });
});

test('accelFromEvent takes the letter from code when Option changes the key', () => {
  assert.deepEqual(ev({ key: '©', code: 'KeyG', altKey: true }), { kind: 'accel', accel: 'Option+G' });
});

test('accelFromEvent maps punctuation by code and never emits + or super', () => {
  assert.deepEqual(ev({ key: '+', code: 'Equal', shiftKey: true, metaKey: true }), { kind: 'accel', accel: 'Shift+Cmd+=' });
  assert.deepEqual(ev({ key: '/', code: 'Slash', ctrlKey: true }), { kind: 'accel', accel: 'Ctrl+/' });
  assert.deepEqual(ev({ key: '`', code: 'Backquote', ctrlKey: true }), { kind: 'accel', accel: 'Ctrl+`' });
  const r = ev({ key: '+', code: 'NumpadAdd', metaKey: true });
  assert.notEqual(r.kind === 'accel' && r.accel.endsWith('+'), true);
  for (const code of ['KeyA', 'Digit1', 'Space', 'Equal', 'ArrowLeft']) {
    const a = ev({ key: 'x', code, metaKey: true });
    assert.equal(a.kind === 'accel' && /super/i.test(a.accel), false);
  }
});

test('accelFromEvent reports held modifiers, a missing modifier and unsupported keys', () => {
  assert.deepEqual(ev({ key: 'Meta', code: 'MetaLeft', metaKey: true, ctrlKey: true }), { kind: 'held', glyphs: '⌃⌘' });
  assert.deepEqual(ev({ key: 'Shift', code: 'ShiftLeft', shiftKey: true }), { kind: 'held', glyphs: '⇧' });
  assert.deepEqual(ev({ key: 'h', code: 'KeyH' }), { kind: 'no_modifier' });
  assert.deepEqual(ev({ key: 'H', code: 'KeyH', shiftKey: true }), { kind: 'no_modifier' });
  assert.deepEqual(ev({ key: 'Home', code: 'Home', metaKey: true }), { kind: 'unsupported' });
});

test('spoken names the modifiers in macOS order, then the key', () => {
  assert.equal(spoken('Cmd+Ctrl+Option+G'), 'Control Option Command G');
  assert.equal(spoken('Cmd+Shift+UP'), 'Shift Command Up Arrow');
  assert.equal(spoken('Ctrl+Option+SPACE'), 'Control Option Space');
});

test('glyphs renders in macOS order ⌃⌥⇧⌘ then the key', () => {
  assert.equal(glyphs('Cmd+Ctrl+Option+G'), '⌃⌥⌘G');
  assert.equal(glyphs('Cmd+Shift+3'), '⇧⌘3');
  assert.equal(glyphs('Ctrl+SPACE'), '⌃Space');
  assert.equal(glyphs('Cmd+UP'), '⌘↑');
  assert.equal(glyphs('Ctrl+Option+Cmd+F13'), '⌃⌥⌘F13');
  assert.equal(glyphs(''), '');
});

test('canonical matches Go: sorted modifiers, upper-cased key', () => {
  assert.equal(canonical('Ctrl+Option+Cmd+G'), 'Cmd+Ctrl+Option+G');
  assert.equal(canonical('Shift+Cmd+k'), 'Cmd+Shift+K');
  assert.equal(canonical('Ctrl+Space'), 'Ctrl+SPACE');
  assert.equal(canonical('Cmd+Ctrl+Option+G'), canonical('Ctrl+Option+Cmd+G'));
});

test('DENYLIST holds the macOS shortcuts in canonical form; hasModifier needs ⌘ ⌥ or ⌃', () => {
  for (const a of ['Cmd+Q', 'Cmd+W', 'Cmd+H', 'Cmd+M', 'Cmd+Tab', 'Cmd+Space', 'Ctrl+Space', 'Shift+Cmd+3', 'Shift+Cmd+4', 'Shift+Cmd+5'])
    assert.equal(DENYLIST.has(canonical(a)), true, a);
  assert.equal(DENYLIST.has(canonical('Ctrl+Option+Cmd+G')), false);
  assert.equal(hasModifier('Shift+G'), false);
  assert.equal(hasModifier('Option+G'), true);
});
