// System notifications for what the web interface announces (new messages, notifications, calls). The page keeps receiving them
// over its live connection, exactly as in a browser; this only turns the page's notification into a phone notification.
import * as Notifications from 'expo-notifications';
import { AppState, Platform } from 'react-native';

export type Permission = 'granted' | 'denied' | 'default';

const SHOW = { shouldPlaySound: true, shouldSetBadge: false, shouldShowBanner: true, shouldShowList: true };
const QUIET = { shouldPlaySound: false, shouldSetBadge: false, shouldShowBanner: false, shouldShowList: false };

// a push that arrives while Hub is open is not shown again: the page already shows it, live
Notifications.setNotificationHandler({
  handleNotification: async (n) => (n.request.trigger && (n.request.trigger as { type?: string }).type === 'push' && AppState.currentState === 'active' ? QUIET : SHOW)
});

let channelReady = false;
async function channel() {
  if (channelReady || Platform.OS !== 'android') return;
  await Notifications.setNotificationChannelAsync('hub', { name: 'Hub', importance: Notifications.AndroidImportance.HIGH, vibrationPattern: [0, 200, 120, 200], lightColor: '#B8F36B',
    lockscreenVisibility: Notifications.AndroidNotificationVisibility.PRIVATE, showBadge: true });
  channelReady = true;
}

const map = (status: string): Permission => (status === 'granted' ? 'granted' : status === 'denied' ? 'denied' : 'default');

// the channel exists from the first start, so pushes that arrive while Hub is closed use it
export async function permission(): Promise<Permission> {
  await channel().catch(() => {});
  try { return map((await Notifications.getPermissionsAsync()).status); } catch (e) { return 'default'; }
}

export async function ask(): Promise<Permission> {
  try {
    await channel();
    return map((await Notifications.requestPermissionsAsync()).status);
  } catch (e) { return 'denied'; }
}

// the page's own id for the notification -> the system's, so the page can close it again
const shown = new Map<string, string>();

export async function show(nid: string, title: string, body: string, tag: string): Promise<void> {
  await channel();
  if (tag) {
    // one notification per chat: a newer one replaces the older
    for (const [k, v] of shown) if (k.startsWith('tag:' + tag + ':')) { Notifications.dismissNotificationAsync(v).catch(() => {}); shown.delete(k); }
  }
  const id = await Notifications.scheduleNotificationAsync({
    content: { title: title || 'Hub', body: body || '', data: { nid }, ...(Platform.OS === 'android' ? { channelId: 'hub' } : {}) } as any,
    trigger: null
  });
  shown.set((tag ? 'tag:' + tag + ':' : '') + nid, id);
}

export function close(nid: string) {
  for (const [k, v] of shown) if (k === nid || k.endsWith(':' + nid)) { Notifications.dismissNotificationAsync(v).catch(() => {}); shown.delete(k); }
}

// a tap on a notification: the page runs what it set as the notification's click
export function onTap(cb: (nid: string) => void) {
  const sub = Notifications.addNotificationResponseReceivedListener((r) => {
    const nid = (r.notification.request.content.data || {}).nid;
    if (typeof nid === 'string') cb(nid);
  });
  return () => sub.remove();
}
