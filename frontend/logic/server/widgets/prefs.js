// Home layout: which widgets this person has, which are hidden, the theme (per person, /me/prefs) and the widget
// settings an admin sets for the whole organization (/settings/widgets). Changes are saved by watching the state.
export const name = 'prefs';

const PERSONAL = ['widgets', 'hiddenWidgets', 'theme', 'hiddenPosts', 'savedPosts', 'dmPrivacy', 'hiddenChatMessages', 'levelsAdded'];
// the settings have a size limit, so only the newest "deleted for me" messages are kept
const trim = (k, v) => (k === 'hiddenChatMessages' && v && typeof v === 'object' ? Object.fromEntries(Object.entries(v).slice(-150)) : v);

export async function load(logic, srv) {
  const prefs = await srv.request('GET', '/me/prefs');
  const patch = {};
  PERSONAL.forEach((k) => { if (prefs[k] !== undefined) patch[k] = prefs[k]; });
  try { patch.wcfg = await srv.request('GET', '/settings/widgets'); } catch (e) { /* not allowed to see it */ }
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
      srv.request('PUT', '/settings/widgets', s.wcfg || {}).catch((e) => console.warn('Could not save the widget settings:', e.message));
    }
  };
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); clearTimeout(timer); timer = setTimeout(flush, 700); };
}
