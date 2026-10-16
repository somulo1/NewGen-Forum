<script setup>
definePageMeta({ path: '/signup', layout: 'auth' });
useHead({ title: 'Create account · Hub' });
const name = ref('');
const email = ref('');
const password = ref('');
const show = ref(false);
const submit = () => { useAuth().signIn(name.value.trim(), email.value.trim()); navigateTo('/'); };
</script>
<template>
  <form novalidate aria-label="Create account" @submit.prevent="submit" style="display: flex; flex-direction: column; gap: 14px">
    <div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Create your account</h1><span style="color: #55645C">It takes a few seconds.</span></div>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Full name<input v-model="name" type="text" autocomplete="name" placeholder="Amani Kariuki" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"></label>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Email<input v-model="email" type="email" autocomplete="email" placeholder="you@company.com" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"></label>
    <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #55645C">Password
      <span style="position: relative; display: block"><input v-model="password" :type="show ? 'text' : 'password'" autocomplete="new-password" placeholder="Choose a password" style="width: 100%; box-sizing: border-box; height: 44px; padding: 0 64px 0 14px; border-radius: 12px; border: 1px solid #C4D0C9; background: #FFFFFF; color: #0F1714; font-size: 15px"><button type="button" :aria-label="show ? 'Hide password' : 'Show password'" @click="show = !show" style="position: absolute; right: 6px; top: 6px; height: 32px; padding: 0 10px; border: 0; border-radius: 8px; background: transparent; color: #55645C; cursor: pointer; font-size: 12px">{{ show ? 'Hide' : 'Show' }}</button></span>
    </label>
    <button type="submit" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">Create account</button>
    <span style="text-align: center; color: #55645C">Already have an account? <NuxtLink to="/login" style="color: #2F7A12; font-weight: 600">Sign in</NuxtLink></span>
  </form>
</template>
