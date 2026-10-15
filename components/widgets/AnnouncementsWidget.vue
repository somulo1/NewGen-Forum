<template>
  <div class="announcements-widget">
    <div class="widget-header">
      <h3>Latest Announcements</h3>
      <NuxtLink to="/announcements" class="view-all">View all →</NuxtLink>
    </div>
    <div class="announcements-list">
      <div v-for="announcement in announcements" :key="announcement.id" class="announcement-item">
        <div class="announcement-badge" :class="`priority-${announcement.priority}`"></div>
        <div class="announcement-info">
          <h4>{{ announcement.title }}</h4>
          <p>{{ announcement.author }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const announcements = [
  { id: 1, title: 'Holiday Schedule Announcement', author: 'HR Department', priority: 'high' },
  { id: 2, title: 'New Office Facilities Opening', author: 'Administration', priority: 'medium' },
  { id: 3, title: 'Security Update Reminder', author: 'IT Operations', priority: 'high' },
]
</script>

<style scoped>
.announcements-widget {
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

.announcements-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.announcement-item {
  display: flex;
  gap: var(--spacing-md);
  align-items: flex-start;
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.announcement-item:last-child {
  border-bottom: none;
}

.announcement-badge {
  width: 3px;
  height: 3px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-top: 6px;
}

.priority-high {
  background: var(--color-coral);
}

.priority-medium {
  background: var(--color-amber);
}

.announcement-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.announcement-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
