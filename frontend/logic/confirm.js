// The confirmation card for anything destructive, shown by ConfirmCard in the app layout. Never the browser's own pop-ups.
import { reactive } from 'vue';

export const confirmState = reactive({ open: false, title: '', body: '', action: '', typed: '', cancelLabel: 'Cancel', tone: 'danger', value: '', resolve: null });

// askConfirm({ title, body, action, typed }) answers true only when confirmed. `typed`: text the person has to type first.
export function askConfirm({ title, body = '', action = 'Delete', typed = '', tone = 'danger' }) {
  if (confirmState.resolve) confirmState.resolve(false);   // only one at a time
  return new Promise((resolve) => Object.assign(confirmState, { open: true, title, body, action, typed, tone, value: '', resolve }));
}

export function answerConfirm(ok) {
  const done = confirmState.resolve;
  Object.assign(confirmState, { open: false, resolve: null, value: '' });
  if (done) done(ok);
}

// confirmed(options, fn): a click handler that asks first. A failed question never runs fn.
export const confirmed = (opts, fn) => (...args) => askConfirm(opts).then((ok) => (ok ? fn(...args) : undefined));
