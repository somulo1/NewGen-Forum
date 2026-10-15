<template>
  <div class="forums-container">
    <div class="page-header">
      <div>
        <h1>Forums</h1>
        <p class="subtitle">Discussion & Collaboration</p>
      </div>
      <button class="btn btn-primary">
        <Icon name="message-square" size="18" />
        New Discussion
      </button>
    </div>

    <div class="forums-layout">
      <aside class="forums-sidebar">
        <div class="search-box">
          <Icon name="search" size="18" />
          <input
            v-model="searchQuery"
            type="text"
            placeholder="Search forums..."
            class="search-input"
          />
        </div>

        <div class="categories">
          <h3 class="category-title">Categories</h3>
          <button
            v-for="cat in categories"
            :key="cat.id"
            :class="['category-btn', { active: activeCategory === cat.id }]"
            @click="activeCategory = cat.id"
          >
            {{ cat.name }}
            <span class="count">{{ cat.count }}</span>
          </button>
        </div>
      </aside>

      <main class="forums-main">
        <div class="forums-grid">
          <div v-for="forum in filteredForums" :key="forum.id" class="forum-card">
            <div class="forum-badge" :class="`priority-${forum.priority}`"></div>
            <div class="forum-content">
              <h3>{{ forum.title }}</h3>
              <p class="forum-description">{{ forum.description }}</p>
              <div class="forum-meta">
                <span class="meta-item">
                  <Icon name="message-square" size="14" />
                  {{ forum.replies }} replies
                </span>
                <span class="meta-item">
                  <Icon name="users" size="14" />
                  {{ forum.views }} views
                </span>
                <span class="meta-item">
                  Last: {{ forum.lastActivity }}
                </span>
              </div>
            </div>
            <div class="forum-author">
              <div class="author-avatar">{{ forum.authorInitials }}</div>
              <div class="author-info">
                <p class="author-name">{{ forum.author }}</p>
                <p class="author-date">{{ forum.date }}</p>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const searchQuery = ref('')
const activeCategory = ref('all')

const categories = [
  { id: 'all', name: 'All Forums', count: 47 },
  { id: 'general', name: 'General', count: 12 },
  { id: 'technical', name: 'Technical', count: 15 },
  { id: 'announcements', name: 'Announcements', count: 8 },
  { id: 'feedback', name: 'Feedback', count: 12 },
]

const forums = [
  {
    id: 1,
    title: 'Q4 Planning & Strategy Discussion',
    description: 'Discussion about our Q4 goals and strategic initiatives',
    category: 'general',
    priority: 'high',
    replies: 24,
    views: 156,
    lastActivity: '2 hours ago',
    author: 'Sarah Martin',
    authorInitials: 'SM',
    date: 'Dec 15, 2024',
  },
  {
    id: 2,
    title: 'Remote Work Policy Updates',
    description: 'New updates to our flexible working policies',
    category: 'general',
    priority: 'medium',
    replies: 18,
    views: 98,
    lastActivity: '4 hours ago',
    author: 'HR Team',
    authorInitials: 'HR',
    date: 'Dec 14, 2024',
  },
  {
    id: 3,
    title: 'New API Documentation',
    description: 'Technical documentation for our latest API updates',
    category: 'technical',
    priority: 'high',
    replies: 12,
    views: 234,
    lastActivity: '30 mins ago',
    author: 'Dev Team',
    authorInitials: 'DT',
    date: 'Dec 15, 2024',
  },
  {
    id: 4,
    title: 'Best Practices: Code Review Process',
    description: 'Share and discuss our code review standards',
    category: 'technical',
    priority: 'medium',
    replies: 8,
    views: 145,
    lastActivity: '1 day ago',
    author: 'John Doe',
    authorInitials: 'JD',
    date: 'Dec 13, 2024',
  },
  {
    id: 5,
    title: 'Holiday Schedule & Time Off',
    description: 'Information about holiday schedules and time off requests',
    category: 'announcements',
    priority: 'high',
    replies: 6,
    views: 67,
    lastActivity: '3 days ago',
    author: 'Management',
    authorInitials: 'MG',
    date: 'Dec 12, 2024',
  },
]

const filteredForums = computed(() => {
  return forums.filter((forum) => {
    const matchesCategory = activeCategory.value === 'all' || forum.category === activeCategory.value
    const matchesSearch = forum.title.toLowerCase().includes(searchQuery.value.toLowerCase()) ||
      forum.description.toLowerCase().includes(searchQuery.value.toLowerCase())
    return matchesCategory && matchesSearch
  })
})
</script>

<style scoped>
.forums-container {
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

.forums-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: var(--spacing-lg);
}

.forums-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.search-box {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  color: var(--color-text-secondary);
}

.search-input {
  flex: 1;
  background: none;
  border: none;
  color: var(--color-text);
  outline: none;
  font-family: 'DM Sans', system-ui, sans-serif;
  font-size: 14px;
}

.search-input::placeholder {
  color: var(--color-text-secondary);
}

.categories {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.category-title {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.category-btn {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--color-bg);
  border: none;
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s ease;
}

.category-btn:hover {
  background: var(--color-border);
  color: var(--color-text);
}

.category-btn.active {
  background: rgba(184, 243, 107, 0.1);
  color: var(--color-lime);
  font-weight: 600;
}

.count {
  background: var(--color-border);
  color: var(--color-text-secondary);
  padding: 2px 6px;
  border-radius: 12px;
  font-size: 11px;
  font-weight: 600;
}

.forums-main {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.forums-grid {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.forum-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  padding: var(--spacing-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  transition: all 0.2s ease;
  cursor: pointer;
  position: relative;
}

.forum-card:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.forum-badge {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  border-radius: var(--radius-lg) 0 0 var(--radius-lg);
}

.priority-high {
  background: var(--color-coral);
}

.priority-medium {
  background: var(--color-amber);
}

.forum-content {
  flex: 1;
  padding-left: var(--spacing-md);
}

.forum-content h3 {
  margin: 0 0 var(--spacing-sm) 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.forum-description {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 14px;
  color: var(--color-text-secondary);
}

.forum-meta {
  display: flex;
  gap: var(--spacing-lg);
  font-size: 12px;
  color: var(--color-text-secondary);
}

.meta-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
}

.forum-author {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  flex-shrink: 0;
}

.author-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 13px;
}

.author-info {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.author-name {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.author-date {
  margin: 0;
  font-size: 11px;
  color: var(--color-text-secondary);
}

@media (max-width: 768px) {
  .forums-layout {
    grid-template-columns: 1fr;
  }

  .forums-sidebar {
    display: none;
  }

  .forum-card {
    flex-direction: column;
    align-items: flex-start;
  }

  .forum-author {
    width: 100%;
    padding-top: var(--spacing-lg);
    border-top: 1px solid var(--color-border);
  }
}
</style>
