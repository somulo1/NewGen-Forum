<!-- Banner shown while previewing the app as another role -->
<template><template v-if="previewing">
<div role="status" style="flex: none; display: flex; flex-wrap: wrap; align-items: center; gap: 12px; padding: 10px 20px; background: #3A3016; color: #FFE7A8; font-size: 14px; border-bottom: 1px solid #5A4A22">
<strong>Previewing as {{ actingRoleName }}.</strong><span>You only see and do what this role's rights allow.</span>
<button @click="exitPreview" style="margin-left: auto; height: 34px; padding: 0 14px; border-radius: 17px; border: 0; background: #FFD66B; color: #0F1714; font-weight: 600; cursor: pointer">Exit preview</button>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'PreviewBanner',
  ...hubBindings(["actingRoleName","exitPreview","previewing"])
};
</script>
