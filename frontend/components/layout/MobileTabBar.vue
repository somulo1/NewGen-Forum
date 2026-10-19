<!-- Mobile bottom tab bar -->
<template><template v-if="showTabbar">
<nav aria-label="Quick navigation" class="hub-tabbar">
<template v-for="(tb, tb__i) in tabbar" :key="tb__i">
<button @click="tb.go" :aria-label="tb.label" :style="tb.style"><span style="position: relative; display: flex"><svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path :d="tb.icon"></path></svg><template v-if="tb.hasDot"><span style="position: absolute; top: -4px; right: -8px; min-width: 16px; height: 16px; padding: 0 4px; box-sizing: border-box; border-radius: 8px; background: #C9B6FF; color: #0F1714; font-size: 10px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ tb.dot }}</span></template></span><span style="font-size: 11px">{{ tb.label }}</span></button>
</template>
</nav>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'MobileTabBar',
  ...hubBindings(["showTabbar","tabbar"])
};
</script>
