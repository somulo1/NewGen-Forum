# Widget Development Guide

This guide explains how to create, register, and integrate custom widgets into the Hub Intranet system.

## Overview

Widgets are Vue 3 components that can be:
- Displayed on the home dashboard
- Shown in module pages
- Registered with metadata for discoverability
- Controlled with permissions
- Customized with configurations

## Widget Anatomy

A complete widget consists of:

1. **Component** - Vue 3 Single File Component (.vue)
2. **Registration** - Metadata and configuration
3. **Styles** - Using CSS variables and Tailwind CSS
4. **Props** - Configuration and data props

## Creating a Widget

### Step 1: Create the Component

Create a file in `components/widgets/YourWidget.vue`:

```vue
<template>
  <div class="your-widget">
    <div class="widget-header">
      <h3>{{ title }}</h3>
      <button class="widget-menu" @click="showMenu = !showMenu">
        <Icon name="more-vertical" size="16" />
      </button>
    </div>

    <div class="widget-body">
      <div v-if="loading" class="loading">Loading...</div>
      <div v-else-if="error" class="error">{{ error }}</div>
      <div v-else class="content">
        <!-- Your widget content -->
        <div v-for="item in items" :key="item.id" class="item">
          {{ item.title }}
        </div>
      </div>
    </div>

    <div class="widget-footer">
      <NuxtLink :to="actionLink" class="footer-link">
        View All →
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/Icon.vue'

interface Props {
  title?: string
  actionLink?: string
  config?: {
    maxItems?: number
    refreshInterval?: number
    [key: string]: any
  }
}

const props = withDefaults(defineProps<Props>(), {
  title: 'Widget Title',
  actionLink: '/',
  config: () => ({}),
})

const showMenu = ref(false)
const loading = ref(false)
const error = ref<string | null>(null)
const items = ref<Array<{ id: string; title: string }>>([])

// Fetch data on mount
const fetchData = async () => {
  loading.value = true
  error.value = null
  try {
    // Replace with your actual data fetching logic
    items.value = [
      { id: '1', title: 'Item 1' },
      { id: '2', title: 'Item 2' },
      { id: '3', title: 'Item 3' },
    ]
  } catch (err) {
    error.value = 'Failed to load data'
    console.error(err)
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
</script>

<style scoped>
.your-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  display: flex;
  flex-direction: column;
  transition: all 0.2s ease;
}

.your-widget:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}

.widget-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-lg);
  border-bottom: 1px solid var(--color-border);
}

.widget-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: var(--color-text);
}

.widget-menu {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: var(--spacing-sm);
  border-radius: var(--radius-sm);
  transition: all 0.2s ease;
}

.widget-menu:hover {
  background: var(--color-border);
  color: var(--color-text);
}

.widget-body {
  flex: 1;
  padding: var(--spacing-lg);
  overflow-y: auto;
}

.loading,
.error {
  text-align: center;
  padding: var(--spacing-xl);
  color: var(--color-text-secondary);
}

.error {
  color: var(--color-coral);
}

.item {
  padding: var(--spacing-md) 0;
  border-bottom: 1px solid var(--color-border);
}

.item:last-child {
  border-bottom: none;
}

.widget-footer {
  padding: var(--spacing-lg);
  border-top: 1px solid var(--color-border);
  background: var(--color-bg);
}

.footer-link {
  color: var(--color-lime);
  text-decoration: none;
  font-size: 13px;
  font-weight: 600;
  transition: opacity 0.2s ease;
}

.footer-link:hover {
  opacity: 0.8;
}
</style>
```

### Step 2: Register the Widget

Add your widget to `plugins/widgets.ts`:

```typescript
import YourWidget from '~/components/widgets/YourWidget.vue'

const widgets: Widget[] = [
  // ... existing widgets ...
  {
    id: 'your-widget-id',
    name: 'Your Widget Name',
    description: 'A brief description of what your widget does',
    icon: 'box', // Use any icon from Icon.vue
    category: 'productivity', // or 'communication', 'management', 'admin', 'tools'
    defaultSize: 'medium', // 'small', 'medium', or 'large'
    component: defineAsyncComponent(() => 
      import('~/components/widgets/YourWidget.vue')
    ),
    permissions: ['widget:your:view'], // Optional: required permissions
    enabled: true,
  },
]
```

