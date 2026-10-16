// renderVals step: wall. Receives the names built by earlier steps and returns the ones later steps need.
export function buildWall(ctx) {
  const { can, flowById, forumsData, makeRequest, openPersonName, s } = ctx;
  const pollOpts = [
    { id: 'a', label: 'FIFA 26', base: 18 },
    { id: 'b', label: 'Call of Duty', base: 11 },
    { id: 'c', label: 'Mario Kart', base: 7 }
  ];
  const total = pollOpts.reduce((n, o) => n + o.base, 0) + (s.vote ? 1 : 0);
  const poll = pollOpts.map((o) => {
    const v = o.base + (s.vote === o.id ? 1 : 0);
    const pct = Math.round((v / total) * 100);
    const mine = s.vote === o.id;
    return {
      label: o.label + (mine ? '  ✓' : ''),
      pct: s.vote ? pct + '%' : '',
      vote: () => this.setState({ vote: o.id }),
      style: 'display:flex;align-items:center;min-height:44px;padding:0 14px;border-radius:10px;cursor:pointer;font-size:14px;color:#E8EFEA;text-align:left;border:1px solid ' + (mine ? '#B8F36B' : '#2A3832') + ';background:' + (s.vote ? 'linear-gradient(90deg,#2C3D34 ' + pct + '%,#121B17 ' + pct + '%)' : '#121B17')
    };
  });
  const posts = s.posts.map((p) => {
    const liked = !!s.liked[p.id];
    return {
      ...p,
      hasPoll: !!p.hasPoll,
      hasPhoto: !!p.photoBg,
      poll: poll,
      pollNote: s.vote ? total + ' votes · thanks for voting' : 'Tap an option to vote · closes Friday',
      likes: p.likes + (liked ? 1 : 0),
      commentList: p.comments,
      images: (p.media || []).filter((m) => m.isImg),
      docs: (p.media || []).filter((m) => !m.isImg),
      hasImages: (p.media || []).some((m) => m.isImg),
      hasDocs: (p.media || []).some((m) => !m.isImg),
      shares: p.shares + (s.shareCount[p.id] || 0),
      shareOpen: s.shareFor === p.id,
      shareNote: s.shareNote[p.id] || '',
      hasShareNote: !!s.shareNote[p.id],
      hasPoster: !!p.poster, posterAward: p.poster ? p.poster.award : '', posterNames: p.poster ? p.poster.names : '', posterWhen: p.poster ? p.poster.when : '',
      hasTags: !!(p.tags && p.tags.length), tagList: (p.tags || []).map((n) => ({ name: n, open: () => openPersonName(n) })),
      previewImg: ((p.media || []).find((m) => m.isImg) || {}).url || '',
      hasPreviewImg: (p.media || []).some((m) => m.isImg),
      hasPreviewBg: !!p.photoBg && !(p.media || []).some((m) => m.isImg),
      previewText: (p.text || 'Shared a file').length > 110 ? p.text.slice(0, 110) + '…' : (p.text || 'Shared a file'),
      shareOutBlocked: !can('wall', 'Share outside Hub'),
      shareOut: can('wall', 'Share outside Hub') ? ['WhatsApp', 'Email', 'LinkedIn', 'X', 'Facebook'].map((app) => ({ label: app, run: () => {
        const f = flowById('f4');
        if (f && f.levels > 0) {
          makeRequest('f4', 'Share ' + p.author + '\'s post to ' + app, { kind: 'share', postId: p.id, app: app });
          this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Waiting for approval before sharing to ' + app + '.' } }));
        } else {
          this.setState((st) => ({ shareFor: null, shareCount: { ...st.shareCount, [p.id]: (st.shareCount[p.id] || 0) + 1 }, shareNote: { ...st.shareNote, [p.id]: 'Shared to ' + app + ' with the image and text.' } }));
        }
      } })) : [],
      toggleShare: () => this.setState((st) => ({ shareFor: st.shareFor === p.id ? null : p.id })),
      copyLink: () => this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Link copied. Paste it anywhere in Hub.' } })),
      shareTargets: forumsData.map((f) => ({ name: f.name, initials: f.initials, color: f.color, pick: () => {
        const snippet = p.text.length > 60 ? p.text.slice(0, 60) + '…' : p.text;
        this.addForumMsg(f.id, 'Shared a wall post from ' + p.author + ': "' + snippet + '"');
        this.setState((st) => ({ shareFor: null, shareCount: { ...st.shareCount, [p.id]: (st.shareCount[p.id] || 0) + 1 }, shareNote: { ...st.shareNote, [p.id]: 'Shared to ' + f.name } }));
      } })),
      commentCount: p.extra + p.comments.length,
      heartFill: liked ? 'currentColor' : 'none',
      like: () => this.setState({ liked: { ...this.state.liked, [p.id]: !liked } }),
      likeStyle: 'height:40px;border-radius:10px;border:0;cursor:pointer;font-size:14px;display:flex;align-items:center;justify-content:center;gap:8px;background:transparent;font-weight:' + (liked ? '600' : '400') + ';color:' + (liked ? '#FF9B8A' : '#C9D6CF'),
      onComment: (e) => {
        const v = (e.target.value || '').trim();
        if (e.key !== 'Enter' || !v) return;
        e.target.value = '';
        this.setState({ posts: this.state.posts.map((q) => q.id === p.id ? { ...q, comments: [...q.comments, { who: 'Sam', initials: 'S', color: '#F49AC1', text: v }] } : q) });
      }
    };
  });
  const moments = [
    { initials: 'BK', color: '#7CC4FF', bg: '#2B4A63', caption: 'Brian at the server room' },
    { initials: 'GN', color: '#FFB86B', bg: '#6B4A2B', caption: 'Team lunch' },
    { initials: 'DO', color: '#F49AC1', bg: '#5A2E47', caption: 'Game night setup' },
    { initials: 'FA', color: '#B8F36B', bg: '#3B5A2B', caption: 'First day!' }
  ];
  const canPost = s.draft.trim().length > 0 || s.wfiles.length > 0;
  const pw = s.pw;
  const panelStyle = 'position:relative;flex:0 0 ' + pw + 'px;width:' + pw + 'px;max-width:60%;min-height:0;box-sizing:border-box;margin:12px;border-radius:22px;border:1px solid #2A3832;background:#131D19;box-shadow:0 12px 36px rgba(0,0,0,0.35);display:flex;flex-direction:column;overflow:hidden';
  const sessions = [
    { title: 'IT daily standup', forum: 'IT Operations', initials: 'IT', color: '#B8F36B', when: 'Weekdays 9:15', live: false, btn: 'Remind me' },
    { title: 'Sprint review', forum: 'Product Builders', initials: 'PB', color: '#7CC4FF', when: 'Now', live: true, btn: 'Join now' },
    { title: 'Choir rehearsal (online)', forum: 'Choir & Worship', initials: 'CW', color: '#F49AC1', when: 'Thu 5:30pm', live: false, btn: 'RSVP' },
    { title: 'IT town hall', forum: 'IT Operations', initials: 'IT', color: '#B8F36B', when: 'Fri 10:00', live: false, btn: 'RSVP' }
  ].map((x) => ({ ...x, join: x.live ? () => this.startCall('video', x.title, x.initials, x.color) : () => this.notify('RSVP sent for ' + x.title + ' (' + x.when + ').', '#7CC4FF'), btnStyle: 'height:40px;padding:0 18px;border-radius:20px;cursor:pointer;font-weight:600;' + (x.live ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA') }));
  const prio = (p) => {
    const c = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' }[p];
    return 'padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;border:1px solid ' + c + ';color:' + c;
  };
  return { canPost, moments, panelStyle, posts, prio, pw, sessions };
}
