// HubLogic actions: messaging. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const messagingActions = {
  postMsg(text, file, replyTo) {
    const gate = this._chatGate;
    if (gate && !gate.allowed) { this.notify(gate.reason, '#FFB86B'); return; }
    if (this.state.panelKind === 'dm' && this.state.dmId) this.addDmMsg(this.state.dmId, text, file, replyTo);
    else this.addForumMsg(this.state.forum, text, file, replyTo);
  },

  addDmMsg(id, text, file, replyTo) {
    this.setState((st) => ({ dms: st.dms.map((d) => (d.id === id ? { ...d, msgs: d.msgs.concat([{ from: 'me', text: text, file: file, replyTo: replyTo || null, time: 'Just now' }]) } : d)) }));
    if (!text) return;
    const replies = ['Got it, thanks!', 'Sure, give me a minute.', 'Okay, I\'ll check and get back to you.', 'Thanks for the update.', 'Can we jump on a quick call?'];
    clearTimeout(this._dmT);
    this._dmT = setTimeout(() => {
      const conv = this.state.dms.find((d) => d.id === id);
      const pr = (conv && this.state.dmPrivacy[conv.with]) || {};
      if (pr.blocked || pr.noChat) return;
      this.setState((st) => ({ dms: st.dms.map((d) => (d.id === id ? { ...d, msgs: d.msgs.concat([{ from: d.with, text: replies[d.msgs.length % replies.length], time: 'Just now' }]) } : d)) }));
    }, 2600);
  },

  openDm(name) {
    if (!name || name === 'Samwel Okoth Omulo') return;
    this.setState((st) => {
      let conv = st.dms.find((d) => d.with === name);
      const dms = conv ? st.dms.map((d) => (d.id === conv.id ? { ...d, unread: 0 } : d)) : [{ id: 'dm-' + Date.now(), with: name, unread: 0, msgs: [] }].concat(st.dms);
      if (!conv) conv = dms[0];
      return { dms: dms, dmId: conv.id, panelKind: 'dm', mode: st.mode === 'full' || st.mode === 'page' ? st.mode : 'docked', menu: false, personOpen: null, newDmOpen: false, dmq: '', fdraft: '' };
    });
  },

  addForumMsg(fid, text, file, replyTo, shared) {
    this.setState((st) => ({ added: { ...st.added, [fid]: [...(st.added[fid] || []), { who: 'Sam', initials: 'S', color: '#F49AC1', time: 'Just now', text: text, file: file, replyTo: replyTo || null, shared: shared || null, replies: 0 }] } }));
  }
};
