# Hub Intranet - Nuxt 3 Widget System

A comprehensive intranet platform built with Nuxt 3 featuring a pluggable widget architecture. This system is designed to be flexible, extensible, and easy to customize.

## 📋 Overview

Hub is a modern intranet system with 22+ pluggable modules/widgets including:

**Communication & Collaboration**
- Announcements - Organization news and updates
- Wall - Community discussion wall
- Forums - Discussion forums
- Chat - Private messaging
- Sessions - Meetings and standups

**Organization & Management**
- Projects - Project tracking
- Tickets - Issue management
- Files - File storage
- Secure Share - Secure link sharing
- Calendar - Event management

**People & Performance**
- People - Team directory
- Awards - Performance awards
- Performance - Performance tracking
- Roles - Role-based access control

**Structure Management**
- Departments - Department dashboards
- Sectors - Sector management
- Branches - Branch management
- Units - Unit management

**Admin & Security**
- Security - Security alerts and logs
- Memos - Internal memos
- Newsletter - Newsletter system

## 🎨 Design System

### Color Palette
The system uses a carefully curated 82-color palette with dark theme defaults:

**Primary Accent Colors:**
- `#B8F36B` - Lime (Primary)
- `#7CC4FF` - Sky Blue (Secondary)
- `#F49AC1` - Pink
- `#6EE7C8` - Teal
- `#C9B6FF` - Lavender
- `#FFB86B` - Orange
- `#FFD66B` - Amber
- `#FF9B8A` - Coral

**Dark Theme (Default):**
- Background: `#121B17`
- Surface: `#1A2520`
- Border: `#2A3832`
- Text: `#E8EFEA`
- Text Secondary: `#9DB0A6`

**Light Theme:**
- Background: `#faf9f5`
- Surface: `#ffffff`
- Border: `#e0e0e0`
- Text: `#1A2520`
- Text Secondary: `#666666`

### Typography
- **UI Font:** DM Sans (400, 500, 600, 700)
- **Display Font:** Space Grotesk (400, 500, 600, 700)
- **Monospace:** ui-monospace

### Layout
- **Canvas Size:** 1440 x 900px (responsive)
- **Sidebar Width:** 320-640px (resizable)
- **Grid System:** Auto-fit columns with 300px minimum
- **Spacing:** 4px, 8px, 12px, 16px, 24px, 32px
- **Border Radius:** 4px, 8px, 12px, 16px, 20px, 24px

## 🏗️ Architecture

### Project Structure

```
intranet-hub/
├── pages/
│   ├── index.vue                 # Home dashboard
│   ├── announcements.vue         # Announcements module
│   ├── forums.vue                # Forums module
│   └── [...slug].vue             # Catch-all for other modules
├── components/
│   ├── HubLayout.vue             # Main layout wrapper
│   ├── HubHeader.vue             # Header with search & user menu
│   ├── HubSidebar.vue            # Navigation sidebar
│   ├── Icon.vue                  # SVG icon component
│   └── Widget*.vue               # Individual widget components
├── composables/
│   ├── useWidgetRegistry.ts      # Widget registry management
│   ├── useTheme.ts               # Theme management
│   └── use*.ts                   # Other composables
├── assets/
│   └── css/
│       └── globals.css           # Global styles & CSS variables
├── layouts/
│   └── default.vue               # Default page layout
├── public/                        # Static assets
├── nuxt.config.ts                # Nuxt configuration
├── tailwind.config.ts            # Tailwind CSS configuration
├── package.json                  # Dependencies
└── README.md                      # This file
```

### Widget System

The widget registry system allows dynamic registration and management of pluggable widgets.

#### useWidgetRegistry Composable

```typescript
import { useWidgetRegistry } from '~/composables/useWidgetRegistry'

const {
  registerWidget,
  registerWidgets,
  getWidget,
  getAllWidgets,
  getWidgetsByCategory,
  createWidgetInstance,
  hasPermission,
} = useWidgetRegistry()
```

#### Widget Interface

```typescript
interface Widget {
  id: string                          // Unique widget identifier
  name: string                        // Display name
  description: string                 // Widget description
  icon: string                        // Icon name
  component: Component                // Vue component
  category: 'communication' | 'management' | 'productivity' | 'admin' | 'tools'
  defaultSize?: 'small' | 'medium' | 'large'
  permissions?: string[]              // Required permissions
  enabled?: boolean                   // Enable/disable widget
}
```

#### Creating Custom Widgets

1. **Create Component:**
```vue
<!-- components/WidgetCustom.vue -->
<template>
  <div class="widget-custom">
    <h3>{{ title }}</h3>
    <p>{{ description }}</p>
  </div>
</template>

<script setup lang="ts">
defineProps({
  config: Object,
})
</script>
```

2. **Register Widget:**
```typescript
import WidgetCustom from '~/components/WidgetCustom.vue'

const { registerWidget } = useWidgetRegistry()

registerWidget({
  id: 'custom-widget',
  name: 'Custom Widget',
  description: 'A custom widget',
  icon: 'box',
  component: WidgetCustom,
  category: 'tools',
  defaultSize: 'medium',
  permissions: ['widget:custom:view'],
})
```

### Module Pages

Each module is a full page that can contain multiple widgets. Pages are created in the `pages/` directory:

- `/` - Home dashboard (index.vue)
- `/announcements` - Announcements (announcements.vue)
- `/forums` - Forums (forums.vue)
- `/[module]` - Catch-all for other modules ([...slug].vue)

### Composables

#### useTheme
Manages light/dark theme toggling

