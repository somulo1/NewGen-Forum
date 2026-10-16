<!-- Widgets: add, remove and set up the widgets on Home -->
<template><template v-if="show.widgets">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Widgets</h1><p style="margin: 4px 0 0; color: #9DB0A6">Pick what your home page should show. You can remove or reorder them anytime.</p></div>
<button @click="closeGallery" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Back to Home</button>
</div>
<label style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #9DB0A6; max-width: 420px">Search widgets<input type="search" :value="wgq" @input="onWgq" placeholder="Search by name or what it does…" style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<div role="tablist" style="align-self: flex-start; display: flex; padding: 3px; border-radius: 12px; background: #16201C; border: 1px solid #24322C">
<button role="tab" @click="wgGo.all" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; ${wgTab.all ? 'background: #E8EFEA; color: #0F1714' : 'background: transparent; color: #C9D6CF'}`">All ({{ wgCounts.all }})</button>
<button role="tab" @click="wgGo.added" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; ${wgTab.added ? 'background: #E8EFEA; color: #0F1714' : 'background: transparent; color: #C9D6CF'}`">On Home ({{ wgCounts.added }})</button>
<button role="tab" @click="wgGo.available" :style="`height: 34px; padding: 0 14px; border-radius: 9px; border: 0; cursor: pointer; font-size: 13px; font-weight: 600; ${wgTab.available ? 'background: #E8EFEA; color: #0F1714' : 'background: transparent; color: #C9D6CF'}`">Available ({{ wgCounts.available }})</button>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(260px, 1fr)); gap: 14px">
<template v-for="(g, g__i) in wgItems" :key="g.title">
<article style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; align-items: center; gap: 8px"><span :style="`width: 10px; height: 10px; border-radius: 50%; background: ${g.tint}`"></span><strong style="font-size: 15px">{{ g.title }}</strong></div>
<span style="flex: 1; font-size: 13px; line-height: 1.45; color: #C9D6CF">{{ g.desc }}</span>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-if="!g.added"><button :disabled="!canManageWidgets" @click="g.add" :style="g.btnStyle">{{ canManageWidgets ? 'Add to home' : 'No access to add' }}</button></template>
<template v-if="g.added"><button @click="g.toggleHide" :aria-label="`${g.hideLabel}: ${g.title}`" style="flex: 1; height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">{{ g.hideLabel }}</button><template v-if="canManageWidgets"><button @click="g.remove" :aria-label="`Remove ${g.title}`" style="height: 38px; padding: 0 14px; border-radius: 10px; border: 1px solid #5A3A36; background: transparent; color: #FF9B8A; font-weight: 600; cursor: pointer">Remove</button></template></template>
<button :aria-label="`Settings for ${g.title}`" title="Settings" @click="g.openSettings" style="flex: none; width: 38px; height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="12" cy="12" r="3"></circle><path d="M12 2v3M12 19v3M2 12h3M19 12h3M5 5l2 2M17 17l2 2M19 5l-2 2M7 17l-2 2"></path></svg></button>
</div>
</article>
</template>
</div>
<template v-if="wgItems.length === 0"><div style="padding: 40px 20px; border-radius: 16px; border: 1px dashed #3F524A; text-align: center; color: #9DB0A6">Nothing here.</div></template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WidgetsSection',
  ...hubBindings(["canManageWidgets","closeGallery","onWgq","show","wgCounts","wgGo","wgItems","wgTab","wgq"])
};
</script>
