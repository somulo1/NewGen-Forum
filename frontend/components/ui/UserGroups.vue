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

<div v-if="err" role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 14px">{{ err }}</div>
<div v-if="ok" role="status" style="padding: 10px 14px; border-radius: 12px; background: var(--k-1E3326,#1E3326); color: #B8F36B; font-size: 14px">{{ ok }}</div>
<span v-if="!groups.length" style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">No user groups yet.</span>

<div v-for="g in groups" :key="g.id" style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px">
<template v-if="edit && edit.id === g.id">
<input v-model="edit.name" maxlength="60" :aria-label="`Name of ${g.name}`" style="flex: 1 1 180px; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)">
<input v-model="edit.desc" maxlength="300" placeholder="What it is for" :aria-label="`About ${g.name}`" style="flex: 2 1 220px; height: 36px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA)">
<button type="button" @click="saveEdit(g)" :disabled="busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save</button>
<button type="button" @click="edit = null" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button>
</template>
<template v-else>
<div style="flex: 1 1 220px; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ g.name }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ g.desc || 'No description' }} · {{ g.roles.length }} {{ g.roles.length === 1 ? 'role' : 'roles' }} · {{ g.count }} {{ g.count === 1 ? 'person' : 'people' }}</span></div>
<template v-if="confirming === g.id"><span style="font-size: 13px; color: var(--k-FFB86B,#FFB86B)">Delete {{ g.name }}? Its rights go with it.</span>
<button type="button" @click="remove(g)" :disabled="busy" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Delete</button>
<button type="button" @click="confirming = ''" style="height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Keep</button></template>
<template v-else>
<button type="button" @click="open = open === g.id ? '' : g.id" :aria-expanded="open === g.id" :style="btn">{{ open === g.id ? 'Hide rights' : 'Rights' }}</button>
<button type="button" @click="edit = { id: g.id, name: g.name, desc: g.desc }" :style="btn">Rename</button>
<button type="button" @click="confirming = g.id" :style="`${btn}; color: var(--k-FF9B8A,#FF9B8A)`">Delete</button></template>
</template>
</div>
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Roles in this group</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><button v-for="r in roles" :key="r.id" type="button" :aria-pressed="g.roles.includes(r.id)" :disabled="busy" @click="toggleRole(g, r.id)" :style="chip(g.roles.includes(r.id))">{{ r.name }}</button></div></div>
<template v-if="open === g.id">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Tick what everyone in {{ g.name }} may do. {{ hint }}</span>
<div v-for="m in modules" :key="m.key" style="border-radius: 12px; border: 1px solid var(--k-24322C,#24322C); overflow: hidden">
<button type="button" @click="toggleMod(g.id + m.key)" :aria-expanded="mods.has(g.id + m.key)" style="display: flex; align-items: center; gap: 8px; width: 100%; padding: 10px 12px; border: 0; background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; font-weight: 600; text-align: left"><span style="flex: 1">{{ m.label }}</span><span style="font-weight: 400; color: var(--k-9DB0A6,#9DB0A6)">{{ onCount(g, m) }} of {{ m.rights.length }}</span></button>
<div v-if="mods.has(g.id + m.key)" style="display: grid; grid-template-columns: repeat(auto-fill, minmax(210px, 1fr)); gap: 4px 12px; padding: 10px 12px">
<label v-for="r in m.rights" :key="r.name" style="display: flex; align-items: center; gap: 8px; font-size: 13px; cursor: pointer"><input type="checkbox" :checked="has(g, m.key, r.name)" :disabled="pending(g, m.key, r.name) || busy" @change="setRight(g, m.key, r.name)"><span>{{ r.label || r.name }}</span><span v-if="pending(g, m.key, r.name)" style="font-size: 9px; font-weight: 700; color: var(--k-FFB86B,#FFB86B)">PENDING</span></label>
</div></div>
</template>
</div>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';
import * as rolesW from '~/logic/server/widgets/roles.js';

export default {
  name: 'UserGroups',
  data() { return { btn: 'height: 34px; padding: 0 16px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer', form: { name: '', desc: '', roles: [] }, edit: null, confirming: '', open: '', mods: new Set(), busy: false, err: '', ok: '' }; },
  computed: {
    logic() { return useHubStore().logic; },
    st() { return this.logic.state; },
    roles() { return (this.st.serverRoles || []).filter((r) => !r.locked).map((r) => ({ id: r.id, name: r.name })); },
    groups() { return (this.st.serverGroups || []).map((g) => ({ ...g, roles: g.roles || [], desc: g.desc || '', count: g.count || 0 })); },
    modules() { return ((this.st.catalog && this.st.catalog.modules) || []).filter((m) => m.rights && m.rights.length); },
    hint() { return 'Changes to rights may need approval first, and show as pending until approved.'; }
  },
  methods: {
    chip(on) { return 'height: 32px; padding: 0 12px; border-radius: 16px; cursor: pointer; font-size: 13px; ' + (on ? 'border: 1px solid var(--k-B8F36B,#B8F36B); background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); font-weight: 600' : 'border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-C9D6CF,#C9D6CF)'); },
    pick(list, id) { const i = list.indexOf(id); if (i === -1) list.push(id); else list.splice(i, 1); },
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
