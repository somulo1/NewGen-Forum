<!-- Voice / video call screen -->
<template><template v-if="visible">
<section aria-label="Meeting" :class="['cs-root', 'cs-' + variant, { 'cs-panel': compact, 'hub-dock': variant === 'dock' }]" :style="rootStyle">
<button v-if="variant === 'dock'" class="hub-grip" aria-label="Drag to resize panel. Arrow keys also resize." @pointerdown="startDrag" @keydown="resizeKey" style="position: absolute; left: 0; top: 0; bottom: 0; width: 14px; padding: 0; border: 0; background: transparent; cursor: ew-resize; z-index: 40; touch-action: none; display: flex; align-items: center; justify-content: center"><span :style="gripStyle"></span></button>
<div :style="`flex: none; display: flex; flex-wrap: ${compact ? 'nowrap' : 'wrap'}; align-items: center; gap: ${compact ? 8 : 12}px; padding: ${compact ? '10px 12px' : '12px 18px'}`">
<span :style="`width: 36px; height: 36px; border-radius: 10px; background: ${call.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 13px`">{{ call.initials }}</span>
<div :style="`flex: 1 1 ${compact ? 0 : 200}px; min-width: 0; display: flex; flex-direction: column`"><strong style="font-size: 15px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ call.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ call.modeLabel }} · {{ call.count }} people</span></div>
<template v-if="call.rec"><span style="height: 30px; padding: 0 12px; border-radius: 15px; background: #3A1E1A; color: #FF9B8A; font-size: 12px; font-weight: 700; display: flex; align-items: center; gap: 6px"><span style="width: 8px; height: 8px; border-radius: 50%; background: #FF9B8A"></span>REC</span></template>
<span style="height: 30px; padding: 0 12px; border-radius: 15px; background: #16201C; font-family: 'Space Grotesk', sans-serif; font-size: 14px; display: flex; align-items: center">{{ call.timer }}</span>
<template v-if="variant === 'main'"><button aria-label="Back to side card" title="Side card" @click="callToDock" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg></button><button aria-label="Minimize call" title="Minimize" @click="callToMin" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button></template>
<template v-if="variant === 'dock'"><button aria-label="Expand call" title="Expand" @click="callToFull" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button><button aria-label="Minimize call" title="Minimize" @click="callToMin" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button></template>
<template v-if="variant === 'panel'"><button aria-label="Expand to full screen" title="Expand" @click="toFull" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5"></path></svg></button><button aria-label="Minimize call" title="Minimize" @click="toMin" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button></template>
<template v-if="variant === 'full'"><button aria-label="Back to side panel" title="Side panel" @click="toDocked" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><rect x="3" y="4" width="18" height="16" rx="2"></rect><path d="M15 4v16"></path></svg></button><button aria-label="Minimize call" title="Minimize" @click="toMin" :style="winBtn"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M5 18h14"></path></svg></button></template>
</div>

<div :style="`flex: 1; min-height: 0; position: relative; display: flex; gap: 12px; padding: ${compact ? '0 12px 8px' : '0 18px 12px'}`">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 12px; overflow-y: auto">

<template v-if="call.isVideo">
<template v-if="call.sharing">
<div class="cs-share" style="flex: 1; min-height: 280px; border-radius: 18px; background: #16201C; border: 2px solid #B8F36B; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12px; padding: 20px; text-align: center">
<svg aria-hidden="true" width="44" height="44" viewBox="0 0 24 24" fill="none" stroke="#B8F36B" stroke-width="1.6" stroke-linejoin="round"><rect x="3" y="4" width="18" height="12" rx="2"></rect><path d="M8 20h8M12 16v4M12 7v6M9 10l3-3 3 3"></path></svg>
<strong style="font-size: 18px">You are presenting your screen</strong>
<span style="color: #9DB0A6; font-size: 13px">Everyone in the meeting can see it.</span>
<button @click="toggleShare" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Stop presenting</button>
</div>
<div style="display: flex; gap: 10px; overflow-x: auto">
<template v-for="(tl, tl__i) in call.tiles" :key="tl__i">
<div :style="tl.stripStyle"><span :style="`width: 40px; height: 40px; border-radius: 50%; background: ${tl.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 14px`">{{ tl.initials }}</span><span style="position: absolute; left: 8px; bottom: 6px; font-size: 11px">{{ tl.name }}</span></div>
</template>
</div>
</template>
<template v-if="call.notSharing">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(var(--cs-tile, 260px), 1fr)); gap: 12px; align-content: start">
<template v-for="(tl, tl__i) in call.tiles" :key="tl__i">
<div class="cs-tile" :style="tl.style">
<template v-if="tl.camOff"><span :style="`width: 76px; height: 76px; border-radius: 50%; background: ${tl.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 26px`">{{ tl.initials }}</span></template>
<template v-if="tl.camOn"><div :style="`position: absolute; inset: 0; background: ${tl.camBg}; display: flex; align-items: flex-end; justify-content: center`"><span :style="`width: 46%; aspect-ratio: 1 / 1.1; border-radius: 50% 50% 18% 18%; background: ${tl.color}; opacity: 0.55`"></span></div></template>
<span style="position: absolute; left: 10px; bottom: 10px; height: 28px; padding: 0 10px; border-radius: 14px; background: rgba(10,16,13,0.75); font-size: 12px; display: flex; align-items: center; gap: 6px">{{ tl.name }}<template v-if="tl.muted"><svg aria-label="Muted" role="img" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#FF9B8A" stroke-width="2" stroke-linecap="round"><path d="M9 9v2a3 3 0 0 0 5 2.2M15 10V5a3 3 0 0 0-6 0M5 11a7 7 0 0 0 11.5 5.3M19 11a7 7 0 0 1-.5 2.5M12 18v3M3 3l18 18"></path></svg></template></span>
<template v-if="tl.hand"><span style="position: absolute; top: 10px; left: 10px; height: 28px; padding: 0 10px; border-radius: 14px; background: #FFD66B; color: #0F1714; font-size: 12px; font-weight: 700; display: flex; align-items: center">Hand raised</span></template>
<template v-if="tl.speaking"><span style="position: absolute; top: 10px; right: 10px; font-size: 11px; font-weight: 600; color: #B8F36B">Speaking</span></template>
</div>
</template>
</div>
</template>
</template>

