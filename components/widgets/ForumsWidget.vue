<template>
  <div class="forums-widget">
    <div class="widget-header">
      <h3>Recent Forums</h3>
      <NuxtLink to="/forums" class="view-all">View all →</NuxtLink>
    </div>
    <div class="forum-list">
      <div v-for="forum in forums" :key="forum.id" class="forum-item">
        <div class="forum-info">
          <h4>{{ forum.title }}</h4>
          <p>{{ forum.replies }} replies</p>
        </div>
        <span class="forum-badge" :class="`badge-${forum.type}`">{{ forum.type }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const forums = [
  { id: 1, title: 'Q4 Planning Discussion', replies: 12, type: 'important' },
  { id: 2, title: 'Remote Work Policy Updates', replies: 8, type: 'general' },
  { id: 3, title: 'Team Building Event Ideas', replies: 24, type: 'general' },
]
</script>

<style scoped>
.forums-widget {
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

.forum-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.forum-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.forum-item:last-child {
  border-bottom: none;
}

.forum-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.forum-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.forum-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: var(--radius-xs);
}

.badge-important {
  background: rgba(255, 155, 138, 0.2);
  color: var(--color-coral);
}

.badge-general {
  background: rgba(124, 196, 255, 0.2);
  color: var(--color-sky);
}
</style>
