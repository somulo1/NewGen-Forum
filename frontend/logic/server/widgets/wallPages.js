// The community wall a batch at a time, like pages of a book. The first batch opens complete, with its pictures and videos
// already decrypted. While the person reads it, the next batch is fetched and its pictures prepared, but not shown. When they reach
// the end, that batch appears at once, and the one after it starts getting ready. Nothing on screen moves while they read: new
// posts from others only ever join at the top.
import * as files from '../files.js';

export const BATCH = 25;               // posts per batch: the first screen opens fast, and the next batch is always ready in time
const FIRST_WAIT_MS = 8000;            // a slow picture does not hold the first batch back longer than this
const NEXT_WAIT_MS = 15000;
const MAX_PAGES_PER_STEP = 8;          // server pages (100 messages each, comments and likes included) read to fill one batch

const wait = (ms) => new Promise((r) => setTimeout(r, ms));

// the posts on screen: everything from the newest down to the boundary (the oldest post revealed)
export function shownCount(state) {
  const posts = state.posts || [];
  if (!state.wallUntil) return Math.min(BATCH, posts.length);
  return posts.filter((p) => p.id >= state.wallUntil).length;
}

export const visible = (state) => {
  const posts = state.posts || [];
  const until = state.wallUntil || (posts[Math.min(BATCH, posts.length) - 1] || {}).id;
  return until ? posts.filter((p) => p.id >= until) : [];
};

function setUrl(logic, fileId, url) {
  logic.setState((st) => ({
    posts: st.posts.map((p) => (p.media && p.media.some((x) => x.id === fileId && x.url !== url) ? { ...p, media: p.media.map((x) => (x.id === fileId ? { ...x, url } : x)) } : p))
  }));
}

// decrypts the pictures and videos of the newest `upTo` posts that are not open yet; resolves when they are (or after ms)
export function prepareMedia(logic, srv, upTo, ms) {
  const jobs = [];
  for (const p of (logic.state.posts || []).slice(0, upTo)) {
    for (const m of p.media || []) if (!m.url && m.ref) jobs.push(files.openUrl(srv, m.ref).then((url) => setUrl(logic, m.ref.fileId, url)).catch(() => {}));
  }
  return Promise.race([Promise.all(jobs), wait(ms)]);
}

// reads older server pages until at least n posts are held (or the wall has no more)
async function hold(logic, more, n) {
  for (let i = 0; i < MAX_PAGES_PER_STEP && (logic.state.posts || []).length < n; i++) {
    if (!(await more())) break;
  }
}

const ready = new WeakMap();   // srv -> the next batch being prepared

// gets the batch after what is shown ready: held, and its pictures open
export function getReady(logic, srv, more) {
  if (ready.has(srv)) return ready.get(srv);
  const job = (async () => {
    const shown = shownCount(logic.state);
    await hold(logic, more, shown + BATCH);
    await prepareMedia(logic, srv, shown + BATCH, NEXT_WAIT_MS);
  })().catch(() => {}).finally(() => ready.delete(srv));
  ready.set(srv, job);
  return job;
}

// The first batch, complete, the first time the wall opens. Later loads keep what was already revealed.
export async function first(logic, srv, more) {
  if (!logic.state.wallUntil) {
    await hold(logic, more, BATCH);
    await prepareMedia(logic, srv, BATCH, FIRST_WAIT_MS);
    const posts = logic.state.posts || [];
    if (posts.length) logic.setState({ wallUntil: posts[Math.min(BATCH, posts.length) - 1].id });
  } else {
    prepareMedia(logic, srv, shownCount(logic.state), NEXT_WAIT_MS);   // new posts at the top
  }
  getReady(logic, srv, more);
}

// The person reached the end of what is shown: the next batch appears (it is normally ready already).
export async function next(logic, srv, more, hasMore) {
  if (logic.state.wallLoadingMore) return;
  const shown = shownCount(logic.state);
  if (shown >= (logic.state.posts || []).length && !hasMore()) return;
  logic.setState({ wallLoadingMore: true });
  try {
    await getReady(logic, srv, more);
    const posts = logic.state.posts || [];
    const end = Math.min(shown + BATCH, posts.length);
    if (end > shown) logic.setState({ wallUntil: posts[end - 1].id });
  } finally {
    logic.setState({ wallLoadingMore: false });
  }
  getReady(logic, srv, more);
}
