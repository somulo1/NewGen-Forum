<!-- People directory -->
<template><template v-if="show.people">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">People</h1><p style="margin: 4px 0 0; color: var(--k-9DB0A6,#9DB0A6)">{{ peopleSummary }}</p></div>
<template v-if="onb.canOnboard || structureEd.canEdit"><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-if="structureEd.canEdit"><button @click="structureEd.openIt" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Organization structure</button></template><template v-if="onb.canOnboard"><button @click="onb.openBulk" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-weight: 600; cursor: pointer">Bulk onboarding</button><button @click="onb.openOne" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Onboard person</button></template></div></template>
</div>
<DataTable :t="peopleTable" />
<div v-if="userActions.open" role="presentation" @click.self="userActions.close" style="position: fixed; inset: 0; z-index: 1000; display: flex; align-items: center; justify-content: center; padding: 20px; background: rgba(4,8,6,0.72)">
<section role="dialog" aria-modal="true" :aria-label="`Actions for ${userActions.name}`" style="width: min(100%, 420px); box-sizing: border-box; padding: 22px; border-radius: 18px; border: 1px solid var(--k-34453D,#34453D); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); box-shadow: 0 24px 70px rgba(0,0,0,0.55); display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; justify-content: space-between; align-items: flex-start; gap: 12px"><div><h2 style="margin: 0; font: 600 20px 'Space Grotesk', sans-serif">User actions</h2><p style="margin: 5px 0 0; color: var(--k-9DB0A6,#9DB0A6); font-size: 13px">{{ userActions.name }} · {{ userActions.email }}</p></div><button aria-label="Close user actions" @click="userActions.close" style="width: 34px; height: 34px; border: 1px solid var(--k-34453D,#34453D); border-radius: 50%; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 18px">×</button></div>
<div style="padding: 12px; border-radius: 12px; background: var(--k-121B17,#121B17); color: var(--k-C9D6CF,#C9D6CF); font-size: 13px">Account status: <strong :style="`color: ${userActions.blocked ? 'var(--k-FF9B8A,#FF9B8A)' : userActions.active ? 'var(--k-B8F36B,#B8F36B)' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ userActions.blocked ? 'Blocked' : userActions.active ? 'Active' : 'Inactive' }}</strong></div>
<div style="display: grid; grid-template-columns: 1fr 1fr; gap: 8px">
<button @click="userActions.activate" style="min-height: 40px; border: 1px solid var(--k-34453D,#34453D); border-radius: 10px; background: var(--k-22302A,#22302A); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">{{ userActions.active ? 'Deactivate user' : 'Activate user' }}</button>
<button @click="userActions.toggleBlocked" style="min-height: 40px; border: 1px solid var(--k-34453D,#34453D); border-radius: 10px; background: var(--k-22302A,#22302A); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">{{ userActions.blocked ? 'Unblock user' : 'Block user' }}</button>
<button @click="userActions.resetPassword" style="min-height: 40px; border: 1px solid var(--k-34453D,#34453D); border-radius: 10px; background: var(--k-22302A,#22302A); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Reset password</button>
<button @click="confirmUserDelete" style="min-height: 40px; border: 1px solid var(--k-8A4842,#8A4842); border-radius: 10px; background: var(--k-3A211F,#3A211F); color: #FFB8AF; cursor: pointer">Delete user</button>
</div>
<div v-if="userActions.tempPassword" role="status" style="padding: 12px; border-radius: 12px; border: 1px solid var(--k-536A46,#536A46); background: var(--k-1E2B1A,#1E2B1A); font-size: 13px">
Temporary password for {{ userActions.name }}:
<code style="display: block; margin-top: 6px; color: #D7F5B8; overflow-wrap: anywhere">{{ userActions.tempPassword }}</code>
<span style="display: block; margin-top: 6px; color: var(--k-9DB0A6,#9DB0A6)">Share it securely; this password is shown only in this dialog.</span>
</div>
<button @click="userActions.close" style="height: 40px; border: 1px solid var(--k-34453D,#34453D); border-radius: 20px; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Done</button>
</section>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PeopleSection',
  ...hubBindings(["onb","structureEd","peopleSummary","peopleTable","show","userActions"]),
  methods: {
    confirmUserDelete() {
      if (window.confirm('Delete ' + this.userActions.name + ' from the people directory?')) this.userActions.remove();
    }
  }
};
</script>
