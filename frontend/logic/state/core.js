// Initial state: core.
export function coreState() {
  return {
    section: 'home', mode: 'docked',
    widgets: ['awards', 'announcements', 'chat', 'calendar', 'tickets', 'projects', 'sessions', 'departments', 'branches', 'performance', 'memos'],
    navOpen: false, panelKind: 'forum', dmId: null, dmsOpen: false, navq: '', newDmOpen: false, dmq: '', forumRead: {},
    secureShares: [], secureSharePasswords: {}, secureShareDraft: { title: '', secret: '', password: '', search: '', recipient: '' },
    hiddenChatMessages: {}, removedChatMessages: {}, messageMenu: null,
    dms: []
  };
}
