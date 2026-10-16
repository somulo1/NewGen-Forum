<!-- Meetings & standups -->
<template><template v-if="show.meetings">
<div style="max-width: 1280px; display: flex; flex-direction: column; gap: 16px; position: relative; min-height: 600px">
<CallScreen v-if="inCall && !callContext" />
<template v-if="!inCall || callContext">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Meetings &amp; standups</h1><p style="margin: 4px 0 0; color: #9DB0A6">{{ meetingsSummary }}</p></div>
<DataTable :t="meetingsTable" />
</template>
</div>
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
