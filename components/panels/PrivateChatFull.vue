<!-- Private chat in full screen -->
<template><template v-if="isFullDm">
<section aria-label="Private chat full screen" style="position: fixed; inset: 0; z-index: 40; background: #0F1714; display: flex; flex-direction: column; overflow: hidden">
<CallScreen v-if="inCall" />
<template v-if="!inCall" style="display: contents">
<div style="flex: none; border-bottom: 1px solid #24322C; background: #0F1714">
<div style="max-width: 900px; margin: 0 auto; padding: 14px clamp(16px, 3vw, 32px); display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
<span :style="`width: 48px; height: 48px; flex: none; border-radius: 50%; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 16px`">{{ cur.initials }}</span>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ cur.name }}</strong><span style="font-size: 13px; color: #9DB0A6"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }} · only you two can see this</span></div>
<button @click="voiceCall" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Voice call</button>
<button @click="videoCall" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Video call</button>
<button @click="toDocked" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Side panel</button>
<button @click="toMin" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Minimize</button>
<button aria-label="Close chat" @click="toClosed" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto">
<div style="max-width: 900px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px); display: flex; flex-direction: column; gap: 12px">
<span style="align-self: center; font-size: 11px; color: #9DB0A6; padding: 4px 10px; border-radius: 10px; background: #1A2520">Today</span>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :style="u.rowStyle">
<div aria-hidden="true" :style="`width: 32px; height: 32px; flex: none; border-radius: 50%; background: ${u.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ u.initials }}</div>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ u.time }}</span></div>
<template v-if="u.hasText"><p style="margin: 4px 0 0; font-size: 14px; line-height: 1.45">{{ u.text }}</p></template>
<template v-if="u.hasFile"><div style="margin-top: 6px; display: flex; gap: 10px; align-items: center; padding: 8px 10px; border-radius: 10px; background: #121B17; border: 1px solid #2A3832"><span style="width: 32px; height: 32px; flex: none; border-radius: 8px; background: #22302A; color: #B8F36B; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg></span><div style="min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ u.fileName }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ u.fileMeta }}</span></div></div></template>
</div>
</div>
</template>
</div>
</div>
<div style="flex: none; border-top: 1px solid #24322C">
<div style="max-width: 900px; margin: 0 auto; position: relative; padding: 12px clamp(16px, 3vw, 32px); display: flex; gap: 4px; align-items: center">
<template v-if="emojiForumOpen">
<div role="menu" aria-label="Emoji" style="position: absolute; left: 14px; right: 14px; bottom: 66px; z-index: 4; padding: 8px; border-radius: 14px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 12px 30px rgba(0,0,0,0.45); display: grid; grid-template-columns: repeat(8, minmax(0, 1fr)); gap: 2px">
<template v-for="(e, e__i) in emojisForum" :key="e__i"><button :aria-label="e.label" @click="e.add" style="height: 38px; border: 0; border-radius: 8px; background: transparent; font-size: 20px; cursor: pointer">{{ e.ch }}</button></template>
</div>
</template>
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #C9D6CF; cursor: pointer"><input type="file" multiple="" aria-label="Attach files" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<button aria-label="Insert emoji" @click="toggleEmojiForum" style="width: 38px; height: 38px; flex: none; border-radius: 50%; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="9"></circle><path d="M8.5 14.5c1 1.2 2.2 1.8 3.5 1.8s2.5-.6 3.5-1.8M9 9.5h.01M15 9.5h.01"></path></svg></button>
<label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Reply in forum</span><input type="text" :value="fdraft" @input="onFDraft" @keydown="fKey" :placeholder="`Message ${cur.name}…`" style="width: 100%; box-sizing: border-box; height: 42px; padding: 0 14px; border-radius: 21px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label>
<button aria-label="Send" @click="sendForum" style="width: 42px; height: 42px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>

</div>
</div>
</template>
</section>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';

export default {
  name: 'PrivateChatFull',
  components: { CallScreen },
  ...hubBindings(["attachForum","cur","emojiForumOpen","emojisForum","fKey","fdraft","inCall","isFullDm","onFDraft","sendForum","toClosed","toDocked","toMin","toggleEmojiForum","videoCall","voiceCall"])
};
</script>
