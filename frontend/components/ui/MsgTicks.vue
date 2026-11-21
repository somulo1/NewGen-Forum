<!-- Delivery ticks on my own messages: one tick while it waits to be sent, two grey ticks once it is delivered, two blue ticks once read. -->
<template>
<svg :width="size" :height="size" viewBox="0 0 24 24" fill="none" :stroke="state === 'read' ? '#4FC3F7' : '#8A988F'" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round" role="img" :aria-label="label" style="vertical-align: -3px; margin-left: 3px; flex: none">
<template v-if="state === 'wait'"><path d="M5 12.8l4.6 4.6L19 7.4"/></template>
<template v-else><path d="M1.5 12.8l4.6 4.6L15 8"/><path d="M9 15.6l1.8 1.8L21.5 7"/></template>
</svg>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({ state: { type: String, default: 'sent' }, size: { type: Number, default: 15 } });
const label = computed(() => ({ wait: 'Waiting to be sent', sent: 'Delivered', read: 'Read' })[props.state] || 'Delivered');
</script>
