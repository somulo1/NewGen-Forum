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
  const TAGS = { CharityRun: ['charity'], NewJoiners: ['new team members', 'joiner', 'welcome'], TournamentNight: ['tournament'] };
  const tagHit = (p) => !s.wallTag || (TAGS[s.wallTag] || [s.wallTag.toLowerCase()]).some((k) => (p.text || '').toLowerCase().indexOf(k) !== -1);
  const countAll = (list) => list.reduce((n, c) => n + 1 + countAll(c.replies || []), 0);
  const newComment = (text) => ({ who: 'Sam', initials: 'S', color: '#F49AC1', text, time: 'Just now', replies: [] });
  // Adds a reply under the comment at `path` (indexes from the top-level list down).
  const addAt = (list, path, text) => list.map((c, i) => (i !== path[0] ? c : path.length === 1 ? { ...c, replies: (c.replies || []).concat([newComment(text)]) } : { ...c, replies: addAt(c.replies || [], path.slice(1), text) }));
  const buildComments = (p, list, path) => list.map((c, i) => {
    const here = path.concat([i]);
    const key = p.id + ':' + here.join('.');
    return { key, who: c.who, initials: c.initials, color: c.color, text: c.text, time: c.time || '', depth: path.length, replying: s.replyFor === key,
      startReply: () => this.setState({ replyFor: key }), cancel: () => this.setState({ replyFor: null }),
      send: (text) => this.setState((st) => ({ replyFor: null, posts: st.posts.map((q) => (q.id === p.id ? { ...q, comments: addAt(q.comments || [], here, text) } : q)) })),
      replies: buildComments(p, c.replies || [], here) };
  });
  const posts = s.posts.filter((p) => !s.hiddenPosts[p.id] && tagHit(p)).map((p) => {
    const liked = !!s.liked[p.id];
    const mineP = p.author === 'Sam';
    const saved = !!s.savedPosts[p.id];
    const closeMenu = { postMenu: null };
    const userPoll = p.pollData ? (() => {
      const my = s.pollVotes[p.id];
      const tot = p.pollData.opts.reduce((n, o) => n + o.votes, 0) + (my ? 1 : 0);
      return { tot, list: p.pollData.opts.map((o) => { const v = o.votes + (my === o.id ? 1 : 0); return { label: o.label + (my === o.id ? '  ✓' : ''), pct: my ? Math.round((v / tot) * 100) + '%' : '', vote: () => this.setState((st) => ({ pollVotes: { ...st.pollVotes, [p.id]: o.id } })),
        style: 'display:flex;align-items:center;min-height:44px;padding:0 14px;border-radius:10px;cursor:pointer;font-size:14px;color:#E8EFEA;text-align:left;border:1px solid ' + (my === o.id ? '#B8F36B' : '#2A3832') + ';background:#121B17' }; }) };
    })() : null;
    return {
      ...p,
      menuOpen: s.postMenu === p.id,
      toggleMenu: () => this.setState((st) => ({ postMenu: st.postMenu === p.id ? null : p.id })),
      menuItems: [
        { label: saved ? 'Unsave post' : 'Save post', run: () => this.setState((st) => ({ ...closeMenu, savedPosts: { ...st.savedPosts, [p.id]: !saved } })) },
        { label: 'Copy link', run: () => this.setState((st) => ({ ...closeMenu, shareNote: { ...st.shareNote, [p.id]: 'Link copied. Paste it anywhere in Hub.' } })) },
        { label: 'Hide this post', run: () => { this.setState((st) => ({ ...closeMenu, hiddenPosts: { ...st.hiddenPosts, [p.id]: true } })); this.notify('Post hidden. It will not show on your wall.', '#9DB0A6'); } },
        mineP ? { label: 'Delete post', danger: true, run: () => { this.setState((st) => ({ ...closeMenu, posts: st.posts.filter((q) => q.id !== p.id) })); this.notify('Your post was deleted.', '#FF9B8A'); } }
          : { label: 'Report post', danger: true, run: () => { this.setState(closeMenu); this.notify('Thanks. The post was reported to the moderators.', '#FFB86B'); } }
      ],
      saved, savedLabel: saved ? 'Saved' : '',
      hasPoll: !!p.hasPoll || !!userPoll,
      hasPhoto: !!p.photoBg,
      poll: userPoll ? userPoll.list : poll,
      pollNote: userPoll ? (s.pollVotes[p.id] ? userPoll.tot + ' votes · thanks for voting' : 'Tap an option to vote') : s.vote ? total + ' votes · thanks for voting' : 'Tap an option to vote · closes Friday',
      likes: p.likes + (liked ? 1 : 0),
      commentsOpen: !!s.commentsOpen[p.id],
      toggleComments: () => this.setState((st) => ({ commentsOpen: { ...st.commentsOpen, [p.id]: !st.commentsOpen[p.id] }, replyFor: null })),
      commentTree: buildComments(p, p.comments || [], []),
      hasComments: (p.comments || []).length > 0,
      addComment: (text) => this.setState((st) => ({ posts: st.posts.map((q) => (q.id === p.id ? { ...q, comments: (q.comments || []).concat([newComment(text)]) } : q)) })),
      media: p.media || [], hasMedia: !!(p.media && p.media.length),
      images: (p.media || []).filter((m) => m.isImg),
      docs: (p.media || []).filter((m) => !m.isImg && !m.isVideo),
      videos: (p.media || []).filter((m) => m.isVideo),
      hasVideos: (p.media || []).some((m) => m.isVideo),
      hasImages: (p.media || []).some((m) => m.isImg),
      hasDocs: (p.media || []).some((m) => !m.isImg && !m.isVideo),
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
      commentCount: p.extra + countAll(p.comments || []),
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
  const baseMoments = [
    { initials: 'BK', color: '#7CC4FF', bg: '#2B4A63', caption: 'Brian at the server room' },
    { initials: 'GN', color: '#FFB86B', bg: '#6B4A2B', caption: 'Team lunch' },
    { initials: 'DO', color: '#F49AC1', bg: '#5A2E47', caption: 'Game night setup' },
    { initials: 'FA', color: '#B8F36B', bg: '#3B5A2B', caption: 'First day!' }
  ];
  const moments = s.myMoments.concat(baseMoments).map((m, i, all) => ({ ...m, open: () => this.setState({ momentOpen: i }), key: i }));
  const pollReady = !!(s.wpoll && s.wpoll.opts.filter((o) => o.trim()).length >= 2 && s.draft.trim());
  const canPost = s.draft.trim().length > 0 && (!s.wpoll || pollReady) || s.wfiles.length > 0;
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
