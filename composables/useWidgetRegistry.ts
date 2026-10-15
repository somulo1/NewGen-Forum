import type { Component } from 'vue'

export interface Widget {
  id: string
  name: string
  description: string
  icon: string
  component: Component
  category: 'communication' | 'management' | 'productivity' | 'admin' | 'tools'
  defaultSize?: 'small' | 'medium' | 'large'
  permissions?: string[]
  enabled?: boolean
}

export interface WidgetInstance {
  id: string
  widgetId: string
  position?: number
  size?: 'small' | 'medium' | 'large'
  config?: Record<string, any>
}

const widgets = new Map<string, Widget>()
const widgetInstances = new Map<string, WidgetInstance>()

export const useWidgetRegistry = () => {
  // Register a new widget
  const registerWidget = (widget: Widget) => {
    widgets.set(widget.id, widget)
  }

  // Register multiple widgets
  const registerWidgets = (newWidgets: Widget[]) => {
    newWidgets.forEach(w => registerWidget(w))
  }

  // Get a widget by ID
  const getWidget = (id: string): Widget | undefined => {
    return widgets.get(id)
  }

  // Get all widgets
  const getAllWidgets = (): Widget[] => {
    return Array.from(widgets.values())
  }

  // Get widgets by category
  const getWidgetsByCategory = (category: Widget['category']): Widget[] => {
    return Array.from(widgets.values()).filter(w => w.category === category)
  }

  // Create widget instance
  const createWidgetInstance = (
    widgetId: string,
    position?: number,
    config?: Record<string, any>
  ): WidgetInstance | null => {
    const widget = getWidget(widgetId)
    if (!widget) return null

    const instance: WidgetInstance = {
      id: `${widgetId}-${Date.now()}`,
      widgetId,
      position,
      size: widget.defaultSize || 'medium',
      config,
    }

    widgetInstances.set(instance.id, instance)
    return instance
  }

  // Get widget instance
  const getWidgetInstance = (id: string): WidgetInstance | undefined => {
    return widgetInstances.get(id)
  }

  // Get all widget instances
  const getAllWidgetInstances = (): WidgetInstance[] => {
    return Array.from(widgetInstances.values())
  }

  // Update widget instance
  const updateWidgetInstance = (id: string, updates: Partial<WidgetInstance>) => {
    const instance = widgetInstances.get(id)
    if (instance) {
      widgetInstances.set(id, { ...instance, ...updates })
    }
  }

  // Remove widget instance
  const removeWidgetInstance = (id: string) => {
    widgetInstances.delete(id)
  }

  // Check permission
  const hasPermission = (widget: Widget, userPermissions: string[]): boolean => {
    if (!widget.permissions || widget.permissions.length === 0) return true
    return widget.permissions.some(p => userPermissions.includes(p))
  }

  return {
    registerWidget,
    registerWidgets,
    getWidget,
    getAllWidgets,
    getWidgetsByCategory,
    createWidgetInstance,
    getWidgetInstance,
    getAllWidgetInstances,
    updateWidgetInstance,
    removeWidgetInstance,
    hasPermission,
  }
}
