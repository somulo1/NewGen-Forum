<!-- Manage organization structure: add, rename, move, merge and remove sectors, branches, units and departments. -->
<template><template v-if="structureEd.open">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
<button type="button" @click="structureEd.close" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">‹ People</button>
<div style="flex: 1 1 260px; min-width: 0"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Organization structure</h1><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ structureEd.summary }}</span></div>
<template v-if="structureEd.busy"><span role="status" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Saving…</span></template>
</div>
<div style="display: flex; flex-direction: column; gap: 16px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); line-height: 1.5">{{ structureEd.note }}</span>

<form novalidate aria-label="Add to the structure" @submit.prevent="structureEd.add" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17)">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Add a
<select :value="structureEd.kind" @change="structureEd.onKind" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="k in structureEd.kinds" :key="k.kind"><option :value="k.kind">{{ k.label }}</option></template></select>
</label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Name
<input type="text" :value="structureEd.name" @input="structureEd.onName" maxlength="80" placeholder="e.g. Nairobi" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"/>
</label>
<template v-if="structureEd.hasParents"><label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Sits under (optional)
<select :value="structureEd.parent" @change="structureEd.onParent" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Top level</option><template v-for="o in structureEd.parentOpts" :key="o.id"><option :value="o.id">{{ o.label }}</option></template></select>
</label></template>
<div style="display: flex; align-items: flex-end"><button type="submit" :disabled="!structureEd.canAdd" :style="`height: 40px; padding: 0 22px; border-radius: 20px; border: 0; font-weight: 600; cursor: ${structureEd.canAdd ? 'pointer' : 'default'}; ${structureEd.canAdd ? 'background: #B8F36B; color: #0F1714' : 'background: var(--k-2A3832,#2A3832); color: var(--k-9DB0A6,#9DB0A6)'}`">Add</button></div>
<span style="grid-column: 1 / -1; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ structureEd.kindBlurb }}</span>
</form>
<template v-if="structureEd.hasOk"><div role="status" style="padding: 10px 14px; border-radius: 12px; background: var(--k-1E3326,#1E3326); color: #B8F36B; font-size: 14px">{{ structureEd.ok }}</div></template>
<template v-if="structureEd.hasErr"><div role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 14px">{{ structureEd.err }}</div></template>

<template v-for="lv in structureEd.levels" :key="lv.kind">
<section :aria-label="lv.plural" style="display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; align-items: baseline; gap: 8px"><strong style="font-size: 15px">{{ lv.plural }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ lv.count }}</span></div>
<template v-if="lv.picked"><div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 12px; background: var(--k-121B17,#121B17)">
<template v-if="lv.canMerge">
<template v-if="lv.asking">
<span style="flex: 1 1 260px; font-size: 13px; color: var(--k-FFB86B,#FFB86B); line-height: 1.5">{{ lv.mergeText }}</span>
<button type="button" @click="lv.merge" :disabled="structureEd.busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #FFB86B; color: #0F1714; font-weight: 600; cursor: pointer">Merge</button>
<button type="button" @click="lv.cancelMerge" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button>
</template>
<template v-else>
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Merge the {{ lv.picked }} ticked into
<select :value="lv.mergeInto" @change="lv.onInto" :aria-label="`${lv.label} to keep`" style="height: 34px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; max-width: 100%"><template v-for="o in lv.mergeOpts" :key="o.id"><option :value="o.id">{{ o.label }}</option></template></select></label>
<button type="button" @click="lv.askMerge" :disabled="structureEd.busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Merge…</button>
<button type="button" @click="lv.clearPick" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Clear</button>
</template>
</template>
<template v-else><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Tick at least two {{ lv.plural.toLowerCase() }} to merge them.</span><button type="button" @click="lv.clearPick" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Clear</button></template>
</div></template>
<template v-if="lv.none"><span style="font-size: 13px; color: var(--k-6F8178,#6F8178)">None yet. {{ lv.plural }} do not appear anywhere in Hub until you add one.</span></template>
<template v-for="r in lv.rows" :key="r.id">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C)">
<template v-if="r.editing">
<input type="text" :value="r.editName" @input="r.onEdit" @keydown.enter.prevent="r.saveEdit" maxlength="80" :aria-label="`Name of ${r.name}`" style="flex: 1 1 200px; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"/>
<button type="button" @click="r.saveEdit" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save</button>
<button type="button" @click="r.cancelEdit" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button>
</template>
<template v-else>
<input type="checkbox" :checked="r.picked" @change="r.onPick" :aria-label="`Select ${r.name} to merge`" style="width: 18px; height: 18px; accent-color: #B8F36B; cursor: pointer"/>
<div style="flex: 1 1 200px; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px">{{ r.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ r.under }} · {{ r.people }}</span></div>
<template v-if="r.confirming">
<span style="font-size: 13px; color: var(--k-FFB86B,#FFB86B)">Remove {{ r.name }}?</span>
<button type="button" @click="r.remove" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Remove</button>
<button type="button" @click="r.keep" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Keep</button>
</template>
<template v-else>
<template v-if="r.canMove"><label style="display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Under
<select :value="r.moveTo" @change="r.onMove" :aria-label="`Where ${r.name} sits`" style="height: 34px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; max-width: 100%"><option value="">Top level</option><template v-for="o in r.moveOpts" :key="o.id"><option :value="o.id">{{ o.label }}</option></template></select></label>
<template v-if="r.moved"><button type="button" @click="r.doMove" :disabled="structureEd.busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Move</button></template></template>
<button type="button" @click="r.startEdit" :aria-label="`Rename ${r.name}`" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Rename</button>
<button type="button" @click="r.askRemove" :aria-label="`Remove ${r.name}`" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer">Remove</button>
</template>
</template>
</div>
</template>
</section>
</template>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'StructureEditSection',
  ...hubBindings(["structureEd"])
};
</script>
