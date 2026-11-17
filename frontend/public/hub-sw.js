// Keeps the Hub pages on the phone so the app opens and moves between pages without a connection. Registered only by the phone
// app. It handles the site's own files (and the fonts) and nothing else: API requests go to the server untouched, and the data
// they return is kept, encrypted, by the page itself.
const VERSION = new URL(self.location.href).searchParams.get('v') || 'app';
const SHELL = 'hub-shell-' + VERSION;
const FONTS = 'hub-fonts';
const HASHED = /^\/_nuxt\/(?!builds\/)/;

self.addEventListener('install', (event) => {
  event.waitUntil((async () => {
    const cache = await caches.open(SHELL);
    let assets = [];
    try { assets = await (await fetch('/_nuxt/offline-assets.json', { cache: 'no-store' })).json(); } catch (e) { /* only the start page then */ }
    const urls = ['/', '/favicon.svg', '/favicon-32.png', '/site.webmanifest'].concat(Array.isArray(assets) ? assets.map((a) => '/_nuxt/' + a) : []);
    // one file that fails does not stop the rest
    await Promise.all(urls.map((u) => cache.add(new Request(u, { cache: 'reload' })).catch(() => {})));
    await self.skipWaiting();
  })());
});

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    for (const name of await caches.keys()) if (name.startsWith('hub-shell-') && name !== SHELL) await caches.delete(name);
    await self.clients.claim();
  })());
});

async function networkFirst(request, fallback) {
  const cache = await caches.open(SHELL);
  try {
    const res = await fetch(request);
    if (res.ok) cache.put(fallback || request, res.clone()).catch(() => {});
    // a server still waking up answers 502/503: the kept copy opens instead, and the page fills in once the server is up
    if (res.status >= 500) return (await cache.match(fallback || request)) || res;
    return res;
  } catch (e) {
    return (await cache.match(fallback || request)) || (await cache.match('/')) || Response.error();
  }
}

async function cacheFirst(name, request) {
  const cache = await caches.open(name);
  const hit = await cache.match(request);
  if (hit) return hit;
  const res = await fetch(request);
  if (res.ok || res.type === 'opaque') cache.put(request, res.clone()).catch(() => {});
  return res;
}

self.addEventListener('fetch', (event) => {
  const { request } = event;
  if (request.method !== 'GET') return;
  const url = new URL(request.url);
  if (url.origin === self.location.origin) {
    // every page of this single-page app is the same start page; the router draws the rest
    if (request.mode === 'navigate') { event.respondWith(networkFirst(request, '/')); return; }
    if (HASHED.test(url.pathname)) { event.respondWith(cacheFirst(SHELL, request)); return; }
    if (url.pathname !== '/hub-sw.js') event.respondWith(networkFirst(request));
    return;
  }
  if (url.hostname === 'fonts.googleapis.com' || url.hostname === 'fonts.gstatic.com') event.respondWith(cacheFirst(FONTS, request));
});
