<template>
  <div class="hub-sidebar-wrapper">
    <!-- Main Navigation -->
    <NuxtLink to="/" :class="['nav-item', { active: isActive('/') }]">
      <Icon name="home" size="18" />
      <span class="nav-label">Home</span>
    </NuxtLink>

    <!-- Messages Section -->
    <div v-if="enabledModules.includes('chat')" class="sidebar-section">
      <h3 class="sidebar-title">MESSAGES</h3>
      <div class="sidebar-section-header">
        <span class="section-count">{{ messageCount }} NEW</span>
        <button class="section-add">+</button>
      </div>
      <nav class="sidebar-nav">
        <NuxtLink
          v-for="person in recentMessages"
          :key="person.id"
          :class="['nav-item', 'chat-item', { active: isActive(`/chat/${person.id}`) }]"
        >
          <div class="avatar-small" :style="{ background: person.color }">{{ person.initials }}</div>
          <span class="nav-label">{{ person.name }}</span>
          <span v-if="person.unread" class="unread-indicator">{{ person.unread }}</span>
        </NuxtLink>
      </nav>
    </div>

    <!-- Main Modules (Added Widgets) -->
    <div v-if="visibleModules.length > 0" class="sidebar-section">
      <h3 class="sidebar-title">MODULES</h3>
      <nav class="sidebar-nav">
        <NuxtLink
          v-for="module in visibleModules"
          :key="module.id"
          :to="module.path"
          :class="['nav-item', { active: isActive(module.path) }]"
        >
          <Icon :name="module.icon" size="18" />
          <span class="nav-label">{{ module.name }}</span>
          <span v-if="module.badge" class="nav-badge">{{ module.badge }}</span>
        </NuxtLink>
      </nav>
    </div>

    <!-- Organization Section -->
    <div v-if="enabledModules.some(m => m.startsWith('org-'))" class="sidebar-section">
      <h3 class="sidebar-title">ORGANIZATION</h3>
      <nav class="sidebar-nav">
        <NuxtLink
          v-for="org in organizationItems"
          :key="org.id"
          :to="org.path"
          :class="['nav-item', { active: isActive(org.path) }]"
        >
          <Icon :name="org.icon" size="18" />
          <span class="nav-label">{{ org.name }}</span>
        </NuxtLink>
      </nav>
    </div>

    <!-- Admin Section -->
    <div v-if="enabledModules.some(m => m.startsWith('admin-'))" class="sidebar-section">
      <h3 class="sidebar-title">ADMIN</h3>
      <nav class="sidebar-nav">
        <NuxtLink
          v-for="admin in adminItems.filter(a => enabledModules.includes(`admin-${a.id}`))"
          :key="admin.id"
          :to="admin.path"
          :class="['nav-item', { active: isActive(admin.path) }]"
        >
          <Icon :name="admin.icon" size="18" />
          <span class="nav-label">{{ admin.name }}</span>
          <span v-if="admin.badge" class="nav-badge">{{ admin.badge }}</span>
        </NuxtLink>
      </nav>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { ref, computed } from 'vue'
import Icon from '~/components/Icon.vue'

const route = useRoute()

// Modules that can be added as widgets
const allModules = [
  { id: 'community-wall', name: 'Community wall', icon: 'message-square', path: '/wall' },
  { id: 'meetings', name: 'Meetings & standups', icon: 'users', path: '/sessions' },
  { id: 'calendar', name: 'Calendar', icon: 'calendar', path: '/admin/calendar' },
  { id: 'communications', name: 'Communications', icon: 'send', path: '/communications' },
  { id: 'tickets', name: 'Tickets', icon: 'edit', path: '/tickets' },
  { id: 'projects', name: 'Projects', icon: 'box', path: '/projects' },
  { id: 'people', name: 'People', icon: 'users', path: '/people' },
  { id: 'approvals', name: 'Approvals', icon: 'check', path: '/approvals' },
]

const organizationItems = [
  { id: 'departments', name: 'Departments', icon: 'folder', path: '/departments' },
  { id: 'sectors', name: 'Sectors', icon: 'git-branch', path: '/sectors' },
  { id: 'branches', name: 'Branches', icon: 'git-branch', path: '/branches' },
  { id: 'units', name: 'Units', icon: 'box', path: '/units' },
]

const adminItems = [
  { id: 'security', name: 'Security & logs', icon: 'shield', path: '/admin/security', badge: '6' },
  { id: 'roles', name: 'Roles & rights', icon: 'lock', path: '/admin/roles' },
]

// Simulated: which modules are currently enabled (added to home)
const enabledModules = ref([
  'community-wall',
  'meetings',
  'calendar',
  'communications',
  'tickets',
  'projects',
  'people',
  'approvals',
  'chat',
  'admin-security',
  'admin-roles',
  'org-departments',
])

const visibleModules = computed(() => {
  return allModules.filter(m => enabledModules.value.includes(m.id))
})

const messageCount = ref(3)
const recentMessages = [
  { id: 1, name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', unread: 2 },
  { id: 2, name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', unread: 1 },
  { id: 3, name: 'Grace Njeri', initials: 'GN', color: '#FFB86B', unread: 0 },
  { id: 4, name: 'David Otieno', initials: 'DO', color: '#F49AC1', unread: 0 },
]

const isActive = (path: string) => {
  return route.path === path
}
</script>

<style scoped>
.hub-sidebar-wrapper {
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.sidebar-section {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-title {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  color: var(--color-text-secondary);
  margin: 0;
  padding: 0 16px;
}

.sidebar-section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  font-size: 11px;
  color: var(--color-text-secondary);
}

.section-count {
  font-weight: 600;
}

.section-add {
  background: transparent;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 16px;
  padding: 0;
}

.sidebar-nav {
  display: flex;
  flex-direction: column;
  gap: 0;
  list-style: none;
  margin: 0;
  padding: 0;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  color: var(--color-text-secondary);
  text-decoration: none;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
  position: relative;
}

.nav-item:hover {
  color: var(--color-text);
  background: transparent;
}

.nav-item.active {
  color: var(--color-lime);
  font-weight: 600;
}

.nav-item.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: var(--color-lime);
  border-radius: 0 2px 2px 0;
}

.nav-item.chat-item {
  padding: 8px 16px;
}

.avatar-small {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 700;
  color: white;
  flex-shrink: 0;
}

.nav-label {
  flex: 1;
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.unread-indicator {
  background: var(--color-coral);
  color: white;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 5px;
  border-radius: 3px;
  margin-left: auto;
  flex-shrink: 0;
}

.nav-badge {
  background: var(--color-coral);
  color: white;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 3px;
  margin-left: auto;
  flex-shrink: 0;
}

@media (max-width: 768px) {
  .hub-sidebar-wrapper {
    padding: 12px 8px;
  }

  .nav-item {
    padding: 8px 12px;
  }

  .nav-label {
    display: none;
  }

  .nav-badge,
  .unread-indicator {
    position: absolute;
    right: 8px;
    top: 8px;
  }
}
</style>
