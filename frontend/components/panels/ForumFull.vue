<!-- Forum, full screen or fitted to the page (beside the sidebar) -->
<template><template v-if="shown">
<section :aria-label="fit === 'page' ? 'Forum fit to page' : 'Forum full screen'" :class="fit === 'page' ? 'hub-fitpage' : 'hub-screen'" :style="rootStyle">
<ChatRail :open="railOpen" @close="railOpen = false" kind="forum" />
<div style="position: relative; flex: 1; min-width: 0; min-height: 0; display: flex; flex-direction: column; overflow: hidden">
<CallScreen variant="full" :fit="fit" />
<template v-if="!showPanelCall">
<div style="width: 100%; flex: 1; min-height: 0; box-sizing: border-box; display: flex; flex-direction: column">
<div style="flex: none; position: relative; z-index: 5; background: var(--k-0F1714,#0F1714); border-bottom: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column">
<div class="ff-bar hub-chat-bar" style="width: 100%; box-sizing: border-box; margin: 0; display: flex; flex-wrap: wrap; gap: 8px 10px; align-items: center; padding: 8px clamp(16px, 3vw, 32px); padding-right: 58px; border: 0; background: var(--k-0F1714,#0F1714)">
<button type="button" class="hub-chatlist-btn" :aria-label="'All forums'" :title="'All forums'" @click="railOpen = true" style="width: 34px; height: 34px; flex: none; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M4 6h16M4 12h16M4 18h16"></path></svg></button>
<span :style="`width: 34px; height: 34px; flex: none; border-radius: 10px; background: ${cur.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 14px`">{{ cur.initials }}</span>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 17px; line-height: 1.25; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ cur.name }}</h1>
<span :title="cur.about" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6); overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ cur.members }} members · {{ cur.online }} online</span>
</div>
<div role="tablist" aria-label="Forum views" class="hub-seg" style="display: inline-flex; flex-wrap: wrap; flex: none; padding: 3px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-2A3832,#2A3832)"><button v-for="v in views" :key="v.k" role="tab" :aria-selected="v.on" @click="v.go" :style="`height: 30px; padding: 0 12px; border: 0; border-radius: 9px; cursor: pointer; font-size: 13px; font-weight: 600; white-space: nowrap; background: ${v.on ? '#B8F36B' : 'transparent'}; color: ${v.on ? '#0F1714' : 'var(--k-C9D6CF,#C9D6CF)'}`">{{ v.label }}</button></div>
<button @click="toggleInvite" aria-label="Invite members" title="Invite members" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0; background: #B8F36B; color: #0F1714; border-color: var(--k-B8F36B,#B8F36B)"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="8" r="4"></circle><path d="M2 21a7 7 0 0 1 14 0M19 8v6M16 11h6"></path></svg></button>
<button v-if="canVoiceCall" @click="voiceCall" aria-label="Voice call" title="Voice call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M5 4h4l2 5-2.5 1.5a11 11 0 0 0 5 5L15 13l5 2v4a2 2 0 0 1-2 2A16 16 0 0 1 3 6a2 2 0 0 1 2-2z"></path></svg></button>
<button v-if="canVideoCall" @click="videoCall" aria-label="Video call" title="Video call" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg></button>
<span class="hub-sep" aria-hidden="true" style="width: 1px; height: 24px; flex: none; align-self: center; background: var(--k-2A3832,#2A3832)"></span>
<template v-if="fit !== 'page'"><button class="hub-ph" @click="toPage" aria-label="Fit to page" title="Fit to page" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M7 12h10M9.5 9.5L7 12l2.5 2.5M14.5 9.5L17 12l-2.5 2.5"></path></svg></button></template>
<button class="hub-ph" @click="toDocked" aria-label="Side panel" title="Side panel" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M14 4v16"></path><rect x="14" y="4" width="7" height="16" rx="1" fill="currentColor" stroke="none" opacity="0.55"></rect></svg></button>
<div class="hub-seg-more" style="position: relative; flex: none">
<button type="button" aria-label="Views: discussions, polls, calls, members, files, about" title="More" aria-haspopup="menu" :aria-expanded="moreOpen" @click="moreOpen = !moreOpen" :style="`width: 34px; height: 34px; padding: 0; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: ${moreOpen ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center`"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="currentColor"><circle cx="12" cy="5" r="1.8"></circle><circle cx="12" cy="12" r="1.8"></circle><circle cx="12" cy="19" r="1.8"></circle></svg></button>
<div v-if="moreOpen" role="menu" aria-label="Show" style="position: absolute; top: 40px; right: 0; z-index: 31; width: 190px; padding: 6px; border-radius: 12px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D); box-shadow: 0 12px 30px rgba(0,0,0,0.35); display: flex; flex-direction: column; gap: 2px">
<button v-for="v in views" :key="v.k" type="button" role="menuitemradio" :aria-checked="v.on" @click="v.go(); moreOpen = false" :style="`display: flex; align-items: center; justify-content: space-between; min-height: 40px; padding: 0 12px; border: 0; border-radius: 8px; cursor: pointer; text-align: left; font-size: 13px; background: ${v.on ? 'var(--k-2C3D34,#2C3D34)' : 'transparent'}; color: var(--k-E8EFEA,#E8EFEA); font-weight: ${v.on ? 600 : 500}`">{{ v.label }}<span v-if="v.on" aria-hidden="true" style="color: #B8F36B">●</span></button>
</div>
</div>
<button @click="toMin" aria-label="Minimize" title="Minimize" style="display: flex; align-items: center; justify-content: center; width: 34px; height: 34px; flex: none; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"></path></svg></button>
</div>
<button v-if="fit !== 'page'" class="hub-chat-close" @click="toClosed" aria-label="Close forum" title="Close forum" style="position: absolute; top: 8px; right: 12px; z-index: 6; display: flex; align-items: center; justify-content: center; width: 38px; height: 38px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>

