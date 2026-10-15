/**
 * Widget Registration Plugin
 *
 * This plugin automatically registers all available widgets on application startup.
 * Add new widgets here by importing the component and registering it.
 */

import type { Component } from 'vue'
import { useWidgetRegistry, type Widget } from '~/composables/useWidgetRegistry'

export default defineNuxtPlugin(() => {
  const { registerWidgets } = useWidgetRegistry()

  // Define all available widgets
  const widgets: Widget[] = [
    // Communication Widgets
    {
      id: 'announcements-widget',
      name: 'Announcements',
      description: 'Latest announcements from your organization',
      icon: 'bell',
      category: 'communication',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/AnnouncementsWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:announcements:view'],
      enabled: true,
    },
    {
      id: 'forums-widget',
      name: 'Recent Forums',
      description: 'Recent forum discussions',
      icon: 'message-square',
      category: 'communication',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/ForumsWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:forums:view'],
      enabled: true,
    },
    {
      id: 'chat-widget',
      name: 'Chat',
      description: 'Private messaging',
      icon: 'message-square',
      category: 'communication',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/ChatWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:chat:view'],
      enabled: true,
    },

    // Management Widgets
    {
      id: 'projects-widget',
      name: 'Projects',
      description: 'Active projects and tasks',
      icon: 'box',
      category: 'management',
      defaultSize: 'large',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/ProjectsWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:projects:view'],
      enabled: true,
    },
    {
      id: 'tickets-widget',
      name: 'Tickets',
      description: 'Issue tracking',
      icon: 'edit',
      category: 'management',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/TicketsWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:tickets:view'],
      enabled: true,
    },
    {
      id: 'files-widget',
      name: 'Files',
      description: 'File storage and sharing',
      icon: 'folder',
      category: 'management',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/FilesWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:files:view'],
      enabled: true,
    },

    // Productivity Widgets
    {
      id: 'calendar-widget',
      name: 'Calendar',
      description: 'Event calendar',
      icon: 'calendar',
      category: 'productivity',
      defaultSize: 'large',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/CalendarWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:calendar:view'],
      enabled: true,
    },
    {
      id: 'tasks-widget',
      name: 'Tasks',
      description: 'Task management',
      icon: 'edit',
      category: 'productivity',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/TasksWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:tasks:view'],
      enabled: true,
    },

    // Admin Widgets
    {
      id: 'security-widget',
      name: 'Security',
      description: 'Security alerts and logs',
      icon: 'shield',
      category: 'admin',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/SecurityWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:security:view', 'admin'],
      enabled: true,
    },
    {
      id: 'roles-widget',
      name: 'Roles & Access',
      description: 'Role-based access control',
      icon: 'lock',
      category: 'admin',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/RolesWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:roles:view', 'admin'],
      enabled: true,
    },

    // Tools
    {
      id: 'people-widget',
      name: 'People',
      description: 'Team directory',
      icon: 'users',
      category: 'tools',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/PeopleWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:people:view'],
      enabled: true,
    },
    {
      id: 'awards-widget',
      name: 'Awards',
      description: 'Performance awards',
      icon: 'award',
      category: 'tools',
      defaultSize: 'medium',
      component: defineAsyncComponent(() =>
        import('~/components/widgets/AwardsWidget.vue').catch(() => ({ template: '<div>Widget loading...</div>' }))
      ),
      permissions: ['widget:awards:view'],
      enabled: true,
    },
  ]

  // Register all widgets
  registerWidgets(widgets)

  return {
    provide: {
      widgets,
    },
  }
})
