<!-- User groups: sets of roles that are given rights together. Everyone whose role is in a group gets what the group is allowed. -->
<template>
<section aria-label="User groups" style="display: flex; flex-direction: column; gap: 14px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); line-height: 1.5">A user group puts several roles together. Tick what the group may do and every role in it can do that, on top of what the role itself allows. Groups also appear next to roles in each widget's settings. A person still has one role.</span>

<form novalidate aria-label="New user group" @submit.prevent="create" style="display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 12px; padding: 14px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Group name<input v-model="form.name" maxlength="60" placeholder="e.g. Team HQ" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">What it is for (optional)<input v-model="form.desc" maxlength="300" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="grid-column: 1 / -1; display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Roles in this group</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><button v-for="r in roles" :key="r.id" type="button" :aria-pressed="form.roles.includes(r.id)" @click="pick(form.roles, r.id)" :style="chip(form.roles.includes(r.id))">{{ r.name }}</button><span v-if="!roles.length" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Make some roles first.</span></div></div>
<div style="grid-column: 1 / -1; display: flex; align-items: center; gap: 10px"><button type="submit" :disabled="!form.name.trim() || busy" :style="`height: 40px; padding: 0 22px; border-radius: 20px; border: 0; font-weight: 600; cursor: pointer; ${form.name.trim() && !busy ? 'background: #B8F36B; color: #0F1714' : 'background: var(--k-2A3832,#2A3832); color: var(--k-9DB0A6,#9DB0A6)'}`">Create group</button></div>
</form>

