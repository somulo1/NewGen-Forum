<template>
  <div class="wall-container">
    <div class="page-header">
      <div>
        <h1>Community wall</h1>
        <p class="subtitle">Everyone's space to post freely: photos, news, shout-outs, polls.</p>
      </div>
    </div>

    <div class="moments-section">
      <button class="moment-btn add-moment">
        <span class="plus-icon">+</span>
        Add a moment
      </button>
      <div v-for="moment in moments" :key="moment.id" class="moment-card" :style="{ background: moment.color }">
        <div class="moment-avatar" :style="{ background: moment.avatarColor }">{{ moment.initials }}</div>
        <div class="moment-text">{{ moment.text }}</div>
      </div>
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

const moments = [
  { id: 1, text: 'Brian at the server room', color: '#2C4563', avatarColor: '#7CC4FF', initials: 'BK' },
  { id: 2, text: 'Team lunch', color: '#5C4936', avatarColor: '#FFB86B', initials: 'GN' },
  { id: 3, text: 'Game night setup', color: '#5C3D5C', avatarColor: '#F49AC1', initials: 'DO' },
  { id: 4, text: 'First day!', color: '#4D5C3D', avatarColor: '#B8F36B', initials: 'FA' },
]

const posts = [
  {
    id: 1,
    author: 'Amina Wanjiru',
    authorInitials: 'AW',
    time: '12 min ago',
    content: 'Huge thanks to everyone who joined the charity run on Saturday. We raised [AMOUNT] for the children\'s home!',
    likes: 48,
    comments: 11,
  },
]
</script>

<style scoped>
.wall-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
}

.page-header h1 {
  margin: 0;
  font-size: 32px;
  font-weight: 600;
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: 8px 0 0 0;
}

.moments-section {
  display: flex;
  gap: 16px;
  overflow-x: auto;
  padding: 4px 0;
}

.moment-btn {
  background: transparent;
  border: none;
  cursor: pointer;
  padding: 0;
}

.add-moment {
  width: 120px;
  height: 160px;
  border: 2px solid var(--color-border);
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--color-text);
  transition: all 0.2s ease;
}

.add-moment:hover {
  background: var(--color-surface);
  border-color: #B8F36B;
}

.plus-icon {
  font-size: 32px;
  color: #B8F36B;
}

.moment-card {
  width: 120px;
  height: 160px;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: white;
  flex-shrink: 0;
  text-align: center;
}

.moment-avatar {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  font-weight: 700;
  color: white;
}

.moment-text {
  font-size: 12px;
  font-weight: 600;
  padding: 0 8px;
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
