// Thin facade over the generated bindings. Provider HTTP happens in Go; this only calls it.
import {
  ActionService,
  GifService,
  LibraryService,
  PanelService,
  SettingsService,
  type DownloadResult,
  type ProviderStatus,
  type Result,
  type KeyStatus,
  type Snapshot as SettingsSnapshot,
} from '../../bindings/gifsmith/internal/services';
import type { Gif, Page } from '../../bindings/gifsmith/internal/provider';
import type { Entry, Snapshot } from '../../bindings/gifsmith/internal/library';
import { Browser, CancelError, type CancellablePromise } from '@wailsio/runtime';
import { CLIPS_ACCESS_MAILTO, GIPHY_KEY_URL, KLIPY_KEY_URL } from './providers';

export type { DownloadResult, Entry, Gif, KeyStatus, Page, ProviderStatus, Result, SettingsSnapshot };

export const search = (p: string, q: string, cursor: string) => GifService.Search(p, q, cursor);
export const trending = (p: string, cursor: string) => GifService.Trending(p, cursor);
// Key state arrives only as booleans; no call returns a key.
export const providers = async () => (await GifService.Providers()) ?? [];
export const lastProvider = () => SettingsService.LastProvider();
export const setLastProvider = (p: string) => SettingsService.SetLastProvider(p);
// Settings: every change applies at once; Go validates and answers with the full snapshot.
export const snapshot = () => SettingsService.Snapshot();
export const setPref = (name: string, value: string) => SettingsService.SetPref(name, value);
// The Search-pane choices Go allows, and what Match System resolves to now.
export const choices = () => SettingsService.Choices();
// Registers or removes the login item now; rejects with cause.kind autostart_failed | not_in_applications.
export const setLaunchAtLogin = (on: boolean) => SettingsService.SetLaunchAtLogin(on);
// Downloads: Go shows the native folder picker and validates the pick; no path crosses from the
// webview. Rejects with cause.kind folder_missing | folder_fat | folder_network | folder_unwritable
// (cause.folder names it) or picker_failed; a cancel resolves to the unchanged snapshot.
export const chooseDownloadDir = () => SettingsService.ChooseDownloadDir();
export const resetDownloadDir = () => SettingsService.ResetDownloadDir();
// The Automatic paste grant, read without prompting.
export const accessibilityStatus = () => SettingsService.AccessibilityStatus();
// Go's services.SettingsError arrives as err.cause: { kind, slot? }.
export const settingsError = (e: unknown): { kind: string; slot?: string; folder?: string } | null => providerError(e);
// Verifies with a real trending call, then stores in the Keychain; rejects with a provider cause.
export const saveKey = (p: string, key: string) => SettingsService.SaveKey(p, key);
// Keys pane: status carries the last 4 characters at most; Test is one uncached trending call;
// Remove deletes the Keychain item (rejects with cause.kind keychain_delete_failed).
export const keyStatus = async () => (await SettingsService.KeyStatus()) ?? [];
export const testKey = (p: string) => SettingsService.TestKey(p);
export const removeKey = (p: string) => SettingsService.RemoveKey(p);
// Hotkeys: Go validates, registers (re-registering the old binding on failure) and stores; rejects
// with cause.kind hotkey_* (hotkey_duplicate names the other slot). The calls run one at a time, in
// order, so a late Suspend can never land after the Resume that ends a recording.
let hotkeyCalls: Promise<unknown> = Promise.resolve();
function inOrder<T>(f: () => Promise<T>): Promise<T> {
  const p = hotkeyCalls.then(f);
  hotkeyCalls = p.catch(() => {});
  return p;
}
export const setHotkey = (slot: string, accel: string) => inOrder(() => SettingsService.SetHotkey(slot, accel));
export const suspendHotkeys = () => inOrder(() => SettingsService.SuspendHotkeys());
export const resumeHotkeys = () => inOrder(() => SettingsService.ResumeHotkeys());
// Arms slot for 5 s: a press emits hotkey:test {slot} before its normal action.
export const testHotkey = (slot: string) => SettingsService.TestHotkey(slot);
// Re-probes Clips once when its result is unknown; resolves to the state, rejects with the probe's cause.
export const checkClips = () => GifService.CheckClips();

