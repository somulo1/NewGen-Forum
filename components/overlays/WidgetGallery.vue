<!-- Add-widget gallery -->
<template><template v-if="galleryOpen">
<div style="position: fixed; inset: 0; z-index: 60; background: rgba(5,10,8,0.72); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box">
<div role="dialog" aria-label="Add a widget" style="width: 100%; max-width: 820px; max-height: 86vh; display: flex; flex-direction: column; border-radius: 22px; background: #16201C; border: 1px solid #2A3832; box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 18px 20px; border-bottom: 1px solid #24322C">
<div style="flex: 1"><h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">Add a widget</h2><span style="font-size: 13px; color: #9DB0A6">Pick what your home page should show. You can remove or reorder them anytime.</span></div>
<button aria-label="Close" @click="closeGallery" style="width: 40px; height: 40px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px 22px; display: grid; grid-template-columns: repeat(auto-fill, minmax(230px, 1fr)); gap: 12px">
<template v-for="(g, g__i) in gallery" :key="g__i">
<div style="padding: 14px; border-radius: 14px; background: #121B17; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; align-items: center; gap: 8px"><span :style="`width: 10px; height: 10px; border-radius: 50%; background: ${g.tint}`"></span><strong style="font-size: 14px">{{ g.title }}</strong></div>
<span style="flex: 1; font-size: 13px; line-height: 1.45; color: #C9D6CF">{{ g.desc }}</span>
<div style="display: flex; gap: 8px"><button @click="g.add" :style="g.btnStyle">{{ g.btnLabel }}</button><button :aria-label="`Settings for ${g.title}`" title="Settings" @click="g.openSettings" style="flex: none; width: 38px; height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="3.5"></circle><path d="M12 2.5v3M12 18.5v3M2.5 12h3M18.5 12h3M5.3 5.3l2.1 2.1M16.6 16.6l2.1 2.1M5.3 18.7l2.1-2.1M16.6 7.4l2.1-2.1"></path></svg></button></div>
</div>
</template>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WidgetGallery',
  ...hubBindings(["closeGallery","gallery","galleryOpen"])
};
</script>
