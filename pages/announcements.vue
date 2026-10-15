<template>
  <div class="announcements-container">
    <div class="page-header">
      <div>
        <h1>Announcements</h1>
        <p class="subtitle">Organization News & Updates</p>
      </div>
      <button class="btn btn-primary">
        <Icon name="edit" size="18" />
        New Announcement
      </button>
    </div>

    <div class="announcements-content">
      <div class="filters">
        <button
          v-for="priority in ['all', 'high', 'medium', 'low']"
          :key="priority"
          :class="['filter-btn', { active: activePriority === priority }]"
          @click="activePriority = priority"
        >
          {{ priority === 'all' ? 'All' : priority.charAt(0).toUpperCase() + priority.slice(1) }}
        </button>
      </div>

      <div class="announcements-list">
        <article v-for="announcement in filteredAnnouncements" :key="announcement.id" class="announcement-card">
          <div class="announcement-header">
            <div>
              <h2>{{ announcement.title }}</h2>
              <p class="announcement-meta">
                {{ announcement.author }} • {{ announcement.date }}
              </p>
            </div>
            <span class="priority-badge" :class="`priority-${announcement.priority}`">
              {{ announcement.priority.charAt(0).toUpperCase() + announcement.priority.slice(1) }}
            </span>
          </div>
          <p class="announcement-body">{{ announcement.content }}</p>
          <div class="announcement-footer">
            <button class="read-btn">Read More</button>
            <span class="engagement">{{ announcement.likes }} likes • {{ announcement.comments }} comments</span>
          </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const activePriority = ref('all')

const announcements = [
  {
    id: 1,
    title: 'Holiday Schedule Announcement',
    author: 'HR Department',
    date: 'Dec 15, 2024',
    priority: 'high',
    content: 'We are pleased to announce the holiday schedule for the upcoming year. The office will be closed on the following dates: December 25-26, January 1, and additional dates as per company policy. Please plan your work accordingly.',
    likes: 45,
    comments: 12,
  },
  {
    id: 2,
    title: 'New Office Facilities Opening',
    author: 'Administration',
    date: 'Dec 14, 2024',
    priority: 'medium',
    content: 'Exciting news! Our new office facility in the downtown area is now open. The facility features modern workspaces, collaboration areas, a fully-equipped cafeteria, and wellness rooms. All employees are welcome to use the new space.',
    likes: 67,
    comments: 28,
  },
  {
    id: 3,
    title: 'Security Update Reminder',
    author: 'IT Operations',
    date: 'Dec 12, 2024',
    priority: 'high',
    content: 'A critical security update is required for all systems. Please update your devices by December 20, 2024. This update includes important patches for vulnerability prevention. Instructions are available on the IT portal.',
    likes: 23,
    comments: 8,
  },
  {
    id: 4,
    title: 'Q4 Performance Review Schedule',
    author: 'HR Team',
    date: 'Dec 10, 2024',
    priority: 'medium',
    content: 'Q4 performance reviews will be conducted between December 18 and January 15, 2025. Please coordinate with your managers to schedule your review session. All employees are expected to complete a self-assessment.',
    likes: 34,
    comments: 15,
  },
  {
    id: 5,
    title: 'New Employee Handbook Available',
    author: 'HR Department',
    date: 'Dec 8, 2024',
    priority: 'low',
    content: 'The updated Employee Handbook for 2025 is now available for download. It includes updated policies, benefits information, and company guidelines. Please review it carefully.',
    likes: 18,
    comments: 5,
  },
]

const filteredAnnouncements = computed(() => {
  if (activePriority.value === 'all') {
    return announcements
  }
  return announcements.filter(a => a.priority === activePriority.value)
})
</script>

<style scoped>
.announcements-container {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
  color: var(--color-text);
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: var(--spacing-sm) 0 0 0;
}

.announcements-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.filters {
  display: flex;
  gap: var(--spacing-md);
}

.filter-btn {
  padding: var(--spacing-sm) var(--spacing-lg);
  border: 1px solid var(--color-border);
  background: var(--color-surface);
  color: var(--color-text-secondary);
  border-radius: var(--radius-lg);
  cursor: pointer;
  font-weight: 500;
  font-size: 13px;
  transition: all 0.2s ease;
}

.filter-btn:hover {
  border-color: var(--color-lime);
  color: var(--color-text);
}

.filter-btn.active {
  background: var(--color-lime);
  color: #000;
  border-color: var(--color-lime);
}

.announcements-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.announcement-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
  transition: all 0.2s ease;
}

.announcement-card:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 16px rgba(184, 243, 107, 0.1);
}

.announcement-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
}

.announcement-header h2 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 20px;
  font-weight: 600;
  color: var(--color-text);
}

.announcement-meta {
  margin: 0;
  font-size: 13px;
  color: var(--color-text-secondary);
}

.priority-badge {
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
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

.priority-low {
  background: rgba(124, 196, 255, 0.2);
  color: var(--color-sky);
}

.announcement-body {
  margin: 0 0 var(--spacing-lg) 0;
  font-size: 15px;
  line-height: 1.6;
  color: var(--color-text);
}

.announcement-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
}

.read-btn {
  background: none;
  border: none;
  color: var(--color-lime);
  cursor: pointer;
  font-weight: 600;
  font-size: 13px;
  transition: opacity 0.2s ease;
}

.read-btn:hover {
  opacity: 0.8;
}

.engagement {
  font-size: 12px;
  color: var(--color-text-secondary);
}

@media (max-width: 768px) {
  .announcement-header {
    flex-direction: column;
    gap: var(--spacing-md);
  }

  .priority-badge {
    align-self: flex-start;
  }
}
</style>
