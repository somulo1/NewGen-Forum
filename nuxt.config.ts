// Hub — Nuxt port of the Claude Design prototype.
// ssr:false because the prototype logic reads window/document; switch on later per page if needed.
export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  components: [{ path: '~/components', pathPrefix: false }],
  css: ['~/assets/hub.css'],
  runtimeConfig: { public: { apiBase: '' } }, // set NUXT_PUBLIC_API_BASE when the backend exists
  vue: { compilerOptions: { whitespace: 'preserve' } },
  app: {
    head: {
      title: 'Hub',
      htmlAttrs: { lang: 'en' },
      meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }, { name: 'color-scheme', content: 'dark' }, { name: 'theme-color', content: '#0F1714' }],
      style: [{ innerHTML: 'html,body,#__nuxt{margin:0;background:#0F1714;color:#E8EFEA;color-scheme:dark}' }],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@500;600;700&family=DM+Sans:wght@400;500;600&display=swap' }
      ]
    }
  },
  compatibilityDate: '2024-10-01'
});
