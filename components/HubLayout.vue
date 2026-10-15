<template>
  <div class="hub-layout">
    <aside class="hub-sidebar" :style="{ width: sidebarWidth }">
      <div class="hub-sidebar-content">
        <slot name="sidebar" />
      </div>
      <div class="hub-sidebar-resize" @mousedown="startResize"></div>
    </aside>

    <div class="hub-main">
      <header class="hub-header">
        <slot name="header" />
      </header>

      <main class="hub-content">
        <slot name="content" />
      </main>
    </div>
  </div>
</template>

<script setup lang="ts">
const sidebarWidth = ref('320px')
const isResizing = ref(false)
const startX = ref(0)
const startWidth = ref(0)

const startResize = (e: MouseEvent) => {
  isResizing.value = true
  startX.value = e.clientX
  const sidebar = document.querySelector('.hub-sidebar') as HTMLElement
  if (sidebar) {
    startWidth.value = sidebar.offsetWidth
  }

  document.addEventListener('mousemove', handleResize)
  document.addEventListener('mouseup', stopResize)
}

const handleResize = (e: MouseEvent) => {
  if (!isResizing.value) return

  const diff = e.clientX - startX.value
  const newWidth = startWidth.value + diff

  if (newWidth >= 320 && newWidth <= 640) {
    sidebarWidth.value = `${newWidth}px`
    localStorage.setItem('hub-sidebar-width', `${newWidth}px`)
  }
}

const stopResize = () => {
  isResizing.value = false
  document.removeEventListener('mousemove', handleResize)
  document.removeEventListener('mouseup', stopResize)
}

onMounted(() => {
  const saved = localStorage.getItem('hub-sidebar-width')
  if (saved) {
    sidebarWidth.value = saved
  }
})
</script>

<style scoped>
.hub-layout {
  display: flex;
  width: 100%;
  height: 100vh;
  background: var(--color-bg);
}

.hub-sidebar {
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-right: 1px solid var(--color-border);
  overflow-y: auto;
  transition: width 0.3s ease;
  position: relative;
}

.hub-sidebar-content {
  flex: 1;
  overflow-y: auto;
}

.hub-sidebar-resize {
  position: absolute;
  right: 0;
  top: 0;
  width: 4px;
  height: 100%;
  cursor: col-resize;
  transition: background-color 0.2s ease;
}

.hub-sidebar-resize:hover {
  background: var(--color-lime);
}

.hub-main {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
}

.hub-header {
  background: var(--color-surface);
  border-bottom: 1px solid var(--color-border);
  height: 64px;
  padding: var(--spacing-lg);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.hub-content {
  flex: 1;
  overflow-y: auto;
  padding: var(--spacing-xl);
}

@media (max-width: 768px) {
  .hub-layout {
    flex-direction: column;
  }

  .hub-sidebar {
    width: 100% !important;
    height: auto;
    max-height: 50vh;
    border-right: none;
    border-bottom: 1px solid var(--color-border);
  }

  .hub-sidebar-resize {
    display: none;
  }

  .hub-main {
    flex: 1;
  }
}
</style>
