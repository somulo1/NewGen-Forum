// Initial state: core.
export function coreState() {
  return {
    section: 'home', mode: 'min',   // chats start folded to their launcher: they open when the person opens them
    levelsAdded: {},
    notifTones: {},   // which notification sounds this person wants (see notifTones.js); empty = the defaults
    widgets: ['awards', 'announcements', 'chat', 'calendar', 'tickets', 'projects', 'sessions', 'departments', 'branches', 'performance', 'memos'],
    navOpen: false, panelKind: 'forum', dmId: null, dmsOpen: false, navq: '', newDmOpen: false, dmq: '', forumRead: {},
    secureShares: [], secureSharePasswords: {}, secureShareDraft: { title: '', secret: '', password: '', search: '', recipient: '' },
    hiddenChatMessages: {}, removedChatMessages: {}, messageMenu: null,
    dms: []
  };
}
