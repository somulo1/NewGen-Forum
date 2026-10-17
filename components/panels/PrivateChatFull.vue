<!-- Private chat, full screen or fitted to the page (beside the sidebar) -->
<template><template v-if="shown">
<section :aria-label="fit === 'page' ? 'Private chat fit to page' : 'Private chat full screen'" :class="fit === 'page' ? 'hub-fitpage' : 'hub-screen'" :style="rootStyle">
<CallScreen variant="full" :fit="fit" />
<template v-if="!showPanelCall">
<div style="flex: none; border-bottom: 1px solid #24322C; background: #0F1714">
<div style="max-width: 900px; margin: 0 auto; padding: 8px clamp(16px, 3vw, 32px); display: flex; flex-wrap: wrap; align-items: center; gap: 8px 10px">
<UserAvatar :name="cur.name" :initials="cur.initials" :color="cur.color" :size="34" :font-size="14"/>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-family: 'Space Grotesk', sans-serif; font-size: 17px">{{ cur.name }}</strong><span style="font-size: 13px; color: #9DB0A6"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }} · only you two can see this</span></div>
<button :aria-expanded="dmPriv.open" aria-label="Privacy and blocking" title="Privacy and blocking" @click="dmPriv.toggleOpen" :style="`display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid ${dmPriv.anyOn ? '#FFB86B' : '#2A3832'}; background: ${dmPriv.open ? '#2C3D34' : 'transparent'}; color: ${dmPriv.anyOn ? '#FFB86B' : '#E8EFEA'}; cursor: pointer`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z"></path></svg></button>
<button @click="voiceCall" aria-label="Voice call" title="Voice call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: #B8F36B; color: #0F1714; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1-2-2z"></path></svg></button>
<button @click="videoCall" aria-label="Video call" title="Video call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<template v-if="fit === 'page'"><button @click="toFull" aria-label="Full screen" title="Full screen" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button></template>
<template v-if="fit !== 'page'"><button @click="toPage" aria-label="Fit to page" title="Fit to page" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg></button></template>
<button @click="toDocked" aria-label="Side panel" title="Side panel" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg></button>
<button @click="toMin" aria-label="Minimize" title="Minimize" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path></svg></button>
<button aria-label="Close chat" title="Close chat" @click="toClosed" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</div>
<div style="max-width: 900px; width: 100%; margin: 0 auto; box-sizing: border-box"><DmPrivacy /></div>
<div style="flex: none; border-bottom: 1px solid #24322C"><div role="tablist" style="max-width: 900px; margin: 0 auto; padding: 0 clamp(8px, 2vw, 24px); display: flex; gap: 4px"><button role="tab" :aria-selected="isPanelChat" @click="goPanelChat" :style="`min-height: 40px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${isPanelChat ? '#B8F36B' : 'transparent'}; color: ${isPanelChat ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelChat ? 600 : 500}`">Chat</button><button role="tab" :aria-selected="isPanelFiles" @click="goPanelFiles" :style="`min-height: 40px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${isPanelFiles ? '#B8F36B' : 'transparent'}; color: ${isPanelFiles ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelFiles ? 600 : 500}`">Files</button><button role="tab" :aria-selected="isPanelCalls" @click="goPanelCalls" :style="`min-height: 40px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${isPanelCalls ? '#B8F36B' : 'transparent'}; color: ${isPanelCalls ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelCalls ? 600 : 500}`">Calls</button></div></div>
<div v-if="isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 900px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)"><CallHistory /></div></div>
<ChatSharedFiles v-if="isPanelFiles" :folders="fileFolders" :selected-folder="selectedChatFolder" :color="cur.color" :open-folder="openChatFolder" :close-folder="closeChatFolder"/>
<div v-if="isPanelChat" style="flex: 1; min-height: 0; overflow-y: auto">
<div style="max-width: 900px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px); display: flex; flex-direction: column; gap: 12px">
<span style="align-self: center; font-size: 11px; color: #9DB0A6; padding: 4px 10px; border-radius: 10px; background: #1A2520">Today</span>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :style="u.rowStyle + ';touch-action:pan-y'" @pointerdown="u.startHold" @pointerup="u.stopHold" @pointerleave="u.stopHold" @pointercancel="u.stopHold" @contextmenu="u.openActions">
<UserAvatar :name="u.who" :initials="u.initials" :color="u.color" :size="32"/>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ u.time }}</span><button type="button" @click.stop="u.openActions($event)" aria-label="Message options" style="margin-left: auto; width: 26px; height: 24px; padding: 0; border: 0; border-radius: 6px; background: transparent; color: #9DB0A6; cursor: pointer">⋯</button><template v-if="u.reply"><button type="button" @click="u.reply" aria-label="Reply to this message" style="height: 22px; padding: 0 8px; border: 0; border-radius: 11px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px">↩ Reply</button></template></div>
<template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid #B8F36B; border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: #9DB0A6; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: #C9D6CF">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template>
<template v-if="u.hasText"><p v-html="u.htmlText" style="margin: 4px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="380" :row="170"/></template>
<template v-if="u.hasSecureShare"><SecureInboxCard :share="u.secureShare"/></template>
<div v-if="u.menuOpen" role="group" aria-label="Message actions" style="display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; padding-top: 8px; border-top: 1px solid #405047">
<button type="button" @click.stop="u.deleteForMe" style="height: 30px; padding: 0 10px; border: 1px solid #526259; border-radius: 8px; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">Delete for me</button>
<button v-if="u.deleteForEveryone" type="button" @click.stop="u.deleteForEveryone" style="height: 30px; padding: 0 10px; border: 1px solid #6A3A34; border-radius: 8px; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 12px">Delete for everyone</button>
<button type="button" @click.stop="u.closeActions" aria-label="Close message actions" style="height: 30px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 12px">Cancel</button>
</div>
</div>
</div>
</template>
</div>
</div>
<div v-if="isPanelChat" style="flex: none; border-top: 1px solid #24322C">
<div v-if="replyBar.active" style="max-width: 900px; margin: 0 auto; padding: 8px clamp(16px, 3vw, 32px) 0"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid #B8F36B; border-radius: 6px; background: #16201C; font-size: 12px; color: #9DB0A6; min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: #E8EFEA">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div></div>
<div style="max-width: 900px; margin: 0 auto; position: relative; padding: 12px clamp(16px, 3vw, 32px); display: grid; grid-template-columns: max-content minmax(0, 1fr) 42px; grid-template-rows: auto auto; gap: 6px; align-items: stretch">
<div style="grid-column: 1; grid-row: 1; display: flex; align-items: center; gap: 4px">
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #C9D6CF; cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="38" @pick="pickEmojiForum"/>
 </div>
