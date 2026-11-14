// Counts a communication as read only when it was really on the person's screen. A memo, announcement, newsletter or job
// post shown on the wall or in a forum is marked read once at least 60% of it has been visible for 2 seconds while the tab
// is in front. Scrolling past, a background tab, or a post below the fold does not count. Elements opt in with the
// attributes data-seen-conv (conversation id) and data-seen-msg (message id); nothing else is tracked.
let sendFn = null;
const pending = new Map();   // "conv|msg" -> {conv, msg}
const done = new Set();      // already recorded in this session
const timers = new Map();    // element -> visibility timer
let io = null, mo = null, flushT = 0;

function flush() {
  flushT = 0;
  if (!sendFn || !pending.size) return;
  const byConv = {};
  pending.forEach(({ conv, msg }) => { (byConv[conv] = byConv[conv] || []).push(msg); });
  pending.clear();
  Object.entries(byConv).forEach(([conv, msgs]) => sendFn(conv, msgs).catch(() => { msgs.forEach((m) => done.delete(conv + '|' + m)); }));
}
function seen(el) {
  const conv = el.getAttribute('data-seen-conv'), msg = el.getAttribute('data-seen-msg');
  if (!conv || !msg) return;
  const k = conv + '|' + msg;
  if (done.has(k)) return;
  done.add(k); pending.set(k, { conv, msg });
  if (!flushT) flushT = setTimeout(flush, 3000);
}
function watch(el) { if (io && el.getAttribute('data-seen-conv')) io.observe(el); }

export function startSeenTracking(send) {
  sendFn = send;
  if (io || typeof window === 'undefined' || !window.IntersectionObserver) return;
  io = new IntersectionObserver((entries) => {
    entries.forEach((e) => {
      const el = e.target;
      if (e.isIntersecting && e.intersectionRatio >= 0.6) {
        if (!timers.has(el)) timers.set(el, setTimeout(() => { timers.delete(el); if (!document.hidden && document.hasFocus()) { seen(el); io.unobserve(el); } }, 2000));
      } else if (timers.has(el)) { clearTimeout(timers.get(el)); timers.delete(el); }
    });
  }, { threshold: [0, 0.6, 1] });
  document.querySelectorAll('[data-seen-conv]').forEach(watch);
  mo = new MutationObserver((list) => list.forEach((r) => r.addedNodes.forEach((n) => {
    if (n.nodeType !== 1) return;
    if (n.hasAttribute('data-seen-conv')) watch(n);
    n.querySelectorAll && n.querySelectorAll('[data-seen-conv]').forEach(watch);
  })));
  mo.observe(document.body, { childList: true, subtree: true });
  window.addEventListener('pagehide', flush);
}
