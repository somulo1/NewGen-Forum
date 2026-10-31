// When a scheduled meeting or standup is on. Everything here is plain clock arithmetic on the entry's date, start and end (no
// timers, no requests): a screen calls it while it draws, and the app asks for one redraw at the next moment anything changes.
// "live" is the time set for the session (start until end). People can enter from LEAD minutes before the start until the end;
// before that no join button shows and nothing is asked of the server.
export const LEAD = 5;
const ZONE = 'Africa/Nairobi';   // the clock the calendar is kept in (the server uses HUB_SCHEDULE_TZ, same default)

const minutes = (hm) => (hm && /^\d\d:\d\d/.test(hm) ? Number(hm.slice(0, 2)) * 60 + Number(hm.slice(3, 5)) : null);

// "Now" in the calendar's clock: the date, and minutes since midnight with seconds as a fraction.
export function clockNow(at = new Date()) {
  const p = Object.fromEntries(new Intl.DateTimeFormat('en-GB', { timeZone: ZONE, hour12: false, year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit', second: '2-digit' }).formatToParts(at).map((x) => [x.type, x.value]));
  const h = Number(p.hour) % 24;
  return { iso: p.year + '-' + p.month + '-' + p.day, min: h * 60 + Number(p.minute) + Number(p.second) / 60 };
}

// Where a session stands on the day `iso` (its date, for a repeating one the day it is shown for):
// 'upcoming' (more than LEAD minutes away), 'soon' (open to join, not started), 'live' (on), 'ended', or 'later' (another day).
export function stateOf(e, iso, now = clockNow()) {
  if (iso !== now.iso) return iso > now.iso ? 'later' : 'ended';
  const s = minutes(e.start);
  if (s === null) return 'later';
  let en = minutes(e.end);
  if (en === null || en <= s) en = s + 60;
  if (now.min >= en) return 'ended';
  if (now.min >= s) return 'live';
  if (now.min >= s - LEAD) return 'soon';
  return 'upcoming';
}

export const canJoin = (st) => st === 'soon' || st === 'live';

// Milliseconds until the next moment any of these sessions (shown for today) changes state: opens, starts or ends. 0 when none.
export function untilNextChange(list, now = clockNow()) {
  let best = Infinity;
  list.forEach((e) => {
    const s = minutes(e.start);
    if (s === null) return;
    let en = minutes(e.end);
    if (en === null || en <= s) en = s + 60;
    [s - LEAD, s, en].forEach((t) => { if (t > now.min && t - now.min < best) best = t - now.min; });
  });
  return best === Infinity ? 0 : Math.ceil(best * 60000) + 500;
}