<RichTextInput toolbar-in-actions style="min-width: 0" :model-value="fdraft" @update:model-value="onFDraftValue" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Message ${cur.name}…`" aria-label="Message in private chat" :rows="1" input-style="font-size: 14px; min-height: 38px"/>
<button aria-label="Send" :disabled="composerOff" @click="sendForum" style="grid-column: 3; grid-row: 1 / 3; align-self: center; width: 42px; height: 42px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>

</div>
</div>
</template>
</section>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';
import ChatSharedFiles from './ChatSharedFiles.vue';
import SecureInboxCard from './SecureInboxCard.vue';

export default {
  name: 'PrivateChatFull',
  components: { CallScreen, ChatSharedFiles, SecureInboxCard },
  props: { fit: { type: String, default: 'screen' } },
  computed: {
    ...hubBindings(["pickEmojiForum","attachForum","closeChatFolder","composerNote","composerOff","cur","dmPriv","fKey","fdraft","fileFolders","isFullDm","isPanelCalls","isPanelChat","isPanelFiles","goPanelChat","goPanelCalls","goPanelFiles","openChatFolder","replyBar","selectedChatFolder","isPageDm","onFDraftValue","sendForum","showPanelCall","toClosed","toDocked","toFull","toMin","toPage","videoCall","voiceCall"]).computed,
    shown() { return this.fit === 'page' ? !!this.isPageDm : !!this.isFullDm; },
    rootStyle() {
      const base = 'background: #0F1714; display: flex; flex-direction: column; overflow: hidden; box-sizing: border-box;';
      return this.fit === 'page'
        ? base + ' position: absolute; inset: 24px 12px 12px 12px; z-index: 30; border-radius: 22px; border: 1px solid #2A3832; box-shadow: 0 12px 36px rgba(0,0,0,0.35)'
        : base + ' position: fixed; inset: 0; z-index: 40';
    }
  }
};
</script>