<template v-if="inviteOpen">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: flex-end; padding: 16px; border-radius: 14px; border: 1px solid var(--k-B8F36B,#B8F36B); background: var(--k-16201C,#16201C)">
<div v-if="canAddToForum" style="flex: 1 1 100%; display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Add someone already in Hub<PeoplePicker instant :people="pickerPeople" :exclude="forumMemberIds" placeholder="Search people by name…" @pick="addPersonToForum"/><span v-if="addErr" role="alert" style="font-size: 12px; color: #FF9B8A">{{ addErr }}</span></div>
<label style="flex: 1 1 280px; display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Add people by name or email<input type="text" :value="inviteText" @input="onInviteText" placeholder="name@company.com, Brian Kiprop…" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)"></label>
<button :disabled="inviteBusy" @click="sendInvites" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ inviteBusy ? 'Adding…' : 'Send invites' }}</button>
<span v-if="inviteErr" role="alert" style="flex: 1 1 100%; font-size: 13px; color: #FF9B8A">{{ inviteErr }}</span>
<button style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Copy invite link</button>
</div>
</template>

<template v-if="ftab.calls"><div style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)"><CallHistory /></div></div></template>

<template v-if="ftab.discussions">
<div style="flex: 1; min-height: 0; display: flex; flex-direction: column; gap: 12px">
<div :ref="chatRef" @click="onActivityClick" style="width: 100%; flex: 1; min-height: 0; box-sizing: border-box; overflow-y: auto">
<div style="max-width: 1200px; min-height: 100%; box-sizing: border-box; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px); display: flex; flex-direction: column; gap: 5px">
<PageSkeleton v-if="chatLoading" kind="chat" bare />
<template v-if="olderCount > 0 || moreOlder"><button type="button" @click="showOlder" style="align-self: center; height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid var(--k-6B7A72,#6B7A72); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">{{ olderCount > 0 ? 'Show ' + olderCount + ' earlier ' + (olderCount === 1 ? 'message' : 'messages') : 'Load earlier messages' }}</button></template>
<template v-for="(u, u__i) in cur.topics" :key="u__i">
<div :class="['ff-msg', u.mine ? 'ff-mine' : 'ff-other']" style="display: flex; gap: 12px; padding: 10px 14px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); touch-action: pan-y" @pointerdown="u.startHold" @pointermove="u.moveHold" @pointerup="u.stopHold" @pointerleave="u.stopHold" @pointercancel="u.stopHold" @contextmenu="u.openActions">
<UserAvatar :name="u.who" :initials="u.initials" :color="u.color" :size="40"/>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6px"><div style="display: flex; gap: 8px; align-items: baseline"><strong>{{ u.who }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ u.time }}</span><button type="button" @click.stop="u.openActions($event)" aria-label="Message options" style="margin-left: auto; width: 26px; height: 24px; padding: 0; border: 0; border-radius: 6px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer">⋯</button></div><template v-if="u.hasReply"><div style="margin: 4px 0 0; padding: 4px 8px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 4px; background: rgba(0,0,0,.2); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap"><strong style="color: var(--k-C9D6CF,#C9D6CF)">{{ u.replyWho }}</strong> {{ u.replySnippet }}</div></template><template v-if="u.hasText && !u.hasFile"><p v-html="u.htmlText" style="margin: 0; font-size: 15px; line-height: 1.5; overflow-wrap: anywhere"></p></template><template v-if="u.hasShared"><SharedPostCard :sid="u.sid"/></template>
<template v-if="u.hasFile"><MediaCollage :files="u.fileList" framed :max-h="380" :row="170"/></template><template v-if="u.hasText && u.hasFile"><p v-html="u.htmlText" style="margin: 6px 0 0; font-size: 15px; line-height: 1.5; overflow-wrap: anywhere"></p></template>
<template v-if="u.hasSecureShare"><SecureInboxCard :share="u.secureShare"/></template>
<div v-if="u.reply" style="margin-top: 6px"><button type="button" @click.stop="u.reply" aria-label="Reply to this message" style="height: 26px; padding: 0 10px; border: 0; border-radius: 13px; background: var(--k-22302A,#22302A); color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; font-size: 12px">↩ Reply<template v-if="u.replies"> · {{ u.replies }} earlier</template></button></div>

<div v-if="u.menuOpen" role="group" aria-label="Message actions" style="display: flex; flex-wrap: wrap; gap: 6px; margin-top: 8px; padding-top: 8px; border-top: 1px solid var(--k-405047,#405047)">
<button type="button" @click.stop="u.deleteForMe" style="height: 30px; padding: 0 10px; border: 1px solid var(--k-526259,#526259); border-radius: 8px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">Delete for me</button>
<button v-if="u.deleteForEveryone" type="button" @click.stop="u.deleteForEveryone" style="height: 30px; padding: 0 10px; border: 1px solid var(--k-6A3A34,#6A3A34); border-radius: 8px; background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer; font-size: 12px">Delete for everyone</button>
<button type="button" @click.stop="u.closeActions" aria-label="Close message actions" style="height: 30px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: var(--k-9DB0A6,#9DB0A6); cursor: pointer; font-size: 12px">Cancel</button>
</div></div>
</div>
</template>
</div>
</div>
</div>
</template>

<template v-if="ftab.sessions">
<div style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)">
<div style="display: flex; flex-direction: column; gap: 14px">
<div style="padding: 24px; border-radius: 16px; background: var(--k-1A2520,#1A2520); border: 1px solid var(--k-2A3832,#2A3832); display: flex; flex-wrap: wrap; gap: 16px; align-items: center">
<div style="flex: 1 1 280px; display: flex; flex-direction: column; gap: 4px"><strong style="font-size: 18px">Start a meeting or standup</strong><span style="color: var(--k-9DB0A6,#9DB0A6)">Video, audio or screen share with forum members. Recordings and notes are saved to the forum.</span></div>
<button v-if="canVideoCall" @click="videoCall" style="height: 44px; padding: 0 22px; border-radius: 22px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Go live now</button>
<button style="height: 44px; padding: 0 22px; border-radius: 22px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Schedule</button>
</div>
<div style="padding: 16px 18px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-wrap: wrap; gap: 12px; align-items: center"><div style="flex: 1 1 240px"><strong>Weekly check-in</strong><div style="color: var(--k-9DB0A6,#9DB0A6); font-size: 13px">Fridays, 10:00 · recurring</div></div><button style="height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">RSVP</button></div>
</div>
</div>
</div>
</template>

