// Home layout: which widgets this person has, which are hidden, the theme (per person, /me/prefs) and the widget
// settings an admin sets for the whole organization (/settings/widgets). Changes are saved by watching the state.
export const name = 'prefs';

const PERSONAL = ['widgets', 'hiddenWidgets', 'theme', 'hiddenPosts', 'savedPosts', 'dmPrivacy', 'hiddenChatMessages', 'levelsAdded', 'notifTones'];
// the settings have a size limit, so only the newest "deleted for me" messages are kept
const trim = (k, v) => (k === 'hiddenChatMessages' && v && typeof v === 'object' ? Object.fromEntries(Object.entries(v).slice(-150)) : v);

export async function load(logic, srv) {
  // if the layout cannot be read, the home still opens (with the defaults) instead of waiting forever
  const prefs = await srv.request('GET', '/me/prefs').catch((e) => { logic.setState({ widgetsReady: true }); throw e; });
  const patch = {};
  PERSONAL.forEach((k) => { if (prefs[k] !== undefined) patch[k] = prefs[k]; });
  try { patch.wcfg = await srv.request('GET', '/settings/widgets'); } catch (e) { /* not allowed to see it */ }
  patch.widgetsReady = true;   // the home page may now decide which widgets to show
  logic.setState(patch);
  watch(logic, srv);
}

function watch(logic, srv) {
  const saved = { personal: JSON.stringify(PERSONAL.map((k) => logic.state[k])), wcfg: JSON.stringify(logic.state.wcfg || {}) };
  let timer = null;
  const flush = async () => {
    const s = logic.state;
    const personal = JSON.stringify(PERSONAL.map((k) => s[k]));
    if (personal !== saved.personal) {
      saved.personal = personal;
      const body = {}; PERSONAL.forEach((k) => { body[k] = trim(k, s[k]); });
      srv.request('PUT', '/me/prefs', body).catch((e) => console.warn('Could not save your home layout:', e.message));
    }
    const wcfg = JSON.stringify(s.wcfg || {});
    if (wcfg !== saved.wcfg) {
      saved.wcfg = wcfg;
      srv.request('PUT', '/settings/widgets', s.wcfg || {}).catch((e) => {
        // not saved (no right, or the server refused it): say so, and show what the organization really has
        logic.notify((e && e.message) || 'The widget settings could not be saved.', '#FF9B8A');
        reloadCfg();
      });
    }
  };
  // what the organization has set, as the server holds it (also after someone else changed it)
  async function reloadCfg() {
    try { const w = await srv.request('GET', '/settings/widgets'); saved.wcfg = JSON.stringify(w); logic.setState({ wcfg: w }); } catch (e) { /* keeps what it has */ }
  }
  // another administrator switched widgets on or off: it shows here now, without a reload
  if (srv.rt) srv.rt.on('widgets.changed', (ev) => { if (ev && ev.actor === srv.me.person) return; reloadCfg(); });
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); clearTimeout(timer); timer = setTimeout(flush, 700); };
}
