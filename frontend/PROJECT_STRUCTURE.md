# Project Structure Overview

## Directory Tree

```
intranet-hub/
├── components/
│   ├── HubLayout.vue              # Main layout container (sidebar + header + content)
│   ├── HubHeader.vue              # Top header with logo, search, theme toggle, user menu
│   ├── HubSidebar.vue             # Left sidebar with navigation
│   ├── Icon.vue                   # SVG icon component
│   └── widgets/
│       ├── ExampleWidget.vue      # Example widget template
│       ├── AnnouncementsWidget.vue # (To be created)
│       ├── ForumsWidget.vue       # (To be created)
│       ├── ChatWidget.vue         # (To be created)
│       ├── ProjectsWidget.vue     # (To be created)
│       ├── TicketsWidget.vue      # (To be created)
│       ├── FilesWidget.vue        # (To be created)
│       ├── CalendarWidget.vue     # (To be created)
│       ├── TasksWidget.vue        # (To be created)
│       ├── SecurityWidget.vue     # (To be created)
│       ├── RolesWidget.vue        # (To be created)
│       ├── PeopleWidget.vue       # (To be created)
│       └── AwardsWidget.vue       # (To be created)
│
├── composables/
│   ├── useWidgetRegistry.ts       # Widget registration and management
│   ├── useTheme.ts                # Light/dark theme management
│   └── use*.ts                    # Other composables (create as needed)
│
├── layouts/
│   └── default.vue                # Default page layout
│
├── pages/
│   ├── index.vue                  # Home dashboard with widgets
│   ├── announcements.vue          # Announcements module
│   ├── forums.vue                 # Forums module
│   └── [...slug].vue              # Catch-all for other modules
│
├── plugins/
│   └── widgets.ts                 # Widget registration plugin
│
├── assets/
│   └── css/
│       └── globals.css            # Global styles and CSS variables
│
├── public/                        # Static assets (favicon, logos, etc)
│   └── (empty initially)
│
├── app.vue                        # Root component
├── nuxt.config.ts                 # Nuxt 3 configuration
├── tailwind.config.ts             # Tailwind CSS configuration
├── tsconfig.json                  # TypeScript configuration
├── package.json                   # Dependencies and scripts
├── .npmrc                         # NPM configuration
├── .gitignore                     # Git ignore rules
├── .env.example                   # Environment variables template
│
├── README.md                      # Main project README
├── WIDGET_GUIDE.md                # Comprehensive widget development guide
└── PROJECT_STRUCTURE.md           # This file
```

## File Descriptions

### Core Configuration Files

**package.json**
- Project metadata and dependencies
- Scripts for dev, build, preview, generate
- Dependencies: nuxt, vue, tailwindcss, lucide-icons

**nuxt.config.ts**
- Nuxt 3 configuration
- Tailwind CSS module setup
- Font imports from Google Fonts
- Head meta tags
- SSR configuration

**tailwind.config.ts**
- Extended color palette (82 colors)
- Custom spacing, border radius, shadows
- Dark mode configuration
- Font family definitions

**tsconfig.json**
- TypeScript compiler options
- Path aliases (@/, ~/)
- Vue 3 settings

**.env.example**
- Template for environment variables
- API configuration
- Feature flags
- Theme defaults

### Layout & Component System

**app.vue**
- Root component
- Theme initialization
- Layout wrapper

**HubLayout.vue**
- Main application layout
- Resizable sidebar (320-640px)
- Header and content areas
- Responsive design

**HubHeader.vue**
- Logo with gradient
- Global search bar
- Notification bell
- Theme toggle
- User menu dropdown

**HubSidebar.vue**
- Module navigation (22 modules)
- Organization structure section
- Admin section
- Active state indicators
- Badge counts

**Icon.vue**
- 20+ SVG icons
- Configurable size
- Custom icon definitions
- Lightweight implementation

### Pages & Modules

**pages/index.vue (Home Dashboard)**
- Quick stats section
- Recent forums widget
- Upcoming events
- Latest announcements
- Team activity feed
- Quick links
- Add widget modal

**pages/announcements.vue**
- Announcement listing
- Priority filtering
- Author information
- Engagement metrics
- Responsive layout

**pages/forums.vue**
- Forum browser
- Category filtering
- Search functionality
- Thread information
- Activity metrics

**pages/[...slug].vue**
- Dynamic module pages
- Catch-all routing
- Placeholder widgets
- Module-specific actions
- Sample data areas

### Widget System

**composables/useWidgetRegistry.ts**
- Widget registration
- Instance creation
- Category filtering
- Permission checking
- Get/update operations

**plugins/widgets.ts**
- Central widget registration
- 12+ widget definitions
- Category organization
- Permission setup
- Async component loading

**components/widgets/ExampleWidget.vue**
- Template for new widgets
- Header with actions
- Loading/error states
- Body and footer sections
- Responsive styling

### Styling & Theme

**assets/css/globals.css**
- CSS variables for colors
- Dark/light theme support
- Base element styling
- Utility classes
- Scrollbar customization
- Typography defaults

