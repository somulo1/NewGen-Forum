<!-- One participant tile in a call: video or avatar, name, mute/hand/speaking marks, host controls, and the enlarge button -->
<template>
<div class="cs-tile" :style="tl.style + (big ? ';aspect-ratio:auto;width:100%;height:100%;min-height:0' : '')" @dblclick="$emit('fullscreen', $event)"><button type="button" class="cs-expand" :aria-label="big ? 'Back to the grid' : 'Enlarge ' + tl.name" :title="big ? 'Back to the grid' : 'Enlarge ' + tl.name" @click.stop="$emit('focus')"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path :d="big ? 'M9 4v5H4M15 4v5h5M9 20v-5H4M15 20v-5h5' : 'M4 9V4h5M20 9V4h-5M4 15v5h5M20 15v5h-5'"></path></svg></button>
<template v-if="tl.camOff"><UserAvatar :name="tl.name" :initials="tl.initials" :color="tl.color" :size="76" :font-size="26"/></template>
<template v-if="tl.camOn && tl.videoKey"><CallVideo :stream-key="tl.videoKey"/></template><template v-if="tl.camOn && !tl.videoKey"><div :style="`position: absolute; inset: 0; background: ${tl.camBg}; display: flex; align-items: flex-end; justify-content: center`"><span :style="`width: 46%; aspect-ratio: 1 / 1.1; border-radius: 50% 50% 18% 18%; background: ${tl.color}; opacity: 0.55`"></span></div></template>
<span style="position: absolute; left: 10px; bottom: 10px; height: 28px; padding: 0 10px; border-radius: 14px; background: rgba(10,16,13,0.75); font-size: 12px; display: flex; align-items: center; gap: 6px">{{ tl.name }}<template v-if="tl.muted"><svg aria-label="Muted" role="img" width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="var(--k-FF9B8A,#FF9B8A)" stroke-width="2" stroke-linecap="round"><path d="M9 9v2a3 3 0 0 0 5 2.2M15 10V5a3 3 0 0 0-6 0M5 11a7 7 0 0 0 11.5 5.3M19 11a7 7 0 0 1-.5 2.5M12 18v3M3 3l18 18"></path></svg></template></span>
<template v-if="tl.netPaused"><span style="position: absolute; top: 10px; left: 10px; font-size: 11px; font-weight: 600; color: var(--k-FFB86B,#FFB86B)">Poor connection</span></template><template v-if="tl.hand"><span style="position: absolute; top: 10px; left: 10px; height: 28px; padding: 0 10px; border-radius: 14px; background: #FFD66B; color: #0F1714; font-size: 12px; font-weight: 700; display: flex; align-items: center">Hand raised</span></template>
<template v-if="tl.speaking"><span style="position: absolute; top: 10px; right: 10px; font-size: 11px; font-weight: 600; color: var(--k-B8F36B,#B8F36B)">Speaking</span></template>
<template v-if="tl.mod"><span style="position: absolute; right: 8px; bottom: 8px"><ModControls :tile="tl"/></span></template>
</div>
</template>

<script>
export default {
  name: 'CallTile',
  props: { tl: { type: Object, required: true }, big: { type: Boolean, default: false } },
  emits: ['focus', 'fullscreen']
};
</script>
