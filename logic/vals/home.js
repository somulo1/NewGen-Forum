// Values and handlers the templates read through hubBindings: home.
export function valsHome(ctx) {
  const { canPost, curView, fileFolders, files, forums, ftab, gallery, go, ids, moments, nav, people, posts, s, selectedChatFolder, sessions, set, show, tab, tabTo, tickets, wfiles, widgets } = ctx;
  return {
    secureShare: {
      draft: s.secureShareDraft,
      recipients: people.filter((person) => (!s.secureShareDraft.search || person.name.toLowerCase().includes(s.secureShareDraft.search.toLowerCase()))).slice(0, 6),
      onTitle: (e) => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, title: e.target.value } })),
      onSecret: (e) => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, secret: e.target.value } })),
      onPassword: (e) => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, password: e.target.value } })),
      onSearch: (e) => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, search: e.target.value, recipient: '' } })),
      selectRecipient: (name) => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, recipient: name, search: name } })),
      clearRecipient: () => this.setState((st) => ({ secureShareDraft: { ...st.secureShareDraft, recipient: '', search: '' } })),
      create: async (sendToRecipient) => {
        const draft = this.state.secureShareDraft;
        if (!draft.title.trim() || !draft.secret || !draft.password) {
          this.notify('Enter a title, secret and link password before sharing.', '#FFB86B');
          return;
        }
        if (sendToRecipient && !draft.recipient) {
          this.notify('Choose a person to send this secure share to.', '#FFB86B');
          return;
        }
        const id = crypto.randomUUID();
        const link = new URL('/s/' + id, window.location.origin).toString();
        const share = { id, title: draft.title.trim(), secret: draft.secret, password: draft.password, recipient: sendToRecipient ? draft.recipient : '', createdAt: Date.now(), expiresAt: Date.now() + 24 * 60 * 60 * 1000, link, openedAt: 0, revealUntil: 0, consumed: false };
        if (sendToRecipient) {
          const message = { id: 'secure-message-' + id, from: 'me', text: 'Shared a one-time secret: ' + share.title, secureShareId: id, time: 'Just now' };
          this.setState((st) => {
            const existing = st.dms.find((dm) => dm.with === share.recipient);
            const dms = existing
              ? st.dms.map((dm) => dm.id === existing.id ? { ...dm, unread: dm.unread + 1, msgs: dm.msgs.concat([message]) } : dm)
              : [{ id: 'dm-' + Date.now(), with: share.recipient, unread: 1, msgs: [message] }, ...st.dms];
            return { secureShares: [share, ...st.secureShares], dms, secureShareDraft: { title: '', secret: '', password: '', search: '', recipient: '' }, section: 'home' };
          });
          this.notify('One-time secure share sent to ' + share.recipient + ' in their private chat.', '#B8F36B');
          return;
        }
        try {
          await navigator.clipboard.writeText(link);
          this.setState((st) => ({ secureShares: [share, ...st.secureShares], secureShareDraft: { title: '', secret: '', password: '', search: '', recipient: '' }, section: 'home' }));
          this.notify('Secure-share link copied. The form has been cleared.', '#B8F36B');
        } catch {
          this.notify('The link could not be copied. Check clipboard permissions and try again.', '#FF9B8A');
        }
      }
    },
    widgets, gallery,
    noWidgets: ids.length === 0,
    openGallery: go.widgets,
    closeGallery: go.home,
    wgTab: { all: s.wgFilter === 'all', added: s.wgFilter === 'added', available: s.wgFilter === 'available' },
    wgGo: { all: set({ wgFilter: 'all' }), added: set({ wgFilter: 'added' }), available: set({ wgFilter: 'available' }) },
    wgq: s.wgq, onWgq: (e) => this.setState({ wgq: e.target.value }),
    wgItems: gallery.filter((g) => (s.wgFilter === 'all' || (s.wgFilter === 'added' ? g.added : !g.added)) && (!s.wgq.trim() || (g.title + ' ' + g.desc).toLowerCase().indexOf(s.wgq.trim().toLowerCase()) !== -1)),
    wgCounts: { added: gallery.filter((g) => g.added).length, available: gallery.filter((g) => !g.added).length, all: gallery.length },
    customizing: s.customize,
    toggleCustomize: () => this.setState((st) => ({ customize: !st.customize })),
    customizeLabel: s.customize ? 'Done' : 'Customize',
    customizeBtnStyle: 'height:40px;padding:0 18px;border-radius:20px;font-weight:600;cursor:pointer;' + (s.customize ? 'border:0;background:#E8EFEA;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA'),
    forums, cur: curView, nav, go, show, tab, tabTo, ftab, posts, sessions, tickets, files, fileFolders, selectedChatFolder, people, moments,
    openChatFolder: (folder) => this.setState({ chatFileFolder: folder.id }),
    closeChatFolder: () => this.setState({ chatFileFolder: null }),
    wfiles,
    hasWFiles: s.wfiles.length > 0,
    pickEmojiForum: (ch) => this.setState((st) => ({ fdraft: st.fdraft + ch })),
    pickEmojiWall: (ch) => this.setState((st) => ({ draft: st.draft + ch })),
    pickEmojiTicket: (ch) => this.setState((st) => ({ tcomment: st.tcomment + ch })),
    pickEmojiCall: (ch) => this.setState((st) => ({ callDraft: st.callDraft + ch })),
    chatRef: (el) => { this.chatEl = el; },
    fdraft: s.fdraft,
    onFDraft: (e) => this.setState({ fdraft: e.target.value }),
    onFDraftValue: (value) => this.setState({ fdraft: value }),
    sendForum: () => {
      const t = this.state.fdraft.trim();
      if (!t) return;
      this.postMsg(t, undefined, this.state.replyTo);
      this.setState({ fdraft: '', emojiF: false, replyTo: null });
    },
    replyBar: { active: !!s.replyTo && s.replyTo.key === this._replyKey, who: s.replyTo ? s.replyTo.who : '', snippet: s.replyTo ? s.replyTo.snippet : '', cancel: () => this.setState({ replyTo: null }) },
    fKey: (e) => {
      if (e.key !== 'Enter') return;
      if (e.shiftKey) return;
      e.preventDefault();
      const t = this.state.fdraft.trim();
      if (!t) return;
      this.postMsg(t, undefined, this.state.replyTo);
      this.setState({ fdraft: '', emojiF: false, replyTo: null });
    },
    attachForum: (e) => {
      const list = Array.from(e.target.files || []);
      const ok = this.capFiles(0, list); if (ok.length) this.postMsg('', ok.map((f) => this.fileInfo(f)));
      e.target.value = '';
    },
    onWallFiles: (e) => {
      const list = this.capFiles(this.state.wfiles.length, Array.from(e.target.files || [])).map((f, i) => {
        const isImg = (f.type || '').indexOf('image/') === 0;
        return { id: Date.now() + '-' + i, name: f.name, meta: this.fmtSize(f.size), isImg: isImg, isVideo: (f.type || '').indexOf('video/') === 0, url: URL.createObjectURL(f) };
      });
      e.target.value = '';
      this.setState((st) => ({ wfiles: st.wfiles.concat(list) }));
    },
    draft: s.draft,
    onDraft: (e) => this.setState({ draft: e.target.value }),
    onDraftValue: (value) => this.setState({ draft: value }),
    onPost: (postBg) => {
      const t = this.state.draft.trim();
      const media = this.state.wfiles;
      const wp = this.state.wpoll;
      const opts = wp ? wp.opts.map((o) => o.trim()).filter(Boolean) : [];
      if (wp && (opts.length < 2 || !t)) { this.notify('A poll needs a question and at least two options.', '#FFB86B'); return; }
      if (!t && !media.length) return;
      const pollData = wp ? { opts: opts.map((label, i) => ({ id: 'o' + i, label, votes: 0 })) } : undefined;
      this.setState({ draft: '', wfiles: [], emojiW: false, wpoll: null, posts: [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: this.state.wallAudience + ' · Just now', text: t, postBg: postBg || undefined, likes: 0, extra: 0, shares: 0, comments: [], media: media, pollData }, ...this.state.posts] });
      return true;
    },
    postBtnStyle: 'height:38px;padding:0 22px;border-radius:19px;border:0;font-weight:600;cursor:' + (canPost ? 'pointer' : 'default') + ';background:' + (canPost ? '#B8F36B' : '#2A3832') + ';color:' + (canPost ? '#0F1714' : '#9DB0A6')
  };
}
