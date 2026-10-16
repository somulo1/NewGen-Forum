<!-- Ticket categories a unit attends to, as one compact table: TAT, priority and the people who always solve them -->
<template><template v-if="catMgr.active">
<div style="display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px">
<div style="flex: 1 1 320px; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">{{ catMgr.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ catMgr.note }}</span></div>
<template v-if="catMgr.canManage && !catMgr.usingDefaults"><div style="display: flex; align-items: center; gap: 8px"><span :style="`font-size: 12px; color: ${catMgr.dirty ? '#FFB86B' : '#9DB0A6'}`">{{ catMgr.dirty ? 'Unsaved changes' : 'All saved' }}</span><button :disabled="!catMgr.dirty" @click="catMgr.discard" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px; opacity: 1">Discard</button><button :disabled="!catMgr.dirty" @click="catMgr.save" :style="`height: 36px; padding: 0 18px; border-radius: 18px; border: 0; font-weight: 600; font-size: 13px; cursor: ${catMgr.dirty ? 'pointer' : 'default'}; ${catMgr.dirty ? 'background: #B8F36B; color: #0F1714' : 'background: #2A3832; color: #9DB0A6'}`">Save changes</button></div></template>
<template v-if="!catMgr.canManage"><span style="font-size: 12px; color: #FFB86B; flex: 1 1 100%">{{ catMgr.noAccessNote }}</span></template>
<template v-if="catMgr.canManage && catMgr.usingDefaults"><button @click="catMgr.startOwn" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">Create our own categories</button></template>
</div>
<div style="overflow-x: auto; border-radius: 12px; border: 1px solid #24322C; background: #16201C">
<table aria-label="Ticket categories" style="width: 100%; min-width: 760px; border-collapse: collapse; font-size: 13px">
<thead><tr>
<th scope="col" style="padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; background: #1B2822">Category</th>
<th scope="col" style="padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; background: #1B2822; width: 96px">TAT (h)</th>
<th scope="col" style="padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; background: #1B2822; width: 110px">Priority</th>
<th scope="col" style="padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; background: #1B2822">Assigned to solve</th>
<th scope="col" style="padding: 8px 12px; background: #1B2822; width: 80px"></th>
</tr></thead>
<tbody>
<template v-for="(c, c__i) in catMgr.rows" :key="c.id">
<tr :aria-label="c.name">
<td style="padding: 6px 12px; border-top: 1px solid #24322C; font-weight: 600; vertical-align: middle">{{ c.name }}</td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C; vertical-align: middle"><input type="number" min="1" :aria-label="`TAT hours for ${c.name}`" :value="c.tat" :disabled="!catMgr.canManage || !c.custom" @change="c.setTat" style="width: 64px; height: 30px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C; vertical-align: middle"><select :aria-label="`Priority for ${c.name}`" :value="c.prio" :disabled="!catMgr.canManage || !c.custom" @change="c.setPrio" style="height: 30px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"><option>High</option><option>Medium</option><option>Low</option></select></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C; vertical-align: middle">
<div style="display: flex; align-items: center; gap: 6px; flex-wrap: nowrap; overflow: hidden">
<template v-for="(m, m__i) in c.summary" :key="m.name"><span :title="m.name" :style="`flex: none; height: 26px; padding: 0 8px 0 3px; border-radius: 13px; background: #121B17; border: 1px solid #2A3832; display: flex; align-items: center; gap: 5px; font-size: 12px; white-space: nowrap`"><span :style="`width: 20px; height: 20px; border-radius: 50%; background: ${m.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 8px`">{{ m.initials }}</span>{{ m.name.split(' ')[0] }}</span></template>
<template v-if="c.extra"><span style="flex: none; font-size: 12px; color: #C9D6CF">+{{ c.extra }} more</span></template>
<template v-if="!c.count"><span style="font-size: 12px; color: #FFB86B">Nobody yet</span></template>
<template v-if="catMgr.canManage && c.custom"><button :aria-label="`Manage who solves ${c.name}`" @click="c.manage" style="flex: none; height: 26px; padding: 0 12px; border-radius: 13px; border: 1px solid #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">{{ c.count ? 'Manage (' + c.count + ')' : 'Assign' }}</button></template>
</div>
</td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C; text-align: right; vertical-align: middle"><template v-if="catMgr.canManage && c.custom"><button :aria-label="`Remove ${c.name}`" @click="c.remove" style="height: 26px; padding: 0 10px; border-radius: 13px; border: 1px solid #6A3A34; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 12px">Remove</button></template></td>
</tr>
</template>
<template v-if="catMgr.canManage && !catMgr.usingDefaults">
<tr style="background: #121B17">
<td style="padding: 6px 12px; border-top: 1px solid #24322C"><input type="text" aria-label="New category name" :value="catMgr.draft.name" @input="catMgr.draft.onName" @keydown.enter.prevent="catMgr.draft.add" placeholder="New category, e.g. Payroll query" style="width: 100%; box-sizing: border-box; height: 30px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C"><input type="number" min="1" aria-label="New category TAT hours" :value="catMgr.draft.tat" @input="catMgr.draft.onTat" style="width: 64px; height: 30px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C"><select aria-label="New category priority" :value="catMgr.draft.prio" @change="catMgr.draft.onPrio" style="height: 30px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><option>High</option><option>Medium</option><option>Low</option></select></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C"><template v-if="catMgr.draft.hasErr"><span role="alert" style="font-size: 12px; color: #FF9B8A">{{ catMgr.draft.err }}</span></template><template v-if="!catMgr.draft.hasErr"><span style="font-size: 12px; color: #9DB0A6">Add it, then assign one or many people to solve it.</span></template></td>
<td style="padding: 6px 12px; border-top: 1px solid #24322C; text-align: right"><button @click="catMgr.draft.add" style="height: 28px; padding: 0 14px; border-radius: 14px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 12px">Add</button></td>
</tr>
</template>
</tbody>
</table>
</div>
<template v-if="catMgr.editor.open">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 70; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px">
<div role="dialog" aria-modal="true" :aria-label="`Assign people to ${catMgr.editor.name}`" style="width: 480px; max-width: 100%; max-height: 100%; display: flex; flex-direction: column; gap: 12px; padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #2A3832">
<div style="display: flex; align-items: center; gap: 10px"><div style="flex: 1; display: flex; flex-direction: column"><strong style="font-size: 16px">Who solves {{ catMgr.editor.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ catMgr.editor.count }} selected. Tick one or many, then apply. Save changes in the table makes it permanent.</span></div><button aria-label="Close" @click="catMgr.editor.cancel" style="width: 34px; height: 34px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">✕</button></div>
<input type="search" aria-label="Search people" :value="catMgr.editor.q" @input="catMgr.editor.onQ" placeholder="Search people…" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<div tabindex="0" aria-label="People to assign" style="max-height: 260px; overflow-y: auto; border-radius: 12px; border: 1px solid #24322C; background: #121B17">
<template v-for="(r, r__i) in catMgr.editor.results" :key="r.name"><label style="display: flex; align-items: center; gap: 10px; min-height: 40px; padding: 0 12px; cursor: pointer; border-top: 1px solid #1C2924"><input type="checkbox" :checked="r.on" @change="r.toggle" style="accent-color: #B8F36B"><span :style="`width: 26px; height: 26px; border-radius: 50%; background: ${r.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 10px`">{{ r.initials }}</span><span style="font-size: 14px">{{ r.name }}</span></label></template>
<template v-if="!catMgr.editor.results.length"><span style="display: block; padding: 14px; font-size: 13px; color: #9DB0A6">Nobody matches.</span></template>
</div>
<template v-if="catMgr.editor.count"><div tabindex="0" aria-label="Selected people" style="max-height: 70px; overflow-y: auto; display: flex; flex-wrap: wrap; gap: 4px"><template v-for="(m, m__i) in catMgr.editor.selected" :key="m.name"><span style="height: 24px; padding: 0 4px 0 8px; border-radius: 12px; background: #2C3D34; display: flex; align-items: center; gap: 4px; font-size: 12px">{{ m.name }}<button :aria-label="`Remove ${m.name}`" @click="m.remove" style="width: 18px; height: 18px; border-radius: 50%; border: 0; background: #3F524A; color: #E8EFEA; cursor: pointer; font-size: 9px; padding: 0">✕</button></span></template></div></template>
<div style="display: flex; gap: 8px; justify-content: flex-end"><button @click="catMgr.editor.cancel" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button><button :disabled="!catMgr.editor.dirty" @click="catMgr.editor.save" :style="`height: 38px; padding: 0 20px; border-radius: 19px; border: 0; font-weight: 600; cursor: ${catMgr.editor.dirty ? 'pointer' : 'default'}; ${catMgr.editor.dirty ? 'background: #B8F36B; color: #0F1714' : 'background: #2A3832; color: #9DB0A6'}`">Apply</button></div>
</div>
</div>
</template>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketCategories',
  ...hubBindings(["catMgr"])
};
</script>