## Widget Interface

```typescript
interface Widget {
  // Unique identifier
  id: string
  
  // Display name
  name: string
  
  // Description shown in widget picker
  description: string
  
  // Icon name from Icon.vue
  icon: string
  
  // Vue component to render
  component: Component
  
  // Widget category for organization
  category: 'communication' | 'management' | 'productivity' | 'admin' | 'tools'
  
  // Default size hint for the widget
  defaultSize?: 'small' | 'medium' | 'large'
  
  // Required permissions to access widget
  permissions?: string[]
  
  // Enable/disable the widget
  enabled?: boolean
}

interface WidgetInstance {
  // Unique instance ID
  id: string
  
  // Reference to widget definition
  widgetId: string
  
  // Position in grid
  position?: number
  
  // Current size
  size?: 'small' | 'medium' | 'large'
  
  // Widget-specific configuration
  config?: Record<string, any>
}
```

## Using Widget Registry

### Register Widgets Programmatically

```typescript
const { registerWidget } = useWidgetRegistry()

registerWidget({
  id: 'my-widget',
  name: 'My Widget',
  description: 'Does something cool',
  icon: 'star',
  category: 'tools',
  component: MyWidgetComponent,
})
```

### Create Widget Instances

```typescript
const { createWidgetInstance } = useWidgetRegistry()

const instance = createWidgetInstance(
  'my-widget',        // widget ID
  0,                  // position (optional)
  { limit: 5 }        // config (optional)
)

// instance.id can be used to reference this instance
// instance.widgetId is the original widget ID
// instance.config contains the passed configuration
```

### Get Widget Information

```typescript
const { 
  getWidget,
  getAllWidgets,
  getWidgetsByCategory,
  hasPermission,
} = useWidgetRegistry()

// Get single widget
const widget = getWidget('my-widget-id')

// Get all widgets
const all = getAllWidgets()

// Get by category
const comms = getWidgetsByCategory('communication')

// Check permissions
const canAccess = hasPermission(widget, ['widget:my:view', 'admin'])
```

## Widget Best Practices

### 1. **Use CSS Variables**

```vue
<style scoped>
.my-widget {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  color: var(--color-text);
  padding: var(--spacing-lg);
  border-radius: var(--radius-lg);
}

.my-widget:hover {
  border-color: var(--color-lime);
  box-shadow: 0 4px 12px rgba(184, 243, 107, 0.1);
}
</style>
```

### 2. **Responsive Design**

```vue
<style scoped>
.widget-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: var(--spacing-lg);
}

@media (max-width: 768px) {
  .widget-container {
    grid-template-columns: 1fr;
  }
}
</style>
```

### 3. **Handle Loading States**

```typescript
const loading = ref(false)
const error = ref<string | null>(null)
const data = ref<Data | null>(null)

const fetchData = async () => {
  loading.value = true
  error.value = null
  try {
    data.value = await api.fetchData()
  } catch (err) {
    error.value = 'Failed to load data'
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  fetchData()
})
```

### 4. **Use TypeScript**

```typescript
interface Item {
  id: string
  title: string
  status: 'active' | 'inactive'
}

interface Config {
  maxItems: number
  sortBy: 'name' | 'date'
}

interface Props {
  title?: string
  config?: Partial<Config>
}
```

### 5. **Implement Error Handling**

```typescript
const fetchData = async () => {
  try {
    data.value = await api.fetch()
  } catch (error) {
    if (error instanceof ValidationError) {
      error.value = 'Invalid input'
    } else if (error instanceof NetworkError) {
      error.value = 'Network error'
    } else {
      error.value = 'Unknown error occurred'
    }
  }
}
```

### 6. **Permission Checks**

