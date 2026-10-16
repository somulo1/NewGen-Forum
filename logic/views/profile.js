// renderVals step: profile. Receives the names built by earlier steps and returns the ones later steps need.
export function buildProfile(ctx) {
  const { reportOf, fmtDay, forums, has, nav, projects, s, skillsByDept, ts } = ctx;
  // ---- My profile ----
  nav.profileCard = 'margin-top:auto;display:flex;align-items:center;gap:10px;width:100%;min-height:58px;padding:8px 10px;border-radius:12px;cursor:pointer;color:#E8EFEA;font-family:inherit;border:1px solid ' + (s.section === 'profile' && !s.profileOf ? '#3F524A' : '#24322C') + ';background:' + (s.section === 'profile' && !s.profileOf ? '#22302A' : '#16201C');
  const pf = s.profile;
  const statusDefs = [
    { label: 'Available', color: '#B8F36B', note: 'Available for chats and calls' },
    { label: 'In a meeting', color: '#FFB86B', note: 'In a meeting, replies may be slow' },
    { label: 'Focusing', color: '#7CC4FF', note: 'Focusing, notifications paused' },
    { label: 'Away', color: '#9DB0A6', note: 'Away from desk' }
  ];
  const curStatus = statusDefs.find((x) => x.label === pf.status) || statusDefs[0];
  const setPf = (patch) => this.setState((st) => ({ profile: Object.assign({}, st.profile, patch) }));
  let localTime = '';
  try { localTime = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi' }); } catch (err) { localTime = ''; }
  const checks = [!!pf.photo, (pf.bio || '').trim().length > 20, pf.skills.length >= 3, true, !!pf.phone.trim()];
  const complete = Math.round((checks.filter(Boolean).length / checks.length) * 100);
  const tip = !pf.photo ? 'Add a profile photo so colleagues recognise you.' : !pf.phone.trim() ? 'Add your phone number to reach 100%.' : (pf.bio || '').trim().length <= 20 ? 'Write a short bio.' : 'Your profile is complete.';
  const myPosts = s.posts.filter((p) => p.author === 'Sam');
  const myRaised = s.tickets.filter((t) => t.raisedBy === 'Sam');
  const myActivity = myPosts.slice(0, 2).map((p) => ({ tag: 'W', color: '#F49AC1', text: 'Posted on the wall: "' + ((p.text || 'a file').length > 70 ? p.text.slice(0, 70) + '…' : (p.text || 'a file')) + '"', time: p.meta.split('· ').pop() }))
    .concat(myRaised.slice(0, 2).map((t) => ({ tag: 'T', color: '#FF9B8A', text: 'Raised ' + t.id + ': ' + t.title, time: t.created })))
    .concat([
      { tag: 'F', color: '#6EE7C8', text: 'Shared "Access review template" in Risk & Compliance', time: 'Yesterday' },
      { tag: 'A', color: '#7CC4FF', text: 'Assigned IT-208 to Brian Kiprop', time: 'Today, 08:40' },
      { tag: 'B', color: '#FFD66B', text: 'Earned the Security Champion badge', time: 'Last month' },
      { tag: 'A', color: '#7CC4FF', text: 'Approved access request for the Finance reporting folder', time: '2 days ago' },
      { tag: 'F', color: '#6EE7C8', text: 'Uploaded "Q3 control test results.xlsx" to Files', time: '3 days ago' },
      { tag: 'T', color: '#FF9B8A', text: 'Closed ticket IT-187: VPN certificate renewal', time: 'Last week' },
      { tag: 'P', color: '#C9B6FF', text: 'Moved "Control testing" to In progress in Audit readiness', time: 'Last week' }
    ]);
  const kindOf = { W: 'Wall post', T: 'Ticket', F: 'File', A: 'Assignment', B: 'Badge', P: 'Project' };
  const withKind = (arr) => arr.map((a) => ({ ...a, kind: kindOf[a.tag] || 'Activity' }));
  const meOwn = {
    isOther: false, initials: 'S', color: '#C2185B', aboutTitle: 'About me', forumsTitle: 'My forums', reportsTo: reportOf('Samwel Okoth Omulo').name || 'Top of the organization',
    name: 'Samwel Okoth Omulo', role: 'Cybersecurity & Technology Risk Officer', dept: 'Risk & Compliance',
    status: curStatus.label, statusColor: curStatus.color, statusNote: curStatus.note, statusMenu: s.statusMenu,
    statuses: statusDefs.map((x) => ({ ...x, pick: () => this.setState((st) => ({ statusMenu: false, profile: { ...st.profile, status: x.label } })) })),
    photo: pf.photo, hasPhoto: !!pf.photo, noPhoto: !pf.photo,
    editing: s.profileEdit, notEditing: !s.profileEdit,
    coverStyle: 'position:relative;height:190px;background:' + pf.cover + ';overflow:hidden;border-radius:21px 21px 0 0',
    covers: [['#1F3A2B', 'Forest'], ['#1E3347', 'Ocean'], ['#3A2433', 'Plum'], ['#3D3320', 'Sand']].map(([c, label]) => ({ label: label + ' cover', pick: () => setPf({ cover: c }),
      style: 'width:26px;height:26px;border-radius:50%;cursor:pointer;background:' + c + ';border:2px solid ' + (pf.cover === c ? '#E8EFEA' : 'rgba(255,255,255,0.25)') })),
    complete, completeTip: tip, completeBar: 'height:100%;width:' + complete + '%;background:' + (complete === 100 ? '#B8F36B' : '#FFB86B') + ';border-radius:3px',
    bio: pf.bio, phone: pf.phone, newSkill: s.newSkill,
    skills: pf.skills.map((n, i) => ({ name: n, editing: s.profileEdit, remove: () => this.setState((st) => ({ profile: { ...st.profile, skills: st.profile.skills.filter((x, j) => j !== i) } })) })),
    facts: [
      { label: 'Department', value: 'Risk & Compliance', sub: 'Cybersecurity team' },
      { label: 'Reports to', value: reportOf('Samwel Okoth Omulo').top ? 'No one' : reportOf('Samwel Okoth Omulo').name, sub: reportOf('Samwel Okoth Omulo').sub },
      { label: 'Location', value: 'Kenya', sub: 'GMT+3 · ' + localTime + ' now' },
      { label: 'Member since', value: '[JOIN DATE]', sub: 'Hub member' }
    ],
    stats: [
      { label: 'Tickets raised', n: myRaised.length, color: '#FF9B8A' },
      { label: 'Tickets handled', n: s.tickets.filter((t) => t.handler === 'Sam' || t.assignees.indexOf('Sam') !== -1).length, color: '#7CC4FF' },
      { label: 'Wall posts', n: myPosts.length, color: '#F49AC1' },
      { label: 'Kudos received', n: 14, color: '#FFD66B' }
    ],
    projects: projects.filter((p) => p.id === 'audit' || p.id === 'network').map((p) => ({ ...p, myRole: p.id === 'audit' ? 'Controls reviewer' : 'Security reviewer' })),
    activity: withKind(myActivity),
    contact: [
      { label: 'Work email', value: '[you@company.com]', color: '#E8EFEA' },
      { label: 'Phone', value: pf.phone.trim() || 'Not added yet', color: pf.phone.trim() ? '#E8EFEA' : '#FFB86B' },
      { label: 'Office', value: '[OFFICE / FLOOR]', color: '#E8EFEA' }
    ],
    localTime, hours: ts.start + ' to ' + ts.end,
    badges: has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf('Samwel Okoth Omulo') !== -1).map((a) => ({ name: a.type, when: fmtDay(a.date), color: '#FFD66B', bg: '#2A2716' })) : [],
    hasBadges: has.awards && s.awardList.some((a) => a.status === 'Approved' && a.people.indexOf('Samwel Okoth Omulo') !== -1),
    forums: forums.filter((f) => f.id === 'risk' || f.id === 'it' || f.id === 'dev').map((f) => ({ ...f, myRole: f.id === 'risk' ? 'Admin' : 'Member' }))
  };
  // Full profile of another person (only for roles with the "View full profiles" right)
  const opDir = s.profileOf ? this.dir.find((x) => x.name === s.profileOf) : null;
  let meOther = null;
  if (opDir) {
    const first = opDir.name.split(' ')[0];
    const mine = (v) => v === opDir.name || v === first;
    const opStatus = statusDefs.find((x) => x.label === opDir.status) || statusDefs[0];
    const opRaised = s.tickets.filter((t) => mine(t.raisedBy));
    const opHandled = s.tickets.filter((t) => mine(t.handler) || (t.assignees || []).some(mine));
    const opPosts = s.posts.filter((p) => mine(p.author));
    const opAwards = has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(opDir.name) !== -1) : [];
    const opProjIds = s.projects.filter((p) => (p.tasks || []).some((t) => mine(t.who))).map((p) => p.id);
    const opActivity = opPosts.slice(0, 2).map((p) => ({ tag: 'W', color: '#F49AC1', text: 'Posted on the wall: "' + ((p.text || 'a file').length > 70 ? p.text.slice(0, 70) + '…' : (p.text || 'a file')) + '"', time: p.meta.split('· ').pop() }))
      .concat(opRaised.slice(0, 2).map((t) => ({ tag: 'T', color: '#FF9B8A', text: 'Raised ' + t.id + ': ' + t.title, time: t.created })))
      .concat(opHandled.slice(0, 2).map((t) => ({ tag: 'A', color: '#7CC4FF', text: 'Handling ' + t.id + ': ' + t.title, time: t.created })))
      .concat(opAwards.map((a) => ({ tag: 'B', color: '#FFD66B', text: 'Earned the ' + a.type + ' award', time: fmtDay(a.date) })))
      .concat([
        { tag: 'F', color: '#6EE7C8', text: 'Shared a document in the ' + opDir.dept + ' forum', time: 'Yesterday' },
        { tag: 'P', color: '#C9B6FF', text: 'Updated a task in ' + opDir.dept + ' planning', time: 'Last week' },
        { tag: 'A', color: '#7CC4FF', text: 'Joined the weekly ' + opDir.dept + ' standup', time: '2 weeks ago' },
        { tag: 'F', color: '#6EE7C8', text: 'Downloaded the "Staff handbook" from Files', time: '2 weeks ago' },
        { tag: 'W', color: '#F49AC1', text: 'Reacted to a post on the community wall', time: 'Last month' }
      ]);
    meOther = {
      isOther: true, initials: opDir.initials, color: opDir.color, aboutTitle: 'About ' + first, forumsTitle: 'Forums', reportsTo: reportOf(opDir.name).name || 'Top of the organization',
      name: opDir.name, role: opDir.title, dept: opDir.dept,
      status: opStatus.label, statusColor: opStatus.color, statusNote: opStatus.note, statusMenu: false, statuses: [],
      photo: '', hasPhoto: false, noPhoto: true, editing: false, notEditing: true,
      coverStyle: 'position:relative;height:190px;background:' + opDir.color + '40;overflow:hidden', covers: [],
      complete: 100, completeTip: '', completeBar: '', bio: opDir.name + ' works in ' + opDir.dept + ' as ' + opDir.title + '.', phone: '', newSkill: '',
      skills: (skillsByDept[opDir.dept] || ['Teamwork']).map((n) => ({ name: n, editing: false, remove: () => {} })),
      facts: [
        { label: 'Department', value: opDir.dept, sub: opDir.title },
        { label: 'Reports to', value: reportOf(opDir.name).top ? 'No one' : reportOf(opDir.name).name, sub: reportOf(opDir.name).sub },
        { label: 'Location', value: opDir.loc || 'Kenya', sub: 'GMT+3 · ' + localTime + ' now' },
        { label: 'Member since', value: '[JOIN DATE]', sub: 'Hub member' }
      ],
      stats: [
        { label: 'Tickets raised', n: opRaised.length, color: '#FF9B8A' },
        { label: 'Tickets handled', n: opHandled.length, color: '#7CC4FF' },
        { label: 'Wall posts', n: opPosts.length, color: '#F49AC1' },
        { label: 'Awards', n: opAwards.length, color: '#FFD66B' }
      ],
      projects: projects.filter((p) => opProjIds.indexOf(p.id) !== -1).map((p) => ({ ...p, myRole: 'Contributor' })),
      activity: withKind(opActivity),
      contact: [
        { label: 'Work email', value: '[' + opDir.name.toLowerCase().replace(/ /g, '.') + '@company.com]', color: '#E8EFEA' },
        { label: 'Phone', value: 'Not shared', color: '#9DB0A6' },
        { label: 'Office', value: opDir.loc || '[OFFICE / FLOOR]', color: '#E8EFEA' }
      ],
      localTime, hours: ts.start + ' to ' + ts.end,
      badges: opAwards.map((a) => ({ name: a.type, when: fmtDay(a.date), color: '#FFD66B', bg: '#2A2716' })), hasBadges: opAwards.length > 0,
      forums: forums.slice(0, 3).map((f) => ({ ...f, myRole: 'Member' }))
    };
  }
  const me = meOther || meOwn;
  const addSkill = () => {
    const v = this.state.newSkill.trim();
    if (!v) return;
    this.setState((st) => ({ newSkill: '', profile: { ...st.profile, skills: st.profile.skills.indexOf(v) === -1 ? st.profile.skills.concat([v]) : st.profile.skills } }));
  };
  return { addSkill, me, setPf };
}
