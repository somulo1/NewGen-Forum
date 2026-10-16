<!-- Error page (404, 403, 500): a short message, centered in the content area, then back to the previous page -->
<template>
<div role="alert" style="width: 100%; min-height: min(70vh, 640px); display: flex; align-items: center; justify-content: center; padding: 24px 0">
<div style="width: 560px; max-width: 100%; box-sizing: border-box; padding: clamp(24px, 4vw, 36px); border-radius: 24px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; align-items: center; gap: 12px; text-align: center">
<span :style="`width: 64px; height: 64px; border-radius: 20px; background: ${c.color}22; color: ${c.color}; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path :d="c.icon"></path></svg></span>
<span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 64px; font-weight: 700; line-height: 1; letter-spacing: -0.03em; color: ${c.color}`">{{ c.code }}</span>
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ c.title }}</h1>
<p style="margin: 0; font-size: 15px; line-height: 1.6; color: #C9D6CF; max-width: 400px">{{ c.message }}</p>
<template v-if="c.ref"><span style="font-size: 12px; color: #6F8178">Reference {{ c.ref }} · {{ stamp }}</span></template>
<div style="display: flex; flex-wrap: wrap; gap: 8px; justify-content: center; margin-top: 4px">
<button @click="back" style="height: 42px; padding: 0 22px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Go back now</button>
<template v-if="code === '500'"><button @click="retry" style="height: 42px; padding: 0 20px; border-radius: 21px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Try again</button></template>
<button @click="home" style="height: 42px; padding: 0 20px; border-radius: 21px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Go to Home</button>
</div>
<div style="font-size: 13px; color: #9DB0A6; display: flex; align-items: center; gap: 10px">
<template v-if="running"><span>Taking you back in <strong style="color: #E8EFEA">{{ left }}</strong> second{{ left === 1 ? '' : 's' }}…</span><button @click="stay" style="border: 0; background: transparent; color: #B8F36B; cursor: pointer; font-size: 13px; text-decoration: underline">Stay here</button></template>
<template v-if="!running"><span>Automatic redirect cancelled.</span></template>
</div>
</div>
</div>
</template>

<script>
const SECONDS = 10;
export default {
  name: 'ErrorPage',
  props: { code: { type: String, default: '404' }, section: { type: String, default: '' }, reason: { type: String, default: 'rights' } },
  emits: ['leave'],
  data() { return { left: SECONDS, running: true, stamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit' }) }; },
  computed: {
    c() {
      const page = this.section ? '"' + this.section + '"' : 'this page';
      if (this.code === '403') return { code: '403', color: '#FFB86B', icon: 'M7 11V8a5 5 0 0 1 10 0v3M5 11h14v9H5z', title: 'No access', message: 'You don\'t have permission to open ' + page + '. Ask an admin if you need it.', ref: '' };
      if (this.code === '500') return { code: '500', color: '#FF9B8A', icon: 'M12 8v5M12 17h.01M10.3 3.9L2.4 18a2 2 0 0 0 1.7 3h15.8a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z', title: 'Something went wrong', message: 'We couldn\'t load this page. Try again in a moment.', ref: 'ERR-' + (Date.now() % 100000).toString(36).toUpperCase() };
      if (this.reason === 'off') return { code: '404', color: '#7CC4FF', icon: 'M4 5h16v14H4zM4 9h16M9 9v10', title: 'Not set up yet', message: page.charAt(0).toUpperCase() + page.slice(1) + ' isn\'t switched on for your organization.', ref: '' };
      return { code: '404', color: '#7CC4FF', icon: 'M11 4a7 7 0 1 0 0 14 7 7 0 0 0 0-14zM20 20l-4-4M9 9a2 2 0 1 1 3 1.7c-.6.4-1 .9-1 1.6', title: 'Page not found', message: 'This page doesn\'t exist.', ref: '' };
    }
  },
  mounted() { this.timer = setInterval(() => { if (!this.running) return; this.left -= 1; if (this.left <= 0) this.back(); }, 1000); },
  beforeUnmount() { clearInterval(this.timer); },
  methods: {
    stay() { this.running = false; },
    leave() { clearInterval(this.timer); this.$emit('leave'); },
    back() {
      this.leave();
      const prev = typeof window !== 'undefined' && window.history.state ? window.history.state.back : null;
      const bad = /^\/(403|404|500)\/?$/;
      this.$router.push(prev && typeof prev === 'string' && !bad.test(prev) ? prev : '/');
    },
    home() { this.leave(); this.$router.push('/'); },
    retry() { this.leave(); if (typeof window !== 'undefined') window.location.reload(); }
  }
};
</script>
