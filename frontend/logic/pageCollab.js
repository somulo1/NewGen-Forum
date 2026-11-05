import * as Y from 'yjs';
import { newId } from './server/e2ee.js';

// Editing a page together. Changes merge by themselves (Yjs) and pass through the server, which stores nothing: the page is saved as usual.
const toB64 = (u) => { let s = ''; for (let i = 0; i < u.length; i += 8192) s += String.fromCharCode.apply(null, u.subarray(i, i + 8192)); return btoa(s); };
const fromB64 = (t) => Uint8Array.from(atob(t), (c) => c.charCodeAt(0));
const WAIT_FOR_PEER = 800;

export function joinPage(srv, project, page, onPeers) {
  const doc = new Y.Doc();
  const sid = newId();
  const peers = new Map();
  let gotState = false, pending = [], timer = null, closed = false;
  const send = (kind, data = '') => srv.request('POST', `/projects/${project}/pages/${page}/live`, { kind, data, sid }).catch(() => {});
  const tell = () => onPeers([...new Set(peers.values())]);
  const off = srv.rt.on('project.live', (ev) => {
    const d = ev.data || {};
    if (d.page !== page || d.sid === sid) return;
    if (d.kind === 'bye') { peers.delete(d.sid); tell(); return; }
    if (!peers.has(d.sid)) { peers.set(d.sid, d.name); tell(); }
    if (d.kind === 'hello') { if (seeded()) send('state', toB64(Y.encodeStateAsUpdate(doc))); return; }
    if (d.data) { gotState = true; Y.applyUpdate(doc, fromB64(d.data), 'remote'); }
  });
  const seeded = () => doc.getXmlFragment('default').length > 0;
  doc.on('update', (u, origin) => {
    if (origin === 'remote' || origin === 'seed' || closed) return;
    pending.push(u);
    if (timer) return;
    timer = setTimeout(() => { timer = null; const m = Y.mergeUpdates(pending); pending = []; send('update', toB64(m)); }, 150);
  });
  return {
    doc,
    // waits for anyone already editing. false: nobody is, so the caller starts from the saved page
    ready: async () => { send('hello'); await new Promise((r) => setTimeout(r, WAIT_FOR_PEER)); return gotState; },
    // The same saved page must always become the same document, so two people starting together do not double its text.
    seed: (fill) => { const id = doc.clientID; doc.clientID = 1; Y.transact(doc, fill, 'seed'); doc.clientID = id; },
    leave: () => { closed = true; clearTimeout(timer); off(); send('bye'); doc.destroy(); }
  };
}
