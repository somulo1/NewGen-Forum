<template>
  <div class="hub-root">
    <HubLayout>
      <template #sidebar>
        <HubSidebar />
      </template>

      <template #header>
        <HubHeader />
      </template>

      <template #content>
        <NuxtPage />
      </template>
    </HubLayout>
  </div>
</template>

<script setup lang="ts">
import { useTheme } from '~/composables/useTheme'

const { initTheme } = useTheme()

onMounted(() => {
  initTheme()
})
</script>

<style scoped>
.hub-root {
  width: 100%;
  min-height: 100vh;
}
</style>