<template v-if="call.isVoice">
<div style="flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 28px; padding: 20px">
<span style="font-size: 13px; color: #9DB0A6">Voice call</span>
<div style="display: flex; flex-wrap: wrap; gap: 36px; justify-content: center">
<template v-for="(tl, tl__i) in call.tiles" :key="tl__i">
<div style="display: flex; flex-direction: column; align-items: center; gap: 10px">
<span class="cs-voice" :style="tl.voiceStyle">{{ tl.initials }}</span>
<span style="font-size: 14px; display: flex; align-items: center; gap: 6px">{{ tl.name }}<template v-if="tl.muted"><svg aria-label="Muted" role="img" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#FF9B8A" stroke-width="2" stroke-linecap="round"><path d="M9 9v2a3 3 0 0 0 5 2.2M15 10V5a3 3 0 0 0-6 0M5 11a7 7 0 0 0 11.5 5.3M19 11a7 7 0 0 1-.5 2.5M12 18v3M3 3l18 18"></path></svg></template></span>
<template v-if="tl.hand"><span style="font-size: 11px; font-weight: 700; color: #FFD66B">Hand raised</span></template>
</div>
</template>
</div>
</div>
</template>
</div>

<template v-if="call.panelOpen">
<aside :aria-label="call.panelTitle" :style="sideStyle">
<div style="flex: none; display: flex; align-items: center; padding: 14px 16px; border-bottom: 1px solid #24322C"><strong style="flex: 1">{{ call.panelTitle }}</strong><button aria-label="Close panel" @click="closeCallPanel" style="width: 34px; height: 34px; border-radius: 50%; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<template v-if="call.showPeople">
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 12px 16px; display: flex; flex-direction: column; gap: 10px">
<template v-for="(tl, tl__i) in call.tiles" :key="tl__i"><div style="display: flex; gap: 10px; align-items: center"><span :style="`width: 34px; height: 34px; border-radius: 50%; background: ${tl.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ tl.initials }}</span><span style="flex: 1; font-size: 14px">{{ tl.name }}</span><span style="font-size: 12px; color: #9DB0A6">{{ tl.stateLabel }}</span></div></template>
<button style="margin-top: 6px; height: 38px; border-radius: 10px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">Invite people</button>
</div>
</template>
<template v-if="call.showChat">
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 12px 16px; display: flex; flex-direction: column; gap: 10px">
<template v-for="(cm, cm__i) in call.msgs" :key="cm__i"><div style="display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 12px">{{ cm.who }}</strong><span style="font-size: 14px; line-height: 1.4; padding: 8px 10px; border-radius: 4px 12px 12px 12px; background: #1F2C26">{{ cm.text }}</span></div></template>
</div>
<div style="flex: none; display: flex; gap: 8px; padding: 10px 12px; border-top: 1px solid #24322C"><label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Message everyone</span><input type="text" :value="call.draft" @input="onCallDraft" @keydown="callKey" placeholder="Message everyone" style="width: 100%; box-sizing: border-box; height: 40px; padding: 0 12px; border-radius: 20px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label><button aria-label="Send" @click="sendCallMsg" style="width: 40px; height: 40px; border-radius: 50%; border: 0; background: #B8F36B; color: #0F1714; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M4 12l16-8-6 16-3-7z"></path></svg></button></div>
</template>
</aside>
</template>
</div>

