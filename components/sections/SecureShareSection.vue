<!-- Create a one-time secret share for a private chat or copy its link. -->
<template>
<template v-if="show.share">
<div style="max-width: 680px; display: flex; flex-direction: column; gap: 16px">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Secure share</h1>
<p style="margin: -8px 0 4px; color: var(--k-9DB0A6,#9DB0A6); line-height: 1.5">Send a secret to a colleague's private chat or copy a one-time link. The recipient needs the password you set. A secret is visible for 10 seconds after opening and expires if unopened for 24 hours.</p>
<div style="padding: 20px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">What is it?<input type="text" :value="secureShare.draft.title" @input="secureShare.onTitle" placeholder="e.g. Staging server login" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Secret<textarea rows="4" :value="secureShare.draft.secret" @input="secureShare.onSecret" placeholder="Username, password, API key…" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; resize: vertical"></textarea></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Link password<input type="password" :value="secureShare.draft.password" @input="secureShare.onPassword" placeholder="Share this separately" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 8px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Send to a colleague<input type="search" :value="secureShare.draft.search" @input="secureShare.onSearch" placeholder="Search by name…" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div v-if="secureShare.draft.search && !secureShare.draft.recipient" style="display: flex; flex-direction: column; gap: 4px; padding: 6px; border: 1px solid var(--k-24322C,#24322C); border-radius: 10px; background: var(--k-121B17,#121B17)">
<template v-for="person in secureShare.recipients" :key="person.name"><button type="button" @click="secureShare.selectRecipient(person.name)" style="padding: 9px 10px; border: 0; border-radius: 7px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left">{{ person.name }}<span style="display: block; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ person.title }}</span></button></template>
<span v-if="!secureShare.recipients.length" style="padding: 8px; color: var(--k-9DB0A6,#9DB0A6); font-size: 13px">No matching colleagues.</span>
</div>
<div v-if="secureShare.draft.recipient" style="display: flex; align-items: center; gap: 8px; padding: 8px 10px; border-radius: 10px; background: var(--k-1A2520,#1A2520)"><span style="flex: 1; font-size: 13px">Sending privately to <strong>{{ secureShare.draft.recipient }}</strong></span><button type="button" @click="secureShare.clearRecipient" aria-label="Change recipient" style="border: 0; background: transparent; color: var(--k-B8F36B,#B8F36B); cursor: pointer">Change</button></div>
</div>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">One open only · 10 seconds visible · expires unopened after 24 hours</span>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<button v-if="secureShare.draft.recipient" type="button" @click="secureShare.create(true)" style="flex: 1 1 210px; height: 44px; border-radius: 12px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 14px; cursor: pointer">Send in private chat</button>
<button type="button" @click="secureShare.create(false)" style="flex: 1 1 210px; height: 44px; border-radius: 12px; border: 1px solid var(--k-6EE7C8,#6EE7C8); background: transparent; color: var(--k-6EE7C8,#6EE7C8); font-weight: 600; font-size: 14px; cursor: pointer">Copy one-time link</button>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'SecureShareSection',
  ...hubBindings(['secureShare', 'show'])
};
</script>
