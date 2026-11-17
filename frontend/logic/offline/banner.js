// A small notice at the top of the phone app: offline, what is waiting to be sent, and a message the server refused.
// It never takes a tap (the page under it keeps working) and hides itself when there is nothing to say.
let el = null, noteTimer = 0, note = '', status = '';

function node() {
  if (el || typeof document === 'undefined' || !document.body) return el;
  el = document.createElement('div');
  el.setAttribute('role', 'status');
  el.setAttribute('aria-live', 'polite');
  el.style.cssText = 'position:fixed;top:8px;left:50%;transform:translateX(-50%);z-index:2147483646;pointer-events:none;' +
    'max-width:calc(100% - 32px);padding:6px 12px;border-radius:999px;font:500 12px/1.3 "DM Sans",system-ui,sans-serif;' +
    'background:rgba(15,23,20,.92);color:#E8EFEA;border:1px solid #22302A;box-shadow:0 4px 14px rgba(0,0,0,.25);' +
    'text-align:center;transition:opacity .2s;opacity:0';
  document.body.appendChild(el);
  return el;
}

function paint() {
  const n = node();
  if (!n) return;
  const text = note || status;
  n.textContent = text;
  n.style.opacity = text ? '1' : '0';
  n.style.color = note ? '#FFB86B' : '#E8EFEA';
}

// offline: no connection; waiting: messages not yet sent
export function setStatus(offline, waiting) {
  const w = waiting ? waiting + (waiting === 1 ? ' message' : ' messages') : '';
  status = offline ? 'No connection · showing saved data' + (w ? ' · ' + w + ' waiting to send' : '')
    : w ? 'Sending ' + w + '…' : '';
  paint();
}

export function flash(text, ms = 8000) {
  note = text;
  clearTimeout(noteTimer);
  noteTimer = setTimeout(() => { note = ''; paint(); }, ms);
  paint();
}
