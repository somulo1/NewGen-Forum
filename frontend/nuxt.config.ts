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
      script: [{ innerHTML: "if (/^\\/(login|signup|reset-password)\\/?$/.test(location.pathname)) { var s = document.createElement('style'); s.textContent = 'html,body,#__nuxt{background:#F2F6F3 !important;color:#0F1714 !important;color-scheme:light !important}'; document.head.appendChild(s); }", tagPosition: 'head' }],
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
  compatibilityDate: '2024-10-01'
});
