<template>
  <div class="files-widget">
    <div class="widget-header">
      <h3>Recent Files</h3>
      <NuxtLink to="/files" class="view-all">View all →</NuxtLink>
    </div>
    <div class="files-list">
      <div v-for="file in files" :key="file.id" class="file-item">
        <div class="file-icon">
          <Icon :name="file.icon" size="16" />
        </div>
        <div class="file-info">
          <h4>{{ file.name }}</h4>
          <p>{{ file.size }} • {{ file.modified }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

const files = [
  { id: 1, name: 'Q4 Report.pdf', size: '2.4 MB', modified: '2 days ago', icon: 'file' },
  { id: 2, name: 'Design System.figma', size: '18 MB', modified: '1 week ago', icon: 'file' },
  { id: 3, name: 'Budget 2025.xlsx', size: '156 KB', modified: '3 days ago', icon: 'file' },
]
</script>

<style scoped>
.files-widget {
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

.files-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.file-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.file-item:last-child {
  border-bottom: none;
}

.file-icon {
  width: 32px;
  height: 32px;
  background: rgba(184, 243, 107, 0.1);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-lime);
  flex-shrink: 0;
}

.file-info {
  flex: 1;
}

.file-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.file-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}
</style>