<div v-if="err" role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: var(--k-FFD3CA,#FFD3CA); font-size: 14px">{{ err }}</div>
<div v-if="ok" role="status" style="padding: 10px 14px; border-radius: 12px; background: var(--k-1E3326,#1E3326); color: #B8F36B; font-size: 14px">{{ ok }}</div>
<span v-if="!groups.length" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">No user groups yet.</span>

<div v-if="groups.length" style="border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); overflow: hidden"><div style="overflow-x: auto">
<table style="width: 100%; border-collapse: collapse; font-size: 14px">
<thead><tr :style="headRow"><th :style="th + '; padding-left: 14px'">Group</th><th :style="th">Roles</th><th :style="th + '; text-align: right'">People</th><th :style="th + '; text-align: right'">Rights</th><th :style="th + '; text-align: right; padding-right: 14px'">Actions</th></tr></thead>
<tbody>
<template v-for="g in groups" :key="g.id">
<tr style="border-top: 1px solid var(--k-24322C,#24322C)">
<td v-if="edit && edit.id === g.id" colspan="5" style="padding: 10px 14px"><div style="display: flex; flex-wrap: wrap; gap: 10px">
<input v-model="edit.name" maxlength="60" :aria-label="`Name of ${g.name}`" @keydown.enter.prevent="saveEdit(g)" style="flex: 1 1 180px; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)">
<input v-model="edit.desc" maxlength="300" placeholder="What it is for" :aria-label="`About ${g.name}`" @keydown.enter.prevent="saveEdit(g)" style="flex: 2 1 220px; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)">
<button type="button" @click="saveEdit(g)" :disabled="busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save</button>
<button type="button" @click="edit = null" :style="btn">Cancel</button></div></td>
<template v-else>
<td style="padding: 12px 8px 12px 14px; vertical-align: top; min-width: 180px"><div style="font-weight: 600">{{ g.name }}</div><div style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); margin-top: 2px">{{ g.desc || 'No description' }}</div></td>
<td style="padding: 12px 8px; vertical-align: top; min-width: 200px"><div style="display: flex; flex-wrap: wrap; gap: 4px"><span v-for="n in roleNames(g)" :key="n" style="padding: 2px 10px; border-radius: 12px; font-size: 12px; background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); white-space: nowrap">{{ n }}</span><span v-if="!g.roles.length" style="font-size: 12px; color: var(--k-6F8178,#6F8178)">No roles</span></div></td>
<td style="padding: 12px 8px; vertical-align: top; text-align: right; font-variant-numeric: tabular-nums">{{ g.count }}</td>
<td style="padding: 12px 8px; vertical-align: top; text-align: right; font-variant-numeric: tabular-nums; white-space: nowrap">{{ granted(g) }}</td>
<td style="padding: 8px 14px 8px 8px; vertical-align: top"><div style="display: flex; justify-content: flex-end; flex-wrap: wrap; gap: 6px">
<template v-if="confirming === g.id"><span style="align-self: center; font-size: 13px; color: var(--k-FFB86B,#FFB86B)">Delete {{ g.name }}? Its rights go with it.</span>
<button type="button" @click="remove(g)" :disabled="busy" :style="small + '; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600'">Delete</button>
<button type="button" @click="confirming = ''" :style="small">Keep</button></template>
<template v-else>
<button type="button" @click="togglePanel(g.id, 'roles')" :aria-expanded="isOpen(g.id, 'roles')" :style="small + (isOpen(g.id, 'roles') ? on : '')">Roles</button>
<button type="button" @click="togglePanel(g.id, 'rights')" :aria-expanded="isOpen(g.id, 'rights')" :style="small + (isOpen(g.id, 'rights') ? on : '')">Rights</button>
<button type="button" @click="edit = { id: g.id, name: g.name, desc: g.desc }" :style="small">Rename</button>
<button type="button" @click="confirming = g.id" :style="small + '; color: var(--k-FF9B8A,#FF9B8A)'">Delete</button></template>
</div></td>
</template>
</tr>
<tr v-if="isOpen(g.id, 'roles')"><td colspan="5" style="padding: 4px 14px 14px; background: var(--k-121B17,#121B17)">
<div style="display: flex; flex-direction: column; gap: 6px; padding-top: 10px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Tap a role to add it to or take it out of {{ g.name }}.</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><button v-for="r in roles" :key="r.id" type="button" :aria-pressed="g.roles.includes(r.id)" :disabled="busy" @click="toggleRole(g, r.id)" :style="chip(g.roles.includes(r.id))">{{ r.name }}</button></div></div></td></tr>
<tr v-if="isOpen(g.id, 'rights')"><td colspan="5" style="padding: 4px 14px 14px; background: var(--k-121B17,#121B17)">
<div style="display: flex; flex-direction: column; gap: 8px; padding-top: 10px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Tick what everyone in {{ g.name }} may do. {{ hint }}</span>
<div v-for="m in modules" :key="m.key" style="border-radius: 12px; border: 1px solid var(--k-24322C,#24322C); overflow: hidden">
<button type="button" @click="toggleMod(g.id + m.key)" :aria-expanded="mods.has(g.id + m.key)" style="display: flex; align-items: center; gap: 8px; width: 100%; padding: 10px 12px; border: 0; background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; font-weight: 600; text-align: left"><span style="flex: 1">{{ m.label }}</span><span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)">{{ onCount(g, m) }} of {{ m.rights.length }}</span></button>
<div v-if="mods.has(g.id + m.key)" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(210px, 1fr)); gap: 4px 12px; padding: 10px 12px">
<label v-for="r in m.rights" :key="r.name" style="display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer"><input type="checkbox" :checked="has(g, m.key, r.name)" :disabled="pending(g, m.key, r.name) || busy" @change="setRight(g, m.key, r.name)"><span>{{ r.label || r.name }}</span><span v-if="pending(g, m.key, r.name)" style="font-size: 9px; font-weight: 700; color: var(--k-FFB86B,#FFB86B)">PENDING</span></label>
</div></div></div></td></tr>
</template>
</tbody>
</table></div></div>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';
import * as rolesW from '~/logic/server/widgets/roles.js';

