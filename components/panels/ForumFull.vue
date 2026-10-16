<!-- Forum, full screen or fitted to the page (beside the sidebar) -->
<template><template v-if="shown">
<section :aria-label="fit === 'page' ? 'Forum fit to page' : 'Forum full screen'" :class="fit === 'page' ? 'hub-fitpage' : 'hub-screen'" :style="rootStyle">
<CallScreen variant="full" :fit="fit" />
<template v-if="!showPanelCall">
<div style="max-width: 1100px; min-height: 100%; box-sizing: border-box; margin: 0 auto; padding: 0 clamp(16px, 3vw, 32px) 8px; display: flex; flex-direction: column; gap: 12px">
<div style="position: sticky; top: 0; z-index: 5; background: #0F1714; padding-top: 8px; display: flex; flex-direction: column; gap: 8px; box-shadow: 0 14px 16px -14px rgba(0,0,0,0.7)">
<div class="ff-bar" style="display: flex; flex-wrap: wrap; gap: 6px 12px; align-items: center; padding: 6px 12px; border-radius: 16px; background: #16201C; border: 1px solid #24322C">
<span :style="`width: 36px; height: 36px; flex: none; border-radius: 10px; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 15px`">{{ cur.initials }}</span>
<div style="flex: 0 1 160px; min-width: 110px; display: flex; flex-direction: column">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 18px; line-height: 1.25; letter-spacing: -0.01em; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ cur.name }}</h1>
<span :title="cur.about" style="font-size: 12px; color: #9DB0A6; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ cur.members }} members · {{ cur.online }} online</span>
</div>
<div role="tablist" class="ff-tabs" style="flex: 1 1 320px; min-width: 0; display: flex; gap: 2px; overflow-x: auto">
<button role="tab" @click="tabTo.discussions" :style="tab.discussions">Discussions</button>
<button role="tab" @click="tabTo.polls" :style="tab.polls">Polls ({{ pollView.count }})</button>
<template v-if="has.sessions"><button role="tab" @click="tabTo.sessions" :style="tab.sessions">Meetings</button></template>
<button role="tab" @click="tabTo.calls" :style="tab.calls">Calls</button>
<button role="tab" @click="tabTo.members" :style="tab.members">Members</button>
<button role="tab" @click="tabTo.files" :style="tab.files">Files</button>
<button role="tab" @click="tabTo.about" :style="tab.about">About</button>
</div>
<div style="display: flex; gap: 6px; flex-wrap: wrap; margin-left: auto">
<button @click="toggleInvite" aria-label="Invite members" title="Invite members" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0; background: #B8F36B; color: #0F1714; border-color: #B8F36B"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="8" r="4"></circle><path d="M2 21a7 7 0 0 1 14 0M19 8v6M16 11h6"></path></svg></button>
<button @click="voiceCall" aria-label="Voice call" title="Voice call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"></path></svg></button>
<button @click="videoCall" aria-label="Video call" title="Video call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<template v-if="fit === 'page'"><button @click="toFull" aria-label="Full screen" title="Full screen" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button></template>
<template v-if="fit !== 'page'"><button @click="toPage" aria-label="Fit to page" title="Fit to page" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M9 4v16"></path></svg></button></template>
<button @click="toDocked" aria-label="Side panel" title="Side panel" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg></button>
<button @click="toMin" aria-label="Minimize" title="Minimize" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path></svg></button>
<button @click="toClosed" aria-label="Close forum" title="Close forum" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</div>

<template v-if="inviteOpen">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: flex-end; padding: 16px; border-radius: 14px; border: 1px solid #B8F36B; background: #16201C">
<label style="flex: 1 1 280px; display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Add people by name or email<input type="text" placeholder="name@company.com, Brian K…" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label>
<button style="height: 42px; padding: 0 18px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Send invites</button>
<button style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Copy invite link</button>
</div>
</template>

</div>

<template v-if="ftab.calls"><CallHistory /></template>

<template v-if="ftab.discussions">
<div style="display: flex; flex-direction: column; gap: 12px">
<template v-for="(u, u__i) in cur.topics" :key="u__i">
<div style="display: flex; gap: 14px; padding: 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<div aria-hidden="true" :style="`width: 40px; height: 40px; flex: none; border-radius: 50%; background: ${u.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700`">{{ u.initials }}</div>
<div style="flex: 1; display: flex; flex-direction: column; gap: 6px"><div style="display: flex; gap: 8px; align-items: baseline"><strong>{{ u.who }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ u.time }}</span></div><template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid #B8F36B; border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: #9DB0A6; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: #C9D6CF">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template><template v-if="u.hasText"><p style="margin: 0; font-size: 15px; line-height: 1.5">{{ u.text }}</p></template><template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="380" :row="170"/></template><button type="button" @click="u.reply" aria-label="Reply to this message" style="align-self: flex-start; padding: 0; border: 0; background: transparent; color: #9DB0A6; cursor: pointer; font-size: 13px">↩ Reply<template v-if="u.replies"> · {{ u.replies }} earlier replies</template></button></div>
</div>
</template>
</div>
</template>

