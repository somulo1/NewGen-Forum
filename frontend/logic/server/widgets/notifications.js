// Notifications: the bell in the header. Opening the list marks everything read on the server.
import { when } from './tickets.js';

export const name = 'notifications';

export async function load(logic, srv) {
  const { notifications, unread } = await srv.request('GET', '/notifications');
  logic.setState({
    notifs: notifications.map((n) => ({ id: n.id, text: n.text, color: n.color || '#7CC4FF', time: when(n.created_at), ticket: n.kind === 'ticket' ? n.ref || null : null })),
    notifSeen: !unread
  });
}

export async function markAllRead(srv) {
  await srv.request('POST', '/notifications/read');
}
