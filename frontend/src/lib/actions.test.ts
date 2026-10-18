import { test } from 'node:test';
import assert from 'node:assert/strict';
import { buildActions, filterActions, type ActionCtx } from './actions.ts';
import type { Gif } from './api';

const gif = (o: Partial<Gif> = {}): Gif =>
  ({ id: '1', provider: 'giphy', title: 't', previewUrl: 'p', downloadUrl: 'd', isClip: false, ...o }) as Gif;
const ctx = (g: Gif): ActionCtx => ({ gif: g, tab: 'provider', inDetails: false, favorited: false });
const giphy = buildActions(ctx(gif({ pageUrl: 'https://giphy.com/gifs/x' })));

test('GIPHY item lists every row in the locked order with hints and groups', () => {
  assert.deepEqual(
    giphy.map((r) => r.id),
    ['copy', 'paste', 'copyLink', 'pasteLink', 'copyMarkdown', 'pasteMarkdown', 'copyPage', 'open', 'download', 'copySquare', 'pasteSquare', 'favorite'],
  );
  assert.deepEqual(
    giphy.map((r) => r.label),
    ['Copy GIF', 'Paste GIF', 'Copy Link', 'Paste Link', 'Copy Markdown', 'Paste Markdown', 'Copy Page Link', 'Open in Browser', 'Download GIF', 'Copy Square GIF', 'Paste Square GIF', 'Add to Favorites'],
  );
  assert.deepEqual(
    giphy.map((r) => r.hint),
    ['↵', '⌘↵', '⌘⇧C', '', '⌘⇧M', '', '', '⌘O', '⌘S', '', '', '⌘⇧F'],
  );
  assert.deepEqual(
    giphy.map((r) => r.group),
    [1, 1, 2, 2, 2, 2, 2, 3, 3, 4, 4, 5],
  );
  assert.equal(giphy[2].keys, 'Meta+Shift+C');
  assert.equal(giphy[3].keys, '');
});

test('item without a page URL omits Copy Page Link and Open in Browser', () => {
  const ids = buildActions(ctx(gif({ provider: 'klipy' }))).map((r) => r.id);
  assert.ok(!ids.includes('copyPage'));
  assert.ok(!ids.includes('open'));
  assert.equal(ids.length, 10);
});

test('Clips read MP4; a Clip without a GIF rendition has no link or Markdown rows', () => {
  const withGif = buildActions(ctx(gif({ isClip: true, gifUrl: 'https://media.giphy.com/x.gif' })));
  assert.deepEqual(withGif.slice(0, 2).map((r) => r.label), ['Copy MP4', 'Paste MP4']);
  assert.ok(withGif.some((r) => r.id === 'copyLink'));
  const bare = buildActions(ctx(gif({ isClip: true }))).map((r) => r.id);
  for (const id of ['copyLink', 'pasteLink', 'copyMarkdown', 'pasteMarkdown'] as const) assert.ok(!bare.includes(id), id);
  assert.deepEqual(bare, ['copy', 'paste', 'download', 'favorite']);
});

test('filterActions matches a trimmed, case-insensitive label substring', () => {
  assert.deepEqual(filterActions(giphy, 'mark').map((r) => r.id), ['copyMarkdown', 'pasteMarkdown']);
  assert.deepEqual(filterActions(giphy, '  LINK ').map((r) => r.id), ['copyLink', 'pasteLink', 'copyPage']);
  assert.deepEqual(filterActions(giphy, 'zzz'), []);
  assert.deepEqual(filterActions(giphy, '   '), giphy);
});

test('filterActions keeps order and buildActions never repeats an id', () => {
  const ids = filterActions(giphy, 'paste').map((r) => r.id);
  assert.deepEqual(ids, ['paste', 'pasteLink', 'pasteMarkdown', 'pasteSquare']);
  assert.equal(new Set(giphy.map((r) => r.id)).size, giphy.length);
});

