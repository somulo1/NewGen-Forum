<template>
  <div class="roles-widget">
    <div class="widget-header">
      <h3>Active Roles</h3>
      <NuxtLink to="/admin/roles" class="view-all">Manage →</NuxtLink>
    </div>
    <div class="roles-list">
      <div v-for="role in roles" :key="role.id" class="role-item">
        <div class="role-icon">
          <Icon name="lock" size="16" />
        </div>
        <div class="role-info">
          <h4>{{ role.name }}</h4>
          <p>{{ role.members }} members</p>
        </div>
        <span class="role-badge">{{ role.permissions }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

const roles = [
  { id: 1, name: 'Administrator', members: 3, permissions: '∞' },
  { id: 2, name: 'Manager', members: 12, permissions: '15' },
  { id: 3, name: 'Team Lead', members: 8, permissions: '12' },
]
</script>

<style scoped>
.roles-widget {
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

.roles-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.role-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.role-item:last-child {
  border-bottom: none;
}

.role-icon {
  width: 32px;
  height: 32px;
  background: rgba(201, 182, 255, 0.2);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-lavender);
  flex-shrink: 0;
}

.role-info {
  flex: 1;
}

.role-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.role-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.role-badge {
  background: var(--color-border);
  color: var(--color-text-secondary);
  padding: 4px 8px;
  border-radius: var(--radius-xs);
  font-size: 11px;
  font-weight: 600;
}
</style>
