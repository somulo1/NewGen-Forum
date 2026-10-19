import * as dmsW from '~/logic/server/widgets/dms.js';
import { dmSamples, forumSamples } from '~/logic/samples.js';
import { richTextHtml } from '~/logic/richText.js';
import { noteText, notesConnected, isMine, permLabel } from '~/logic/noteText.js';
import * as forumsW from '~/logic/server/widgets/forums.js';
// renderVals step: forums. Receives the names built by earlier steps and returns the ones later steps need.
export function buildForums(ctx) {
  const { s } = ctx;
  const sampleForums = [
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
  sampleForums.forEach((f, fi) => {
    const pool = lines[f.id];
    for (let i = 0; i < 26; i++) {
      const c = chatters[(i + fi) % chatters.length];
      f.updates.push({ who: c.who, initials: c.initials, color: c.color, time: ago[Math.min(ago.length - 1, Math.floor(i / 2.5))], text: pool[i % pool.length], replies: (i * 3 + fi) % 6 });
    }
  });
  // Connected to the server: the forums you are in. Not connected (no sign-in to an organization): the built-in samples.
  const live = s.forumsLive || (this.srv ? [] : null);
  const forumsData = live ? live.map((f) => ({ ...f, updates: f.updates.slice() })) : sampleForums;
  const NO_FORUM = { id: '', name: 'No forums yet', initials: '', color: '#9DB0A6', unread: 0, members: 0, online: 0, privacy: '', about: 'You are not in any forum yet.', updates: [], memberList: [] };
  const cur = forumsData.find((f) => f.id === s.forum) || forumsData[0] || NO_FORUM;
  const openMode = this.isMobileViewport() ? 'page' : s.mode === 'full' || s.mode === 'page' ? s.mode : 'docked';
  const forums = forumsData.map((f) => {
    const active = f.id === cur.id && s.mode !== 'closed';
    const unread = f.live ? f.unread : (s.forumRead && s.forumRead[f.id] ? 0 : f.unread);
    return {
      ...f,
      unread,
      hasUnread: unread > 0,
      open: () => { this.setState((st) => ({ forumRead: { ...st.forumRead, [f.id]: true }, forum: f.id, panelKind: 'forum', mode: openMode, menu: false, tab: 'discussions', invite: false })); if (f.live) forumsW.markRead(this, this.srv, f.id); },
      rowStyle: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;color:var(--k-E8EFEA,#E8EFEA);background:' + (active ? 'var(--k-22302A,#22302A)' : 'transparent'),
      chipStyle: 'flex:none;width:36px;height:36px;border-radius:10px;cursor:pointer;font-family:Space Grotesk,sans-serif;font-weight:700;font-size:12px;color:#0F1714;background:' + f.color + ';border:2px solid ' + (f.id === cur.id ? '#E8EFEA' : 'transparent') + ';opacity:' + (f.id === cur.id ? '1' : '0.65'),
      unreadStyle: 'position:absolute;top:-5px;right:-6px;min-width:17px;height:17px;padding:0 4px;box-sizing:border-box;border-radius:9px;background:#B8F36B;color:#0F1714;font:700 10px sans-serif;display:flex;align-items:center;justify-content:center;pointer-events:none'
    };
  });
  const snip = (m) => { const x = m.text || (m.file ? [].concat(m.file)[0].name : ''); return x.length > 90 ? x.slice(0, 90) + '…' : x; };
  const dmNow = s.panelKind === 'dm' ? s.dms.find((d) => d.id === s.dmId) : null;
  const replyKey = dmNow ? dmNow.id : cur.id;
  this._replyKey = replyKey;
  const replyFn = (who, m) => () => this.setState({ replyTo: { key: replyKey, who, snippet: snip(m), mid: m.mid || '' } });
  const quote = (m) => ({ hasReply: !!m.replyTo, replyWho: m.replyTo ? m.replyTo.who : '', replySnippet: m.replyTo ? m.replyTo.snippet : '' });
  const fingerprint = (message) => encodeURIComponent([message.who || '', message.time || '', message.text || '', [].concat(message.file || []).map((file) => file.name || '').join(',')].join('|'));
  const norm = (u, id) => ({ replies: 0, ...u, id: u.id || id, source: u, htmlText: richTextHtml(u.text), hasText: !!u.text, hasFile: !!u.file, fileList: [].concat(u.file || []), ...quote(u), hasShared: !!u.shared, sid: u.shared ? u.shared.sid : '', reply: replyFn(u.who === 'Sam' ? 'You' : u.who, u) });
  const addedSource = s.added[cur.id] || [];
  const addedHere = addedSource.map((message) => norm(message, 'forum-' + cur.id + '-added-' + fingerprint(message)));
  const baseMsgs = (live ? cur.updates : cur.updates.concat(forumSamples())).map((message, index) => norm(message, 'forum-' + cur.id + '-sample-' + index));
  const messageActions = (message, key, deleteForEveryone) => ({
    messageKey: key,
    menuOpen: s.messageMenu === key,
    startHold: (event) => {
      if (event.pointerType === 'mouse' && event.button !== 0) return;
      clearTimeout(this._messageHoldTimer);
      this._messageHoldTimer = setTimeout(() => this.setState({ messageMenu: key }), 550);
    },
    stopHold: () => clearTimeout(this._messageHoldTimer),
    openActions: (event) => { event.preventDefault(); this.setState({ messageMenu: key }); },
    closeActions: () => this.setState({ messageMenu: null }),
    deleteForMe: () => this.setState((st) => ({ hiddenChatMessages: { ...st.hiddenChatMessages, [key]: true }, messageMenu: null })),
    deleteForEveryone: deleteForEveryone ? () => { deleteForEveryone(); this.setState({ messageMenu: null }); } : null
  });
  const fStyle = (m) => (m.who === 'Sam' || m.who === 'You'
    ? { ...m, who: 'You', mine: true, rowStyle: 'display:flex;gap:10px;flex-direction:row-reverse', bubbleStyle: 'max-width:78%;min-width:0;padding:10px 14px;border-radius:14px 4px 14px 14px;background:var(--k-2C3D34,#2C3D34)' }
    : { ...m, mine: false, rowStyle: 'display:flex;gap:10px', bubbleStyle: 'flex:1;min-width:0;padding:10px 12px;border-radius:4px 12px 12px 12px;background:var(--k-1A2520,#1A2520)' });
  const forumMessages = baseMsgs.slice().reverse().concat(addedHere).map((message) => {
    const key = 'forum:' + cur.id + ':' + message.id;
    const source = message.source;
    const canDeleteForEveryone = (message.who === 'Sam' || message.who === 'You') && (addedSource.includes(source) || !!source.mid);
    const actions = messageActions(message, key, canDeleteForEveryone ? () => (source.mid ? forumsW.remove(this, this.srv, cur.id, source.mid).catch((e) => this.notify(e.message || 'Could not delete the message.', '#FF9B8A')) : this.setState((st) => ({ added: { ...st.added, [cur.id]: (st.added[cur.id] || []).filter((item) => item !== source) } }))) : null);
    return { ...fStyle(message), ...actions };
  }).filter((message) => !s.hiddenChatMessages[message.messageKey]);
  const forumView = { ...cur, chat: forumMessages, topics: forumMessages, subline: cur.online + ' online · live updates', dotColor: '#B8F36B' };
  // Private messages (share the forum panel)
  const dmStatusCol = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
  const dmPerson = (n) => this.dir.find((d) => d.name === n) || { name: n, initials: n.split(' ').map((x) => x[0]).join(''), color: '#C9B6FF', status: 'Available', title: '' };
  const curDm = s.panelKind === 'dm' ? s.dms.find((d) => d.id === s.dmId) : null;
  const isDM = !!curDm;
  let dmView = null;
  if (curDm) {
    const pr = dmPerson(curDm.with);
    const secureShares = s.secureShares.filter((share) => share.recipient === curDm.with).map((share) => {
      const now = Date.now();
      const expired = !share.openedAt && now >= share.expiresAt;
      const showingSecret = !!share.openedAt && !share.consumed && now < share.revealUntil;
      return {
        ...share, expired, showingSecret, unavailable: expired || !!share.openedAt, passwordValue: s.secureSharePasswords[share.id] || '',
        onPassword: (e) => this.setState((st) => ({ secureSharePasswords: { ...st.secureSharePasswords, [share.id]: e.target.value } })),
        unlock: () => {
          const current = this.state.secureShares.find((item) => item.id === share.id);
          if (!current || current.consumed || current.openedAt || Date.now() >= current.expiresAt) {
            this.notify('This secure share has expired or has already been opened.', '#FFB86B');
            return;
          }
          if ((this.state.secureSharePasswords[share.id] || '') !== current.password) {
            this.notify('That link password is not correct.', '#FF9B8A');
            return;
          }
          const revealUntil = Date.now() + 10000;
          this.setState((st) => ({ secureShares: st.secureShares.map((item) => item.id === share.id ? { ...item, openedAt: Date.now(), revealUntil } : item), secureSharePasswords: { ...st.secureSharePasswords, [share.id]: '' } }));
          this.notify('Secure share opened. Its contents will disappear in 10 seconds.', '#B8F36B');
          setTimeout(() => this.setState((st) => ({ secureShares: st.secureShares.map((item) => item.id === share.id ? { ...item, consumed: true, revealUntil: 0 } : item) })), 10000);
        }
      };
    });
    const sharesById = Object.fromEntries(secureShares.map((share) => [share.id, share]));
    const dmMessages = (this.srv ? [] : dmSamples(curDm.with)).concat(curDm.msgs).map((m, index) => {
      const mine = m.from === 'me';
      const who = mine ? 'You' : m.from;
      const pp = mine ? { initials: 'S', color: '#C2185B' } : dmPerson(m.from);
      const key = 'dm:' + curDm.id + ':' + (m.id || 'legacy-' + fingerprint({ who: m.from, ...m }));
      const actions = messageActions(m, key, mine ? () => {
        if (m.mid && this.srv) { dmsW.remove(this, this.srv, curDm.id, m.mid).catch((e) => this.notify(e.message || 'Could not delete it.', '#FF9B8A')); return; }
        if (curDm.msgs.includes(m)) {
          this.setState((st) => ({
            dms: st.dms.map((dm) => dm.id === curDm.id ? { ...dm, msgs: dm.msgs.filter((item) => item !== m) } : dm),
            secureShares: m.secureShareId ? st.secureShares.filter((share) => share.id !== m.secureShareId) : st.secureShares
          }));
        } else {
          this.setState((st) => ({ removedChatMessages: { ...st.removedChatMessages, [key]: true } }));
        }
      } : null);
      return {
        id: key, who, initials: pp.initials, color: pp.color, time: m.time, text: m.text || '', htmlText: richTextHtml(m.text),
        hasText: !!m.text, hasFile: !!m.file, fileList: [].concat(m.file || []), ...quote(m), reply: replyFn(who, m),
        replies: 0, mine, secureShare: sharesById[m.secureShareId] || null, hasSecureShare: !!sharesById[m.secureShareId],
        rowStyle: 'display:flex;gap:10px;' + (mine ? 'flex-direction:row-reverse' : ''),
        bubbleStyle: 'max-width:78%;min-width:0;padding:10px 14px;border-radius:' + (mine ? '14px 4px 14px 14px;background:var(--k-2C3D34,#2C3D34)' : '4px 14px 14px 14px;background:var(--k-1A2520,#1A2520)'),
        ...actions
      };
    }).filter((message) => !s.hiddenChatMessages[message.messageKey] && !s.removedChatMessages[message.messageKey]);
    dmView = { id: curDm.id, name: curDm.with, initials: pr.initials, color: pr.color, online: '', about: pr.title, subline: (pr.status || 'Available') + ' · private chat', dotColor: dmStatusCol[pr.status] || '#B8F36B', title: pr.title,
      chat: dmMessages, topics: [] };
  }
  const curView = isDM ? dmView : forumView;
  const fileKindColor = { PDF: '#FF9B8A', DOC: '#7CC4FF', DOCX: '#7CC4FF', XLS: '#B8F36B', XLSX: '#B8F36B', PPT: '#FFB86B', PPTX: '#FFB86B', ZIP: '#C9B6FF', RAR: '#C9B6FF', '7Z': '#C9B6FF', IMG: '#F49AC1', VIDEO: '#6EE7C8', NOTE: '#FFD66B' };
  const fileFolders = curView.chat.flatMap((message, messageIndex) => {
    if (!message.fileList.length) return [];
    const attachedFiles = message.fileList.map((file) => {
      if (file.noteId) {
        // A shared note: a live link. It stops opening if the share is stopped, the note is deleted or the Notes widget is removed.
        const note = s.notes.find((x) => x.container === file.noteId) || s.notes.find((x) => x.id === file.noteId);
        const share = note && isMine(note) ? (note.shares || []).find((x) => x.id === file.shareId) : null;
        const live = !!note && notesConnected(s) && (isMine(note) ? !!share : true);
        const perm = note ? (isMine(note) ? (share && share.perm) : note.access) : 'view';
        return { ...file, kind: 'NOTE', previewKind: 'note', color: '#FFD66B', isNote: true, live, noteLive: live, notePreview: live ? noteText(note.body).split('\n').filter((l) => l.trim()).slice(0, 6).join('\n') : '',
          meta: live ? 'Note · ' + permLabel(perm) + ' · live link' : 'Note · access removed', noteOpen: () => this.openNote(note.id) };
      }
      const extension = (file.name || '').split('.').pop().toUpperCase();
      const kind = file.isImg ? 'IMG' : file.isVideo ? 'VIDEO' : ({ PNG: 'IMG', JPG: 'IMG', JPEG: 'IMG', GIF: 'IMG', SVG: 'IMG', WEBP: 'IMG', MP4: 'VIDEO', WEBM: 'VIDEO', MOV: 'VIDEO' })[extension] || extension || 'FILE';
      const previewKind = kind === 'IMG' ? 'image' : kind === 'VIDEO' ? 'video' : kind === 'PDF' ? 'pdf' : ['XLS', 'XLSX', 'CSV'].includes(kind) ? 'sheet' : ['PPT', 'PPTX'].includes(kind) ? 'slides' : ['ZIP', 'RAR', '7Z'].includes(kind) ? 'archive' : ['DOC', 'DOCX', 'TXT'].includes(kind) ? 'document' : 'file';
      return { ...file, kind, previewKind, color: fileKindColor[kind] || '#9DB0A6' };
    });
    return [{
      id: curView.id + ':' + messageIndex,
      name: message.text || 'Shared files',
      by: message.who,
      time: message.time,
      files: attachedFiles,
      count: attachedFiles.length,
      samples: attachedFiles.slice(0, 3).map((file, index) => ({ kind: file.kind, color: file.color, height: [88, 76, 82][index] }))
    }];
  });
  const selectedChatFolder = fileFolders.find((folder) => folder.id === s.chatFileFolder) || null;
  this._chatKey = (isDM ? curDm.id : cur.id) + ':' + curView.chat.length;
  const wfiles = s.wfiles.map((w) => ({ ...w, notImg: !w.isImg, remove: () => this.setState((st) => ({ wfiles: st.wfiles.filter((x) => x.id !== w.id) })) }));
  return { cur, curDm, curView, dmPerson, dmStatusCol, dmView, fileFolders, forums, forumsData, isDM, selectedChatFolder, wfiles };
}
