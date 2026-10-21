// HubLogic actions: messaging. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
import * as forumsW from '~/logic/server/widgets/forums.js';
import * as dmsW from '~/logic/server/widgets/dms.js';
import * as mentionsW from '~/logic/server/mentions.js';

export const messagingActions = {
  postMsg(text, file, replyTo) {
    const gate = this._chatGate;
    if (gate && !gate.allowed) { this.notify(gate.reason, '#FFB86B'); return; }
    if (this.state.panelKind === 'dm' && this.state.dmId) this.addDmMsg(this.state.dmId, text, file, replyTo);
    else {
      this.addForumMsg(this.state.forum, text, file, replyTo);
      // people tagged with @ in a forum message are told (bell and email); the text itself stays encrypted
      if (this.srv && text && text.indexOf('@') !== -1) {
        const f = forumsW.liveForum(this, this.state.forum);
        if (f) mentionsW.tell(this.srv, 'forum', f.name, mentionsW.names(this.srv, text).filter((d) => f.memberList.some((m) => m.person === d.uid)));
      }
    }
  },

  // Shows the message at once, then seals and sends it; the stored copy replaces it when the server echoes it back.
  addDmMsg(id, text, file, replyTo) {
    const tmp = crypto.randomUUID();
    this.setState((st) => ({ dms: st.dms.map((d) => (d.id === id ? { ...d, msgs: d.msgs.concat([{ id: tmp, from: 'me', text: text, file: file, replyTo: replyTo || null, time: 'Just now', pending: true }]) } : d)) }));
    if (!this.srv) return;
    const drop = () => this.setState((st) => ({ dms: st.dms.map((d) => ({ ...d, msgs: d.msgs.filter((m) => m.id !== tmp) })) }));
    dmsW.send(this, this.srv, id, { text: text || '', files: [].concat(file || []).map((f) => f.file).filter(Boolean), replyTo: replyTo || null })
      .then((conv) => dmsW.refresh(this, this.srv, conv).then(drop))
      .catch((e) => { drop(); this.notify(e.message || 'Could not send the message.', '#FF9B8A'); });
  },

  openDm(name) {
    if (!name || name === 'Samwel Okoth Omulo' || (this.srv && name === this.srv.me.name)) return;
    const known = this.state.dms.find((d) => d.with === name);
    if (this.srv && known) dmsW.markRead(this, this.srv, known.id);
    this.setState((st) => {
      let conv = st.dms.find((d) => d.with === name);
      const dms = conv ? st.dms.map((d) => (d.id === conv.id ? { ...d, unread: 0 } : d)) : [{ id: 'dm-' + Date.now(), with: name, unread: 0, msgs: [] }].concat(st.dms);
      if (!conv) conv = dms[0];
      return { dms: dms, dmId: conv.id, panelKind: 'dm', panelTab: 'chat', chatFileFolder: null, mode: this.isMobileViewport() ? 'page' : st.mode === 'full' || st.mode === 'page' ? st.mode : 'docked', menu: false, personOpen: null, newDmOpen: false, dmq: '', fdraft: '' };
    });
  },

  // The message shows at once; when the forum is on the server it is sealed and sent, then replaced by the stored copy.
  addForumMsg(fid, text, file, replyTo, shared) {
    const tmp = crypto.randomUUID();
    this.setState((st) => ({ added: { ...st.added, [fid]: [...(st.added[fid] || []), { id: tmp, who: 'Sam', initials: 'S', color: '#F49AC1', time: 'Just now', text: text, file: file, replyTo: replyTo || null, shared: shared || null, replies: 0 }] } }));
    if (!forumsW.isLive(this.srv, fid)) return;
    const dropTmp = () => this.setState((st) => ({ added: { ...st.added, [fid]: (st.added[fid] || []).filter((m) => m.id !== tmp) } }));
    forumsW.send(this, this.srv, fid, { text, files: [].concat(file || []).map((f) => f.file).filter(Boolean), replyTo: replyTo && replyTo.mid, shared })
      .then(dropTmp)
      .catch((e) => { dropTmp(); this.notify(e.message || 'Could not send the message.', '#FF9B8A'); });
  }
};
