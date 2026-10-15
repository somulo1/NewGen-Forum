<template>
  <div class="tasks-widget">
    <div class="widget-header">
      <h3>My Tasks</h3>
      <button class="add-task" title="Add task">+</button>
    </div>
    <div class="tasks-list">
      <div v-for="task in tasks" :key="task.id" class="task-item">
        <input type="checkbox" :checked="task.completed" class="task-checkbox" />
        <div class="task-info" :class="{ completed: task.completed }">
          <h4>{{ task.title }}</h4>
          <p>{{ task.dueDate }}</p>
        </div>
        <span class="task-priority" :class="`priority-${task.priority}`">•</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const tasks = [
  { id: 1, title: 'Complete project documentation', dueDate: 'Due Dec 20', completed: false, priority: 'high' },
  { id: 2, title: 'Review pull requests', dueDate: 'Due Dec 18', completed: false, priority: 'medium' },
  { id: 3, title: 'Update team members', dueDate: 'Due Dec 22', completed: true, priority: 'low' },
]
</script>

<style scoped>
.tasks-widget {
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

.add-task {
  background: var(--color-lime);
  color: #000;
  border: none;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  font-weight: 600;
  transition: opacity 0.2s ease;
}

.add-task:hover {
  opacity: 0.8;
}

.tasks-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.task-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.task-item:last-child {
  border-bottom: none;
}

.task-checkbox {
  width: 16px;
  height: 16px;
  cursor: pointer;
  flex-shrink: 0;
}

.task-info {
  flex: 1;
}

.task-info.completed h4 {
  text-decoration: line-through;
  color: var(--color-text-secondary);
}

.task-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.task-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.task-priority {
  font-size: 16px;
  margin-left: auto;
}

.priority-high {
  color: var(--color-coral);
}

.priority-medium {
  color: var(--color-amber);
}

.priority-low {
  color: var(--color-sky);
}
</style>
