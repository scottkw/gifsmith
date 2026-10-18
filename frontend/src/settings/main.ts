import { mount } from 'svelte';
import { Events } from '@wailsio/runtime';
import '../app.css';
import GeneralPane from './GeneralPane.svelte';
import { snapshot, type SettingsSnapshot } from '../lib/api';
import { applyTheme } from '../lib/theme';

const s = await snapshot();
applyTheme(s.theme); // before mount: no white flash in dark mode
Events.On('settings:changed', (e) => applyTheme((e.data as SettingsSnapshot).theme));
mount(GeneralPane, { target: document.getElementById('app')!, props: { initial: s } });
