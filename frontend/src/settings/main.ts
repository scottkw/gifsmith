import { mount } from 'svelte';
import { Events } from '@wailsio/runtime';
import '../app.css';
import GeneralPane from './GeneralPane.svelte';
import { snapshot, type SettingsSnapshot } from '../lib/api';

function setTheme(theme: string) {
  if (theme === 'light' || theme === 'dark') document.documentElement.dataset.theme = theme;
  else delete document.documentElement.dataset.theme;
}

const s = await snapshot();
setTheme(s.theme);
Events.On('settings:changed', (e) => setTheme((e.data as SettingsSnapshot).theme));
mount(GeneralPane, { target: document.getElementById('app')!, props: { initial: s } });
