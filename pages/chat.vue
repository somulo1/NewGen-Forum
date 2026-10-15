<template>
  <div class="chat-container">
    <div class="chat-layout">
      <aside class="chats-panel">
        <div class="panel-header">
          <h2>Chats</h2>
        </div>

        <div class="conversations-list">
          <div
            v-for="conversation in conversations"
            :key="conversation.id"
            :class="['conversation-item', { active: activeChat === conversation.id }]"
            @click="activeChat = conversation.id"
          >
            <div class="conv-avatar" :style="{ background: conversation.color }">{{ conversation.initials }}</div>
            <div class="conv-content">
              <h4>{{ conversation.name }}</h4>
              <p>{{ conversation.lastMessage }}</p>
            </div>
            <span v-if="conversation.unread" class="unread-badge">{{ conversation.unread }}</span>
          </div>
        </div>
      </aside>

      <div v-if="activeChat" class="chat-main">
        <div class="chat-header">
          <div class="header-info">
            <div class="header-avatar" :style="{ background: activeChatData.color }">{{ activeChatData.initials }}</div>
            <div>
              <h3>{{ activeChatData.name }}</h3>
              <p>{{ activeChatData.status }}</p>
            </div>
          </div>
          <div class="header-actions">
            <button class="header-btn"><Icon name="phone" size="18" /></button>
            <button class="header-btn"><Icon name="video" size="18" /></button>
            <button class="header-btn"><Icon name="more-vertical" size="18" /></button>
          </div>
        </div>

        <div class="messages-area">
          <div v-for="msg in messages" :key="msg.id" :class="['message-group', msg.type]">
            <div v-if="msg.type === 'received'" class="message-avatar" :style="{ background: activeChatData.color }">
              {{ activeChatData.initials }}
            </div>
            <div class="message-bubble">{{ msg.text }}</div>
            <span v-if="msg.time" class="message-time">{{ msg.time }}</span>
          </div>
        </div>

        <div class="message-input-area">
          <button class="input-btn"><Icon name="plus" size="20" /></button>
          <input
            v-model="inputMessage"
            type="text"
            placeholder="Message Brian Kiprop..."
            class="message-input"
            @keyup.enter="sendMessage"
          />
          <button class="input-btn emoji"><Icon name="smile" size="20" /></button>
          <button class="send-btn" @click="sendMessage"><Icon name="send" size="18" /></button>
        </div>
      </div>

      <div v-else class="empty-state">
        <Icon name="message-square" size="48" />
        <p>Select a conversation to start chatting</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const activeChat = ref(1)
const inputMessage = ref('')

const conversations = [
  {
    id: 1,
    name: 'Brian Kiprop',
    initials: 'BK',
    lastMessage: 'I have 15 minutes before 4...',
    color: '#7CC4FF',
    unread: 2,
    status: 'Available · private chat',
  },
  {
    id: 2,
    name: 'Amina Wanjiru',
    initials: 'AW',
    lastMessage: 'Do you have the audit...',
    color: '#6EE7C8',
    unread: 1,
    status: 'Available · private chat',
  },
  {
    id: 3,
    name: 'Grace Njeri',
    initials: 'GN',
    lastMessage: 'You: Perfect, I\'ll check...',
    color: '#FFB86B',
    unread: 0,
    status: 'Away · private chat',
  },
  {
    id: 4,
    name: 'David Otieno',
    initials: 'DO',
    lastMessage: 'Yes, picking it up tomorrow...',
    color: '#F49AC1',
    unread: 0,
    status: 'Away · private chat',
  },
]

const messages = [
  { id: 1, type: 'received', text: 'Hi Sam, the VPN certificate issue is fixed.', time: '2:10 PM' },
  { id: 2, type: 'sent', text: 'Great, thanks!', time: '2:12 PM' },
  { id: 3, type: 'received', text: 'Can we do a quick call about the SIEM rollout?', time: '3:31 PM' },
  { id: 4, type: 'sent', text: 'I have 15 minutes before 4.', time: null },
]

