// A copy of the Hub sign-in in the phone's secure storage (Keychain / Keystore), next to the phone's key. The web page keeps its
// sign-in in the web view's storage, which a cache cleaner, an Android storage clean-up or a web view update can erase; the app
// puts this copy back before the page starts, so none of those ever signs the person out. Signing out deletes it.
import * as SecureStore from 'expo-secure-store';

const SLOT = 'hub.session.v1';
const OPTS: SecureStore.SecureStoreOptions = { keychainAccessible: SecureStore.WHEN_UNLOCKED_THIS_DEVICE_ONLY };
const MAX = 8 * 1024;

export async function load(): Promise<string> {
  try {
    const v = await SecureStore.getItemAsync(SLOT, OPTS);
    return v && v.length <= MAX ? v : '';
  } catch (e) {
    return '';
  }
}

// the page's current sign-in (JSON), or '' when it signed out
export async function keep(json: string): Promise<void> {
  try {
    if (json && json.length <= MAX && json.startsWith('{')) await SecureStore.setItemAsync(SLOT, json, OPTS);
    else if (!json) await SecureStore.deleteItemAsync(SLOT, OPTS);
  } catch (e) { /* kept again with the next change */ }
}
