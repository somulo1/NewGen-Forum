<script setup>
definePageMeta({ path: '/signup', layout: 'auth' });
useHead({ title: 'Create account · Hub' });
const strict = AUTH_STRICT;
const name = ref('');
const email = ref('');
const password = ref('');
const show = ref(false);
const questionList = SECURITY_QUESTIONS;
const qs = ref([SECURITY_QUESTIONS[0], SECURITY_QUESTIONS[1], SECURITY_QUESTIONS[2]]);
const ans = ref(['', '', '']);
const err = ref('');
const submit = async () => {
  err.value = '';
  const e = email.value.trim();
  if (strict) {
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(e)) { err.value = 'Enter a valid email address.'; return; }
    if (password.value.length < 8) { err.value = 'Use a password of at least 8 characters.'; return; }
    if (new Set(qs.value).size < 3) { err.value = 'Choose three different security questions.'; return; }
    if (ans.value.some((a) => a.trim().length < 2)) { err.value = 'Answer all three security questions.'; return; }
  }
  await useAccounts().create(name.value.trim(), e, password.value, qs.value.map((q, i) => ({ q, a: ans.value[i] })));
  useAuth().signIn(name.value.trim(), e);
  navigateTo('/');
};
</script>
<template>
  <form novalidate aria-label="Create account" @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
    <div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Create your account</h1></div>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Full name<input v-model="name" type="text" autocomplete="name" placeholder="Amani Kariuki" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"></label>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Email<input v-model="email" type="email" autocomplete="email" placeholder="you@company.com" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"></label>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Password
      <span style="position: relative; display: block"><input v-model="password" :type="show ? 'text' : 'password'" autocomplete="new-password" placeholder="Choose a password" style="width: 100%; box-sizing: border-box; height: 44px; padding: 0 64px 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"><button type="button" :aria-label="show ? 'Hide password' : 'Show password'" @click="show = !show" style="position: absolute; right: 6px; top: 6px; height: 32px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: #55645C; cursor: pointer; font-size: 12px">{{ show ? 'Hide' : 'Show' }}</button></span>
    </label>
    <fieldset style="margin: 0; padding: 12px; border-radius: 14px; border: 1px solid #DCE5DF; display: flex; flex-direction: column; gap: 10px; min-width: 0">
      <legend style="padding: 0 6px; font-size: 13px; font-weight: 700">Security questions</legend>
      <template v-for="i in 3" :key="i">
        <select v-model="qs[i - 1]" :aria-label="`Security question ${i}`" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px">
          <option v-for="q in questionList" :key="q" :value="q">{{ q }}</option>
        </select>
        <input v-model="ans[i - 1]" type="text" autocomplete="off" :aria-label="`Answer ${i}`" placeholder="Your answer" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px">
      </template>
    </fieldset>
    <span v-if="err" role="alert" style="color: #C0392B; font-size: 13px">{{ err }}</span>
    <button type="submit" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">Create account</button>
    <span style="text-align: center; color: #55645C">Already have an account? <NuxtLink to="/login" style="color: #2F7A12; font-weight: 600">Sign in</NuxtLink></span>
  </form>
</template>
