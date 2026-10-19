<!-- Notes: notebooks (like folders) holding notes. Opened from the sidebar or the My notes widget. -->
<template><template v-if="show.notes">
<div style="width: 100%; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<template v-if="!notesPage.editor.open && !notesPage.inBook">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 260px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Notes</h1><p style="margin: 4px 0 0; color: #9DB0A6">Your private notes, kept in notebooks. {{ notesPage.total }} in all.</p></div>
<button @click="notesPage.newNote" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #FFD66B; color: #0F1714; font-weight: 600; cursor: pointer">New note</button>
<button @click="notesPage.newBook.toggle" :aria-expanded="notesPage.newBook.open" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">New notebook</button>
</div>
<form v-if="notesPage.newBook.open" novalidate @submit.prevent="notesPage.newBook.create" style="display: flex; flex-wrap: wrap; gap: 8px; padding: 12px; border-radius: 14px; background: #16201C; border: 1px solid #24322C"><input type="text" :value="notesPage.newBook.name" @input="notesPage.newBook.onName" placeholder="Notebook name" aria-label="Notebook name" style="flex: 1 1 200px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><button type="submit" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create</button></form>
<input type="search" :value="notesPage.q" @input="notesPage.onQ" placeholder="Search all notes" aria-label="Search all notes" style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px">
<template v-if="notesPage.searching">
<span v-if="!notesPage.matches.length" style="color: #9DB0A6">Nothing matches "{{ notesPage.q }}".</span>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 220px), 1fr)); gap: 14px"><template v-for="n in notesPage.matches" :key="n.id"><div style="display: flex; flex-direction: column; border-radius: 16px; border: 1px solid #24322C; background: #16201C; overflow: hidden; min-width: 0">
<button @click="n.open" :aria-label="`Open note ${n.title}`" style="border: 0; padding: 16px 0; background: #121B17; cursor: pointer; display: flex; align-items: center; justify-content: center; min-height: 190px">
<div style="width: 150px; height: 170px; border-radius: 4px; background: #FBF6E3; padding: 12px 12px; box-sizing: border-box; display: flex; flex-direction: column; gap: 5px; overflow: hidden; box-shadow: 0 6px 14px rgba(0,0,0,0.35); text-align: left"><span style="font-size: 10px; font-weight: 600; line-height: 1.3; color: #2A2716; overflow: hidden; max-height: 2.7em">{{ n.title }}</span><span style="height: 2px; width: 36%; background: #E0B93A"></span><span style="font-size: 7.5px; line-height: 1.5; color: #5A5434; white-space: pre-line; overflow: hidden; flex: 1">{{ n.body }}</span></div>
</button>
<div style="padding: 12px 14px; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: center"><span :style="`flex: none; padding: 1px 6px; border-radius: 4px; background: ${n.color}; color: #0F1714; font-size: 10px; font-weight: 700`">{{ n.kindTag }}</span><strong style="font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0">{{ n.title }}</strong><span v-if="n.pinned" title="Pinned" style="margin-left: auto; font-size: 11px; color: #FFD66B">Pinned</span></div>
<span style="font-size: 12px; color: #9DB0A6">{{ n.meta }}</span>
<div style="display: flex; gap: 6px; margin-top: 6px"><button @click="n.open" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Open</button><button @click="n.togglePin" :aria-pressed="n.pinned" :aria-label="n.pinned ? 'Unpin note' : 'Pin note'" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ n.pinned ? 'Unpin' : 'Pin' }}</button></div>
</div></div></template></div>
</template>
<div v-else style="display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 14px">
<template v-for="bk in notesPage.books" :key="bk.id">
<button @click="bk.open" :aria-label="`Open notebook ${bk.name}`" style="text-align: left; padding: 0; border-radius: 16px; border: 1px solid #24322C; background: #16201C; color: #E8EFEA; cursor: pointer; display: flex; flex-direction: column; overflow: hidden; font-family: inherit">
<div style="height: 110px; padding: 14px 14px 0; box-sizing: border-box; background: #121B17; display: flex; gap: 8px; align-items: flex-end; overflow: hidden">
<template v-for="(sm, i) in bk.samples" :key="i">
<span :style="`flex: 1; min-width: 0; height: ${sm.h}; border-radius: 8px 8px 0 0; background: #FBF6E3; padding: 8px; box-sizing: border-box; display: flex; flex-direction: column; gap: 4px; overflow: hidden`">
<span :style="`align-self: flex-start; padding: 1px 5px; border-radius: 4px; background: ${sm.color}; color: #0F1714; font-size: 9px; font-weight: 700`">{{ sm.kind }}</span>
<span style="font-size: 8px; line-height: 1.3; color: #2A2716; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ sm.title }}</span>
<span style="height: 3px; border-radius: 2px; background: #E4DAB4; width: 90%"></span>
<span style="height: 3px; border-radius: 2px; background: #E4DAB4; width: 70%"></span>
</span>
</template>
<span v-if="!bk.samples.length" style="flex: 1; height: 72px; border-radius: 8px 8px 0 0; border: 1px dashed #3A4A42; border-bottom: 0; display: flex; align-items: center; justify-content: center; font-size: 12px; color: #9DB0A6">Empty notebook</span>
</div>
<div style="padding: 14px; display: flex; gap: 12px; align-items: center">
<svg aria-hidden="true" width="26" height="26" viewBox="0 0 24 24" :fill="bk.color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ bk.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ bk.count }} {{ bk.count === 1 ? 'note' : 'notes' }} · updated {{ bk.updated }}</span></div>
</div>
</button>
</template>
</div>
</template>
<template v-else-if="!notesPage.editor.open">
<nav aria-label="Breadcrumb" style="display: flex; align-items: center; gap: 8px; font-size: 14px"><button @click="notesPage.back" style="border: 0; background: transparent; color: #B8F36B; cursor: pointer; padding: 0; font-size: 14px">Notes</button><span style="color: #9DB0A6">/</span><span>{{ notesPage.book.name }}</span></nav>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px"><svg aria-hidden="true" width="32" height="32" viewBox="0 0 24 24" :fill="notesPage.book.color" stroke="none"><path d="M3 6.5A1.5 1.5 0 0 1 4.5 5h5l2 2.5h8A1.5 1.5 0 0 1 21 9v9.5a1.5 1.5 0 0 1-1.5 1.5h-15A1.5 1.5 0 0 1 3 18.5z"></path></svg><div style="flex: 1 1 200px; min-width: 0"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px">{{ notesPage.book.name }}</h1><span style="color: #9DB0A6; font-size: 13px">{{ notesPage.book.count }} {{ notesPage.book.count === 1 ? 'note' : 'notes' }}</span></div>
<input type="search" :value="notesPage.q" @input="notesPage.onQ" placeholder="Search this notebook" aria-label="Search this notebook" style="height: 40px; padding: 0 14px; border-radius: 20px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px; flex: 0 1 240px; min-width: 0">
<button @click="notesPage.newNote" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #FFD66B; color: #0F1714; font-weight: 600; cursor: pointer">New note</button></div>
<span v-if="!notesPage.hasNotes" style="color: #9DB0A6">No notes here yet.</span>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(min(100%, 220px), 1fr)); gap: 14px"><template v-for="n in notesPage.notes" :key="n.id"><div style="display: flex; flex-direction: column; border-radius: 16px; border: 1px solid #24322C; background: #16201C; overflow: hidden; min-width: 0">
<button @click="n.open" :aria-label="`Open note ${n.title}`" style="border: 0; padding: 16px 0; background: #121B17; cursor: pointer; display: flex; align-items: center; justify-content: center; min-height: 190px">
<div style="width: 150px; height: 170px; border-radius: 4px; background: #FBF6E3; padding: 12px 12px; box-sizing: border-box; display: flex; flex-direction: column; gap: 5px; overflow: hidden; box-shadow: 0 6px 14px rgba(0,0,0,0.35); text-align: left"><span style="font-size: 10px; font-weight: 600; line-height: 1.3; color: #2A2716; overflow: hidden; max-height: 2.7em">{{ n.title }}</span><span style="height: 2px; width: 36%; background: #E0B93A"></span><span style="font-size: 7.5px; line-height: 1.5; color: #5A5434; white-space: pre-line; overflow: hidden; flex: 1">{{ n.body }}</span></div>
</button>
<div style="padding: 12px 14px; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: center"><span :style="`flex: none; padding: 1px 6px; border-radius: 4px; background: ${n.color}; color: #0F1714; font-size: 10px; font-weight: 700`">{{ n.kindTag }}</span><strong style="font-size: 14px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0">{{ n.title }}</strong><span v-if="n.pinned" title="Pinned" style="margin-left: auto; font-size: 11px; color: #FFD66B">Pinned</span></div>
<span style="font-size: 12px; color: #9DB0A6">{{ n.meta }}</span>
<div style="display: flex; gap: 6px; margin-top: 6px"><button @click="n.open" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">Open</button><button @click="n.togglePin" :aria-pressed="n.pinned" :aria-label="n.pinned ? 'Unpin note' : 'Pin note'" style="flex: 1; height: 34px; border-radius: 9px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ n.pinned ? 'Unpin' : 'Pin' }}</button></div>
</div></div></template></div>
</template>
<div v-if="notesPage.hasFlash" role="status" style="position: fixed; left: 50%; transform: translateX(-50%); top: 76px; z-index: 80; padding: 10px 16px; border-radius: 12px; background: #E8EFEA; color: #0F1714; font-size: 14px">{{ notesPage.flash }}</div>

