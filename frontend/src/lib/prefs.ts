// Search-pane option lists (UI-SPEC "Search"): Go owns the allowed codes, the webview names them.
export type Option = { value: string; label: string };

export function languageOptions(codes: string[], system: string): Option[] {
  return [];
}

export function countryOptions(codes: string[], system: string): Option[] {
  return [];
}

export function ratingLabel(v: string): string {
  return v;
}
