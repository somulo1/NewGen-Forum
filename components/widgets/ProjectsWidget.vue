<template>
  <div class="projects-widget">
    <div class="widget-header">
      <h3>Active Projects</h3>
      <NuxtLink to="/projects" class="view-all">View all →</NuxtLink>
    </div>
    <div class="projects-list">
      <div v-for="project in projects" :key="project.id" class="project-item">
        <div class="project-info">
          <h4>{{ project.name }}</h4>
          <div class="project-meta">
            <span>{{ project.tasks }} tasks</span>
            <span>{{ project.progress }}% done</span>
          </div>
        </div>
        <div class="progress-bar">
          <div class="progress-fill" :style="{ width: project.progress + '%' }"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const projects = [
  { id: 1, name: 'Website Redesign', tasks: 12, progress: 75 },
  { id: 2, name: 'Mobile App Development', tasks: 24, progress: 45 },
  { id: 3, name: 'API Integration', tasks: 8, progress: 90 },
]
</script>

<style scoped>
.projects-widget {
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

.projects-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.project-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.project-item:last-child {
  border-bottom: none;
}

.project-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.project-meta {
  display: flex;
  gap: var(--spacing-lg);
  font-size: 12px;
  color: var(--color-text-secondary);
}

.progress-bar {
  width: 100%;
  height: 6px;
  background: var(--color-border);
  border-radius: var(--radius-xs);
  overflow: hidden;
}

.progress-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-lime) 0%, var(--color-teal) 100%);
  border-radius: var(--radius-xs);
  transition: width 0.3s ease;
}
</style>
