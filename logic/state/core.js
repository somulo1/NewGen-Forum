// Initial state: core.
export function coreState() {
  return {
    section: 'home', mode: 'docked',
    widgets: ['awards', 'announcements', 'chat', 'calendar', 'tickets', 'projects', 'sessions', 'departments', 'branches', 'performance', 'memos'],
    navOpen: false, panelKind: 'forum', dmId: null, dmsOpen: false, newDmOpen: false, dmq: '',
    dms: [
      { id: 'dm1', with: 'Brian Kiprop', unread: 2, msgs: [{ from: 'Brian Kiprop', text: 'Hi Sam, the VPN certificate issue is fixed.', time: '2:10 PM' }, { from: 'me', text: 'Great, thanks!', time: '2:12 PM' }, { from: 'Brian Kiprop', text: 'Can we do a quick call about the SIEM rollout?', time: '3:31 PM' }, { from: 'Brian Kiprop', text: 'I have 15 minutes before 4.', time: '3:32 PM' }] },
      { id: 'dm2', with: 'Amina Wanjiru', unread: 1, msgs: [{ from: 'Amina Wanjiru', text: 'Do you have the audit evidence list?', time: '1:05 PM' }] },
      { id: 'dm3', with: 'Grace Njeri', unread: 0, msgs: [{ from: 'Grace Njeri', text: 'The access review sign-offs are in the HR folder.', time: '11:20 AM' }, { from: 'me', text: 'Perfect, I\'ll check.', time: '11:24 AM' }] },
      { id: 'dm4', with: 'David Otieno', unread: 0, msgs: [{ from: 'me', text: 'Is DEV-95 on your radar?', time: 'Yesterday' }, { from: 'David Otieno', text: 'Yes, picking it up tomorrow.', time: 'Yesterday' }] }
    ]
  };
}
