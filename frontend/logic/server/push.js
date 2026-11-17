// Push notifications in the Hub phone app. The page hands the phone's push address to the backend for this sign-in, so new
// messages, calls and notifications reach the phone while Hub is in the background or closed, and a tap on one opens what it
// was about. In a browser there is no HubNative and nothing here runs.
const OPEN_TRIES = 40;
const OPEN_EVERY_MS = 500;

const native = () => (typeof window !== 'undefined' && window.HubNative && window.HubNative.pushAddress ? window.HubNative : null);

// what a tap leads to; the chat lists may still be loading when the tap started the app, so it waits for them a little
function open(logic, srv, d, tries = 0) {
  const again = () => { if (tries < OPEN_TRIES) setTimeout(() => open(logic, srv, d, tries + 1), OPEN_EVERY_MS); };
  if (d.kind === 'dm') {
    const row = (logic.state.dms || []).find((x) => x.id === d.conversation);
    if (row) logic.openDm(row.with); else again();
  } else if (d.kind === 'forum') {
    const local = srv.forumLocalId && srv.forumLocalId.get(d.conversation);
    if (local) logic.setState({ forum: local, panelKind: 'forum', mode: logic.isMobileViewport() ? 'page' : 'docked', tab: 'discussions' }); else again();
  } else if (d.kind === 'notification') {
    logic.setState({ notifOpen: true });
  }
  // a call: its card is already showing from the live connection
}

export async function startPush(logic, srv) {
  const n = native();
  if (!n) return;
  const register = async (addr) => {
    if (!addr || !addr.token) return;
    await srv.request('POST', '/push/devices', { token: addr.token, platform: addr.platform });
    n.pushReady(true);
  };
  window.__hubPushAddress = (addr) => { register(addr).catch(() => n.pushReady(false)); };
  window.__hubPushTap = () => { n.pushOpened().then((d) => { if (d) open(logic, srv, d); }).catch(() => {}); };
  try { await register(await n.pushAddress()); } catch (e) { n.pushReady(false); }
  window.__hubPushTap();
}

// signing out on the phone: this address stops receiving pushes (the backend also stops once the sign-in ends)
export async function stopPush(call) {
  const n = native();
  if (!n) return;
  try {
    const addr = await n.pushAddress();
    if (addr && addr.token) await call('DELETE', '/push/devices', { token: addr.token, platform: addr.platform });
  } catch (e) { /* the ended sign-in already stops the pushes */ }
  n.pushReady(false);
}
