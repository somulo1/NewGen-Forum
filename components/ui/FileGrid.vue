<!-- All files of one message: thumbnails and pills that wrap, so many files stay tidy in any chat width. -->
<template>
<div style="display: flex; flex-wrap: wrap; gap: 6px; margin-top: 6px; min-width: 0; max-width: 100%">
<template v-for="(f, i) in files" :key="i"><FileChip :name="f.name" :meta="f.meta" :url="f.url" :is-img="!!f.isImg" :size="size"/></template>
</div>
</template>

<script>
export default { name: 'FileGrid', props: { files: { type: Array, default: () => [] }, size: { type: Number, default: 132 } } };
</script>
