// What keeps one person to one place and signed in only while active:
//  * one tab: opening Hub in a second tab closes the first, which says so and stops (like WhatsApp Web);
//  * one browser: a sign-in somewhere else ends this session (the server tells us over the realtime socket, and refuses
//    every request from the old one);
//  * activity: clicks and key presses are reported to the server, which signs the person out once the organization's
//    limit passes without any. This tab also signs out on its own at the limit, even when nothing else is being asked.
const REPORT_MS = 10000;
const MESSAGES = {
  session_replaced: 'You signed in on another browser or device, so you were signed out here.',
  device_unverified: 'This sign-in could not prove it is on this device, so it was ended. Sign in again.',
  session_expired: 'You were signed out after being inactive. Sign in again to continue.',
  phone_replaced: 'Your account is now on another phone, so this phone was signed out. To use Hub here again, sign in and confirm with the code we email you.',
  session_week: 'For your security the website signs you out once a week. Sign in again to continue.',
  access_blocked: 'An administrator blocked your access to this organization, so you were signed out.',
  access_deactivated: 'An administrator turned off your access to this organization, so you were signed out.',
  access_removed: 'You were removed from this organization, so you were signed out.'
};
export const signOutMessage = (code) => MESSAGES[code] || '';

// legacy settings stored hours
const toMinutes = (v) => { const n = Number(v) || 0; return n > 0 && n < 15 ? ({ 1: 60, 4: 300, 8: 720, 12: 720 }[n] || 1440) : n; };

export function startSessionGuard({ logic, be, goLogin }) {
  const stops = [];
  const tabId = (crypto.randomUUID && crypto.randomUUID()) || String(Math.random());
  const born = Date.now();
  let yielded = false;

  const stopAll = () => { stops.splice(0).forEach((fn) => { try { fn(); } catch (e) { /* already stopped */ } }); };
  const teardown = () => {
    try { if (logic.state.call && logic.endCall) logic.endCall(); } catch (e) { /* not in a call */ }
    try { if (logic.srv && logic.srv.rt) logic.srv.rt.close(); } catch (e) { /* closed */ }
    stopAll();
  };

  // ---- one tab ----
  if (typeof BroadcastChannel !== 'undefined') {
    const bc = new BroadcastChannel('hub-tab');
    bc.onmessage = (e) => {
      const m = e.data || {};
      if (m.type !== 'claim' || m.id === tabId || yielded) return;
      if (m.at < born || (m.at === born && m.id < tabId)) return;   // the other tab is older: it will yield, not us
      yielded = true;
      teardown();
      try { window.close(); } catch (x) { /* only tabs opened by script can close */ }
      showClosed();
    };
    bc.postMessage({ type: 'claim', id: tabId, at: born });
    stops.push(() => bc.close());
  }

  // ---- ended by the server ----
  const onEnded = (e) => { teardown(); goLogin(e && e.detail); };
  window.addEventListener('hub:signed-out', onEnded);
  stops.push(() => window.removeEventListener('hub:signed-out', onEnded));
  // another tab of this browser signed out
  const onStorage = (e) => { if (e.key === 'hub-session' && !e.newValue && !yielded) { teardown(); goLogin(''); } };
  window.addEventListener('storage', onStorage);
  stops.push(() => window.removeEventListener('storage', onStorage));

  // a sign-in elsewhere: ask the server whether this session is still the live one (a 401 ends it)
  let rtOff = null;
  const watchRealtime = () => {
    const rt = logic.srv && logic.srv.rt;
    if (!rt) { const t = setTimeout(watchRealtime, 1000); stops.push(() => clearTimeout(t)); return; }
    rtOff = rt.on('session.replaced', () => { be.request('GET', '/auth/session').catch(() => {}); });
    // blocked, deactivated or removed by an administrator: the server has already ended the session; asking it signs out here now
    const offEnded = rt.on('session.ended', () => { be.request('GET', '/auth/session').catch(() => {}); });
    stops.push(offEnded);
    // back after a dropped connection: a sign-in elsewhere meanwhile is caught here
    const reOff = rt.on('open', (ev) => { if (ev && ev.reconnect) be.request('GET', '/auth/session').catch(() => {}); });
    stops.push(() => { if (rtOff) rtOff(); reOff(); });
  };
  watchRealtime();
  // A sign-in on another browser or phone ends this session: the live connection says so at once (session.replaced). Without
  // polling, the server is also asked the moment this tab is looked at again. A 401 for an ended session signs this tab out at once.
  const probe = () => { if (!document.hidden) be.request('GET', '/auth/session').catch(() => {}); };
  document.addEventListener('visibilitychange', probe);
  window.addEventListener('focus', probe);
  stops.push(() => { document.removeEventListener('visibilitychange', probe); window.removeEventListener('focus', probe); });

  // ---- activity ----
  let lastActive = Date.now(), lastSent = 0, timer = 0, limitMs = 0;
  // the phone app has no inactivity limit: it stays signed in like a chat app (the server agrees, see auth/phone.go)
  const inApp = typeof window !== 'undefined' && !!(window.HubNative && window.HubNative.app);
  const applyLimit = (minutes) => { limitMs = minutes > 0 && !inApp ? minutes * 60000 : 0; };
  applyLimit(toMinutes((logic.state.orgSecurity || {}).session));
  const send = () => {
    timer = 0; lastSent = Date.now();
    be.request('POST', '/auth/active', {}).then((r) => { if (r && typeof r.idle_minutes === 'number') applyLimit(r.idle_minutes); }).catch(() => {});
  };
  const onActivity = () => {
    lastActive = Date.now();
    if (timer) return;
    const wait = REPORT_MS - (lastActive - lastSent);
    if (wait <= 0) send(); else timer = setTimeout(send, wait);
  };
  ['pointerdown', 'keydown', 'wheel', 'touchstart'].forEach((t) => { window.addEventListener(t, onActivity, { passive: true, capture: true }); stops.push(() => window.removeEventListener(t, onActivity, true)); });
  stops.push(() => clearTimeout(timer));
  send();   // opening the app counts

  // this tab signs out on its own when the limit passes, to the second
  const tick = setInterval(() => {
    if (!limitMs || Date.now() - lastActive < limitMs) return;
    be.logout().finally(() => { try { sessionStorage.setItem('hub-signout', 'session_expired'); } catch (e) { /* ok */ } teardown(); goLogin('session_expired'); });
  }, 1000);
  stops.push(() => clearInterval(tick));

  return teardown;
}

function showClosed() {
  const d = document.createElement('div');
  d.setAttribute('role', 'alert');
  d.style.cssText = 'position:fixed;inset:0;z-index:2147483647;background:#0F1714;color:#E8EFEA;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:12px;padding:24px;text-align:center;font-family:DM Sans,system-ui,sans-serif';
  d.innerHTML = '<strong style="font-size:22px;font-family:Space Grotesk,sans-serif">Hub is open in another tab</strong><span style="color:#9DB0A6;max-width:420px;line-height:1.5">This tab was closed because you opened Hub in a newer one. Carry on there. You can close this tab.</span>';
  document.body.appendChild(d);
}
