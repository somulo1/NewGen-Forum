<!-- Last resort: errors Nuxt itself raises outside the app layout. -->
<script setup lang="ts">
const props = defineProps<{ error: { statusCode?: number } }>();
const code = props.error && props.error.statusCode === 404 ? '404' : props.error && props.error.statusCode === 403 ? '403' : '500';
const leave = () => clearError();
</script>
<template>
  <div style="min-height: 100vh; background: #0F1714; color: #E8EFEA; font-family: 'DM Sans', system-ui, sans-serif; display: flex; align-items: center; padding: 16px">
    <ErrorPage :code="code" @leave="leave" />
  </div>
</template>
