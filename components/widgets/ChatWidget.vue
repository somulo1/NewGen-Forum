<template>
  <div class="chat-widget">
    <div class="widget-header">
      <h3>Messages</h3>
      <NuxtLink to="/chat" class="view-all">View all →</NuxtLink>
    </div>
    <div class="chat-list">
      <div v-for="chat in chats" :key="chat.id" class="chat-item">
        <div class="chat-avatar">{{ chat.initials }}</div>
        <div class="chat-info">
          <h4>{{ chat.name }}</h4>
          <p>{{ chat.lastMessage }}</p>
        </div>
        <span v-if="chat.unread" class="unread-badge">{{ chat.unread }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const chats = [
  { id: 1, name: 'John Doe', initials: 'JD', lastMessage: 'Thanks for the update', unread: 2 },
  { id: 2, name: 'Sarah Martin', initials: 'SM', lastMessage: 'Can you check the files?', unread: 0 },
  { id: 3, name: 'Dev Team', initials: 'DT', lastMessage: 'Meeting at 2 PM', unread: 1 },
]
</script>

<style scoped>
.chat-widget {
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

.chat-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.chat-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.chat-item:last-child {
  border-bottom: none;
}

.chat-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 12px;
  flex-shrink: 0;
}

.chat-info {
  flex: 1;
}

.chat-info h4 {
  margin: 0 0 var(--spacing-xs) 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.chat-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.unread-badge {
  background: var(--color-coral);
  color: white;
  font-size: 11px;
  font-weight: 600;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
