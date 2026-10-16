// renderVals step: projectStandup. A project standup is a calendar event for the whole team.
// Before it starts the project shows the time; from the start time until it ends the button becomes "Join standup".
export function buildProjectStandup(ctx) {
  const { cur_p, fmtDay, isoAdd, joinEvent, occurs, pj, s, todayISO } = ctx;
  if (!cur_p) return {};
  const nowHM = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi' });
  const mine = s.events.filter((e) => e.type === 'Standup' && e.pid === cur_p.id);
  let next = null;
  for (let k = 0; k < 31 && !next; k++) {
    const iso = isoAdd(todayISO, k);
    next = mine.filter((e) => occurs(e, iso) && !(k === 0 && e.end <= nowHM)).sort((a, b) => a.start.localeCompare(b.start)).map((e) => ({ e, iso }))[0] || null;
  }
  const live = !!next && next.iso === todayISO && next.e.start <= nowHM;
  const addMinutes = (hm, n) => { const t = Number(hm.slice(0, 2)) * 60 + Number(hm.slice(3, 5)) + n; return String(Math.floor(t / 60) % 24).padStart(2, '0') + ':' + String(t % 60).padStart(2, '0'); };
  const f = s.standupForm;
  const setF = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ standupForm: { ...st.standupForm, [k]: v, err: '' } })); };
  pj.standup = {
    state: live ? 'live' : next ? 'upcoming' : 'none', isLive: live, isUpcoming: !!next && !live, isNone: !next,
    label: next ? (next.iso === todayISO ? 'Today' : fmtDay(next.iso)) + ' · ' + next.e.start + (next.e.repeat && next.e.repeat !== 'None' ? ' · repeats ' + next.e.repeat.toLowerCase() : '') : '',
    join: () => next && joinEvent(next.e),
    cancel: () => { if (!next) return; this.setState((x) => ({ events: x.events.filter((y) => y.id !== next.e.id) })); this.notify('Standup cancelled. The team was told.', '#7CC4FF'); },
    openForm: () => this.setState({ standupForm: { date: todayISO, start: addMinutes(nowHM, 5), repeat: 'None', err: '' } }),
    formOpen: !!f, form: f ? { date: f.date, start: f.start, repeat: f.repeat, err: f.err, hasErr: !!f.err, onDate: setF('date'), onStart: setF('start'), onRepeat: setF('repeat'), min: todayISO,
      close: () => this.setState({ standupForm: null }),
      submit: () => {
        if (!f.date || !f.start) { this.setState({ standupForm: { ...f, err: 'Pick the date and the start time.' } }); return; }
        if (f.date < todayISO || (f.date === todayISO && f.start <= nowHM)) { this.setState({ standupForm: { ...f, err: 'Pick a time that is still ahead.' } }); return; }
        const item = { id: 'sp' + Date.now(), title: cur_p.name + ' standup', type: 'Standup', date: f.date, start: f.start, end: addMinutes(f.start, 15), people: cur_p.team.slice(), forum: '', repeat: f.repeat, by: 'Samwel Okoth Omulo', notes: '', visibility: 'Project team', remind: 'None', project: cur_p.name, pid: cur_p.id };
        this.setState((x) => ({ standupForm: null, events: x.events.concat([item]) }));
        this.notify('Standup scheduled for ' + (f.date === todayISO ? 'today' : fmtDay(f.date)) + ' at ' + f.start + '. The button turns into "Join standup" when it starts.', '#B8F36B');
      } } : {} };
  return {};
}