<template v-if="ftab.polls"><div style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)"><ForumPolls /></div></div></template>

<template v-if="ftab.members">
<div style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)">
<div style="display: flex; align-items: center; gap: 10px; margin-bottom: 12px; flex-wrap: wrap"><strong style="font-size: 15px; flex: 1">{{ cur.members }} {{ cur.members === 1 ? 'member' : 'members' }}</strong>
<button v-if="canAddToForum" type="button" @click="toggleAdd" :aria-expanded="addOpen" aria-label="Add a person to this forum" style="display: inline-flex; align-items: center; gap: 6px; height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; font-size: 13px; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg>Add</button></div>
<div v-if="canAddToForum && addOpen" style="margin-bottom: 14px; padding: 12px; border-radius: 12px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-B8F36B,#B8F36B); display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Search someone already in Hub. Picking them adds them to {{ cur.name }} straight away.</span>
<PeoplePicker instant :people="pickerPeople" :exclude="forumMemberIds" placeholder="Search people by name…" @pick="addPersonToForum"/>
<span v-if="addBusy" role="status" style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Adding…</span>
<span v-if="addErr" role="alert" style="font-size: 12px; color: #FF9B8A">{{ addErr }}</span></div>
<div style="margin-bottom: 12px"><ForumChatControls /></div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px">
<template v-for="(m, m__i) in chatCtl.members" :key="m__i">
<div style="padding: 14px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; gap: 12px; align-items: center">
<UserAvatar :name="m.name" :initials="m.initials" :color="m.color" :size="40"/>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong>{{ m.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ m.forumRole }} · <span :style="`color: ${m.stateColor}`">{{ m.stateLabel }}</span></span></div>
<template v-if="chatCtl.canManage"><button @click="m.toggle" :aria-label="`${m.actionLabel}: ${m.name}`" :style="`flex: none; height: 30px; padding: 0 12px; border-radius: 15px; cursor: pointer; font-size: 12px; ${m.can ? 'border: 1px solid var(--k-6A3A34,#6A3A34); background: transparent; color: var(--k-FF9B8A,#FF9B8A)' : 'border: 0; background: #B8F36B; color: #0F1714; font-weight: 600'}`">{{ m.actionLabel }}</button></template>
</div>
</template>
</div>
</div>
</div>
</template>

