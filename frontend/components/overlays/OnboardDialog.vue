<!-- Onboard people: one at a time, or many from a CSV. Structure fields follow the organization. -->
<template><template v-if="onb.open">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 70; background: rgba(5,10,8,0.6); display: flex; align-items: center; justify-content: center; padding: 16px">
<div role="dialog" aria-modal="true" aria-label="Onboard people" style="width: 720px; max-width: 100%; max-height: 100%; display: flex; flex-direction: column; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 18px 22px 0">
<h2 style="flex: 1; margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">{{ onb.heading }}</h2>
<button type="button" aria-label="Close" @click="onb.close" style="width: 36px; height: 36px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">✕</button>
</div>
<div v-if="onb.showTabs" role="tablist" style="flex: none; display: flex; gap: 4px; padding: 10px 22px 0; border-bottom: 1px solid var(--k-24322C,#24322C)">
<button role="tab" :aria-selected="onb.isOne" @click="onb.goOne" :style="`min-height: 42px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${onb.isOne ? 'var(--k-B8F36B,#B8F36B)' : 'transparent'}; color: ${onb.isOne ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)'}; font-weight: ${onb.isOne ? 600 : 500}`">One person</button>
<button role="tab" :aria-selected="onb.isBulk" @click="onb.goBulk" :style="`min-height: 42px; padding: 0 16px; border: 0; background: transparent; cursor: pointer; font-size: 14px; border-bottom: 2px solid ${onb.isBulk ? 'var(--k-B8F36B,#B8F36B)' : 'transparent'}; color: ${onb.isBulk ? 'var(--k-E8EFEA,#E8EFEA)' : 'var(--k-9DB0A6,#9DB0A6)'}; font-weight: ${onb.isBulk ? 600 : 500}`">Bulk onboarding</button>
</div>

<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 22px 22px; display: flex; flex-direction: column; gap: 14px">
<template v-if="onb.doneOne"><div role="status" style="padding: 12px 14px; border-radius: 12px; background: var(--k-1B2A1E,#1B2A1E); color: var(--k-B8F36B,#B8F36B); font-size: 14px">{{ onb.done.one }} was onboarded. You can add another person below.</div></template>
<template v-if="onb.doneBulk"><div role="status" style="padding: 12px 14px; border-radius: 12px; background: var(--k-1B2A1E,#1B2A1E); color: var(--k-B8F36B,#B8F36B); font-size: 14px">{{ onb.done.count }} {{ onb.done.count === 1 ? 'person was' : 'people were' }} onboarded<template v-if="onb.done.skipped">. {{ onb.done.skipped }} row{{ onb.done.skipped === 1 ? '' : 's' }} skipped because of errors</template>.</div></template>

