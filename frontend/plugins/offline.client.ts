import { keepShell } from '~/logic/offline/index.js';

// Inside the Hub phone app the pages are kept on the phone for offline use; in a browser this does nothing.
export default defineNuxtPlugin(() => { keepShell(); });
