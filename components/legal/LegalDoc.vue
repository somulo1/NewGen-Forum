<!-- One legal document from logic/legal.js, plus links to the others. -->
<template>
<article v-if="doc" style="display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-direction: column; gap: 6px">
<button type="button" @click="back" style="align-self: flex-start; border: 0; background: transparent; color: #2F7A12; cursor: pointer; font-size: 14px; padding: 0">‹ Back</button>
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: clamp(26px, 5vw, 36px); letter-spacing: -0.02em">{{ doc.title }}</h1>
<span style="font-size: 13px; color: #6B7A72">Last updated {{ doc.updated }}</span>
</div>
<nav aria-label="On this page" style="display: flex; flex-wrap: wrap; gap: 6px"><a v-for="(s, i) in doc.sections" :key="i" :href="`#s${i}`" style="padding: 4px 12px; border-radius: 14px; background: #fff; border: 1px solid #DCE5DF; color: #33443B; font-size: 12px; text-decoration: none">{{ s.h.replace(/^\d+\.\s*/, '') }}</a></nav>
<section v-for="(s, i) in doc.sections" :id="`s${i}`" :key="i" style="padding: 18px 20px; border-radius: 16px; background: #fff; border: 1px solid #DCE5DF; display: flex; flex-direction: column; gap: 10px; scroll-margin-top: 16px">
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 18px">{{ s.h }}</h2>
<p v-for="(t, j) in (s.p || [])" :key="j" style="margin: 0; color: #33443B">{{ t }}</p>
<ul v-if="s.list" style="margin: 0; padding-left: 20px; display: flex; flex-direction: column; gap: 6px; color: #33443B"><li v-for="(t, j) in s.list" :key="j">{{ t }}</li></ul>
</section>
<p style="margin: 0; font-size: 12px; color: #6B7A72">This document is a starting draft. A lawyer will review it before relying on it.</p>
<div style="display: flex; flex-direction: column; gap: 8px; padding-top: 8px; border-top: 1px solid #DCE5DF"><strong style="font-size: 14px">Other documents</strong>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><NuxtLink v-for="d in others" :key="d.slug" :to="d.path" style="padding: 6px 14px; border-radius: 16px; border: 1px solid #C4D0C9; background: #fff; color: #2F7A12; font-weight: 600; font-size: 13px; text-decoration: none">{{ d.title }}</NuxtLink></div></div>
</article>
</template>

<script setup>
import { LEGAL_DOCS, docBySlug } from '~/logic/legal.js';
const props = defineProps({ slug: { type: String, required: true } });
const doc = computed(() => docBySlug(props.slug));
const others = computed(() => LEGAL_DOCS.filter((d) => d.slug !== props.slug));
useHead(() => ({ title: (doc.value ? doc.value.title : 'Legal') + ' · Hub' }));
const router = useRouter();
const back = () => { if (window.history.length > 1) router.back(); else navigateTo('/'); };
</script>
