// The realtime connection: one WebSocket per tab. The API hands out a 30-second single-use ticket, the socket attaches the
// organization and account channels on its own, and conversation channels are joined with { op: 'sub', conv }.
// Events arrive as { type, channel, org, actor, data }. Reconnects by itself and re-joins what was subscribed.
const base = () => {
  try {
    const cfg = window.__NUXT__ && window.__NUXT__.config && window.__NUXT__.config.public;
    if (cfg && cfg.apiBase) return String(cfg.apiBase).replace(/\/$/, '');
  } catch (e) { /* fall through */ }
  return window.location.protocol + '//' + window.location.hostname + ':8080';
};

import { retryDelay, spread } from './backoff.js';

export function connectRealtime(srv) {
  const rt = { ws: null, convs: new Set(), on: null, sub: null, unsub: null, close: null, listeners: new Map(), retry: 0, opened: 0, timer: 0, ping: 0, closed: false };

  const emit = (type, ev) => { (rt.listeners.get(type) || []).forEach((fn) => { try { fn(ev); } catch (e) { console.warn('realtime handler failed', e); } }); };
  const send = (o) => { if (rt.ws && rt.ws.readyState === 1) rt.ws.send(JSON.stringify(o)); };

  async function open() {
    if (rt.closed) return;
    let ticket;
    try { ticket = (await srv.request('POST', '/realtime/ticket', {})).ticket; } catch (e) { return again(); }
    const url = base().replace(/^http/, 'ws') + '/v1/realtime?ticket=' + encodeURIComponent(ticket);
    const ws = new WebSocket(url);
    rt.ws = ws;
    ws.onopen = () => {
      rt.retry = 0;
      const reconnect = rt.opened++ > 0;   // false for the first connection of this tab; true after a drop, when what was missed must be caught up
      rt.convs.forEach((c) => send({ op: 'sub', conv: c }));
      clearInterval(rt.ping);
      rt.ping = setInterval(() => send({ op: 'ping' }), 25000);
      emit('open', { reconnect });
    };
    ws.onmessage = (e) => { try { const ev = JSON.parse(e.data); if (ev && ev.type) emit(ev.type, ev); } catch (err) { /* not an event */ } };
    ws.onclose = () => { clearInterval(rt.ping); if (rt.ws === ws) rt.ws = null; again(); };
    ws.onerror = () => { try { ws.close(); } catch (e) { /* already closed */ } };
  }
  // Waits a random time (see backoff.js) so a crowd that lost its connection together does not come back together. A hidden tab
  // does not keep trying: it waits until it is shown again, then a short random pause, so thousands of background tabs cost nothing.
  function again() {
    if (rt.closed) return;
    clearTimeout(rt.timer);
    rt.retry += 1;
    if (typeof document !== 'undefined' && document.hidden) { rt.waitingVisible = true; return; }
    rt.timer = setTimeout(open, retryDelay(rt.retry));
  }
  function resume() {
    if (rt.closed || !rt.waitingVisible) return;
    rt.waitingVisible = false;
    clearTimeout(rt.timer);
    rt.timer = setTimeout(open, spread(2500));
  }
  const onVisible = () => { if (!document.hidden) resume(); };
  const onOnline = () => { if (!rt.ws && !rt.closed) { clearTimeout(rt.timer); rt.waitingVisible = false; rt.timer = setTimeout(open, spread(3000)); } };
  if (typeof document !== 'undefined') { document.addEventListener('visibilitychange', onVisible); window.addEventListener('online', onOnline); }

  rt.on = (type, fn) => {
    const l = rt.listeners.get(type) || new Set();
    l.add(fn); rt.listeners.set(type, l);
    return () => l.delete(fn);
  };
  rt.sub = (conv) => { if (rt.convs.has(conv)) return; rt.convs.add(conv); send({ op: 'sub', conv }); };
  rt.unsub = (conv) => { rt.convs.delete(conv); send({ op: 'unsub', conv }); };
  rt.typing = (conv, on = true) => send(on ? { op: 'typing', conv } : { op: 'typing', conv, on: false });
  rt.close = () => { rt.closed = true; if (typeof document !== 'undefined') { document.removeEventListener('visibilitychange', onVisible); window.removeEventListener('online', onOnline); } clearTimeout(rt.timer); clearInterval(rt.ping); if (rt.ws) rt.ws.close(); };
  open();
  return rt;
}
