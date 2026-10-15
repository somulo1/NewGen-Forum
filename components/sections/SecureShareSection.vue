<!-- Secure share (one-time links) -->
<template><template v-if="show.share">
<div style="max-width: 620px; display: flex; flex-direction: column; gap: 16px">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Secure share</h1>
<p style="margin: -8px 0 4px; color: #9DB0A6; line-height: 1.5">Share a password or credential through a one-time link. The receiver needs the password you set, and the link is destroyed the moment it's opened.</p>
<div style="padding: 20px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">What is it?<input type="text" placeholder="e.g. Staging server login" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Secret<textarea rows="3" placeholder="Username, password, API key…" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; resize: vertical"></textarea></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Link password<input type="password" placeholder="Share this separately" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<span style="padding: 6px 12px; border-radius: 16px; background: #22302A; font-size: 13px">Expires after 1 view</span>
<span style="padding: 6px 12px; border-radius: 16px; background: #22302A; font-size: 13px">Or after 24 hours unopened</span>
</div>
<button @click="genLink" style="height: 44px; border-radius: 12px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 15px; cursor: pointer">Generate one-time link</button>
<template v-if="linkMade">
<div style="padding: 14px; border-radius: 12px; border: 1px solid #B8F36B; background: #1A2520; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; gap: 10px; align-items: center"><code style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; font-size: 14px; color: #D4FA9E">hub.[yourdomain]/s/7Kq2-xR9p</code><button style="height: 36px; padding: 0 14px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Copy</button></div>
<span style="font-size: 12px; color: #9DB0A6">Self-destructs on first open. You'll be notified when it's used.</span>
</div>
</template>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'SecureShareSection',
  ...hubBindings(["genLink","linkMade","show"])
};
</script>
