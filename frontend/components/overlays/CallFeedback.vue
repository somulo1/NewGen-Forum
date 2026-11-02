<!-- After a call, meeting or standup someone else also joined: rate it (1 to 5 stars) and leave a remark if you like. -->
<template><template v-if="callFb">
<div role="dialog" aria-modal="true" aria-label="Rate the call" style="position: fixed; inset: 0; z-index: 95; background: rgba(0,0,0,0.45); display: flex; align-items: center; justify-content: center; padding: 16px">
<div style="width: 380px; max-width: 100%; box-sizing: border-box; padding: 20px; border-radius: 18px; background: var(--k-1D2A25,#1D2A25); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 16px 40px rgba(0,0,0,0.5); display: flex; flex-direction: column; gap: 12px; color: var(--k-E8EFEA,#E8EFEA)">
<div style="display: flex; flex-direction: column; gap: 3px"><strong style="font-size: 17px">How was the {{ callFb.what }}?</strong><span style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">{{ callFb.when }}<template v-if="callFb.with"> · with {{ callFb.with }}</template></span></div>
<div role="radiogroup" aria-label="Rating" style="display: flex; gap: 6px; justify-content: center"><template v-for="st in callFb.starList" :key="st.n"><button type="button" role="radio" :aria-checked="st.on" :aria-label="st.label" @click="st.pick" :style="`width: 42px; height: 42px; border: 0; background: transparent; cursor: pointer; font-size: 32px; line-height: 1; color: ${st.on ? '#F5C451' : 'var(--k-3F524A,#3F524A)'}`">★</button></template></div>
<textarea :value="callFb.remark" @input="callFb.onRemark" placeholder="Anything to add? (optional)" aria-label="Remark" rows="3" maxlength="2000" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font: inherit; font-size: 14px; resize: vertical"></textarea>
<template v-if="callFb.err"><span role="alert" style="font-size: 13px; color: var(--k-FF9B8A,#FF9B8A)">{{ callFb.err }}</span></template>
<div style="display: flex; gap: 8px"><button @click="callFb.later" :disabled="callFb.busy" style="flex: 1; height: 42px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 14px">Later</button><button @click="callFb.send" :disabled="callFb.busy" style="flex: 1; height: 42px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 14px">{{ callFb.busy ? 'Sending…' : 'Send' }}</button></div>
</div>
</div>
</template></template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CallFeedback',
  ...hubBindings(["callFb"])
};
</script>
