// One shared Hub logic instance for the whole app.
// Components read values computed by HubLogic.renderVals() through hubBindings([...names]).
import { shallowReactive, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { HubLogic } from '~/logic/HubLogic.js';
import { sectionRoutes } from '~/logic/routes.js';
import { startServerSync } from '~/logic/server/boot.js';

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

  const start = sectionFor(route.path);
  if (logic.state.section !== start) logic.setState({ section: start });

  const known = (p) => Object.keys(sectionRoutes).some((k) => sectionRoutes[k] === p);
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

  onMounted(() => {
    const be = useBackend();
    const session = be.session();
    if (session && session.org) startServerSync(logic, be.request, session).catch((e) => console.error('Could not connect to the server:', e));
    if (logic.componentDidMount) logic.componentDidMount();
    document.addEventListener('click', openInAppSecureShare, true);
  });
  watch(vals, () => { if (logic.componentDidUpdate) logic.componentDidUpdate(); }, { flush: 'post' });
  onBeforeUnmount(() => {
    document.removeEventListener('click', openInAppSecureShare, true);
    if (logic.componentWillUnmount) logic.componentWillUnmount();
  });
}
