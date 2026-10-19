<template>
<div style="flex: 1; min-height: 0; overflow-y: auto">
<div style="max-width: 1100px; margin: 0 auto; padding: 18px clamp(16px, 3vw, 32px)">
<template v-if="selectedFolder">
<nav aria-label="Breadcrumb" style="display: flex; align-items: center; gap: 8px; margin-bottom: 16px; font-size: 14px">
<button type="button" @click="closeFolder" style="height: 34px; padding: 0 12px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Files</button>
<span style="color: #9DB0A6">/</span><strong style="overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ selectedFolder.name }}</strong>
</nav>
<div style="margin-bottom: 16px; display: flex; align-items: center; gap: 12px">
<svg aria-hidden="true" width="30" height="30" viewBox="0 0 24 24" :fill="color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg>
<div><h2 style="margin: 0; font-size: 20px">{{ selectedFolder.name }}</h2><span style="font-size: 12px; color: #9DB0A6">{{ selectedFolder.count }} {{ selectedFolder.count === 1 ? 'file' : 'files' }} · shared by {{ selectedFolder.by }} · {{ selectedFolder.time }}</span></div>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px">
<template v-for="(file, fileIndex) in selectedFolder.files" :key="fileIndex">
<article style="min-width: 0; border-radius: 14px; border: 1px solid #24322C; background: #16201C; overflow: hidden">
<div style="height: 190px; padding: 12px; box-sizing: border-box; background: #121B17; display: flex; align-items: center; justify-content: center">
<img v-if="file.previewKind === 'image' && file.url" :src="file.url" :alt="file.name" style="width: 100%; height: 100%; object-fit: contain; border-radius: 8px">
<div v-else-if="file.previewKind === 'image'" role="img" :aria-label="file.name" style="width: 100%; height: 100%; border-radius: 8px; background: #24405A"></div>
<video v-else-if="file.previewKind === 'video' && file.url" :src="file.url" controls preload="metadata" playsinline :aria-label="file.name" style="width: 100%; max-height: 100%; border-radius: 8px; background: #000"></video>
<div v-else-if="file.previewKind === 'video'" style="color: #9DB0A6; font-size: 13px">Video preview unavailable</div>
<iframe v-else-if="file.previewKind === 'pdf' && file.url" :src="file.url" :title="file.name" style="width: 100%; height: 100%; border: 0; border-radius: 6px; background: #fff"></iframe>
<div v-else-if="file.previewKind === 'note'" :style="`width: 150px; height: 100%; max-height: 166px; border-radius: 4px; background: #FBF6E3; padding: 12px; box-sizing: border-box; display: flex; flex-direction: column; gap: 5px; overflow: hidden; box-shadow: 0 6px 14px rgba(0,0,0,0.35); opacity: ${file.live ? 1 : 0.45}`"><span style="font-size: 10px; font-weight: 600; line-height: 1.3; color: #2A2716; overflow: hidden; max-height: 2.7em">{{ file.name.replace(/\.note$/, '') }}</span><span style="height: 2px; width: 36%; background: #E0B93A"></span><span style="font-size: 7.5px; line-height: 1.5; color: #5A5434; white-space: pre-line; overflow: hidden; flex: 1">{{ file.live ? file.notePreview : 'This note is no longer shared.' }}</span></div>
<div v-else-if="file.previewKind === 'sheet'" style="width: 100%; max-width: 190px; border-radius: 5px; background: #EDF1EE; padding: 6px; box-sizing: border-box; display: grid; grid-template-columns: repeat(4, 1fr); gap: 2px">
<template v-for="cell in 24" :key="cell"><span :style="`height: 12px; background: ${cell <= 4 ? '#8FC64A' : cell % 4 === 1 ? '#D6DED8' : '#E3E9E5'}`"></span></template>
</div>
<div v-else-if="file.previewKind === 'slides'" style="width: 100%; max-width: 200px; aspect-ratio: 16 / 9; border-radius: 5px; background: #EDF1EE; padding: 10px; box-sizing: border-box; display: flex; flex-direction: column; gap: 8px">
<span :style="`height: 8px; width: 65%; border-radius: 2px; background: ${file.color}`"></span><div style="flex: 1; display: flex; gap: 8px"><span style="flex: 1; border-radius: 3px; background: #C3CCC6"></span><span style="flex: 1; display: flex; flex-direction: column; gap: 4px"><i style="height: 4px; background: #C3CCC6"></i><i style="height: 4px; width: 80%; background: #C3CCC6"></i><i style="height: 4px; width: 90%; background: #C3CCC6"></i></span></div>
</div>
<div v-else-if="file.previewKind === 'archive'" style="display: flex; flex-direction: column; align-items: center; gap: 8px; color: #C9B6FF">
<svg aria-hidden="true" width="62" height="62" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M11 3v2h2v2h-2v2h2v2h-2v3"></path></svg><span style="font-size: 12px; color: #9DB0A6">Compressed archive</span>
</div>
<div v-else style="width: 112px; height: 150px; border-radius: 4px; background: #EDF1EE; padding: 12px 10px; box-sizing: border-box; display: flex; flex-direction: column; gap: 7px">
<span :style="`height: 7px; width: 70%; border-radius: 2px; background: ${file.color}`"></span><span style="height: 4px; border-radius: 2px; background: #C3CCC6"></span><span style="height: 4px; width: 88%; border-radius: 2px; background: #C3CCC6"></span><span style="height: 4px; width: 95%; border-radius: 2px; background: #C3CCC6"></span><span style="height: 4px; width: 65%; border-radius: 2px; background: #C3CCC6"></span>
</div>
</div>
<div style="padding: 12px 14px; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: center; min-width: 0"><span :style="`flex: none; padding: 1px 6px; border-radius: 4px; background: ${file.color}; color: #0F1714; font-size: 10px; font-weight: 700`">{{ file.kind }}</span><strong :title="file.name" style="min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 14px">{{ file.name }}</strong></div>
<span style="font-size: 12px; color: #9DB0A6">{{ file.meta || 'Shared in this chat' }}</span>
<button v-if="file.isNote" type="button" :disabled="!file.live" @click="file.noteOpen" :style="`height: 32px; margin-top: 5px; border: 1px solid #2A3832; border-radius: 9px; background: transparent; color: ${file.live ? '#E8EFEA' : '#6C7B73'}; cursor: ${file.live ? 'pointer' : 'default'}; font-size: 13px`">{{ file.live ? 'Open note' : 'Disconnected' }}</button>
<a v-if="file.url" :href="file.url" :download="file.name" style="height: 32px; margin-top: 5px; border: 1px solid #2A3832; border-radius: 9px; color: #E8EFEA; text-decoration: none; display: flex; align-items: center; justify-content: center; font-size: 13px">Download</a>
</div>
</article>
</template>
</div>
</template>
<template v-else-if="folders.length">
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 14px">
<template v-for="(folder, folderIndex) in folders" :key="folder.id">
<button type="button" @click="openFolder(folder)" :aria-label="`Open ${folder.name} folder`" style="width: 100%; text-align: left; color: #E8EFEA; cursor: pointer; border-radius: 16px; border: 1px solid #24322C; background: #16201C; display: flex; flex-direction: column; overflow: hidden; padding: 0">
<div style="height: 110px; padding: 14px 14px 0; box-sizing: border-box; background: #121B17; display: flex; gap: 8px; align-items: flex-end; overflow: hidden">
<template v-for="(sample, sampleIndex) in folder.samples" :key="sampleIndex"><span :style="`flex: 1; min-width: 0; height: ${sample.height}px; border-radius: 8px 8px 0 0; background: #EDF1EE; padding: 8px; box-sizing: border-box; display: flex; flex-direction: column; gap: 4px; overflow: hidden`"><span :style="`align-self: flex-start; padding: 1px 5px; border-radius: 4px; background: ${sample.color}; color: #0F1714; font-size: 9px; font-weight: 700`">{{ sample.kind }}</span><span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 90%"></span><span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 70%"></span><span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 82%"></span></span></template>
</div>
<div style="padding: 14px; display: flex; gap: 12px; align-items: center">
<svg aria-hidden="true" width="26" height="26" viewBox="0 0 24 24" :fill="color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong :title="folder.name" style="font-size: 15px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ folder.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ folder.count }} {{ folder.count === 1 ? 'file' : 'files' }} · shared by {{ folder.by }} · {{ folder.time }}</span></div>
</div>
</button>
</template>
</div>
</template>
<div v-else style="padding: 28px; border: 1px solid #24322C; border-radius: 16px; background: #16201C; text-align: center; color: #9DB0A6">No files have been shared in this chat yet.</div>
</div>
</div>
</template>

<script>
export default {
  name: 'ChatSharedFiles',
  props: {
    folders: { type: Array, default: () => [] },
    selectedFolder: { type: Object, default: null },
    color: { type: String, default: '#B8F36B' },
    openFolder: { type: Function, required: true },
    closeFolder: { type: Function, required: true }
  }
};
</script>
