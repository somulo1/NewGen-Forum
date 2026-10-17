// Values and handlers the templates read through hubBindings: home.
export function valsHome(ctx) {
  const { canPost, curView, files, forums, ftab, gallery, go, ids, moments, nav, people, posts, s, sessions, set, show, tab, tabTo, tickets, wfiles, widgets } = ctx;
  return {
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
    forums, cur: curView, nav, go, show, tab, tabTo, ftab, posts, sessions, tickets, files, people, moments,
    wfiles,
    hasWFiles: s.wfiles.length > 0,
    pickEmojiForum: (ch) => this.setState((st) => ({ fdraft: st.fdraft + ch })),
    pickEmojiWall: (ch) => this.setState((st) => ({ draft: st.draft + ch })),
    pickEmojiTicket: (ch) => this.setState((st) => ({ tcomment: st.tcomment + ch })),
    pickEmojiCall: (ch) => this.setState((st) => ({ callDraft: st.callDraft + ch })),
    chatRef: (el) => { this.chatEl = el; },
    fdraft: s.fdraft,
    onFDraft: (e) => this.setState({ fdraft: e.target.value }),
    sendForum: () => {
      const t = this.state.fdraft.trim();
      if (!t) return;
      this.postMsg(t, undefined, this.state.replyTo);
      this.setState({ fdraft: '', emojiF: false, replyTo: null });
    },
    replyBar: { active: !!s.replyTo && s.replyTo.key === this._replyKey, who: s.replyTo ? s.replyTo.who : '', snippet: s.replyTo ? s.replyTo.snippet : '', cancel: () => this.setState({ replyTo: null }) },
    fKey: (e) => {
      if (e.key !== 'Enter') return;
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
