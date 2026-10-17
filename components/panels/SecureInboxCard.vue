<template>
<article style="padding: 14px; border-radius: 14px; background: #17231D; border: 1px solid #6EE7C8; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#6EE7C8" stroke-width="1.8" stroke-linejoin="round"><rect x="4" y="10" width="16" height="11" rx="2"></rect><path d="M8 10V7a4 4 0 0 1 8 0v3"></path></svg><strong style="flex: 1; min-width: 0; overflow-wrap: anywhere">{{ share.title }}</strong><span style="font-size: 11px; color: #9DB0A6">Secure share</span></div>
<template v-if="share.showingSecret"><pre style="margin: 0; padding: 12px; border-radius: 10px; background: #0F1714; color: #E8EFEA; white-space: pre-wrap; overflow-wrap: anywhere; font: inherit">{{ share.secret }}</pre><span style="font-size: 12px; color: #FFB86B">This secret is visible for 10 seconds and can only be opened once.</span></template>
<template v-else-if="share.unavailable"><span role="status" style="font-size: 13px; color: #FFB86B">{{ share.expired ? 'This share expired unopened.' : 'This share has already been opened and destroyed.' }}</span></template>
<template v-else><span style="font-size: 13px; color: #C9D6CF">Enter the password shared with you. This link expires if unopened within 24 hours.</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><input type="password" :value="share.passwordValue" @input="share.onPassword" :aria-label="`Password for ${share.title}`" placeholder="Link password" style="flex: 1 1 180px; min-width: 0; height: 38px; box-sizing: border-box; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"><button type="button" @click="share.unlock" style="height: 38px; padding: 0 14px; border: 0; border-radius: 10px; background: #6EE7C8; color: #0F1714; font-weight: 600; cursor: pointer">Open once</button></div></template>
</article>
</template>

<script>
export default { name: 'SecureInboxCard', props: { share: { type: Object, required: true } } };
</script>
