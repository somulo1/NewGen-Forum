import * as forumsW from '~/logic/server/widgets/forums.js';
// Values and handlers the templates read through hubBindings: panels.
export function valsPanels(ctx) {
  const { chatCtl, chatGate, dmPerson, dmPriv, dmStatusCol, forums, incomingView, isDM, panelStyle, pollView, s, set } = ctx;
  return {
    panelStyle,
    chatCtl, chatGate, composerOff: !chatGate.allowed, composerNote: chatGate.reason, dmPriv, incomingView, pollView,
    panelTab: s.panelTab, isPanelChat: s.panelTab === 'chat', isPanelPolls: s.panelTab === 'polls', isPanelCalls: s.panelTab === 'calls', isPanelFiles: s.panelTab === 'files',
    goPanelChat: set({ panelTab: 'chat' }), goPanelPolls: set({ panelTab: 'polls' }), goPanelCalls: set({ panelTab: 'calls' }), goPanelFiles: set({ panelTab: 'files' }),
    callHistory: (() => {
      const { cur, curDm, dmView } = ctx;
      const mine = (e) => (isDM ? e.ctx === 'dm' && curDm && e.ref === curDm.with : e.ctx === 'forum' && cur && e.ref === cur.id);
      const fmtDur = (n) => (n >= 3600 ? Math.floor(n / 3600) + ' h ' + Math.round((n % 3600) / 60) + ' min' : n >= 60 ? Math.floor(n / 60) + ' min ' + (n % 60) + ' s' : n + ' s');
      const fmtWhen = (t) => { const d = new Date(t), now = new Date(); const day = d.toDateString() === now.toDateString() ? 'Today' : d.toDateString() === new Date(now - 86400000).toDateString() ? 'Yesterday' : d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short' }); return day + ' · ' + d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' }); };
      const label = { out: 'Outgoing', in: 'Incoming', missed: 'Missed', declined: 'Declined', blocked: 'Blocked' };
      const color = { out: '#B8F36B', in: '#7CC4FF', missed: '#FF5C5C', declined: '#FFB86B', blocked: '#C9B6FF' };
      const rows = s.callLog.filter(mine).sort((a, b) => b.at - a.at).map((e) => ({ id: e.id, title: (label[e.dir] || 'Call') + ' ' + (e.kind === 'video' ? 'video call' : 'voice call'), color: color[e.dir] || '#9DB0A6', arrow: e.dir === 'in' ? '↙' : e.dir === 'out' ? '↗' : e.dir === 'blocked' ? '⦸' : '✕',
        sub: fmtWhen(e.at) + (e.dur ? ' · ' + fmtDur(e.dur) : ''), isVideo: e.kind === 'video' }));
      return { rows, empty: rows.length === 0, who: isDM ? (curDm ? curDm.with : '') : (cur ? cur.name : ''),
        callVoice: isDM && curDm ? () => this.startDmCall('voice', curDm.with, dmView.initials, dmView.color) : () => this.startCall('voice', cur.name, cur.initials, cur.color, [], 'forum', cur.id),
        callVideo: isDM && curDm ? () => this.startDmCall('video', curDm.with, dmView.initials, dmView.color) : () => this.startCall('video', cur.name, cur.initials, cur.color, [], 'forum', cur.id) };
    })(),
    gripStyle: 'display:block;width:4px;height:44px;border-radius:2px;background:' + (s.dragging ? '#B8F36B' : 'var(--k-3F524A,#3F524A)'),
    selectMode: s.dragging ? 'none' : 'auto',
    startDrag: (e) => this.startDrag(e),
    resizeKey: (e) => {
      if (e.key === 'ArrowLeft') this.setState({ pw: Math.min(this.panelMax(), this.state.pw + 20) });
      if (e.key === 'ArrowRight') this.setState({ pw: Math.max(320, this.state.pw - 20) });
    },
    isDocked: s.mode === 'docked',
    isFull: s.mode === 'full', isFullForum: s.mode === 'full' && !isDM, isFullDm: s.mode === 'full' && isDM, isDM: isDM,
    isPageForum: s.mode === 'page' && !isDM, isPageDm: s.mode === 'page' && isDM, isPage: s.mode === 'page',
    // The strip shows this kind of conversation (forums, or chats) and, after a divider, anything of the other kind with unread
    // messages, so you can see and open what is waiting from either screen.
    panelChips: (() => {
      const mine = (isDM ? s.dms.map((d) => { const pr = dmPerson(d.with); const on = d.id === s.dmId; return { name: d.with, initials: pr.initials, unread: d.unread, hasUnread: d.unread > 0, open: () => this.openDm(d.with),
      chipStyle: 'flex:none;width:36px;height:36px;border-radius:50%;cursor:pointer;font-weight:700;font-size:12px;color:#0F1714;font-family:inherit;background:' + pr.color + ';border:2px solid ' + (on ? '#E8EFEA' : 'transparent') + ';opacity:' + (on ? '1' : '0.65'),
      unreadStyle: 'position:absolute;top:-5px;right:-6px;min-width:17px;height:17px;padding:0 4px;box-sizing:border-box;border-radius:9px;background:#C9B6FF;color:#0F1714;font:700 10px sans-serif;display:flex;align-items:center;justify-content:center;pointer-events:none' }; }) : forums).map((c) => ({ ...c, avatar: isDM, sep: false }));
      const dmChip = (d) => { const pr = dmPerson(d.with); return { name: d.with, initials: pr.initials, unread: d.unread, hasUnread: true, avatar: true, color: pr.color, open: () => this.openDm(d.with),
        chipStyle: 'flex:none;width:36px;height:36px;border-radius:50%;cursor:pointer;font-weight:700;font-size:12px;color:#0F1714;font-family:inherit;background:' + pr.color + ';border:2px solid transparent',
        unreadStyle: 'position:absolute;top:-5px;right:-6px;min-width:17px;height:17px;padding:0 4px;box-sizing:border-box;border-radius:9px;background:#C9B6FF;color:#0F1714;font:700 10px sans-serif;display:flex;align-items:center;justify-content:center;pointer-events:none' }; };
      const other = isDM ? forums.filter((f) => f.hasUnread).map((f) => ({ ...f, avatar: false })) : s.dms.filter((d) => d.unread > 0).map(dmChip);
      if (other.length) other[0] = { ...other[0], sep: true };
      return mine.concat(other);
    })(),
    dmList: s.dms.map((d) => { const pr = dmPerson(d.with); const last = d.msgs[d.msgs.length - 1]; const on = isDM && d.id === s.dmId && s.mode !== 'closed';
      return { name: d.with, initials: pr.initials, color: pr.color, dot: dmStatusCol[pr.status] || '#B8F36B', unread: d.unread, hasUnread: d.unread > 0,
        last: last ? (last.from === 'me' ? 'You: ' : '') + (last.text || (last.file ? 'Sent a file' : '')) : 'Say hello', open: () => this.openDm(d.with),
        style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:50px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-family:inherit;color:var(--k-E8EFEA,#E8EFEA);background:' + (on ? 'var(--k-22302A,#22302A)' : 'transparent') }; }),
    dmsOpen: s.dmsOpen, toggleDms: () => this.setState((x) => ({ dmsOpen: !x.dmsOpen })),
    dmsChevron: 'transition:transform 0.2s;transform:rotate(' + (s.dmsOpen ? '0' : '-90') + 'deg)',
    dmUnreadLabel: (() => { const n = s.dms.reduce((a, d) => a + d.unread, 0); return n ? n + ' new' : String(s.dms.length); })(),
    newDmOpen: s.newDmOpen, toggleNewDm: () => this.setState((x) => ({ newDmOpen: !x.newDmOpen, dmsOpen: true, dmq: '' })),
    dmq: s.dmq, onDmq: (e) => this.setState({ dmq: e.target.value }),
    dmMatches: (s.dmq.trim() ? this.dir.filter((d) => d.name.toLowerCase().indexOf(s.dmq.trim().toLowerCase()) !== -1) : this.dir.slice(-9)).slice(0, 6).map((d) => ({ name: d.name, initials: d.initials, color: d.color, title: d.title, open: () => this.openDm(d.name) })),
    isMin: s.mode === 'min',
    menuOpen: s.menu,
    inviteOpen: s.invite,
    createOpen: s.create,
    linkMade: s.link,
    toDocked: set({ mode: 'docked', menu: false }),
    toFull: set({ mode: 'full', menu: false }),
    toPage: set({ mode: 'page', menu: false }),
    toMin: set({ mode: 'min', menu: false }),
    toClosed: set({ mode: 'closed', menu: false }),
    toggleMenu: () => this.setState({ menu: !s.menu }),
    toggleInvite: () => this.setState({ invite: !s.invite }),
    toggleCreate: () => this.setState({ create: !s.create }),
    // Make a forum (named, with the people to start with) and invite more people into the one that is open. Both need the server.
    forumForm: s.forumForm,
    onForumName: (e) => this.setState({ forumForm: { ...s.forumForm, name: e.target.value, err: '' } }),
    onForumMembers: (e) => this.setState({ forumForm: { ...s.forumForm, members: e.target.value, err: '' } }),
    createForumNow: () => {
      const f = this.state.forumForm;
      if (!this.srv) { this.notify('Sign in to an organization to create forums.', '#FFB86B'); return; }
      this.setState({ forumForm: { ...f, busy: true, err: '' } });
      forumsW.createForum(this, this.srv, f).then((id) => {
        this.setState({ forumForm: { name: '', members: '', err: '', busy: false }, create: false, forum: id, panelKind: 'forum', mode: this.isMobileViewport() ? 'page' : 'docked', menu: false, tab: 'discussions', forumsOpen: true });
        this.notify('Forum created.', '#B8F36B');
      }).catch((e) => this.setState({ forumForm: { ...f, busy: false, err: e.message || 'Could not create the forum.' } }));
    },
    inviteText: s.inviteText, inviteErr: s.inviteErr, inviteBusy: s.inviteBusy,
    onInviteText: (e) => this.setState({ inviteText: e.target.value, inviteErr: '' }),
    sendInvites: () => {
      if (!this.srv || !forumsW.isLive(this.srv, this.state.forum)) { this.notify('This forum is not on the server.', '#FFB86B'); return; }
      this.setState({ inviteBusy: true, inviteErr: '' });
      forumsW.invite(this, this.srv, this.state.forum, this.state.inviteText)
        .then((n) => { this.setState({ inviteBusy: false, inviteText: '', invite: false }); this.notify(n + (n === 1 ? ' person' : ' people') + ' added to the forum.', '#B8F36B'); })
        .catch((e) => this.setState({ inviteBusy: false, inviteErr: e.message || 'Could not add them.' }));
    },
    genLink: set({ link: true })
  };
}
