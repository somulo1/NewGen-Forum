// Typed API client for when the backend exists (see docs/03-API-CONTRACT.md).
// Nothing calls this yet — the app runs on in-memory sample data in logic/HubLogic.js.
// Set NUXT_PUBLIC_API_BASE (e.g. https://hub.example.com) and start replacing sample data module by module.
import type {
  User, Org, Forum, ForumMessage, DirectConversation, DirectMessage, Post, CalendarEvent, Ticket,
  TicketCategory, Project, Task, Folder, FileRef, Communication, Award, ApprovalRequest, ApprovalFlow,
  Role, RightOverride, AuditLog, DetectionRule, SecurityAlert, Notification, Level, UserId
} from '~/types/hub';

export interface Page<T> { data: T[]; total: number }
export interface ListQuery { page?: number; pageSize?: number; sort?: string; filter?: Record<string, string> }
export interface Pending { approvalRequestId: string }

function toQuery(q: ListQuery = {}) {
  const out: Record<string, string | number> = {};
  if (q.page != null) out.page = q.page;
  if (q.pageSize != null) out.pageSize = q.pageSize;
  if (q.sort) out.sort = q.sort;
  Object.entries(q.filter || {}).forEach(([k, v]) => { out[`filter[${k}]`] = v; });
  return out;
}

export function useApi() {
  const base = (useRuntimeConfig().public.apiBase as string) || '';
  const api = $fetch.create({ baseURL: base, credentials: 'include' });
  const get = <T>(url: string, query?: Record<string, unknown>) => api<T>(url, { query });
  const send = <T>(method: 'POST' | 'PATCH' | 'DELETE', url: string, body?: unknown) => api<T>(url, { method, body: body as Record<string, unknown> });

  return {
    me: { get: () => get<User>('/api/me'), update: (b: Partial<User>) => send<User>('PATCH', '/api/me', b) },
    org: {
      get: () => get<Org>('/api/org'),
      setWidgets: (widgets: Org['widgets']) => send<Org>('PATCH', '/api/org/widgets', { widgets }),
      setWidgetConfig: (id: string, cfg: unknown) => send<Org>('PATCH', `/api/org/widgets/${id}/config`, cfg),
      setCommRule: (level: Level, type: 'Memo' | 'Announcement', rule: unknown) => send<Org>('PATCH', `/api/org/comm-rules/${level}/${type}`, rule),
      dashboard: (level: Level, id: string, from?: string, to?: string) => get<Record<string, unknown>>(`/api/org/${level}/${id}/dashboard`, { from, to })
    },
    people: { list: (q?: ListQuery) => get<Page<User>>('/api/people', toQuery(q)), get: (id: UserId) => get<User>(`/api/people/${id}`) },
    forums: {
      list: () => get<Forum[]>('/api/forums'),
      messages: (id: string) => get<ForumMessage[]>(`/api/forums/${id}/messages`),
      post: (id: string, b: { text?: string; fileId?: string }) => send<ForumMessage>('POST', `/api/forums/${id}/messages`, b)
    },
    dms: {
      list: () => get<DirectConversation[]>('/api/dms'),
      open: (userId: UserId) => send<DirectConversation>('POST', '/api/dms', { userId }),
      messages: (id: string) => get<DirectMessage[]>(`/api/dms/${id}/messages`),
      post: (id: string, b: { text?: string; fileId?: string }) => send<DirectMessage>('POST', `/api/dms/${id}/messages`, b),
      read: (id: string) => send<void>('POST', `/api/dms/${id}/read`)
    },
    wall: {
      list: (q?: ListQuery) => get<Page<Post>>('/api/posts', toQuery(q)),
      create: (b: Partial<Post>) => send<Post>('POST', '/api/posts', b),
      share: (id: string, b: { forumId?: string; app?: string }) => send<Post | Pending>('POST', `/api/posts/${id}/share`, b)
    },
    calendar: {
      list: (from: string, to: string) => get<CalendarEvent[]>('/api/events', { from, to }),
      save: (e: Partial<CalendarEvent>) => (e.id ? send<CalendarEvent>('PATCH', `/api/events/${e.id}`, e) : send<CalendarEvent>('POST', '/api/events', e)),
      remove: (id: string) => send<void>('DELETE', `/api/events/${id}`)
    },
    calls: { start: (b: { mode: 'voice' | 'video'; title: string; participantIds: UserId[]; private: boolean; eventId?: string }) => send<{ id: string; token: string }>('POST', '/api/calls', b) },
    tickets: {
      list: (q?: ListQuery) => get<Page<Ticket>>('/api/tickets', toQuery(q)),
      get: (id: string) => get<Ticket>(`/api/tickets/${id}`),
      create: (b: Partial<Ticket>) => send<Ticket>('POST', '/api/tickets', b),
      action: (id: string, action: 'comments' | 'files' | 'assign' | 'start' | 'hold' | 'resume' | 'solve' | 'close' | 'reopen', b?: unknown) => send<Ticket | Pending>('POST', `/api/tickets/${id}/${action}`, b),
      escalationTargets: (ticketId: string) => get<{ level: Level | 'desk'; entities: { id: string; name: string }[] }[]>('/api/escalation/targets', { ticketId }),
      escalationCategories: (level: string, id: string) => get<TicketCategory[]>('/api/escalation/categories', { level, id }),
      escalationAssignees: (level: string, id: string, categoryId: string) => get<User[]>('/api/escalation/assignees', { level, id, categoryId }),
      escalate: (id: string, b: { level: string; id: string; categoryId: string; assigneeId: UserId; reason?: string }) => send<Ticket>('POST', `/api/tickets/${id}/escalate`, b)
    },
    projects: {
      list: () => get<Project[]>('/api/projects'),
      get: (id: string) => get<Project & { tasks: Task[] }>(`/api/projects/${id}`),
      create: (b: Partial<Project>) => send<Project | Pending>('POST', '/api/projects', b),
      moveTask: (taskId: string, column: Task['column']) => send<Task>('POST', `/api/tasks/${taskId}/move`, { column })
    },
    files: { folders: () => get<Folder[]>('/api/folders'), list: (folderId: string) => get<FileRef[]>(`/api/folders/${folderId}/files`) },
    comms: { list: (q?: ListQuery) => get<Page<Communication>>('/api/comms', toQuery(q)), send: (b: Partial<Communication>) => send<Communication | Pending>('POST', '/api/comms', b) },
    awards: { list: () => get<Award[]>('/api/awards'), give: (b: FormData) => api<Award | Pending>('/api/awards', { method: 'POST', body: b }) },
    approvals: {
      list: (status?: string) => get<ApprovalRequest[]>('/api/approvals', { status }),
      decide: (id: string, decision: 'approve' | 'reject', note?: string) => send<ApprovalRequest>('POST', `/api/approvals/${id}/${decision}`, { note }),
      flows: () => get<ApprovalFlow[]>('/api/approval-flows')
    },
    rights: { roles: () => get<Role[]>('/api/roles'), list: () => get<RightOverride[]>('/api/rights'), change: (b: RightOverride) => send<Pending>('PATCH', '/api/rights', b) },
    security: {
      logs: (q?: ListQuery) => get<Page<AuditLog>>('/api/audit-logs', toQuery(q)),
      rules: () => get<DetectionRule[]>('/api/security/rules'),
      alerts: () => get<SecurityAlert[]>('/api/security/alerts')
    },
    notifications: { list: () => get<Notification[]>('/api/notifications'), read: (ids: string[]) => send<void>('POST', '/api/notifications/read', { ids }) }
  };
}