<div v-if="notesPage.editor.open" style="display: flex; flex-direction: column; gap: 14px; width: 100%; min-width: 0">
<nav aria-label="Breadcrumb" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px; font-size: 14px"><button @click="notesPage.editor.close" style="border: 0; background: transparent; color: #B8F36B; cursor: pointer; padding: 0; font-size: 14px">Notes</button><span style="color: #9DB0A6">/</span><span :style="`color: ${notesPage.editor.bookColor}`">{{ notesPage.editor.bookName }}</span><span style="color: #9DB0A6">/</span><strong style="font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 40ch">{{ notesPage.editor.title || 'Untitled' }}</strong></nav>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px 12px; padding: 10px 14px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<span :style="`padding: 1px 6px; border-radius: 4px; background: ${notesPage.editor.bookColor}; color: #0F1714; font-size: 10px; font-weight: 700`">NOTE</span>
<span style="font-size: 12px; color: #9DB0A6">{{ notesPage.editor.edited }} · {{ notesPage.editor.words }}</span>
<span :style="`font-size: 12px; color: ${notesPage.editor.dirty ? '#FFB86B' : '#9DB0A6'}`">{{ notesPage.editor.dirty ? 'Unsaved changes' : 'All changes saved' }}</span>
<label style="display: flex; align-items: center; gap: 6px; font-size: 12px; color: #9DB0A6; margin-left: auto">Notebook<select :value="notesPage.editor.book" @change="notesPage.editor.onBook" style="height: 34px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"><template v-for="o in notesPage.editor.books" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select></label>
<button v-if="notesPage.editor.canPin" type="button" @click="notesPage.editor.togglePin" :aria-pressed="notesPage.editor.pinned" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ notesPage.editor.pinned ? 'Unpin' : 'Pin' }}</button>
<button v-if="notesPage.editor.remove" type="button" @click="notesPage.editor.remove" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #FF9B8A; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 13px">Delete</button>
<button type="button" :disabled="!notesPage.editor.dirty" @click="notesPage.editor.save" :style="`height: 34px; padding: 0 18px; border-radius: 17px; border: 0; font-weight: 600; font-size: 13px; cursor: ${notesPage.editor.dirty ? 'pointer' : 'default'}; background: ${notesPage.editor.dirty ? '#FFD66B' : '#3A3620'}; color: ${notesPage.editor.dirty ? '#0F1714' : '#9A9066'}`">{{ notesPage.editor.isNew ? 'Create note' : 'Save changes' }}</button>
</div>
<span v-if="notesPage.editor.hasErr" role="alert" style="font-size: 13px; color: #FF9B8A">{{ notesPage.editor.err }}</span>
<div class="note-paper" style="width: 100%; max-width: 960px; margin: 0 auto; box-sizing: border-box; padding: clamp(20px, 4vw, 44px) clamp(18px, 5vw, 56px) 56px; border-radius: 6px; background: #FBF6E3; box-shadow: 0 10px 30px rgba(0,0,0,0.4); min-height: 440px; display: flex; flex-direction: column; gap: 10px">
<input type="text" :value="notesPage.editor.title" @input="notesPage.editor.onTitle" placeholder="Title" aria-label="Note title" style="border: 0; border-bottom: 2px solid #E0B93A; padding: 4px 0 10px; background: transparent; color: #2A2716; font-family: 'Space Grotesk', sans-serif; font-size: clamp(22px, 3.4vw, 30px); font-weight: 600; outline: none; width: 100%; box-sizing: border-box">
<NoteEditor :modelValue="notesPage.editor.body" @update:modelValue="notesPage.editor.onBody"/>
</div>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';
export default { name: 'NotesSection', ...hubBindings(['notesPage', 'show']) };
</script>