<template v-if="ftab.files">
<ChatSharedFiles :folders="fileFolders" :selected-folder="selectedChatFolder" :color="cur.color" :open-folder="openChatFolder" :close-folder="closeChatFolder"/>
</template>

<template v-if="ftab.about">
<div style="flex: 1; min-height: 0; overflow-y: auto"><div style="max-width: 1200px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 12px">
<div style="padding: 18px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); text-transform: uppercase; letter-spacing: 0.06em">Description</span><span>{{ cur.about }}</span></div>
<div style="padding: 18px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); text-transform: uppercase; letter-spacing: 0.06em">Who can join</span><span>{{ cur.privacy }}</span></div>
<div style="padding: 18px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); text-transform: uppercase; letter-spacing: 0.06em">Rules</span><span>Be respectful. Keep topics on-theme. No sharing secrets in chat; use Secure share.</span></div>
</div>
</div>
</div>
</template>

<template v-if="ftab.discussions">
<div style="flex: none; border-top: 1px solid var(--k-24322C,#24322C); background: var(--k-0F1714,#0F1714)">
<div v-if="cur.typingText" role="status" aria-live="polite" style="max-width: 1200px; margin: 0 auto; padding: 6px clamp(16px, 3vw, 32px) 0; font-size: 12px; color: var(--k-7CC4FF,#7CC4FF); font-style: italic">{{ cur.typingText }}</div>
<template v-if="replyBar.active"><div style="max-width: 1200px; margin: 0 auto; padding: 8px clamp(16px, 3vw, 32px) 0"><div style="display: flex; align-items: center; gap: 8px; margin: 0 0 6px; padding: 6px 10px; border-left: 3px solid var(--k-B8F36B,#B8F36B); border-radius: 6px; background: var(--k-16201C,#16201C); font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); min-width: 0"><span style="flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">Replying to <strong style="color: var(--k-E8EFEA,#E8EFEA)">{{ replyBar.who }}</strong>: {{ replyBar.snippet }}</span><button type="button" aria-label="Cancel reply" @click="replyBar.cancel" style="width: 24px; height: 24px; border: 0; border-radius: 12px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button></div></div></template>
<ChatAttachBar/>
<div class="hub-chat-composer" style="max-width: 1200px; margin: 0 auto; position: relative; padding: 12px clamp(16px, 3vw, 32px); display: grid; grid-template-columns: max-content minmax(0, 1fr) 42px; grid-template-rows: auto auto; gap: 6px; align-items: stretch">
<div style="grid-column: 1; grid-row: 1; display: flex; align-items: center; gap: 4px">
<label title="Attach files" style="position: relative; width: 38px; height: 38px; flex: none; border-radius: 50%; display: flex; align-items: center; justify-content: center; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer"><input type="file" multiple="" aria-label="Attach files" :disabled="composerOff" @change="attachForum" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg></label>
<EmojiButton :size="38" @pick="pickEmojiForum"/>
<button aria-label="Create a poll" title="Create a poll" :disabled="!pollView.canCreate" @click="pollView.compose" style="width: 38px; height: 38px; flex: none; border-radius: 50%; border: 0; background: transparent; color: var(--k-C9D6CF,#C9D6CF); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="19" height="19" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M5 20V10M12 20V4M19 20v-7"></path></svg></button>
 </div>