```typescript
const { theme, setTheme, toggleTheme, initTheme } = useTheme()
```

#### useWidgetRegistry
Manages widget registration and instances (see Widget System above)

## 🚀 Getting Started

### Installation

```bash
# Install dependencies
npm install

# Development server
npm run dev

# Build for production
npm run build

# Run the production server
npm start

# Generate static site
npm run generate
```

### Development

The development server runs on `http://localhost:3000`

Features:
- Hot module replacement (HMR)
- Auto imports for composables and components
- Tailwind CSS compilation
- Development tools enabled

### Building

```bash
# Build for production
npm run build

# Preview production build
npm run preview
```

### Render Deployment

Configure the Render web service with:

- **Build Command:** `npm ci && npm run build`
- **Start Command:** `npm start`

Do not use `npm run dev` as the service start command. The production script runs
Nuxt's Nitro server, which listens on Render's assigned port and network
interface. Clear Render's build cache before the first deployment with the
locked dependencies to remove any stale Nuxt installation.

## 🎯 Key Features

### 1. **Pluggable Widget System**
- Register widgets dynamically
- Group widgets by category
- Permission-based access control
- Widget instances with configurations

### 2. **Dark/Light Theme Support**
- Toggle between themes
- Persistent theme preference
- System preference detection
- Smooth transitions

### 3. **Responsive Design**
- Mobile-first approach
- Adaptive grid layouts
- Resizable sidebar
- Flexible component sizing

### 4. **Navigation**
- Organized sidebar with module grouping
- Quick access badges for notifications
- Breadcrumb navigation
- Active page indicators

### 5. **Search & Filter**
- Global search bar
- Module-specific filtering
- Category-based organization
- Real-time filtering

## 🎨 Customization

### Colors

Edit `tailwind.config.ts` to customize colors:

```typescript
colors: {
  lime: '#B8F36B',
  sky: '#7CC4FF',
  // ... more colors
}
```

Or use CSS variables in `assets/css/globals.css`:

```css
:root {
  --color-lime: #B8F36B;
  --color-sky: #7CC4FF;
  /* ... */
}
```

### Typography

Fonts are loaded from Google Fonts. Change in `nuxt.config.ts`:

```typescript
link: [
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=YourFont:wght@400;500;600;700&display=swap'
  }
]
```

### Layout

Adjust spacing, sizing, and spacing in `tailwind.config.ts`:

```typescript
spacing: {
  xs: '4px',
  sm: '8px',
  md: '12px',
  lg: '16px',
  xl: '24px',
},
```

## 📦 Dependencies

- **Nuxt 3** - Meta framework for Vue
- **Vue 3** - Progressive JavaScript framework
- **Tailwind CSS** - Utility-first CSS framework
- **Lucide Icons** - Beautiful SVG icon library (optional)

## 🔐 Security

### Access Control

Widgets support permission-based access:

```typescript
registerWidget({
  id: 'secure-widget',
  permissions: ['widget:secure:view', 'widget:secure:edit']
})
```

Check permissions:

```typescript
const { hasPermission } = useWidgetRegistry()
const canAccess = hasPermission(widget, userPermissions)
```

### Best Practices

1. Validate user input
2. Use computed properties for data binding
3. Implement proper error handling
4. Use TypeScript for type safety
5. Follow Vue 3 composition API patterns

## 📱 Responsive Breakpoints

- **Mobile:** < 640px
- **Tablet:** 640px - 1024px
- **Desktop:** > 1024px

Media queries are built into components with mobile-first approach.

## 🔧 Configuration

### Nuxt Config

`nuxt.config.ts` includes:
- Tailwind CSS module
- Global CSS imports
- Font imports
- SSR configuration
- Development tools

### Tailwind Config

`tailwind.config.ts` provides:
- Extended color palette
- Custom spacing
- Font family definitions
- Border radius scale
- Box shadow scale
- Dark mode support

## 📚 Page Components

### Home (index.vue)
Dashboard with:
- Quick stats
- Recent forums
- Upcoming events
- Latest announcements
- Team activity
- Quick links
- Widget management modal

### Announcements (announcements.vue)
Announcement listing with:
- Priority filtering
- Author information
- Engagement metrics
- Formatted timestamps

### Forums (forums.vue)
Forum browser with:
- Category filtering
- Search functionality
- Thread information
- Activity metrics
- Author details

### Catch-all ([...slug].vue)
Dynamic module pages with:
- Auto-generated titles
- Module-specific actions
- Placeholder widgets
- Sample data areas

## 🛠️ Development Tips

### Creating New Pages

```vue
<template>
  <div class="my-module">
    <!-- Content here -->
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'default',
})
</script>
```

### Using the Widget Registry

```typescript
const { registerWidget, createWidgetInstance } = useWidgetRegistry()

// Register
registerWidget(myWidget)

// Create instance
const instance = createWidgetInstance('widget-id', 0)
```

### Styling Components

Use Tailwind classes or scoped CSS:

```vue
<style scoped>
.my-component {
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  padding: var(--spacing-lg);
}
</style>
```

## 📝 License

This project is part of the NewGen Forum Intranet Hub initiative.

## 🤝 Contributing

To add new widgets:

1. Create component in `components/`
2. Use `useWidgetRegistry` to register
3. Add to module pages or home dashboard
4. Follow design system conventions
5. Ensure responsive design

## 🆘 Support

For issues or questions about the widget system, refer to:
- `composables/useWidgetRegistry.ts` - Widget API documentation
- Component files for implementation examples
- This README for architecture overview
