// Initial state: forum polls, forum chat rules and private-chat privacy.
export function forumsState() {
  return {
    polls: [], chatFileFolder: null,
    pollForm: null, pollMenu: null, panelTab: 'chat',
    forumChat: {},
    forumForm: { name: '', members: '', err: '', busy: false }, inviteText: '', inviteErr: '', inviteBusy: false,
    dmPrivacy: {}, privacyOpen: false, incoming: null
  };
}