<template v-if="onb.isOne">
<form novalidate aria-label="Onboard a person" @submit.prevent="onb.submit" style="display: flex; flex-direction: column; gap: 14px">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">All details marked * are required.</span>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 12px">
<template v-for="(fd, fd__i) in onb.fields" :key="fd.key">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ fd.label }} *
<template v-if="fd.isSelect"><select :value="fd.value" @change="fd.onInput" :aria-invalid="fd.hasErr" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Choose…</option><template v-for="(go, go__i) in fd.options" :key="go__i"><option :value="go">{{ go }}</option></template></select></template>
<template v-if="!fd.isSelect"><input :type="fd.type" :value="fd.value" @input="fd.onInput" :disabled="onb.isEdit && fd.key === 'email'" :title="onb.isEdit && fd.key === 'email' ? 'The email is the person\'s sign-in and cannot be changed here' : ''" :aria-invalid="fd.hasErr" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></template>
<template v-if="fd.hasErr"><span role="alert" style="font-size: 12px; color: var(--k-FF9B8A,#FF9B8A)">{{ fd.err }}</span></template>
</label>
</template>
</div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Job title (optional)<input type="text" :value="onb.title" @input="onb.onTitle" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17)">
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ onb.structureNote }}</span>
<template v-if="onb.hasStructure">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(240px, 1fr)); gap: 12px">
<template v-for="(st, st__i) in onb.structure" :key="st.key">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ st.label }}<template v-if="!st.label.includes('optional')"> *</template>
<select :value="st.value" @change="st.onChange" :aria-invalid="st.hasErr" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">Choose…</option><template v-for="(op, op__i) in st.options" :key="op__i"><option :value="op">{{ op }}</option></template></select>
<template v-if="st.hasErr"><span role="alert" style="font-size: 12px; color: var(--k-FF9B8A,#FF9B8A)">{{ st.err }}</span></template>
</label>
</template>
</div>
</template>
</div>
<template v-if="onb.hasAccess">
<div style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17)">
<strong style="font-size: 14px">Role and rights</strong>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Rights start from the role. Changes made here apply to this person only.</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Role<select :value="onb.access.roleValue" :disabled="!onb.access.canEdit" @change="onb.access.onRole" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option v-if="!onb.access.roleValue" value="">No role (choose one)</option><template v-for="r in onb.access.roles" :key="r.v"><option :value="r.v">{{ r.l }}</option></template></select></label>
<label v-if="onb.access.reports.show" style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Reports to ({{ onb.access.reports.roleName }})<select :value="onb.access.reports.value" :disabled="!onb.access.reports.canEdit" @change="onb.access.reports.onChange" aria-label="Who this person reports to" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><template v-for="o in onb.access.reports.options" :key="o.v"><option :value="o.v">{{ o.l }}</option></template></select><span style="font-size: 11px">Leave as the role to show "{{ onb.access.reports.roleName }}" on the profile, or pick the exact person.</span></label>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px"><button type="button" :aria-expanded="onb.access.open" @click="onb.access.toggleOpen" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">{{ onb.access.open ? 'Hide individual rights' : 'Edit individual rights' }}</button>
<span :style="`font-size: 12px; color: ${onb.access.hasCustom ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-9DB0A6,#9DB0A6)'}`">{{ onb.access.hasCustom ? onb.access.customNote : 'No individual changes' }}</span>
<button v-if="onb.access.hasCustom && onb.access.canEdit" type="button" @click="onb.access.reset" style="height: 32px; padding: 0 14px; border-radius: 16px; border: 0; background: transparent; color: var(--k-B8F36B,#B8F36B); cursor: pointer; font-size: 12px; text-decoration: underline">Reset to role</button></div>
<span v-if="!onb.access.canEdit" style="font-size: 12px; color: var(--k-FFD66B,#FFD66B)">Ask an admin for the "Manage roles &amp; rights" right to change these.</span>
<div v-if="onb.access.open" style="max-height: 280px; overflow-y: auto; display: flex; flex-direction: column; gap: 10px; padding-right: 4px">
<template v-for="m in onb.access.modules" :key="m.name"><fieldset style="margin: 0; padding: 10px 12px; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 6px; min-width: 0"><legend :style="`padding: 0 6px; font-size: 12px; font-weight: 700; color: ${m.tint}`">{{ m.name }}</legend>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(min(100%, 200px), 1fr)); gap: 6px">
<template v-for="r in m.rights" :key="r.label"><label :style="`display: flex; align-items: center; gap: 8px; min-height: 36px; padding: 0 10px; border-radius: 10px; font-size: 13px; cursor: ${r.disabled ? 'default' : 'pointer'}; background: ${r.custom ? 'var(--k-2A2314,#2A2314)' : 'var(--k-16201C,#16201C)'}; border: 1px solid ${r.custom ? 'var(--k-FFB86B,#FFB86B)' : 'var(--k-24322C,#24322C)'}`"><input type="checkbox" :checked="r.on" :disabled="r.disabled" @change="r.toggle" :aria-label="`${m.name}: ${r.label}`"><span style="flex: 1; min-width: 0">{{ r.label }}</span><span v-if="r.custom" style="font-size: 10px; color: var(--k-FFB86B,#FFB86B)">custom</span></label></template>
</div></fieldset></template>
</div>
</div>
</template>
<div v-if="onb.formError" role="alert" style="padding: 10px 12px; border-radius: 10px; background: rgba(255,155,138,0.14); color: #FF9B8A; font-size: 13px">{{ onb.formError }}</div>
<div style="display: flex; gap: 8px; justify-content: flex-end"><button type="button" @click="onb.close" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Close</button><button type="submit" :disabled="onb.saving" :style="`height: 40px; padding: 0 22px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: ${onb.saving ? 'default' : 'pointer'}; opacity: ${onb.saving ? 0.7 : 1}`">{{ onb.saving ? 'Saving…' : onb.submitLabel }}</button></div>
</form>
</template>

