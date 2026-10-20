// Hub — Nuxt port of the Claude Design prototype.
// ssr:false because the prototype logic reads window/document; switch on later per page if needed.
export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  components: [{ path: '~/components', pathPrefix: false }],
  css: ['~/assets/hub.css', '~/assets/theme-light.css'],
  runtimeConfig: { public: { apiBase: 'http://localhost:8080' } }, // override with NUXT_PUBLIC_API_BASE
  vue: { compilerOptions: { whitespace: 'preserve' } },
  app: {
    head: {
      title: 'Hub',
      htmlAttrs: { lang: 'en' },
      meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }, { name: 'color-scheme', content: 'dark' }, { name: 'theme-color', content: '#0F1714' }],
      style: [{ innerHTML: 'html,body,#__nuxt{margin:0;background:#0F1714;color:#E8EFEA;color-scheme:dark}' }],
      script: [{ innerHTML: "try { if (localStorage.getItem('hub-theme') === 'light') { var t = document.createElement('style'); t.textContent = 'html,body,#__nuxt{background:#E4ECE9 !important;color:#0F1714 !important;color-scheme:light !important}'; document.head.appendChild(t); } } catch (e) {} if (/^\\/(login|signup|reset-password)\\/?$/.test(location.pathname)) { var s = document.createElement('style'); s.textContent = 'html,body,#__nuxt{background:#F2F6F3 !important;color:#0F1714 !important;color-scheme:light !important}'; document.head.appendChild(s); }", tagPosition: 'head' }],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/favicon.svg' },
        { rel: 'icon', type: 'image/png', sizes: '32x32', href: '/favicon-32.png' },
        { rel: 'icon', href: '/favicon.ico', sizes: 'any' },
        { rel: 'apple-touch-icon', href: '/apple-touch-icon.png' },
        { rel: 'manifest', href: '/site.webmanifest' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@500;600;700&family=DM+Sans:wght@400;500;600&display=swap' }
      ]
    }
  },
  // The Hub logic is one shared instance, so Vue's component-by-component hot update can leave a docked panel pointing at a
  // torn-down child ("can't access property emitsOptions of null"). In dev, reload the page instead of hot-patching.
  vite: {
    // Pre-bundle deps Vite would otherwise discover lazily (each discovery triggers a full page reload in dev).
    optimizeDeps: {
      include: ['livekit-client', '@noble/curves/ed25519', '@noble/hashes/hkdf', '@noble/hashes/hmac', '@noble/hashes/sha256', 'qrcode']
    },
    plugins: [{
      name: 'hub-full-reload',
      apply: 'serve',
      handleHotUpdate({ file, server }) {
        if (!/\/(components|layouts|pages|logic|composables)\/.*\.(vue|js|ts)$/.test(file)) return;
        server.ws.send({ type: 'full-reload' });
        return [];
      }
    }]
  },
  // Browser protections for the production site (the dev server is left alone). The content policy is report-only first: it lists
  // what a page may load and connect to; switch it to enforcing once the browser console shows no violations on your pages.
  $production: {
    routeRules: {
      '/**': {
        headers: {
          'X-Content-Type-Options': 'nosniff',
          'X-Frame-Options': 'DENY',
          'Referrer-Policy': 'no-referrer',
          'Permissions-Policy': 'camera=(self), microphone=(self), display-capture=(self), geolocation=(), payment=()',
          'Content-Security-Policy-Report-Only': "default-src 'self'; script-src 'self' 'unsafe-inline' 'wasm-unsafe-eval'; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; font-src 'self' https://fonts.gstatic.com; img-src 'self' data: blob:; media-src 'self' blob:; connect-src 'self' https: wss: blob:; worker-src 'self' blob:; frame-ancestors 'none'; base-uri 'self'; form-action 'self'; object-src 'none'"
        }
      }
    }
  },
  compatibilityDate: '2024-10-01'
});
