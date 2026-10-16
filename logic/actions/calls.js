// HubLogic actions: calls. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const callsActions = {
  callBusy() {
    if (!this.state.call) return false;
    this.setState({ toast: 'You are already in a call. Leave "' + this.state.call.title + '" before starting another.', toastColor: '#FFB86B' });
    clearTimeout(this._toastT);
    this._toastT = setTimeout(() => this.setState({ toast: '' }), 3800);
    return true;
  },

  startCall(mode, title, initials, color, people, context, ref) {
    if (this.callBusy()) return;
    clearInterval(this._ct);
    this._ct = setInterval(() => this.forceUpdate(), 1000);
    const ns = { call: { mode: mode, title: title, initials: initials, color: color, start: Date.now(), people: people || null, context: context || null, ref: ref || null }, callMode: 'docked', personOpen: null, mic: true, cam: mode === 'video', sharing: false, hand: false, rec: false, callPanel: null, menu: false };
    if ((context === 'dm' || context === 'forum') && this.state.mode !== 'docked' && this.state.mode !== 'full') ns.mode = 'docked';
    this.setState(ns);
  },

  startDmCall(kind, name, initials, color) {
    if (this.callBusy()) return;
    this.openDm(name);
    this.startCall(kind, 'Private call with ' + name, initials, color, [name], 'dm', name);
  },

  endCall() {
    clearInterval(this._ct);
    this.setState({ call: null, callMode: 'docked' });
  }
};
