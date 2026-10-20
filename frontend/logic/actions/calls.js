import { startRingback, stopRingback, hangup } from '~/logic/tones.js';
import * as callsW from '~/logic/server/calls.js';

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
    const ns = { call: { mode: mode, title: title, initials: initials, color: color, start: Date.now(), people: people || null, context: context || null, ref: ref || null, dir: dir || 'out' }, callMode: 'docked', personOpen: null, mic: true, callOff: {}, cam: mode === 'video', sharing: false, hand: false, rec: false, callPanel: null, menu: false };
    if ((context === 'dm' || context === 'forum') && this.state.mode !== 'docked' && this.state.mode !== 'full' && this.state.mode !== 'page') ns.mode = 'docked';
    this.setState(ns);
    if (this.srv) {
      this.setState({ callMsgs: [], callDraft: '', sharing: false, hand: false });
      // a real call: the server opens the room (or this.joinCallId joins the one that rang)
      const job = this._joinCallId ? callsW.join(this, this.srv, this._joinCallId) : callsW.start(this, this.srv, { kind: mode, context, ref });
      this._joinCallId = null;
      job.then(() => callsW.applyLocal(this)).catch((e) => { this.notify(e.message || 'Could not start the call.', '#FF9B8A'); this.endCall(); });
      return;
    }
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

  // A moderator can only switch another person's mic or camera OFF. Turning it back on is up to them.
  modOff(name, what) {
    if (this.srv) callsW.moderate(this, name, what);
    this.setState((st) => ({ callOff: { ...st.callOff, [name]: { ...(st.callOff[name] || {}), [what]: true } } }));
    this.notify((what === 'mic' ? 'Muted ' : 'Turned off the video of ') + name + '.', '#FFB86B');
  },

  // Room controls in a big call (host or moderator): mute, mute_all, cam_off, remove, lock, unlock, end, stage_on, stage_off.
  callCtl(action, identity) {
    if (!this.srv) return;
    if (action === 'lock' || action === 'unlock') this.setState({ callLocked: action === 'lock' });
    callsW.roomControl(this, this.srv, action, identity || '');
    if (action === 'mute_all') this.notify('Muted everyone.', '#FFB86B');
  },

  endCall() {
    clearInterval(this._ct);
    const c = this.state.call;
    if (this.srv) { callsW.leaveRoom(); this.setState({ call: null, callMode: 'docked', callOff: {}, sharing: false, hand: false, callFind: '', callLocked: false }); setTimeout(() => callsW.loadHistory(this, this.srv).catch(() => {}), 1500); return; }
    stopRingback(); hangup();
    if (c && (c.context === 'dm' || c.context === 'forum')) this.logCall({ ctx: c.context, ref: c.ref, kind: c.mode === 'video' ? 'video' : 'voice', dir: c.dir || 'out', at: c.start, dur: Math.round((Date.now() - c.start) / 1000) });
    this.setState({ call: null, callMode: 'docked', callOff: {} });
  }
};