export default {
  name: 'UserGroups',
  data() {
    return {
      btn: 'height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer',
      small: 'height: 30px; padding: 0 12px; border-radius: 15px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; cursor: pointer; white-space: nowrap',
      on: '; border-color: var(--k-B8F36B,#B8F36B); background: var(--k-2C3D34,#2C3D34)',
      headRow: 'text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)',
      th: 'padding: 10px 8px; font-weight: 600; white-space: nowrap',
      form: { name: '', desc: '', roles: [] }, edit: null, confirming: '', open: '', mods: new Set(), busy: false, err: '', ok: ''
    };
  },
  computed: {
    logic() { return useHubStore().logic; },
    st() { return this.logic.state; },
    roles() { return (this.st.serverRoles || []).filter((r) => !r.locked).map((r) => ({ id: r.id, name: r.name })); },
    groups() { return (this.st.serverGroups || []).map((g) => ({ ...g, roles: g.roles || [], desc: g.desc || '', count: g.count || 0 })); },
    roleById() { const m = {}; for (const r of (this.st.serverRoles || [])) m[r.id] = r.name; return m; },
    modules() { return ((this.st.catalog && this.st.catalog.modules) || []).filter((m) => m.rights && m.rights.length); },
    hint() { return 'Changes to rights may need approval first, and show as pending until approved.'; }
  },
  methods: {
    chip(on) { return 'height: 32px; padding: 0 12px; border-radius: 16px; cursor: pointer; font-size: 13px; ' + (on ? 'border: 1px solid var(--k-B8F36B,#B8F36B); background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); font-weight: 600' : 'border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-C9D6CF,#C9D6CF)'); },
    pick(list, id) { const i = list.indexOf(id); if (i === -1) list.push(id); else list.splice(i, 1); },
    roleNames(g) { return g.roles.map((id) => this.roleById[id]).filter(Boolean); },
    granted(g) { const n = this.modules.reduce((t, m) => t + this.onCount(g, m), 0); return n ? n + (n === 1 ? ' right' : ' rights') : 'None'; },
    isOpen(id, kind) { return this.open === id + ':' + kind; },
    togglePanel(id, kind) { this.open = this.isOpen(id, kind) ? '' : id + ':' + kind; },
    toggleMod(k) { const s = new Set(this.mods); if (s.has(k)) s.delete(k); else s.add(k); this.mods = s; },
    has(g, mod, right) { const r = (this.st.groupRights || {})[g.id]; return !!(r && r[mod] && r[mod][right]); },
    pending(g, mod, right) { return (this.st.pendingRights || {})['g:' + g.id + '|' + mod + '|' + right] !== undefined; },
    onCount(g, m) { return m.rights.filter((r) => this.has(g, m.key, r.name)).length; },
    async run(job, ok) {
      if (this.busy) return;
      this.busy = true; this.err = ''; this.ok = '';
      try { await job(); this.ok = ok || ''; } catch (e) { this.err = (e && e.message) || 'That did not go through.'; }
      this.busy = false;
    },
    create() {
      const f = this.form;
      return this.run(async () => { await rolesW.createGroup(this.logic, this.logic.srv, { name: f.name.trim(), desc: f.desc.trim(), roles: f.roles.slice() }); this.form = { name: '', desc: '', roles: [] }; }, 'User group created.');
    },
    saveEdit(g) { const e = this.edit; return this.run(async () => { await rolesW.updateGroup(this.logic, this.logic.srv, g.id, { name: e.name, desc: e.desc }); this.edit = null; }, 'Saved.'); },
    toggleRole(g, id) { const next = g.roles.includes(id) ? g.roles.filter((x) => x !== id) : g.roles.concat([id]); return this.run(() => rolesW.updateGroup(this.logic, this.logic.srv, g.id, { roles: next })); },
    remove(g) { return this.run(async () => { await rolesW.deleteGroup(this.logic, this.logic.srv, g.id); this.confirming = ''; }, g.name + ' was deleted.'); },
    setRight(g, mod, right) {
      const nv = !this.has(g, mod, right);
      return this.run(async () => { const r = await rolesW.setGroupRight(this.logic, this.logic.srv, g.id, mod, right, nv); this.ok = r === 'applied' ? 'Saved.' : 'Sent for approval. It shows as pending until approved.'; });
    }
  }
};
</script>
