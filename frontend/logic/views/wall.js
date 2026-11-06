import * as forumsW from '~/logic/server/widgets/forums.js';
import { shareUrl } from '~/logic/shareOut.js';
import * as wallW from '~/logic/server/widgets/wall.js';
import { richTextHtml, mentionLinker, hashtagsOf, linkHashtags } from '~/logic/richText.js';
// renderVals step: wall. Receives the names built by earlier steps and returns the ones later steps need.
export function buildWall(ctx) {
  const { can, flowById, forumsData, makeRequest, openPersonName, s } = ctx;
  // a tagged person (@Name) is a link to their profile
  const linkMentions = mentionLinker(this.dir.map((d) => d.name).concat(this.srv ? [this.srv.me.name] : []));
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
      style: 'display:flex;align-items:center;min-height:44px;padding:0 14px;border-radius:10px;cursor:pointer;font-size:14px;color:var(--k-E8EFEA,#E8EFEA);text-align:left;border:1px solid ' + (mine ? '#B8F36B' : '#2A3832') + ';background:' + (s.vote ? 'linear-gradient(90deg,var(--k-2C3D34,#2C3D34) ' + pct + '%,#121B17 ' + pct + '%)' : 'var(--k-121B17,#121B17)')
    };
  });
  const tagHit = (p) => !s.wallTag || hashtagsOf(p.text).some((t) => t.toLowerCase() === s.wallTag.toLowerCase());
  const countAll = (list) => list.reduce((n, c) => n + 1 + countAll(c.replies || []), 0);
  const newComment = (text) => ({ who: 'Sam', initials: 'S', color: '#F49AC1', text, time: 'Just now', replies: [] });
  // Adds a reply under the comment at `path` (indexes from the top-level list down).
  const addAt = (list, path, text) => list.map((c, i) => (i !== path[0] ? c : path.length === 1 ? { ...c, replies: (c.replies || []).concat([newComment(text)]) } : { ...c, replies: addAt(c.replies || [], path.slice(1), text) }));
  const buildComments = (p, list, path) => list.map((c, i) => {
    const here = path.concat([i]);
    const key = p.id + ':' + here.join('.');
    return { key, openWho: () => openPersonName(c.mine ? 'Sam' : c.who), who: c.who, uid: c.uid || '', self: !!c.mine, initials: c.initials, color: c.color, text: c.text, htmlText: linkHashtags(linkMentions(richTextHtml(c.text))), time: c.time || '', depth: path.length, replying: s.replyFor === key,
      startReply: () => this.setState({ replyFor: key }), cancel: () => this.setState({ replyFor: null }),
      send: (text) => { this.setState({ replyFor: null }); wallW.comment(this, this.srv, c.id || p.id, text).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); },
      id: c.id || '', likes: c.likes || 0, dislikes: c.dislikes || 0, iLiked: !!c.iLiked, iDisliked: !!c.iDisliked, canReact: !!(this.srv && c.id),
      react: (kind) => {
        if (!this.srv || !c.id || (this._reacting || (this._reacting = new Set())).has(c.id)) return;
        this._reacting.add(c.id);
        wallW.reactToComment(this, this.srv, c.id, kind).catch((e) => this.notify((e && e.message) || 'That did not work.', '#FF9B8A')).finally(() => { this._reacting.delete(c.id); this.forceUpdate(); });
      },
      replies: buildComments(p, c.replies || [], here) };
  });
  // Every option always shows how many voted for it and what share that is, so everyone can follow the poll before and after voting.
  const userPollOf = (p) => (p.pollData ? (() => {
      const my = s.pollVotes[p.id];
      const tot = p.pollData.opts.reduce((n, o) => n + o.votes, 0) + (my ? 1 : 0);
      return { tot, list: p.pollData.opts.map((o) => { const v = o.votes + (my === o.id ? 1 : 0); const pc = tot ? Math.round((v / tot) * 100) : 0;
        return { label: o.label + (my === o.id ? '  ✓' : ''), count: v, pct: pc + '%',
          vote: () => { if (my) { this.notify('You have already voted on this poll. Your first vote is final.', '#FFB86B'); return; } this.setState((st) => ({ pollVotes: { ...st.pollVotes, [p.id]: o.id } })); wallW.vote(this, this.srv, p.id, o.id).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); },
          style: 'display:flex;align-items:center;gap:10px;min-height:44px;padding:0 14px;border-radius:10px;cursor:' + (my ? 'default' : 'pointer') + ';font-size:14px;color:var(--k-E8EFEA,#E8EFEA);text-align:left;border:1px solid ' + (my === o.id ? '#B8F36B' : '#2A3832') + ';background:linear-gradient(90deg,' + (my === o.id ? 'rgba(184,243,107,0.28)' : 'var(--k-2C3D34,#2C3D34)') + ' ' + pc + '%,var(--k-121B17,#121B17) ' + pc + '%);transition:background .5s ease' }; }) };
  })() : null);
  const posts = s.posts.filter((p) => !s.hiddenPosts[p.id] && tagHit(p)).map((p) => {
    const liked = !!s.liked[p.id];
    const mineP = p.author === 'Sam' || !!p.mine;
    const saved = !!s.savedPosts[p.id];
    const closeMenu = { postMenu: null };
    const userPoll = userPollOf(p);
    return {
      ...p, htmlText: linkHashtags(linkMentions(richTextHtml(p.text))),
      menuOpen: s.postMenu === p.id,
      toggleMenu: () => this.setState((st) => ({ postMenu: st.postMenu === p.id ? null : p.id })),
      menuItems: [
        { label: saved ? 'Unsave post' : 'Save post', run: () => this.setState((st) => ({ ...closeMenu, savedPosts: { ...st.savedPosts, [p.id]: !saved } })) },
        { label: 'Copy link', run: () => this.setState((st) => ({ ...closeMenu, shareNote: { ...st.shareNote, [p.id]: 'Link copied. Paste it anywhere in Hub.' } })) },
        ...(!this.srv || can('wall', 'Hide posts') ? [{ label: 'Hide this post', run: () => { this.setState((st) => ({ ...closeMenu, hiddenPosts: { ...st.hiddenPosts, [p.id]: true } })); this.notify('Post hidden. It will not show on your wall.', '#9DB0A6'); } }] : []),
        // "Delete own posts" lets someone delete what they wrote, "Delete others' posts" what anyone wrote. It leaves the wall at once and the
        // server erases it for good, with its pictures and videos, so the space is free straight away.
        ...((mineP && (!this.srv || can('wall', 'Delete own posts'))) || (this.srv && can('wall', "Delete others' posts"))
          ? [{ label: 'Delete post', danger: true, run: () => { this.setState(closeMenu); wallW.remove(this, this.srv, p.id).then(() => this.notify(mineP ? 'Your post was deleted.' : 'The post was deleted.', '#9DB0A6')).catch((e) => this.notify(e.message || 'The post could not be deleted.', '#FF9B8A')); } }]
          : []),
        ...(mineP ? [] : [{ label: 'Report post', danger: true, run: () => { this.setState(closeMenu); this.notify('Thanks. The post was reported to the moderators.', '#FFB86B'); } }])
      ],
      saved, savedLabel: saved ? 'Saved' : '',
      hasPoll: !!p.hasPoll || !!userPoll,
      hasPhoto: !!p.photoBg,
      poll: userPoll ? userPoll.list : poll,
      pollNote: userPoll ? (userPoll.tot + (userPoll.tot === 1 ? ' vote' : ' votes') + (s.pollVotes[p.id] ? ' · thanks for voting' : ' · tap an option to vote')) : s.vote ? total + ' votes · thanks for voting' : 'Tap an option to vote · closes Friday',
      likes: p.likes + (liked ? 1 : 0),
      commentsOpen: !!s.commentsOpen[p.id],
      replyingAny: !!s.replyFor && s.replyFor.indexOf(p.id + ':') === 0,
      toggleComments: () => this.setState((st) => ({ commentsOpen: { ...st.commentsOpen, [p.id]: !st.commentsOpen[p.id] }, replyFor: null })),
      commentTree: buildComments(p, p.comments || [], []),
      hasComments: (p.comments || []).length > 0,
      addComment: (text) => wallW.comment(this, this.srv, p.id, text).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')),
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
      openAuthor: () => openPersonName(p.mine ? 'Sam' : p.author), hasTags: !!(p.tags && p.tags.length), tagList: (p.tags || []).map((n) => ({ name: n, open: () => openPersonName(n) })),
      previewImg: ((p.media || []).find((m) => m.isImg) || {}).url || '',
      hasPreviewImg: (p.media || []).some((m) => m.isImg),
      hasPreviewBg: !!p.photoBg && !(p.media || []).some((m) => m.isImg),
      previewText: (p.text || 'Shared a file').length > 110 ? p.text.slice(0, 110) + '…' : (p.text || 'Shared a file'),
      shareOutBlocked: !can('wall', 'Share outside Hub'),
      shareOut: can('wall', 'Share outside Hub') ? ['WhatsApp', 'Email', 'LinkedIn', 'X', 'Facebook'].map((app) => ({ label: app, run: () => {
        const f = flowById('f4');
        if (f && f.levels > 0) {
          makeRequest('f4', 'Share ' + p.author + '\'s post to ' + app, { kind: 'share', postId: p.id, app: app });
          this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Waiting for approval. Once approved you get a notification that opens ' + app + ' with the post.' } }));
        } else {
          const win = typeof window !== 'undefined' ? window.open(shareUrl(app, p, pollInfo(p).list.map((o) => o.label.replace(/\s+✓$/, ''))), '_blank') : null;
          if (win) win.opener = null;
          this.setState((st) => ({ shareFor: null, shareCount: { ...st.shareCount, [p.id]: (st.shareCount[p.id] || 0) + 1 }, shareNote: { ...st.shareNote, [p.id]: win ? 'Opened ' + app + ' with the post. Finish posting there; reactions are not tracked once it is out.' : 'Your browser blocked the ' + app + ' window. Allow pop-ups and try again.' } }));
        }
      } })) : [],
      toggleShare: () => this.setState((st) => ({ shareFor: st.shareFor === p.id ? null : p.id })),
      copyLink: () => this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Link copied. Paste it anywhere in Hub.' } })),
      shareTargets: forumsData.map((f) => ({ name: f.name, initials: f.initials, color: f.color, pick: () => {
        const sid = 'sp' + Date.now();
        // The post goes to the forum as it is: text, images and poll. Reactions do not follow it; polls stay one poll.
        this.addForumMsg(f.id, '', undefined, undefined, { sid, postId: p.id, snap: { author: p.author, uid: p.uid || '', initials: p.initials, color: p.color, meta: p.meta, text: p.text || '', media: p.media || [] } });
        this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Shared to ' + f.name + ' as it is.' } }));
        wallW.shared(this, this.srv, p.id, f.name).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A'));
      } })),
      commentCount: p.extra + countAll(p.comments || []),
      heartFill: liked ? 'currentColor' : 'none',
      like: () => { this.setState({ liked: { ...this.state.liked, [p.id]: !liked } }); wallW.toggleReaction(this, this.srv, p.id, 'like').catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); },
      likeStyle: 'height:40px;border-radius:10px;border:0;cursor:pointer;font-size:14px;display:flex;align-items:center;justify-content:center;gap:8px;background:transparent;font-weight:' + (liked ? '600' : '400') + ';color:' + (liked ? '#FF9B8A' : '#C9D6CF'),
      onComment: (e) => {
        const v = (e.target.value || '').trim();
        if (e.key !== 'Enter' || !v) return;
        e.target.value = '';
        wallW.comment(this, this.srv, p.id, v).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A'));
      }
    };
  });
  // One place that knows a post's poll, so the wall and a forum the post was shared to always show the same votes.
  const pollInfo = (p) => {
    const u = userPollOf(p);
    if (u) return { list: u.list, note: u.tot + (u.tot === 1 ? ' vote' : ' votes') + (s.pollVotes[p.id] ? ' · thanks for voting' : ' · tap an option to vote'), isPoll: true };
    if (p.hasPoll) return { list: poll, note: s.vote ? total + ' votes · thanks for voting' : 'Tap an option to vote · closes Friday', isPoll: true };
    return { list: [], note: '', isPoll: false };
  };
  const sharedViews = {};
  const addedMsgs = Object.values(s.added || {}).reduce((acc, arr) => acc.concat(arr), []);
  forumsData.forEach((f) => (f.updates || []).forEach((m) => { if (m.shared) addedMsgs.push({ ...m, __forum: f.id }); }));
  addedMsgs.filter((m) => m.shared).forEach((m) => {
    const sh = m.shared, orig = s.posts.find((q) => q.id === sh.postId), snap = sh.snap || {};
    const pi = orig ? pollInfo(orig) : { list: [], note: '', isPoll: false };
    // Polls stay one poll with one thread; anything else keeps its own comments in the forum only.
    const own = pi.isPoll && orig ? (orig.comments || []) : (s.sharedComments[sh.sid] || []);
    sharedViews[sh.sid] = { ...snap, uid: (orig && orig.uid) || snap.uid || '', hasMedia: !!(snap.media && snap.media.length), isPoll: pi.isPoll, poll: pi.list, pollNote: pi.note, comments: own.map((c) => ({ who: c.who, uid: c.uid || '', self: !!c.mine, initials: c.initials, color: c.color, text: c.text, time: c.time || '' })), hasComments: own.length > 0,
      syncNote: pi.isPoll ? 'Votes and comments are shared with the original post.' : 'Reactions stay on the wall. Comments here stay in this forum.',
      addComment: (text) => {
        // saved on the server: a comment on a shared poll is a comment on the wall post; any other comment is a reply in the forum
        if (this.srv && pi.isPoll && orig) { wallW.comment(this, this.srv, orig.id, text).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); return; }
        if (this.srv && m.__forum && m.mid) { forumsW.send(this, this.srv, m.__forum, { text, replyTo: m.mid }).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); return; }
        return this.setState((st) => (pi.isPoll && orig
        ? { posts: st.posts.map((q) => (q.id === orig.id ? { ...q, comments: (q.comments || []).concat([{ who: 'Sam', initials: 'S', color: '#F49AC1', text, time: 'Just now', replies: [] }]) } : q)) }
        : { sharedComments: { ...st.sharedComments, [sh.sid]: (st.sharedComments[sh.sid] || []).concat([{ who: 'Sam', initials: 'S', color: '#F49AC1', text, time: 'Just now' }]) } }));
      } };
  });
  const moments = s.moments.map((m, i) => ({ ...m, key: m.id, mine: m.owner === (this.srv ? this.srv.me.name : 'Sam'), hasMedia: !!m.media, mediaUrl: m.media ? m.media.url : '', mediaIsImg: !!(m.media && m.media.isImg), mediaIsVideo: !!(m.media && m.media.isVideo),
    likeCount: m.likes.length, loveCount: m.loves.length, open: () => this.setState({ momentOpen: i }) }));
  const pollReady = !!(s.wpoll && s.wpoll.opts.filter((o) => o.trim()).length >= 2 && s.draft.trim());
  const canPost = s.draft.trim().length > 0 && (!s.wpoll || pollReady) || s.wfiles.length > 0;
  const pw = Math.min(s.pw, this.panelMax());
  const panelStyle = 'position:relative;flex:0 0 ' + pw + 'px;width:' + pw + 'px;max-width:60%;min-height:0;box-sizing:border-box;margin:12px;border-radius:22px;border:2px solid var(--k-6B7A72,#6B7A72);background:var(--k-131D19,#131D19);box-shadow:none;display:flex;flex-direction:column;overflow:hidden';
  const prio = (p) => {
    const c = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' }[p];
    return 'padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;border:1px solid ' + c + ';color:' + c;
  };
  return { canPost, moments, panelStyle, posts, prio, pw, sharedViews };
}