```typescript
import { useWidgetRegistry } from '~/composables/useWidgetRegistry'

const { hasPermission } = useWidgetRegistry()
const widget = getWidget('my-widget')
const userPermissions = ['widget:my:view'] // from auth

const canAccess = computed(() => 
  hasPermission(widget, userPermissions)
)
```

## Widget Sizes

Widgets can be sized for different layouts:

```typescript
// Small widget (1/4 width on desktop)
defaultSize: 'small'

// Medium widget (1/2 width on desktop)
defaultSize: 'medium'

// Large widget (full width on desktop)
defaultSize: 'large'
```

Apply in CSS:

```css
.widget-small {
  grid-column: span 1;
}

.widget-medium {
  grid-column: span 2;
}

.widget-large {
  grid-column: span 4;
}

@media (max-width: 1024px) {
  .widget-small,
  .widget-medium {
    grid-column: span 2;
  }
  
  .widget-large {
    grid-column: span 4;
  }
}

@media (max-width: 640px) {
  .widget-small,
  .widget-medium,
  .widget-large {
    grid-column: span 1;
  }
}
```

## Widget Configuration

Pass configuration to widgets:

```typescript
// Register with default config
const widget = {
  id: 'my-widget',
  // ... other props
  config: {
    maxItems: 5,
    sortBy: 'date',
    showImages: true,
  }
}

// Instance with custom config
const instance = createWidgetInstance('my-widget', 0, {
  maxItems: 10, // Override
})
```

Access in component:

```vue
<script setup lang="ts">
interface Props {
  config?: {
    maxItems?: number
    sortBy?: string
  }
}

const props = withDefaults(defineProps<Props>(), {
  config: () => ({
    maxItems: 5,
    sortBy: 'date',
  })
})

// Use props.config.maxItems, etc.
</script>
```

## Testing Widgets

Create a test file `__tests__/YourWidget.spec.ts`:

```typescript
import { mount } from '@vue/test-utils'
import YourWidget from '~/components/widgets/YourWidget.vue'

describe('YourWidget', () => {
  it('renders', () => {
    const wrapper = mount(YourWidget)
    expect(wrapper.exists()).toBe(true)
  })

  it('accepts props', () => {
    const wrapper = mount(YourWidget, {
      props: {
        title: 'Test Title',
        config: { maxItems: 10 }
      }
    })
    expect(wrapper.find('h3').text()).toBe('Test Title')
  })

  it('loads data on mount', async () => {
    const wrapper = mount(YourWidget)
    await wrapper.vm.$nextTick()
    expect(wrapper.find('.content').exists()).toBe(true)
  })
})
```

## Publishing Your Widget

1. **Follow naming conventions**: `YourWidgetName.vue`
2. **Add to `plugins/widgets.ts`**: Register with metadata
3. **Include documentation**: Add prop descriptions
4. **Style consistently**: Use CSS variables and theme colors
5. **Handle responsive design**: Test on mobile and desktop
6. **Add error handling**: Graceful failure modes
7. **Test permissions**: Verify access control works

## Icons Available

See `components/Icon.vue` for the full list of available icons:

- home, bell, sun, moon, user, settings, help-circle, log-out
- search, folder, message-square, users, calendar, file, award
- lock, shield, git-branch, box, edit, trash-2, more-vertical
- arrow-right, check, x

And many more. Add new icon paths to `Icon.vue` as needed.

## Troubleshooting

### Widget not appearing
- Check if it's registered in `plugins/widgets.ts`
- Verify the component path is correct
- Check browser console for import errors

### Styling not applied
- Use `var(--color-*)` CSS variables
- Check scoped styles are not too specific
- Ensure Tailwind CSS classes are spelled correctly

### Props not updating
- Use `watch` or `computed` for reactive changes
- Check that props are properly typed
- Use `withDefaults` for default values

### Permission issues
- Verify permissions in widget registration
- Check user permissions array
- Use `hasPermission` composable to debug

## Resources

- Vue 3 Composition API: https://vuejs.org/guide/extras/composition-api-faq.html
- Tailwind CSS: https://tailwindcss.com/docs
- Nuxt 3: https://nuxt.com/docs
- TypeScript: https://www.typescriptlang.org/docs/
