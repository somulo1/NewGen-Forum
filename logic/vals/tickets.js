import { pickHandler } from '~/logic/ticketRouting.js';

// Values and handlers the templates read through hubBindings: tickets.
export function valsTickets(ctx) {
  const { catalog, cur_t, dayH, esc, notifyRules, ntCats, ntCatSel, ntPrios, ntTargets, pauseRules, ticketDeskOk, s, set, setTs, sid, sw, tagNames, tagPeople, tatText, ticketStats, tickets, tk, ts, tsetCats } = ctx;
  return {
    sw, settingsOpen: !!(sid && catalog[sid]),
    closeSettings: set({ settingsFor: null }),
    saveSettings: () => { const wc = this.state.wcfg[sid] || {}; if (wc.see === 'Specific groups' && !(wc.groups || []).length) { this.notify('Pick at least one group, or choose Everyone.', '#FFB86B'); return; } this.setState({ settingsFor: null }); this.notify('Settings saved for ' + ((s.wcfg[sid] && s.wcfg[sid].title) || catalog[sid].title) + '.', '#B8F36B'); },
    tset: ts, tsetCats, pauseRules, notifyRules,
    dayHoursLabel: (Math.round(dayH * 10) / 10) + ' hours',
    setNewCat: (e) => setTs({ newCat: e.target.value }),
    setNewTat: (e) => setTs({ newTat: e.target.value }),
    setNewUnit: (e) => setTs({ newUnit: e.target.value }),
    addCat: () => setTs((t) => (t.newCat.trim() ? { cats: t.cats.concat([{ name: t.newCat.trim(), tat: Math.max(1, Number(t.newTat) || 1), unit: t.newUnit }]), newCat: '', newTat: '' } : {})),
    setStart: (e) => setTs({ start: e.target.value }),
    setEnd: (e) => setTs({ end: e.target.value }),
    setNearPct: (e) => setTs({ nearPct: Number(e.target.value) }),
    setEscalate: (e) => setTs({ escalate: e.target.value }),
    tickets, ticketStats, tk, tagPeople,
    showTicketList: s.section === 'tickets' && !cur_t,
    showDeskList: s.section === 'ticketdesk' && ticketDeskOk && !cur_t,
    ticketDeskOk,
    showTicketView: (s.section === 'tickets' || s.section === 'ticketdesk') && !!cur_t,
    closeTicket: () => { const from = this.state.ticketFrom; if (from) this.setState({ ticketOpen: null, ticketFrom: null, section: 'orgunit', ou: from.ou, ouTab: from.tab, ouItem: null }); else this.setState({ ticketOpen: null }); },
    ticketBackLabel: s.ticketFrom ? '‹ Back to ' + s.ticketFrom.label : 'All tickets',
    esc, tcomment: s.tcomment, tagOpen: s.tagOpen,
    onTComment: (e) => this.setState({ tcomment: e.target.value }),
    onTCommentValue: (value) => this.setState({ tcomment: value }),
    toggleTag: () => this.setState((st) => ({ tagOpen: !st.tagOpen })),
    postTComment: () => {
      const text = this.state.tcomment.trim();
      if (!text || !cur_t || cur_t.step === 4) return;
      const tagged = tagNames.filter((n) => text.indexOf('@' + n) !== -1);
      this.updTicket(cur_t.id, (t) => Object.assign({}, t, { activity: this.act(t, { kind: 'comment', text: text }) }));
      this.setState({ tcomment: '', tagOpen: false });
      if (tagged.length) this.notify('You tagged ' + tagged.join(', ') + ' on ' + cur_t.id + '. They have been notified.', '#C9B6FF', cur_t.id);
    },
    attachTicket: (e) => {
      const list = this.capFiles(0, Array.from(e.target.files || []));
      e.target.value = '';
      if (!cur_t || !list.length) return;
      this.updTicket(cur_t.id, (t) => Object.assign({}, t, { activity: list.reduce((acc, f) => acc.concat([{ who: 'Sam', time: 'Just now', kind: 'file', ...(i => ({ text: i.name, meta: i.meta, url: i.url, isImg: i.isImg }))(this.fileInfo(f)) }]), t.activity) }));
    },
    handover: (e) => {
      const name = e.target.value;
      if (!name || !cur_t || name === cur_t.handler) return;
      const verb = cur_t.handler ? 'Handed it over to ' : 'Assigned it to ';
      this.updTicket(cur_t.id, (t) => Object.assign({}, t, { handler: name, handlerSince: 'Just now', assignees: t.assignees.indexOf(name) === -1 ? t.assignees.concat([name]) : t.assignees, step: Math.max(1, t.step), activity: this.act(t, { text: verb + name }) }));
      if (ts.notify.assigned) this.notify(cur_t.id + (cur_t.handler ? ' handed over to ' : ' assigned to ') + name + '. They have been notified.', '#7CC4FF', cur_t.id);
    },
    ntOpen: s.ntOpen, nt: s.nt, ntCats, ntPrios,
    toggleNewTicket: () => this.setState((st) => ({ ntOpen: !st.ntOpen })),
    ntSetTitle: (e) => { const v = e.target.value; this.setState((st) => ({ nt: { ...st.nt, title: v } })); },
    ntAttach: (e) => { const list = this.capFiles((this.state.nt.files || []).length, Array.from(e.target.files || [])).map((f) => this.fileInfo(f)); e.target.value = ''; this.setState((st) => ({ nt: { ...st.nt, files: (st.nt.files || []).concat(list) } })); },
    ntRemoveFile: (i) => this.setState((st) => ({ nt: { ...st.nt, files: (st.nt.files || []).filter((_, k) => k !== i) } })),
    ntSetDesc: (e) => { const v = typeof e === 'string' ? e : e.target.value; this.setState((st) => ({ nt: { ...st.nt, desc: v } })); },
    ntBtnStyle: 'height:40px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (s.nt.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:var(--k-2A3832,#2A3832);color:var(--k-9DB0A6,#9DB0A6)'),
    createTicket: () => {
      const nt = this.state.nt;
      const cat = ntCatSel;
      if (!nt.title.trim()) return;
      const handler = pickHandler(cat, this.state.tickets);
      const pre = /Software/.test(cat.name) ? 'DEV-' : /HR/.test(cat.name) ? 'HR-' : /Network|Hardware|Access/.test(cat.name) ? 'IT-' : 'TK-';
      const id = pre + (300 + this.state.tickets.length);
      const t = { id: id, title: nt.title.trim(), cat: cat.name, prio: cat.custom ? cat.prio : nt.prio, step: 0, hold: false, raisedBy: 'Sam', assignees: handler ? [handler] : [], handler: handler, handlerSince: handler ? 'Just now' : '', created: 'Just now', elapsed: 0, sync: 'Not synced', dept: ntTargets.name || 'IT Operations', to: ntTargets.type === 'desk' ? { type: 'desk', name: 'Ticket desk' } : { type: ntTargets.type, name: ntTargets.name }, fromDept: 'Risk & Compliance', createdAt: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), escalated: false, tatH: cat.custom ? cat.tat : 0, desc: nt.desc.trim() || 'No details given.', activity: [{ who: 'Sam', kind: 'event', text: 'Raised the ticket', time: 'Just now' }].concat((nt.files || []).map((f) => ({ who: 'Sam', kind: 'file', time: 'Just now', text: f.name, meta: f.meta, url: f.url, isImg: f.isImg }))) };
      this.setState((st) => ({ tickets: [t].concat(st.tickets), ntOpen: false, nt: { title: '', cat: st.nt.cat, prio: 'Medium', desc: '', files: [], toType: st.nt.toType, toName: st.nt.toName }, ticketOpen: id }));
      this.notify(id + ' raised in ' + cat.name + ' and sent to ' + (ntTargets.type === 'desk' ? 'the ticket desk' : ntTargets.name) + '. TAT ' + (cat.custom ? cat.tat + ' hours' : tatText(cat.name)) + (handler ? '. Assigned to ' + handler : '') + '.', '#7CC4FF', id);
    }
  };
}
