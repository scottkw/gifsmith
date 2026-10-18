import { mount } from 'svelte'
import './app.css'
import App from './App.svelte'
import { snapshot } from './lib/api'
import { applyTheme } from './lib/theme'

applyTheme((await snapshot()).theme) // before mount: no white flash in dark mode
mount(App, { target: document.getElementById('app')! })
