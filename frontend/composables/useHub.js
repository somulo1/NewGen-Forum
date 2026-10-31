// One shared Hub logic instance for the whole app.
// Components read values computed by HubLogic.renderVals() through hubBindings([...names]).
import { shallowReactive, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { HubLogic } from '~/logic/HubLogic.js';
import { sectionRoutes } from '~/logic/routes.js';
import { startServerSync } from '~/logic/server/boot.js';
import { startSessionGuard } from '~/composables/useSessionGuard.js';

let hub = null;

export function useHubStore() {
  if (!hub) {
    const logic = new HubLogic({});
    logic.state = shallowReactive(logic.state || {});
    const vals = computed(() => {
      logic.__tick.value; // forceUpdate()
      return logic.renderVals() || {};
    });
    hub = { logic, vals };
    if (import.meta.dev && typeof window !== 'undefined') window.__hub = hub;   // for measuring in development
  }
  return hub;
}

// Options-API mixin: exposes each named value from renderVals() as a computed property.
export function hubBindings(names) {
  const map = {};
  names.forEach((n) => { map[n] = function () { return useHubStore().vals.value[n]; }; });
  return { computed: map };
}

// Used once by layouts/default.vue: lifecycle hooks + keeps the URL and the section in sync.
export function useHubShell() {
  const { logic, vals } = useHubStore();
  const route = useRoute();
  const router = useRouter();
  const sectionFor = (p) => Object.keys(sectionRoutes).find((k) => sectionRoutes[k] === p) || 'home';

  try { logic._accessPending = !!(useBackend().session() || {}).org; } catch (e) { logic._accessPending = false; }
  const start = sectionFor(route.path);
  if (logic.state.section !== start) logic.setState({ section: start });

  const known = (p) => Object.keys(sectionRoutes).some((k) => sectionRoutes[k] === p);
  // Every page change swaps the session's tokens for new ones, so a copied token goes stale within one click.
  watch(() => route.path, () => { try { const be = useBackend(); if (be.session()) be.rotate().catch(() => {}); } catch (e) { /* not signed in */ } });
  watch(() => route.path, (p) => {
    if (!known(p)) return;
    const s = sectionFor(p);
    if (logic.state.section !== s) logic.setState({ section: s, ticketOpen: null, projectOpen: null, folder: null, navOpen: false });
  });
  watch(() => logic.state.section, (s) => {
    const r = sectionRoutes[s] || '/';
    if (route.path !== r) router.push(r);
  });

  const openInAppSecureShare = (event) => {
    if (!(event.target instanceof Element)) return;
    const anchor = event.target.closest('a[href]');
    if (!anchor) return;
    const target = new URL(anchor.href, window.location.origin);
    if (target.origin !== window.location.origin || !/^\/s\/[^/]+\/?$/.test(target.pathname)) return;
    event.preventDefault();
    router.push(target.pathname + target.search + target.hash);
  };

  let guard = null;
  // The session lives in the browser's storage, shared by every tab. When another tab signs someone else in (or out), this tab must
  // not go on showing, or acting as, the person it started with.
  const onStorage = (e) => {
    if (e.key !== 'hub-session') return;
    const mine = logic.srv && logic.srv.session && logic.srv.session.account;
    let now = null;
    try { now = e.newValue ? JSON.parse(e.newValue).account : null; } catch (x) { /* unreadable: treat as signed out */ }
    if (mine && now !== mine) window.location.reload();
  };
  // Coming back to this tab after a while: ask the server who this is and what they may do now, instead of trusting what was loaded
  // earlier. A different person (or a signed-out session) reloads the page; a changed role or rights refresh in place.
  let hiddenAt = 0, checking = false;
  const onVisible = async () => {
    if (document.visibilityState === 'hidden') { hiddenAt = Date.now(); return; }
    if (!logic.srv || checking || !hiddenAt || Date.now() - hiddenAt < 30000) return;
    checking = true;
    try {
      const me = await logic.srv.request('GET', '/me');
      if (me.account !== logic.srv.session.account || me.person !== logic.srv.me.person) { window.location.reload(); return; }
      const now = JSON.stringify([me.role, me.rights]);
      if (now !== logic.srv.meSeen) {
        logic.srv.meSeen = now;
        await Promise.all([import('~/logic/server/widgets/profile.js').then((m) => m.load(logic, logic.srv)), import('~/logic/server/widgets/roles.js').then((m) => m.load(logic, logic.srv))]);
        logic.forceUpdate();
      }
    } catch (err) {
      if (err && (err.status === 401 || err.status === 403)) window.location.reload();
    } finally { checking = false; }
  };
  onMounted(() => {
    window.addEventListener('storage', onStorage);
    document.addEventListener('visibilitychange', onVisible);
    const be = useBackend();
    const session = be.session();
    if (session && session.org) startServerSync(logic, be.request, session, be.bytes).catch((e) => { logic._accessPending = false; logic.forceUpdate(); console.error('Could not connect to the server:', e); });
    if (session && session.org) {
      guard = startSessionGuard({ logic, be, goLogin: (why) => { window.location.assign('/login' + (why ? '?why=' + encodeURIComponent(why) : '')); } });
    }
    if (logic.componentDidMount) logic.componentDidMount();
    document.addEventListener('click', openInAppSecureShare, true);
  });
  watch(vals, () => { if (logic.componentDidUpdate) logic.componentDidUpdate(); }, { flush: 'post' });
  onBeforeUnmount(() => {
    if (guard) guard();
    window.removeEventListener('storage', onStorage);
    document.removeEventListener('visibilitychange', onVisible);
    document.removeEventListener('click', openInAppSecureShare, true);
    if (logic.componentWillUnmount) logic.componentWillUnmount();
  });
}
