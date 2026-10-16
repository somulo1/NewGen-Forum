<!-- An attached file: image preview (opens full size) or a document chip, each with Open and Download. -->
<template>
<div :style="`display: flex; flex-direction: column; gap: 8px; padding: 8px 10px; border-radius: 10px; background: #121B17; border: 1px solid #2A3832; max-width: ${isImg && url ? 320 : 360}px; align-self: flex-start`">
<a v-if="isImg && url" :href="url" target="_blank" rel="noopener" :aria-label="`Open ${name}`" style="display: block"><img :src="url" :alt="name" style="display: block; width: 100%; max-height: 240px; object-fit: contain; border-radius: 8px; background: #0F1714"></a>
<div style="display: flex; gap: 10px; align-items: center">
<span style="width: 32px; height: 32px; flex: none; border-radius: 8px; background: #22302A; color: #B8F36B; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg></span>
<div style="min-width: 0; flex: 1; display: flex; flex-direction: column"><strong style="font-size: 13px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ name }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ meta }}</span></div>
<template v-if="url">
<a :href="url" target="_blank" rel="noopener" style="font-size: 12px; color: #B8F36B; text-decoration: none; padding: 6px 4px">Open</a>
<a :href="url" :download="name" style="font-size: 12px; color: #B8F36B; text-decoration: none; padding: 6px 4px">Download</a>
</template>
</div>
</div>
</template>

<script>
export default { name: 'FileChip', props: { name: String, meta: String, url: String, isImg: Boolean } };
</script>
