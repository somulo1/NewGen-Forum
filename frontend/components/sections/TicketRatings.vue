<!-- Ratings on closed tickets: for the people who worked on them and the heads of the place each ticket was with. -->
<template><template v-if="tkFeedback">
<section aria-label="Ticket ratings" style="padding: 16px 18px; border-radius: 16px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 10px">
<button @click="tkFeedback.toggle" :aria-expanded="tkFeedback.open" style="display: flex; flex-wrap: wrap; align-items: center; gap: 6px 12px; border: 0; background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; padding: 0; text-align: left; font: inherit">
<strong style="font-size: 15px; flex: 1 1 200px">Ratings on tickets you handled or head</strong>
<span style="color: #F5C451; letter-spacing: 1px">{{ tkFeedback.stars }}</span><strong>{{ tkFeedback.average }}</strong><span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ tkFeedback.total }} rated</span><span aria-hidden="true" style="color: var(--k-9DB0A6,#9DB0A6)">{{ tkFeedback.open ? '▴' : '▾' }}</span>
</button>
<template v-if="tkFeedback.open">
<template v-if="tkFeedback.err"><span role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ tkFeedback.err }}</span></template>
<div :style="`display: flex; flex-direction: column; gap: 8px; opacity: ${tkFeedback.loading ? 0.6 : 1}`">
<template v-for="x in tkFeedback.items" :key="x.id"><button @click="x.open" style="display: flex; flex-direction: column; gap: 4px; padding: 10px 12px; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; text-align: left; font: inherit">
<span style="display: flex; flex-wrap: wrap; gap: 8px; align-items: baseline"><strong style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ x.ref }}</strong><strong style="font-size: 14px; flex: 1 1 160px; min-width: 0; overflow-wrap: anywhere">{{ x.title }}</strong><span :aria-label="`${x.rating} of 5`" style="color: #F5C451; letter-spacing: 1px">{{ x.stars }}</span></span>
<span v-if="x.remark" style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF); white-space: pre-line; overflow-wrap: anywhere">“{{ x.remark }}”</span>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ x.meta }}</span>
</button></template>
</div>
<template v-if="tkFeedback.many"><div style="display: flex; align-items: center; justify-content: space-between; gap: 8px"><button @click="tkFeedback.prev" :disabled="!tkFeedback.hasPrev || tkFeedback.loading" :style="`height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; opacity: ${tkFeedback.hasPrev ? 1 : 0.4}`">‹ Newer</button><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ tkFeedback.pageLabel }}</span><button @click="tkFeedback.next" :disabled="!tkFeedback.hasNext || tkFeedback.loading" :style="`height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; opacity: ${tkFeedback.hasNext ? 1 : 0.4}`">Older ›</button></div></template>
</template>
</section>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'TicketRatings',
  ...hubBindings(["tkFeedback"])
};
</script>
