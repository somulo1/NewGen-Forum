<template>
  <div class="tickets-widget">
    <div class="widget-header">
      <h3>Tickets</h3>
      <NuxtLink to="/tickets" class="view-all">View all →</NuxtLink>
    </div>
    <div class="tickets-list">
      <div v-for="ticket in tickets" :key="ticket.id" class="ticket-item">
        <div class="ticket-id">{{ ticket.id }}</div>
        <div class="ticket-info">
          <h4>{{ ticket.title }}</h4>
          <p>{{ ticket.status }}</p>
        </div>
        <span class="priority-badge" :class="`priority-${ticket.priority}`">{{ ticket.priority }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const tickets = [
  { id: 'TK-101', title: 'Fix login issue', status: 'In Progress', priority: 'high' },
  { id: 'TK-102', title: 'Update documentation', status: 'Open', priority: 'medium' },
  { id: 'TK-103', title: 'Performance optimization', status: 'In Progress', priority: 'medium' },
]
</script>

<style scoped>
.tickets-widget {
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

.tickets-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.ticket-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.ticket-item:last-child {
  border-bottom: none;
}

.ticket-id {
  font-family: monospace;
  font-size: 12px;
  font-weight: 600;
  color: var(--color-lime);
  min-width: 60px;
}

.ticket-info {
  flex: 1;
}

.ticket-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.ticket-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.priority-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: var(--radius-xs);
  white-space: nowrap;
}

.priority-high {
  background: rgba(255, 155, 138, 0.2);
  color: var(--color-coral);
}

.priority-medium {
  background: rgba(255, 214, 107, 0.2);
  color: var(--color-amber);
}
</style>
