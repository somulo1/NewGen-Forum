// renderVals step: forums. Receives the names built by earlier steps and returns the ones later steps need.
export function buildForums(ctx) {
  const { s } = ctx;
  const forumsData = [
    { id: 'it', name: 'IT Operations', initials: 'IT', color: '#B8F36B', unread: 3, members: 42, online: 9, privacy: 'Invite only', about: 'Day-to-day IT support, outages and change announcements.',
      updates: [
        { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '2 min ago', text: 'Email server patch is done. Restart Outlook if you see sync errors.', replies: 4 },
        { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '14 min ago', text: 'Ticket IT-208 moved to In Progress in Jira.', replies: 1 },
        { who: 'David O.', initials: 'DO', color: '#F49AC1', time: '1 h ago', text: 'Reminder: VPN certificates expire Friday. Renew from the self-service page.', replies: 7 },
        { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: '3 h ago', text: 'Shared "Network diagram v3" in Files.', replies: 0 }
      ] },
    { id: 'dev', name: 'Product Builders', initials: 'PB', color: '#7CC4FF', unread: 5, members: 18, online: 6, privacy: 'Members of Engineering', about: 'Developers, designers and testers shipping the product.',
      updates: [
        { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: 'Just now', text: 'PR #312 merged into main via GitHub. Staging deploy running.', replies: 2 },
        { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '20 min ago', text: 'Sprint review session starts at 3pm. Join from the Sessions tab.', replies: 5 },
        { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '1 h ago', text: 'Shared staging credentials via a one-time link. Check your DMs.', replies: 0 }
      ] },
    { id: 'game', name: 'Weekend Gamers', initials: 'WG', color: '#FFB86B', unread: 0, members: 64, online: 12, privacy: 'Open to everyone', about: 'FIFA nights, tournaments and banter. Everyone welcome.',
      updates: [
        { who: 'David O.', initials: 'DO', color: '#F49AC1', time: '10 min ago', text: 'Tournament bracket is up. Vote on the game in the feed poll!', replies: 11 },
        { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '2 h ago', text: 'Who is on for Saturday 6pm?', replies: 9 }
      ] },
    { id: 'choir', name: 'Choir & Worship', initials: 'CW', color: '#F49AC1', unread: 1, members: 27, online: 3, privacy: 'Request to join', about: 'Rehearsal schedules, song lists and announcements.',
      updates: [
        { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '30 min ago', text: 'Rehearsal moved to Thursday 5:30pm, main hall.', replies: 3 }
      ] },
    { id: 'risk', name: 'Risk & Compliance', initials: 'RC', color: '#6EE7C8', unread: 0, members: 11, online: 2, privacy: 'Invite only', about: 'Policies, audits and control reviews.',
      updates: [
        { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: '1 h ago', text: 'New access-review policy uploaded. Please acknowledge by Monday.', replies: 2 }
      ] }
  ];
  const chatters = [
    { who: 'Brian K.', initials: 'BK', color: '#7CC4FF' },
    { who: 'Grace N.', initials: 'GN', color: '#FFB86B' },
    { who: 'David O.', initials: 'DO', color: '#F49AC1' },
    { who: 'Amina W.', initials: 'AW', color: '#6EE7C8' },
    { who: 'Faith A.', initials: 'FA', color: '#B8F36B' },
    { who: 'Kevin M.', initials: 'KM', color: '#C9B6FF' }
  ];
  const lines = {
    it: ['Printer on 2nd floor is back online.', 'Anyone else seeing slow Wi-Fi in the boardroom?', 'Yes, checking the access point now.', 'Fixed. It was a firmware hang.', 'Please lock your screens when stepping away.', 'New laptops arrive Monday, list is in Files.', 'Backup job finished successfully overnight.', 'Password reset requests go through Tickets please.', 'Thanks for the quick fix this morning!', 'Antivirus update rolling out at 1pm.', 'Server room AC was serviced today.', 'Who has the spare HDMI adapter?', 'It is in the IT cabinet, top drawer.', 'Scheduled downtime Saturday 8 to 10pm.'],
    dev: ['Build is green again.', 'Can someone review PR #309?', 'On it after standup.', 'Staging DB was refreshed from yesterday\'s snapshot.', 'Found the bug: timezone offset on reports.', 'Nice catch!', 'Merging the hotfix now.', 'Design for the mobile settings screen is in Files.', 'Who is on call this weekend?', 'Me, ping me on the forum.', 'Unit test coverage is now 78%.', 'Let\'s demo the new export at the sprint review.', 'Release notes draft is ready for comments.', 'API rate limit raised to 600/min.'],
    game: ['GG last night!', 'Rematch on Saturday?', 'I\'m in.', 'Bringing my own controller this time.', 'Who won the last round?', 'Kevin, again.', 'We need a rule against Kevin.', 'Snacks are on David this week.', 'Projector or the big TV?', 'Big TV, less lag.', 'Bracket goes up Thursday.', 'Can guests join?', 'Yes, just invite them to the forum.', 'Let\'s go!'],
    choir: ['Song list for Sunday is in Files.', 'Altos, please practice the bridge.', 'Can we start 15 minutes early?', 'Fine by me.', 'Keyboard needs new batteries.', 'I\'ll bring some.', 'Great rehearsal everyone.', 'Recording from Thursday is uploaded.', 'Uniform colour this week is white.', 'Who is leading the opening song?', 'Faith is.', 'Practice the harmonies at home please.', 'Thank you all!', 'See you Sunday.'],
    risk: ['Q3 control testing starts next week.', 'Please submit evidence by Friday.', 'Access review template is in Files.', 'Any exceptions to report?', 'Two, logged in the register.', 'Good, let\'s discuss Monday.', 'Policy acknowledgment is at 82%.', 'Reminder sent to the remaining staff.', 'Audit committee pack is due on the 30th.', 'Draft is ready for review.', 'Incident from last week is closed.', 'Root cause documented.', 'Thanks team.', 'Training session booked for Tuesday.']
  };
  const ago = ['4 h ago', '5 h ago', '6 h ago', 'Yesterday', 'Yesterday', 'Yesterday', '2 days ago', '2 days ago', '3 days ago', '4 days ago', 'Last week'];
  forumsData.forEach((f, fi) => {
    const pool = lines[f.id];
    for (let i = 0; i < 26; i++) {
      const c = chatters[(i + fi) % chatters.length];
      f.updates.push({ who: c.who, initials: c.initials, color: c.color, time: ago[Math.min(ago.length - 1, Math.floor(i / 2.5))], text: pool[i % pool.length], replies: (i * 3 + fi) % 6 });
    }
  });
  const cur = forumsData.find((f) => f.id === s.forum) || forumsData[0];
  const openMode = s.mode === 'full' || s.mode === 'page' ? s.mode : 'docked';
  const forums = forumsData.map((f) => {
    const active = f.id === cur.id && s.mode !== 'closed';
    return {
      ...f,
      hasUnread: f.unread > 0,
      open: () => this.setState({ forum: f.id, panelKind: 'forum', mode: openMode, menu: false, tab: 'discussions', invite: false }),
      rowStyle: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;color:#E8EFEA;background:' + (active ? '#22302A' : 'transparent'),
      chipStyle: 'flex:none;width:36px;height:36px;border-radius:10px;cursor:pointer;font-family:Space Grotesk,sans-serif;font-weight:700;font-size:12px;color:#0F1714;background:' + f.color + ';border:2px solid ' + (f.id === cur.id ? '#E8EFEA' : 'transparent') + ';opacity:' + (f.id === cur.id ? '1' : '0.65')
    };
  });
  const norm = (u) => ({ replies: 0, ...u, hasText: !!u.text, hasFile: !!u.file, fileName: u.file ? u.file.name : '', fileMeta: u.file ? u.file.meta : '' });
  const addedHere = (s.added[cur.id] || []).map(norm);
  const baseMsgs = cur.updates.map(norm);
  const fStyle = (m) => (m.who === 'Sam' || m.who === 'You'
    ? { ...m, who: 'You', mine: true, rowStyle: 'display:flex;gap:10px;flex-direction:row-reverse', bubbleStyle: 'max-width:78%;min-width:0;padding:10px 14px;border-radius:14px 4px 14px 14px;background:#2C3D34' }
    : { ...m, mine: false, rowStyle: 'display:flex;gap:10px', bubbleStyle: 'flex:1;min-width:0;padding:10px 12px;border-radius:4px 12px 12px 12px;background:#1A2520' });
  const forumView = { ...cur, chat: baseMsgs.slice().reverse().concat(addedHere).map(fStyle), topics: addedHere.slice().reverse().concat(baseMsgs), subline: cur.online + ' online · live updates', dotColor: '#B8F36B' };
  // Private messages (share the forum panel)
  const dmStatusCol = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
  const dmPerson = (n) => this.dir.find((d) => d.name === n) || { name: n, initials: n.split(' ').map((x) => x[0]).join(''), color: '#C9B6FF', status: 'Available', title: '' };
  const curDm = s.panelKind === 'dm' ? s.dms.find((d) => d.id === s.dmId) : null;
  const isDM = !!curDm;
  let dmView = null;
  if (curDm) {
    const pr = dmPerson(curDm.with);
    dmView = { id: curDm.id, name: curDm.with, initials: pr.initials, color: pr.color, online: '', about: pr.title, subline: (pr.status || 'Available') + ' · private chat', dotColor: dmStatusCol[pr.status] || '#B8F36B', title: pr.title,
      chat: curDm.msgs.map((m) => { const mine = m.from === 'me'; const who = mine ? 'You' : m.from; const pp = mine ? { initials: 'S', color: '#C2185B' } : dmPerson(m.from);
        return { who: who, initials: pp.initials, color: pp.color, time: m.time, text: m.text || '', hasText: !!m.text, hasFile: !!m.file, fileName: m.file ? m.file.name : '', fileMeta: m.file ? m.file.meta : '', replies: 0, mine: mine,
          rowStyle: 'display:flex;gap:10px;' + (mine ? 'flex-direction:row-reverse' : ''), bubbleStyle: 'max-width:78%;min-width:0;padding:10px 14px;border-radius:' + (mine ? '14px 4px 14px 14px;background:#2C3D34' : '4px 14px 14px 14px;background:#1A2520') }; }),
      topics: [] };
  }
  const curView = isDM ? dmView : forumView;
  this._chatKey = (isDM ? curDm.id : cur.id) + ':' + curView.chat.length;
  const emojiSet = [['😀','grinning'],['😂','laughing'],['😍','heart eyes'],['😎','cool'],['🤔','thinking'],['😢','sad'],['🙏','thank you'],['👍','thumbs up'],['👏','clap'],['🎉','party'],['🔥','fire'],['❤️','heart'],['💯','hundred'],['🚀','rocket'],['✅','done'],['👀','eyes']];
  const emojisForum = emojiSet.map(([ch, label]) => ({ ch, label, add: () => this.setState((st) => ({ fdraft: st.fdraft + ch })) }));
  const emojisWall = emojiSet.map(([ch, label]) => ({ ch, label, add: () => this.setState((st) => ({ draft: st.draft + ch })) }));
  const wfiles = s.wfiles.map((w) => ({ ...w, notImg: !w.isImg, remove: () => this.setState((st) => ({ wfiles: st.wfiles.filter((x) => x.id !== w.id) })) }));
  return { cur, curDm, curView, dmPerson, dmStatusCol, dmView, emojisForum, emojisWall, forums, forumsData, isDM, wfiles };
}
