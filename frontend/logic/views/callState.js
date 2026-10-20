import * as callsW from '~/logic/server/calls.js';
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
    const othersAll = [];   // who is in a call comes from the call room itself
    const pick = s.call.context === 'forum' && !(s.call.people || []).length ? null : s.call.people;
    const others = pick ? pick.filter((n) => n && n !== 'Everyone' && n !== 'Samwel Okoth Omulo').slice(0, 8).map((n, i) => {
      const c = pColor[n] || ((this.dir.find((d) => d.name === n) || {}).color) || '#C9B6FF';
      return { name: n, initials: ini(n), color: c, bg: c + '26', camBg: c + '40', muted: i % 3 === 1, camOn: i % 2 === 0 };
    }) : othersAll;
    const live = !!(this.srv && s.call.peers);
    const othersLive = live ? s.call.peers.map((p) => {
      const d = this.dir.find((x) => x.name === p.name);
      const c = (d && d.color) || '#C9B6FF';
      return { name: p.name, initials: d ? d.initials : ini(p.name), color: c, bg: c + '26', camBg: c + '40', muted: !p.mic, camOn: p.cam || p.sharing, hand: p.hand, speaking: !!p.speaking, videoKey: p.id };
    }) : null;
    const me = { name: 'You', initials: 'S', color: '#F49AC1', bg: '#2B2A1E', camBg: '#3D3B29', muted: !s.mic, camOn: isVideo && (s.cam || s.sharing), you: true, videoKey: live ? 'local' : '' };
    // Moderator = the person who started a meeting, standup or forum call. One-to-one calls have no moderator.
    const big = !!s.call.sfu;
    const isMod = big ? s.call.myRole === 'host' || s.call.myRole === 'moderator' : s.call.context !== 'dm' && s.call.dir !== 'in';
    const off = s.callOff || {};
    const all = [me].concat((othersLive || others).map((o) => ({ ...o, muted: o.muted || !!(off[o.name] && off[o.name].mic), camOn: o.camOn && !(off[o.name] && off[o.name].cam) })));
    const talkers = all.filter((p) => !p.muted);
    const speakerName = live ? '' : talkers.length ? talkers[Math.floor(secs / 4) % talkers.length].name : '';
    const tiles = all.map((p) => {
      const speaking = p.name === speakerName || !!p.speaking;
      const hand = p.you ? s.hand : !!p.hand;
      const mod = isMod && !p.you;
      return { ...p, speaking, hand, mod, micOn: !p.muted, camOnNow: p.camOn && isVideo, camAvail: isVideo,
        muteMic: () => this.modOff(p.name, 'mic'), stopCam: () => this.modOff(p.name, 'cam'), camOff: !p.camOn || !isVideo, camOn: p.camOn && isVideo,
        stateLabel: p.muted ? 'Muted' : speaking ? 'Speaking' : 'Listening',
        style: 'position:relative;aspect-ratio:16/9;min-height:150px;border-radius:18px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;overflow:hidden;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
        stripStyle: 'position:relative;flex:none;width:160px;height:96px;border-radius:12px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
        voiceStyle: 'width:104px;height:104px;border-radius:50%;background:' + p.color + ';color:#0F1714;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:34px;' + (speaking ? 'animation:ring 1.2s ease-out infinite' : 'opacity:0.85') };
    });
    // The people list. In a small room it is the tiles; in a big one it is the whole room, searchable and capped so hundreds of
    // people do not all draw at once. Hosts and moderators can bring a listener on stage or send a speaker back.
    const find = (s.callFind || '').trim().toLowerCase();
    const roster = (s.call.roster || []).filter((r) => !find || (r.name || '').toLowerCase().includes(find));
    const rows = big ? roster.slice(0, 150).map((r) => {
      const d = this.dir.find((x) => x.name === r.name);
      const c = (d && d.color) || '#C9B6FF';
      return { name: r.name, initials: d ? d.initials : ini(r.name), color: c, mod: isMod && r.role !== 'host', micOn: r.mic, camOnNow: r.cam, camAvail: s.call.mode === 'video',
        muteMic: () => this.callCtl('mute', r.id), stopCam: () => this.callCtl('cam_off', r.id),
        stateLabel: r.hand ? 'Hand raised' : r.stage ? (r.mic ? 'On stage' : 'On stage, muted') : 'Listening',
        stageLabel: isMod && r.role === 'participant' ? (r.stage ? 'Send to audience' : r.hand ? 'Bring on stage' : 'Invite to speak') : '',
        toggleStage: () => this.callCtl(r.stage ? 'stage_off' : 'stage_on', r.id) };
    }) : tiles;
    const ctl = (on, danger) => 'width:66px;height:60px;border-radius:16px;border:0;cursor:pointer;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;' + (danger ? 'background:#FF9B8A;color:#0F1714' : on ? 'background:#B8F36B;color:#0F1714' : 'background:var(--k-1D2A25,#1D2A25);color:var(--k-E8EFEA,#E8EFEA)');
    call = {
      ...s.call, timer, isVideo, isVoice: !isVideo, count: big ? (s.call.total || all.length) : all.length, big, isMod, rows,
      moreRows: big && roster.length > 150 ? roster.length - 150 : 0, find: s.callFind || '',
      listening: big && s.call.canSpeak === false,
      limitNote: big && s.call.limit ? 'Up to ' + s.call.limit + ' people' : '',
      e2eeNote: big ? (s.call.e2ee ? 'End-to-end encrypted' : 'Encrypted to the media server') : '',
      onFind: (e) => this.setState({ callFind: e.target.value }),
      muteAll: () => this.callCtl('mute_all'), lockRoom: () => this.callCtl(s.callLocked ? 'unlock' : 'lock'), lockLabel: s.callLocked ? 'Unlock room' : 'Lock room',
      endForAll: () => this.callCtl('end'),
      modeLabel: isVideo ? 'Video meeting' : 'Voice call',
      rec: s.rec, sharing: s.sharing && isVideo, notSharing: !(s.sharing && isVideo), tiles,
      panelOpen: !!s.callPanel, showChat: s.callPanel === 'chat', showPeople: s.callPanel === 'people',
      panelTitle: s.callPanel === 'people' ? 'People (' + (big ? s.call.total || all.length : all.length) + ')' : 'Meeting chat',
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
    if (this.srv) callsW.sendChat(t);
  };
  const allNotifs = taskNotifs.map((n) => ({ ...n, open: () => this.setState({ notifOpen: false, section: 'projects', projectOpen: n.task.pid, taskOpen: n.task }) })).concat(notifList);
  const roleNames = Object.fromEntries((this.state.serverRoles || []).map((r) => [r.id, r.name]));
  return { allNotifs, call, callMiniStyle, restoreCall, roleNames, sendCall, showCallMini, showMeetingDock, showMeetingFull, showPanelCall };
}
