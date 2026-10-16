<!-- Polls for the open forum. Results update live while the page is open. -->
<template>
<div style="display: flex; flex-direction: column; gap: 12px; min-width: 0">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px">
<span style="flex: 1 1 auto; font-size: 12px; color: #B8F36B; font-weight: 600">● Results update live</span>
<template v-if="pollView.canCreate"><button v-if="!pollView.form.open" @click="pollView.openForm" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">New poll</button></template>
<template v-if="!pollView.canCreate"><span style="font-size: 12px; color: #FFB86B">{{ pollView.createHint }}</span></template>
</div>

<template v-if="pollView.form.open">
<form novalidate @submit.prevent="pollView.form.submit" aria-label="New poll" style="padding: 14px; border-radius: 14px; background: #121B17; border: 1px solid #B8F36B; display: flex; flex-direction: column; gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Question<input type="text" :value="pollView.form.question" @input="pollView.form.onQuestion" placeholder="What do you want to ask?" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<template v-for="(op, op__i) in pollView.form.options" :key="op__i">
<div style="display: flex; gap: 6px"><input type="text" :aria-label="op.placeholder" :value="op.value" @input="op.onInput" :placeholder="op.placeholder" style="flex: 1; min-width: 0; height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><template v-if="op.canRemove"><button type="button" :aria-label="`Remove ${op.placeholder}`" @click="op.remove" style="width: 38px; height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></template></div>
</template>
<template v-if="pollView.form.canAdd"><button type="button" @click="pollView.form.addOption" style="align-self: flex-start; height: 32px; padding: 0 14px; border-radius: 16px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">Add option</button></template>
<template v-if="pollView.form.hasErr"><span role="alert" style="font-size: 13px; color: #FF9B8A">{{ pollView.form.err }}</span></template>
<div style="display: flex; gap: 8px"><button type="submit" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create poll</button><button type="button" @click="pollView.form.cancel" style="height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></div>
</form>
</template>

<template v-if="pollView.noPolls"><div style="padding: 24px; border-radius: 14px; border: 1px dashed #3F524A; text-align: center; color: #9DB0A6; font-size: 14px">No polls in this forum yet.</div></template>

<template v-for="(pl, pl__i) in pollView.polls" :key="pl.id">
<article :aria-label="pl.question" style="padding: 14px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-start; gap: 8px">
<div style="flex: 1 1 160px; min-width: 0; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px; line-height: 1.35">{{ pl.question }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ pl.by }} · {{ pl.when }}</span></div>
<span :style="`height: 24px; padding: 0 10px; border-radius: 12px; font-size: 11px; font-weight: 700; display: flex; align-items: center; background: ${pl.statusColor}22; color: ${pl.statusColor}`">{{ pl.statusLabel }}</span>
<template v-if="pl.canControl"><button :aria-label="`Poll controls for ${pl.question}`" aria-haspopup="menu" @click="pl.toggleMenu" style="width: 30px; height: 30px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">⋯</button></template>
</div>

<template v-if="pl.menuOpen">
<div role="menu" style="padding: 10px; border-radius: 12px; background: #1D2A25; border: 1px solid #34453D; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; flex-wrap: wrap; gap: 8px"><button role="menuitem" @click="pl.toggleLock" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">{{ pl.lockLabel }}</button><button role="menuitem" @click="pl.remove" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #6A3A34; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 13px">Delete poll</button></div>
<span style="font-size: 12px; color: #9DB0A6">Block specific people from voting on this poll{{ pl.blockedCount ? ' · ' + pl.blockedCount + ' blocked' : '' }}</span>
<template v-for="(vt, vt__i) in pl.voters" :key="vt__i"><div style="display: flex; align-items: center; gap: 8px"><span aria-hidden="true" :style="`width: 26px; height: 26px; border-radius: 50%; background: ${vt.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 10px`">{{ vt.initials }}</span><span style="flex: 1; font-size: 13px">{{ vt.name }}</span><button @click="vt.toggle" :style="`height: 28px; padding: 0 12px; border-radius: 14px; cursor: pointer; font-size: 12px; ${vt.blocked ? 'border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600' : 'border: 1px solid #2A3832; background: transparent; color: #E8EFEA'}`">{{ vt.label }}</button></div></template>
</div>
</template>

<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(op, op__i) in pl.options" :key="op.id">
<button :disabled="!op.canVote" :aria-pressed="op.mine" @click="op.vote" :style="op.rowStyle">
<span style="display: flex; align-items: center; gap: 8px; font-size: 14px"><span style="flex: 1; text-align: left">{{ op.label }}<template v-if="op.mine"> ✓</template></span><span style="font-size: 12px; color: #C9D6CF">{{ op.count }} · {{ op.pct }}%</span></span>
<span style="display: block; height: 6px; border-radius: 3px; background: #22302A; overflow: hidden"><span :style="op.barStyle" style="display: block"></span></span>
</button>
</template>
</div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: #C9D6CF">{{ pl.totalLabel }}</span><span style="font-size: 12px; color: #9DB0A6">{{ pl.note }}</span></div>
</article>
</template>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ForumPolls',
  ...hubBindings(["pollView"])
};
</script>
