<template>
  <aside class="hub-sidebar" :class="{ collapsed }">
    <nav class="sidebar-nav">
      <NuxtLink to="/" class="nav-link home-link" active-class="active">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"></path>
        </svg>
        <span>Home</span>
      </NuxtLink>

      <div v-for="section in sections" :key="section.title" class="nav-section">
        <h3 class="section-title">{{ section.title }}</h3>
        <NuxtLink v-for="item in section.items" :key="item.path" :to="item.path" class="nav-link" active-class="active">
          <svg width="18" height="18" :viewBox="item.icon.viewBox" fill="none" stroke="currentColor" stroke-width="2" v-html="item.icon.path"></svg>
          <span>{{ item.label }}</span>
          <span v-if="item.badge" class="badge">{{ item.badge }}</span>
        </NuxtLink>
      </div>
    </nav>

    <div class="sidebar-footer">
      <button class="collapse-btn" @click="collapsed = !collapsed" :title="collapsed ? 'Expand' : 'Collapse'">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="15 18 9 12 15 6"></polyline>
        </svg>
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const collapsed = ref(false)

const sections = [
  {
    title: 'MODULES',
    items: [
      { path: '/wall', label: 'Community wall', icon: { viewBox: '0 0 24 24', path: '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>' } },
      { path: '/sessions', label: 'Meetings & standups', icon: { viewBox: '0 0 24 24', path: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path>' } },
      { path: '/admin/calendar', label: 'Calendar', icon: { viewBox: '0 0 24 24', path: '<rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line>' } },
      { path: '/communications', label: 'Communications', icon: { viewBox: '0 0 24 24', path: '<path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>' }, badge: '3' },
      { path: '/tickets', label: 'Tickets', icon: { viewBox: '0 0 24 24', path: '<path d="M9 11l3 3L22 4"></path><path d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>' } },
      { path: '/projects', label: 'Projects', icon: { viewBox: '0 0 24 24', path: '<rect x="3" y="3" width="18" height="18" rx="2" ry="2"></rect><line x1="9" y1="9" x2="15" y2="9"></line><line x1="9" y1="15" x2="15" y2="15"></line>' } },
      { path: '/people', label: 'People', icon: { viewBox: '0 0 24 24', path: '<path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2"></path><circle cx="9" cy="7" r="4"></circle><path d="M23 21v-2a4 4 0 0 0-3-3.87"></path><path d="M16 3.13a4 4 0 0 1 0 7.75"></path>' } },
      { path: '/approvals', label: 'Approvals', icon: { viewBox: '0 0 24 24', path: '<polyline points="20 6 9 17 4 12"></polyline>' }, badge: '3' },
    ],
  },
  {
    title: 'ADMIN',
    items: [
      { path: '/admin/security', label: 'Security & logs', icon: { viewBox: '0 0 24 24', path: '<path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>' }, badge: '6' },
      { path: '/admin/roles', label: 'Roles & rights', icon: { viewBox: '0 0 24 24', path: '<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle>' } },
    ],
  },
]
</script>

<style scoped>
.hub-sidebar {
  width: 220px;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
  transition: width 0.3s ease;
}

.hub-sidebar.collapsed {
  width: 60px;
}

.sidebar-nav {
  flex: 1;
  padding: 12px 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.nav-section {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px 0;
  border-top: 1px solid var(--color-border);
}

.section-title {
  margin: 0;
  padding: 0 16px;
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--color-text-secondary);
  letter-spacing: 0.5px;
}

.home-link {
  padding: 0 16px;
}

.nav-link {
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
  white-space: nowrap;
  overflow: hidden;
}

.nav-link:hover {
  color: var(--color-text);
  background: var(--color-bg);
}

.nav-link.active {
  color: #B8F36B;
  font-weight: 600;
}

.nav-link.active::before {
  content: '';
  position: absolute;
  left: 0;
  top: 50%;
  transform: translateY(-50%);
  width: 3px;
  height: 20px;
  background: #B8F36B;
  border-radius: 0 2px 2px 0;
}

.nav-link svg {
  flex-shrink: 0;
}

.badge {
  margin-left: auto;
  background: #FF9B8A;
  color: #0F1714;
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 3px;
  flex-shrink: 0;
}

.hub-sidebar.collapsed .nav-link span {
  display: none;
}

.hub-sidebar.collapsed .badge {
  position: absolute;
  right: 4px;
  top: 4px;
  margin-left: 0;
}

.sidebar-footer {
  padding: 12px;
  border-top: 1px solid var(--color-border);
  display: flex;
  justify-content: center;
}

.collapse-btn {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.collapse-btn:hover {
  border-color: #B8F36B;
  color: #B8F36B;
}

@media (max-width: 768px) {
  .hub-sidebar {
    position: fixed;
    left: 0;
    top: 60px;
    height: calc(100vh - 60px);
    z-index: 100;
    box-shadow: 2px 0 8px rgba(0, 0, 0, 0.3);
  }

  .hub-sidebar.collapsed {
    width: 0;
    overflow: hidden;
  }
}
</style>
