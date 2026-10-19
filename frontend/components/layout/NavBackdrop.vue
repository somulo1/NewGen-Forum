<!-- Mobile: dark backdrop behind the slide-out menu -->
<template><template v-if="navOpen"><button class="hub-navback hub-m" aria-label="Close menu" @click="closeNav"></button></template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'NavBackdrop',
  ...hubBindings(["closeNav","navOpen"])
};
</script>
