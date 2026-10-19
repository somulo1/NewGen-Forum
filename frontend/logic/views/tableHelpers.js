// renderVals step: tableHelpers. Receives the names built by earlier steps and returns the ones later steps need.
export function buildTableHelpers(ctx) {
  const { ini, nowMs, pColor } = ctx;
  // =============== Shared helpers for tables ===============
  const tz9 = { timeZone: 'Africa/Nairobi' };
  const todayISO = new Date(nowMs).toLocaleDateString('en-CA', tz9);
  const isoAdd = (iso, d) => new Date(new Date(iso + 'T12:00:00+03:00').getTime() + d * 86400000).toLocaleDateString('en-CA', tz9);
  const dowOf = (iso) => new Date(iso + 'T12:00:00+03:00').getUTCDay();
  const fmtDay = (iso) => new Date(iso + 'T12:00:00+03:00').toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'Africa/Nairobi' });
  const personOf = (n) => ({ name: n, initials: n === 'Samwel Okoth Omulo' ? 'S' : ini(n), color: pColor[n] || ((this.dir.find((d) => d.name === n) || {}).color) || '#C9B6FF' });
  const C = {
    t: (text, o) => Object.assign({ text: String(text == null ? '' : text) }, o || {}),
    b: (text, color, o) => Object.assign({ kind: 'badge', text: text, color: color }, o || {}),
    p: (name, o) => { const pp = personOf(name); return Object.assign({ kind: 'person', text: name, initials: pp.initials, color: pp.color }, o || {}); },
    a: (buttons) => ({ kind: 'actions', buttons: buttons, fv: '' }),
    bar: (pct, color, text) => ({ kind: 'bar', pct: pct, color: color, text: text || (pct + '%'), fv: String(pct) }),
    chk: (on, toggle, o) => Object.assign({ kind: 'check', on: on, toggle: toggle, fv: on ? 'Yes' : 'No' }, o || {}),
    sel: (value, options, onChange, o) => Object.assign({ kind: 'select', value: value, options: options, onChange: onChange, fv: value }, o || {})
  };
  const corePeople = ['Brian Kiprop', 'David Otieno', 'Amina Wanjiru', 'Grace Njeri', 'Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'John Kamau', 'Lilian Chebet'];
  return { C, corePeople, dowOf, fmtDay, isoAdd, personOf, todayISO, tz9 };
}
