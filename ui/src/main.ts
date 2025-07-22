import {mount} from 'svelte';
import './app.css';
import App from './App.svelte';

import {type NMEALoggerData} from './types.js';

import 'leaflet/dist/leaflet.css';

declare global {
    interface Window {
        nmeaLogger: NMEALoggerData;
    }
}

const app = mount(App, {
    target: document.getElementById('app')!,
});

export default app;
