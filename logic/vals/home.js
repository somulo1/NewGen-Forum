// Values and handlers the templates read through hubBindings: home.
export function valsHome(ctx) {
  const { canPost, curView, emojisForum, emojisWall, files, forums, ftab, gallery, go, ids, moments, nav, people, posts, s, sessions, set, show, tab, tabTo, tickets, wfiles, widgets } = ctx;
  return {
    widgets, gallery,
    noWidgets: ids.length === 0,
    galleryOpen: s.gallery,
    openGallery: set({ gallery: true }),
    closeGallery: set({ gallery: false }),
    customizing: s.customize,
    toggleCustomize: () => this.setState((st) => ({ customize: !st.customize })),
    customizeLabel: s.customize ? 'Done' : 'Customize',
    customizeBtnStyle: 'height:40px;padding:0 18px;border-radius:20px;font-weight:600;cursor:pointer;' + (s.customize ? 'border:0;background:#E8EFEA;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA'),
    forums, cur: curView, nav, go, show, tab, tabTo, ftab, posts, sessions, tickets, files, people, moments,
    emojisForum, emojisWall, wfiles,
    hasWFiles: s.wfiles.length > 0,
    emojiForumOpen: s.emojiF,
    emojiWallOpen: s.emojiW,
    toggleEmojiForum: () => this.setState((st) => ({ emojiF: !st.emojiF })),
    toggleEmojiWall: () => this.setState((st) => ({ emojiW: !st.emojiW })),
    chatRef: (el) => { this.chatEl = el; },
    fdraft: s.fdraft,
    onFDraft: (e) => this.setState({ fdraft: e.target.value }),
    sendForum: () => {
      const t = this.state.fdraft.trim();
      if (!t) return;
      this.postMsg(t);
      this.setState({ fdraft: '', emojiF: false });
    },
    fKey: (e) => {
      if (e.key !== 'Enter') return;
      const t = this.state.fdraft.trim();
      if (!t) return;
      this.postMsg(t);
      this.setState({ fdraft: '', emojiF: false });
    },
    attachForum: (e) => {
      const list = Array.from(e.target.files || []);
      list.forEach((f) => this.postMsg('', { name: f.name, meta: this.fmtSize(f.size) }));
      e.target.value = '';
    },
    onWallFiles: (e) => {
      const list = Array.from(e.target.files || []).map((f, i) => {
        const isImg = (f.type || '').indexOf('image/') === 0;
        return { id: Date.now() + '-' + i, name: f.name, meta: this.fmtSize(f.size), isImg: isImg, url: isImg ? URL.createObjectURL(f) : '' };
      });
      e.target.value = '';
      this.setState((st) => ({ wfiles: st.wfiles.concat(list) }));
    },
    draft: s.draft,
    onDraft: (e) => this.setState({ draft: e.target.value }),
    onPost: () => {
      const t = this.state.draft.trim();
      const media = this.state.wfiles;
      if (!t && !media.length) return;
      this.setState({ draft: '', wfiles: [], emojiW: false, posts: [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: 'Everyone · Just now', text: t, likes: 0, extra: 0, shares: 0, comments: [], media: media }, ...this.state.posts] });
    },
    postBtnStyle: 'height:38px;padding:0 22px;border-radius:19px;border:0;font-weight:600;cursor:' + (canPost ? 'pointer' : 'default') + ';background:' + (canPost ? '#B8F36B' : '#2A3832') + ';color:' + (canPost ? '#0F1714' : '#9DB0A6')
  };
}
