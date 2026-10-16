<!-- One attached file, kept thin: an image thumbnail (opens full size) or a document pill. Download is always one tap. -->
<template>
<div v-if="isImg && url" :style="`position: relative; flex: none; width: ${size}px; height: ${Math.round(size * 0.75)}px; border-radius: 10px; overflow: hidden; border: 1px solid #2A3832; background: #0F1714`">
<a :href="url" target="_blank" rel="noopener" :title="name" :aria-label="`Open ${name}`" style="display: block; width: 100%; height: 100%"><img :src="url" :alt="name" style="display: block; width: 100%; height: 100%; object-fit: cover"></a>
<a :href="url" :download="name" :aria-label="`Download ${name}`" title="Download" style="position: absolute; right: 4px; bottom: 4px; width: 24px; height: 24px; border-radius: 12px; background: rgba(15,23,20,0.78); color: #E8EFEA; display: flex; align-items: center; justify-content: center; text-decoration: none; font-size: 13px">↓</a>
</div>
<div v-else :style="'display: inline-flex; align-items: center; gap: 8px; max-width: 100%; min-width: 0; height: 36px; padding: 0 6px 0 10px; box-sizing: border-box; border-radius: 18px; background: #121B17; border: 1px solid #2A3832'">
<svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#B8F36B" stroke-width="1.8" stroke-linejoin="round" style="flex: none"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg>
<a v-if="url" :href="url" target="_blank" rel="noopener" :title="`Open ${name}`" style="min-width: 0; font-size: 13px; color: #E8EFEA; text-decoration: none; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ name }}</a>
<span v-else style="min-width: 0; font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ name }}</span>
<span style="flex: none; font-size: 11px; color: #9DB0A6">{{ meta }}</span>
<a v-if="url" :href="url" :download="name" :aria-label="`Download ${name}`" title="Download" style="flex: none; width: 26px; height: 26px; border-radius: 13px; display: flex; align-items: center; justify-content: center; color: #B8F36B; text-decoration: none">↓</a>
</div>
</template>

<script>
export default { name: 'FileChip', props: { name: String, meta: String, url: String, isImg: Boolean, size: { type: Number, default: 132 } } };
</script>
