// renderVals step: tickets. Receives the names built by earlier steps and returns the ones later steps need.
export function buildTickets(ctx) {
  const { has, prio, s } = ctx;
  // ---- Tickets & TAT ----
  const ts = s.tset;
  const toMin = (v) => { const p = String(v || '0:0').split(':'); return (Number(p[0]) || 0) * 60 + (Number(p[1]) || 0); };
  const dayH = Math.max(1, (toMin(ts.end) - toMin(ts.start)) / 60);
  const pColor = { 'Brian Kiprop': '#7CC4FF', 'David Otieno': '#F49AC1', 'Amina Wanjiru': '#6EE7C8', 'Grace Njeri': '#FFB86B', 'Faith Achieng': '#B8F36B', 'Kevin Mutua': '#C9B6FF', 'Mercy Wambui': '#FFB86B', 'Sam': '#F49AC1', 'John Kamau': '#7CC4FF', 'Lilian Chebet': '#F49AC1' };
  const ini = (n) => (n === 'Sam' ? 'S' : n.split(' ').map((x) => x[0]).join('').slice(0, 2));
  const fmtH = (h) => {
    let m = Math.round(Math.abs(h) * 60);
    const dm = Math.round(dayH * 60);
    const d = Math.floor(m / dm); m -= d * dm;
    const hh = Math.floor(m / 60), mm = m % 60;
    const parts = [];
    if (d) parts.push(d + 'd');
    if (hh) parts.push(hh + 'h');
    if (mm && !d) parts.push(mm + 'm');
    return parts.join(' ') || '0m';
  };
  const catOf = (name) => ts.cats.find((c) => c.name === name) || { name: name, tat: 1, unit: 'days' };
  const tatHours = (name) => { const c = catOf(name); return c.unit === 'days' ? c.tat * dayH : c.tat; };
  const tatText = (name) => { const c = catOf(name); return c.tat + (c.unit === 'days' ? (c.tat === 1 ? ' working day' : ' working days') : (c.tat === 1 ? ' hour' : ' hours')); };
  const stepNames = ['Raised', 'Assigned', 'In progress', 'Resolved', 'Closed'];
  const badge = (c) => 'display:inline-block;padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;white-space:nowrap;background:' + c + '22;color:' + c + ';border:1px solid ' + c + '66';
  const calc = (t) => {
    const tat = tatHours(t.cat);
    const paused = t.step === 2 && t.hold && ts.pause.onhold;
    const rem = tat - t.elapsed;
    const pct = Math.min(100, Math.round((t.elapsed / tat) * 100));
    let sla, color, remainingLabel;
    if (t.step >= 3) { sla = t.late ? 'Solved late' : 'Solved on time'; color = t.late ? '#FFB86B' : '#B8F36B'; remainingLabel = 'Solved in ' + fmtH(t.elapsed); }
    else if (paused) { sla = 'TAT paused'; color = '#7CC4FF'; remainingLabel = fmtH(Math.max(0, rem)) + ' left'; }
    else if (rem < 0) { sla = 'Overdue'; color = '#FF9B8A'; remainingLabel = fmtH(-rem) + ' overdue'; }
    else if (pct >= ts.nearPct) { sla = 'Due soon'; color = '#FFB86B'; remainingLabel = fmtH(rem) + ' left'; }
    else { sla = 'On time'; color = '#B8F36B'; remainingLabel = fmtH(rem) + ' left'; }
    return { tat, rem, pct, paused, sla, color, remainingLabel };
  };
  const tickets = s.tickets.map((t) => {
    const c = calc(t);
    return { ...t, prioStyle: prio(t.prio), statusLabel: t.step === 2 && t.hold ? 'On hold' : stepNames[t.step], slaLabel: c.sla, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, handlerLabel: t.handler || 'Unassigned',
      open: () => this.setState({ ticketOpen: t.id, tagOpen: false, tcomment: '' }), _c: c };
  });
  const nOpen = tickets.filter((t) => t.step < 3).length;
  const nOver = tickets.filter((t) => t._c.sla === 'Overdue').length;
  const nSoon = tickets.filter((t) => t._c.sla === 'Due soon').length;
  const nDone = tickets.filter((t) => t.step >= 3).length;
  const ticketStats = [
    { label: 'Open', n: nOpen, color: '#E8EFEA' },
    { label: 'Overdue', n: nOver, color: '#FF9B8A' },
    { label: 'Due soon', n: nSoon, color: '#FFB86B' },
    { label: 'Solved', n: nDone, color: '#B8F36B' }
  ];
  const cur_t = s.tickets.find((t) => t.id === s.ticketOpen);
  let tk = {};
  if (cur_t) {
    const c = calc(cur_t);
    const id = cur_t.id;
    const primary = 'height:42px;border-radius:10px;border:0;background:#B8F36B;color:#0F1714;font-weight:600;cursor:pointer;font-size:14px';
    const secondary = 'height:42px;border-radius:10px;border:1px solid #2A3832;background:transparent;color:#E8EFEA;cursor:pointer;font-size:14px';
    const setStep = (step, text, extra) => this.updTicket(id, (t) => Object.assign({}, t, { step: step, hold: false, activity: this.act(t, { text: text }) }, extra || {}));
    const actions = [];
    if (cur_t.step === 1) actions.push({ label: 'Start work', style: primary, run: () => setStep(2, 'Started work') });
    if (cur_t.step === 2) {
      actions.push({ label: 'Mark as solved', style: primary, run: () => {
        const late = cur_t.elapsed > c.tat;
        setStep(3, 'Marked it solved', { late: late });
        if (ts.notify.solved) this.notify(id + ' solved' + (late ? ' (late)' : ' on time') + '. ' + cur_t.raisedBy + ' has been notified.', '#B8F36B', id);
      } });
      actions.push({ label: cur_t.hold ? 'Resume work' : 'Put on hold', style: secondary, run: () => this.updTicket(id, (t) => Object.assign({}, t, { hold: !t.hold, activity: this.act(t, { text: t.hold ? 'Resumed work' : 'Put the ticket on hold' }) })) });
    }
    if (cur_t.step === 3) {
      actions.push({ label: 'Close ticket', style: primary, run: () => setStep(4, 'Closed the ticket') });
      actions.push({ label: 'Reopen', style: secondary, run: () => setStep(2, 'Reopened the ticket') });
    }
    if (cur_t.step === 4) actions.push({ label: 'Reopen', style: secondary, run: () => setStep(2, 'Reopened the ticket') });
    const pauses = [ts.pause.weekends && 'weekends', ts.pause.holidays && 'public holidays', ts.pause.offhours && ('outside ' + ts.start + '–' + ts.end), ts.pause.onhold && 'on-hold time', ts.pause.awaiting && 'waiting for the requester'].filter(Boolean);
    tk = {
      ...cur_t,
      prioStyle: prio(cur_t.prio),
      slaLabel: c.sla, slaColor: c.color, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, paused: c.paused,
      slaBarStyle: 'height:100%;width:' + Math.max(4, c.pct) + '%;background:' + c.color + ';border-radius:4px',
      tatLabel: tatText(cur_t.cat), usedLabel: fmtH(cur_t.elapsed),
      pauseNote: pauses.length ? 'Not counted: ' + pauses.join(', ') + '.' : 'TAT counts all the time.',
      hasHandler: !!cur_t.handler, noHandler: !cur_t.handler,
      handlerColor: pColor[cur_t.handler] || '#9DB0A6', handlerInitials: cur_t.handler ? ini(cur_t.handler) : '',
      handlerValue: cur_t.handler || '',
      handoverLabel: cur_t.handler ? 'Hand over to' : 'Assign to',
      wasEscalated: !!cur_t.escalatedTo, escalatedToLabel: cur_t.escalatedTo || '',
      assignees: cur_t.assignees.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })),
      actions: actions, noActions: actions.length === 0,
      steps: stepNames.map((label, i) => {
        const done = i < cur_t.step, now = i === cur_t.step;
        const notes = [cur_t.created, cur_t.handler ? 'To ' + cur_t.handler : 'Waiting for someone', cur_t.step === 2 && cur_t.hold ? 'On hold' : (cur_t.step >= 2 ? 'Work started' : ''), cur_t.step >= 3 ? (cur_t.late ? 'After the TAT' : 'Within the TAT') : 'TAT ' + tatText(cur_t.cat), cur_t.step === 4 ? 'Done' : ''];
        return {
          label: label, num: done ? '✓' : String(i + 1), note: notes[i],
          barStyle: 'display:block;height:4px;border-radius:2px;background:' + (done || now ? '#B8F36B' : '#2A3832'),
          dotStyle: 'width:22px;height:22px;flex:none;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;' + (done ? 'background:#B8F36B;color:#0F1714' : now ? 'border:2px solid #B8F36B;color:#B8F36B;box-sizing:border-box' : 'border:1px solid #3F524A;color:#9DB0A6;box-sizing:border-box'),
          labelStyle: 'font-weight:' + (now ? '700' : '500') + ';color:' + (done || now ? '#E8EFEA' : '#9DB0A6')
        };
      }),
      activity: cur_t.activity.map((a) => ({ ...a, initials: ini(a.who), color: pColor[a.who] || '#9DB0A6', isEvent: a.kind === 'event', isComment: a.kind === 'comment', isFile: a.kind === 'file', meta: a.meta || '' }))
    };
  }
  const tagNames = ['Brian Kiprop', 'David Otieno', 'Amina Wanjiru', 'Grace Njeri', 'Faith Achieng', 'Kevin Mutua'];
  const tagPeople = tagNames.map((n) => ({ name: n, tag: () => this.setState((st) => ({ tcomment: (st.tcomment && !/\s$/.test(st.tcomment) ? st.tcomment + ' ' : st.tcomment) + '@' + n + ' ', tagOpen: false })) }));
  const autoNotifs = [];
  if (has.tickets) tickets.forEach((t) => {
    if (t._c.sla === 'Overdue' && ts.notify.overdue) autoNotifs.push({ text: t.id + ' is overdue by ' + fmtH(-t._c.rem) + '. Escalated to ' + ts.escalate + '.', color: '#FF9B8A', time: 'Now', ticket: t.id });
    if (t._c.sla === 'Due soon' && ts.notify.nearing) autoNotifs.push({ text: t.id + ' is due soon: ' + fmtH(t._c.rem) + ' left on its TAT.', color: '#FFB86B', time: 'Now', ticket: t.id });
  });
  const notifList = autoNotifs.concat(s.notifs).map((n) => ({ ...n, open: () => this.setState({ notifOpen: false, section: n.ticket ? 'tickets' : this.state.section, ticketOpen: n.ticket || this.state.ticketOpen }) }));
  const chip = (on) => 'height:36px;padding:0 14px;border-radius:18px;cursor:pointer;font-size:13px;' + (on ? 'border:1px solid #B8F36B;background:#2C3D34;color:#E8EFEA;font-weight:600' : 'border:1px solid #2A3832;background:#121B17;color:#C9D6CF');
  const ntCats = ts.cats.map((c) => ({ label: c.name + ' · ' + tatText(c.name), pick: () => this.setState((st) => ({ nt: { ...st.nt, cat: c.name } })), style: chip(s.nt.cat === c.name) }));
  const ntPrios = ['High', 'Medium', 'Low'].map((p) => ({ label: p, pick: () => this.setState((st) => ({ nt: { ...st.nt, prio: p } })), style: chip(s.nt.prio === p) }));
  const setTs = (patch) => this.setState((st) => ({ tset: Object.assign({}, st.tset, typeof patch === 'function' ? patch(st.tset) : patch) }));
  const tsetCats = ts.cats.map((c, i) => ({
    ...c,
    setTat: (e) => setTs((t) => ({ cats: t.cats.map((x, j) => (j === i ? { ...x, tat: Math.max(1, Number(e.target.value) || 1) } : x)) })),
    setUnit: (e) => setTs((t) => ({ cats: t.cats.map((x, j) => (j === i ? { ...x, unit: e.target.value } : x)) })),
    remove: () => setTs((t) => ({ cats: t.cats.filter((x, j) => j !== i) }))
  }));
  const pauseRules = [
    { key: 'weekends', label: 'Weekends', desc: 'Saturday and Sunday do not count.' },
    { key: 'holidays', label: 'Public holidays', desc: 'Days on the holiday calendar do not count.' },
    { key: 'offhours', label: 'Outside working hours', desc: 'Only time between the start and end of the working day counts.' },
    { key: 'onhold', label: 'While a ticket is on hold', desc: 'Putting a ticket on hold stops its clock.' },
    { key: 'awaiting', label: 'While waiting for the requester', desc: 'The clock stops until the person who raised it replies.' }
  ].map((r) => ({ ...r, checked: !!ts.pause[r.key], toggle: () => setTs((t) => ({ pause: { ...t.pause, [r.key]: !t.pause[r.key] } })) }));
  const notifyRules = [
    { key: 'solved', label: 'When a ticket is solved', desc: 'Tell the person who raised it.' },
    { key: 'overdue', label: 'When a ticket goes overdue', desc: 'Alert the handler and escalate.' },
    { key: 'nearing', label: 'Before the TAT runs out', desc: 'Warn the handler when time is nearly up.' },
    { key: 'assigned', label: 'When a ticket is assigned or handed over', desc: 'Tell the new handler.' }
  ].map((r) => ({ ...r, checked: !!ts.notify[r.key], toggle: () => setTs((t) => ({ notify: { ...t.notify, [r.key]: !t.notify[r.key] } })) }));
  const files = [
    { kind: 'PDF', color: '#FF9B8A', name: 'Leave policy 2026', meta: 'HR · v3 · 2 days ago' },
    { kind: 'XLS', color: '#B8F36B', name: 'Q3 budget tracker', meta: 'Finance · v12 · today' },
    { kind: 'DOC', color: '#7CC4FF', name: 'Onboarding guide', meta: 'HR · v5 · last week' },
    { kind: 'PNG', color: '#FFB86B', name: 'Network diagram v3', meta: 'IT Operations · 3 h ago' },
    { kind: 'PPT', color: '#F49AC1', name: 'Town hall slides', meta: 'Management · yesterday' },
    { kind: 'ZIP', color: '#6EE7C8', name: 'Brand assets', meta: 'Marketing · last month' }
  ];
  const people = [
    { name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', role: 'Finance Analyst', forumRole: 'Admin' },
    { name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', role: 'Systems Engineer', forumRole: 'Moderator' },
    { name: 'Grace Njeri', initials: 'GN', color: '#FFB86B', role: 'HR Officer', forumRole: 'Member' },
    { name: 'David Otieno', initials: 'DO', color: '#F49AC1', role: 'Developer', forumRole: 'Member' },
    { name: 'Faith Achieng', initials: 'FA', color: '#B8F36B', role: 'Customer Care', forumRole: 'Member' },
    { name: 'Kevin Mutua', initials: 'KM', color: '#C9B6FF', role: 'Sales Lead', forumRole: 'Guest' }
  ];
  return { badge, chip, cur_t, dayH, files, ini, nOpen, nOver, nSoon, notifList, notifyRules, ntCats, ntPrios, pColor, pauseRules, people, setTs, tagNames, tagPeople, tatText, ticketStats, tickets, tk, ts, tsetCats };
}
