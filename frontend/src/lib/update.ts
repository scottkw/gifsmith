// Updates copy (UI-SPEC "Updates": U2-U7, U-E1-U-E4, T12). Pure: no DOM or runtime.
import type { UpdateStatus } from './api';

export function updateText(st: UpdateStatus): string {
  switch (st.state) {
    case 'checking':
      return 'Checking for updates…';
    case 'current':
      return `Gifsmith ${st.current} is up to date.`;
    case 'available':
      return `Gifsmith ${st.latest} is available. You have ${st.current}.`;
    case 'downloading':
      return `Downloading Gifsmith ${st.latest}…` + (st.total > 0 ? ` ${Math.floor((st.done * 100) / st.total)}%` : '');
    case 'verifying':
      return "Checking the download's signature…";
    case 'opened':
      return `The Gifsmith ${st.latest} disk image is open. Quit Gifsmith, then drag the new Gifsmith to Applications and choose Replace.`;
  }
  return '';
}

export function updateErrorText(kind: string, st: UpdateStatus): string {
  switch (kind) {
    case 'update_download_failed':
      return `Couldn't download Gifsmith ${st.latest}. Check your connection, then try again.`;
    case 'update_verify_failed':
      return "The download didn't pass the signature check, so Gifsmith deleted it. Nothing was opened.";
    case 'update_open_failed':
      return "Gifsmith couldn't open the downloaded disk image. Try again.";
  }
  return "Couldn't check for updates. Check your connection, then try again.";
}

export function updateToastText(v: string): string {
  return `Gifsmith ${v} is available.`;
}
