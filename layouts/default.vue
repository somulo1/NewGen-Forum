<template>
  <HubLayout>
    <template #sidebar>
      <HubSidebar />
    </template>

    <template #header>
      <HubHeader />
    </template>

    <template #content>
      <slot />
    </template>
  </HubLayout>
</template>

<script setup lang="ts">
</script>
