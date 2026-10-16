<script setup>
import { LEGAL_DOCS } from '~/logic/legal.js';
definePageMeta({ path: '/legal', layout: 'legal' });
useHead({ title: 'Legal · Hub' });
const docs = LEGAL_DOCS;
</script>
<template>
  <div style="display: flex; flex-direction: column; gap: 18px">
    <h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: clamp(26px, 5vw, 36px); letter-spacing: -0.02em">Legal</h1>
    <NuxtLink v-for="d in docs" :key="d.slug" :to="d.path" style="padding: 16px 18px; border-radius: 16px; background: #fff; border: 1px solid #DCE5DF; text-decoration: none; color: inherit; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 16px; color: #2F7A12">{{ d.title }}</strong><span style="font-size: 14px; color: #55645C">{{ d.summary }}</span></NuxtLink>
  </div>
</template>
