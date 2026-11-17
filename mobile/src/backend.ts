import { API_URL, isSecure } from './config';

const WAKE_TIMEOUT_MS = 60000;

// The hosted backend sleeps when idle. Asking it once at launch starts the wake-up while the web page is still loading, so the
// page's first API calls do not wait the whole cold start. Nothing about the person or the phone is sent.
export function wakeBackend(): void {
  if (!API_URL || (!isSecure(API_URL) && !__DEV__)) return;
  const ctl = new AbortController();
  const t = setTimeout(() => ctl.abort(), WAKE_TIMEOUT_MS);
  fetch(API_URL + '/healthz', { cache: 'no-store', credentials: 'omit', signal: ctl.signal })
    .catch(() => {})
    .finally(() => clearTimeout(t));
}
