// Hub — Nuxt port of the Claude Design prototype.
// ssr:false because the prototype logic reads window/document; switch on later per page if needed.
export default defineNuxtConfig({
  ssr: false,
  devtools: { enabled: false },
  components: [{ path: '~/components', pathPrefix: false }],
  css: ['~/assets/hub.css'],
  vue: { compilerOptions: { whitespace: 'preserve' } },
  app: {
    head: {
      title: 'Hub',
      htmlAttrs: { lang: 'en' },
      meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Space+Grotesk:wght@500;600;700&family=DM+Sans:wght@400;500;600&display=swap' }
      ]
    }
  },
  compatibilityDate: '2024-10-01'
});
