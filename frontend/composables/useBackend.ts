// Session and HTTP client for the Hub backend (/v1). Tokens live in localStorage; an expired access token is refreshed once
// (refresh tokens rotate, so concurrent callers share one refresh) before a request is retried.
const KEY = 'hub-session';

export interface Session {
  access: string; refresh: string; expiresAt: number;
  account: string; name: string; email: string;
  org: { id: string; slug: string; name: string; person?: string } | null;
}
export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) { super(message); }
}

const read = (): Session | null => { try { const v = localStorage.getItem(KEY); return v ? JSON.parse(v) : null; } catch (e) { return null; } };
const write = (s: Session | null) => { try { if (s) localStorage.setItem(KEY, JSON.stringify(s)); else localStorage.removeItem(KEY); } catch (e) { /* storage blocked: session lasts until reload */ } };

let refreshing: Promise<Session | null> | null = null;

export const useBackend = () => {
  const base = ((useRuntimeConfig().public.apiBase as string) || '').replace(/\/$/, '');

  const raw = async (method: string, path: string, body?: unknown, token?: string): Promise<any> => {
    let res: Response;
    try {
      res = await fetch(base + '/v1' + path, {
        method,
        headers: { ...(body !== undefined ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: 'Bearer ' + token } : {}) },
        body: body !== undefined ? JSON.stringify(body) : undefined
      });
    } catch (e) {
      throw new ApiError(0, 'network', 'Cannot reach the server. Check that the backend is running.');
    }
    if (res.status === 204) return null;
    const text = await res.text();
    let data: any = null;
    try { data = text ? JSON.parse(text) : null; } catch (e) { /* non-JSON body */ }
    if (!res.ok) throw new ApiError(res.status, data?.error?.code || 'error', data?.error?.message || 'Something went wrong (' + res.status + ').');
    return data;
  };

  const toSession = (t: any, email: string): Session => ({
    access: t.access_token, refresh: t.refresh_token, expiresAt: Date.now() + (t.expires_in - 30) * 1000,
    account: t.account, name: t.name, email, org: t.org || null
  });

  const refresh = (): Promise<Session | null> => {
    const s = read();
    if (!s) return Promise.resolve(null);
    if (!refreshing) {
      refreshing = raw('POST', '/auth/refresh', { refresh_token: s.refresh, org: s.org?.slug || '' })
        .then((t) => { const n = toSession(t, s.email); write(n); return n; })
        .catch(() => { write(null); return null; })
        .finally(() => { refreshing = null; });
    }
    return refreshing;
  };

  const request = async (method: string, path: string, body?: unknown): Promise<any> => {
    let s = read();
    if (s && Date.now() >= s.expiresAt) s = await refresh();
    if (!s) throw new ApiError(401, 'signed_out', 'Sign in to continue.');
    try {
      return await raw(method, path, body, s.access);
    } catch (e) {
      if (e instanceof ApiError && e.status === 401) {
        const n = await refresh();
        if (n) return raw(method, path, body, n.access);
      }
      throw e;
    }
  };

  return {
    session: read,
    signedIn: () => !!read(),
    async login(email: string, password: string, totp = ''): Promise<Session> {
      const e = email.trim();
      const t = await raw('POST', '/auth/login', { email: e, password, ...(totp ? { totp } : {}) });
      const s = toSession(t, e.toLowerCase());
      write(s);
      return s;
    },
    async signup(name: string, email: string, password: string, quiz: { question: string; answer: string }[]) {
      return raw('POST', '/auth/signup', { name: name.trim(), email: email.trim(), password, quiz });
    },
    async logout() {
      const s = read();
      write(null);
      if (s) { try { await raw('POST', '/auth/logout', { refresh_token: s.refresh }, s.access); } catch (e) { /* already signed out server-side */ } }
    },
    request,
    get: (path: string) => request('GET', path),
    post: (path: string, body?: unknown) => request('POST', path, body ?? {}),
    put: (path: string, body?: unknown) => request('PUT', path, body ?? {}),
    patch: (path: string, body?: unknown) => request('PATCH', path, body ?? {}),
    del: (path: string) => request('DELETE', path)
  };
};