// The webview may only ask the OS to open these fixed URLs.
const EXTERNAL = new Set([GIPHY_KEY_URL, KLIPY_KEY_URL, CLIPS_ACCESS_MAILTO]);
export function openExternal(url: string) {
  if (!EXTERNAL.has(url)) throw new Error('refusing to open an unlisted URL');
  return Browser.OpenURL(url);
}

// One in-flight call per slot ("first:<provider>", "page:<provider>"): a new call cancels the
// previous one (its Go ctx and HTTP request). A cancelled or superseded call resolves to
// undefined and must not render; only the slot's latest call yields a result or an error.
const slots = new Map<string, CancellablePromise<unknown>>();

export function latest<T>(slot: string, p: CancellablePromise<T>): Promise<T | undefined> {
  slots.get(slot)?.cancel();
  slots.set(slot, p);
  const mine = () => slots.get(slot) === p;
  return p.then(
    (v) => {
      if (!mine()) return undefined;
      slots.delete(slot); // settled: a later cancel(slot) must not send a stray CancelCall
      return v;
    },
    (e) => {
      if (e instanceof CancelError || !mine()) return undefined;
      slots.delete(slot);
      throw e;
    },
  );
}

export function cancel(slot: string) {
  slots.get(slot)?.cancel();
  slots.delete(slot);
}

// Go's provider.Error arrives as err.cause (Options.MarshalError); anything else has no cause.
export type ProviderErr = { kind: string; provider: string; status?: number; retryAfterSec?: number };
export function providerError(e: unknown): ProviderErr | null {
  const c = (e as { cause?: unknown } | null)?.cause as ProviderErr | undefined;
  return c && typeof c.kind === 'string' ? c : null;
}

// Go's services.ActionError arrives as err.cause too: { kind } only.
export const actionError = (e: unknown): { kind: string } | null => providerError(e);

export const copy = (g: Gif) => ActionService.Copy(g);
// Go builds the text (and validates every URL) from g; the webview only names the kind.
export type TextKind = 'link' | 'markdown' | 'page';
export const copyText = (g: Gif, kind: TextKind) => ActionService.CopyText(g, kind);
export const pasteText = (g: Gif, kind: 'link' | 'markdown') => ActionService.PasteText(g, kind);
// Per-item URLs open only through Go (platform.Fake in the UAT build), never Browser.OpenURL.
export const openPage = (g: Gif) => ActionService.OpenPage(g);
export const openSource = (g: Gif) => ActionService.OpenSource(g);
export const paste = (g: Gif) => ActionService.Paste(g);
// Go downloads, crops and stages the square GIF; Clips are refused.
export const copySquare = (g: Gif) => ActionService.CopySquare(g);
export const pasteSquare = (g: Gif) => ActionService.PasteSquare(g);
// Saves into the download folder Go chooses; the result is the file and folder names, never a path.
export const download = (g: Gif) => ActionService.Download(g);
// Reveals the file Go saved last; no path crosses from the webview.
export const showDownload = () => ActionService.ShowDownload();
export const requestAccessibility = () => ActionService.RequestAccessibility();
export const hide = () => PanelService.Hide();
export const openSettings = () => PanelService.OpenSettings();

// Favorites, Recents and recent searches: local files only, never a provider request (D-09).
// Go never sends null lists; the generated types allow it, so normalize once here.
export type Library = { favorites: Entry[]; recents: Entry[]; searches: Record<string, string[]>; unreadable: string[]; canUndo: boolean };
const library = (s: Snapshot): Library => ({
  favorites: s.favorites ?? [],
  recents: s.recents ?? [],
  searches: (s.searches ?? {}) as Record<string, string[]>,
  unreadable: s.unreadable ?? [],
  canUndo: s.canUndo,
});
export const loadLibrary = async () => library(await LibraryService.Load());
// Each mutation resolves to the fresh lists; failures reject with cause.kind library_read |
// library_write | invalid_item, and the previous lists stay as they were.
export const setFavorite = async (g: Gif, on: boolean) => library(await LibraryService.SetFavorite(g, on));
export const removeRecent = async (g: Gif) => library(await LibraryService.RemoveRecent(g.provider, g.id));
export const clearRecents = async () => library(await LibraryService.ClearRecents());
export const undoLibrary = async () => library(await LibraryService.Undo());
export const saveSearch = (p: string, q: string) => LibraryService.SaveSearch(p, q);
export const removeSearch = async (p: string, q: string) => library(await LibraryService.RemoveSearch(p, q));
