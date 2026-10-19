<!-- Toast notification -->
<template><template v-if="hasToast">
<div role="status" class="hub-toast" style="position: fixed; left: 50%; transform: translateX(-50%); top: 76px; z-index: 95; width: max-content; max-width: min(420px, calc(100vw - 24px)); box-sizing: border-box; padding: 14px 16px; border-radius: 14px; background: var(--k-E8EFEA,#E8EFEA); color: var(--k-0F1714,#0F1714); font-size: 14px; font-weight: 500; box-shadow: 0 14px 34px rgba(0,0,0,0.45); display: flex; gap: 10px; align-items: flex-start"><span :style="`width: 8px; height: 8px; margin-top: 6px; flex: none; border-radius: 50%; background: ${toastColor}`"></span>{{ toast }}</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'AppToast',
  ...hubBindings(["hasToast","toast","toastColor"])
};
</script>