const activeChatData = computed(() => conversations[activeChat.value - 1] || conversations[0])

const sendMessage = () => {
  if (inputMessage.value.trim()) {
    messages.push({
      id: messages.length + 1,
      type: 'sent',
      text: inputMessage.value,
      time: new Date().toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' }),
    })
    inputMessage.value = ''
  }
}
</script>

<style scoped>
.chat-container {
  display: flex;
  height: calc(100vh - 120px);
  background: var(--color-bg);
}

.chat-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  width: 100%;
  height: 100%;
}

.chats-panel {
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.panel-header {
  padding: 16px;
  border-bottom: 1px solid var(--color-border);
}

.panel-header h2 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: var(--color-text);
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
}

.conversation-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  cursor: pointer;
  border-left: 3px solid transparent;
  transition: all 0.2s ease;
}

.conversation-item:hover {
  background: var(--color-border);
}

.conversation-item.active {
  background: rgba(184, 243, 107, 0.1);
  border-left-color: var(--color-lime);
}

.conv-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
  flex-shrink: 0;
}

.conv-content {
  flex: 1;
  min-width: 0;
}

.conv-content h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.conv-content p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.unread-badge {
  background: var(--color-coral);
  color: white;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: 700;
  flex-shrink: 0;
}

.chat-main {
  display: flex;
  flex-direction: column;
  background: var(--color-bg);
}

.chat-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  padding: 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.header-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.header-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 13px;
}

.header-info h3 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--color-text);
}

.header-info p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.header-actions {
  display: flex;
  gap: 8px;
}

.header-btn {
  width: 36px;
  height: 36px;
  background: var(--color-border);
  border: none;
  border-radius: 8px;
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.header-btn:hover {
  background: var(--color-lime);
  color: #000;
}

.messages-area {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.message-group {
  display: flex;
  gap: 8px;
  align-items: flex-end;
}

.message-group.sent {
  justify-content: flex-end;
  align-items: flex-end;
}

.message-avatar {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 11px;
  flex-shrink: 0;
}

.message-bubble {
  max-width: 50%;
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.5;
  word-wrap: break-word;
}

.message-group.received .message-bubble {
  background: var(--color-surface);
  color: var(--color-text);
}

.message-group.sent .message-bubble {
  background: var(--color-lime);
  color: #000;
}

.message-time {
  font-size: 11px;
  color: var(--color-text-secondary);
  padding: 0 4px;
}

.message-input-area {
  background: var(--color-surface);
  border-top: 1px solid var(--color-border);
  padding: 12px 16px;
  display: flex;
  align-items: flex-end;
  gap: 8px;
}

.input-btn {
  width: 36px;
  height: 36px;
  background: transparent;
  border: none;
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  transition: all 0.2s ease;
}

.input-btn:hover {
  background: var(--color-border);
}

.message-input {
  flex: 1;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  border-radius: 20px;
  padding: 10px 16px;
  color: var(--color-text);
  font-family: 'DM Sans', sans-serif;
  font-size: 13px;
}

.message-input::placeholder {
  color: var(--color-text-secondary);
}

.message-input:focus {
  outline: none;
  border-color: var(--color-lime);
}

.send-btn {
  width: 36px;
  height: 36px;
  background: var(--color-lime);
  border: none;
  color: #000;
  border-radius: 8px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: opacity 0.2s ease;
}

.send-btn:hover {
  opacity: 0.9;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 16px;
  color: var(--color-text-secondary);
}

.empty-state :deep(svg) {
  opacity: 0.5;
}

@media (max-width: 768px) {
  .chat-layout {
    grid-template-columns: 1fr;
  }

  .chats-panel {
    display: none;
  }

  .message-bubble {
    max-width: 70%;
  }
}
</style>
