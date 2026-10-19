// Endpoint helpers for the Hub backend (/v1, see backend/docs/API.md). Responses are untyped for now; widgets are wired to
// these one module at a time, replacing the sample data in logic/. Encrypted content (chat, forums, wall, polls, notes,
// calendar details) additionally needs the client-side crypto layer before it can be read or written.
export function useApi() {
  const b = useBackend();
  const q = (o: Record<string, unknown> = {}) => {
    const p = new URLSearchParams();
    Object.entries(o).forEach(([k, v]) => { if (v !== undefined && v !== null && v !== '') p.set(k, String(v)); });
    const s = p.toString();
    return s ? '?' + s : '';
  };
  return {
    me: { get: () => b.get('/me'), update: (body: unknown) => b.patch('/me', body), prefs: () => b.get('/me/prefs'), setPrefs: (body: unknown) => b.put('/me/prefs', body) },
    org: { structure: () => b.get('/org/structure'), addUnit: (body: unknown) => b.post('/org/units', body), editUnit: (id: string, body: unknown) => b.patch(`/org/units/${id}`, body) },
    people: { list: (query?: Record<string, unknown>) => b.get('/people' + q(query)), get: (id: string) => b.get(`/people/${id}`), presence: () => b.get('/presence') },
    tickets: {
      list: (query?: Record<string, unknown>) => b.get('/tickets' + q(query)),
      get: (id: string) => b.get(`/tickets/${id}`),
      create: (body: unknown) => b.post('/tickets', body),
      act: (id: string, body: unknown) => b.post(`/tickets/${id}/act`, body),
      categories: () => b.get('/tickets/categories'),
      settings: () => b.get('/tickets/settings')
    },
    projects: { list: () => b.get('/projects'), create: (body: unknown) => b.post('/projects', body), update: (id: string, body: unknown) => b.patch(`/projects/${id}`, body) },
    awards: { list: () => b.get('/awards'), give: (body: unknown) => b.post('/awards', body) },
    comms: { list: (query?: Record<string, unknown>) => b.get('/comms' + q(query)), send: (body: unknown) => b.post('/comms', body), read: (id: string) => b.post(`/comms/${id}/read`), rules: () => b.get('/comms/rules') },
    approvals: { list: () => b.get('/approvals'), flows: () => b.get('/approvals/flows'), approve: (id: string, body?: unknown) => b.post(`/approvals/${id}/approve`, body), reject: (id: string, body?: unknown) => b.post(`/approvals/${id}/reject`, body) },
    files: { inFolder: (id: string) => b.get(`/folders/${id}/files`), remove: (id: string) => b.del(`/files/${id}`), shares: () => b.get('/shares') },
    calendar: { list: (query?: Record<string, unknown>) => b.get('/calendar' + q(query)) },
    calls: { history: () => b.get('/calls/history'), start: (body: unknown) => b.post('/calls', body) },
    notifications: { list: () => b.get('/notifications'), read: (body: unknown) => b.post('/notifications/read', body) },
    security: { logs: (query?: Record<string, unknown>) => b.get('/security/logs' + q(query)), rules: () => b.get('/security/rules'), alerts: () => b.get('/security/alerts') },
    roles: { list: () => b.get('/roles') },
    settings: { get: () => b.get('/settings') }
  };
}
