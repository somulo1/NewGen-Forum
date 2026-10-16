// Values and handlers the templates read through hubBindings: panels.
export function valsPanels(ctx) {
  const { chatCtl, chatGate, dmPerson, dmPriv, dmStatusCol, forums, incomingView, isDM, panelStyle, pollView, s, set } = ctx;
  return {
    panelStyle,
    chatCtl, chatGate, composerOff: !chatGate.allowed, composerNote: chatGate.reason, dmPriv, incomingView, pollView,
    panelTab: s.panelTab, isPanelChat: s.panelTab !== 'polls', isPanelPolls: s.panelTab === 'polls',
    goPanelChat: set({ panelTab: 'chat' }), goPanelPolls: set({ panelTab: 'polls' }),
    gripStyle: 'display:block;width:4px;height:44px;border-radius:2px;background:' + (s.dragging ? '#B8F36B' : '#3F524A'),
    selectMode: s.dragging ? 'none' : 'auto',
    startDrag: (e) => this.startDrag(e),
    resizeKey: (e) => {
      if (e.key === 'ArrowLeft') this.setState({ pw: Math.min(640, this.state.pw + 20) });
      if (e.key === 'ArrowRight') this.setState({ pw: Math.max(320, this.state.pw - 20) });
    },
    isDocked: s.mode === 'docked',
    isFull: s.mode === 'full', isFullForum: s.mode === 'full' && !isDM, isFullDm: s.mode === 'full' && isDM, isDM: isDM,
    isPageForum: s.mode === 'page' && !isDM, isPageDm: s.mode === 'page' && isDM, isPage: s.mode === 'page',
    panelChips: isDM ? s.dms.map((d) => { const pr = dmPerson(d.with); const on = d.id === s.dmId; return { name: d.with, initials: pr.initials, open: () => this.openDm(d.with),
      chipStyle: 'flex:none;width:36px;height:36px;border-radius:50%;cursor:pointer;font-weight:700;font-size:12px;color:#0F1714;font-family:inherit;background:' + pr.color + ';border:2px solid ' + (on ? '#E8EFEA' : 'transparent') + ';opacity:' + (on ? '1' : '0.65') }; }) : forums,
    dmList: s.dms.map((d) => { const pr = dmPerson(d.with); const last = d.msgs[d.msgs.length - 1]; const on = isDM && d.id === s.dmId && s.mode !== 'closed';
      return { name: d.with, initials: pr.initials, color: pr.color, dot: dmStatusCol[pr.status] || '#B8F36B', unread: d.unread, hasUnread: d.unread > 0,
        last: last ? (last.from === 'me' ? 'You: ' : '') + (last.text || (last.file ? 'Sent a file' : '')) : 'Say hello', open: () => this.openDm(d.with),
        style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:50px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-family:inherit;color:#E8EFEA;background:' + (on ? '#22302A' : 'transparent') }; }),
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
    genLink: set({ link: true })
  };
}
