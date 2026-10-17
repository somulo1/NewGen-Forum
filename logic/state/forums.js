import { samplePolls } from '~/logic/polls.js';

// Initial state: forum polls, forum chat rules and private-chat privacy.
export function forumsState() {
  return {
    polls: samplePolls(Date.now()), chatFileFolder: null,
    pollForm: null, pollMenu: null, panelTab: 'chat',
    forumChat: { risk: { mode: 'roles', roles: ['Admin', 'Moderator'], blocked: [], allowed: [] } },
    dmPrivacy: {}, privacyOpen: false, incoming: null
  };
}
