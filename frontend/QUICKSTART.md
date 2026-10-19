# Quick Start Guide

Get the NewGen Hub system up and running in minutes.

## Prerequisites

- Node.js 16+ (LTS recommended)
- npm 8+ or yarn/pnpm
- Modern web browser
- Git (optional, for version control)

## Installation

### 1. Install Dependencies

```bash
cd intranet-hub
npm install
```

This installs:
- Nuxt 3
- Vue 3
- Tailwind CSS
- PostCSS

### 2. Start Development Server

```bash
npm run dev
```

Output:
```
[nitro] Server built in 235ms
➜ Local: http://localhost:3000/
```

### 3. Open in Browser

Navigate to `http://localhost:3000`

You should see:
- **Home Dashboard** with widgets
- **Dark theme** by default
- **Responsive layout** adapting to your screen size

## Using the Application

### Navigation

1. **Sidebar** on the left contains:
   - 22 modules grouped by category
   - Active page indicator
   - Notification badges

2. **Header** at the top has:
   - Hub logo
   - Search bar
   - Notifications (3)
   - Theme toggle
   - User menu

3. **Main Content** in the center:
   - Module-specific content
   - Widgets and components
   - Interactive elements

### Switching Modules

Click any module in the sidebar:
- **Home** (/) - Dashboard with quick stats
- **Announcements** - Latest news
- **Forums** - Discussion boards
- **Tickets** - Issue tracking
- **Projects** - Project management
- And 17 more...

### Theme Toggle

Click the **sun/moon icon** in the header to:
- Switch between dark and light themes
- Theme preference is saved automatically
- Smooth transition between themes

### Search

Use the **search bar** to:
- Search across modules
- Filter by keyword
- Quick access to content

## File Structure

```
intranet-hub/
├── components/           # Vue components
├── pages/               # Page components
├── composables/         # Vue composables
├── layouts/             # Layout components
├── plugins/             # Nuxt plugins
├── assets/              # CSS and styles
├── public/              # Static files
├── nuxt.config.ts       # Configuration
├── tailwind.config.ts   # Tailwind config
└── package.json         # Dependencies
```

## Building Custom Widgets

### Example: Create a Simple Widget

1. **Create component:**
```vue
<!-- components/widgets/MyWidget.vue -->
<template>
  <div class="my-widget">
    <h3>{{ title }}</h3>
    <p>My custom widget content</p>
  </div>
</template>

<script setup lang="ts">
defineProps({
  title: String,
})
</script>

<style scoped>
.my-widget {
  padding: var(--spacing-lg);
}
</style>
```

2. **Register widget:**
Edit `plugins/widgets.ts`:
```typescript
{
  id: 'my-widget',
  name: 'My Widget',
  description: 'My custom widget',
  icon: 'box',
  category: 'tools',
  component: defineAsyncComponent(() =>
    import('~/components/widgets/MyWidget.vue')
  ),
  enabled: true,
}
```

3. **Use on home page:**
Edit `pages/index.vue` and add to widgets-grid.

## Common Tasks

### Add a New Module

1. Create `pages/my-module.vue`
2. Add to sidebar in `HubSidebar.vue`
3. Implement module UI

### Change Colors

Edit `tailwind.config.ts`:
```typescript
colors: {
  lime: '#YOUR_COLOR_HEX',
  // ... other colors
}
```

### Change Fonts

Edit `nuxt.config.ts`:
```typescript
link: [
  {
    rel: 'stylesheet',
    href: 'https://fonts.googleapis.com/css2?family=YourFont&display=swap'
  }
]
```

And `tailwind.config.ts`:
```typescript
fontFamily: {
  sans: ['Your Font', 'sans-serif'],
}
```

### Add Navigation Links

Edit `HubSidebar.vue` - modify arrays:
- `modules` - Communication modules
- `organizationItems` - Organization modules
- `adminItems` - Admin modules

### Create a Data Module

1. Create API utility in `composables/` or service
2. Use in widget component with `onMounted`
3. Handle loading/error states
4. Display fetched data

## Production Build

### Build for Production

```bash
npm run build
```

Creates optimized build in `.output/` or `dist/`

### Preview Production Build

```bash
npm run preview
```

Tests production build locally before deployment.

### Generate Static Site

