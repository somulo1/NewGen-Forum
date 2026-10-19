import { shareUrl } from '~/logic/shareOut.js';
// renderVals step: roles. Receives the names built by earlier steps and returns the ones later steps need.
export function buildRoles(ctx) {
  const { badge, flowById, ini, modInfo, pColor, rKey, rightOf, rightsDef, roleDefs, s, set, tabS, toggleRight } = ctx;
  // ---- Roles page ----
  const memberSamples = { superadmin: ['Samwel Okoth Omulo', 'Brian Kiprop'], admin: ['Samwel Okoth Omulo', 'David Otieno', 'Grace Njeri'], hod: ['Grace Njeri', 'Amina Wanjiru', 'Kevin Mutua'], it: ['Brian Kiprop', 'David Otieno', 'John Kamau'], finance: ['Amina Wanjiru', 'Mercy Wambui'], staff: ['Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'John Kamau', 'Lilian Chebet'], guest: ['External auditor'] };
  const roleSelDef = roleDefs.find((r) => r.id === s.roleSel) || roleDefs[5];
  const roleList = roleDefs.map((r) => ({ ...r, pick: () => this.setState({ roleSel: r.id }), style: 'display:flex;align-items:center;gap:8px;width:100%;min-height:52px;padding:6px 12px;border-radius:10px;cursor:pointer;color:var(--k-E8EFEA,#E8EFEA);border:1px solid ' + (s.roleSel === r.id ? '#3F524A' : 'transparent') + ';background:' + (s.roleSel === r.id ? 'var(--k-22302A,#22302A)' : 'transparent') }));
  const roleSel = { ...roleSelDef,
    members: (memberSamples[roleSelDef.id] || []).map((n) => ({ name: n, initials: n === 'Samwel Okoth Omulo' ? 'S' : ini(n), color: pColor[n] || '#C9B6FF' })),
    modules: Object.keys(rightsDef).map((m) => {
      const rights = rightsDef[m].map((r) => { const k = rKey(roleSelDef.id, m, r); return { name: r, on: rightOf(roleSelDef.id, m, r), locked: !!roleSelDef.locked, pending: s.pendingRights[k] !== undefined, toggle: () => toggleRight(roleSelDef.id, m, r) }; });
      return { label: modInfo[m][0], tint: modInfo[m][1], rights, onCount: rights.filter((x) => x.on).length, total: rights.length };
    }) };
  const roleTab = { rights: tabS(s.roleTab === 'rights'), members: tabS(s.roleTab === 'members'), flows: tabS(s.roleTab === 'flows'),
    rightsGo: set({ roleTab: 'rights' }), membersGo: set({ roleTab: 'members' }), flowsGo: set({ roleTab: 'flows' }),
    isRights: s.roleTab === 'rights', isMembers: s.roleTab === 'members', isFlows: s.roleTab === 'flows' };
  const flowItem = (f) => ({ ...f, modName: modInfo[f.mod] ? modInfo[f.mod][0] : f.mod, levelsVal: String(f.levels), hasLevels: f.levels > 0,
    setLevels: (e) => { const v = Number(e.target.value); this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, levels: v } : x)) })); },
    levels: f.approvers.slice(0, f.levels).map((who, i) => ({ n: i + 1, who, set: (e) => { const v = e.target.value; this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, approvers: x.approvers.map((a, j) => (j === i ? v : a)) } : x)) })); } })) });
  const flowsAll = s.flows.map(flowItem);
  const f8 = flowById('f8');

  // ---- Approvals inbox ----
  const reqItems = s.requests.map((rq) => {
    const f = flowById(rq.flowId) || { levels: 1, approvers: ['Line manager'], mod: 'home' };
    const L = Math.max(1, f.levels);
    const pending = rq.status === 'Pending';
    const stC = rq.status === 'Approved' ? '#B8F36B' : rq.status === 'Rejected' ? '#FF9B8A' : '#FFB86B';
    const steps = Array.from({ length: L }, (_, i) => {
      const lvl = i + 1;
      const h = rq.history.find((x) => x.level === lvl);
      const done = !!h && h.decision === 'Approved';
      const rej = !!h && h.decision === 'Rejected';
      const curL = pending && lvl === rq.level;
      const c = done ? '#B8F36B' : rej ? '#FF9B8A' : curL ? '#FFB86B' : '#9DB0A6';
      return { n: done ? '✓' : String(lvl), who: f.approvers[i] || 'Approver', color: c,
        state: h ? h.decision + ' by ' + h.who + ' · ' + h.when : curL ? 'Waiting now' : 'Not reached yet',
        bar: 'display:block;height:4px;border-radius:2px;background:' + (done ? '#B8F36B' : rej ? '#FF9B8A' : curL ? '#FFB86B' : 'var(--k-2A3832,#2A3832)'),
        dot: 'width:22px;height:22px;flex:none;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;box-sizing:border-box;' + (done ? 'background:#B8F36B;color:#0F1714' : 'border:2px solid ' + c + ';color:' + c) };
    });
    const decide = (ok) => () => {
      const cur = this.state.requests.find((x) => x.id === rq.id);
      const lvl = cur.level;
      const hist = cur.history.concat([{ level: lvl, who: 'You', decision: ok ? 'Approved' : 'Rejected', when: 'Just now' }]);
      if (!ok) {
        this.setState((st) => ({ requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, status: 'Rejected' } : x)), pendingRights: rq.kind === 'rights' ? Object.fromEntries(Object.entries(st.pendingRights).filter(([k]) => k !== rq.key)) : st.pendingRights }));
        this.notify(rq.id + ' rejected. ' + rq.by + ' has been told.', '#FF9B8A');
        return;
      }
      if (lvl < L) {
        this.setState((st) => ({ requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, level: lvl + 1 } : x)) }));
        this.notify(rq.id + ' approved at level ' + lvl + '. Sent to ' + (f.approvers[lvl] || 'the next approver') + '.', '#B8F36B');
      } else {
        this.setState((st) => {
          const out = { requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, status: 'Approved' } : x)) };
          if (rq.kind === 'project' && rq.project && st.widgets.includes('projects')) {
            out.projects = [rq.project].concat(st.projects.filter((p) => p.id !== rq.project.id));
            out.section = 'projects';
            out.projectOpen = rq.project.id;
            out.projectForm = null;
          }
          if (rq.kind === 'rights') { out.rightsOv = { ...st.rightsOv, [rq.key]: rq.val }; out.pendingRights = Object.fromEntries(Object.entries(st.pendingRights).filter(([k]) => k !== rq.key)); }
          if (rq.kind === 'award') {
            const a = st.awardList.find((x) => x.id === rq.awardId);
            out.awardList = st.awardList.map((x) => (x.id === rq.awardId ? { ...x, status: 'Approved' } : x));
            if (a && a.wall) {
              const names = a.people.join(', ');
              out.posts = [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: 'Performance award · Just now', text: 'Congratulations ' + a.people.map((n) => '@' + n).join(', ') + ' on the ' + a.type + ' award!' + (a.citation ? '\n' + a.citation : ''), likes: 0, extra: 0, shares: 0, comments: [],
                media: a.poster ? [{ id: 'pst' + Date.now(), name: 'Award poster', meta: '', isImg: true, url: a.poster }] : [],
                poster: a.poster ? null : { award: a.type, names: names, when: 'Today' }, tags: a.people.slice() }].concat(out.posts || st.posts);
            }
          }
          if (rq.kind === 'comm') {
            const cmItem = st.comms.find((c) => c.id === rq.commId);
            out.comms = st.comms.map((c) => (c.id === rq.commId ? { ...c, status: 'Published', read: 0 } : c));
            if (cmItem && cmItem.to.indexOf('Community wall') !== -1) out.posts = [{ id: Date.now(), author: cmItem.by === 'Samwel Okoth Omulo' ? 'Sam' : cmItem.by, initials: cmItem.by === 'Samwel Okoth Omulo' ? 'S' : ini(cmItem.by), color: '#F49AC1', meta: cmItem.type + ' · Just now', text: cmItem.title + '\n' + cmItem.body, likes: 0, extra: 0, shares: 0, comments: [] }].concat(st.posts);
          }
          if (rq.kind === 'share') {
            const sp = st.posts.find((q) => q.id === rq.postId);
            out.shareCount = { ...st.shareCount, [rq.postId]: (st.shareCount[rq.postId] || 0) + 1 };
            out.shareNote = { ...st.shareNote, [rq.postId]: 'Approved. Open ' + rq.app + ' from your notifications to post it.' };
            if (sp) out.notifs = [{ text: 'Approved: your post can go to ' + rq.app + '. Tap to open it with the full post.', color: '#B8F36B', time: 'Just now', ticket: null, url: shareUrl(rq.app, sp, []) }].concat(st.notifs);
          }
          return out;
        });
        if (rq.kind === 'award') {
          const a = this.state.awardList.find((x) => x.id === rq.awardId);
          if (a) (a.forums || []).forEach((fid) => this.addForumMsg(fid, 'Award: ' + a.type + ' · Congratulations ' + a.people.map((n) => '@' + n).join(', ') + '!' + (a.poster ? ' (poster attached)' : ''), a.poster ? { name: 'Award poster.png', meta: 'Image', url: a.poster, isImg: true } : undefined));
        }
        if (rq.kind === 'personEdit') this.applyPersonEdit(rq.changes);
        const projectReady = rq.kind === 'project' && this.state.projects.some((p) => rq.project && p.id === rq.project.id);
        this.notify(rq.id + ' fully approved: ' + rq.title + (rq.kind === 'award' ? '. Everyone awarded was tagged and notified.' : projectReady ? '. The project is now open in Projects.' : rq.kind === 'project' ? '. Projects is disabled, so it was not added.' : '.'), '#B8F36B');
      }
    };
    return { ...rq, modName: modInfo[f.mod] ? modInfo[f.mod][0] : '', steps, statusLabel: pending ? 'Level ' + rq.level + ' of ' + L : rq.status, statusStyle: badge(stC),
      levelsLabel: L === 1 ? 'Direct approval' : L + '-level approval', hasNote: !!rq.note,
      canAct: pending, currentApprover: f.approvers[rq.level - 1] || 'approver', approveLabel: rq.level < L ? 'Approve & send on' : 'Approve',
      approve: decide(true), reject: decide(false) };
  });
  const pendingCount = reqItems.filter((r) => r.status === 'Pending').length;
  const requestsShown = s.apTab === 'waiting' ? reqItems.filter((r) => r.status === 'Pending') : reqItems;
  return { f8, flowItem, flowsAll, pendingCount, requestsShown, roleList, roleSel, roleTab };
}
