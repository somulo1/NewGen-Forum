import * as profileW from '~/logic/server/widgets/profile.js';
// renderVals step: profile. Receives the names built by earlier steps and returns the ones later steps need.
export function buildProfile(ctx) {
  const { reportOf, fmtDay, forums, has, nav, projects, s, skillsByDept, ts } = ctx;
  // ---- My profile ----
  nav.profileCard = 'margin-top:auto;display:flex;align-items:center;gap:10px;width:100%;min-height:58px;padding:8px 10px;border-radius:12px;cursor:pointer;color:var(--k-E8EFEA,#E8EFEA);font-family:inherit;border:1px solid ' + (s.section === 'profile' && !s.profileOf ? '#3F524A' : '#24322C') + ';background:' + (s.section === 'profile' && !s.profileOf ? 'var(--k-22302A,#22302A)' : 'var(--k-16201C,#16201C)');
  const pf = s.profile;
  const statusDefs = [
    { label: 'Available', color: '#B8F36B', note: 'Available for chats and calls' },
    { label: 'In a meeting', color: '#FFB86B', note: 'In a meeting, replies may be slow' },
    { label: 'Focusing', color: '#7CC4FF', note: 'Focusing, notifications paused' },
    { label: 'Away', color: '#9DB0A6', note: 'Away from desk' }
  ];
  const curStatus = statusDefs.find((x) => x.label === pf.status) || statusDefs[0];
  const setPf = (patch) => this.setState((st) => ({ profile: Object.assign({}, st.profile, patch) }));
  // status and cover save straight away; text fields save with "Save" in edit mode
  const saveNow = (patch) => { setPf(patch); if (this.srv && !s.profileEdit) profileW.save(this, this.srv).catch((e) => this.notify(e.message || 'Could not save.', '#FF9B8A')); };
  const mi = s.meInfo || { name: 'Samwel Okoth Omulo', initials: 'S', color: '#C2185B', email: '', title: '', dept: '', branch: '', role: '', joined: '' };
  const joinedLabel = mi.joined ? fmtDay(mi.joined) : '';
  // The organization's time zone (Settings > Regional); the profile shows the time there and its offset from GMT.
  const orgTz = (s.orgProfile && s.orgProfile.timezone) || 'Africa/Nairobi';
  let localTime = '', gmt = 'GMT';
  try {
    localTime = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: orgTz });
    const off = new Intl.DateTimeFormat('en-GB', { timeZone: orgTz, timeZoneName: 'longOffset' }).formatToParts(new Date()).find((x) => x.type === 'timeZoneName');
    gmt = off ? off.value.replace(/:00$/, '').replace(/([+-])0(\d)/, '$1$2').replace(/^GMT$/, 'GMT+0') : 'GMT';
  } catch (err) { localTime = ''; }
  const where = (branch) => branch || (s.orgProfile && s.orgProfile.country) || 'Not set';
  // Who a person answers to, from the server (their manager, the person their role reports to, or the role). Without it, the local rules.
  const reportLine = (rt, fallback) => (!rt ? fallback : rt.kind === 'person' ? { display: rt.person.name, sub: rt.role ? rt.role + ' · reporting line' : 'Manager' } : rt.kind === 'role' ? { display: rt.role, sub: 'Role' } : { display: 'No one', sub: 'Top of the organization' });
  // What someone did in the forums you can see: their latest messages.
  const forumActivity = (who) => (s.forumsLive || []).flatMap((f) => f.updates.filter((u) => u.who === who && u.text).slice(0, 1).map((u) => ({ tag: 'F', color: '#6EE7C8', text: 'Posted in ' + f.name + ': "' + (u.text.length > 70 ? u.text.slice(0, 70) + '…' : u.text) + '"', time: u.time, at: u.at || '' }))).sort((a, b) => (a.at < b.at ? 1 : -1)).slice(0, 2);
  const checks = [!!pf.photo, (pf.bio || '').trim().length > 20, pf.skills.length >= 3, true, !!pf.phone.trim()];
  const complete = Math.round((checks.filter(Boolean).length / checks.length) * 100);
  const tip = !pf.photo ? 'Add a profile photo so colleagues recognise you.' : !pf.phone.trim() ? 'Add your phone number to reach 100%.' : (pf.bio || '').trim().length <= 20 ? 'Write a short bio.' : 'Your profile is complete.';
  const myPosts = s.posts.filter((p) => p.author === 'Sam');
  const myRaised = s.tickets.filter((t) => t.raisedBy === 'Sam');
  const myActivity0 = myPosts.slice(0, 2).map((p) => ({ tag: 'W', color: '#F49AC1', text: 'Posted on the wall: "' + ((p.text || 'a file').length > 70 ? p.text.slice(0, 70) + '…' : (p.text || 'a file')) + '"', time: p.meta.split('· ').pop() }))
    .concat(myRaised.slice(0, 2).map((t) => ({ tag: 'T', color: '#FF9B8A', text: 'Raised ' + t.id + ': ' + t.title, time: t.created })))
    .concat(s.tickets.filter((t) => t.handler === mi.name).slice(0, 2).map((t) => ({ tag: 'A', color: '#7CC4FF', text: 'Handling ' + t.id + ': ' + t.title, time: t.handlerSince || t.created })))
    .concat(s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(mi.name) !== -1).map((a) => ({ tag: 'B', color: '#FFD66B', text: 'Earned the ' + a.type + ' award', time: fmtDay(a.date) })))
    .concat(forumActivity('Sam'));
  const myActivity = myActivity0;
  const myReport = reportLine(mi.reportsTo, reportOf(mi.name));
  const kindOf = { W: 'Wall post', T: 'Ticket', F: 'File', A: 'Assignment', B: 'Badge', P: 'Project' };
  const withKind = (arr) => arr.map((a) => ({ ...a, kind: kindOf[a.tag] || 'Activity' }));
  const meOwn = {
    isOther: false, initials: mi.initials, color: mi.color, aboutTitle: 'About me', forumsTitle: 'My forums', reportsTo: myReport.display,
    name: mi.name, email: mi.email, role: mi.title, dept: mi.dept,
    status: curStatus.label, statusColor: curStatus.color, statusNote: curStatus.note, statusMenu: s.statusMenu,
    statuses: statusDefs.map((x) => ({ ...x, pick: () => { this.setState({ statusMenu: false }); saveNow({ status: x.label }); } })),
    photo: pf.photo, hasPhoto: !!pf.photo, noPhoto: !pf.photo,
    editing: s.profileEdit, notEditing: !s.profileEdit,
    coverStyle: 'position:relative;height:190px;background:' + pf.cover + ';overflow:hidden;border-radius:21px 21px 0 0',
    covers: [['#1F3A2B', 'Forest'], ['#1E3347', 'Ocean'], ['#3A2433', 'Plum'], ['#3D3320', 'Sand']].map(([c, label]) => ({ label: label + ' cover', pick: () => saveNow({ cover: c }),
      style: 'width:26px;height:26px;border-radius:50%;cursor:pointer;background:' + c + ';border:2px solid ' + (pf.cover === c ? '#E8EFEA' : 'rgba(255,255,255,0.25)') })),
    complete, completeTip: tip, completeBar: 'height:100%;width:' + complete + '%;background:' + (complete === 100 ? '#B8F36B' : '#FFB86B') + ';border-radius:3px',
    bio: pf.bio, phone: pf.phone, newSkill: s.newSkill,
    skills: pf.skills.map((n, i) => ({ name: n, editing: s.profileEdit, remove: () => this.setState((st) => ({ profile: { ...st.profile, skills: st.profile.skills.filter((x, j) => j !== i) } })) })),
    facts: [
      { label: 'Department', value: mi.dept || 'Not set', sub: mi.unit || mi.title },
      { label: 'Reports to', value: myReport.display, sub: myReport.sub },
      { label: 'Location', value: where(mi.branch), sub: gmt + ' · ' + localTime + ' now' },
      { label: 'Member since', value: joinedLabel || 'Not recorded', sub: mi.role || 'Hub member' }
    ],
    stats: [
      { label: 'Tickets raised', n: myRaised.length, color: '#FF9B8A' },
      { label: 'Tickets handled', n: s.tickets.filter((t) => t.handler === 'Sam' || t.assignees.indexOf('Sam') !== -1).length, color: '#7CC4FF' },
      { label: 'Wall posts', n: myPosts.length, color: '#F49AC1' },
      { label: 'Awards', n: s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(mi.name) !== -1).length, color: '#FFD66B' }
    ],
    projects: projects.filter((p) => s.projects.some((x) => x.id === p.id && x.team.indexOf(mi.name) !== -1)).map((p) => { const x = s.projects.find((y) => y.id === p.id); return { ...p, myRole: x && x.owner === mi.name ? 'Owner' : 'Team member' }; }),
    activity: withKind(myActivity),
    contact: [
      { label: 'Work email', value: mi.email || 'Not set', color: '#E8EFEA' },
      { label: 'Phone', value: pf.phone.trim() || 'Not added yet', color: pf.phone.trim() ? '#E8EFEA' : '#FFB86B' },
      { label: 'Office', value: [mi.branch, mi.unit].filter(Boolean).join(' · ') || 'Not set', color: '#E8EFEA' }
    ],
    localTime, hours: ts.start + ' to ' + ts.end,
    badges: has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(mi.name) !== -1).map((a) => ({ name: a.type, when: fmtDay(a.date), color: '#FFD66B', bg: '#2A2716' })) : [],
    hasBadges: has.awards && s.awardList.some((a) => a.status === 'Approved' && a.people.indexOf(mi.name) !== -1),
    forums: forums.filter((f) => !f.notMember).slice(0, 6).map((f) => ({ ...f, myRole: f.myRole || 'Member' }))
  };
  // Full profile of another person (only for roles with the "View full profiles" right)
  const opDir = s.profileOf ? this.dir.find((x) => x.name === s.profileOf) : null;
  let meOther = null;
  if (opDir) {
    const det = this.srv && opDir.uid ? (this.srv.people || {})[opDir.uid] : null;
    if (this.srv && opDir.uid && det === undefined) profileW.loadPerson(this, this.srv, opDir.uid);
    const pd = det && det !== 'loading' ? det : null;
    const first = opDir.name.split(' ')[0];
    const mine = (v) => v === opDir.name || v === first;
    const opStatus = statusDefs.find((x) => x.label === opDir.status) || statusDefs[0];
    const opRaised = s.tickets.filter((t) => mine(t.raisedBy));
    const opHandled = s.tickets.filter((t) => mine(t.handler) || (t.assignees || []).some(mine));
    const opPosts = s.posts.filter((p) => mine(p.author));
    const opReport = reportLine(pd && pd.reports_to, reportOf(opDir.name));
    const opAwards = has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(opDir.name) !== -1) : [];
    const opProjIds = s.projects.filter((p) => (p.tasks || []).some((t) => mine(t.who))).map((p) => p.id);
    const opActivity = opPosts.slice(0, 2).map((p) => ({ tag: 'W', color: '#F49AC1', text: 'Posted on the wall: "' + ((p.text || 'a file').length > 70 ? p.text.slice(0, 70) + '…' : (p.text || 'a file')) + '"', time: p.meta.split('· ').pop() }))
      .concat(opRaised.slice(0, 2).map((t) => ({ tag: 'T', color: '#FF9B8A', text: 'Raised ' + t.id + ': ' + t.title, time: t.created })))
      .concat(opHandled.slice(0, 2).map((t) => ({ tag: 'A', color: '#7CC4FF', text: 'Handling ' + t.id + ': ' + t.title, time: t.created })))
      .concat(opAwards.map((a) => ({ tag: 'B', color: '#FFD66B', text: 'Earned the ' + a.type + ' award', time: fmtDay(a.date) })))
      .concat(forumActivity(opDir.name));
    meOther = {
      isOther: true, initials: opDir.initials, color: opDir.color, aboutTitle: 'About ' + first, forumsTitle: 'Forums', reportsTo: opReport.display,
      name: opDir.name, email: (pd && pd.email) || opDir.email || '', role: opDir.title, dept: opDir.dept,
      status: opStatus.label, statusColor: opStatus.color, statusNote: opStatus.note, statusMenu: false, statuses: [],
      photo: '', hasPhoto: false, noPhoto: true, editing: false, notEditing: true,
      coverStyle: 'position:relative;height:190px;background:' + ((pd && pd.cover) || opDir.color + '40') + ';overflow:hidden', covers: [],
      complete: 100, completeTip: '', completeBar: '', bio: (pd && pd.bio) || opDir.name + ' works in ' + opDir.dept + ' as ' + opDir.title + '.', phone: (pd && pd.phone) || '', newSkill: '',
      skills: ((pd && pd.skills && pd.skills.length ? pd.skills : null) || skillsByDept[opDir.dept] || ['Teamwork']).map((n) => ({ name: n, editing: false, remove: () => {} })),
      facts: [
        { label: 'Department', value: opDir.dept, sub: opDir.title },
        { label: 'Reports to', value: opReport.display, sub: opReport.sub },
        { label: 'Location', value: where(opDir.loc), sub: gmt + ' · ' + localTime + ' now' },
        { label: 'Member since', value: pd && pd.joined_on ? fmtDay(pd.joined_on) : 'Not recorded', sub: (pd && pd.role_name) || 'Hub member' }
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
        { label: 'Work email', value: (pd && pd.email) || opDir.email || 'Not shared', color: '#E8EFEA' },
        { label: 'Phone', value: (pd && pd.phone) || 'Not shared', color: (pd && pd.phone) ? '#E8EFEA' : '#9DB0A6' },
        { label: 'Office', value: opDir.loc || '[OFFICE / FLOOR]', color: '#E8EFEA' }
      ],
      localTime, hours: ts.start + ' to ' + ts.end,
      badges: opAwards.map((a) => ({ name: a.type, when: fmtDay(a.date), color: '#FFD66B', bg: '#2A2716' })), hasBadges: opAwards.length > 0,
      forums: (s.forumsLive ? forums.filter((f) => (f.memberList || []).some((m) => m.name === opDir.name)).slice(0, 6).map((f) => ({ ...f, myRole: ((f.memberList.find((m) => m.name === opDir.name) || {}).role || 'member').replace(/^./, (c) => c.toUpperCase()) })) : forums.slice(0, 3).map((f) => ({ ...f, myRole: 'Member' })))
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
