<!-- Private chat, full screen or fitted to the page (beside the sidebar) -->
<template><template v-if="shown">
<section :aria-label="fit === 'page' ? 'Private chat fit to page' : 'Private chat full screen'" :class="fit === 'page' ? 'hub-fitpage' : 'hub-screen'" :style="rootStyle">
<CallScreen variant="full" :fit="fit" />
<template v-if="!showPanelCall">
<div style="flex: none; border-bottom: 1px solid #24322C; background: #0F1714">
<div style="max-width: 900px; margin: 0 auto; padding: 14px clamp(16px, 3vw, 32px); display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
<span :style="`width: 48px; height: 48px; flex: none; border-radius: 50%; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 16px`">{{ cur.initials }}</span>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ cur.name }}</strong><span style="font-size: 13px; color: #9DB0A6"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }} · only you two can see this</span></div>
<button :aria-expanded="dmPriv.open" @click="dmPriv.toggleOpen" :style="`height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid ${dmPriv.anyOn ? '#FFB86B' : '#2A3832'}; background: ${dmPriv.open ? '#2C3D34' : 'transparent'}; color: ${dmPriv.anyOn ? '#FFB86B' : '#E8EFEA'}; cursor: pointer`">Privacy</button>
<button @click="voiceCall" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Voice call</button>
<button @click="videoCall" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Video call</button>
<template v-if="fit === 'page'"><button @click="toFull" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Full screen</button></template>
<template v-if="fit !== 'page'"><button @click="toPage" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Fit to page</button></template>
<button @click="toDocked" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Side panel</button>
<button @click="toMin" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Minimize</button>
<button aria-label="Close chat" @click="toClosed" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</div>
<div style="max-width: 900px; width: 100%; margin: 0 auto; box-sizing: border-box"><DmPrivacy /></div>
<div style="flex: none; border-bottom: 1px solid #24322C"><div role="tablist" style="max-width: 900px; margin: 0 auto; padding: 0 clamp(8px, 2vw, 24px); display: flex; gap: 4px"><button role="tab" :aria-selected="!isPanelCalls" @click="goPanelChat" :style="`min-height: 40px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${!isPanelCalls ? '#B8F36B' : 'transparent'}; color: ${!isPanelCalls ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${!isPanelCalls ? 600 : 500}`">Chat</button><button role="tab" :aria-selected="isPanelCalls" @click="goPanelCalls" :style="`min-height: 40px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${isPanelCalls ? '#B8F36B' : 'transparent'}; color: ${isPanelCalls ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelCalls ? 600 : 500}`">Calls</button></div></div>
<div v-if="isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 900px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)"><CallHistory /></div></div>
<div v-if="!isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto">
<div style="max-width: 900px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px); display: flex; flex-direction: column; gap: 12px">
<span style="align-self: center; font-size: 11px; color: #9DB0A6; padding: 4px 10px; border-radius: 10px; background: #1A2520">Today</span>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :style="u.rowStyle">
<div aria-hidden="true" :style="`width: 32px; height: 32px; flex: none; border-radius: 50%; background: ${u.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ u.initials }}</div>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ u.time }}</span><template v-if="u.reply"><button type="button" @click="u.reply" aria-label="Reply to this message" style="margin-left: auto; height: 22px; padding: 0 8px; border: 0; border-radius: 11px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px">↩ Reply</button></template></div>
<template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid #B8F36B; border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: #9DB0A6; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: #C9D6CF">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template>
<template v-if="u.hasText"><p style="margin: 4px 0 0; font-size: 14px; line-height: 1.45">{{ u.text }}</p></template>
<template v-if="u.hasFile"><FileGrid :files="u.fileList" :size="140"/></template>
</div>
</div>
</template>
</div>
</div>
<div v-if="!isPanelCalls" style="flex: none; border-top: 1px solid #24322C">
<div v-if="replyBar.active" style="max-width: 900px; margin: 0 auto; padding: 8px clamp(16px, 3vw, 32px) 0"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid #B8F36B; border-radius: 6px; background: #16201C; font-size: 12px; color: #9DB0A6; min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: #E8EFEA">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div></div>
<div style="max-width: 900px; margin: 0 auto; position: relative; padding: 12px clamp(16px, 3vw, 32px); display: flex; gap: 4px; align-items: center">
<template v-if="emojiForumOpen">
<div role="menu" aria-label="Emoji" style="position: absolute; left: 14px; right: 14px; bottom: 66px; z-index: 4; padding: 8px; border-radius: 14px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 12px 30px rgba(0,0,0,0.45); display: grid; grid-template-columns: repeat(8, minmax(0, 1fr)); gap: 2px">
<template v-for="(e, e__i) in emojisForum" :key="e__i"><button :aria-label="e.label" @click="e.add" style="height: 38px; border: 0; border-radius: 8px; background: transparent; font-size: 20px; cursor: pointer">{{ e.ch }}</button></template>
</div>
</template>
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #C9D6CF; cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<button aria-label="Insert emoji" @click="toggleEmojiForum" style="width: 38px; height: 38px; flex: none; border-radius: 50%; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="9"></circle><path d="M8.5 14.5c1 1.2 2.2 1.8 3.5 1.8s2.5-.6 3.5-1.8M9 9.5h.01M15 9.5h.01"></path></svg></button>
<label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Reply in forum</span><input type="text" :value="fdraft" @input="onFDraft" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Message ${cur.name}…`" style="width: 100%; box-sizing: border-box; height: 42px; padding: 0 14px; border-radius: 21px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label>
<button aria-label="Send" :disabled="composerOff" @click="sendForum" style="width: 42px; height: 42px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>

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
  props: { fit: { type: String, default: 'screen' } },
  computed: {
    ...hubBindings(["attachForum","composerNote","composerOff","cur","dmPriv","emojiForumOpen","emojisForum","fKey","fdraft","isFullDm","replyBar","isPanelCalls","goPanelChat","goPanelCalls","isPageDm","onFDraft","sendForum","showPanelCall","toClosed","toDocked","toFull","toMin","toPage","toggleEmojiForum","videoCall","voiceCall"]).computed,
    shown() { return this.fit === 'page' ? !!this.isPageDm : !!this.isFullDm; },
    rootStyle() {
      const base = 'background: #0F1714; display: flex; flex-direction: column; overflow: hidden; box-sizing: border-box;';
      return this.fit === 'page'
        ? base + ' position: absolute; inset: 12px; z-index: 30; border-radius: 22px; border: 1px solid #2A3832; box-shadow: 0 12px 36px rgba(0,0,0,0.35)'
        : base + ' position: fixed; inset: 0; z-index: 40';
    }
  }
};
</script>
