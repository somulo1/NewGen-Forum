<template>
  <div class="wall-container">
    <div class="page-header">
      <div>
        <h1>Community Wall</h1>
        <p class="subtitle">Share updates with your team</p>
      </div>
      <button class="btn btn-primary">
        <Icon name="edit" size="18" />
        New Post
      </button>
    </div>

    <div class="wall-layout">
      <div class="wall-feed">
        <!-- Post Creator -->
        <div class="post-creator">
          <div class="creator-avatar">SM</div>
          <input type="text" placeholder="What's on your mind?" class="creator-input" />
        </div>

        <!-- Posts -->
        <div v-for="post in posts" :key="post.id" class="post-card">
          <div class="post-header">
            <div class="post-author">
              <div class="author-avatar">{{ post.authorInitials }}</div>
              <div class="author-info">
                <h3>{{ post.author }}</h3>
                <span class="post-time">{{ post.time }}</span>
              </div>
            </div>
            <button class="post-menu">
              <Icon name="more-vertical" size="16" />
            </button>
          </div>

          <div class="post-content">
            <p>{{ post.content }}</p>
          </div>

          <div class="post-footer">
            <button class="post-action">
              <Icon name="heart" size="16" />
              {{ post.likes }}
            </button>
            <button class="post-action">
              <Icon name="message-square" size="16" />
              {{ post.comments }}
            </button>
            <button class="post-action">
              <Icon name="share-2" size="16" />
              Share
            </button>
          </div>
        </div>
      </div>

      <aside class="wall-sidebar">
        <div class="sidebar-card">
          <h3>Trending Topics</h3>
          <div class="trending-list">
            <a href="#" class="trending-item">#ProjectX</a>
            <a href="#" class="trending-item">#TeamBuilding</a>
            <a href="#" class="trending-item">#Innovation</a>
            <a href="#" class="trending-item">#Culture</a>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const posts = [
  {
    id: 1,
    author: 'John Doe',
    authorInitials: 'JD',
    time: '2 hours ago',
    content: 'Excited to announce our new product launch! Check back next week for more details. #ProductX',
    likes: 24,
    comments: 5,
  },
  {
    id: 2,
    author: 'Sarah Martin',
    authorInitials: 'SM',
    time: '4 hours ago',
    content: 'Great team meeting today! Lots of positive energy and great ideas. Looking forward to the next sprint!',
    likes: 18,
    comments: 3,
  },
  {
    id: 3,
    author: 'Marcus Davis',
    authorInitials: 'MD',
    time: '1 day ago',
    content: 'Just released a new feature update. Thanks to everyone who contributed! 🚀',
    likes: 42,
    comments: 8,
  },
]
</script>

<style scoped>
.wall-container {
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
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: var(--spacing-sm) 0 0 0;
}

.wall-layout {
  display: grid;
  grid-template-columns: 1fr 280px;
  gap: var(--spacing-lg);
}

.wall-feed {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.post-creator {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.creator-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  flex-shrink: 0;
}

.creator-input {
  flex: 1;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  color: var(--color-text);
  font-family: 'DM Sans', sans-serif;
}

.creator-input::placeholder {
  color: var(--color-text-secondary);
}

.post-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
  transition: all 0.2s ease;
}

.post-card:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.post-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: var(--spacing-lg);
  padding-bottom: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.post-author {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.author-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  flex-shrink: 0;
}

.author-info h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
}

.post-time {
  font-size: 12px;
  color: var(--color-text-secondary);
}

.post-menu {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: var(--spacing-sm);
}

.post-content {
  margin-bottom: var(--spacing-lg);
}

.post-content p {
  margin: 0;
  color: var(--color-text);
  line-height: 1.6;
}

.post-footer {
  display: flex;
  gap: var(--spacing-lg);
  padding-top: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
}

.post-action {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 13px;
  transition: color 0.2s ease;
}

.post-action:hover {
  color: var(--color-lime);
}

.wall-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.sidebar-card {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.sidebar-card h3 {
  margin: 0 0 var(--spacing-lg) 0;
  font-size: 14px;
  font-weight: 600;
}

.trending-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.trending-item {
  padding: var(--spacing-md);
  background: var(--color-bg);
  border-radius: var(--radius-md);
  color: var(--color-lime);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.trending-item:hover {
  background: rgba(184, 243, 107, 0.1);
}

@media (max-width: 1024px) {
  .wall-layout {
    grid-template-columns: 1fr;
  }

  .wall-sidebar {
    display: none;
  }
}
</style>
