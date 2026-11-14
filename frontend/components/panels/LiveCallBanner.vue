<!-- A voice or video call is in progress in this chat or forum and I am not in it: a glowing bar with Join -->
<template><div v-if="curLive" class="hub-livebar" role="status">
<span class="hub-livebar-dot hub-live-glow" aria-hidden="true"></span>
<svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path v-if="curLive.video" d="M3 6h13v12H3zM16 10l5-3v10l-5-3z"></path><path v-else d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"></path></svg>
<span class="hub-livebar-text">{{ curLive.text }}</span>
<button type="button" class="hub-live-glow" @click="curLive.join">Join</button>
</div></template>

<script>
import { hubBindings } from '~/composables/useHub';
export default { name: 'LiveCallBanner', ...hubBindings(['curLive']) };
</script>

<style>
.hub-livebar { flex: none; display: flex; align-items: center; gap: 10px; margin: 8px 12px; padding: 8px 8px 8px 14px; border-radius: 14px; background: rgba(184,243,107,0.12); border: 1px solid #B8F36B; color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; font-weight: 600; }
.hub-livebar-dot { width: 10px; height: 10px; flex: none; border-radius: 50%; background: #B8F36B; }
.hub-livebar-text { flex: 1; min-width: 0; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.hub-livebar button { flex: none; height: 32px; padding: 0 16px; border-radius: 16px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; cursor: pointer; }
</style>
