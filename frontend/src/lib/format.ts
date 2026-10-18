// Details pane formatters (UI-SPEC "Details pane"). No Svelte or DOM imports so node --test can load it.
import type { Gif } from './api';

export function fileSize(_bytes?: number): string {
  return '';
}
export function dimensions(_w?: number, _h?: number): string {
  return '';
}
export function linkText(_url: string): string {
  return '';
}
export function tagsSummary(_tags?: string[] | null): string {
  return '';
}
export function providerLine(_gif: Gif, _favorited: boolean): string {
  return '';
}
