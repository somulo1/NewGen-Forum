<template>
  <header class="hub-header">
    <div class="header-left">
      <button class="menu-btn" @click="$emit('toggleSidebar')">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="4" y1="7" x2="20" y2="7"></line>
          <line x1="4" y1="12" x2="20" y2="12"></line>
          <line x1="4" y1="17" x2="20" y2="17"></line>
        </svg>
      </button>
      
      <div class="logo">
        <div class="logo-icon">H</div>
        <span class="logo-text">Hub</span>
        <span class="org-label">[YOUR ORGANIZATION]</span>
      </div>
    </div>

    <div class="header-search">
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <circle cx="11" cy="11" r="7"></circle>
        <path d="M20 20l-4-4"></path>
      </svg>
      <input type="text" placeholder="Search people, forums, files, tickets…">
    </div>

    <div class="header-right">
      <button class="theme-btn" @click="toggleTheme">
        <svg v-if="isDark" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="4"></circle>
          <path d="M12 2v2M12 20v2M2 12h2M20 12h2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"></path>
        </svg>
        <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M20 14.5A8 8 0 0 1 9.5 4a8 8 0 1 0 10.5 10.5z"></path>
        </svg>
        <span>{{ isDark ? 'Light' : 'Dark' }}</span>
      </button>

      <button class="notif-btn" @click="toggleNotifications">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M6 16v-5a6 6 0 0 1 12 0v5l1.5 2h-15z"></path>
          <path d="M10 20.5a2 2 0 0 0 4 0"></path>
        </svg>
        <span v-if="hasNotifications" class="notif-badge">1</span>
      </button>

      <button class="profile-btn" @click="goToProfile">S</button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const isDark = ref(true)
const hasNotifications = ref(true)

const toggleTheme = () => {
  isDark.value = !isDark.value
}

const toggleNotifications = () => {
  // Handle notifications
}

const goToProfile = () => {
  // Navigate to profile
}
</script>

<style scoped>
.hub-header {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 20px;
  border-bottom: 1px solid var(--color-border);
  background: var(--color-bg);
  height: 60px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 0 0 auto;
}

.menu-btn {
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.menu-btn:hover {
  border-color: #B8F36B;
  color: #B8F36B;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}

.logo-icon {
  width: 34px;
  height: 34px;
  border-radius: 9px;
  background: #B8F36B;
  color: #0F1714;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 17px;
}

.logo-text {
  font-family: 'Space Grotesk', sans-serif;
  font-weight: 700;
  font-size: 18px;
  color: var(--color-text);
}

.org-label {
  font-size: 13px;
  color: var(--color-text-secondary);
  padding-left: 10px;
  border-left: 1px solid var(--color-border);
}

.header-search {
  flex: 1;
  max-width: 560px;
  display: flex;
  align-items: center;
  gap: 10px;
  height: 42px;
  padding: 0 14px;
  border-radius: 21px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
}

.header-search svg {
  flex-shrink: 0;
}

.header-search input {
  flex: 1;
  background: transparent;
  border: none;
  outline: none;
  color: var(--color-text);
  font-size: 14px;
}

.header-search input::placeholder {
  color: var(--color-text-secondary);
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 0 0 auto;
}

.theme-btn {
  height: 40px;
  padding: 0 14px;
  border-radius: 20px;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: all 0.2s ease;
}

.theme-btn:hover {
  border-color: #B8F36B;
  color: #B8F36B;
}

.notif-btn {
  position: relative;
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.notif-btn:hover {
  border-color: #B8F36B;
  color: #B8F36B;
}

.notif-badge {
  position: absolute;
  top: -4px;
  right: -4px;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  border-radius: 9px;
  background: #FF9B8A;
  color: #0F1714;
  font-size: 11px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
}

.profile-btn {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  border: 2px solid #F49AC1;
  background: #C2185B;
  color: white;
  font-weight: 600;
  font-size: 15px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.profile-btn:hover {
  border-color: #B8F36B;
  background: #D81B60;
}

@media (max-width: 768px) {
  .header-search {
    display: none;
  }

  .org-label {
    display: none;
  }
}
</style>
