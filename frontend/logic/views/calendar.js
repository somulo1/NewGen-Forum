import * as calendarW from '~/logic/server/widgets/calendar.js';
// renderVals step: calendar. Receives the names built by earlier steps and returns the ones later steps need.
export function buildCalendar(ctx) {
  const { C, chip, corePeople, dowOf, fmtDay, has, isoAdd, personOf, s, set, todayISO } = ctx;
  // =============== Calendar ===============
  const typeColor = { Meeting: '#7CC4FF', Standup: '#6EE7C8', Task: '#FFB86B', Event: '#C9B6FF', Personal: '#F49AC1', Deadline: '#FF9B8A' };
  const occurs = (e, iso) => {
    if (e.repeat === 'None' || !e.repeat) return e.date === iso;
    if (iso < e.date) return false;
    const d = dowOf(iso);
    if (e.repeat === 'Daily') return true;
    if (e.repeat === 'Weekdays') return d >= 1 && d <= 5;
    if (e.repeat === 'Weekly') return d === dowOf(e.date);
    return false;
  };
  const deadlines = [];
  if (has.projects) s.projects.forEach((p) => p.tasks.forEach((t) => { if (t.col < 3) deadlines.push({ id: 'dl-' + p.id + t.title, title: t.title, type: 'Deadline', date: t.due, start: '', end: '', people: [t.primary].concat(t.secondary ? [t.secondary] : []), forum: '', repeat: 'None', project: p.name }); }));
  const itemsOn = (iso) => s.events.filter((e) => !s.calHide[e.type] && (has.sessions || (e.type !== 'Meeting' && e.type !== 'Standup')) && occurs(e, iso)).concat(s.calHide.Deadline ? [] : deadlines.filter((d) => d.date === iso)).sort((a, b) => (a.start || '99').localeCompare(b.start || '99'));
  const cy = Number(s.calMonth.slice(0, 4)), cmn = Number(s.calMonth.slice(5, 7));
  const firstISO = s.calMonth + '-01';
  const offset = (dowOf(firstISO) + 6) % 7;
  const dim = new Date(Date.UTC(cy, cmn, 0)).getUTCDate();
  const weeks = Math.ceil((offset + dim) / 7);
  const monthShift = (d) => { let y = cy, m = cmn + d; if (m < 1) { m = 12; y--; } if (m > 12) { m = 1; y++; } return y + '-' + (m < 10 ? '0' : '') + m; };
  const chipStyleC = (c) => 'display:block;max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding:2px 6px;border-radius:5px;font-size:10px;font-weight:600;background:' + c + '26;color:' + c;
  // Past items are crossed out. Upcoming ones can be closed, but only by the person who set them.
  const ME = this.srv ? this.srv.me.name : 'Samwel Okoth Omulo';   // who the signed-in person is, as events name their organizer
  const nowHM = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi', hour12: false });
  const isPast = (e, iso) => iso < todayISO || (iso === todayISO && !!(e.end || e.start) && (e.end || e.start) < nowHM);
  const isClosed = (e, iso) => (e.closedDates || []).indexOf(iso) !== -1;
  const ended = (e, iso) => isPast(e, iso) || isClosed(e, iso);
  const isMine = (e) => e.type !== 'Deadline' && (e.by === ME || e.by === 'Sam');
  const joinEvent = (e) => this.startCall(e.type === 'Standup' ? 'voice' : 'video', e.title, e.title.split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase(), typeColor[e.type], e.people, null, e.container || null, e.by && e.by !== ME && e.by !== 'Sam' ? 'in' : 'out');
  const cal = {
    monthLabel: new Date(firstISO + 'T12:00:00+03:00').toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'Africa/Nairobi' }),
    prevMonth: () => { this.setState({ calMonth: monthShift(-1) }); calendarW.ensureMonth(this, this.srv, monthShift(-1)); }, nextMonth: () => { this.setState({ calMonth: monthShift(1) }); calendarW.ensureMonth(this, this.srv, monthShift(1)); },
    today: () => this.setState({ calMonth: todayISO.slice(0, 7), calSel: todayISO }),
    weekdays: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map((l) => ({ l: l })),
    legend: [['Personal', 'My own'], ['Meeting', 'Meetings'], ['Standup', 'Standups'], ['Task', 'Tasks'], ['Event', 'Events'], ['Deadline', 'Project deadlines']].filter(([k]) => (k !== 'Deadline' || has.projects) && ((k !== 'Meeting' && k !== 'Standup') || has.sessions)).map(([k, l]) => ({ label: l, color: typeColor[k], on: !s.calHide[k],
      toggle: () => this.setState((x) => ({ calHide: { ...x.calHide, [k]: !x.calHide[k] } })),
      style: 'height:32px;padding:0 12px;border-radius:16px;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:6px;' + (s.calHide[k] ? 'border:1px dashed var(--k-3F524A,#3F524A);background:transparent;color:var(--k-9DB0A6,#9DB0A6)' : 'border:1px solid var(--k-2A3832,#2A3832);background:var(--k-16201C,#16201C);color:var(--k-E8EFEA,#E8EFEA)') })),
    days: Array.from({ length: weeks * 7 }, (_, i) => {
      const iso = isoAdd(firstISO, i - offset);
      const inMonth = iso.slice(0, 7) === s.calMonth;
      const its = itemsOn(iso);
      const isSel = iso === s.calSel, isToday = iso === todayISO;
      return { n: Number(iso.slice(8, 10)), aria: fmtDay(iso) + ', ' + its.length + ' items',
        pick: () => this.setState({ calSel: iso }),
        style: 'min-height:108px;padding:6px;border:0;border-right:1px solid var(--k-24322C,#24322C);border-bottom:1px solid var(--k-24322C,#24322C);display:flex;flex-direction:column;gap:3px;text-align:left;cursor:pointer;font-family:inherit;color:var(--k-E8EFEA,#E8EFEA);overflow:hidden;min-width:0;background:' + (isSel ? 'var(--k-22302A,#22302A)' : inMonth ? 'var(--k-16201C,#16201C)' : 'var(--k-121B17,#121B17)') + ';' + (inMonth ? '' : 'opacity:0.5'),
        numStyle: isToday ? 'align-self:flex-start;min-width:24px;height:24px;padding:0 6px;box-sizing:border-box;border-radius:12px;background:#B8F36B;color:#0F1714;font-weight:700;font-size:12px;display:flex;align-items:center;justify-content:center' : 'font-size:12px;font-weight:600;color:var(--k-C9D6CF,#C9D6CF);padding:2px 4px',
        items: its.slice(0, 3).map((e) => ({ label: (e.start ? e.start + ' ' : '') + e.title, style: chipStyleC(typeColor[e.type]) + (ended(e, iso) ? ';text-decoration:line-through;opacity:0.55' : '') })),
        hasMore: its.length > 3, more: its.length - 3 };
    }),
    selLabel: (s.calSel === todayISO ? 'Today · ' : '') + fmtDay(s.calSel),
    agenda: itemsOn(s.calSel).map((e) => ({
      ended: ended(e, s.calSel), endedLabel: isClosed(e, s.calSel) ? 'Closed' : 'Ended',
      cardStyle: 'padding:12px;border-radius:12px;background:var(--k-121B17,#121B17);border:1px solid ' + typeColor[e.type] + ';display:flex;flex-direction:column;gap:6px;opacity:' + (ended(e, s.calSel) ? '0.65' : '1'),
      title: e.type === 'Deadline' ? 'Due: ' + e.title : e.title, type: e.type === 'Deadline' ? 'Project deadline' : e.type, color: typeColor[e.type],
      typeStyle: 'padding:2px 8px;border-radius:8px;font-size:11px;font-weight:600;background:' + typeColor[e.type] + '22;color:' + typeColor[e.type],
      time: e.start ? e.start + (e.end ? '–' + e.end : '') : 'All day',
      sub: [e.project, e.forum ? e.forum + ' forum' : '', e.repeat && e.repeat !== 'None' ? 'Repeats ' + e.repeat.toLowerCase() : '', e.notes || '', e.visibility ? 'Visible to: ' + e.visibility.toLowerCase() : '', e.remind && e.remind !== 'None' ? 'Reminder: ' + e.remind.toLowerCase() : ''].filter(Boolean).join(' · '),
      hasSub: !!(e.project || e.forum || (e.repeat && e.repeat !== 'None') || e.notes || e.visibility),
      people: e.people.slice(0, 6).map(personOf), hasPeople: e.people.length > 0, peopleLabel: e.people.length + (e.people.length === 1 ? ' person' : ' people'),
      canJoin: (e.type === 'Meeting' || e.type === 'Standup') && s.calSel === todayISO && !ended(e, s.calSel), join: () => joinEvent(e),
      canClose: isMine(e) && !ended(e, s.calSel), closeLabel: e.type === 'Task' ? 'Mark done' : 'Close',
      closeIt: () => { calendarW.update(this, this.srv, e, { closedDates: (e.closedDates || []).concat([s.calSel]) }).catch((x) => this.notify(x.message, '#FF9B8A')); this.notify('"' + e.title + '" ' + (e.type === 'Task' ? 'marked done.' : 'closed.'), '#7CC4FF'); },
      canRemove: isMine(e) && !ended(e, s.calSel), removeLabel: 'Cancel',
      remove: () => { calendarW.remove(this, this.srv, e).catch((x) => this.notify(x.message, '#FF9B8A')); this.notify(e.type === 'Task' ? '"' + e.title + '" marked done.' : '"' + e.title + '" cancelled. Participants were told.', '#7CC4FF'); }
    }))
  };
  cal.noAgenda = cal.agenda.length === 0;
  const openEv = (type) => this.setState({ evOpen: true, ev: { title: '', type: type, date: s.calSel || todayISO, start: type === 'Standup' ? '09:15' : '10:00', end: type === 'Standup' ? '09:30' : '11:00', repeat: type === 'Standup' ? 'Weekdays' : 'None', forum: '', people: [], pq: '', notes: '', visibility: 'Only me', remind: '15 minutes before' } });
  const evS = s.ev;
  const setEv = (patch) => this.setState((x) => ({ ev: Object.assign({}, x.ev, patch) }));
  const evNeeds = evS.type === 'Meeting' || evS.type === 'Standup';
  const pqv = evS.pq.trim().toLowerCase();
  const candidates = pqv ? corePeople.concat(this.dir.map((d) => d.name)).filter((n, i, a) => a.indexOf(n) === i && n.toLowerCase().indexOf(pqv) !== -1).slice(0, 8) : corePeople.slice(0, 8);
  const titleOf = (n) => ((this.dir.find((d) => d.name === n) || {}).title) || 'Staff';
  const ev = {
    heading: 'New ' + evS.type.toLowerCase(), title: evS.title, date: evS.date, start: evS.start, end: evS.end, repeat: evS.repeat, forum: evS.forum, pq: evS.pq,
    setTitle: (e) => setEv({ title: e.target.value }), setDate: (e) => setEv({ date: e.target.value }), setStart: (e) => setEv({ start: e.target.value }), setEnd: (e) => setEv({ end: e.target.value }),
    setRepeat: (e) => setEv({ repeat: e.target.value }), setForum: (e) => setEv({ forum: e.target.value }), setPq: (e) => setEv({ pq: e.target.value }),
    types: [['Personal', 'Just for me'], ['Task', 'Task'], ['Event', 'Event'], ['Meeting', 'Meeting'], ['Standup', 'Standup']].filter(([t]) => (t !== 'Meeting' && t !== 'Standup') || has.sessions).map(([t, l]) => ({ label: l, pick: () => setEv({ type: t, repeat: t === 'Standup' ? 'Weekdays' : evS.repeat }), style: chip(evS.type === t) })),
    needsPeople: evNeeds, isOwn: !evNeeds, count: evS.people.length,
    notes: evS.notes || '', visibility: evS.visibility || 'Only me', remind: evS.remind || 'None',
    setNotes: (e) => setEv({ notes: e.target.value }), setVisibility: (e) => setEv({ visibility: e.target.value }), setRemind: (e) => setEv({ remind: e.target.value }),
    chosen: evS.people.map((n) => ({ ...personOf(n), remove: () => setEv({ people: evS.people.filter((x) => x !== n) }) })),
    matches: candidates.map((n) => { const on = evS.people.indexOf(n) !== -1; return { ...personOf(n), title: titleOf(n), on: on, mark: on ? 'Added' : '+ Add',
      toggle: () => setEv({ people: on ? evS.people.filter((x) => x !== n) : evS.people.concat([n]) }),
      style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;cursor:pointer;color:var(--k-E8EFEA,#E8EFEA);font-family:inherit;background:' + (on ? 'var(--k-22302A,#22302A)' : 'transparent') }; }),
    close: set({ evOpen: false }),
    saveLabel: evNeeds ? 'Schedule & send invites' : 'Add to calendar',
    saveStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (evS.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:var(--k-2A3832,#2A3832);color:var(--k-9DB0A6,#9DB0A6)'),
    save: () => {
      const e = this.state.ev;
      if (!e.title.trim()) { this.notify('Give it a title first.', '#FFB86B'); return; }
      if ((e.type === 'Meeting' || e.type === 'Standup') && e.people.length === 0) { this.notify('Pick at least one participant.', '#FFB86B'); return; }
      const item = { id: 'e' + Date.now(), title: e.title.trim(), type: e.type, date: e.date || todayISO, start: e.start, end: e.end, people: (e.type === 'Meeting' || e.type === 'Standup') ? e.people : [], forum: e.forum, repeat: (e.type === 'Meeting' || e.type === 'Standup') ? e.repeat : 'None', by: 'Samwel Okoth Omulo', notes: (e.type === 'Meeting' || e.type === 'Standup') ? '' : (e.notes || '').trim(), visibility: (e.type === 'Meeting' || e.type === 'Standup') ? '' : e.visibility, remind: e.remind };
      this.setState({ evOpen: false, calSel: item.date, calMonth: item.date.slice(0, 7) });
      calendarW.create(this, this.srv, { ...item, by: this.srv.me.name }).catch((x) => this.notify(x.message || 'Could not save it.', '#FF9B8A'));
      this.notify((e.type === 'Meeting' || e.type === 'Standup') ? e.type + ' scheduled for ' + fmtDay(item.date) + '. Invites sent to ' + e.people.length + (e.people.length === 1 ? ' person.' : ' people.') : 'Added to your calendar for ' + fmtDay(item.date) + '.', '#B8F36B');
    }
  };
  const calUpcoming = [];
  for (let k = 0; k < 14 && calUpcoming.length < 4; k++) { const iso = isoAdd(todayISO, k); itemsOn(iso).forEach((e) => { if (calUpcoming.length < 4) calUpcoming.push({ ...e, iso: iso }); }); }
  const calItems = calUpcoming.map((e) => ({ title: e.type === 'Deadline' ? 'Due: ' + e.title : e.title, color: typeColor[e.type], time: e.start || 'All day', day: e.iso === todayISO ? 'Today' : e.iso === isoAdd(todayISO, 1) ? 'Tomorrow' : fmtDay(e.iso).split(' ').slice(0, 2).join(' '), type: e.type === 'Deadline' ? 'Project deadline' : e.type, peopleLabel: e.people.length ? ' · ' + e.people.length + ' people' : '' }));
  const meetingEvents = s.events.filter((e) => e.type === 'Meeting' || e.type === 'Standup').slice().sort((a, b) => (a.date + a.start).localeCompare(b.date + b.start));
  const meetingsTable = { id: 'meetings', pageSize: 8, minWidth: '960px', emptyText: 'No meetings scheduled yet.',
    cols: [{ label: 'Meeting', filter: 'text' }, { label: 'Type', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Time' }, { label: 'Repeats', filter: 'select' }, { label: 'Participants' }, { label: 'Organizer', filter: 'select' }, { label: '', align: 'right' }],
    rows: meetingEvents.map((e) => ({ cells: [
      C.t(e.title, { strong: true, sub: e.forum ? e.forum + ' forum' : '' }), C.b(e.type, typeColor[e.type]), C.t(fmtDay(e.date), { fv: e.date, nowrap: true }),
      C.t(e.start + '–' + e.end, { nowrap: true, mono: true }), C.t(e.repeat === 'None' ? 'Once' : e.repeat),
      { kind: 'avatars', list: e.people.slice(0, 5).map(personOf), sub: e.people.length + (e.people.length === 1 ? ' person' : ' people'), fv: '' },
      C.t(e.by === ME || e.by === 'Sam' ? 'You' : e.by, { fv: e.by === ME || e.by === 'Sam' ? 'You' : e.by }),
      C.a([{ label: 'Join', kind: 'primary', run: () => joinEvent(e) }, { label: 'Cancel', kind: 'danger', run: () => { calendarW.remove(this, this.srv, e).catch((x) => this.notify(x.message, '#FF9B8A')); this.notify('"' + e.title + '" cancelled. Participants were told.', '#7CC4FF'); } }])
    ] })) };
  // The Home page's "Meetings & standups": the real meetings and standups on this person's calendar for the next 7 days
  // (a repeating one shows once, at its next time), live while it is on.
  const repeatName = { Daily: 'every day', Weekdays: 'on weekdays', Weekly: 'every week' };
  const sessions = [];
  const seen = new Set();
  for (let k = 0; k < 7; k++) {
    const iso = isoAdd(todayISO, k);
    meetingEvents.forEach((e) => {
      if (seen.has(e.id) || !occurs(e, iso) || isClosed(e, iso) || (k === 0 && isPast(e, iso))) return;
      seen.add(e.id);
      const live = k === 0 && !!e.start && e.start <= nowHM && (!e.end || nowHM < e.end);
      const day = k === 0 ? 'Today' : k === 1 ? 'Tomorrow' : fmtDay(iso).split(' ').slice(0, 2).join(' ');
      const mine = isMine(e);
      const going = !!s.rsvp[e.id];
      const pill = (on) => 'flex:none;height:30px;padding:0 12px;border-radius:15px;cursor:pointer;font-size:12px;font-weight:600;' + (on ? 'border:0;background:var(--k-2C3D34,#2C3D34);color:var(--k-B8F36B,#B8F36B)' : 'border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA)');
      sessions.push({ id: e.id, iso, start: e.start || '', title: e.title, type: e.type, color: typeColor[e.type], initials: e.title.split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase(),
        when: (live ? 'Live now' + (e.end ? ' · until ' + e.end : '') : day + ' ' + e.start) + (e.repeat && e.repeat !== 'None' ? ' · ' + (repeatName[e.repeat] || e.repeat.toLowerCase()) : '') + (e.forum ? ' · ' + e.forum : ''),
        live, btn: live ? 'Join now' : mine || !this.srv || !e.container ? '' : going ? 'Going' : 'RSVP',
        btnStyle: live ? 'flex:none;height:30px;padding:0 14px;border-radius:15px;border:0;background:#B8F36B;color:#0F1714;cursor:pointer;font-size:12px;font-weight:700' : pill(going),
        join: live ? () => joinEvent(e) : () => calendarW.toggleRsvp(this, this.srv, e).catch((x) => this.notify(x.message || 'Could not save your answer.', '#FF9B8A')) });
    });
  }
  sessions.sort((a, b) => (a.live === b.live ? (a.iso + a.start).localeCompare(b.iso + b.start) : a.live ? -1 : 1));
  const meetingsSummary = meetingEvents.length + ' meeting' + (meetingEvents.length === 1 ? '' : 's') + ' scheduled · ' + meetingEvents.filter((e) => e.people.length > 0).length + ' with participants';
  return { cal, calItems, ev, joinEvent, meetingsSummary, sessions: sessions.slice(0, 4), meetingsTable, occurs, openEv, titleOf };
}
