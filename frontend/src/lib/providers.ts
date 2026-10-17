// The four fixed provider slots (UI-SPEC Provider Tabs Contract). ⌘1–⌘4 always map to the same slot.
export type SlotId = 'giphy' | 'clips' | 'klipy' | 'finergifs';

export type Slot = {
  id: SlotId;
  label: string;
  shortcut: string;
  placeholder: string;
  attribution: 'giphy' | 'klipy' | 'finer';
  keyProvider?: 'giphy' | 'klipy'; // set when this slot shows the setup card without its own key
  keyLabel?: string;
  getKeyUrl?: string;
};

export const GIPHY_KEY_URL = 'https://developers.giphy.com/dashboard/';
export const KLIPY_KEY_URL = 'https://partner.klipy.com';
// C2: GIPHY grants Clips access by email (D-21); pre-addressed with a subject.
export const CLIPS_ACCESS_MAILTO = 'mailto:clips@giphy.com?subject=GIPHY%20Clips%20API%20access%20request';

export const SLOTS: readonly Slot[] = [
  { id: 'giphy', label: 'GIPHY', shortcut: '⌘1', placeholder: 'Search GIPHY', attribution: 'giphy', keyProvider: 'giphy', keyLabel: 'GIPHY', getKeyUrl: GIPHY_KEY_URL },
  { id: 'clips', label: 'Clips', shortcut: '⌘2', placeholder: 'Search GIPHY Clips', attribution: 'giphy' },
  { id: 'klipy', label: 'Klipy', shortcut: '⌘3', placeholder: 'Search KLIPY', attribution: 'klipy', keyProvider: 'klipy', keyLabel: 'Klipy', getKeyUrl: KLIPY_KEY_URL },
  { id: 'finergifs', label: 'Finer Gifs', shortcut: '⌘4', placeholder: 'Search The Finer Gifs Club', attribution: 'finer' },
];

export const slotById = (id: string): Slot => SLOTS.find((s) => s.id === id) ?? SLOTS[0];
