<!-- Layout for login, sign up and password reset: no sidebar, one centered card -->
<template>
  <div style="min-height: 100vh; box-sizing: border-box; padding: 24px 16px; background: #F2F6F3; color: #0F1714; font-family: 'DM Sans', system-ui, sans-serif; font-size: 14px; display: flex; align-items: center; justify-content: center">
    <div style="width: 420px; max-width: 100%; display: flex; flex-direction: column; gap: 20px; align-items: stretch">
      <div style="display: flex; align-items: center; gap: 10px; justify-content: center">
        <span aria-hidden="true" style="width: 40px; height: 40px; border-radius: 12px; background: #B8F36B; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 20px">H</span>
        <span style="font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 22px">Hub</span>
      </div>
      <main style="padding: clamp(22px, 5vw, 32px); border-radius: 24px; background: #FFFFFF; border: 1px solid #DCE5DF; box-shadow: 0 8px 30px rgba(15,23,20,0.08); display: flex; flex-direction: column; gap: 18px">
        <slot />
      </main>
      <p style="margin: 0; text-align: center; font-size: 12px; color: #6B7A72"><NuxtLink to="/terms" target="_blank" style="color: inherit">Terms</NuxtLink> · <NuxtLink to="/privacy" target="_blank" style="color: inherit">Privacy</NuxtLink></p>
    </div>
  </div>
</template>
