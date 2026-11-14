<!-- Private chat, full screen or fitted to the page (beside the sidebar) -->
<template><template v-if="shown">
<section :aria-label="fit === 'page' ? 'Private chat fit to page' : 'Private chat full screen'" :class="fit === 'page' ? 'hub-fitpage' : 'hub-screen'" :style="rootStyle">
<ChatRail :open="railOpen" @close="railOpen = false" kind="dm" />
<div style="position: relative; flex: 1; min-width: 0; min-height: 0; display: flex; flex-direction: column; overflow: hidden">
<CallScreen variant="full" :fit="fit" />
<template v-if="!showPanelCall">
<div style="position: relative; flex: none; border-bottom: 1px solid var(--k-24322C,#24322C); background: var(--k-0F1714,#0F1714)">
<div class="hub-chat-bar" style="margin: 0; padding: 8px clamp(16px, 3vw, 32px); padding-right: 58px; display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px">
<UserAvatar :name="cur.name" :initials="cur.initials" :color="cur.color" :size="34" :font-size="14"/>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-family: 'Space Grotesk', sans-serif; font-size: 17px">{{ cur.name }}</strong><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }}</span></div>
<div role="tablist" aria-label="Chat views" class="hub-seg" style="display: inline-flex; flex: none; padding: 3px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832)"><button v-for="v in views" :key="v.k" role="tab" :aria-selected="v.on" @click="v.go" :style="`height: 30px; padding: 0 14px; border: 0; border-radius: 9px; cursor: pointer; font-size: 13px; font-weight: 600; white-space: nowrap; background: ${v.on ? '#B8F36B' : 'transparent'}; color: ${v.on ? '#0F1714' : 'var(--k-C9D6CF,#C9D6CF)'}`">{{ v.label }}</button></div>
<span class="hub-sep" aria-hidden="true" style="width: 1px; height: 24px; flex: none; background: var(--k-2A3832,#2A3832)"></span>
<button :aria-expanded="dmPriv.open" aria-label="Privacy and blocking" title="Privacy and blocking" @click="dmPriv.toggleOpen" :style="`display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid ${dmPriv.anyOn ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-2A3832,#2A3832)'}; background: ${dmPriv.open ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: ${dmPriv.anyOn ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-E8EFEA,#E8EFEA)'}; cursor: pointer`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z"></path></svg></button>
<button v-if="canVoiceCall" @click="voiceCall" aria-label="Voice call" title="Voice call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: #B8F36B; color: #0F1714; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1-2-2z"></path></svg></button>
<button v-if="canVideoCall" @click="videoCall" aria-label="Video call" title="Video call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<span class="hub-sep" aria-hidden="true" style="width: 1px; height: 24px; flex: none; background: var(--k-2A3832,#2A3832)"></span>
<template v-if="fit !== 'page'"><button class="hub-ph" @click="toPage" aria-label="Fit to page" title="Fit to page" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M7 12h10M9.5 9.5L7 12l2.5 2.5M14.5 9.5L17 12l-2.5 2.5"></path></svg></button></template>
<button class="hub-ph" @click="toDocked" aria-label="Side panel" title="Side panel" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M14 4v16"></path><rect x="14" y="4" width="7" height="16" rx="1" fill="currentColor" stroke="none" opacity="0.55"></rect></svg></button>
<div class="hub-seg-more" style="position: relative; flex: none">
<button type="button" aria-label="Views: chat, files, calls" title="More" aria-haspopup="menu" :aria-expanded="moreOpen" @click="moreOpen = !moreOpen" :style="`width: 34px; height: 34px; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: ${moreOpen ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="5" r="1.8"></circle><circle cx="12" cy="12" r="1.8"></circle><circle cx="12" cy="19" r="1.8"></circle></svg></button>
<div v-if="moreOpen" role="menu" aria-label="Show" style="position: absolute; top: 40px; right: 0; z-index: 31; width: 190px; padding: 6px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 12px 30px rgba(0,0,0,0.35); display: flex; flex-direction: column; gap: 2px">
<button v-for="v in views" :key="v.k" type="button" role="menuitemradio" :aria-checked="v.on" @click="v.go(); moreOpen = false" :style="`display: flex; align-items: center; justify-content: space-between; min-height: 40px; padding: 0 12px; border: 0; border-radius: 8px; cursor: pointer; text-align: left; font-size: 13px; background: ${v.on ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); font-weight: ${v.on ? 600 : 500}`">{{ v.label }}<span v-if="v.on" aria-hidden="true" style="color: #B8F36B">●</span></button>
</div>
</div>
<button @click="toMin" aria-label="Minimize" title="Minimize" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path></svg></button>
</div>
<button v-if="fit !== 'page'" class="hub-chat-close" aria-label="Close chat" title="Close chat" @click="toClosed" style="position: absolute; top: 8px; right: 12px; z-index: 6; display: flex; align-items: center; justify-content: center; width: 38px; height: 38px; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="max-width: 1200px; width: 100%; margin: 0 auto; box-sizing: border-box"><DmPrivacy /></div>

<div v-if="isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)"><CallHistory /></div></div>
<ChatSharedFiles v-if="isPanelFiles" :folders="fileFolders" :selected-folder="selectedChatFolder" :color="cur.color" :open-folder="openChatFolder" :close-folder="closeChatFolder"/>
<div v-if="isPanelChat" style="flex: 1; min-height: 0; overflow-y: auto">
<div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px); display: flex; flex-direction: column; gap: 5px">
<span style="align-self: center; font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); padding: 4px 10px; border-radius: 10px; background: var(--k-1A2520,#1A2520)">Today</span>
<template v-if="olderCount > 0 || moreOlder"><button type="button" @click="showOlder" style="align-self: center; height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid var(--k-6B7A72,#6B7A72); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">{{ olderCount > 0 ? 'Show ' + olderCount + ' earlier ' + (olderCount === 1 ? 'message' : 'messages') : 'Load earlier messages' }}</button></template>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :style="u.rowStyle + ';touch-action:pan-y'" @pointerdown="u.startHold" @pointermove="u.moveHold" @pointerup="u.stopHold" @pointerleave="u.stopHold" @pointercancel="u.stopHold" @contextmenu="u.openActions">
<UserAvatar :name="u.who" :initials="u.initials" :color="u.color" :size="32"/>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ u.time }}</span><button type="button" @click.stop="u.openActions($event)" aria-label="Message options" style="margin-left: auto; width: 26px; height: 24px; padding: 0; border: 0; border-radius: 6px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer">⋯</button></div>
<template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: var(--k-C9D6CF,#C9D6CF)">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template>
<template v-if="u.hasText && !u.hasFile"><p v-html="u.htmlText" style="margin: 4px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="380" :row="170"/></template><template v-if="u.hasText && u.hasFile"><p v-html="u.htmlText" style="margin: 6px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasSecureShare"><SecureInboxCard :share="u.secureShare"/></template>
<div v-if="u.reply" style="margin-top: 6px"><button type="button" @click.stop="u.reply" aria-label="Reply to this message" style="height: 26px; padding: 0 10px; border: 0; border-radius: 13px; background: var(--k-22302A,#22302A); color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; font-size: 12px">↩ Reply<template v-if="u.replies"> · {{ u.replies }} earlier</template></button></div>

<div v-if="u.menuOpen" role="group" aria-label="Message actions" style="display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; padding-top: 8px; border-top: 1px solid var(--k-405047,#405047)">
<button type="button" @click.stop="u.deleteForMe" style="height: 30px; padding: 0 10px; border: 1px solid var(--k-526259,#526259); border-radius: 8px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">Delete for me</button>
<button v-if="u.deleteForEveryone" type="button" @click.stop="u.deleteForEveryone" style="height: 30px; padding: 0 10px; border: 1px solid var(--k-6A3A34,#6A3A34); border-radius: 8px; background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer; font-size: 12px">Delete for everyone</button>
<button type="button" @click.stop="u.closeActions" aria-label="Close message actions" style="height: 30px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 12px">Cancel</button>
</div>
</div>
</div>
</template>
</div>
</div>
<div v-if="isPanelChat" style="flex: none; border-top: 1px solid var(--k-24322C,#24322C)">
<div v-if="cur.typingText" role="status" aria-live="polite" style="max-width: 1200px; margin: 0 auto; padding: 6px clamp(16px, 3vw, 32px) 0; font-size: 12px; color: var(--k-7CC4FF,#7CC4FF); font-style: italic">{{ cur.typingText }}</div>
<div v-if="replyBar.active" style="max-width: 1200px; margin: 0 auto; padding: 8px clamp(16px, 3vw, 32px) 0"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 6px; background: var(--k-16201C,#16201C); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: var(--k-E8EFEA,#E8EFEA)">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button></div></div>
<ChatAttachBar/>
<div class="hub-chat-composer" style="max-width: 1200px; margin: 0 auto; position: relative; padding: 12px clamp(16px, 3vw, 32px); display: grid; grid-template-columns: max-content minmax(0, 1fr) 42px; grid-template-rows: auto auto; gap: 6px; align-items: stretch">
<div style="grid-column: 1; grid-row: 1; display: flex; align-items: center; gap: 4px">
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="38" @pick="pickEmojiForum"/>
 </div>
<RichTextInput toolbar-in-actions style="min-width: 0" :model-value="fdraft" @update:model-value="onFDraftValue" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Message ${cur.name}…`" aria-label="Message in private chat" :rows="1" input-style="font-size: 14px; min-height: 38px"/>
<button aria-label="Send" :disabled="composerOff" @click="sendForum" style="grid-column: 3; grid-row: 1 / 3; align-self: center; width: 42px; height: 42px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>

</div>
</div>
</template>
</div>
</section>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';
import ChatRail from './ChatRail.vue';
import ChatSharedFiles from './ChatSharedFiles.vue';
import SecureInboxCard from './SecureInboxCard.vue';

export default {
  name: 'PrivateChatFull',
  components: { CallScreen, ChatRail, ChatSharedFiles, SecureInboxCard },
  props: { fit: { type: String, default: 'screen' } },
  data() { return { moreOpen: false, railOpen: false }; },
  computed: {
    ...hubBindings(["canVoiceCall","canVideoCall","olderCount","moreOlder","showOlder","pickEmojiForum","attachForum","closeChatFolder","composerNote","composerOff","cur","dmPriv","fKey","fdraft","fileFolders","isFullDm","isPanelCalls","isPanelChat","isPanelFiles","goPanelChat","goPanelCalls","goPanelFiles","openChatFolder","replyBar","selectedChatFolder","isPageDm","onFDraftValue","sendForum","showPanelCall","toClosed","toDocked","toFull","toMin","toPage","videoCall","voiceCall"]).computed,
    views() { return [['chat', 'Chat', this.isPanelChat, this.goPanelChat], ['files', 'Files', this.isPanelFiles, this.goPanelFiles], ['calls', 'Calls', this.isPanelCalls, this.goPanelCalls]].map(([k, label, on, go]) => ({ k, label, on: !!on, go })); },
    shown() { return this.fit === 'page' ? !!this.isPageDm : !!this.isFullDm; },
    rootStyle() {
      const base = 'background: var(--k-0F1714,#0F1714); display: flex; flex-direction: row; overflow: hidden; box-sizing: border-box;';
      return this.fit === 'page'
        ? base + ' position: absolute; inset: 0; z-index: 30; border-radius: 0; border: 0; box-shadow: none'
        : base + ' position: fixed; inset: 0; z-index: 40';
    }
  }
};
</script>
