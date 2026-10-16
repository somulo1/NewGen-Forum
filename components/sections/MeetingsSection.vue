<!-- Meetings & standups -->
<template><template v-if="show.meetings">
<template v-if="inCall && !callContext">
<div style="width: 100%; min-height: 600px; display: flex; flex-direction: column; background: #0F1714">
<CallScreen />
</div>
</template>
<template v-if="!inCall || callContext">
<div style="max-width: 1280px; display: flex; flex-direction: column; gap: 16px">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Meetings &amp; standups</h1><p style="margin: 4px 0 0; color: #9DB0A6">{{ meetingsSummary }}</p></div>
<DataTable :t="meetingsTable" />
</div>
</template>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';
import CallScreen from '../overlays/CallScreen.vue';

export default {
  name: 'MeetingsSection',
  components: { CallScreen },
  ...hubBindings(["call","inCall","meetingsSummary","meetingsTable","show"]),
  computed: {
    callContext() {
      return this.call?.context || null;
    }
  }
};
</script>
