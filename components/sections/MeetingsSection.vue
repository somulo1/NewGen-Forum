<!-- Meetings & standups -->
<template><template v-if="show.meetings">
<div style="max-width: 1280px; display: flex; flex-direction: column; gap: 16px">
<div><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Meetings &amp; standups</h1><p style="margin: 4px 0 0; color: #9DB0A6">{{ meetingsSummary }}</p></div>
<DataTable :t="meetingsTable" />
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'MeetingsSection',
  ...hubBindings(["meetingsSummary","meetingsTable","show"])
};
</script>
