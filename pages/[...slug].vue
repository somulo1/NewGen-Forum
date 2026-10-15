<template>
  <div class="module-container">
    <div class="page-header">
      <div>
        <h1>{{ moduleTitle }}</h1>
        <p class="subtitle">{{ moduleDescription }}</p>
      </div>
      <button v-if="moduleActions.length > 0" class="btn btn-primary">
        <Icon :name="moduleActions[0].icon" size="18" />
        {{ moduleActions[0].label }}
      </button>
    </div>

    <div class="module-content">
      <div class="placeholder-card">
        <Icon :name="moduleIcon" size="48" />
        <h2>{{ moduleTitle }}</h2>
        <p>{{ moduleDescription }}</p>
        <p class="placeholder-hint">This module is ready to be customized with widgets and components.</p>
        <button class="btn btn-secondary">Configure Widget</button>
      </div>

      <div class="sample-widgets">
        <h3>Sample Data</h3>
        <div class="widget-grid">
          <div v-for="i in 4" :key="i" class="sample-widget">
            <div class="sample-widget-header">
              <h4>Widget {{ i }}</h4>
            </div>
            <div class="sample-widget-body">
              <p>Sample content area for widget {{ i }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

definePageMeta({
  layout: 'default',
})

const route = useRoute()

// Map paths to module info
const moduleMap: Record<string, { title: string; description: string; icon: string; actions: Array<{icon: string; label: string}> }> = {
  'wall': {
    title: 'Wall',
    description: 'Community wall for posts and updates',
    icon: 'message-square',
    actions: [{ icon: 'edit', label: 'New Post' }]
  },
  'tickets': {
    title: 'Tickets',
    description: 'Issue tracking and management',
    icon: 'edit',
    actions: [{ icon: 'edit', label: 'Create Ticket' }]
  },
  'projects': {
    title: 'Projects',
    description: 'Project tracking and management',
    icon: 'box',
    actions: [{ icon: 'box', label: 'New Project' }]
  },
  'files': {
    title: 'Files',
    description: 'File storage and sharing',
    icon: 'folder',
    actions: [{ icon: 'folder', label: 'Upload File' }]
  },
  'secure-share': {
    title: 'Secure Share',
    description: 'Secure link sharing and management',
    icon: 'lock',
    actions: [{ icon: 'lock', label: 'Create Link' }]
  },
  'sessions': {
    title: 'Sessions',
    description: 'Meetings and standups',
    icon: 'users',
    actions: [{ icon: 'calendar', label: 'Schedule Meeting' }]
  },
  'awards': {
    title: 'Awards',
    description: 'Performance awards system',
    icon: 'award',
    actions: [{ icon: 'award', label: 'Give Award' }]
  },
  'people': {
    title: 'People',
    description: 'Team directory and profiles',
    icon: 'users',
    actions: []
  },
  'chat': {
    title: 'Chat',
    description: 'Private messaging',
    icon: 'message-square',
    actions: [{ icon: 'message-square', label: 'New Chat' }]
  },
  'departments': {
    title: 'Departments',
    description: 'Department dashboards',
    icon: 'folder',
    actions: []
  },
  'sectors': {
    title: 'Sectors',
    description: 'Sector management',
    icon: 'git-branch',
    actions: []
  },
  'branches': {
    title: 'Branches',
    description: 'Branch management',
    icon: 'git-branch',
    actions: []
  },
  'units': {
    title: 'Units',
    description: 'Unit management',
    icon: 'box',
    actions: []
  },
  'admin/security': {
    title: 'Security',
    description: 'Security alerts and logs',
    icon: 'shield',
    actions: []
  },
  'admin/roles': {
    title: 'Roles & Rights',
    description: 'Roles and access control',
    icon: 'lock',
    actions: [{ icon: 'edit', label: 'New Role' }]
  },
  'admin/calendar': {
    title: 'Calendar',
    description: 'Event calendar management',
    icon: 'calendar',
    actions: [{ icon: 'calendar', label: 'Add Event' }]
  },
  'admin/memos': {
    title: 'Memos',
    description: 'Internal memos',
    icon: 'file',
    actions: [{ icon: 'edit', label: 'Write Memo' }]
  },
  'admin/newsletter': {
    title: 'Newsletter',
    description: 'Newsletter system',
    icon: 'file',
    actions: [{ icon: 'edit', label: 'Send Newsletter' }]
  },
  'admin/performance': {
    title: 'Performance',
    description: 'Performance tracking',
    icon: 'award',
    actions: []
  }
}

const slug = computed(() => {
  const routeSlug = route.params.slug
  if (Array.isArray(routeSlug)) {
    return routeSlug.join('/')
  }
  return routeSlug || ''
})

const moduleData = computed(() => {
  return moduleMap[slug.value] || {
    title: slug.value.split('/').map(s => s.charAt(0).toUpperCase() + s.slice(1)).join(' '),
    description: 'Module content area',
    icon: 'box',
    actions: []
  }
})

const moduleTitle = computed(() => moduleData.value.title)
const moduleDescription = computed(() => moduleData.value.description)
const moduleIcon = computed(() => moduleData.value.icon)
const moduleActions = computed(() => moduleData.value.actions)
</script>

<style scoped>
.module-container {
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
  color: var(--color-text);
}

.subtitle {
  color: var(--color-text-secondary);
  font-size: 14px;
  margin: var(--spacing-sm) 0 0 0;
}

.module-content {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xl);
}

.placeholder-card {
  background: var(--color-surface);
  border: 2px dashed var(--color-border);
  border-radius: var(--radius-lg);
  padding: var(--spacing-xl);
  text-align: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--spacing-lg);
  color: var(--color-text-secondary);
}

.placeholder-card :deep(svg) {
  opacity: 0.5;
}

.placeholder-card h2 {
  margin: 0;
  color: var(--color-text);
}

.placeholder-card p {
  margin: 0;
  font-size: 14px;
}

.placeholder-hint {
  max-width: 400px;
}

.sample-widgets {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.sample-widgets h3 {
  margin: 0;
  color: var(--color-text);
  font-size: 16px;
}

.widget-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: var(--spacing-lg);
}

.sample-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
}

.sample-widget-header {
  padding: var(--spacing-lg);
  background: var(--color-border);
  border-bottom: 1px solid var(--color-border);
}

.sample-widget-header h4 {
  margin: 0;
  color: var(--color-text);
  font-size: 14px;
}

.sample-widget-body {
  padding: var(--spacing-lg);
  min-height: 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.sample-widget-body p {
  margin: 0;
  color: var(--color-text-secondary);
  font-size: 13px;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: var(--spacing-lg);
  }

  .widget-grid {
    grid-template-columns: 1fr;
  }
}
</style>
