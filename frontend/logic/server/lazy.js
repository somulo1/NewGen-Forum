// Lazy loading of the widgets. At sign-in only what every page needs is fetched (the directory, rights, organization, preferences,
// and the chats, notifications and settings that the sidebar and header show). Everything else loads the first time a page needs it:
// opening Projects fetches projects (and what its tables show next to them), not the security logs or the files. Each widget loads
// once, and again only when a live event says it changed. A page shows skeleton blocks until the widgets it needs have answered.
import { spread } from './backoff.js';

// the widgets each page reads from (a page may show another widget's records next to its own, so these are listed together)
export const NEEDS = {
  home: ['comms', 'calendar', 'tickets', 'projects', 'folders', 'awards', 'notes'],
  feed: ['wall'],
  meetings: ['calendar', 'projects'],
  calendar: ['calendar', 'projects'],
  comms: ['comms', 'approvals'],
  tickets: ['tickets'],
  ticketdesk: ['tickets'],
  projects: ['projects', 'tickets', 'folders', 'calendar'],
  approvals: ['approvals', 'awards', 'comms'],
  security: ['security'],
  roles: ['approvals', 'awards'],
  files: ['folders'],
  share: ['folders'],
  people: ['awards'],
  profile: ['awards', 'wall', 'projects', 'tickets'],
  structure: ['tickets', 'comms', 'projects', 'approvals'],
  orgunit: ['tickets', 'comms', 'projects', 'approvals'],
  orgedit: ['tickets', 'comms', 'projects', 'approvals'],
  widgets: ['approvals'],
  notes: ['notes'],
  settings: []
};

// what the open page and any open card or dialog need right now
export function needsOf(state) {
  const need = new Set(NEEDS[state.section] || []);
  if (state.personOpen) need.add('awards');       // a person's card shows their awards
  if (state.cmOpen) need.add('comms');            // writing a communication
  if (state.ticketOpen) need.add('tickets');
  return [...need];
}

export function createLazy(logic, modules, { concurrency = 3, retryAfterMs = 15000, now = Date.now } = {}) {
  const byName = new Map(modules.map((m) => [m.name, m]));
  const inflight = new Map();   // name -> promise
  const failed = new Map();     // name -> when it failed; a failed widget counts as answered (the page shows what it has) and is tried again later
  const waiting = [];
  let running = 0;

  const done = (name) => { logic.loaded.add(name); logic.forceUpdate(); };

  const pump = () => {
    while (running < concurrency && waiting.length) {
      const name = waiting.shift();
      running++;
      const mod = byName.get(name);
      const t0 = typeof performance !== 'undefined' ? performance.now() : 0;
      const p = Promise.resolve().then(() => mod.load(logic, logic.srv)).then(
        () => { failed.delete(name); },
        (e) => { failed.set(name, now()); if (typeof console !== 'undefined') console.warn('Could not load ' + name + ':', e); }
      ).then(() => {
        running--; inflight.delete(name); done(name);
        if (typeof console !== 'undefined' && console.debug) console.debug('[hub] ' + name + ' loaded in ' + Math.round((typeof performance !== 'undefined' ? performance.now() : 0) - t0) + ' ms (lazy)');
        pump();
      });
      inflight.set(name, p);
    }
  };

  const api = {
    // Starts loading what is not loaded yet. Safe to call on every render: a widget already loaded or loading is skipped.
    want(names) {
      for (const name of names) {
        if (!byName.has(name) || inflight.has(name) || waiting.includes(name)) continue;
        const f = failed.get(name);
        if (logic.loaded.has(name) && !(f && now() - f > retryAfterMs)) continue;
        waiting.push(name);
      }
      pump();
    },
    ready(names) { return names.every((n) => !byName.has(n) || logic.loaded.has(n)); },
    // widgets nobody has asked for yet, loaded quietly a little after sign-in so they are usually in by the time they are needed
    later(names, afterMs = 1200) {
      setTimeout(() => api.want(names), afterMs + spread(1500));
    },
    // After a live event: refresh a widget only if it was ever loaded (an unvisited page costs nothing)
    refresh(mod) {
      if (!logic.loaded.has(mod.name)) return Promise.resolve();
      return Promise.resolve(mod.load(logic, logic.srv));
    },
    isLoaded: (n) => logic.loaded.has(n),
    inflight: () => inflight.size
  };
  return api;
}
