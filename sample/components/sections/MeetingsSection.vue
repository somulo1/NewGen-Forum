<!-- Meetings & standups -->
<template><template v-if="show.meetings">
<div style="max-width: 860px; display: flex; flex-direction: column; gap: 16px">
<h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Meetings &amp; standups</h1>
<p style="margin: -8px 0 4px; color: #9DB0A6">Schedule meetings and standups, pick who joins, and start them from here.</p>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<button @click="newMeeting" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Schedule a meeting</button>
<button @click="newStandup" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Schedule a standup</button>
<button @click="startInstant" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Start now</button>
</div>
<DataTable :t="meetingsTable" />
<template v-for="(s, s__i) in emptyList" :key="s__i">
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 16px 18px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<span :style="`width: 40px; height: 40px; border-radius: 11px; background: ${s.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700`">{{ s.initials }}</span>
<div style="flex: 1 1 240px; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">{{ s.title }}</strong><span style="color: #9DB0A6; font-size: 13px">{{ s.forum }} · {{ s.when }}</span></div>
<template v-if="s.live"><span style="font-size: 12px; font-weight: 600; color: #FF9B8A">● LIVE</span></template>
<button @click="s.join" :style="s.btnStyle">{{ s.btn }}</button>
</div>
</template>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'MeetingsSection',
  ...hubBindings(["emptyList","meetingsTable","newMeeting","newStandup","show","startInstant"])
};
</script>
