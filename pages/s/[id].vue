<template>
<section aria-label="Secure share link" style="width: 100%; min-height: 60vh; box-sizing: border-box; display: flex; align-items: center; justify-content: center; padding: 24px 12px">
<div style="width: min(100%, 560px); display: flex; flex-direction: column; gap: 14px">
<SecureInboxCard v-if="share" :share="share"/>
<div v-else style="padding: 20px; border: 1px solid #2A3832; border-radius: 14px; background: #16201C; display: flex; flex-direction: column; gap: 8px">
<strong style="font-size: 17px">Secure share unavailable</strong>
<span role="status" style="color: #9DB0A6; line-height: 1.5">This link is invalid, expired, or is not available in this app session.</span>
<NuxtLink to="/" style="align-self: flex-start; margin-top: 4px; color: #B8F36B">Return to Hub</NuxtLink>
</div>
</div>
</section>
</template>

<script>
import { computed, onBeforeUnmount, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useHubStore } from '~/composables/useHub';
import SecureInboxCard from '~/components/panels/SecureInboxCard.vue';

export default {
  name: 'SecureShareLinkPage',
  components: { SecureInboxCard },
  setup() {
    const route = useRoute();
    const { logic } = useHubStore();
    const clock = ref(Date.now());
    const timer = setInterval(() => { clock.value = Date.now(); }, 1000);
    onBeforeUnmount(() => clearInterval(timer));

    const share = computed(() => {
      void clock.value;
      const item = logic.state.secureShares.find((entry) => entry.id === route.params.id && !entry.recipient);
      if (!item) return null;
      const expired = !item.openedAt && Date.now() >= item.expiresAt;
      const showingSecret = !!item.openedAt && !item.consumed && Date.now() < item.revealUntil;
      return {
        ...item,
        expired,
        showingSecret,
        unavailable: expired || !!item.openedAt,
        passwordValue: logic.state.secureSharePasswords[item.id] || '',
        onPassword: (event) => logic.setState((state) => ({
          secureSharePasswords: { ...state.secureSharePasswords, [item.id]: event.target.value }
        })),
        unlock: () => {
          const current = logic.state.secureShares.find((entry) => entry.id === item.id && !entry.recipient);
          if (!current || current.consumed || current.openedAt || Date.now() >= current.expiresAt) {
            logic.notify('This secure share has expired or has already been opened.', '#FFB86B');
            return;
          }
          if ((logic.state.secureSharePasswords[item.id] || '') !== current.password) {
            logic.notify('That link password is not correct.', '#FF9B8A');
            return;
          }
          const openedAt = Date.now();
          logic.setState((state) => ({
            secureShares: state.secureShares.map((entry) => entry.id === item.id ? { ...entry, openedAt, revealUntil: openedAt + 10000 } : entry),
            secureSharePasswords: { ...state.secureSharePasswords, [item.id]: '' }
          }));
          logic.notify('Secure share opened. Its contents will disappear in 10 seconds.', '#B8F36B');
          setTimeout(() => logic.setState((state) => ({
            secureShares: state.secureShares.map((entry) => entry.id === item.id ? { ...entry, consumed: true, revealUntil: 0 } : entry)
          })), 10000);
        }
      };
    });
    return { share };
  }
};
</script>
