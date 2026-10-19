<!-- Roles & rights -->
<template><template v-if="show.roles">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Roles &amp; rights</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">People get roles. Roles get rights in each widget. People can only act within their role's rights.</p></div>
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Preview the app as<select :value="actingRole" @change="setActingRole" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in previewOptions" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select></label>
</div>
<div role="tablist" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid var(--k-24322C,#24322C)"><template v-for="(rt, rt__i) in rolesTabs" :key="rt__i"><button role="tab" @click="rt.go" :style="rt.style">{{ rt.label }}</button></template></div>
<template v-if="!rolesView.isFlows && !rolesView.isReporting">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 14px 20px; padding: 14px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-B8F36B,#B8F36B)">
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Working on role<select :value="roleBar.value" @change="setRoleSel" aria-label="Role you are working on" style="height: 40px; min-width: 200px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; font-weight: 600"><template v-for="(o, o__i) in roleBar.options" :key="o__i"><option :value="o.v">{{ o.l }}</option></template></select></label>
<div style="flex: 1 1 260px; min-width: 0; display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ roleBar.desc }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ roleBar.scope }} · {{ roleBar.count }} {{ roleBar.count === 1 ? 'person' : 'people' }} in this role<template v-if="roleBar.locked"> · rights are locked</template></span></div>
<template v-if="roleBar.hasRole && !rolesView.isMembers">
<template v-if="!roleBar.editing"><button type="button" @click="roleBar.startEdit" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Rename</button>
<button type="button" @click="roleBar.toggleDefault" :aria-pressed="roleBar.isDefault" :style="`height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; ${roleBar.isDefault ? 'border-color: #B8F36B; color: #B8F36B' : ''}`">{{ roleBar.isDefault ? 'Default for new people ✓' : 'Make default for new people' }}</button>
<button v-if="roleBar.canDelete" type="button" @click="roleBar.remove" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px; border-color: #FF9B8A; color: #FF9B8A">Delete role</button></template>
<form v-else novalidate @submit.prevent="roleBar.saveEdit" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; flex: 1 1 100%"><input type="text" :value="roleBar.editName" @input="roleBar.onEditName" aria-label="Role name" placeholder="Role name" style="flex: 1 1 180px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><input type="text" :value="roleBar.editDesc" @input="roleBar.onEditDesc" aria-label="What this role is for" placeholder="What this role is for" style="flex: 2 1 260px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><button type="submit" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save</button><button type="button" @click="roleBar.cancelEdit" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Cancel</button></form>
</template>
<template v-if="rolesView.isMembers"><label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Add a person<select value="" @change="addToRole" aria-label="Add a person to this role" style="height: 40px; min-width: 200px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Choose a person…</option><template v-for="(a, a__i) in roleBar.addOpts" :key="a__i"><option :value="a.v">{{ a.l }}</option></template></select></label></template>
</div>
</template>
<template v-if="rolesView.isRoles">
<form novalidate @submit.prevent="newRoleForm.create" aria-label="Create a role" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; padding: 14px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)">
<strong style="font-size: 14px; flex: 1 1 100%">New role</strong>
<input type="text" :value="newRoleForm.name" @input="newRoleForm.onName" aria-label="New role name" placeholder="Role name, e.g. Records officer" style="flex: 1 1 200px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px">
<input type="text" :value="newRoleForm.desc" @input="newRoleForm.onDesc" aria-label="What the new role is for" placeholder="What this role is for (optional)" style="flex: 2 1 260px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px">
<select :value="newRoleForm.copyFrom" @change="newRoleForm.onCopy" aria-label="Start from" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in newRoleForm.copyOpts" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select>
<button type="submit" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create role</button>
<span v-if="newRoleForm.hasErr" role="alert" style="flex: 1 1 100%; font-size: 13px; color: #FF9B8A">{{ newRoleForm.err }}</span>
</form>
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Roles are made here and given rights on the Rights tab. Nothing is built in: the first role you make is up to you.</span><DataTable :t="rolesTable" /></template>
<template v-if="rolesView.isRights"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ rightsHint }}</span>
<details style="padding: 12px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C)"><summary style="cursor: pointer; font-size: 14px; font-weight: 600">Add your own right or page</summary>
<form novalidate @submit.prevent="newRightForm.add" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; margin-top: 10px">
<select :value="newRightForm.module" @change="newRightForm.onModule" aria-label="Page for the new right" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in newRightForm.moduleOptions" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select>
<input type="text" :value="newRightForm.name" @input="newRightForm.onName" aria-label="Name of the new right" placeholder="e.g. Approve refunds" style="flex: 1 1 200px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px">
<select :value="newRightForm.kind" @change="newRightForm.onKind" aria-label="Kind of right" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in newRightForm.kindOptions" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select>
<button type="submit" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Add right</button>
<span v-if="newRightForm.hasErr" role="alert" style="flex: 1 1 100%; font-size: 13px; color: #FF9B8A">{{ newRightForm.err }}</span>
</form>
<form novalidate @submit.prevent="newModuleForm.add" style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center; margin-top: 10px">
<input type="text" :value="newModuleForm.name" @input="newModuleForm.onName" aria-label="Name of the new page" placeholder="A new page, e.g. Procurement" style="flex: 1 1 240px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px">
<button type="submit" style="height: 36px; padding: 0 14px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Add page</button>
<span v-if="newModuleForm.hasErr" role="alert" style="flex: 1 1 100%; font-size: 13px; color: #FF9B8A">{{ newModuleForm.err }}</span>
</form></details>
<DataTable :t="rightsTable" /></template>
<template v-if="rolesView.isMembers"><DataTable :t="membersTable" /></template>
<template v-if="rolesView.isReporting">
<span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6); line-height: 1.5">Pick the role each role reports to, then the person in that role they report to (for example IT support reports to Head of IT, and that person is Hannah). Profiles show that person. If no person is chosen, profiles show the role name. Anyone can still be given a different manager in their own edit form.</span>
<DataTable :t="reportingTable" />
<div v-if="reportingBar.canEdit" style="position: sticky; bottom: 0; z-index: 5; display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 12px 16px; border-radius: 14px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-34453D,#34453D)">
<span :style="`flex: 1 1 140px; font-size: 13px; color: ${reportingBar.dirty ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ reportingBar.dirty ? 'You have unsaved changes' : 'All changes saved' }}</span>
<button :disabled="!reportingBar.dirty" @click="reportingBar.discard" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Discard</button>
<button :disabled="!reportingBar.dirty" @click="reportingBar.save" :style="`height: 40px; padding: 0 22px; border-radius: 20px; border: 0; font-weight: 600; cursor: ${reportingBar.dirty ? 'pointer' : 'default'}; background: ${reportingBar.dirty ? '#B8F36B' : 'var(--k-2A3832,#2A3832)'}; color: ${reportingBar.dirty ? '#0F1714' : 'var(--k-9DB0A6,#9DB0A6)'}`">Save changes</button>
</div>
<span v-else style="font-size: 12px; color: var(--k-FFD66B,#FFD66B)">You can view reporting lines. Ask an admin for the "Manage roles & rights" right to change them.</span>
</template>
<template v-if="rolesView.isFlows"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Pick how many approval levels each action needs. Who approves a level is not typed in: it is whoever holds the right "Approve level N" on that page, which you give to roles on the Rights tab.</span><DataTable :t="flowsTable" /></template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'RolesSection',
  ...hubBindings(["reportingTable","reportingBar","actingRole","flowsTable","membersTable","rightsHint","rightsTable","roleBar","rolesTable","rolesTabs","newRoleForm","newRightForm","newModuleForm","previewOptions","rolesView","setActingRole","setRoleSel","addToRole","show"])
};
</script>