<template v-if="ftab.sessions">
<div style="display: flex; flex-direction: column; gap: 14px">
<div style="padding: 24px; border-radius: 16px; background: #1A2520; border: 1px solid #2A3832; display: flex; flex-wrap: wrap; gap: 16px; align-items: center">
<div style="flex: 1 1 280px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 18px">Start a meeting or standup</strong><span style="color: #9DB0A6">Video, audio or screen share with forum members. Recordings and notes are saved to the forum.</span></div>
<button @click="videoCall" style="height: 44px; padding: 0 22px; border-radius: 22px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Go live now</button>
<button style="height: 44px; padding: 0 22px; border-radius: 22px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Schedule</button>
</div>
<div style="padding: 16px 18px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-wrap: wrap; gap: 12px; align-items: center"><div style="flex: 1 1 240px"><strong>Weekly check-in</strong><div style="color: #9DB0A6; font-size: 13px">Fridays, 10:00 · recurring</div></div><button style="height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">RSVP</button></div>
</div>
</template>

<template v-if="ftab.polls"><ForumPolls /></template>

<template v-if="ftab.members">
<ForumChatControls />
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px">
<template v-for="(m, m__i) in chatCtl.members" :key="m__i">
<div style="padding: 14px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; gap: 12px; align-items: center">
<div aria-hidden="true" :style="`width: 40px; height: 40px; flex: none; border-radius: 50%; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700`">{{ m.initials }}</div>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong>{{ m.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ m.forumRole }} · <span :style="`color: ${m.stateColor}`">{{ m.stateLabel }}</span></span></div>
<template v-if="chatCtl.canManage"><button @click="m.toggle" :aria-label="`${m.actionLabel}: ${m.name}`" :style="`flex: none; height: 30px; padding: 0 12px; border-radius: 15px; cursor: pointer; font-size: 12px; ${m.can ? 'border: 1px solid #6A3A34; background: transparent; color: #FF9B8A' : 'border: 0; background: #B8F36B; color: #0F1714; font-weight: 600'}`">{{ m.actionLabel }}</button></template>
</div>
</template>
</div>
</template>

<template v-if="ftab.files">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 12px">
<template v-for="(d, d__i) in files" :key="d__i">
<div style="padding: 16px; border-radius: 14px; border: 1px solid #24322C; background: #16201C; display: flex; flex-direction: column; gap: 8px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 11px; color: ${d.color}`">{{ d.kind }}</span><strong style="font-weight: 600">{{ d.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ d.meta }}</span></div>
</template>
</div>
</template>

<template v-if="ftab.about">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 12px">
<div style="padding: 18px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: #9DB0A6; text-transform: uppercase; letter-spacing: 0.06em">Description</span><span>{{ cur.about }}</span></div>
<div style="padding: 18px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: #9DB0A6; text-transform: uppercase; letter-spacing: 0.06em">Who can join</span><span>{{ cur.privacy }}</span></div>
<div style="padding: 18px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: #9DB0A6; text-transform: uppercase; letter-spacing: 0.06em">Rules</span><span>Be respectful. Keep topics on-theme. No sharing secrets in chat; use Secure share.</span></div>
</div>
</template>

<template v-if="ftab.discussions">
<div style="margin-top: auto; position: sticky; bottom: 0; z-index: 5; padding: 12px 0 16px; background: #0F1714; box-shadow: 0 -14px 16px -14px rgba(0,0,0,0.7)">
<template v-if="replyBar.active"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid #B8F36B; border-radius: 6px; background: #16201C; font-size: 12px; color: #9DB0A6; min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: #E8EFEA">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div></template>
<div style="position: relative; display: flex; gap: 6px; align-items: center; padding: 6px; border-radius: 14px; border: 1px solid #2A3832; background: #16201C">
<label title="Attach files" style="position: relative; width: 42px; height: 42px; flex: none; border-radius: 10px; display: flex; align-items: center; justify-content: center; color: #C9D6CF; cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="42" @pick="pickEmojiForum"/>
<button aria-label="Create a poll" title="Create a poll" :disabled="!pollView.canCreate" @click="pollView.compose" style="width: 42px; height: 42px; flex: none; border-radius: 10px; border: 0; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg></button>
<label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Start a topic</span><input type="text" :value="fdraft" @input="onFDraft" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Start a new topic in ${cur.name}…`" style="width: 100%; box-sizing: border-box; height: 42px; padding: 0 10px; border: 0; outline: none; background: transparent; color: #E8EFEA; font-size: 15px"></label>
<button :disabled="composerOff" @click="sendForum" style="height: 42px; padding: 0 20px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Post</button>
</div>
</div>
</template>
</div>
</template>
</section>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';

export default {
  name: 'ForumFull',
  components: { CallScreen },
  props: { fit: { type: String, default: 'screen' } },
  computed: {
    ...hubBindings(["pickEmojiForum","attachForum","replyBar","chatCtl","composerNote","composerOff","cur","fKey","fdraft","files","ftab","has","inviteOpen","isFullForum","isPageForum","onFDraft","pollView","sendForum","showPanelCall","tab","tabTo","toClosed","toDocked","toFull","toMin","toPage","toggleInvite","videoCall","voiceCall"]).computed,
    shown() { return this.fit === 'page' ? !!this.isPageForum : !!this.isFullForum; },
    rootStyle() {
      const base = 'background: #0F1714; overflow-y: auto; box-sizing: border-box;';
      return this.fit === 'page'
        ? base + ' position: absolute; inset: 24px 12px 12px 12px; z-index: 30; border-radius: 22px; border: 1px solid #2A3832; box-shadow: 0 12px 36px rgba(0,0,0,0.35)'
        : base + ' position: fixed; inset: 0; z-index: 40';
    }
  }
};
</script>
