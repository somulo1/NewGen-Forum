<!-- Files and folder view -->
<template><template v-if="showFolderList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Files</h1><p style="margin: 4px 0 0; color: #9DB0A6">Folders shared across the organization. Open one to preview what's inside.</p></div>
<button style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">New folder</button>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 14px">
<template v-for="(fo, fo__i) in folders" :key="fo__i">
<button @click="fo.open" style="text-align: left; padding: 0; border-radius: 16px; border: 1px solid #24322C; background: #16201C; color: #E8EFEA; cursor: pointer; display: flex; flex-direction: column; overflow: hidden">
<div style="height: 110px; padding: 14px 14px 0; box-sizing: border-box; background: #121B17; display: flex; gap: 8px; align-items: flex-end; overflow: hidden">
<template v-for="(sm, sm__i) in fo.samples" :key="sm__i">
<span :style="`flex: 1; min-width: 0; height: ${sm.h}; border-radius: 8px 8px 0 0; background: #EDF1EE; padding: 8px; box-sizing: border-box; display: flex; flex-direction: column; gap: 4px; overflow: hidden`">
<span :style="`align-self: flex-start; padding: 1px 5px; border-radius: 4px; background: ${sm.color}; color: #0F1714; font-size: 9px; font-weight: 700`">{{ sm.kind }}</span>
<span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 90%"></span>
<span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 70%"></span>
<span style="height: 3px; border-radius: 2px; background: #C3CCC6; width: 82%"></span>
</span>
</template>
</div>
<div style="padding: 14px; display: flex; gap: 12px; align-items: center">
<svg aria-hidden="true" width="26" height="26" viewBox="0 0 24 24" :fill="fo.color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ fo.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ fo.count }} files · updated {{ fo.updated }}</span></div>
</div>
</button>
</template>
</div>
</div>
</template>

