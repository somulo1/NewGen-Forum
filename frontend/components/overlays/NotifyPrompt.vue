<!-- Asks once (browsers only allow it from a click) to show Hub's messages, calls and notices as system notifications
     when Hub is in a background tab. "Not now" hides it for 7 days. -->
<template><div v-if="visible" role="dialog" aria-label="Turn on notifications" class="hub-notifyprompt">
<span style="font-size: 22px" aria-hidden="true">🔔</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 14px">Get alerts when Hub is in another tab</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">New messages, calls and notices pop up like WhatsApp, even while you work elsewhere.</span></div>
<div style="display: flex; gap: 6px; flex: none"><button type="button" @click="turnOn" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Turn on</button><button type="button" @click="later" style="height: 34px; padding: 0 12px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Not now</button></div>
</div></template>

<script>
import { supported, permission, ask } from '~/logic/browserNotify.js';
const KEY = 'hub.notifyPromptLater';

export default {
  name: 'NotifyPrompt',
  data() { return { visible: false }; },
  mounted() {
    // shown a little after sign-in, only when the browser can notify and nobody has answered yet
    setTimeout(() => {
      let later = 0; try { later = +localStorage.getItem(KEY) || 0; } catch (e) { /* private window */ }
      this.visible = supported() && permission() === 'default' && Date.now() > later && !!(window.location && !/^\/(auth|s\/|privacy|terms)/.test(window.location.pathname));
    }, 4000);
  },
  methods: {
    async turnOn() { this.visible = false; await ask(); },
    later() { this.visible = false; try { localStorage.setItem(KEY, String(Date.now() + 7 * 86400000)); } catch (e) { /* private window */ } }
  }
};
</script>

<style>
.hub-notifyprompt { position: fixed; left: 20px; bottom: 20px; z-index: 70; width: 440px; max-width: calc(100vw - 28px); box-sizing: border-box; display: flex; align-items: center; gap: 12px; flex-wrap: wrap; padding: 14px 16px; border-radius: 16px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 16px 40px rgba(0,0,0,0.4); color: var(--k-E8EFEA,#E8EFEA); }
@media (max-width: 900px) { .hub-notifyprompt { left: 12px; right: 12px; bottom: 76px; width: auto; } }
</style>
