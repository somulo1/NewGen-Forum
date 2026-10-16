<!-- Legal documents: readable single column with links to every policy. Open to signed-out visitors. -->
<template>
  <div style="min-height: 100vh; background: #F2F6F3; color: #0F1714; font-family: 'DM Sans', system-ui, sans-serif; font-size: 15px; line-height: 1.6">
    <header style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px 16px; padding: 14px clamp(16px, 4vw, 40px); background: #fff; border-bottom: 1px solid #DCE5DF">
      <NuxtLink to="/" aria-label="Hub home" style="display: flex; align-items: center; gap: 10px; text-decoration: none; color: inherit"><span aria-hidden="true" style="width: 34px; height: 34px; border-radius: 9px; background: #B8F36B; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 17px">H</span><span style="font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 18px">Hub</span></NuxtLink>
      <nav aria-label="Legal documents" style="display: flex; flex-wrap: wrap; gap: 4px 14px; margin-left: auto; font-size: 14px"><NuxtLink v-for="d in docs" :key="d.slug" :to="d.path" style="color: #2F7A12; text-decoration: none; font-weight: 600" active-class="legal-on">{{ d.title }}</NuxtLink></nav>
    </header>
    <main style="max-width: 780px; margin: 0 auto; padding: clamp(20px, 5vw, 48px) clamp(16px, 4vw, 24px) 64px"><slot /></main>
    <footer style="padding: 20px; text-align: center; font-size: 12px; color: #6B7A72; border-top: 1px solid #DCE5DF">Hub Platforms Limited · Westlands Business Park, Nairobi</footer>
  </div>
</template>

<script setup>
import { LEGAL_DOCS } from '~/logic/legal.js';
const docs = LEGAL_DOCS;
</script>

<style>
.legal-on { text-decoration: underline !important; }
</style>
