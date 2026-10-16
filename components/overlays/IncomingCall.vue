<!-- Someone is calling. Only shown when your privacy settings let the call through. -->
<template><template v-if="incomingView">
<div role="alertdialog" aria-label="Incoming call" style="position: fixed; top: 84px; right: 20px; z-index: 90; width: 320px; max-width: calc(100vw - 28px); box-sizing: border-box; padding: 16px; border-radius: 18px; background: #1D2A25; border: 1px solid #B8F36B; box-shadow: 0 16px 40px rgba(0,0,0,0.5); display: flex; flex-direction: column; gap: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: #B8F36B; font-weight: 600">● Incoming {{ incomingView.kindLabel.toLowerCase() }}</span><strong style="font-size: 17px">{{ incomingView.name }}</strong></div>
<div style="display: flex; gap: 8px"><button @click="incomingView.accept" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; cursor: pointer">Accept</button><button @click="incomingView.decline" style="flex: 1; height: 42px; border-radius: 21px; border: 0; background: #E5484D; color: #fff; font-weight: 700; cursor: pointer">Decline</button></div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'IncomingCall',
  ...hubBindings(["incomingView"])
};
</script>
