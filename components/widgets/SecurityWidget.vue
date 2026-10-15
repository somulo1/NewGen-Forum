<template>
  <div class="security-widget">
    <div class="widget-header">
      <h3>Security Alerts</h3>
      <NuxtLink to="/admin/security" class="view-all">View all →</NuxtLink>
    </div>
    <div class="alerts-list">
      <div v-for="alert in alerts" :key="alert.id" class="alert-item" :class="`severity-${alert.severity}`">
        <Icon :name="alert.icon" size="16" />
        <div class="alert-info">
          <h4>{{ alert.title }}</h4>
          <p>{{ alert.message }}</p>
        </div>
        <span class="alert-time">{{ alert.time }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

const alerts = [
  { id: 1, title: 'Failed Login Attempt', message: 'Multiple failed attempts from IP 192.168.x.x', icon: 'shield', time: '2 hours ago', severity: 'warning' },
  { id: 2, title: 'Security Update Available', message: 'Critical patch available for system', icon: 'shield', time: '1 day ago', severity: 'info' },
  { id: 3, title: 'Permission Change', message: 'Admin role added to user account', icon: 'lock', time: '3 days ago', severity: 'info' },
]
</script>

<style scoped>
.security-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.widget-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.view-all {
  color: var(--color-lime);
  text-decoration: none;
  font-size: 12px;
  font-weight: 600;
}

.alerts-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.alert-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  border-radius: var(--radius-md);
  border-left: 3px solid transparent;
}

.severity-warning {
  background: rgba(255, 214, 107, 0.1);
  border-left-color: var(--color-amber);
  color: var(--color-amber);
}

.severity-info {
  background: rgba(124, 196, 255, 0.1);
  border-left-color: var(--color-sky);
  color: var(--color-sky);
}

.alert-info {
  flex: 1;
}

.alert-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.alert-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.alert-time {
  font-size: 11px;
  color: var(--color-text-secondary);
  white-space: nowrap;
}
</style>
