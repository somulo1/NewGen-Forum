import * as dmsW from '~/logic/server/widgets/dms.js';
import { richTextHtml, mentionLinker } from '~/logic/richText.js';
import { noteText, notesConnected, isMine, permLabel } from '~/logic/noteText.js';
import * as forumsW from '~/logic/server/widgets/forums.js';
// renderVals step: forums. Receives the names built by earlier steps and returns the ones later steps need.
export function buildForums(ctx) {
  const { s } = ctx;
  const sampleForums = [];   // forums come from the server
  // Connected to the server: the forums you are in. Not connected (no sign-in to an organization): the built-in samples.
  const live = s.forumsLive || (this.srv ? [] : null);
  const forumsData = live ? live.map((f) => ({ ...f, updates: f.updates.slice() })) : sampleForums;
  const loadingForums = !!this.srv && !s.forumsLive;   // signed in, and the forums have not arrived yet
  const NO_FORUM = { id: '', name: loadingForums ? 'Loading your forums…' : 'No forums yet', initials: '', color: '#9DB0A6', unread: 0, members: 0, online: 0, privacy: '', about: loadingForums ? 'Your forums are loading.' : 'You are not in any forum yet.', updates: [], memberList: [] };
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
  // a tagged person (@Name) is a link to their profile
  const linkMentions = mentionLinker(this.dir.map((d) => d.name).concat(this.srv ? [this.srv.me.name] : []));
  const norm = (u, id) => ({ replies: 0, ...u, id: u.id || id, source: u, htmlText: linkMentions(richTextHtml(u.text)), hasText: !!u.text, hasFile: !!u.file, fileList: [].concat(u.file || []), ...quote(u), hasShared: !!u.shared, sid: u.shared ? u.shared.sid : '', reply: replyFn(u.who === 'Sam' ? 'You' : u.who, u) });
  const addedSource = s.added[cur.id] || [];
  const addedHere = addedSource.map((message) => norm(message, 'forum-' + cur.id + '-added-' + fingerprint(message)));
  const baseMsgs = cur.updates.map((message, index) => norm(message, 'forum-' + cur.id + '-sample-' + index));
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
    const dmMessages = curDm.msgs.map((m, index) => {
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
