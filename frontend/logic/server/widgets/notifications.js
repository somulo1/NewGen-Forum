// Notifications: the bell in the header and its unread badge. New ones arrive live over the realtime socket (the server
// sends each person's notifications to their own account channel) and play the notification chime. Opening the list marks
// everything read on the server.
import { when } from './tickets.js';
import { connectRealtime } from '../realtime.js';
import { chime, loadTones } from '../../tones.js';

export const name = 'notifications';

const toItem = (n) => ({ id: n.id, text: n.text, color: n.color || '#7CC4FF', time: when(n.created_at || new Date().toISOString()), ticket: n.kind === 'ticket' ? n.ref || null : null, kind: n.kind, read: !!n.read_at });

export async function load(logic, srv) {
  // the app's sounds come from the server, so they can be tuned without a release
  const base = (srv.apiBase || (typeof window !== 'undefined' && window.__NUXT__ && window.__NUXT__.config && window.__NUXT__.config.public.apiBase) || 'http://localhost:8080');
  loadTones(base);
  const { notifications, unread } = await srv.request('GET', '/notifications');
  logic.setState({ notifs: notifications.map(toItem), notifSeen: !unread, notifUnread: unread || 0 });
  const rt = srv.rt || (srv.rt = connectRealtime(srv));
  if (srv.notifLive) return;
  srv.notifLive = true;
  rt.on('notification', (ev) => {
    const n = ev.data || {};
    if (logic.state.notifs.some((x) => x.id && x.id === n.id)) return;
    logic.setState((st) => ({ notifs: [toItem(n)].concat(st.notifs), notifSeen: false, notifUnread: (st.notifUnread || 0) + 1 }));
    chime();
    // a notification about something on screen: refresh it
    if (n.kind === 'ticket' && srv.reload.tickets) srv.reload.tickets();
    if ((n.kind === 'comm' || n.kind === 'award' || n.kind === 'approval') && srv.reload.approvals) srv.reload.approvals();
  });
}

export async function markAllRead(srv, logic) {
  await srv.request('POST', '/notifications/read');
  if (logic) logic.setState((st) => ({ notifUnread: 0, notifSeen: true, notifs: st.notifs.map((n) => ({ ...n, read: true })) }));
}

export async function clearAll(srv, logic) {
  await srv.request('DELETE', '/notifications');
  if (logic) logic.setState({ notifs: [], notifUnread: 0, notifSeen: true });
}
