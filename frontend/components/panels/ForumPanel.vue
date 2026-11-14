<!-- Docked side panel for a forum or private chat (full screen on mobile) -->
<template><template v-if="isDocked">
<aside aria-label="Forum panel" class="hub-dock" :style="panelStyle">
<CallScreen variant="panel" />
<button class="hub-grip" aria-label="Drag to resize panel. Arrow keys also resize." @pointerdown="startDrag" @keydown="resizeKey" style="position: absolute; left: 0; top: 0; bottom: 0; width: 14px; padding: 0; border: 0; background: transparent; cursor: ew-resize; z-index: 40; touch-action: none; display: flex; align-items: center; justify-content: center"><span :style="gripStyle"></span></button>
<template v-if="!showPanelCall">
<div class="hub-panel-header" style="position: relative; display: flex; align-items: center; gap: 10px; padding: 14px 58px 12px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)">
<UserAvatar :name="cur.name" :initials="cur.initials" :color="cur.color" :size="34" :font-size="13" radius="10px"/>
<div class="hub-panel-name" style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ cur.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span :style="`color: ${cur.dotColor}`">●</span> {{ cur.subline }}</span></div>
<div class="hub-panel-actions" style="display: flex; align-items: center; gap: 2px; flex: none">
<template v-if="isDM"><button aria-label="Privacy and blocking" :aria-expanded="dmPriv.open" @click="dmPriv.toggleOpen" :style="`width: 36px; height: 36px; border-radius: 9px; border: 0; background: ${dmPriv.open ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: ${dmPriv.anyOn ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-C9D6CF,#C9D6CF)'}; cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M12 3l7 3v5c0 4.5-3 8-7 10-4-2-7-5.5-7-10V6z"></path></svg></button></template>
<div v-if="canAddToForum" style="flex: none">
<button aria-label="Add people to this forum" title="Add people" aria-haspopup="dialog" :aria-expanded="addOpen" @click="toggleAdd" :style="`width: 36px; height: 36px; border-radius: 9px; border: 0; cursor: pointer; display: flex; align-items: center; justify-content: center; background: ${addOpen ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-C9D6CF,#C9D6CF)`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="8" r="3.5"></circle><path d="M2.5 20c0-3.6 2.9-6 6.5-6s6.5 2.4 6.5 6M19 8v6M16 11h6"></path></svg></button>
<template v-if="addOpen"><div role="dialog" aria-label="Add people to this forum" style="position: absolute; top: calc(100% + 6px); right: 12px; z-index: 35; width: min(290px, calc(100% - 24px)); box-sizing: border-box; padding: 12px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 14px 30px rgba(0,0,0,0.45); display: flex; flex-direction: column; gap: 8px">
<strong style="font-size: 13px">Add people to {{ cur.name }}</strong>
<PeoplePicker instant :people="pickerPeople" :exclude="forumMemberIds" placeholder="Search people already in Hub…" @pick="addPersonToForum"/>
<span v-if="addBusy" role="status" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Adding…</span>
<span v-if="addErr" role="alert" style="font-size: 12px; color: #FF9B8A">{{ addErr }}</span>
</div></template>
</div>
<button v-if="canVoiceCall" aria-label="Voice call" @click="voiceCall" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"></path></svg></button>
<button v-if="canVideoCall" aria-label="Video call" @click="videoCall" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<div style="position: relative; flex: none">
<button class="hub-ph" aria-label="Fit to page" title="Fit to page" @click="toPage" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M7 12h10M9.5 9.5L7 12l2.5 2.5M14.5 9.5L17 12l-2.5 2.5"></path></svg></button>
</div>
<button aria-label="Minimize forum" @click="toMin" style="width: 36px; height: 36px; border-radius: 9px; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button>
<div style="position: relative; flex: none">
<button aria-label="More: chat, members, polls, calls" title="More" aria-haspopup="menu" :aria-expanded="moreOpen" @click="moreOpen = !moreOpen" :style="`width: 36px; height: 36px; border-radius: 9px; border: 0; background: ${moreOpen ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="5" r="1.8"></circle><circle cx="12" cy="12" r="1.8"></circle><circle cx="12" cy="19" r="1.8"></circle></svg></button>
<template v-if="moreOpen"><div role="menu" aria-label="Show" style="position: absolute; top: 42px; right: 0; z-index: 31; width: 190px; padding: 6px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 12px 30px rgba(0,0,0,0.35); display: flex; flex-direction: column; gap: 2px">
<button v-for="it in moreItems" :key="it.label" role="menuitemradio" :aria-checked="it.on" @click="it.go(); moreOpen = false" :style="`display: flex; align-items: center; justify-content: space-between; min-height: 40px; padding: 0 12px; border: 0; border-radius: 8px; cursor: pointer; text-align: left; font-size: 13px; background: ${it.on ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); font-weight: ${it.on ? 600 : 500}`">{{ it.label }}<span v-if="it.on" aria-hidden="true" style="color: #B8F36B">●</span></button>
</div></template>
</div>
</div>
<button class="hub-panel-close" :aria-label="isDM ? 'Close chat panel' : 'Close forum panel'" :title="isDM ? 'Close chat' : 'Close forum'" @click="toClosed" style="position: absolute; top: 10px; right: 10px; z-index: 41; width: 38px; height: 38px; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<LiveCallBanner />
<div v-if="panelChips.length > 1" data-no-busy class="hub-panel-chips" style="flex: none; display: flex; gap: 8px; padding: 10px 14px; border-bottom: 1px solid var(--k-24322C,#24322C); overflow-x: auto">
<template v-for="(f, f__i) in panelChips" :key="f__i">
<span v-if="f.sep" aria-hidden="true" style="flex: none; width: 1px; margin: 4px 4px; background: var(--k-34453D,#34453D)"></span>
<span style="position: relative; flex: none; width: 36px; height: 36px">
<button :aria-label="`${f.name}${f.hasUnread ? `, ${f.unread} unread messages` : ''}`" :title="f.name" @click="f.open" :class="{ 'hub-live-glow': !!(liveConv || {})[f.conv || f.sid] }" :style="f.chipStyle"><UserAvatar v-if="f.avatar" :name="f.name" :initials="f.initials" :color="f.color" :size="32" :font-size="12"/><template v-else>{{ f.initials }}</template></button>
<span v-if="f.hasUnread" :style="f.unreadStyle">{{ f.unread > 99 ? '99+' : f.unread }}</span>
</span>
</template>
</div>
<DmPrivacy />
<template v-if="true">