<div :style="`flex: none; display: flex; flex-wrap: wrap; justify-content: center; gap: ${compact ? 6 : 10}px; padding: ${compact ? '8px 10px 12px' : '12px 18px 18px'}`">
<button :aria-label="call.micLabel" @click="toggleMic" class="cs-ctl" :style="call.micStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round"><rect x="9" y="3" width="6" height="11" rx="3"></rect><path d="M5 11a7 7 0 0 0 14 0M12 18v3"></path></svg><span style="font-size: 11px">{{ call.micLabel }}</span></button>
<template v-if="call.isVideo"><button :aria-label="call.camLabel" @click="toggleCam" class="cs-ctl" :style="call.camStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3z"></path></svg><span style="font-size: 11px">{{ call.camLabel }}</span></button></template>
<button class="cs-ctl" aria-label="Share screen" @click="toggleShare" :style="call.shareStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"><rect x="3" y="4" width="18" height="12" rx="2"></rect><path d="M8 20h8M12 16v4"></path></svg><span style="font-size: 11px">Share</span></button>
<button class="cs-ctl" aria-label="Raise hand" @click="toggleHand" :style="call.handStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round" stroke-linejoin="round"><path d="M8 13V5.5a1.5 1.5 0 0 1 3 0V12M11 11V4a1.5 1.5 0 0 1 3 0v7M14 11V5.5a1.5 1.5 0 0 1 3 0V14a7 7 0 0 1-7 7 6 6 0 0 1-5-3l-2.5-4a1.5 1.5 0 0 1 2.6-1.5L8 15"></path></svg><span style="font-size: 11px">Hand</span></button>
<button class="cs-ctl" aria-label="Record" @click="toggleRec" :style="call.recStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9"><circle cx="12" cy="12" r="8"></circle><circle cx="12" cy="12" r="3.5" fill="currentColor"></circle></svg><span style="font-size: 11px">Record</span></button>
<button class="cs-ctl" aria-label="Chat" @click="openCallChat" :style="call.chatStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linejoin="round"><path d="M4 5h16v11H8l-4 4z"></path></svg><span style="font-size: 11px">Chat</span></button>
<button class="cs-ctl" aria-label="People" @click="openCallPeople" :style="call.peopleStyle"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.9" stroke-linecap="round"><circle cx="9" cy="8" r="3.5"></circle><path d="M2 20c0-3.5 3-6 7-6s7 2.5 7 6M16 4.5a3.5 3.5 0 0 1 0 7M18 14c2.5.6 4 2.8 4 6"></path></svg><span style="font-size: 11px">People</span></button>
<button class="cs-leave" @click="endCall" style="height: 60px; padding: 0 26px; border-radius: 30px; border: 0; background: #E5484D; color: #fff; font-weight: 700; font-size: 15px; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linejoin="round"><path d="M3 14.5c5-4.5 13-4.5 18 0l-2 2.5-3.5-1.5V13a12 12 0 0 0-7 0v2.5L5 17z"></path></svg>Leave</button>
</div>
</section>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CallScreen',
  props: { variant: { type: String, default: 'main' } },
  computed: {
    ...hubBindings(["call","callKey","closeCallPanel","endCall","onCallDraft","openCallChat","openCallPeople","sendCallMsg","gripStyle","panelStyle","resizeKey","showMeetingDock","showMeetingFull","showPanelCall","startDrag","callToDock","callToFull","callToMin","toDocked","toFull","toMin","toggleCam","toggleHand","toggleMic","toggleRec","toggleShare"]).computed,
    visible() { return this.variant === 'main' ? !!this.showMeetingFull : this.variant === 'dock' ? !!this.showMeetingDock : !!this.showPanelCall; },
    compact() { return this.variant === 'panel' || this.variant === 'dock'; },
    winBtn() { return 'width: 34px; height: 34px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center; flex: none'; },
    rootStyle() {
      const base = 'background: #0A100D; color: #E8EFEA; display: flex; flex-direction: column; overflow: hidden; box-sizing: border-box; z-index: 30; position: absolute;';
      if (this.variant === 'main') return base + ' inset: 12px; border-radius: 22px; border: 1px solid #2A3832; box-shadow: 0 12px 36px rgba(0,0,0,0.35)';
      if (this.variant === 'dock') return this.panelStyle + ';background: #0A100D;color: #E8EFEA';
      return base + ' inset: 0; border-radius: inherit';
    },
    sideStyle() {
      const card = 'background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; overflow: hidden; border-radius: 18px; ';
      if (this.compact) return card + 'position: absolute; left: 8px; right: 8px; top: 0; bottom: 8px; z-index: 5';
      return card + 'flex: 0 0 340px; max-width: 45%';
    }
  }
};
</script>

<style>
.cs-main, .cs-full { --cs-tile: 260px; }
.cs-panel { --cs-tile: 120px; }
.cs-panel .cs-tile { min-height: 0 !important; border-radius: 12px !important; }
.cs-panel .cs-share { min-height: 150px !important; }
.cs-panel .cs-voice { width: 64px !important; height: 64px !important; font-size: 22px !important; }
.cs-panel .cs-ctl { width: 44px !important; height: 44px !important; border-radius: 14px !important; }
.cs-panel .cs-ctl span { display: none; }
.cs-panel .cs-leave { height: 44px !important; padding: 0 16px !important; border-radius: 22px !important; font-size: 13px !important; }
@media (max-width: 900px) { .cs-main { inset: 0 0 64px 0 !important; border-radius: 0 !important; border: 0 !important; } }
</style>
