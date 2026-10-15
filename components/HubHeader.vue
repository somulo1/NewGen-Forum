<template>
  <div class="hub-header-container">
    <div class="hub-header-left">
      <h1 class="hub-logo">Hub</h1>
    </div>

    <div class="hub-header-center">
      <div class="search-bar">
        <Icon name="search" size="18" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Search forums, files, people..."
          class="search-input"
        />
      </div>
    </div>

    <div class="hub-header-right">
      <button class="header-icon-btn" title="Notifications">
        <Icon name="bell" size="20" />
        <span class="notification-badge">3</span>
      </button>

      <button class="header-icon-btn" @click="toggleTheme" title="Toggle theme">
        <Icon :name="theme === 'dark' ? 'sun' : 'moon'" size="20" />
      </button>

      <div class="user-menu">
        <button class="user-button" @click="showUserMenu = !showUserMenu">
          <div class="user-avatar">SM</div>
          <Icon name="chevron-down" size="16" />
        </button>

        <div v-if="showUserMenu" class="user-menu-dropdown">
          <a href="#" class="menu-item">
            <Icon name="user" size="16" />
            Profile
          </a>
          <a href="#" class="menu-item">
            <Icon name="settings" size="16" />
            Settings
          </a>
          <a href="#" class="menu-item">
            <Icon name="help-circle" size="16" />
            Help
          </a>
          <hr class="menu-divider" />
          <a href="#" class="menu-item">
            <Icon name="log-out" size="16" />
            Logout
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useTheme } from '~/composables/useTheme'
import Icon from '~/components/Icon.vue'

const { theme, toggleTheme } = useTheme()
const searchQuery = ref('')
const showUserMenu = ref(false)

// Close menu when clicking outside
watch(showUserMenu, (newVal) => {
  if (newVal) {
    nextTick(() => {
      document.addEventListener('click', closeMenu)
    })
  }
})

const closeMenu = () => {
  showUserMenu.value = false
  document.removeEventListener('click', closeMenu)
}
</script>

<style scoped>
.hub-header-container {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  height: 100%;
  gap: var(--spacing-lg);
}

.hub-header-left {
  flex-shrink: 0;
}

.hub-logo {
  font-family: 'Space Grotesk', system-ui, sans-serif;
  font-size: 24px;
  font-weight: 700;
  background: linear-gradient(135deg, var(--color-lime) 0%, var(--color-sky) 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  margin: 0;
}

.hub-header-center {
  flex: 1;
  max-width: 400px;
}

.search-bar {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) var(--spacing-md);
  background: var(--color-bg);
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

.hub-header-right {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
  flex-shrink: 0;
}

.header-icon-btn {
  position: relative;
  background: none;
  border: none;
  color: var(--color-text);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-icon-btn:hover {
  background: var(--color-border);
}

.notification-badge {
  position: absolute;
  top: 2px;
  right: 2px;
  background: var(--color-coral);
  color: white;
  font-size: 10px;
  font-weight: 600;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.user-menu {
  position: relative;
}

.user-button {
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  background: var(--color-border);
  border: none;
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-lg);
  cursor: pointer;
  color: var(--color-text);
  transition: all 0.2s ease;
}

.user-button:hover {
  background: var(--color-dark-gray);
}

.user-avatar {
  width: 28px;
  height: 28px;
  background: linear-gradient(135deg, var(--color-sky) 0%, var(--color-teal) 100%);
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 12px;
}

.user-menu-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  min-width: 180px;
  margin-top: var(--spacing-sm);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
  z-index: 100;
}

.menu-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  color: var(--color-text);
  text-decoration: none;
  transition: all 0.2s ease;
}

.menu-item:hover {
  background: var(--color-border);
  color: var(--color-lime);
}

.menu-divider {
  border: none;
  border-top: 1px solid var(--color-border);
  margin: var(--spacing-sm) 0;
}

@media (max-width: 768px) {
  .hub-header-center {
    max-width: 200px;
  }

  .search-bar {
    padding: var(--spacing-sm);
  }

  .search-input {
    font-size: 12px;
  }
}
</style>