</template>
<div v-if="!isDM && isPanelMembers" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px">
<div style="display: flex; align-items: center; gap: 10px; margin-bottom: 12px; flex-wrap: wrap"><strong style="font-size: 15px; flex: 1">{{ cur.members }} {{ cur.members === 1 ? 'member' : 'members' }}</strong>
<button v-if="canAddToForum" type="button" @click="toggleAdd" :aria-expanded="addOpen" aria-label="Add a person to this forum" style="display: inline-flex; align-items: center; gap: 6px; height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 13px; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg>Add</button></div>
<div v-if="canAddToForum && addOpen" style="margin-bottom: 14px; padding: 12px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-B8F36B,#B8F36B); display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Search someone already in Hub. Picking them adds them to {{ cur.name }} straight away.</span>
<PeoplePicker instant :people="pickerPeople" :exclude="forumMemberIds" placeholder="Search people by name…" @pick="addPersonToForum"/>
<span v-if="addBusy" role="status" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Adding…</span>
<span v-if="addErr" role="alert" style="font-size: 12px; color: #FF9B8A">{{ addErr }}</span></div>
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(m, m__i) in chatCtl.members" :key="m__i"><div style="padding: 10px 12px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; gap: 10px; align-items: center"><UserAvatar :name="m.name" :initials="m.initials" :color="m.color" :size="34"/><div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ m.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ m.forumRole }} · <span :style="`color: ${m.stateColor}`">{{ m.stateLabel }}</span></span></div></div></template>
</div></div>
<div v-if="!isDM && isPanelPolls" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px"><ForumPolls /></div>
<div v-if="isPanelCalls" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px"><CallHistory /></div>
<template v-if="isPanelChat || (isDM && !isPanelCalls)">
<div :ref="chatRef" @click="onActivityClick" style="flex: 1; min-height: 0; overflow-y: auto; padding: 14px 14px 14px 20px; display: flex; flex-direction: column; gap: 5px">
<span style="align-self: center; font-size: 11px; color: var(--k-9DB0A6,#9DB0A6); padding: 4px 10px; border-radius: 10px; background: var(--k-1A2520,#1A2520)">Today</span>
<PageSkeleton v-if="chatLoading" kind="chat" bare />
<template v-if="olderCount > 0 || moreOlder"><button type="button" @click="showOlder" style="align-self: center; height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid var(--k-6B7A72,#6B7A72); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">{{ olderCount > 0 ? 'Show ' + olderCount + ' earlier ' + (olderCount === 1 ? 'message' : 'messages') : 'Load earlier messages' }}</button></template>
<template v-for="(u, u__i) in cur.chat" :key="u__i">
<div :data-seen-conv="u.seenConv || null" :data-seen-msg="u.seenMsg || null" :style="u.rowStyle + ';touch-action:pan-y'" @pointerdown="u.startHold" @pointermove="u.moveHold" @pointerup="u.stopHold" @pointerleave="u.stopHold" @pointercancel="u.stopHold" @contextmenu="u.openActions">
<UserAvatar :name="u.who" :initials="u.initials" :color="u.color" :size="32"/>
<div :style="u.bubbleStyle">
<div style="display: flex; gap: 8px; align-items: baseline"><strong style="font-size: 13px">{{ u.who }}</strong><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ u.time }}</span><button type="button" @click.stop="u.openActions($event)" aria-label="Message options" style="margin-left: auto; width: 26px; height: 24px; padding: 0; border: 0; border-radius: 6px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer">⋯</button></div>
<template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: var(--k-C9D6CF,#C9D6CF)">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template>
<template v-if="u.hasText && !u.hasFile"><p v-html="u.htmlText" style="margin: 4px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="260" :row="110"/></template><template v-if="u.hasText && u.hasFile"><p v-html="u.htmlText" style="margin: 6px 0 0; font-size: 14px; line-height: 1.45; overflow-wrap: anywhere"></p></template>
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
<div style="flex: none; border-top: 1px solid var(--k-24322C,#24322C)"><div v-if="cur.typingText" role="status" aria-live="polite" style="max-width: 900px; margin: 0 auto; padding: 6px clamp(16px, 3vw, 32px) 0; font-size: 12px; color: var(--k-7CC4FF,#7CC4FF); font-style: italic">{{ cur.typingText }}</div><div v-if="replyBar.active" style="padding: 8px 14px 0 18px"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 6px; background: var(--k-16201C,#16201C); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: var(--k-E8EFEA,#E8EFEA)">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button></div></div>
<ChatAttachBar/>
<div class="hub-panel-composer hub-chat-composer" style="position: relative; padding: 12px 14px 12px 18px; display: grid; grid-template-columns: max-content minmax(0, 1fr) 42px; grid-template-rows: auto auto; gap: 6px; align-items: stretch">
<div style="grid-column: 1; grid-row: 1; display: flex; align-items: center; gap: 4px">
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="38" @pick="pickEmojiForum"/>
<template v-if="!isDM"><button aria-label="Create a poll" title="Create a poll" :disabled="!pollView.canCreate" @click="pollView.compose" style="width: 38px; height: 38px; flex: none; border-radius: 50%; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg></button></template>
 </div>
<RichTextInput toolbar-in-actions :mentions="mentionList" style="min-width: 0" :model-value="fdraft" @update:model-value="onFDraftValue" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Message ${cur.name}…`" aria-label="Message in chat" :rows="1" input-style="font-size: 13px; min-height: 38px"/>
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
  data() { return { moreOpen: false }; },
  computed: {
    ...hubBindings(["liveConv","canVoiceCall","canVideoCall","chatLoading","olderCount","moreOlder","showOlder","mentionList","onActivityClick","isPanelMembers","goPanelMembers","chatCtl","canAddToForum","addOpen","addErr","addBusy","toggleAdd","addPersonToForum","pickerPeople","forumMemberIds","pickEmojiForum","attachForum","chatRef","composerNote","composerOff","cur","dmPriv","fKey","fdraft","goPanelChat","goPanelPolls","gripStyle","isDM","isDocked","isPanelChat","replyBar","isPanelCalls","goPanelCalls","isPanelPolls","menuOpen","onFDraftValue","panelChips","panelStyle","pollView","resizeKey","sendForum","showPanelCall","startDrag","toClosed","toFull","toMin","toPage","toggleMenu","videoCall","voiceCall"]).computed,
    // the chat's views sit behind the ⋯ button so the side panel keeps its height for the conversation
    moreItems() {
      const v = this; const dm = v.isDM;
      return [{ label: 'Chat', on: v.isPanelChat || (dm && !v.isPanelCalls), go: v.goPanelChat }]
        .concat(dm ? [] : [{ label: 'Members', on: v.isPanelMembers, go: v.goPanelMembers }, { label: 'Polls (' + ((v.pollView && v.pollView.count) || 0) + ')', on: v.isPanelPolls, go: v.goPanelPolls }])
        .concat([{ label: 'Calls', on: v.isPanelCalls, go: v.goPanelCalls }]);
    }
  },
};
</script>
