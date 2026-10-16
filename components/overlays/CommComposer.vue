<!-- Memo / announcement / newsletter composer -->
<template><template v-if="cmOpen">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 66; background: rgba(5,10,8,0.65); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box">
<div role="dialog" :aria-label="cm.heading" style="width: 100%; max-width: 660px; max-height: 92vh; display: flex; flex-direction: column; border-radius: 20px; background: #16201C; border: 1px solid #2A3832; box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 16px 20px; border-bottom: 1px solid #24322C"><h2 style="margin: 0; flex: 1; font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ cm.heading }}</h2><button aria-label="Close" @click="cm.close" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(ct, ct__i) in cm.types" :key="ct__i"><button @click="ct.pick" :style="ct.style">{{ ct.label }}</button></template></div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Title<input type="text" :value="cm.title" @input="cm.setTitle" placeholder="Give it a clear title" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 15px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Message<textarea rows="5" :value="cm.body" @input="cm.setBody" placeholder="Write the message" style="padding: 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; line-height: 1.5; resize: vertical"></textarea></label>
<div style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: #121B17; border: 1px solid #24322C">
<strong style="font-size: 14px">Where should it go?</strong>
<template v-if="cm.wallAllowed"><label style="display: flex; gap: 10px; align-items: center; font-size: 14px; cursor: pointer"><input type="checkbox" :checked="cm.wall" @change="cm.toggleWall">Community wall</label></template>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">A forum<select :value="cm.forum" @change="cm.setForum" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><option value="">No forum</option><template v-for="(f, f__i) in forums" :key="f__i"><option :value="f.id">{{ f.name }}</option></template></select></label>
<template v-if="cm.hasRuleNote"><span style="font-size: 12px; color: #C9D6CF; line-height: 1.5; padding: 8px 10px; border-radius: 8px; background: #16201C">{{ cm.ruleNote }}</span></template>
<template v-for="(gr, gr__i) in cm.groups" :key="gr__i">
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: #C9D6CF">{{ gr.label }}</span><div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(gi, gi__i) in gr.items" :key="gi__i"><button @click="gi.toggle" :aria-pressed="gi.on" :style="gi.style">{{ gi.name }}</button></template></div></div>
</template>
<template v-if="cm.noStructure"><span style="font-size: 12px; color: #9DB0A6">Department, sector, branch and unit options appear here once your organization adds those widgets.</span></template>
</div>
<template v-if="cm.needsApproval"><span style="font-size: 13px; color: #FFB86B">{{ cm.approvalNote }}</span></template>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid #24322C"><button @click="cm.close" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button><button @click="cm.publish" :style="cm.publishStyle">{{ cm.publishLabel }}</button></div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CommComposer',
  ...hubBindings(["cm","cmOpen","forums"])
};
</script>
