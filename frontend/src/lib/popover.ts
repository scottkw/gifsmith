// Action menu placement (UI-SPEC "Action menu placement"). No Svelte or DOM imports so node --test can load it.
export const MENU_W = 320;
const GAP = 8; // anchor offset and window inset
const FOOTER = 32;

// The menu's bottom stays GAP above the footer (P19: y 440 in the 480px panel) so attribution shows.
const maxBottom = (h: number) => h - FOOTER - GAP;

// Right of the tile when it fits, else left of it, for 3, 4 or 5 columns; top aligns with the tile.
// Then everything is clamped GAP inside the window and above the footer.
export function placeMenu(
  tile: { x: number; y: number; width: number; height: number },
  menuH: number,
  win = { w: 720, h: 480 },
): { left: number; top: number } {
  const right = tile.x + tile.width + GAP;
  const left = right + MENU_W <= win.w - GAP ? right : tile.x - GAP - MENU_W;
  const clamp = (v: number, max: number) => Math.max(GAP, Math.min(v, max));
  return { left: clamp(left, win.w - GAP - MENU_W), top: clamp(tile.y, maxBottom(win.h) - menuH) };
}

// In Details the menu hangs 16px inside the content area's top-right corner (content top = search 48 + tabs 36).
const CONTENT_TOP = 84;
const DETAILS_INSET = 16;
export function placeInDetails(menuH: number): { left: number; top: number } {
  return { left: 720 - DETAILS_INSET - MENU_W, top: Math.max(GAP, Math.min(CONTENT_TOP + DETAILS_INSET, maxBottom(480) - menuH)) };
}
