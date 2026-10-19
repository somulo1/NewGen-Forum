<!-- Moderator controls on a participant: turn their mic or camera OFF. Nobody can switch another person's mic or camera ON. -->
<template>
<span style="display: inline-flex; gap: 4px">
<button type="button" :disabled="!tile.micOn" @click.stop="tile.muteMic" :aria-label="tile.micOn ? `Mute ${tile.name}` : `${tile.name} is muted. Only they can unmute`" :title="tile.micOn ? `Mute ${tile.name}` : `${tile.name} is muted. Only they can unmute`" :style="btn(tile.micOn)">
<svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 3a3 3 0 0 0-3 3v5a3 3 0 0 0 6 0V6a3 3 0 0 0-3-3zM5 11a7 7 0 0 0 14 0M12 18v3"></path><path v-if="!tile.micOn" d="M3 3l18 18"></path></svg>
</button>
<button v-if="tile.camAvail" type="button" :disabled="!tile.camOnNow" @click.stop="tile.stopCam" :aria-label="tile.camOnNow ? `Turn off ${tile.name}'s video` : `${tile.name}'s video is off. Only they can turn it on`" :title="tile.camOnNow ? `Turn off ${tile.name}'s video` : `${tile.name}'s video is off. Only they can turn it on`" :style="btn(tile.camOnNow)">
<svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="6" width="13" height="12" rx="2"></rect><path d="M16 10l5-3v10l-5-3"></path><path v-if="!tile.camOnNow" d="M3 3l18 18"></path></svg>
</button>
</span>
</template>

<script>
export default {
  name: 'ModControls',
  props: { tile: { type: Object, required: true } },
  methods: { btn(on) { return `width: 28px; height: 28px; border-radius: 50%; border: 0; display: flex; align-items: center; justify-content: center; background: ${on ? 'rgba(10,16,13,0.8)' : 'rgba(255,155,138,0.22)'}; color: ${on ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-FF9B8A,#FF9B8A)'}; cursor: ${on ? 'pointer' : 'not-allowed'}`; } }
};
</script>