<template v-if="onb.isBulk">
<div style="display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">
<span>Upload a CSV or paste rows. The first row names the columns:</span>
<code style="padding: 8px 10px; border-radius: 8px; background: var(--k-121B17,#121B17); font-size: 12px; color: var(--k-B8F36B,#B8F36B); overflow-x: auto; white-space: nowrap">{{ onb.bulk.columns }}</code>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ onb.bulk.structureNote }} Dates can be 1994-03-21 or 21/03/1994.</span>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center">
<button type="button" @click="onb.bulk.template" style="height: 38px; padding: 0 16px; border-radius: 19px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 13px">Download template</button>
<label style="position: relative; height: 38px; padding: 0 16px; border-radius: 19px; border: 0; background: var(--k-E8EFEA,#E8EFEA); color: var(--k-0F1714,#0F1714); font-weight: 600; cursor: pointer; font-size: 13px; display: flex; align-items: center">Choose CSV file<input type="file" accept=".csv,text/csv,text/plain" aria-label="Choose CSV file" @change="onb.bulk.onFile" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"></label>
<template v-if="onb.bulk.fileName"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ onb.bulk.fileName }}</span></template>
</div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Or paste rows here<textarea rows="5" :value="onb.bulk.text" @input="onb.bulk.onText" placeholder="first_name,second_name,date_of_birth,gender,email,phone" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px; font-family: ui-monospace, monospace; line-height: 1.45; resize: vertical"></textarea></label>
<template v-if="onb.bulk.hasMissing"><span role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ onb.bulk.missingText }}</span></template>
<template v-if="onb.bulk.hasUnknown"><span style="font-size: 12px; color: var(--k-FFB86B,#FFB86B)">{{ onb.bulk.unknownText }}</span></template>
<template v-if="onb.bulk.hasRows">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">{{ onb.bulk.summary }}</span>
<div tabindex="0" aria-label="Rows read from your file" style="max-height: 260px; overflow: auto; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C)">
<table style="width: 100%; min-width: 520px; border-collapse: separate; border-spacing: 0; font-size: 13px">
<thead><tr><th style="position: sticky; top: 0; padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)">Row</th><th style="position: sticky; top: 0; padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)">Person</th><th style="position: sticky; top: 0; padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)">Status</th><th style="position: sticky; top: 0; padding: 8px 12px; text-align: left; font-size: 11px; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6); background: var(--k-1B2822,#1B2822)">What to fix</th></tr></thead>
<tbody>
<template v-for="(r, r__i) in onb.bulk.rows" :key="r.n">
<tr><td style="padding: 8px 12px; border-top: 1px solid var(--k-24322C,#24322C); color: var(--k-9DB0A6,#9DB0A6)">{{ r.n }}</td><td style="padding: 8px 12px; border-top: 1px solid var(--k-24322C,#24322C)"><strong style="font-weight: 600">{{ r.name }}</strong><br><span style="color: var(--k-9DB0A6,#9DB0A6); font-size: 12px">{{ r.email }}</span></td><td style="padding: 8px 12px; border-top: 1px solid var(--k-24322C,#24322C)"><span :style="`font-size: 12px; font-weight: 700; color: ${r.statusColor}`">{{ r.statusLabel }}</span></td><td style="padding: 8px 12px; border-top: 1px solid var(--k-24322C,#24322C); color: var(--k-FF9B8A,#FF9B8A); line-height: 1.4">{{ r.errorText }}</td></tr>
</template>
</tbody>
</table>
</div>
<template v-if="onb.bulk.moreRows"><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">Showing the first 200 rows. {{ onb.bulk.moreRows }} more rows will be checked and imported too.</span></template>
</template>
<div style="display: flex; gap: 8px; justify-content: flex-end"><button type="button" @click="onb.close" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Close</button><button type="button" :disabled="!onb.bulk.canImport" @click="onb.bulk.import" :style="`height: 40px; padding: 0 22px; border-radius: 20px; border: 0; font-weight: 600; cursor: ${onb.bulk.canImport ? 'pointer' : 'default'}; ${onb.bulk.canImport ? 'background: #B8F36B; color: #0F1714' : 'background: var(--k-2A3832,#2A3832); color: var(--k-9DB0A6,#9DB0A6)'}`">{{ onb.bulk.importLabel }}</button></div>
</div>
</template>
</div>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'OnboardDialog',
  ...hubBindings(["onb"])
};
</script>
