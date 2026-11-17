// A slim strip right under the header of the phone app: offline, what is waiting to be sent, and a message the server refused.
// It is part of the page (it moves the content down a line instead of covering it), follows the light and dark themes, and is
// gone when all is well.
let el = null, keep = 0, noteTimer = 0, note = '', status = '';

function node() {
  if (el || typeof document === 'undefined') return el;
  el = document.createElement('div');
  el.setAttribute('role', 'status');
  el.setAttribute('aria-live', 'polite');
  el.style.cssText = 'flex:none;padding:5px 16px;text-align:center;font:500 12px/1.35 "DM Sans",system-ui,sans-serif;' +
    'background:var(--k-121B17,#121B17);border-bottom:1px solid var(--k-24322C,#24322C);display:none';
  return el;
}

// under the header; a page without one (signing in) shows it at the top
function place() {
  const h = document.querySelector('.hub-header');
  if (h && h.nextElementSibling !== el) h.after(el);
  else if (!h && el.parentNode !== document.body && document.body) document.body.prepend(el);
}

function paint() {
  const n = node();
  if (!n) return;
  const text = note || status;
  n.textContent = text;
  n.style.display = text ? 'block' : 'none';
  n.style.color = note ? '#FFB86B' : 'var(--k-9DB0A6,#9DB0A6)';
  place();
  // the layout may draw its header again (another page); the strip follows it while it has something to say
  clearInterval(keep);
  if (text) keep = setInterval(place, 1500);
}

// offline: no connection; waiting: messages and changes not yet on the server
export function setStatus(offline, waiting) {
  const w = waiting ? waiting + (waiting === 1 ? ' change' : ' changes') : '';
  status = offline ? 'No connection · showing saved data' + (w ? ' · ' + w + ' waiting to sync' : '')
    : w ? 'Syncing ' + w + '…' : '';
  paint();
}

export function flash(text, ms = 8000) {
  note = text;
  clearTimeout(noteTimer);
  noteTimer = setTimeout(() => { note = ''; paint(); }, ms);
  paint();
}
