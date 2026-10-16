<script setup>
definePageMeta({ path: '/login', layout: 'auth' });
useHead({ title: 'Sign in · Hub' });
const route = useRoute();
const email = ref('');
const password = ref('');
const show = ref(false);
const submit = () => {
  useAuth().signIn('', email.value.trim());
  const next = typeof route.query.next === 'string' && route.query.next.startsWith('/') ? route.query.next : '/';
  navigateTo(next);
};
</script>
<template>
  <div v-if="route.query.reset" role="status" style="padding: 10px 12px; border-radius: 10px; background: #EAF6DD; color: #2B5A0B; font-size: 13px; margin-bottom: 14px">Password updated. Sign in with your new password.</div>
  <form novalidate aria-label="Sign in" @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
    <div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Welcome back</h1></div>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Email<input v-model="email" type="email" autocomplete="username" placeholder="you@company.com" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"></label>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Password
      <span style="position: relative; display: block"><input v-model="password" :type="show ? 'text' : 'password'" autocomplete="current-password" placeholder="Your password" style="width: 100%; box-sizing: border-box; height: 44px; padding: 0 64px 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"><button type="button" :aria-label="show ? 'Hide password' : 'Show password'" @click="show = !show" style="position: absolute; right: 6px; top: 6px; height: 32px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: #55645C; cursor: pointer; font-size: 12px">{{ show ? 'Hide' : 'Show' }}</button></span>
    </label>
    <div style="text-align: right"><NuxtLink to="/reset-password" style="color: #2F7A12; font-size: 13px">Forgot password?</NuxtLink></div>
    <button type="submit" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">Sign in</button>
    <span style="text-align: center; color: #55645C">New here? <NuxtLink to="/signup" style="color: #2F7A12; font-weight: 600">Create an account</NuxtLink></span>
  </form>
</template>