<RichTextInput toolbar-in-actions :mentions="mentionList" style="min-width: 0" :model-value="fdraft" @update:model-value="onFDraftValue" @keydown="fKey" :disabled="composerOff" :placeholder="composerOff ? composerNote : `Start a topic in ${cur.name}…`" aria-label="Start a topic" :rows="2" input-style="font-size: 14px; min-height: 42px"/>
<button aria-label="Post" title="Post" :disabled="composerOff" @click="sendForum" style="grid-column: 3; grid-row: 1 / 3; align-self: center; width: 42px; height: 42px; flex: none; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button>
</div>
</div>
</template>
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
  name: 'ForumFull',
  components: { CallScreen, ChatRail, ChatSharedFiles, SecureInboxCard },
  props: { fit: { type: String, default: 'screen' } },
  data() { return { moreOpen: false, railOpen: false }; },
  computed: {
    ...hubBindings(["canVoiceCall","canVideoCall","chatLoading","olderCount","moreOlder","showOlder","mentionList","onActivityClick","canAddToForum","addErr","addOpen","addBusy","toggleAdd","addPersonToForum","pickerPeople","forumMemberIds","pickEmojiForum","attachForum","chatRef","replyBar","chatCtl","closeChatFolder","composerNote","composerOff","cur","fKey","fdraft","fileFolders","ftab","has","inviteOpen","inviteText","inviteErr","inviteBusy","onInviteText","sendInvites","isFullForum","isPageForum","onFDraftValue","openChatFolder","pollView","selectedChatFolder","sendForum","showPanelCall","tab","tabTo","toClosed","toDocked","toFull","toMin","toPage","toggleInvite","videoCall","voiceCall"]).computed,
    // the forum's views as one switch in the header: Discussions, Polls, Calls, Members, Files, About
    views() {
      const t = this.ftab || {}, go = this.tabTo || {};
      return [['discussions', 'Discussions'], ['polls', 'Polls (' + ((this.pollView && this.pollView.count) || 0) + ')'], ['calls', 'Calls'], ['members', 'Members'], ['files', 'Files'], ['about', 'About']]
        .map(([k, label]) => ({ k, label, on: !!t[k], go: go[k] }));
    },
    shown() { return this.fit === 'page' ? !!this.isPageForum : !!this.isFullForum; },
    rootStyle() {
      const base = 'background: var(--k-0F1714,#0F1714); display: flex; flex-direction: row; overflow: hidden; box-sizing: border-box;';
      return this.fit === 'page'
        ? base + ' position: absolute; inset: 0; z-index: 30; border-radius: 0; border: 0; box-shadow: none'
        : base + ' position: fixed; inset: 0; z-index: 40';
    }
  }
};
</script>

<style>
/* phones and tablets (the sizes where a chat opens as a page, as in the side panel): the views move behind the vertical ⋮ so the name and the icon buttons keep their room */
.hub-seg-more { display: none !important; }
@media (max-width: 900px) {
  .hub-chat-bar .hub-seg { display: none !important; }
  .hub-chat-bar .hub-seg-more { display: block !important; }
  .hub-chat-bar .hub-sep { display: none !important; }
  .hub-chat-bar { flex-wrap: nowrap !important; gap: 6px !important; }
  .hub-chat-bar strong, .hub-chat-bar h1 { display: block; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
}
/* fitted to the page the chat is a page like any other: no close button, so its header needs no room for one */
.hub-fitpage .hub-chat-bar { padding-right: clamp(16px, 3vw, 32px) !important; }
/* The forum reads like a chat on every screen: my messages sit on the right, everyone else's on the left */
.ff-msg { box-sizing: border-box; width: fit-content; min-width: min(100%, 380px); max-width: 85%; }
.ff-mine { margin-left: auto; flex-direction: row-reverse; background: var(--k-2C3D34,#2C3D34) !important; border-color: var(--k-3A5446,#3A5446) !important; border-radius: 14px 4px 14px 14px !important; }
.ff-other { margin-right: auto; border-radius: 4px 14px 14px 14px !important; }
@media (max-width: 700px) {
  .ff-msg { padding: 8px 10px !important; gap: 8px !important; width: 88%; max-width: 88%; }
}
</style>
