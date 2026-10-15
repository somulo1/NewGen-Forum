<template>
  <div class="people-widget">
    <div class="widget-header">
      <h3>Team Members</h3>
      <NuxtLink to="/people" class="view-all">View all →</NuxtLink>
    </div>
    <div class="people-list">
      <div v-for="person in people" :key="person.id" class="person-item">
        <div class="person-avatar">{{ person.initials }}</div>
        <div class="person-info">
          <h4>{{ person.name }}</h4>
          <p>{{ person.role }}</p>
        </div>
        <span class="status-badge" :class="`status-${person.status}`"></span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const people = [
  { id: 1, name: 'John Doe', initials: 'JD', role: 'Senior Developer', status: 'online' },
  { id: 2, name: 'Sarah Martin', initials: 'SM', role: 'Product Manager', status: 'online' },
  { id: 3, name: 'Marcus Davis', initials: 'MD', role: 'Designer', status: 'away' },
]
</script>

<style scoped>
.people-widget {
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

.people-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.person-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
  position: relative;
}

.person-item:last-child {
  border-bottom: none;
}

.person-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.person-info {
  flex: 1;
}

.person-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.person-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.status-badge {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}

.status-online {
  background: var(--color-lime);
}

.status-away {
  background: var(--color-amber);
}

.status-offline {
  background: var(--color-text-secondary);
}
</style>
