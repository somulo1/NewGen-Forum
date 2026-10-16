// Values and handlers the templates read through hubBindings: shell.
export function valsShell(ctx) {
  const { denied, can, forums, go, myUnits, ticketDeskOk, allNotifs, call, callMiniStyle, cur, curDm, curSec, dmView, has, isDM, restoreCall, s, sendCall, set, showCallMini, showMeetingDock, showMeetingFull, showPanelCall } = ctx;
  const nq = (s.navq || '').trim().toLowerCase();
  const hit = (txt) => txt.toLowerCase().indexOf(nq) !== -1;
  const clear = () => this.setState({ navq: '' });
  const pages = [['Home', 'home', true], ['Community wall', 'feed', true], ['Meetings & standups', 'meetings', has.sessions], ['Calendar', 'calendar', has.calendar], ['Communications', 'comms', has.comms],
    ['My tickets', 'tickets', has.tickets], ['Ticket desk', 'ticketdesk', ticketDeskOk], ['Projects', 'projects', has.projects], ['Files', 'files', has.files], ['Secure share', 'share', has.secureshare],
    ['People', 'people', true], ['Approvals', 'approvals', true], ['Security & logs', 'security', can('security', 'View alerts') || can('security', 'View logs')], ['Roles & rights', 'roles', can('roles', 'View roles')],
    ['My profile', 'profile', true], ['Widgets', 'widgets', can('home', 'Open widgets page')]];
  const results = [];
  if (nq) {
    pages.filter((p) => p[2] && hit(p[0])).forEach((p) => results.push({ group: 'Pages', label: p[0], sub: '', initials: p[0][0], color: '#E8EFEA', run: () => { go[p[1]](); clear(); } }));
    forums.filter((f) => hit(f.name)).forEach((f) => results.push({ group: 'Forums', label: f.name, sub: f.members + ' members', initials: f.initials, color: f.color, run: () => { f.open(); clear(); } }));
    myUnits.filter((u) => hit(u.name)).forEach((u) => results.push({ group: 'Organization', label: u.name, sub: u.type, initials: u.short, color: u.color, run: () => { u.open(); clear(); } }));
    if (has.chat) this.dir.filter((d) => hit(d.name)).slice(0, 6).forEach((d) => results.push({ group: 'People', label: d.name, sub: d.title, initials: d.initials, color: d.color, run: () => { this.openDm(d.name); clear(); } }));
  }
  const navSearch = { q: s.navq || '', active: !!nq, onInput: (e) => this.setState({ navq: e.target.value }), clear, none: !!nq && results.length === 0, results: results.slice(0, 20) };
  return {
    navSearch, denied,
    notifs: allNotifs, noNotifs: allNotifs.length === 0,
    notifOpen: s.notifOpen,
    hasBadge: !s.notifSeen && allNotifs.length > 0,
    badgeCount: allNotifs.length > 9 ? '9+' : String(allNotifs.length),
    toggleNotifs: () => this.setState((st) => ({ notifOpen: !st.notifOpen, notifSeen: true })),
    hasToast: !!s.toast, toast: s.toast, toastColor: s.toastColor,
    inCall: !!s.call, call,
    navOpen: s.navOpen, navClass: 'hub-nav' + (s.navOpen ? ' open' : ''),
    openNav: () => this.setState({ navOpen: true, notifOpen: false }), closeNav: () => this.setState({ navOpen: false }),
    navClick: (e) => { if (!this.state.navOpen) return; const b = e.target && e.target.closest ? e.target.closest('button') : null; if (b && !b.closest('[data-keep]')) this.setState({ navOpen: false }); },
    showTabbar: s.mode !== 'full' && !s.call,
    tabbar: (() => {
      const unread = s.dms.reduce((a, d) => a + d.unread, 0);
      const st = (on) => 'flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;border:0;background:transparent;cursor:pointer;font-family:inherit;color:' + (on ? '#B8F36B' : '#9DB0A6') + ';font-weight:' + (on ? '600' : '500');
      const items = [
        { label: 'Home', icon: 'M3 11l9-7 9 7v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z', on: curSec === 'home', go: () => this.setState({ section: 'home', navOpen: false, ticketOpen: null, projectOpen: null }) },
        { label: 'Wall', icon: 'M5 3h14v18H5zM8 8h8M8 12h8M8 16h5', on: curSec === 'feed', go: () => this.setState({ section: 'feed', navOpen: false }) }
      ];
      if (has.chat) items.push({ label: 'Messages', icon: 'M4 5h16v11H9l-5 4z', on: isDM && (s.mode === 'docked'), dot: unread, go: () => { if (s.dms.length) this.openDm((s.dms.find((d) => d.unread) || s.dms[0]).with); else this.setState({ navOpen: true, dmsOpen: true, newDmOpen: true }); } });
      if (has.tickets) items.push({ label: 'Tickets', icon: 'M3 7h18v4a2 2 0 0 0 0 4v4H3v-4a2 2 0 0 0 0-4z', on: curSec === 'tickets', go: () => this.setState({ section: 'tickets', ticketOpen: null, navOpen: false }) });
      else if (has.calendar) items.push({ label: 'Calendar', icon: 'M3 5h18v16H3zM3 10h18M8 3v4M16 3v4', on: curSec === 'calendar', go: () => this.setState({ section: 'calendar', navOpen: false }) });
      items.push({ label: 'Menu', icon: 'M4 7h16M4 12h16M4 17h16', on: s.navOpen, go: () => this.setState((x) => ({ navOpen: !x.navOpen })) });
      return items.map((t) => ({ ...t, hasDot: !!t.dot, dot: t.dot || 0, style: st(t.on) }));
    })(),
    startInstant: () => this.startCall('video', 'Instant meeting', 'S', '#F49AC1', []),
    voiceCall: isDM ? () => this.startDmCall('voice', curDm.with, dmView.initials, dmView.color) : () => this.startCall('voice', cur.name, cur.initials, cur.color, [], 'forum', cur.id),
    videoCall: isDM ? () => this.startDmCall('video', curDm.with, dmView.initials, dmView.color) : () => this.startCall('video', cur.name, cur.initials, cur.color, [], 'forum', cur.id),
    endCall: () => this.endCall(),
    showMeetingFull, showMeetingDock, showPanelCall, showCallMini, callMiniStyle, restoreCall,
    callToFull: set({ callMode: 'full' }), callToDock: set({ callMode: 'docked' }), callToMin: set({ callMode: 'min' }),
    toggleMic: () => this.setState((st) => ({ mic: !st.mic })),
    toggleCam: () => this.setState((st) => ({ cam: !st.cam })),
    toggleShare: () => this.setState((st) => ({ sharing: !st.sharing })),
    toggleHand: () => this.setState((st) => ({ hand: !st.hand })),
    toggleRec: () => this.setState((st) => ({ rec: !st.rec })),
    openCallChat: () => this.setState((st) => ({ callPanel: st.callPanel === 'chat' ? null : 'chat' })),
    openCallPeople: () => this.setState((st) => ({ callPanel: st.callPanel === 'people' ? null : 'people' })),
    closeCallPanel: set({ callPanel: null }),
    onCallDraft: (e) => this.setState({ callDraft: e.target.value }),
    callKey: (e) => { if (e.key === 'Enter') sendCall(); },
    sendCallMsg: sendCall
  };
}
