// renderVals step: awards. Receives the names built by earlier steps and returns the ones later steps need.
export function buildAwards(ctx) {
  const { can, chip, chipS, flowById, fmtDay, forums, has, makeRequest, openPersonName, personOf, roleDefs, s, set, todayISO, visibleRoles } = ctx;
  // =============== Performance awards ===============
  const awardsOn = s.widgets.indexOf('awards') !== -1;
  const approvedAwards = s.awardList.filter((a) => a.status === 'Approved');
  const stripBase = [];
  approvedAwards.forEach((a) => a.people.forEach((n) => {
    const isMe = n === 'Samwel Okoth Omulo';
    const pp = personOf(n);
    stripBase.push({ name: n, initials: isMe ? 'S' : pp.initials, color: pp.color, award: a.type, when: fmtDay(a.date), photo: isMe ? s.profile.photo : '', hasPhoto: isMe && !!s.profile.photo, noPhoto: !(isMe && s.profile.photo), open: () => openPersonName(n) });
  }));
  const awardStrip = {
    hidden: s.hiddenWidgets.indexOf('awards') !== -1, hideLabel: s.hiddenWidgets.indexOf('awards') !== -1 ? 'Show' : 'Hide',
    toggleHide: () => this.setState((st) => ({ hiddenWidgets: st.hiddenWidgets.indexOf('awards') !== -1 ? st.hiddenWidgets.filter((x) => x !== 'awards') : st.hiddenWidgets.concat(['awards']) })),
    title: 'Performance awards', sub: stripBase.length + ' people recognised recently · tap anyone to see their profile',
    items: stripBase.length ? stripBase.concat(stripBase) : [],
    empty: stripBase.length === 0,
    trackStyle: 'display:flex;gap:14px;width:max-content;padding:0 7px;animation:hubMarquee ' + Math.max(30, stripBase.length * 6) + 's linear infinite',
    canGive: can('awards', 'Nominate') || can('awards', 'Approve awards'),
    openSettings: () => this.setState({ settingsFor: 'awards' }),
    remove: () => this.setState((x) => ({ widgets: x.widgets.filter((w) => w !== 'awards') }))
  };
  const awS = s.awd;
  const setAw = (patch) => this.setState((x) => ({ awd: Object.assign({}, x.awd, typeof patch === 'function' ? patch(x.awd) : patch) }));
  const personRole = (n) => s.memberRoles[n] || ({ 'Samwel Okoth Omulo': 'admin', 'Brian Kiprop': 'it', 'Amina Wanjiru': 'finance', 'Grace Njeri': 'hod' }[n]) || (/Lead|Supervisor/.test(((this.dir.find((d) => d.name === n) || {}).title) || '') ? (has.branches ? 'branch' : 'hod') : 'staff');
  const awq = awS.q.trim().toLowerCase();
  const awPool = this.dir.filter((d) => (awS.role === 'all' || personRole(d.name) === awS.role) && (!awq || (d.name + ' ' + d.title + ' ' + d.dept).toLowerCase().indexOf(awq) !== -1));
  const awMatches = awPool.slice(0, 40);
  const f7 = flowById('f7');
  const awTypes = ['Employee of the month', 'Star performer', 'Customer hero', 'Team player', 'Innovation award', 'Long service'];
  const aw = {
    type: awS.type, q: awS.q, role: awS.role, citation: awS.citation, poster: awS.poster, hasPoster: !!awS.poster, noPoster: !awS.poster, wall: awS.wall,
    types: awTypes.map((t) => ({ label: t, pick: () => setAw({ type: t }), style: chip(awS.type === t) })),
    roleOpts: visibleRoles.filter((r) => r.id !== 'superadmin').map((r) => ({ v: r.id, l: r.name })),
    setQ: (e) => setAw({ q: e.target.value }), setRole: (e) => setAw({ role: e.target.value }), setCitation: (e) => setAw({ citation: e.target.value }),
    setPoster: (e) => { const f = (e.target.files || [])[0]; e.target.value = ''; if (f) setAw({ poster: URL.createObjectURL(f) }); },
    clearPoster: () => setAw({ poster: '' }),
    toggleWall: () => setAw((x) => ({ wall: !x.wall })),
    forumChips: forums.map((f) => { const on = awS.forums.indexOf(f.id) !== -1; return { name: f.name, on: on, style: chipS(on), toggle: () => setAw((x) => ({ forums: on ? x.forums.filter((y) => y !== f.id) : x.forums.concat([f.id]) })) }; }),
    countLabel: awS.people.length ? awS.people.length + (awS.people.length === 1 ? ' person selected' : ' people selected') : 'Nobody selected yet',
    chosen: awS.people.map((n) => ({ ...personOf(n), remove: () => setAw((x) => ({ people: x.people.filter((y) => y !== n) })) })),
    matchLabel: awPool.length > 40 ? 'Showing 40 of ' + awPool.length + ' matches. Search to narrow it down.' : awPool.length + (awPool.length === 1 ? ' match' : ' matches'),
    selectAll: () => setAw((x) => ({ people: x.people.concat(awMatches.map((d) => d.name).filter((n) => x.people.indexOf(n) === -1)) })),
    matches: awMatches.map((d) => { const on = awS.people.indexOf(d.name) !== -1; return { name: d.name, initials: d.initials, color: d.color, sub: d.title + ' · ' + (roleDefs.find((r) => r.id === personRole(d.name)) || {}).name, on: on, mark: on ? 'Selected' : '+ Add',
      toggle: () => setAw((x) => ({ people: on ? x.people.filter((y) => y !== d.name) : x.people.concat([d.name]) })),
      style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;cursor:pointer;color:var(--k-E8EFEA,#E8EFEA);font-family:inherit;background:' + (on ? 'var(--k-24220F,#24220F)' : 'transparent') }; }),
    posterNames: awS.people.length ? (awS.people.length > 3 ? awS.people.slice(0, 3).join(', ') + ' +' + (awS.people.length - 3) : awS.people.join(', ')) : 'Awardee names',
    approvalNote: f7 && f7.levels > 0 ? 'Awards are approved before they appear (' + (f7.levels === 1 ? 'direct approval by ' + f7.approvers[0] : f7.levels + ' levels: ' + f7.approvers.slice(0, f7.levels).join(' → ')) + ').' : 'Awards appear straight away.',
    submitLabel: f7 && f7.levels > 0 ? 'Send for approval' : 'Give award',
    submitStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:700;cursor:pointer;' + (awS.people.length ? 'background:#FFD66B;color:#0F1714' : 'background:var(--k-2A3832,#2A3832);color:var(--k-9DB0A6,#9DB0A6)'),
    close: set({ awOpen: false }),
    submit: () => {
      const a0 = this.state.awd;
      if (!a0.people.length) { this.notify('Pick at least one person to award.', '#FFB86B'); return; }
      const id = 'AW-' + (13 + this.state.awardList.length);
      const needs = f7 && f7.levels > 0;
      const rec = { id: id, type: a0.type, people: a0.people.slice(), citation: a0.citation.trim(), date: todayISO, status: needs ? 'Pending' : 'Approved', poster: a0.poster, wall: a0.wall, forums: a0.forums.slice() };
      this.setState((x) => ({ awOpen: false, awardList: [rec].concat(x.awardList), awd: { type: 'Employee of the month', people: [], role: 'all', q: '', citation: '', poster: '', wall: true, forums: [] } }));
      const label = a0.type + ' for ' + (a0.people.length > 2 ? a0.people.length + ' people' : a0.people.join(' and '));
      if (needs) makeRequest('f7', label, { kind: 'award', awardId: id });
      else this.notify(label + ' is live on Home.', '#B8F36B');
    }
  };
  return { aw, awardStrip, awardsOn };
}