**Design System Details:**
- 82 unique colors
- 6 accent colors (lime, sky, pink, teal, lavender, orange)
- 9 dark theme colors
- 6 spacing levels
- 5 border radius sizes
- 5 box shadow levels

### Documentation

**README.md** (Main)
- Project overview
- 22 modules listed
- Design system documentation
- Architecture explanation
- Widget system basics
- Getting started guide
- Customization options
- Feature highlights

**WIDGET_GUIDE.md**
- Complete widget development guide
- Step-by-step widget creation
- Widget interface documentation
- Best practices (6 sections)
- Widget sizes and configuration
- Testing examples
- Troubleshooting guide
- Available icons

**PROJECT_STRUCTURE.md**
- This file
- Complete directory tree
- File descriptions
- Organization details

## Module Organization

### Communication (5 modules)
- Announcements
- Wall
- Forums
- Chat
- Sessions

### Management (4 modules)
- Projects
- Tickets
- Files
- Secure Share

### Productivity (3 modules)
- Calendar
- Memos
- Newsletter

### Organization (4 modules)
- Departments
- Sectors
- Branches
- Units

### People & Performance (2 modules)
- People
- Awards

### Admin & Security (3 modules)
- Security
- Roles & Rights
- Performance

## Color Palette

### Primary Colors
- Lime: #B8F36B
- Sky: #7CC4FF
- Pink: #F49AC1
- Teal: #6EE7C8
- Lavender: #C9B6FF
- Orange: #FFB86B
- Amber: #FFD66B
- Coral: #FF9B8A

### Dark Theme (Default)
- Background: #121B17
- Surface: #1A2520
- Border: #2A3832
- Text: #E8EFEA
- Text Secondary: #9DB0A6

### Light Theme
- Background: #faf9f5
- Surface: #ffffff
- Border: #e0e0e0
- Text: #1A2520
- Text Secondary: #666666

## Responsive Breakpoints

- **Mobile:** < 640px
- **Tablet:** 640px - 1024px
- **Desktop:** > 1024px

Mobile-first approach with media queries for larger screens.

## Key Features Implemented

✅ **Pluggable Widget System**
- Dynamic widget registration
- Widget instances with config
- Category organization
- Permission-based access

✅ **Theme Support**
- Dark theme (default)
- Light theme
- Toggle functionality
- Persistent preferences
- System preference detection

✅ **Responsive Design**
- Mobile-friendly layouts
- Adaptive grids
- Resizable sidebar
- Touch-friendly navigation

✅ **Navigation**
- Sidebar with modules
- Section grouping
- Badge indicators
- Active state styling

✅ **Component System**
- Reusable components
- Composition API
- TypeScript support
- CSS variables for theming

✅ **Search & Filter**
- Global search bar
- Module-specific filtering
- Real-time filtering
- Category-based search

## Dependencies

```json
{
  "nuxt": "^3.8.0",
  "vue": "^3.3.0",
  "@nuxtjs/tailwindcss": "^6.10.0",
  "tailwindcss": "^3.3.0",
  "postcss": "^8.4.0"
}
```

## Getting Started

1. **Install dependencies:**
   ```bash
   npm install
   ```

2. **Start development server:**
   ```bash
   npm run dev
   ```

3. **Open browser:**
   ```
   http://localhost:3000
   ```

4. **Build for production:**
   ```bash
   npm run build
   ```

## Next Steps

### To Add New Widgets:
1. Create component in `components/widgets/`
2. Register in `plugins/widgets.ts`
3. Add to pages as needed
4. Follow WIDGET_GUIDE.md for best practices

### To Add New Modules:
1. Create page in `pages/`
2. Add navigation in `HubSidebar.vue`
3. Define module configuration
4. Implement module layout

### To Customize Design:
1. Edit color variables in `tailwind.config.ts`
2. Update CSS variables in `globals.css`
3. Adjust spacing and sizing
4. Modify typography

### To Extend Functionality:
1. Create new composables in `composables/`
2. Add utility functions as needed
3. Implement new widgets
4. Add plugins for features

## File Size Optimization

- **Component lazy loading** via `defineAsyncComponent()`
- **CSS variable system** for efficient theming
- **Tailwind CSS** for minimal CSS output
- **Composition API** for tree-shaking

## Browser Support

- Modern browsers (Chrome, Firefox, Safari, Edge)
- Vue 3 support required
- ES2020+ JavaScript
- CSS Grid and Flexbox support

## Performance Considerations

- Sidebar width stored in localStorage
- Theme preference persisted
- Async widget component loading
- Computed properties for reactive data
- Watch for theme changes
- Minimal re-renders with proper scoping

## Security Notes

- XSS protection via Vue's built-in escaping
- CSRF protection headers recommended
- Permission-based widget access
- User input validation required
- API security best practices

## Maintenance

- Update Nuxt 3 regularly
- Keep dependencies current
- Monitor security advisories
- Test widget compatibility
- Update documentation
- Track performance metrics

---

**Last Updated:** December 15, 2024
**Version:** 1.0.0
**License:** Internal Use
