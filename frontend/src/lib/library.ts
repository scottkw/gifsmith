// skeleton (RED)
import type { Gif } from './api';
import type { SlotId } from './providers';
export type Section = { provider: SlotId; label: string; items: Gif[]; start: number };
export type Row = { kind: 'header'; section: number; size: number } | { kind: 'tiles'; section: number; start: number; size: number };
export function sections(_gifs: Gif[], _filter: string): Section[] {
  return [];
}
export function flatItems(_secs: Section[]): Gif[] {
  return [];
}
export function rows(_secs: Section[]): Row[] {
  return [];
}
