<!-- Docked side panel for a forum or private chat (full screen on mobile) -->
<template><template v-if="isDocked">
<aside aria-label="Forum panel" class="hub-dock" :style="panelStyle">
<CallScreen variant="panel" />
<button class="hub-grip" aria-label="Drag to resize panel. Arrow keys also resize." @pointerdown="startDrag" @keydown="resizeKey" style="position: absolute; left: 0; top: 0; bottom: 0; width: 14px; padding: 0; border: 0; background: transparent; cursor: ew-resize; z-index: 40; touch-action: none; display: flex; align-items: center; justify-content: center"><span :style="gripStyle"></span></button>
<template v-if="!showPanelCall">
<div class="hub-panel-header" style="position: relative; display: flex; align-items: center; gap: 10px; padding: 14px 58px 12px 20px; border-bottom: 1px solid #24322C">
<UserAvatar :name="cur.name" :initials="cur.initials" :color="cur.color" :size="34" :font-size="13" radius="10px"/>
<div class="hub-panel-name" style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ cur.name }}</strong><span style="font-size: 12px; color: #9DB0A6"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }}</span></div>
<div class="hub-panel-actions" style="display: flex; align-items: center; gap: 2px; flex: none">
<template v-if="isDM"><button aria-label="Privacy and blocking" :aria-expanded="dmPriv.open" @click="dmPriv.toggleOpen" :style="`width: 36px; height: 36px; border-radius: 9px; border: 0; background: ${dmPriv.open ? '#2C3D34' : 'transparent'}; color: ${dmPriv.anyOn ? '#FFB86B' : '#C9D6CF'}; cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z"></path></svg></button></template>
<button aria-label="Voice call" @click="voiceCall" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"></path></svg></button>
<button aria-label="Video call" @click="videoCall" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<div style="position: relative; flex: none">
<button aria-label="Expand chat" aria-haspopup="menu" :aria-expanded="menuOpen" @click="toggleMenu" :style="`width: 36px; height: 36px; border-radius: 9px; border: 0; background: ${menuOpen ? '#2C3D34' : 'transparent'}; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button>
<template v-if="menuOpen"><div role="menu" aria-label="Expand chat" style="position: absolute; top: 42px; right: 0; z-index: 30; width: 230px; padding: 6px; border-radius: 12px; background: #1D2A25; border: 1px solid #34453D; box-shadow: 0 14px 30px rgba(0,0,0,0.45); display: flex; flex-direction: column; gap: 2px">
<button role="menuitem" @click="toPage" style="display: flex; align-items: center; gap: 10px; min-height: 52px; padding: 6px 10px; border: 0; border-radius: 8px; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg><span style="display: flex; flex-direction: column; gap: 1px"><span style="font-size: 14px; font-weight: 600">Fit to page</span><span style="font-size: 12px; color: #9DB0A6">Fills the space beside the sidebar</span></span></button>
<button role="menuitem" @click="toFull" style="display: flex; align-items: center; gap: 10px; min-height: 52px; padding: 6px 10px; border: 0; border-radius: 8px; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg><span style="display: flex; flex-direction: column; gap: 1px"><span style="font-size: 14px; font-weight: 600">Full screen</span><span style="font-size: 12px; color: #9DB0A6">Covers the whole window</span></span></button>
</div></template>
</div>
<button aria-label="Minimize forum" @click="toMin" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button>
</div>
<button :aria-label="isDM ? 'Close chat panel' : 'Close forum panel'" :title="isDM ? 'Close chat' : 'Close forum'" @click="toClosed" style="position: absolute; top: 10px; right: 10px; z-index: 41; width: 38px; height: 38px; padding: 0; border-radius: 50%; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="display: flex; gap: 8px; padding: 10px 14px; border-bottom: 1px solid #24322C; overflow-x: auto">
<template v-for="(f, f__i) in panelChips" :key="f__i">
<span style="position: relative; flex: none; width: 36px; height: 36px">
<button :aria-label="`${f.name}${f.hasUnread ? `, ${f.unread} unread messages` : ''}`" :title="f.name" @click="f.open" :style="f.chipStyle"><UserAvatar v-if="isDM" :name="f.name" :initials="f.initials" :color="f.color" :size="32" :font-size="12"/><template v-else>{{ f.initials }}</template></button>
<span v-if="f.hasUnread" :style="f.unreadStyle">{{ f.unread > 99 ? '99+' : f.unread }}</span>
</span>
</template>
</div>
<DmPrivacy />
<template v-if="true">
<div role="tablist" style="flex: none; display: flex; gap: 4px; padding: 6px 14px 0 20px; border-bottom: 1px solid #24322C"><button role="tab" :aria-selected="isPanelChat" @click="goPanelChat" :style="`min-height: 38px; padding: 0 14px; border: 0; background: transparent; cursor: pointer; font-size: 13px; border-bottom: 2px solid ${isPanelChat ? '#B8F36B' : 'transparent'}; color: ${isPanelChat ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelChat ? 600 : 500}`">Chat</button><button v-if="!isDM" role="tab" :aria-selected="isPanelPolls" @click="goPanelPolls" :style="`min-height: 38px; padding: 0 14px; border: 0; background: transparent; cursor: pointer; font-size: 13px; border-bottom: 2px solid ${isPanelPolls ? '#B8F36B' : 'transparent'}; color: ${isPanelPolls ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelPolls ? 600 : 500}`">Polls ({{ pollView.count }})</button><button role="tab" :aria-selected="isPanelCalls" @click="goPanelCalls" :style="`min-height: 38px; padding: 0 14px; border: 0; background: transparent; cursor: pointer; font-size: 13px; border-bottom: 2px solid ${isPanelCalls ? '#B8F36B' : 'transparent'}; color: ${isPanelCalls ? '#E8EFEA' : '#9DB0A6'}; font-weight: ${isPanelCalls ? 600 : 500}`">Calls</button></div>
</template>
<div v-if="!isDM && isPanelPolls" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px"><ForumPolls /></div>
<div v-if="isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px"><CallHistory /></div>
<template v-if="isPanelChat || (isDM && !isPanelCalls)">
<div :ref="chatRef" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px; display: flex; flex-direction: column; gap: 12px">
<span style="align-self: center; font-size: 11px; color: #9DB0A6; padding: 4px 10px; border-radius: 10px; background: #1A2520">Today</span>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :style="u.rowStyle + ';touch-action:pan-y'" @pointerdown="u.startHold" @pointerup="u.stopHold" @pointerleave="u.stopHold" @pointercancel="u.stopHold" @contextmenu="u.openActions">
<UserAvatar :name="u.who" :initials="u.initials" :color="u.color" :size="32"/>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ u.time }}</span><button type="button" @click.stop="u.openActions($event)" aria-label="Message options" style="margin-left: auto; width: 26px; height: 24px; padding: 0; border: 0; border-radius: 6px; background: transparent; color: #9DB0A6; cursor: pointer">⋯</button><template v-if="u.reply"><button type="button" @click="u.reply" aria-label="Reply to this message" style="height: 22px; padding: 0 8px; border: 0; border-radius: 11px; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 11px">↩ Reply</button></template></div>
<template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid #B8F36B; border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: #9DB0A6; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: #C9D6CF">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template>
<template v-if="u.hasText"><p v-html="u.htmlText" style="margin: 4px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="260" :row="110"/></template>
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
<div style="flex: none; border-top: 1px solid #24322C"><div v-if="replyBar.active" style="padding: 8px 14px 0 18px"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid #B8F36B; border-radius: 6px; background: #16201C; font-size: 12px; color: #9DB0A6; min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: #E8EFEA">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div></div>
<div style="position: relative; padding: 12px 14px 12px 18px; display: grid; grid-template-columns: max-content minmax(0, 1fr) 42px; grid-template-rows: auto auto; gap: 6px; align-items: stretch">
<div style="grid-column: 1; grid-row: 1; display: flex; align-items: center; gap: 4px">
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: #C9D6CF; cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="38" @pick="pickEmojiForum"/>
<template v-if="!isDM"><button aria-label="Create a poll" title="Create a poll" :disabled="!pollView.canCreate" @click="pollView.compose" style="width: 38px; height: 38px; flex: none; border-radius: 50%; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg></button></template>
 </div>
<RichTextInput toolbar-in-actions style="min-width: 0" :model-value="fdraft" @update:model-value="onFDraftValue" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Message ${cur.name}…`" aria-label="Message in chat" :rows="1" input-style="font-size: 13px; min-height: 38px"/>
<button aria-label="Send" :disabled="composerOff" @click="sendForum" style="grid-column: 3; grid-row: 1 / 3; align-self: center; width: 42px; height: 42px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>
</div>
</div>
</template>
</template>
</aside>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';
import SecureInboxCard from './SecureInboxCard.vue';

export default {
  name: 'ForumPanel',
  components: { CallScreen, SecureInboxCard },
  ...hubBindings(["pickEmojiForum","attachForum","chatRef","composerNote","composerOff","cur","dmPriv","fKey","fdraft","goPanelChat","goPanelPolls","gripStyle","isDM","isDocked","isPanelChat","replyBar","isPanelCalls","goPanelCalls","isPanelPolls","menuOpen","onFDraftValue","panelChips","panelStyle","pollView","resizeKey","sendForum","showPanelCall","startDrag","toClosed","toFull","toMin","toPage","toggleMenu","videoCall","voiceCall"])
};
</script>
