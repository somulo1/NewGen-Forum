<script setup>
definePageMeta({ path: '/reset-password', layout: 'auth' });
useHead({ title: 'Reset password · Hub' });
const email = ref('');
const sent = ref(false);
const password = ref('');
const done = ref(false);
const send = () => { sent.value = true; };
// Demo: the "link in your email" is the next step on this same page.
const save = () => { done.value = true; };
</script>
<template>
  <div style="display: flex; flex-direction: column; gap: 14px">
    <template v-if="!sent">
      <form novalidate aria-label="Reset password" @submit.prevent="send" style="display: flex; flex-direction: column; gap: 14px">
        <div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Reset your password</h1><span style="color: #9DB0A6">Enter your email and we'll send you a link.</span></div>
        <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #9DB0A6">Email<input v-model="email" type="email" autocomplete="email" placeholder="you@company.com" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 15px"></label>
        <button type="submit" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">Send reset link</button>
      </form>
    </template>
    <template v-if="sent && !done">
      <form novalidate aria-label="Choose a new password" @submit.prevent="save" style="display: flex; flex-direction: column; gap: 14px">
        <div style="display: flex; flex-direction: column; gap: 4px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">Check your email</h1><span style="color: #9DB0A6">We sent a link to <strong style="color: #E8EFEA">{{ email || 'your email' }}</strong>. Opening it brings you here to choose a new password.</span></div>
        <label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #9DB0A6">New password<input v-model="password" type="password" autocomplete="new-password" placeholder="Choose a new password" style="height: 44px; padding: 0 14px; border-radius: 12px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 15px"></label>
        <button type="submit" style="height: 46px; border-radius: 23px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; cursor: pointer">Save new password</button>
      </form>
    </template>
    <template v-if="done">
      <div role="status" style="display: flex; flex-direction: column; gap: 14px; align-items: center; text-align: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px">Password updated</h1><span style="color: #9DB0A6">You can sign in with your new password.</span><NuxtLink to="/login" style="width: 100%; box-sizing: border-box; height: 46px; border-radius: 23px; background: #B8F36B; color: #0F1714; font-weight: 700; font-size: 15px; display: flex; align-items: center; justify-content: center; text-decoration: none">Back to sign in</NuxtLink></div>
    </template>
    <span v-if="!done" style="text-align: center"><NuxtLink to="/login" style="color: #B8F36B; font-weight: 600">Back to sign in</NuxtLink></span>
  </div>
</template>