<template v-if="showFolderView">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<nav aria-label="Breadcrumb" style="display: flex; align-items: center; gap: 8px; font-size: 14px"><button @click="closeFolder" style="height: 34px; padding: 0 12px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Files</button><span style="color: #9DB0A6">/</span><strong>{{ fv.name }}</strong></nav>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
<svg aria-hidden="true" width="34" height="34" viewBox="0 0 24 24" :fill="fv.color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg>
<div style="flex: 1 1 260px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ fv.name }}</h1><span style="font-size: 13px; color: #9DB0A6">{{ fv.count }} files · shared with {{ fv.audience }}</span></div>
<label v-if="fv.isDocs" style="position: relative; height: 40px; padding: 0 18px; border-radius: 20px; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 8px"><input type="file" multiple="" aria-label="Upload to this folder" @change="uploadToFolder" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 16V4M7 9l5-5 5 5M4 20h16"></path></svg>Upload here</label>
</div>
<template v-if="fv.isNotes">
<p v-if="fv.sn.disconnected" role="status" style="margin: 0; padding: 12px 14px; border-radius: 12px; background: #2A2418; border: 1px solid #5A4A1E; color: #FFD66B; font-size: 13px">The Notes widget is not on your home page, so notes you shared are disconnected. Add the My notes widget again to reconnect them.</p>
<section aria-label="Shared with you" style="display: flex; flex-direction: column; gap: 12px">
<h2 style="margin: 0; font-size: 16px">Shared with you</h2>
<span v-if="!fv.sn.hasIn" style="color: #9DB0A6; font-size: 14px">No one has shared a note with you yet.</span>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 220px), 1fr)); gap: 14px">
<template v-for="n in fv.sn.in" :key="n.id"><div style="display: flex; flex-direction: column; border-radius: 16px; border: 1px solid #24322C; background: #16201C; overflow: hidden; min-width: 0">
<button @click="n.open" :aria-label="`Open note ${n.title}`" style="border: 0; padding: 16px 0; background: #121B17; cursor: pointer; display: flex; align-items: center; justify-content: center; min-height: 190px"><div style="width: 150px; height: 170px; border-radius: 4px; background: #FBF6E3; padding: 12px; box-sizing: border-box; display: flex; flex-direction: column; gap: 5px; overflow: hidden; box-shadow: 0 6px 14px rgba(0,0,0,0.35); text-align: left"><span style="font-size: 10px; font-weight: 600; line-height: 1.3; color: #2A2716; overflow: hidden; max-height: 2.7em">{{ n.title }}</span><span style="height: 2px; width: 36%; background: #E0B93A"></span><span style="font-size: 7.5px; line-height: 1.5; color: #5A5434; white-space: pre-line; overflow: hidden; flex: 1">{{ n.preview }}</span></div></button>
<div style="padding: 12px 14px; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: center"><span style="flex: none; padding: 1px 6px; border-radius: 4px; background: #FFD66B; color: #0F1714; font-size: 10px; font-weight: 700">NOTE</span><strong style="font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0">{{ n.title }}</strong></div>
<span style="font-size: 12px; color: #9DB0A6">From {{ n.by }} · {{ n.when }}</span>
<span :style="`align-self: flex-start; padding: 1px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; background: ${n.canEdit ? '#2C3D34' : '#2A2F3A'}; color: ${n.canEdit ? '#B8F36B' : '#9DB0C6'}`">{{ n.accessLabel }}</span>
<div style="display: flex; gap: 6px; margin-top: 6px"><button @click="n.open" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ n.canEdit ? 'Open & edit' : 'Open' }}</button><button @click="n.remove" aria-label="Remove from my files" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 13px">Remove</button></div>
</div></div></template>
</div>
</section>
<section v-if="fv.sn.hasOut" aria-label="Shared by you" style="display: flex; flex-direction: column; gap: 12px">
<h2 style="margin: 0; font-size: 16px">Shared by you</h2>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 260px), 1fr)); gap: 10px">
<template v-for="n in fv.sn.out" :key="n.shareId"><div style="display: flex; flex-direction: column; gap: 6px; padding: 12px 14px; border-radius: 14px; border: 1px solid #24322C; background: #16201C">
<strong style="font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ n.title }}</strong>
<span style="font-size: 12px; color: #9DB0A6">To {{ n.to }} · {{ n.accessLabel }}</span>
<span :style="`font-size: 12px; color: ${n.live ? '#B8F36B' : '#FFB86B'}`">{{ n.live ? 'Connected' : 'Disconnected' }}</span>
<div style="display: flex; gap: 6px"><button @click="n.open" style="flex: 1; height: 32px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Open</button><button @click="n.stop" style="flex: 1; height: 32px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 13px">Stop sharing</button></div>
</div></template>
</div>
</section>
</template>
<div v-if="fv.isDocs" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(220px, 1fr)); gap: 14px">
<template v-for="(fl, fl__i) in fv.files" :key="fl__i">
<div style="border-radius: 16px; border: 1px solid #24322C; background: #16201C; display: flex; flex-direction: column; overflow: hidden">
<div style="height: 168px; background: #121B17; display: flex; align-items: center; justify-content: center; padding: 14px; box-sizing: border-box">
<template v-if="fl.isDoc">
<div style="width: 112px; height: 140px; border-radius: 4px; background: #EDF1EE; padding: 12px 10px; box-sizing: border-box; display: flex; flex-direction: column; gap: 6px; box-shadow: 0 6px 14px rgba(0,0,0,0.35)">
<span :style="`height: 7px; width: 70%; border-radius: 2px; background: ${fl.color}`"></span>
<span style="height: 4px; width: 100%; border-radius: 2px; background: #C3CCC6"></span>
<span style="height: 4px; width: 92%; border-radius: 2px; background: #C3CCC6"></span>
<span style="height: 4px; width: 96%; border-radius: 2px; background: #C3CCC6"></span>
<span style="height: 4px; width: 64%; border-radius: 2px; background: #C3CCC6"></span>
<span style="height: 4px; width: 88%; border-radius: 2px; background: #C3CCC6; margin-top: 4px"></span>
<span style="height: 4px; width: 94%; border-radius: 2px; background: #C3CCC6"></span>
<span style="height: 4px; width: 58%; border-radius: 2px; background: #C3CCC6"></span>
</div>
</template>
<template v-if="fl.isSheet">
<div style="width: 160px; border-radius: 4px; background: #EDF1EE; padding: 6px; box-sizing: border-box; display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 2px; box-shadow: 0 6px 14px rgba(0,0,0,0.35)">
<span style="height: 12px; background: #8FC64A"></span><span style="height: 12px; background: #8FC64A"></span><span style="height: 12px; background: #8FC64A"></span><span style="height: 12px; background: #8FC64A"></span>
<span style="height: 12px; background: #D6DED8"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span>
<span style="height: 12px; background: #D6DED8"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span>
<span style="height: 12px; background: #D6DED8"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span>
<span style="height: 12px; background: #D6DED8"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span>
<span style="height: 12px; background: #D6DED8"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span><span style="height: 12px; background: #E3E9E5"></span>
</div>
</template>
<template v-if="fl.isSlide">
<div style="width: 176px; aspect-ratio: 16 / 9; border-radius: 4px; background: #EDF1EE; padding: 10px; box-sizing: border-box; display: flex; flex-direction: column; gap: 8px; box-shadow: 0 6px 14px rgba(0,0,0,0.35)">
<span :style="`height: 9px; width: 60%; border-radius: 2px; background: ${fl.color}`"></span>
<div style="flex: 1; display: flex; gap: 8px"><span style="flex: 1; border-radius: 3px; background: #C3CCC6"></span><span style="flex: 1; display: flex; flex-direction: column; gap: 4px"><span style="height: 4px; background: #C3CCC6; border-radius: 2px"></span><span style="height: 4px; width: 80%; background: #C3CCC6; border-radius: 2px"></span><span style="height: 4px; width: 90%; background: #C3CCC6; border-radius: 2px"></span></span></div>
</div>
</template>
<template v-if="fl.isImg">
<template v-if="fl.hasUrl"><img :src="fl.url" :alt="fl.name" style="width: 100%; height: 140px; object-fit: cover; border-radius: 8px"></template>
<template v-if="fl.noUrl"><div role="img" :aria-label="fl.name" :style="`width: 100%; height: 140px; border-radius: 8px; background: ${fl.imgBg}; display: flex; align-items: flex-end; overflow: hidden`"><svg aria-hidden="true" width="100%" height="70%" viewBox="0 0 200 80" preserveAspectRatio="none"><circle cx="150" cy="22" r="10" fill="rgba(255,255,255,0.55)"></circle><path d="M0 80 L55 30 L95 62 L130 40 L200 80 Z" fill="rgba(255,255,255,0.25)"></path></svg></div></template>
</template>
<template v-if="fl.isZip">
<div style="display: flex; flex-direction: column; align-items: center; gap: 8px; color: #C9B6FF"><svg aria-hidden="true" width="56" height="56" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M11 3v2h2v2h-2v2h2v2h-2v3"></path></svg><span style="font-size: 12px; color: #9DB0A6">Compressed archive</span></div>
</template>
</div>
<div style="padding: 12px 14px; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: center"><span :style="`flex: none; padding: 1px 6px; border-radius: 4px; background: ${fl.color}; color: #0F1714; font-size: 10px; font-weight: 700`">{{ fl.kind }}</span><strong style="font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ fl.name }}</strong></div>
<span style="font-size: 12px; color: #9DB0A6">{{ fl.size }} · {{ fl.by }} · {{ fl.when }}</span>
<div style="display: flex; gap: 6px; margin-top: 6px"><button style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Open</button><button style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Share</button></div>
</div>
</div>
</template>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'FilesSection',
  ...hubBindings(["closeFolder","folders","fv","showFolderList","showFolderView","uploadToFolder"])
};
</script>
