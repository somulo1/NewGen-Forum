// Browser (desktop/phone) notifications, like WhatsApp Web: when Hub is open but you are in another tab or app, a new
// message, notification or call pops up from the operating system, and the tab title shows how many arrived ("(3) Hub").
// Uses the browser Notification API, so it works while a Hub tab is open (even in the background). Notifications while
// every Hub tab is closed need Web Push (a service worker plus a push service on the server); that is a later step.
const PREF = 'hub.desktopNotify';   // 'on' | 'off' (this browser only)
let missed = 0;
let baseTitle = '';
let listening = false;

export const supported = () => typeof window !== 'undefined' && 'Notification' in window;
export const permission = () => (supported() ? Notification.permission : 'unsupported');   // 'default' | 'granted' | 'denied'
export function enabled() {
  if (permission() !== 'granted') return false;
  try { return localStorage.getItem(PREF) !== 'off'; } catch (e) { return true; }
}
export function setEnabled(on) { try { localStorage.setItem(PREF, on ? 'on' : 'off'); } catch (e) { /* private window */ } }

// Browsers only allow the question from a click; call this from a button.
export async function ask() {
  if (!supported()) return 'unsupported';
  const r = await Notification.requestPermission();
  if (r === 'granted') setEnabled(true);
  return r;
}

const away = () => typeof document !== 'undefined' && (document.hidden || !document.hasFocus());

function watchFocus() {
  if (listening || typeof document === 'undefined') return;
  listening = true;
  const back = () => { if (!away()) { missed = 0; if (baseTitle) document.title = baseTitle; } };
  document.addEventListener('visibilitychange', back);
  window.addEventListener('focus', back);
}

// the tab title counts what arrived while you were away
function bumpTitle() {
  if (typeof document === 'undefined') return;
  watchFocus();
  baseTitle = (document.title || 'Hub').replace(/^\(\d+\+?\)\s*/, '');
  missed++;
  document.title = '(' + (missed > 99 ? '99+' : missed) + ') ' + baseTitle;
}

/**
 * Shows a system notification, only when the person is not looking at Hub right now.
 * @param {{title: string, body?: string, tag?: string, icon?: string, onClick?: Function, urgent?: boolean, always?: boolean}} n
 *  tag: one notification per chat (a newer message replaces the older one); urgent: stays until clicked (calls);
 *  always: show even when Hub is in front (e.g. a call while another app has focus is handled by `away` anyway).
 */
export function show(n) {
  if (!n || (!n.always && !away())) return null;
  bumpTitle();
  if (!enabled()) return null;
  try {
    const note = new Notification(n.title, { body: (n.body || '').slice(0, 180), tag: n.tag, renotify: !!n.tag, icon: n.icon || '/favicon.ico', badge: '/favicon.ico', requireInteraction: !!n.urgent, silent: false });
    note.onclick = () => { try { window.focus(); } catch (e) { /* ignore */ } if (n.onClick) n.onClick(); note.close(); };
    if (!n.urgent) setTimeout(() => note.close(), 8000);
    return note;
  } catch (e) { return null; }   // some mobile browsers only allow notifications from a service worker
}

// a short, safe preview of a chat message
export function preview(m) {
  if (!m) return '';
  if (m.text) return String(m.text).replace(/\*\*/g, '').replace(/\s+/g, ' ').trim();
  if (m.file) return 'Sent a file';
  return 'New message';
}
