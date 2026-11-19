<!-- Ticket timing policy: working hours, when TAT stops counting, and ticket notifications (Settings > Tickets) -->
<template>
<div style="display: flex; flex-direction: column; gap: 22px; max-width: 760px">
<section style="display: flex; flex-direction: column; gap: 12px">
<span :style="heading">Working hours</span>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px">
<label :style="label">Day starts<input type="time" :value="tset.start" @input="setStart" :style="field + time"></label>
<label :style="label">Day ends<input type="time" :value="tset.end" @input="setEnd" :style="field + time"></label>
</div>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">One working day = {{ dayHoursLabel }} of TAT.</span>
</section>
<section style="display: flex; flex-direction: column; gap: 10px">
<span :style="heading">TAT does not count during</span>
<template v-for="pr in pauseRules" :key="pr.label"><label :style="row"><input type="checkbox" :checked="pr.checked" @change="pr.toggle"><span style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ pr.label }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ pr.desc }}</span></span></label></template>
</section>
<section style="display: flex; flex-direction: column; gap: 10px">
<span :style="heading">Notifications</span>
<template v-for="nr in notifyRules" :key="nr.label"><label :style="row"><input type="checkbox" :checked="nr.checked" @change="nr.toggle"><span style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ nr.label }}</span><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ nr.desc }}</span></span></label></template>
<label :style="label">Warn when this much of the TAT is used: {{ tset.nearPct }}%<input type="range" min="50" max="95" step="5" :value="tset.nearPct" @input="setNearPct"></label>
<label :style="label">Escalate overdue tickets to<select :value="tset.escalate" @change="setEscalate" :style="field"><option value="Head of IT Operations">Head of IT Operations</option><option value="Head of Software Development">Head of Software Development</option><option value="HR Manager">HR Manager</option><option value="The requester's department head">The requester's department head</option></select></label>
</section>
</div>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketPolicy',
  data() {
    return {
      heading: 'font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)',
      label: 'display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)',
      field: 'height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px',
      time: '; color-scheme: dark',
      row: 'display: flex; gap: 12px; align-items: flex-start; padding: 10px 12px; border-radius: 12px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C); cursor: pointer'
    };
  },
  ...hubBindings(["dayHoursLabel","notifyRules","pauseRules","setEnd","setEscalate","setNearPct","setStart","tset"])
};
</script>
