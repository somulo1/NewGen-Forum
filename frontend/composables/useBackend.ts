// Session and HTTP client for the Hub backend (/v1). Tokens live in localStorage; an expired access token is refreshed once
// (refresh tokens rotate, so concurrent callers share one refresh) before a request is retried.
const KEY = 'hub-session';

export interface Session {
  access: string; refresh: string; expiresAt: number; at?: number;
  account: string; name: string; email: string;
  org: { id: string; slug: string; name: string; person?: string } | null;
}
export class ApiError extends Error {
  constructor(public status: number, public code: string, message: string) { super(message); }
}

const read = (): Session | null => { try { const v = localStorage.getItem(KEY); return v ? JSON.parse(v) : null; } catch (e) { return null; } };
const write = (s: Session | null) => { try { if (s) localStorage.setItem(KEY, JSON.stringify(s)); else localStorage.removeItem(KEY); } catch (e) { /* storage blocked: session lasts until reload */ } };

let refreshing: Promise<Session | null> | null = null;

// The server ended this session on purpose: a sign-in somewhere else, or too long without activity. The login page says why.
const ENDED = ['session_replaced', 'session_expired'];
export const SIGNOUT_KEY = 'hub-signout';
const ended = (code: string) => {
  write(null);
  try { sessionStorage.setItem(SIGNOUT_KEY, code); } catch (e) { /* the login page just will not explain */ }
  try { window.dispatchEvent(new CustomEvent('hub:signed-out', { detail: code })); } catch (e) { /* not in a browser */ }
};

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
    access: t.access_token, refresh: t.refresh_token, at: Date.now(), expiresAt: Date.now() + (t.expires_in - 30) * 1000,
    account: t.account, name: t.name, email, org: t.org || null
  });

  // One refresh at a time across every tab of this browser: a refresh token works once, and using it twice looks like theft
  // (the server ends the whole session). The token is read again inside the lock, so a second tab uses the new one.
  // maxAge: skip when this browser already rotated that recently.
  const refresh = (maxAge = 0): Promise<Session | null> => {
    if (!refreshing) {
      const go = async (): Promise<Session | null> => {
        const s = read();
        if (!s) return null;
        if (maxAge && s.at && Date.now() - s.at < maxAge) return s;
        try {
          const t = await raw('POST', '/auth/refresh', { refresh_token: s.refresh, org: s.org?.slug || '' });
          const n = toSession(t, s.email); write(n); return n;
        } catch (e) {
          if (e instanceof ApiError && ENDED.includes(e.code)) ended(e.code); else write(null);
          return null;
        }
      };
      const locks: any = typeof navigator !== 'undefined' ? (navigator as any).locks : null;
      refreshing = (locks ? locks.request('hub-refresh', go) : go()).finally(() => { refreshing = null; });
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
        if (ENDED.includes(e.code)) { ended(e.code); throw e; }
        // another request or tab may already have rotated the token: use the newest one before refreshing again
        const cur = read();
        const n = cur && cur.access !== s.access ? cur : await refresh();
        if (n) return raw(method, path, body, n.access);
      }
      throw e;
    }
  };

  // Binary transfer (encrypted file contents): PUT bytes up, or GET bytes down.
  const bytes = async (method: 'GET' | 'PUT', path: string, body?: Uint8Array): Promise<Uint8Array | null> => {
    let s = read();
    if (s && Date.now() >= s.expiresAt) s = await refresh();
    if (!s) throw new ApiError(401, 'signed_out', 'Sign in to continue.');
    const res = await fetch(base + '/v1' + path, { method, headers: { Authorization: 'Bearer ' + s.access, ...(body ? { 'Content-Type': 'application/octet-stream' } : {}) }, body });
    if (!res.ok) {
      let msg = 'Transfer failed (' + res.status + ').';
      try { msg = (await res.json()).error.message || msg; } catch (e) { /* not JSON */ }
      throw new ApiError(res.status, 'error', msg);
    }
    return method === 'GET' ? new Uint8Array(await res.arrayBuffer()) : null;
  };

  return {
    session: read,
    // a call that needs no sign-in (forgot password)
    open: (method: string, path: string, body?: unknown) => raw(method, path, body),
    // A new access and refresh token on each page change; the old ones stop working at once.
    rotate: () => refresh(4000),
    bytes,
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