test('Download comes right after Open in Browser, or after Paste Markdown without a page', () => {
  const ids = giphy.map((r) => r.id);
  assert.equal(ids[ids.indexOf('open') + 1], 'download');
  const dl = giphy.find((r) => r.id === 'download')!;
  assert.deepEqual([dl.label, dl.hint, dl.keys, dl.group], ['Download GIF', '⌘S', 'Meta+S', 3]);
  const klipy = buildActions(ctx(gif({ provider: 'klipy' }))).map((r) => r.id);
  assert.equal(klipy[klipy.indexOf('pasteMarkdown') + 1], 'download');
  const clip = buildActions(ctx(gif({ isClip: true })));
  assert.equal(clip.find((r) => r.id === 'download')?.label, 'Download MP4');
});

test('Copy Square GIF follows Download in group 4 and is omitted for Clips (D-15)', () => {
  const ids = giphy.map((r) => r.id);
  assert.equal(ids[ids.indexOf('download') + 1], 'copySquare');
  const sq = giphy.find((r) => r.id === 'copySquare')!;
  assert.deepEqual([sq.label, sq.hint, sq.keys, sq.group], ['Copy Square GIF', '', '', 4]);
  const clip = buildActions(ctx(gif({ isClip: true, gifUrl: 'https://media.giphy.com/x.gif', pageUrl: 'https://giphy.com/clips/x' })));
  assert.ok(!clip.some((r) => r.id === 'copySquare' || r.label.includes('Square')));
});

test('Paste Square GIF follows Copy Square GIF; a Clip never lists a square row', () => {
  const ids = giphy.map((r) => r.id);
  assert.equal(ids[ids.indexOf('copySquare') + 1], 'pasteSquare');
  const sq = giphy.find((r) => r.id === 'pasteSquare')!;
  assert.deepEqual([sq.label, sq.hint, sq.keys, sq.group], ['Paste Square GIF', '', '', 4]);
  for (const g of [gif({ isClip: true }), gif({ isClip: true, gifUrl: 'https://media.giphy.com/x.gif' })]) {
    assert.ok(!buildActions(ctx(g)).some((r) => r.id === 'copySquare' || r.id === 'pasteSquare'));
  }
});

test('favorite row toggles by state; Recents adds Remove from Recents and Clear Recents (group 6)', () => {
  const g = gif({ pageUrl: 'https://giphy.com/gifs/x' });
  const row = (rows: ReturnType<typeof buildActions>, id: string) => rows.find((r) => r.id === id);
  const add = row(buildActions(ctx(g)), 'favorite')!;
  assert.deepEqual([add.label, add.hint, add.keys, add.group, add.destructive], ['Add to Favorites', '⌘⇧F', 'Meta+Shift+F', 5, false]);
  const fav = row(buildActions({ ...ctx(g), favorited: true }), 'favorite')!;
  assert.deepEqual([fav.label, fav.destructive], ['Remove from Favorites', true]);

  const provider = buildActions({ ...ctx(g), favorited: true }).map((r) => r.id);
  assert.ok(!provider.includes('removeRecent') && !provider.includes('clearRecents'));

  const recents = buildActions({ ...ctx(g), tab: 'recents' });
  assert.deepEqual(
    recents.slice(-3).map((r) => [r.id, r.label, r.hint, r.group, r.destructive]),
    [
      ['favorite', 'Add to Favorites', '⌘⇧F', 5, false],
      ['removeRecent', 'Remove from Recents', '⌘⌫', 6, true],
      ['clearRecents', 'Clear Recents', '', 6, true],
    ],
  );
  const favorites = buildActions({ ...ctx(g), tab: 'favorites', favorited: true });
  assert.deepEqual(
    favorites.slice(-1).map((r) => [r.id, r.label, r.destructive]),
    [['favorite', 'Remove from Favorites', true]],
  );
  assert.ok(!favorites.some((r) => r.id === 'removeRecent' || r.id === 'clearRecents'));
});
