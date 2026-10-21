// renderVals step: people. Receives the names built by earlier steps and returns the ones later steps need.
export function buildPeople(ctx) {
  const { has, reportOf, s, set } = ctx;
  // ---- People directory ----
  const PAGE = 24;
  const q = s.pq.trim().toLowerCase();
  const statusCol = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
  const base = this.peopleDir().filter((p) => (!q || (p.name + ' ' + p.title + ' ' + p.dept).toLowerCase().indexOf(q) !== -1) && (!s.pAvail || p.status === 'Available'));
  const inLoc = base.filter((p) => s.ploc === 'All' || p.loc === s.ploc);
  const results = inLoc.filter((p) => s.pdept === 'All' || p.dept === s.pdept);
  const pages = Math.max(1, Math.ceil(results.length / PAGE));
  const pg = Math.min(s.ppage, pages - 1);
  const deptNames = Array.from(new Set(this.peopleDir().map((p) => p.dept))).sort();
  const facetStyle = (on) => 'display:flex;align-items:center;gap:8px;width:100%;min-height:36px;padding:0 10px;border-radius:8px;border:0;cursor:pointer;font-size:13px;' + (on ? 'background:var(--k-2C3D34,#2C3D34);color:var(--k-E8EFEA,#E8EFEA);font-weight:600' : 'background:transparent;color:var(--k-C9D6CF,#C9D6CF)');
  const chipS = (on) => 'height:32px;padding:0 12px;border-radius:16px;cursor:pointer;font-size:12px;' + (on ? 'border:1px solid var(--k-B8F36B,#B8F36B);background:var(--k-2C3D34,#2C3D34);color:var(--k-E8EFEA,#E8EFEA);font-weight:600' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-C9D6CF,#C9D6CF)');
  const segS = (on) => 'height:34px;padding:0 14px;border-radius:9px;border:0;cursor:pointer;font-size:13px;font-weight:600;' + (on ? 'background:var(--k-E8EFEA,#E8EFEA);color:var(--k-0F1714,#0F1714)' : 'background:transparent;color:var(--k-C9D6CF,#C9D6CF)');
  const pagerS = (dis) => 'height:38px;padding:0 16px;border-radius:19px;border:1px solid var(--k-2A3832,#2A3832);background:transparent;font-size:13px;cursor:' + (dis ? 'default;color:var(--k-4A5C54,#4A5C54)' : 'pointer;color:var(--k-E8EFEA,#E8EFEA)');
  const dir = {
    total: this.peopleDir().length, deptCount: deptNames.length, onlineCount: this.peopleDir().filter((p) => p.status === 'Available').length,
    q: s.pq, onQ: (e) => this.setState({ pq: e.target.value, ppage: 0 }),
    avail: s.pAvail, toggleAvail: () => this.setState((st) => ({ pAvail: !st.pAvail, ppage: 0 })),
    hasFilters: s.pdept !== 'All' || s.ploc !== 'All' || s.pAvail || !!q,
    clear: set({ pdept: 'All', ploc: 'All', pAvail: false, pq: '', ppage: 0 }),
    depts: [{ name: 'All departments', n: inLoc.length, key: 'All' }].concat(deptNames.map((d) => ({ name: d, key: d, n: inLoc.filter((p) => p.dept === d).length }))).map((d) => ({ ...d, style: facetStyle(s.pdept === d.key), pick: () => this.setState({ pdept: d.key, ppage: 0 }) })),
    locs: ['All'].concat(has.branches ? this.org.branches.map((b) => b.name) : []).map((l) => ({ name: l === 'All' ? 'All locations' : l, style: chipS(s.ploc === l), pick: () => this.setState({ ploc: l, ppage: 0 }) })),
    isGrid: s.pview === 'grid', isList: s.pview === 'list', gridBtn: segS(s.pview === 'grid'), listBtn: segS(s.pview === 'list'),
    setGrid: set({ pview: 'grid' }), setList: set({ pview: 'list' }),
    none: results.length === 0,
    showing: results.length ? 'Showing ' + (pg * PAGE + 1) + '–' + Math.min(results.length, (pg + 1) * PAGE) + ' of ' + results.length + ' people' : '',
    page: results.slice(pg * PAGE, (pg + 1) * PAGE).map((p) => ({ ...p, statusColor: statusCol[p.status], open: () => this.setState({ personOpen: p.id }) })),
    pageNo: pg + 1, pages, atStart: pg === 0, atEnd: pg >= pages - 1,
    prev: () => this.setState({ ppage: Math.max(0, pg - 1) }), next: () => this.setState({ ppage: Math.min(pages - 1, pg + 1) }),
    prevStyle: pagerS(pg === 0), nextStyle: pagerS(pg >= pages - 1)
  };
  const po = s.personOpen !== null && s.personOpen !== undefined ? this.peopleDir()[s.personOpen] : null;
  const skillsByDept = { 'IT Operations': ['Networking', 'Windows server', 'Help desk'], 'Software Development': ['Laravel', 'Vue', 'Testing'], 'Finance': ['Reporting', 'Excel', 'Budgeting'], 'Risk & Compliance': ['Audit', 'Policy', 'Controls'], 'Sales': ['Negotiation', 'CRM'], 'HR': ['Recruiting', 'Onboarding'], 'Customer Care': ['Customer service', 'Complaints handling'], 'Marketing': ['Content', 'Social media'], 'Operations': ['Process improvement'], 'Logistics': ['Fleet', 'Scheduling'], 'Legal': ['Contracts', 'Compliance'] };
  const psn = po ? { ...po, statusColor: statusCol[po.status], cover: po.color + '55', manager: reportOf(po.name).top ? 'Top of the organization' : reportOf(po.name).display, hubRole: /Lead|Supervisor|Manager/.test(po.title) ? 'Head of department' : 'Staff',
    securityCode: (() => { const v = this.srv && this.srv.vault; if (!v || !po.hasAccount) return ''; const a = v.acctOfPerson.get(po.uid); const c = a ? v.safetyCodeOf(a) : ''; if (!c) v.accountOfPerson(po.uid).then((x) => { if (x) this.forceUpdate(); }); return c; })(),
    email: po.email || '[' + po.name.toLowerCase().replace(' ', '.') + '@company.com]', skills: (skillsByDept[po.dept] || ['Teamwork']).map((x) => ({ name: x })),
    message: () => this.openDm(po.name),
    awards: has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(po.name) !== -1).map((a) => ({ type: a.type, when: new Date(a.date + 'T00:00:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }), citation: a.citation })) : [],
    hasAwards: has.awards && s.awardList.some((a) => a.status === 'Approved' && a.people.indexOf(po.name) !== -1),
    canMessage: has.chat && !(this.srv && po.uid === this.srv.me.person),
    voice: () => this.startDmCall('voice', po.name, po.initials, po.color), video: () => this.startDmCall('video', po.name, po.initials, po.color) } : {};
  return { chipS, dir, po, psn, segS, skillsByDept };
}
