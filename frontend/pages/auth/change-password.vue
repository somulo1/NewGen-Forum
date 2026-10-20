<script setup>
// Shown after signing in with a temporary password (a new owner account, or a password an administrator reset).
definePageMeta({ path: '/change-password', layout: 'auth' });
useHead({ title: 'Choose a new password · Hub' });
const be = useBackend();
const current = ref(''), pw = ref(''), pw2 = ref(''), show = ref(false), err = ref(''), busy = ref(false);
const rules = computed(() => [['8 to 128 characters', pw.value.length >= 8 && pw.value.length <= 128], ['An uppercase and a lowercase letter', /[a-z]/.test(pw.value) && /[A-Z]/.test(pw.value)], ['A number', /\d/.test(pw.value)]]);
const submit = async () => {
  err.value = '';
  if (!current.value) { err.value = 'Enter the temporary password you signed in with.'; return; }
  if (rules.value.some((r) => !r[1])) { err.value = 'The new password does not meet the rules.'; return; }
  if (pw.value !== pw2.value) { err.value = 'The new passwords do not match.'; return; }
  busy.value = true;
  try {
    await be.post('/auth/password', { current_password: current.value, new_password: pw.value });
  } catch (e) {
    err.value = (e && e.message) || 'Could not change the password.';
    busy.value = false;
    return;
  }
  // the session that held the temporary password ends; sign in again with the new one
  await be.logout();
  busy.value = false;
  await navigateTo('/login?reset=1');
};
const field = 'height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid var(--k-C9D6CF,#C9D6CF); background: #fff; color: var(--k-0F1714,#0F1714); font-size: 15px';
const label = 'display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-55645C,#55645C)';
</script>
<template>
  <form novalidate aria-label="Choose a new password" @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
    <h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Choose a new password</h1>
    <p style="margin: 0; color: var(--k-55645C,#55645C); font-size: 14px">You signed in with a temporary password. Choose your own before you continue.</p>
    <label :style="label">Temporary password<input v-model="current" :type="show ? 'text' : 'password'" autocomplete="current-password" aria-label="Temporary password" :style="field"></label>
    <label :style="label">New password<input v-model="pw" :type="show ? 'text' : 'password'" autocomplete="new-password" aria-label="New password" :style="field"></label>
    <label :style="label">Confirm new password<input v-model="pw2" :type="show ? 'text' : 'password'" autocomplete="new-password" aria-label="Confirm new password" :style="field"></label>
    <label style="display: flex; gap: 8px; align-items: center; font-size: 13px"><input v-model="show" type="checkbox"> Show passwords</label>
    <ul style="margin: 0; padding: 0; list-style: none; display: flex; flex-direction: column; gap: 2px; font-size: 12px"><li v-for="r in rules" :key="r[0]" :style="`color: ${r[1] ? '#2B7A12' : 'var(--k-8A988F,#8A988F)'}`">{{ r[1] ? '\u2713' : '\u25CB' }} {{ r[0] }}</li></ul>
    <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
    <button type="submit" :disabled="busy" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">{{ busy ? 'Saving…' : 'Save new password' }}</button>
    <button type="button" @click="be.logout().then(() => navigateTo('/login'))" style="border: 0; background: transparent; color: var(--k-55645C,#55645C); cursor: pointer; font-size: 13px">Sign out</button>
  </form>
</template>
