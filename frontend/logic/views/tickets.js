import * as ticketsW from '~/logic/server/widgets/tickets.js';
import { recipientOf, recipientLabel } from '~/logic/ticketRouting.js';
import { richTextHtml, mentionLinker } from '~/logic/richText.js';

// renderVals step: tickets. Receives the names built by earlier steps and returns the ones later steps need.
export function buildTickets(ctx) {
  const { can, has, prio, s } = ctx;
  const withMentions = mentionLinker(this.dir.map((d) => d.name).concat(['Sam']));
  // ---- Tickets & TAT ----
  const ts = s.tset;
  const toMin = (v) => { const p = String(v || '0:0').split(':'); return (Number(p[0]) || 0) * 60 + (Number(p[1]) || 0); };
  const dayH = Math.max(1, (toMin(ts.end) - toMin(ts.start)) / 60);
  const pColor = new Proxy({ Sam: '#F49AC1' }, { get: (o, k) => (k in o ? o[k] : ((this.dir.find((d) => d.name === k) || {}).color)) });
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
    const tat = t.tatH || tatHours(t.cat);
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
    const to = recipientOf(t, this.org, has);
    return { ...t, to, toLabel: recipientLabel(to), isMine: t.raisedBy === 'Sam', prioStyle: prio(t.prio), statusLabel: t.step === 2 && t.hold ? 'On hold' : stepNames[t.step], slaLabel: c.sla, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, handlerLabel: t.handler || 'Unassigned',
      open: () => this.setState({ ticketOpen: t.id, ticketFrom: null, tagOpen: false, tcomment: '' }), _c: c };
  });
  const nOpen = tickets.filter((t) => t.step < 3).length;
  const nOver = tickets.filter((t) => t._c.sla === 'Overdue').length;
  const nSoon = tickets.filter((t) => t._c.sla === 'Due soon').length;
  const nDone = tickets.filter((t) => t.step >= 3).length;
  const statsOf = (list) => [
    { label: 'Open', n: list.filter((t) => t.step < 3).length, color: '#E8EFEA' },
    { label: 'Overdue', n: list.filter((t) => t._c.sla === 'Overdue').length, color: '#FF9B8A' },
    { label: 'Due soon', n: list.filter((t) => t._c.sla === 'Due soon').length, color: '#FFB86B' },
    { label: 'Solved', n: list.filter((t) => t.step >= 3).length, color: '#B8F36B' }
  ];
  const ticketStats = statsOf(tickets);
  const myTickets = tickets.filter((t) => t.isMine);
  const myTicketStats = statsOf(myTickets);
  const deskTickets = tickets.filter((t) => t.to.type === 'desk');
  const deskStats = statsOf(deskTickets);
  const cur_t = s.tickets.find((t) => t.id === s.ticketOpen);
  let tk = {};
  if (cur_t) {
    const c = calc(cur_t);
    const id = cur_t.id;
    const primary = 'height:42px;border-radius:10px;border:0;background:#B8F36B;color:#0F1714;font-weight:600;cursor:pointer;font-size:14px';
    const secondary = 'height:42px;border-radius:10px;border:1px solid var(--k-2A3832,#2A3832);background:transparent;color:var(--k-E8EFEA,#E8EFEA);cursor:pointer;font-size:14px';
    const fail = (e) => this.notify(e.message || 'That did not work.', '#FF9B8A');
    const setStep = (step, done, extra) => ticketsW.setStep(this, this.srv, cur_t, step, extra).then(done).catch(fail);
    // resolving and closing open a small form first: a comment (required to resolve) and stars (closing only)
    const fin = s.tkFinish && s.tkFinish.id === id && s.tkFinish.step === cur_t.step + 1 ? s.tkFinish : null;
    const openFinish = (step) => this.setState({ tkFinish: { id, step, note: '', stars: 0, err: '', busy: false } });
    const actions = [];
    // nobody works on a ticket they raised: the person who raised it can only close it once it is solved
    const raisedByMe = cur_t.raisedBy === 'Sam';
    // the person on it, the head of the place it is with (the server says so), or the Solve right; never who raised it
    const mayWork = !!cur_t.canWork || (!raisedByMe && (cur_t.handler === 'Sam' || !this.srv || can('tickets', 'Solve')));   // a head also works on what they raised to their own place
    // three actions only. Resolve: from the moment it reaches the place, until it is resolved. Close: once resolved (Hub closes it
    // itself 30 minutes after that). Escalate: forward it to another place, open, resolved or even closed; it opens there on that
    // place's own TAT.
    if (cur_t.step < 3 && mayWork) actions.push({ label: 'Resolve', style: primary, run: () => openFinish(3) });
    if (cur_t.step === 3) actions.push({ label: 'Close', style: primary, run: () => openFinish(4) });
    if (this.srv && cur_t.canEscalate !== false && cur_t.canEscalate) actions.push({ label: 'Escalate', style: secondary, run: () => this.setState({ escOpen: true, esc: { type: '', id: '', cat: '', person: '', why: '' } }) });
    let finish = null;
    if (fin) {
      const patch = (p) => this.setState((st) => ({ tkFinish: st.tkFinish ? { ...st.tkFinish, ...p } : st.tkFinish }));
      const solving = fin.step === 3;
      finish = {
        solving, title: solving ? 'Resolve ticket' : 'Close ticket',
        hint: solving ? 'Say what was done to resolve it. ' + cur_t.raisedBy + ' sees this comment.' : 'Every closure is rated. Pick 1 to 5 stars for how it was handled and say why. (If nobody closes it, Hub closes it 30 minutes after it is resolved with 5 stars.)',
        placeholder: solving ? 'What was done to resolve it (required)' : 'Comment on closing it (required)',
        note: fin.note, err: fin.err, busy: fin.busy, stars: fin.stars,
        starList: [1, 2, 3, 4, 5].map((n) => ({ n, on: n <= fin.stars, label: n + (n === 1 ? ' star' : ' stars'), pick: () => patch({ stars: fin.stars === n ? 0 : n }) })),
        onNote: (e) => patch({ note: e.target.value, err: '' }),
        cancel: () => this.setState({ tkFinish: null }),
        submit: () => {
          const note = (this.state.tkFinish && this.state.tkFinish.note || '').trim();
          // every action carries a comment
          if (!solving && !fin.stars) return patch({ err: 'Rate it: pick 1 to 5 stars.' });
          if (!note) return patch({ err: solving ? 'Write what was done to resolve it.' : 'Write a comment on closing it.' });
          if (note.length > 2000) return patch({ err: 'Keep the comment under 2000 characters.' });
          if (fin.busy) return;
          patch({ busy: true });
          const late = cur_t.elapsed > c.tat;
          const extra = solving ? { comment: note } : { comment: note, rating: fin.stars || 0 };
          ticketsW.setStep(this, this.srv, cur_t, fin.step, extra).then(() => {
            this.setState({ tkFinish: null });
            if (solving && ts.notify.solved) this.notify(id + ' solved' + (late ? ' (late)' : ' on time') + '. ' + cur_t.raisedBy + ' has been notified.', '#B8F36B', id);
          }).catch((e) => patch({ busy: false, err: e.message || 'That did not work.' }));
        }
      };
    }
    const pauses = [ts.pause.weekends && 'weekends', ts.pause.holidays && 'public holidays', ts.pause.offhours && ('outside ' + ts.start + '–' + ts.end), ts.pause.onhold && 'on-hold time', ts.pause.awaiting && 'waiting for the requester'].filter(Boolean);
    tk = {
      ...cur_t,
      descHtml: richTextHtml(cur_t.desc),
      prioStyle: prio(cur_t.prio),
      slaLabel: c.sla, slaColor: c.color, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, paused: c.paused,
      slaBarStyle: 'height:100%;width:' + Math.max(4, c.pct) + '%;background:' + c.color + ';border-radius:4px',
      tatLabel: tatText(cur_t.cat), usedLabel: fmtH(cur_t.elapsed),
      pauseNote: pauses.length ? 'Not counted: ' + pauses.join(', ') + '.' : 'TAT counts all the time.',
      hasHandler: !!cur_t.handler, noHandler: !cur_t.handler,
      handlerColor: pColor[cur_t.handler] || '#9DB0A6', handlerInitials: cur_t.handler ? ini(cur_t.handler) : '',
      handlerValue: cur_t.handler || '',
      handoverLabel: cur_t.handler ? 'Hand over to' : 'Assign to',
      // handing over or reassigning needs the right; the person who raised it is never offered as the one to work on it
      // only the head of the department the ticket is with assigns or reassigns (the server decides; the person working on it does not)
      canHandover: cur_t.step < 4 && (!this.srv || cur_t.canAssign === true),
      // the people this place's category names to solve it: suggested first. A category's people only take tickets raised after they
      // were added, so an older ticket waits for someone to assign it; one click does that here.
      suggested: (() => {
        if (!this.srv || cur_t.handler || cur_t.step >= 3) return [];
        const TB = { departments: 'department', units: 'unit', branches: 'branch', sectors: 'sector', desk: 'desk' };
        const to = cur_t.to || { type: 'desk' };
        const ref = to.type === 'desk' ? null : (this.srv.structure[to.type] && this.srv.structure[to.type].get(to.name));
        const cats = this.srv.ticketCats || [];
        const c = cats.find((x) => x.name === cur_t.cat && x.unit_type === TB[to.type] && (x.unit_ref || null) === (ref || null)) || cats.find((x) => x.name === cur_t.cat && x.unit_type === 'desk');
        return ((c && c.members) || []).map((id) => (this.dir || []).find((p) => p.uid === id)).filter((p) => p && p.name !== cur_t.raisedBy)
          .map((p) => ({ name: p.name, assign: () => ticketsW.assign(this, this.srv, cur_t, p.name).then(() => this.notify(cur_t.id + ' assigned to ' + p.name + '. They have been notified.', '#7CC4FF', cur_t.id)).catch((x) => this.notify(x.message || 'Could not assign it.', '#FF9B8A')) }));
      })(),
      // assigning: search anyone by name (there can be many people), or take it yourself in one click
      // the members of the place it is with: the head hands it over to one of them
      placeMembers: (() => {
        const to = cur_t.to || { type: 'desk' };
        const field = { departments: 'dept', branches: 'loc', sectors: 'sector', units: 'unit' }[to.type];
        const ex = new Set([cur_t.raisedBy, cur_t.handler === 'Sam' && this.srv ? this.srv.me.name : cur_t.handler]);
        return (this.dir || []).filter((p) => (p.status || 'active') === 'active' && (!field || p[field] === to.name) && !ex.has(p.name))
          .map((p) => ({ value: p.name, label: (this.srv && p.name === this.srv.me.name ? 'Me (' + p.name + ')' : p.name) + (p.title ? ' · ' + p.title : '') })).sort((a, b) => a.label.localeCompare(b.label));
      })(),
      onMember: (e) => { const v = e.target.value; e.target.value = ''; if (v) ticketsW.assign(this, this.srv, cur_t, this.srv && v === this.srv.me.name ? 'Sam' : v).then(() => this.notify(cur_t.id + (cur_t.handler ? ' handed over to ' : ' assigned to ') + v + '. They have been notified.', '#7CC4FF', cur_t.id)).catch((x) => this.notify(x.message || 'Could not hand it over.', '#FF9B8A')); },
      pickPeople: (this.dir || []).filter((p) => (p.status || 'active') === 'active').map((p) => ({ uid: p.uid || p.name, name: p.name, title: p.title, dept: p.dept, initials: p.initials, color: p.color })),
      pickExclude: (this.dir || []).filter((p) => p.name === cur_t.raisedBy || (raisedByMe && this.srv && p.name === this.srv.me.name) || (cur_t.handler && (p.name === cur_t.handler || (cur_t.handler === 'Sam' && this.srv && p.name === this.srv.me.name)))).map((p) => p.uid || p.name),
      canAssignMe: !!this.srv && !raisedByMe && cur_t.handler !== 'Sam' && cur_t.step < 3,
      assignTo: (name) => ticketsW.assign(this, this.srv, cur_t, this.srv && name === this.srv.me.name ? 'Sam' : name).then(() => this.notify(cur_t.id + (cur_t.handler ? ' handed over to ' : ' assigned to ') + (this.srv && name === this.srv.me.name ? 'you' : name) + '.', '#7CC4FF', cur_t.id)).catch((x) => this.notify(x.message || 'Could not assign it.', '#FF9B8A')),
      assignMe: () => ticketsW.assign(this, this.srv, cur_t, 'Sam').then(() => this.notify(cur_t.id + ' is yours now.', '#7CC4FF', cur_t.id)).catch((x) => this.notify(x.message || 'Could not assign it to you.', '#FF9B8A')),
      handoverOpts: (this.dir || []).filter((p) => p.name !== cur_t.raisedBy && !(raisedByMe && p.name === (this.srv && this.srv.me.name)) && (p.status || 'active') === 'active')
        .map((p) => ({ value: this.srv && p.name === this.srv.me.name ? 'Sam' : p.name, label: (this.srv && p.name === this.srv.me.name ? 'Me (' + p.name + ')' : p.name) + (p.dept ? ' · ' + p.dept : '') }))
        .sort((a, b) => a.label.localeCompare(b.label)),
      ownNote: raisedByMe && !cur_t.canWork && cur_t.step < 4 ? 'You raised this ticket, so someone else works on it. You can close it once it is solved.' : '',
      // business requests and communications: the receiving place acknowledges them
      isAckable: cur_t.kind === 'request' || cur_t.kind === 'communication', kindLabel: cur_t.kind === 'request' ? 'business request' : cur_t.kind === 'communication' ? 'communication' : 'ticket',
      acks: cur_t.acks || [], hasAcks: (cur_t.acks || []).length > 0,
      canAck: !!cur_t.canAck && !(cur_t.acks || []).some((a) => this.srv && a.name === this.srv.me.name),
      acknowledge: () => ticketsW.acknowledge(this, this.srv, cur_t).then(() => this.notify('Acknowledged.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not acknowledge it.', '#FF9B8A')),
      wasEscalated: !!cur_t.escalatedTo, escalatedToLabel: cur_t.escalatedTo || '',
      assignees: cur_t.assignees.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })),
      finish, actions: finish ? [] : actions,
      ratingStars: cur_t.rating ? '★★★★★'.slice(0, cur_t.rating) + '☆☆☆☆☆'.slice(cur_t.rating) : '',
      closedNote: cur_t.step === 4 ? (cur_t.autoClosed ? 'Closed automatically 30 minutes after it was resolved.' : 'Closed' + (cur_t.closedAt ? ' ' + cur_t.closedAt : '') + '.') : '',
      closed: cur_t.step === 4, noActions: !finish && actions.length === 0 && cur_t.step !== 4,
      steps: stepNames.map((label, i) => {
        const done = i < cur_t.step, now = i === cur_t.step;
        const notes = [cur_t.created, cur_t.handler ? 'To ' + cur_t.handler : 'Waiting for someone', cur_t.step === 2 && cur_t.hold ? 'On hold' : (cur_t.step >= 2 ? 'Work started' : ''), cur_t.step >= 3 ? (cur_t.late ? 'After the TAT' : 'Within the TAT') : 'TAT ' + tatText(cur_t.cat), cur_t.step === 4 ? 'Done' : ''];
        return {
          label: label, num: done ? '✓' : String(i + 1), note: notes[i],
          barStyle: 'display:block;height:4px;border-radius:2px;background:' + (done || now ? '#B8F36B' : 'var(--k-2A3832,#2A3832)'),
          dotStyle: 'width:22px;height:22px;flex:none;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;' + (done ? 'background:#B8F36B;color:#0F1714' : now ? 'border:2px solid var(--k-B8F36B,#B8F36B);color:var(--k-B8F36B,#B8F36B);box-sizing:border-box' : 'border:1px solid var(--k-3F524A,#3F524A);color:var(--k-9DB0A6,#9DB0A6);box-sizing:border-box'),
          labelStyle: 'font-weight:' + (now ? '700' : '500') + ';color:' + (done || now ? '#E8EFEA' : '#9DB0A6')
        };
      }),
      activity: cur_t.activity.map((a) => ({ ...a, htmlText: withMentions(richTextHtml(a.text)), initials: ini(a.who), color: pColor[a.who] || '#9DB0A6', isEvent: a.kind === 'event', isComment: a.kind === 'comment', isFile: a.kind === 'file', isImg: !!a.isImg, url: a.url || '', meta: a.meta || '' }))
    };
  }
  // tagging on a ticket is only for the people who own it: whoever raised it, whoever is assigned to resolve it, and the heads of
  // the dashboard it went to and the one it came from. They can tag each other, and nobody else can tag.
  const tagCircle = cur_t ? [...new Set([cur_t.raisedBy, cur_t.handler].concat(cur_t.assignees || [], cur_t.toHeads || [], cur_t.fromHeads || []))].filter(Boolean) : [];
  const myName = this.srv ? this.srv.me.name : 'Sam';
  const canTag = !!cur_t && (tagCircle.includes('Sam') || tagCircle.includes(myName));
  const ticketPeople = tagCircle.filter((n) => n !== 'Sam' && n !== myName);
  const tagNames = canTag ? ticketPeople : [];
  const tagPeople = tagNames.map((n) => ({ name: n, tag: () => this.setState((st) => ({ tcomment: (st.tcomment && !/\s$/.test(st.tcomment) ? st.tcomment + ' ' : st.tcomment) + '@' + n + ' ', tagOpen: false })) }));
  const autoNotifs = [];
  // only for tickets this person raised or is working on now: someone it was handed away from hears nothing more (unless tagged)
  const meNow = this.srv ? this.srv.me.name : 'Sam';
  if (has.tickets) tickets.filter((t) => !this.srv || t.raisedBy === 'Sam' || t.handler === meNow || t.handler === 'Sam').forEach((t) => {
    if (t._c.sla === 'Overdue' && ts.notify.overdue) autoNotifs.push({ text: t.id + ' is overdue by ' + fmtH(-t._c.rem) + '. Escalated to ' + ts.escalate + '.', color: '#FF9B8A', time: 'Now', ticket: t.id });
    if (t._c.sla === 'Due soon' && ts.notify.nearing) autoNotifs.push({ text: t.id + ' is due soon: ' + fmtH(t._c.rem) + ' left on its TAT.', color: '#FFB86B', time: 'Now', ticket: t.id });
  });
  const notifList = autoNotifs.concat(s.notifs).map((n) => ({ ...n, open: n.url ? () => { if (typeof window !== 'undefined') { const w = window.open(n.url, '_blank'); if (w) w.opener = null; } this.setState({ notifOpen: false }); } : n.kind === 'comm' && n.ref && (this.state.comms || []).some((c) => c.id === n.ref) ? () => {
    // a reminder or notice about a memo or announcement opens it (and opening it counts as reading it)
    const c = this.state.comms.find((x) => x.id === n.ref);
    this.setState({ notifOpen: false, section: 'comms', commPage: '', commOpen: c.uid, commView: 'read' });
    if (this.srv && c.status === 'Published' && !c.iRead) this.srv.request('POST', '/comms/' + c.uid + '/read').then(() => this.setState((st) => ({ comms: st.comms.map((x) => (x.uid === c.uid ? { ...x, iRead: true } : x)) }))).catch(() => {});
  } : () => this.setState({ ticketFrom: null, notifOpen: false, section: n.ticket ? 'tickets' : this.state.section, ticketOpen: n.ticket || this.state.ticketOpen }) }));
  const chip = (on) => 'height:36px;padding:0 14px;border-radius:18px;cursor:pointer;font-size:13px;' + (on ? 'border:1px solid var(--k-B8F36B,#B8F36B);background:var(--k-2C3D34,#2C3D34);color:var(--k-E8EFEA,#E8EFEA);font-weight:600' : 'border:1px solid var(--k-2A3832,#2A3832);background:var(--k-121B17,#121B17);color:var(--k-C9D6CF,#C9D6CF)');
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
  const people = [];
  const defaultCats = ts.cats.map((c) => ({ name: c.name, tat: tatHours(c.name) }));
  return { defaultCats, deskStats, deskTickets, myTicketStats, myTickets, badge, chip, cur_t, dayH, ini, nOpen, nOver, nSoon, notifList, notifyRules, ntCats, ntPrios, pColor, pauseRules, people, setTs, tagNames, ticketPeople, canTag, tagPeople, tatText, ticketStats, tickets, tk, ts, tsetCats };
}
