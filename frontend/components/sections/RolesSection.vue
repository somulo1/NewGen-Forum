<!-- Roles & rights -->
<template><template v-if="show.roles">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Roles &amp; rights</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">People get roles. Roles get rights in each widget. People can only act within their role's rights.</p></div>
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Preview the app as<select :value="actingRole" @change="setActingRole" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="admin">Admin (you)</option><option value="hod">Head of department</option><option value="it">IT support</option><option value="finance">Finance approver</option><option value="staff">Staff</option><option value="guest">Guest</option></select></label>
</div>
<div role="tablist" style="display: flex; flex-wrap: wrap; gap: 4px; border-bottom: 1px solid var(--k-24322C,#24322C)"><template v-for="(rt, rt__i) in rolesTabs" :key="rt__i"><button role="tab" @click="rt.go" :style="rt.style">{{ rt.label }}</button></template></div>
<template v-if="!rolesView.isFlows && !rolesView.isReporting">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 14px 20px; padding: 14px 16px; border-radius: 14px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-B8F36B,#B8F36B)">
<label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Working on role<select :value="roleBar.value" @change="setRoleSel" aria-label="Role you are working on" style="height: 40px; min-width: 200px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; font-weight: 600"><template v-for="(o, o__i) in roleBar.options" :key="o__i"><option :value="o.v">{{ o.l }}</option></template></select></label>
<div style="flex: 1 1 260px; min-width: 0; display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ roleBar.desc }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ roleBar.scope }} · {{ roleBar.count }} {{ roleBar.count === 1 ? 'person' : 'people' }} in this role<template v-if="roleBar.locked"> · rights are locked</template></span></div>
<template v-if="rolesView.isMembers"><label style="display: flex; align-items: center; gap: 8px; font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">Add a person<select value="" @change="addToRole" aria-label="Add a person to this role" style="height: 40px; min-width: 200px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Choose a person…</option><template v-for="(a, a__i) in roleBar.addOpts" :key="a__i"><option :value="a.v">{{ a.l }}</option></template></select></label></template>
</div>
</template>
<template v-if="rolesView.isRoles"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Each role has a scope. Department, sector, branch and unit roles only appear when your organization uses that level.</span><DataTable :t="rolesTable" /></template>
<template v-if="rolesView.isRights"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ rightsHint }}</span><DataTable :t="rightsTable" /></template>
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
<template v-if="rolesView.isFlows"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Pick how many approval levels each action needs and who approves at each level.</span><DataTable :t="flowsTable" /></template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'RolesSection',
  ...hubBindings(["reportingTable","reportingBar","actingRole","flowsTable","membersTable","rightsHint","rightsTable","roleBar","rolesTable","rolesTabs","rolesView","setActingRole","setRoleSel","addToRole","show"])
};
</script>