```bash
npm run generate
```

Creates static HTML files for static hosting.

## Environment Variables

Create `.env.local` (copy from `.env.example`):

```bash
NUXT_PUBLIC_API_BASE_URL=https://api.example.com
NUXT_PUBLIC_APP_NAME=NewGen Hub
NUXT_PUBLIC_DEFAULT_THEME=dark
```

Access in components:
```typescript
const apiUrl = useRuntimeConfig().public.apiBaseUrl
```

## Debugging

### Browser DevTools

1. Open DevTools (F12)
2. Vue extension for DevTools (install if needed)
3. Inspect components
4. Check console for errors

### Nuxt DevTools

Enabled by default in development:
- Open with right-click → "Inspect with Nuxt DevTools"
- Or press `Shift + Alt + I`

### Common Issues

**Port 3000 already in use:**
```bash
npm run dev -- -p 3001
```

**Dependencies not installing:**
```bash
rm -rf node_modules package-lock.json
npm install
```

**Styles not loading:**
- Check Tailwind CSS config
- Restart dev server
- Clear browser cache

## TypeScript Support

Full TypeScript support included. Create typed components:

```typescript
interface Props {
  title: string
  count?: number
}

interface Item {
  id: string
  label: string
}

const props = withDefaults(defineProps<Props>(), {
  count: 0,
})

const items = ref<Item[]>([])
```

## Performance Tips

1. **Use lazy loading** for heavy components
2. **Optimize images** before importing
3. **Minimize re-renders** with proper scoping
4. **Use computed** for expensive operations
5. **Implement pagination** for large lists

## Deployment Options

### Vercel (Recommended)

1. Push to GitHub
2. Connect Vercel
3. Auto-deploying

### Netlify

1. Build locally: `npm run generate`
2. Push `dist/` folder
3. Configure build command

### Self-Hosted

1. Build: `npm run build`
2. Start server: `node .output/server/index.mjs`
3. Requires Node.js on server

### Docker

Create `Dockerfile`:
```dockerfile
FROM node:18-alpine
WORKDIR /app
COPY . .
RUN npm install && npm run build
CMD ["node", ".output/server/index.mjs"]
```

## Next Steps

1. ✅ Explore the home dashboard
2. ✅ Navigate through different modules
3. ✅ Try toggling the theme
4. ✅ Read WIDGET_GUIDE.md for custom widgets
5. ✅ Customize colors and typography
6. ✅ Add your own modules
7. ✅ Build your first widget
8. ✅ Deploy to production

## Resources

- **Documentation:** README.md, WIDGET_GUIDE.md, PROJECT_STRUCTURE.md
- **Vue 3 Docs:** https://vuejs.org/
- **Nuxt 3 Docs:** https://nuxt.com/
- **Tailwind CSS:** https://tailwindcss.com/
- **TypeScript:** https://www.typescriptlang.org/

## Support

For issues or questions:
1. Check the documentation files
2. Review the example components
3. Check browser console for errors
4. Review NODE environment and versions

## Tips & Tricks

### Sidebar Resizing
- Click and drag the right edge of sidebar
- Width saved to localStorage
- Resizable between 320-640px

### Theme Preference
- Click sun/moon icon to toggle
- Preference saved in localStorage
- Defaults to dark theme

### Module Badges
- Show notification counts
- Update in `HubSidebar.vue`
- Edit the module array

### Search Functionality
- Global search in header
- Module-specific search on module pages
- Implement with filtering logic

### Active Navigation
- Current page highlighted in sidebar
- Blue left border on active item
- Lime green text color

## Customization Examples

### Change Primary Color

`tailwind.config.ts`:
```typescript
colors: {
  primary: {
    500: '#YOUR_HEX_COLOR',
  }
}
```

### Adjust Spacing

`tailwind.config.ts`:
```typescript
spacing: {
  lg: '20px', // Change from 16px
}
```

### Add New Icon

`Icon.vue`:
```typescript
const icons: Record<string, string> = {
  'my-icon': '<path d="..."/>',
}
```

Use: `<Icon name="my-icon" />`

## Congratulations! 🎉

You're ready to build with NewGen Hub. Start by:
1. Creating your first widget
2. Adding a custom module
3. Customizing the design
4. Building your features

Happy coding!
