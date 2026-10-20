// The theme pref ("" = System, "light", "dark") resolved to what <html data-theme> holds.
export type Resolved = 'light' | 'dark';

export function resolveTheme(pref: string, systemDark: boolean): Resolved {
  if (pref === 'light' || pref === 'dark') return pref;
  return systemDark ? 'dark' : 'light';
}

let stopFollowing = () => {};
const listeners = new Set<(t: Resolved) => void>();

// Sets data-theme now; while pref is System it follows macOS live (one listener, replaced per call).
export function applyTheme(pref: string) {
  stopFollowing();
  const mq = matchMedia('(prefers-color-scheme: dark)');
  const set = () => {
    const t = resolveTheme(pref, mq.matches);
    document.documentElement.dataset.theme = t;
    listeners.forEach((f) => f(t));
  };
  set();
  if (pref !== 'light' && pref !== 'dark') {
    mq.addEventListener('change', set);
    stopFollowing = () => mq.removeEventListener('change', set);
  } else stopFollowing = () => {};
}

export const resolvedTheme = (): Resolved => (document.documentElement.dataset.theme === 'dark' ? 'dark' : 'light');

// For script-chosen images (the GIPHY mark): called with each newly applied theme; returns unsubscribe.
export function onTheme(f: (t: Resolved) => void) {
  listeners.add(f);
  return () => listeners.delete(f);
}
