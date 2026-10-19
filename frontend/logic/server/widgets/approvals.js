// Approvals: the inbox (Approvals page) and the approval flows set on Roles & rights. The server runs each flow: when the
// last level approves it applies the change (publishes the memo, grants the award, opens the project ...).
import { when } from './tickets.js';
import * as comms from './comms.js';
import * as awards from './awards.js';
import * as projects from './projects.js';

export const name = 'approvals';

const KIND = { 'comm.publish': 'comm', 'award.grant': 'award', 'project.activate': 'project', 'rights.role': 'rights', 'person.update': 'personEdit', 'share.activate': 'share' };

const mapRequest = (r) => ({
  id: r.ref, uid: r.id, flowId: r.flow, title: r.title, by: r.by, when: when(r.created_at), level: r.level, status: r.status, kind: KIND[r.kind] || r.kind,
  history: (r.history || []).map((h) => ({ level: h.level, who: h.who, decision: h.decision, when: when(h.at), comment: h.comment || '' })),
  note: r.note || '', canDecide: !!r.can_decide, waitingFor: r.waiting_for || ''
});

export async function load(logic, srv) {
  const [{ flows }, { requests }] = await Promise.all([srv.request('GET', '/approvals/flows'), srv.request('GET', '/approvals?limit=200')]);
  srv.flowSnap = JSON.stringify(flows);
  logic.setState({ flows, requests: requests.map(mapRequest) });
  if (!srv.watchingFlows) { srv.watchingFlows = true; watchFlows(logic, srv); }
}

export async function decide(logic, srv, rq, approve, comment = '') {
  await srv.request('POST', `/approvals/${rq.uid}/${approve ? 'approve' : 'reject'}`, { comment });
  await load(logic, srv);
  // the approved change happened on the server: refresh what it touched
  if (rq.kind === 'comm') await comms.load(logic, srv);
  if (rq.kind === 'award') await awards.load(logic, srv);
  if (rq.kind === 'project') await projects.load(logic, srv);
  if (rq.kind === 'rights' || rq.kind === 'personEdit') { const r = await import('./roles.js'); await r.load(logic, srv); }
}

// Roles & rights > Approval flows edits state.flows in place; save each changed flow.
function watchFlows(logic, srv) {
  let timer = null;
  const flush = async () => {
    const now = JSON.stringify(logic.state.flows || []);
    if (now === srv.flowSnap) return;
    const before = JSON.parse(srv.flowSnap);
    srv.flowSnap = now;
    for (const f of logic.state.flows) {
      const b = before.find((x) => x.id === f.id);
      if (b && b.levels === f.levels) continue;   // only the number of levels is edited here; approvers follow from the "Approve level N" rights
      try { await srv.request('PUT', `/approvals/flows/${f.id}`, { levels: f.levels }); } catch (e) { logic.notify(e.message || 'Could not save the approval flow.', '#FF9B8A'); }
    }
    await load(logic, srv);   // the approver names per level come back from the server
  };
  const orig = logic.setState.bind(logic);
  logic.setState = (update, done) => { orig(update, done); clearTimeout(timer); timer = setTimeout(flush, 800); };
}
