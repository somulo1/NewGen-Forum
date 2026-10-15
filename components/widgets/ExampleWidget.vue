<template>
  <div class="widget-container">
    <div class="widget-header">
      <h3>{{ title }}</h3>
      <button v-if="showActions" class="widget-action" @click="showMenu = !showMenu">
        <Icon name="more-vertical" size="16" />
      </button>
    </div>

    <div class="widget-body">
      <slot>
        <p class="widget-placeholder">Widget content goes here</p>
      </slot>
    </div>

    <div v-if="showActions" class="widget-footer">
      <NuxtLink :to="actionLink" class="widget-link">
        View all →
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

interface Props {
  title: string
  actionLink?: string
  showActions?: boolean
  config?: Record<string, any>
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Widget',
  actionLink: '/',
  showActions: true,
  config: () => ({}),
})

const showMenu = ref(false)
</script>

<style scoped>
.widget-container {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  height: 100%;
  transition: all 0.2s ease;
}

.widget-container:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.widget-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.widget-action {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.widget-action:hover {
  background: var(--color-border);
  color: var(--color-text);
}

.widget-body {
  flex: 1;
  padding: var(--spacing-lg);
  overflow-y: auto;
}

.widget-placeholder {
  margin: 0;
  color: var(--color-text-secondary);
  text-align: center;
  font-size: 14px;
}

.widget-footer {
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg);
}

.widget-link {
  color: var(--color-lime);
  text-decoration: none;
  font-size: 13px;
  font-weight: 600;
  transition: opacity 0.2s ease;
}

.widget-link:hover {
  opacity: 0.8;
}
</style>
