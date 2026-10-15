<template>
  <div class="files-container">
    <div class="page-header">
      <div>
        <h1>Files & Storage</h1>
        <p class="subtitle">{{ storageUsed }}% storage used</p>
      </div>
      <button class="btn btn-primary">
        <Icon name="upload" size="18" />
        Upload File
      </button>
    </div>

    <div class="files-layout">
      <aside class="files-sidebar">
        <nav class="nav-folders">
          <button :class="['nav-item', { active: activeFolder === 'all' }]" @click="activeFolder = 'all'">
            <Icon name="folder" size="16" />
            All Files
          </button>
          <button :class="['nav-item', { active: activeFolder === 'recent' }]" @click="activeFolder = 'recent'">
            <Icon name="clock" size="16" />
            Recent
          </button>
          <button :class="['nav-item', { active: activeFolder === 'shared' }]" @click="activeFolder = 'shared'">
            <Icon name="share-2" size="16" />
            Shared with Me
          </button>
          <button :class="['nav-item', { active: activeFolder === 'starred' }]" @click="activeFolder = 'starred'">
            <Icon name="star" size="16" />
            Starred
          </button>
          <button :class="['nav-item', { active: activeFolder === 'trash' }]" @click="activeFolder = 'trash'">
            <Icon name="trash-2" size="16" />
            Trash
          </button>
        </nav>

        <div class="storage-info">
          <h4>Storage</h4>
          <div class="storage-bar">
            <div class="storage-fill" :style="{ width: storageUsed + '%' }"></div>
          </div>
          <p>{{ storageUsed }}% of 100 GB used</p>
        </div>

        <div class="my-folders">
          <h4>My Folders</h4>
          <div class="folder-list">
            <a href="#" class="folder-item">
              <Icon name="folder" size="14" />
              Projects
            </a>
            <a href="#" class="folder-item">
              <Icon name="folder" size="14" />
              Documents
            </a>
            <a href="#" class="folder-item">
              <Icon name="folder" size="14" />
              Media
            </a>
          </div>
        </div>
      </aside>

      <main class="files-main">
        <div class="files-toolbar">
          <input type="text" placeholder="Search files..." class="search-input" />
          <div class="view-options">
            <button class="view-btn active" title="List view">
              <Icon name="list" size="16" />
            </button>
            <button class="view-btn" title="Grid view">
              <Icon name="grid" size="16" />
            </button>
          </div>
        </div>

        <div class="files-list">
          <div v-for="file in files" :key="file.id" class="file-row">
            <div class="file-checkbox">
              <input type="checkbox" />
            </div>
            <div class="file-icon">
              <Icon :name="file.icon" size="16" />
            </div>
            <div class="file-info">
              <h4>{{ file.name }}</h4>
              <p>{{ file.size }} • {{ file.modified }}</p>
            </div>
            <div class="file-owner">{{ file.owner }}</div>
            <div class="file-actions">
              <button class="action-btn">
                <Icon name="share-2" size="14" />
              </button>
              <button class="action-btn">
                <Icon name="star" size="14" />
              </button>
              <button class="action-btn">
                <Icon name="more-vertical" size="14" />
              </button>
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

const activeFolder = ref('all')
const storageUsed = ref(65)

const files = [
  { id: 1, name: 'Q4 Report 2024.pdf', size: '2.4 MB', modified: '2 days ago', owner: 'You', icon: 'file' },
  { id: 2, name: 'Design System v3.figma', size: '18 MB', modified: '1 week ago', owner: 'Sarah Martin', icon: 'file' },
  { id: 3, name: 'Budget 2025.xlsx', size: '156 KB', modified: '3 days ago', owner: 'You', icon: 'file' },
  { id: 4, name: 'Marketing Presentation.pptx', size: '5.2 MB', modified: '4 days ago', owner: 'John Doe', icon: 'file' },
  { id: 5, name: 'Project Assets', size: '--', modified: '1 week ago', owner: 'You', icon: 'folder' },
  { id: 6, name: 'Team Meeting Notes.docx', size: '328 KB', modified: '5 days ago', owner: 'You', icon: 'file' },
  { id: 7, name: 'Video Recording.mp4', size: '456 MB', modified: '2 weeks ago', owner: 'Marcus Davis', icon: 'file' },
  { id: 8, name: 'Contract - Draft.pdf', size: '892 KB', modified: '1 day ago', owner: 'Legal Team', icon: 'file' },
]
</script>

