// renderVals step: callState. Receives the names built by earlier steps and returns the ones later steps need.
export function buildCallState(ctx) {
  const { cur, curDm, ini, isDM, notifList, pColor, pw, s, taskNotifs } = ctx;
  // ---- Meeting ----
  let call = {};
  if (s.call) {
    const secs = Math.max(0, Math.floor((Date.now() - s.call.start) / 1000));
    const pad = (n) => (n < 10 ? '0' : '') + n;
    const timer = (secs >= 3600 ? Math.floor(secs / 3600) + ':' : '') + pad(Math.floor((secs % 3600) / 60)) + ':' + pad(secs % 60);
    const isVideo = s.call.mode === 'video';
    const othersAll = [
      { name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', bg: '#1B2E40', camBg: '#24405A', muted: false, camOn: true },
      { name: 'Grace Njeri', initials: 'GN', color: '#FFB86B', bg: '#33271A', camBg: '#4A3826', muted: true, camOn: false },
      { name: 'David Otieno', initials: 'DO', color: '#F49AC1', bg: '#33202C', camBg: '#4A2E40', muted: false, camOn: true },
      { name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', bg: '#16332B', camBg: '#1F4A3E', muted: true, camOn: false },
      { name: 'Faith Achieng', initials: 'FA', color: '#B8F36B', bg: '#22301A', camBg: '#324624', muted: false, camOn: true }
    ];
    const pick = s.call.people;
    const others = pick ? pick.filter((n) => n && n !== 'Everyone' && n !== 'Samwel Okoth Omulo').slice(0, 8).map((n, i) => {
      const c = pColor[n] || ((this.dir.find((d) => d.name === n) || {}).color) || '#C9B6FF';
      return { name: n, initials: ini(n), color: c, bg: c + '26', camBg: c + '40', muted: i % 3 === 1, camOn: i % 2 === 0 };
    }) : (isVideo ? othersAll : othersAll.slice(0, 3));
    const me = { name: 'You', initials: 'S', color: '#F49AC1', bg: '#2B2A1E', camBg: '#3D3B29', muted: !s.mic, camOn: isVideo && s.cam, you: true };
    const all = [me].concat(others);
    const talkers = all.filter((p) => !p.muted);
    const speakerName = talkers.length ? talkers[Math.floor(secs / 4) % talkers.length].name : '';
    const tiles = all.map((p) => {
      const speaking = p.name === speakerName;
      const hand = p.you ? s.hand : false;
      return { ...p, speaking, hand, camOff: !p.camOn || !isVideo, camOn: p.camOn && isVideo,
        stateLabel: p.muted ? 'Muted' : speaking ? 'Speaking' : 'Listening',
        style: 'position:relative;aspect-ratio:16/9;min-height:150px;border-radius:18px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;overflow:hidden;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
        stripStyle: 'position:relative;flex:none;width:160px;height:96px;border-radius:12px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
        voiceStyle: 'width:104px;height:104px;border-radius:50%;background:' + p.color + ';color:#0F1714;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:34px;' + (speaking ? 'animation:ring 1.2s ease-out infinite' : 'opacity:0.85') };
    });
    const ctl = (on, danger) => 'width:66px;height:60px;border-radius:16px;border:0;cursor:pointer;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;' + (danger ? 'background:#FF9B8A;color:#0F1714' : on ? 'background:#B8F36B;color:#0F1714' : 'background:#1D2A25;color:#E8EFEA');
    call = {
      ...s.call, timer, isVideo, isVoice: !isVideo, count: all.length,
      modeLabel: isVideo ? 'Video meeting' : 'Voice call',
      rec: s.rec, sharing: s.sharing && isVideo, notSharing: !(s.sharing && isVideo), tiles,
      panelOpen: !!s.callPanel, showChat: s.callPanel === 'chat', showPeople: s.callPanel === 'people',
      panelTitle: s.callPanel === 'people' ? 'People (' + all.length + ')' : 'Meeting chat',
      msgs: s.callMsgs, draft: s.callDraft,
      micLabel: s.mic ? 'Mute' : 'Unmute', micStyle: ctl(false, !s.mic),
      camLabel: s.cam ? 'Stop video' : 'Start video', camStyle: ctl(false, !s.cam),
      shareStyle: ctl(s.sharing, false), handStyle: ctl(s.hand, false), recStyle: ctl(false, s.rec),
      chatStyle: ctl(s.callPanel === 'chat', false), peopleStyle: ctl(s.callPanel === 'people', false)
    };
  }
  const panelCallMatch = !!s.call && ((s.call.context === 'dm' && isDM && curDm.with === s.call.ref) || (s.call.context === 'forum' && !isDM && cur.id === s.call.ref));
  const isMeetingCall = !!s.call && !s.call.context;
  const showMeetingFull = isMeetingCall && s.callMode === 'full' && s.mode !== 'full' && s.mode !== 'page';
  const showMeetingDock = isMeetingCall && s.callMode === 'docked' && s.mode !== 'full';
  const showPanelCall = panelCallMatch && (s.mode === 'docked' || s.mode === 'full' || s.mode === 'page');
  const showCallMini = !!s.call && !showMeetingFull && !showMeetingDock && !showPanelCall;
  const callMiniStyle = 'position:fixed;z-index:60;right:' + (s.mode === 'docked' ? pw + 36 : 20) + 'px;bottom:' + (s.mode === 'min' ? 92 : 20) + 'px';
  const restoreCall = () => {
    const c = this.state.call;
    if (!c) return;
    if (!c.context) this.setState({ callMode: 'docked', mode: this.state.mode === 'full' ? 'docked' : this.state.mode });
    else if (c.context === 'dm') this.openDm(c.ref);
    else this.setState({ forum: c.ref, panelKind: 'forum', mode: this.state.mode === 'full' || this.state.mode === 'page' ? this.state.mode : 'docked', menu: false, tab: 'discussions' });
  };
  const sendCall = () => {
    const t = this.state.callDraft.trim();
    if (!t) return;
    this.setState((st) => ({ callMsgs: st.callMsgs.concat([{ who: 'You', text: t }]), callDraft: '' }));
  };
  const allNotifs = taskNotifs.map((n) => ({ ...n, open: () => this.setState({ notifOpen: false, section: 'projects', projectOpen: n.task.pid, taskOpen: n.task }) })).concat(notifList);
  const roleNames = { admin: 'Admin', hod: 'Head of department', it: 'IT support', finance: 'Finance approver', staff: 'Staff', guest: 'Guest' };
  return { allNotifs, call, callMiniStyle, restoreCall, roleNames, sendCall, showCallMini, showMeetingDock, showMeetingFull, showPanelCall };
}
