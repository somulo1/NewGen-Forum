// Stops a button sending the same thing twice, for every button, with no per-button code.
// A click "arms" its button. Any write the API client starts shortly after belongs to that button: it ignores clicks and,
// if the write is not instant, shows a spinner until the write (and any step chained straight after it) has finished.
// Opt out with data-no-busy on the button (or any ancestor).
const ARM_MS = 1500;         // a write this soon after the click came from it (allows for encrypting first)
const REPEAT_MS = 250;       // a second click this soon is a double-click, even before any write has started
const SETTLE_MS = 300;       // steps chained one after another keep the spinner on between them
const SPINNER_AFTER_MS = 200;   // quick writes (a like) never flash a spinner
const GIVE_UP_MS = 120000;   // a request that never ends must not lock the button for good

const held = new Map();      // button -> { count, spinTimer, settleTimer, giveUpTimer }
let armed = null;            // { button, until }
let last = null;             // { button, at }

const exempt = (button) => !!button.closest('[data-no-busy],[role="toolbar"],[role="listbox"],[role="tablist"]');

function showSpinner(button) {
  if (!held.has(button)) return;
  const css = getComputedStyle(button);
  button.style.setProperty('--hub-busy-color', css.color);
  if (css.position === 'static') button.setAttribute('data-hub-busy-rel', '');
  button.setAttribute('data-hub-busy', 'spin');
}

function release(button) {
  const h = held.get(button);
  if (!h) return;
  clearTimeout(h.spinTimer); clearTimeout(h.settleTimer); clearTimeout(h.giveUpTimer);
  held.delete(button);
  button.removeAttribute('data-hub-busy');
  button.removeAttribute('data-hub-busy-rel');
  button.removeAttribute('aria-busy');
  button.style.removeProperty('--hub-busy-color');
}

function hold(button, promise) {
  let h = held.get(button);
  if (!h) {
    h = { count: 0, spinTimer: setTimeout(() => showSpinner(button), SPINNER_AFTER_MS), settleTimer: null, giveUpTimer: setTimeout(() => release(button), GIVE_UP_MS) };
    held.set(button, h);
    button.setAttribute('data-hub-busy', 'lock');
    button.setAttribute('aria-busy', 'true');
  }
  clearTimeout(h.settleTimer);
  h.count++;
  const done = () => {
    h.count--;
    if (h.count > 0) return;
    if (armed && armed.button === button) armed.until = Date.now() + SETTLE_MS;
    h.settleTimer = setTimeout(() => release(button), SETTLE_MS);
  };
  promise.then(done, done);
}

// The API client passes every write through here.
export function trackWrite(promise) {
  if (armed && Date.now() <= armed.until && armed.button.isConnected) hold(armed.button, promise);
  return promise;
}

export function installSubmitGuard(root = document) {
  root.addEventListener('click', (e) => {
    const button = e.target && e.target.closest ? e.target.closest('button') : null;
    if (!button || button.disabled || exempt(button)) return;
    const now = Date.now();
    if (held.has(button) || (last && last.button === button && now - last.at < REPEAT_MS)) {
      e.preventDefault();
      e.stopImmediatePropagation();
      return;
    }
    last = { button, at: now };
    armed = { button, until: now + ARM_MS };
  }, true);
}
