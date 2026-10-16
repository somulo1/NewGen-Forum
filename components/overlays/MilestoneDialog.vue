<!-- Milestone dialog: create a milestone, or close / reopen one with a required comment -->
<template><template v-if="msd.open">
<div style="position: fixed; inset: 0; z-index: 70; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px">
<form role="dialog" aria-modal="true" novalidate :aria-label="msd.title" @submit.prevent="msd.submit" style="width: 520px; max-width: 100%; max-height: 100%; overflow-y: auto; padding: 22px; border-radius: 20px; background: #16201C; border: 1px solid #2A3832; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; align-items: center; gap: 12px"><h2 style="flex: 1; margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">{{ msd.title }}</h2><button type="button" aria-label="Close" @click="msd.cancel" style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div>

<template v-if="msd.isNew">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Project
<template v-if="msd.lockProject"><span style="height: 40px; padding: 0 12px; border-radius: 10px; background: #121B17; border: 1px solid #2A3832; color: #E8EFEA; font-size: 14px; display: flex; align-items: center">{{ msd.projectName }}</span></template>
<template v-if="!msd.lockProject"><select :value="msd.pid" @change="msd.onProject" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option value="">Choose a project…</option><template v-for="(po, po__i) in msd.projectOpts" :key="po__i"><option :value="po.v">{{ po.l }}</option></template></select></template>
</label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Milestone name<input type="text" :value="msd.name" @input="msd.onName" placeholder="For example: Beta release" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Due date<input type="date" :value="msd.date" :min="msd.minDate" :max="msd.maxDate" @input="msd.onDate" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><span style="font-size: 12px">{{ msd.rangeNote }}</span></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">What will be true when this is reached? (optional)<textarea rows="3" :value="msd.desc" @input="msd.onDesc" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; line-height: 1.45; resize: vertical"></textarea></label>
<span style="font-size: 12px; color: #9DB0A6">An objective is created for this milestone automatically. Tasks you add under it count toward the milestone's progress.</span>
</template>

<template v-if="!msd.isNew">
<div style="padding: 12px 14px; border-radius: 12px; background: #121B17; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">{{ msd.msName }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ msd.msSummary }}</span></div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">{{ msd.commentLabel }}<textarea rows="4" :value="msd.comment" @input="msd.onComment" placeholder="Write a short note. It stays with the milestone and appears in the project activity." style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; line-height: 1.45; resize: vertical"></textarea></label>
</template>

<template v-if="msd.hasErr"><span role="alert" style="font-size: 13px; color: #FF9B8A">{{ msd.err }}</span></template>
<div style="display: flex; gap: 8px; justify-content: flex-end"><button type="button" @click="msd.cancel" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button><button type="submit" style="height: 40px; padding: 0 20px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ msd.submitLabel }}</button></div>
</form>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'MilestoneDialog',
  ...hubBindings(["msd"])
};
</script>
