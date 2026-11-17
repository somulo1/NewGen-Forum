// Push notifications: the address Android (Firebase) or iOS (Apple) gives this phone, handed to the Hub page so it can register
// it with the backend for this sign-in, and the taps on pushes that open the app. While the app is open the page shows its own
// alerts, so pushes are only shown by the system when Hub is in the background or closed.
import Constants, { ExecutionEnvironment } from 'expo-constants';
import * as Notifications from 'expo-notifications';
import { Platform } from 'react-native';

export type PushAddress = { token: string; platform: 'android' | 'ios' };
export type PushData = Record<string, string>;

const supported = () => Platform.OS === 'ios' || (Platform.OS === 'android' && Constants.executionEnvironment !== ExecutionEnvironment.StoreClient);
const toAddress = (t: Notifications.DevicePushToken): PushAddress | null =>
  typeof t.data === 'string' && t.data ? { token: t.data, platform: t.type === 'ios' ? 'ios' : 'android' } : null;

// null when this phone cannot receive pushes (Expo Go on Android, notifications turned off, no Firebase set up in the build)
export async function address(): Promise<PushAddress | null> {
  if (!supported()) return null;
  try {
    if ((await Notifications.getPermissionsAsync()).status !== 'granted') return null;
    return toAddress(await Notifications.getDevicePushTokenAsync());
  } catch (e) {
    return null;
  }
}

// the platform may change the address at any time; the page registers the new one
export function onNewAddress(cb: (a: PushAddress) => void) {
  const sub = Notifications.addPushTokenListener((t) => { const a = toAddress(t); if (a) cb(a); });
  return () => sub.remove();
}

const pushData = (r: Notifications.NotificationResponse | null): PushData | null => {
  const d = (r && r.notification.request.content.data) || {};
  return typeof d.kind === 'string' ? Object.fromEntries(Object.entries(d).map(([k, v]) => [k, String(v)])) : null;
};

// a tap on a push, including the one that started the app
export function onTap(cb: (d: PushData) => void) {
  Notifications.getLastNotificationResponseAsync().then((r) => {
    const d = pushData(r);
    if (!d) return;
    Notifications.clearLastNotificationResponseAsync().catch(() => {});   // opened once, not again at the next start
    cb(d);
  }).catch(() => {});
  const sub = Notifications.addNotificationResponseReceivedListener((r) => { const d = pushData(r); if (d) cb(d); });
  return () => sub.remove();
}
