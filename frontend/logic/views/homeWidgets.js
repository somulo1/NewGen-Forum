import * as calendarW from '~/logic/server/widgets/calendar.js';
import * as commsW from '~/logic/server/widgets/comms.js';
import * as dmsW from '~/logic/server/widgets/dms.js';
// renderVals step: homeWidgets. Receives the names built by earlier steps and returns the ones later steps need.
export function buildHomeWidgets(ctx) {
  const { todayISO, calItems, fmtDay, sharedNotes, forums, has, ini, myUnits, nOpen, nOver, nSoon, openCm, openEv, orgWidgetItems, people, projStats, projects, s, sessions, tickets } = ctx;
  // ---- Files ----
  const kindColor = { PDF: '#FF9B8A', DOC: '#7CC4FF', XLS: '#B8F36B', PPT: '#FFB86B', IMG: '#F49AC1', ZIP: '#C9B6FF' };
  // "Shared notes" is a built-in folder: notes shared with me (live links) and the notes I share.
  const sn = sharedNotes;
  const noteFolder = { id: 'sharednotes', name: 'Shared notes', color: '#FFD66B', audience: 'you and the people you share notes with', updated: sn.in.length ? sn.in[0].when : 'no notes yet', files: [], count: sn.in.length + sn.out.length,
    samples: sn.in.concat(sn.out).slice(0, 3).map((x, i) => ({ kind: 'NOTE', color: '#FFD66B', h: [86, 74, 80][i] + 'px' })),
    open: () => this.setState({ folder: 'sharednotes' }) };
  const folders = [noteFolder].concat(s.folders.map((f) => ({ ...f, count: f.files.length,
    samples: f.files.slice(0, 3).map((x, i) => ({ kind: x.kind, color: kindColor[x.kind] || '#9DB0A6', h: [86, 74, 80][i] + 'px' })),
    open: () => this.setState({ folder: f.id }) })));
  const cur_f = s.folder === 'sharednotes' ? { id: 'sharednotes', name: 'Shared notes', color: '#FFD66B', audience: 'you and the people you share notes with', files: [], isNotes: true } : s.folders.find((f) => f.id === s.folder);
  const fv = cur_f ? { ...cur_f, sn, isDocs: !cur_f.isNotes, count: cur_f.isNotes ? sn.in.length + sn.out.length : cur_f.files.length, files: cur_f.files.map((x) => ({ ...x, color: kindColor[x.kind] || '#9DB0A6',
    isDoc: x.kind === 'PDF' || x.kind === 'DOC', isSheet: x.kind === 'XLS', isSlide: x.kind === 'PPT', isImg: x.kind === 'IMG', isZip: x.kind === 'ZIP' || !kindColor[x.kind],
    hasUrl: !!x.url, noUrl: !x.url, imgBg: x.imgBg || '#3A3150' })) } : {};

  const catalog = {
    announcements: { title: 'Announcements', sub: 'From management', tint: '#B8F36B', desc: 'Pinned company news. Staff mark each one as read, so admins can see who has seen it.' },
    events: { title: 'Upcoming events', sub: 'Next few weeks', tint: '#FFB86B', desc: 'Events across the organization with one-tap RSVP.' },
    clocks: { title: 'World clock', sub: 'Team time zones', tint: '#6EE7C8', desc: 'Live time in the cities your teams work from.' },
    birthdays: { title: 'Celebrations', sub: 'Birthdays & anniversaries', tint: '#F49AC1', desc: 'Who is celebrating today, with a quick way to send wishes.' },
    tickets: { title: 'Tickets', sub: 'Turnaround time tracking', tint: '#FF9B8A', desc: 'Open, overdue and due-soon tickets, the most urgent ones to view, and a quick way to raise a new one.' },
    projects: { title: 'Projects', sub: 'Progress and health', tint: '#7CC4FF', desc: 'Every project with its status, progress and due date, plus a View button into each one.' },
    awards: { title: 'Performance awards', sub: 'Recognition', tint: '#FFD66B', desc: 'Spotlight top performers, show award winners and let staff nominate colleagues.' },
    secureshare: { title: 'Secure share', sub: 'One-time links', tint: '#6EE7C8', desc: 'Share a password or credential through a link that self-destructs after one view.' },
    wall: { title: 'Community wall', sub: 'Latest posts', tint: '#F49AC1', desc: 'The newest posts from the wall, so nobody misses what colleagues are sharing.' },
    forums: { title: 'My forums', sub: 'Unread activity', tint: '#7CC4FF', desc: 'Your forums with unread counts. Open any one straight into the side panel.' },
    files: { title: 'Recent files', sub: 'Across your forums', tint: '#C9B6FF', desc: 'The latest documents shared with you.' },
    joiners: { title: 'New joiners', sub: 'This week', tint: '#B8F36B', desc: 'Welcome new colleagues and help them settle in.' },
    sessions: { title: 'Meetings & standups', sub: 'Live & upcoming', tint: '#6EE7C8', desc: 'Live meetings and daily standups you can join, plus a button to start one.' },
    notes: { title: 'My notes', sub: 'Only you can see this', tint: '#FFD66B', desc: 'A private sticky note for your own reminders.' },
    branches: { title: 'Branches', sub: 'Your locations', tint: '#7CC4FF', desc: 'For organizations with several locations: each branch, its head, staff count and contacts.' },
    sectors: { title: 'Sectors', sub: 'Top level of the structure', tint: '#FFB86B', desc: 'Groups branches (or departments) into sectors under the organization. Turns on sector dashboards and the Sector manager role.' },
    departments: { title: 'Departments', sub: 'Teams and how they perform', tint: '#B8F36B', desc: 'Departments sit under their branch or sector. Tickets go to departments, and each department gets a dashboard and a Head of department role.' },
    units: { title: 'Units', sub: 'Teams inside branches', tint: '#C9B6FF', desc: 'Units are the smallest level and sit inside a branch. Turns on unit dashboards and the Unit lead role.' },
    performance: { title: 'Our performance', sub: 'Against its own target', tint: '#6EE7C8', desc: 'How your department, branch, sector or unit resolves what it receives on time, against its target.' },
    calendar: { title: 'Calendar', sub: 'Coming up', tint: '#7CC4FF', desc: 'Your meetings, standups, tasks and events, with project deadlines. Schedule straight from here.' },
    memos: { title: 'Memos', sub: 'Latest memos', tint: '#FFD66B', desc: 'Send formal memos to the wall, a forum or chosen departments, branches, sectors or units.' },
    jobposts: { title: 'Job posts', sub: 'Open roles', tint: '#7CC4FF', desc: 'Post open roles and see how many people read and reacted to each.' },
    newsletter: { title: 'Newsletter', sub: 'Latest issues', tint: '#F49AC1', desc: 'Publish a regular newsletter and see how many people read it.' },
    chat: { title: 'Private messages', sub: 'One-to-one chats', tint: '#C9B6FF', desc: 'Private one-to-one chats with private voice and video calls. Opens in the same side panel, full screen or bubble as forums.' }
  };
  if (catalog.branches) catalog.branches.desc = 'Branches sit under their sector (if you use sectors) and hold units and departments. Turns on branch dashboards and the Branch manager role.';
  const pillBtn = (on) => 'flex:none;height:30px;padding:0 12px;border-radius:15px;cursor:pointer;font-size:12px;font-weight:600;' + (on ? 'border:0;background:var(--k-2C3D34,#2C3D34);color:var(--k-B8F36B,#B8F36B)' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA)');
  const tz = (zone) => { try { return new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: zone }); } catch (err) { return '--:--'; } };
  const tzDay = (zone) => { try { return new Date().toLocaleDateString('en-GB', { weekday: 'long', timeZone: zone }); } catch (err) { return ''; } };
  // ---- people-based widgets (real directory data) ----
  const mmdd = (iso) => (iso || '').slice(5, 10);
  const today = todayISO.slice(5, 10);
  const inDays = (n) => mmdd(new Date(new Date(todayISO + 'T12:00:00+03:00').getTime() + n * 86400000).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }));
  const soon = [0, 1, 2, 3, 4, 5, 6, 7].map(inDays);
  const celebrations = [];
  (this.dir || []).forEach((p) => {
    const d = soon.indexOf(mmdd(p.dob));
    if (p.dob && d !== -1) celebrations.push({ id: 'b-' + p.uid, name: p.name, initials: p.initials, color: p.color, what: d === 0 ? 'Birthday today' : d === 1 ? 'Birthday tomorrow' : 'Birthday ' + new Date(p.dob.slice(0, 4) && todayISO.slice(0, 4) + p.dob.slice(4) + 'T12:00:00+03:00').toLocaleDateString('en-GB', { weekday: 'long', timeZone: 'Africa/Nairobi' }), at: d });
    const yrs = p.joined && mmdd(p.joined) === today ? Number(todayISO.slice(0, 4)) - Number(p.joined.slice(0, 4)) : 0;
    if (yrs >= 1) celebrations.push({ id: 'a-' + p.uid, name: p.name, initials: p.initials, color: p.color, what: yrs + (yrs === 1 ? ' year' : ' years') + ' with the team today', at: 0 });
  });
  celebrations.sort((a, b) => a.at - b.at);
  const recent = (iso) => iso && (new Date(todayISO + 'T12:00:00+03:00') - new Date(iso + 'T12:00:00+03:00')) / 86400000 <= 14;
  const newJoiners = (this.dir || []).filter((p) => recent(p.joined)).sort((a, b) => (a.joined < b.joined ? 1 : -1)).slice(0, 3).map((p) => ({ name: p.name, role: [p.title, p.dept].filter(Boolean).join(' · '), initials: p.initials, color: p.color }));
  const itemsFor = {
    chat: s.dms.slice(0, 4).map((d) => { const pr = this.dir.find((x) => x.name === d.with) || { initials: ini(d.with), color: '#C9B6FF' }; const last = d.msgs[d.msgs.length - 1];
      return { name: d.with, initials: pr.initials, color: pr.color, unread: d.unread, hasUnread: d.unread > 0, last: last ? (last.from === 'me' ? 'You: ' : '') + (last.text || 'Sent a file') : 'Say hello', open: () => this.openDm(d.with) }; }),
    announcements: s.comms.filter((c) => c.type === 'Announcement' && c.status === 'Published').slice(0, 3).map((c) => ({ id: c.id, tag: (c.byDept || 'News').split(' ')[0], title: c.title, meta: c.by + ' · ' + fmtDay(c.date), ackLabel: c.iRead ? 'Read' : 'Mark read', ackStyle: pillBtn(!!c.iRead), ack: () => commsW.markRead(this, this.srv, c) })),
    events: calendarW.upcomingEvents(s.events, todayISO).map((e) => ({ id: e.id, mon: new Date(e.date + 'T12:00:00+03:00').toLocaleDateString('en-GB', { month: 'short', timeZone: 'Africa/Nairobi' }).toUpperCase(), day: String(Number(e.date.slice(8, 10))), title: e.title,
      meta: new Date(e.date + 'T12:00:00+03:00').toLocaleDateString('en-GB', { weekday: 'long', timeZone: 'Africa/Nairobi' }) + (e.start ? ' · ' + e.start : '') + (e.forum ? ' · ' + e.forum : ''),
      rsvpLabel: s.rsvp[e.id] ? 'Going' : 'RSVP', rsvpStyle: pillBtn(!!s.rsvp[e.id]), rsvp: () => calendarW.toggleRsvp(this, this.srv, e).catch((x) => this.notify(x.message, '#FF9B8A')) })),
    clocks: [['Nairobi', 'Africa/Nairobi'], ['London', 'Europe/London'], ['New York', 'America/New_York'], ['Dubai', 'Asia/Dubai']].map(([city, zone]) => ({ city, time: tz(zone), day: tzDay(zone) })),
    // Celebrations and new joiners come from the people directory (birthdays only where the viewer may see them).
    // "Send wish" sends a real private message (end-to-end encrypted like any chat). Whether it was sent is read back from
    // the chat itself, so it stays "Sent" after a reload and never goes twice on the same day.
    birthdays: celebrations.map((b) => {
      const mine = !!this.srv && b.name === this.srv.me.name;
      const first = b.name.split(' ')[0];
      const text = b.id.startsWith('b-') ? 'Happy birthday, ' + first + '! \uD83C\uDF82 Wishing you a wonderful day.' : 'Happy work anniversary, ' + first + '! \uD83C\uDF89 Congratulations on ' + b.what.replace(' with the team today', '') + ' with the team. Thank you for everything you do.';
      const dm = this.srv ? (s.dms || []).find((d) => d.with === b.name && !String(d.id).startsWith('dm-')) : null;
      const day = (iso) => new Date(iso).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' });
      const sent = this.srv ? !!dm && dm.msgs.some((m) => m.from === 'me' && m.text === text && m.at && day(m.at) === todayISO) : !!s.wished[b.id];
      const busy = !!(this._wishing && this._wishing.has(b.id));
      return { ...b, wishLabel: mine ? 'Your day' : sent ? 'Sent' : busy ? 'Sending…' : 'Send wish', wishStyle: pillBtn(mine || sent || busy),
        wish: () => {
          if (mine || sent || busy) return;
          if (!this.srv) { this.setState((st) => ({ wished: { ...st.wished, [b.id]: true } })); return; }
          (this._wishing || (this._wishing = new Set())).add(b.id);
          this.forceUpdate();
          dmsW.ensure(this, this.srv, b.name)
            .then((id) => dmsW.send(this, this.srv, id, { text }).then((conv) => dmsW.refresh(this, this.srv, conv)))
            .then(() => this.notify('Your wish was sent to ' + first + '.', '#B8F36B'))
            .catch((e) => this.notify((e && e.message) || 'Could not send your wish.', '#FF9B8A'))
            .finally(() => { this._wishing.delete(b.id); this.forceUpdate(); });
        } };
    }),
    tickets: [{ n: nOpen, label: 'Open', color: '#E8EFEA' }, { n: nOver, label: 'Overdue', color: '#FF9B8A' }, { n: nSoon, label: 'Due soon', color: '#FFB86B' }],
    // only files that really exist in the folders this person can open, newest first
    files: s.folders.flatMap((f) => f.files.map((x) => ({ kind: x.kind, color: kindColor[x.kind] || '#9DB0A6', name: x.name, meta: f.name + ' · ' + x.when, at: x.at }))).sort((a, b) => (a.at < b.at ? 1 : -1)).slice(0, 4),
    joiners: newJoiners,
    sessions: sessions,
    projects: projects,
    wall: s.posts.slice(0, 2).map((p) => ({ author: p.author, initials: p.initials, color: p.color, text: (p.text || 'Shared a file').length > 90 ? p.text.slice(0, 90) + '…' : (p.text || 'Shared a file'), likes: p.likes + (s.liked[p.id] ? 1 : 0) })),
    forums: forums.filter((f) => f.hasUnread).concat(forums.filter((f) => !f.hasUnread)).slice(0, 4),
  };
  // structure widgets follow the organization's structure, not the person's choice
  const LEVELS = ['departments', 'sectors', 'branches', 'units'];
  const ids = s.widgets.filter((x) => catalog[x] && (LEVELS.indexOf(x) === -1 || has[x]));
  const isHidden = (id) => s.hiddenWidgets.indexOf(id) !== -1;
  const toggleHide = (id) => () => this.setState((st) => ({ hiddenWidgets: st.hiddenWidgets.indexOf(id) !== -1 ? st.hiddenWidgets.filter((x) => x !== id) : st.hiddenWidgets.concat([id]) }));
  const removeWidget = (id) => this.setState((st) => {
    const patch = { widgets: st.widgets.filter((x) => x !== id) };
    if (id === 'projects') Object.assign(patch, {
      section: st.section === 'projects' ? 'home' : st.section,
      projectOpen: null, projectForm: null, taskForm: null, taskOpen: null, msDialog: null
    });
    return patch;
  });
  const widgets = ids.filter((x) => x !== 'awards' && (s.customize || !isHidden(x))).map((id, i) => ({
    hidden: isHidden(id), hideLabel: isHidden(id) ? 'Show' : 'Hide', toggleHide: toggleHide(id),
    ...catalog[id],
    title: (s.wcfg[id] && s.wcfg[id].title) || catalog[id].title,
    openSettings: () => this.setState({ settingsFor: id }),
    t: { [id]: true },
    items: id === 'calendar' ? calItems : (id === 'memos' || id === 'newsletter' || id === 'jobposts') ? s.comms.filter((c) => c.type === (id === 'memos' ? 'Memo' : id === 'jobposts' ? 'Job post' : 'Newsletter') && c.status === 'Published').slice(0, 3).map((c) => ({ ...c, by: c.by === 'Samwel Okoth Omulo' ? 'You' : c.by, dateLabel: fmtDay(c.date), audience: c.to.length > 2 ? c.to.slice(0, 2).join(', ') + ' +' + (c.to.length - 2) : c.to.join(', '), readBar: 'height:100%;width:' + c.read + '%;background:#B8F36B' })) : (orgWidgetItems[id] || itemsFor[id] || []),
    isCommList: id === 'memos' || id === 'newsletter' || id === 'jobposts',
    noComms: (id === 'memos' || id === 'newsletter' || id === 'jobposts') && !s.comms.some((c) => c.type === (id === 'memos' ? 'Memo' : id === 'jobposts' ? 'Job post' : 'Newsletter')),
    compose: () => openCm(id === 'memos' ? 'Memo' : id === 'newsletter' ? 'Newsletter' : id === 'jobposts' ? 'Job post' : 'Announcement'),
    composeLabel: id === 'memos' ? 'New memo' : id === 'newsletter' ? 'New issue' : id === 'jobposts' ? 'New job post' : 'New announcement',
    goComms: () => this.setState({ section: 'comms' }),
    newEvent: () => openEv('Personal'), goCalendar: () => this.setState({ section: 'calendar' }),
    isOrgList: id === 'departments' || id === 'units',
    orgNote: id === 'departments' ? 'Each sits under ' + (has.branches ? 'its branch' : has.sectors ? 'its sector' : 'the organization') + '. % = received items resolved on time.' : 'Units sit inside ' + (has.branches ? 'their branch' : 'the organization') + '.',
    goStructure: () => this.setState({ section: 'structure' }),
    noUnits: id === 'performance' && myUnits.length === 0,
    customizing: s.customize,
    note: s.note,
    onNote: (e) => this.setState({ note: e.target.value }),
    goTickets: () => this.setState({ section: 'tickets', ticketOpen: null }),
    goProjects: () => this.setState({ section: 'projects', projectOpen: null }),
    urgent: tickets.filter((t) => t.step < 3).slice().sort((a, b) => ['Overdue', 'Due soon', 'On time', 'TAT paused'].indexOf(a.slaLabel) - ['Overdue', 'Due soon', 'On time', 'TAT paused'].indexOf(b.slaLabel)).slice(0, 3),
    pstats: projStats.slice(1),
    goMeetings: () => this.setState({ section: 'meetings' }),
    goWall: () => this.setState({ section: 'feed' }),
    goPeople: () => this.setState({ section: 'people' }),
    nominated: !!s.nominated,
    nominateLabel: s.nominated ? 'Nomination sent' : 'Nominate a colleague',
    nominate: () => this.setState({ nominated: true }),
    shareMade: false,
    shareIdle: true,
    makeShare: () => this.setState({ section: 'share' }),
    resetShare: () => this.setState({ section: 'share' }),
    cardStyle: 'break-inside:avoid;margin:0 0 16px;box-sizing:border-box;width:100%;padding:18px;border-radius:18px;background:var(--k-16201C,#16201C);display:flex;flex-direction:column;gap:14px;border:' + (s.customize ? '1px dashed var(--k-4A5C54,#4A5C54)' : '1px solid var(--k-24322C,#24322C)'),
    remove: () => removeWidget(id),
    moveUp: () => this.setState((st) => {
      const w = st.widgets.slice(); const k = w.indexOf(id);
      if (k > 0) { const tmp = w[k - 1]; w[k - 1] = w[k]; w[k] = tmp; }
      return { widgets: w };
    })
  }));
  const gallery = Object.keys(catalog).filter((id) => (id !== 'performance' || has.structure) && (LEVELS.indexOf(id) === -1 || has[id])).map((id) => {
    const added = ids.indexOf(id) !== -1;
    return { ...catalog[id], added, hidden: added && isHidden(id), hideLabel: isHidden(id) ? 'Show on Home' : 'Hide on Home', toggleHide: toggleHide(id), remove: () => removeWidget(id), btnLabel: added ? 'Added' : 'Add to home',
      openSettings: () => this.setState({ settingsFor: id }),
      btnStyle: 'flex:1;height:38px;border-radius:10px;font-weight:600;cursor:' + (added ? 'default' : 'pointer') + ';' + (added ? 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-9DB0A6,#9DB0A6)' : 'border:0;background:#B8F36B;color:#0F1714'),
      add: () => { if (!added) this.setState((st) => ({ widgets: st.widgets.concat(id) })); } };
  });
  // today's birthdays, for the Community wall card
  const bToday = celebrations.filter((c) => c.at === 0 && /^Birthday/.test(c.what));
  const wallBirthdays = { has: bToday.length > 0, text: bToday.length ? bToday[0].name + (bToday.length > 1 ? ' and ' + (bToday.length - 1) + (bToday.length === 2 ? ' other' : ' others') : '') + '. Send them a wish on the wall.' : '', names: bToday.map((c) => c.name) };
  return { catalog, cur_f, folders, fv, gallery, ids, isHidden, toggleHide, wallBirthdays, widgets };
}
