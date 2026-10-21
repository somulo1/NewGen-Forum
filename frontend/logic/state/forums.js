// Initial state: forum polls, forum chat rules and private-chat privacy.
export function forumsState() {
  return {
    polls: [], chatFileFolder: null,
    pollForm: null, pollMenu: null, panelTab: 'chat',
    forumChat: {},
    forumForm: { name: '', members: '', picked: [], err: '', busy: false }, addOpen: false, addErr: '', addBusy: false, inviteText: '', inviteErr: '', inviteBusy: false,
    dmPrivacy: {}, privacyOpen: false, incoming: null
  };
}