<style scoped>
.files-container {
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

.files-layout {
  display: grid;
  grid-template-columns: 200px 1fr;
  gap: var(--spacing-lg);
}

.files-sidebar {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.nav-folders {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md) var(--spacing-lg);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  color: var(--color-text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}

.nav-item:hover {
  border-color: var(--color-lime);
  color: var(--color-text);
}

.nav-item.active {
  background: rgba(184, 243, 107, 0.1);
  border-color: var(--color-lime);
  color: var(--color-lime);
}

.storage-info {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.storage-info h4 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.storage-bar {
  width: 100%;
  height: 6px;
  background: var(--color-border);
  border-radius: var(--radius-xs);
  overflow: hidden;
  margin-bottom: var(--spacing-md);
}

.storage-fill {
  height: 100%;
  background: linear-gradient(90deg, var(--color-lime) 0%, var(--color-sky) 100%);
}

.storage-info p {
  margin: 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.my-folders {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-lg);
}

.my-folders h4 {
  margin: 0 0 var(--spacing-md) 0;
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
}

.folder-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.folder-item {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-sm) var(--spacing-md);
  color: var(--color-text-secondary);
  text-decoration: none;
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
  font-size: 13px;
}

.folder-item:hover {
  background: var(--color-border);
  color: var(--color-lime);
}

.files-main {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.files-toolbar {
  display: flex;
  align-items: center;
  gap: var(--spacing-lg);
}

.search-input {
  flex: 1;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-md) var(--spacing-lg);
  color: var(--color-text);
  font-family: 'DM Sans', sans-serif;
}

.search-input::placeholder {
  color: var(--color-text-secondary);
}

.view-options {
  display: flex;
  gap: var(--spacing-sm);
}

.view-btn {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text-secondary);
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.view-btn:hover {
  border-color: var(--color-lime);
  color: var(--color-text);
}

.view-btn.active {
  background: var(--color-lime);
  border-color: var(--color-lime);
  color: #000;
}

.files-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.file-row {
  display: grid;
  grid-template-columns: 40px 40px 1fr 120px 80px;
  gap: var(--spacing-lg);
  align-items: center;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-bottom: none;
  padding: var(--spacing-lg);
  transition: all 0.2s ease;
}

.file-row:first-child {
  border-radius: var(--radius-lg) var(--radius-lg) 0 0;
}

.file-row:last-child {
  border-bottom: 1px solid var(--color-border);
  border-radius: 0 0 var(--radius-lg) var(--radius-lg);
}

.file-row:hover {
  background: var(--color-border);
}

.file-checkbox input {
  width: 16px;
  height: 16px;
  cursor: pointer;
}

.file-icon {
  width: 36px;
  height: 36px;
  background: rgba(184, 243, 107, 0.1);
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-lime);
}

.file-info h4 {
  margin: 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--color-text);
}

.file-info p {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--color-text-secondary);
}

.file-owner {
  font-size: 13px;
  color: var(--color-text-secondary);
}

.file-actions {
  display: flex;
  gap: var(--spacing-sm);
  justify-content: flex-end;
}

.action-btn {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.action-btn:hover {
  background: var(--color-surface);
  color: var(--color-lime);
}

@media (max-width: 1024px) {
  .files-layout {
    grid-template-columns: 1fr;
  }

  .file-row {
    grid-template-columns: 40px 1fr 40px;
  }

  .file-owner {
    display: none;
  }

  .files-sidebar {
    display: none;
  }
}
</style>
