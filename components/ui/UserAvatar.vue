<template>
<span :aria-hidden="!ariaLabel" :aria-label="ariaLabel || undefined" :role="ariaLabel ? 'img' : undefined" :style="`position: relative; display: flex; align-items: center; justify-content: center; flex: none; width: ${size}px; height: ${size}px; box-sizing: border-box; overflow: hidden; border-radius: ${radius}; background: ${color}; color: ${textColor}; font-weight: 700; font-size: ${fontSize}px; ${border ? `border: ${border};` : ''}`">
<img v-if="isSelf && photo" :src="photo" :alt="ariaLabel ? '' : undefined" style="position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover">
<span v-else>{{ initials }}</span>
</span>
</template>

<script>
import { computed } from 'vue';
import { useHubStore } from '~/composables/useHub';

export default {
  name: 'UserAvatar',
  props: {
    name: { type: String, default: '' },
    initials: { type: String, default: '' },
    color: { type: String, default: '#C9B6FF' },
    size: { type: Number, default: 36 },
    fontSize: { type: Number, default: 12 },
    radius: { type: String, default: '50%' },
    border: { type: String, default: '' },
    textColor: { type: String, default: '#0F1714' },
    ariaLabel: { type: String, default: '' }
  },
  setup(props) {
    const { logic } = useHubStore();
    const isSelf = computed(() => ['sam', 'samwel okoth omulo', 'you'].includes(String(props.name || '').trim().toLowerCase()));
    const photo = computed(() => logic.state.profile.photo);
    return { isSelf, photo };
  }
};
</script>
