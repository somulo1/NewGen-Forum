import { startRingback, stopRingback, hangup } from '~/logic/tones.js';

// HubLogic actions: calls. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const callsActions = {
  callBusy() {
    if (!this.state.call) return false;
    this.setState({ toast: 'You are already in a call. Leave "' + this.state.call.title + '" before starting another.', toastColor: '#FFB86B' });
    clearTimeout(this._toastT);
    this._toastT = setTimeout(() => this.setState({ toast: '' }), 3800);
    return true;
  },

  startCall(mode, title, initials, color, people, context, ref, dir) {
    if (this.callBusy()) return;
    clearInterval(this._ct);
    this._ct = setInterval(() => this.forceUpdate(), 1000);
    const ns = { call: { mode: mode, title: title, initials: initials, color: color, start: Date.now(), people: people || null, context: context || null, ref: ref || null, dir: dir || 'out' }, callMode: 'docked', personOpen: null, mic: true, cam: mode === 'video', sharing: false, hand: false, rec: false, callPanel: null, menu: false };
    if ((context === 'dm' || context === 'forum') && this.state.mode !== 'docked' && this.state.mode !== 'full' && this.state.mode !== 'page') ns.mode = 'docked';
    this.setState(ns);
    // The caller hears ringing until the other side answers; the person who accepted a call does not.
    if ((context === 'dm' || context === 'forum') && dir !== 'in') startRingback(3200);
  },

  startDmCall(kind, name, initials, color, dir) {
    if (this.callBusy()) return;
    if ((this.state.dmPrivacy[name] || {}).blocked) { this.logCall({ ctx: 'dm', ref: name, kind: kind === 'video' ? 'video' : 'voice', dir: 'blocked' }); this.notify('You blocked ' + name + '. Unblock them to call.', '#FFB86B'); return; }
    this.openDm(name);
    this.startCall(kind, 'Private call with ' + name, initials, color, [name], 'dm', name, dir);
  },

  // Adds a line to the call history shown in each chat's Calls tab.
  logCall(entry) {
    this.setState((st) => ({ callLog: [{ id: 'cl' + Date.now() + Math.random().toString(36).slice(2, 5), at: Date.now(), dur: 0, ...entry }].concat(st.callLog) }));
  },

  endCall() {
    clearInterval(this._ct);
    const c = this.state.call;
    stopRingback(); hangup();
    if (c && (c.context === 'dm' || c.context === 'forum')) this.logCall({ ctx: c.context, ref: c.ref, kind: c.mode === 'video' ? 'video' : 'voice', dir: c.dir || 'out', at: c.start, dur: Math.round((Date.now() - c.start) / 1000) });
    this.setState({ call: null, callMode: 'docked' });
  }
};
