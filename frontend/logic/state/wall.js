// Initial state: wall.
export function wallState() {
  return {
    pw: 380, dragging: false, draft: '', fdraft: '', added: {}, emojiF: false, emojiW: false, wfiles: [], wallAudience: 'Everyone', wallTag: '', hiddenPosts: {}, savedPosts: {}, moments: [], momentOpen: null, reactorsOpen: false, addMomentOpen: false, postMenu: null, wpoll: null, pollVotes: {}, replyTo: null, sharedComments: {}, commentsOpen: {}, replyFor: null, shareFor: null, shareNote: {}, shareCount: {},
    posts: []
  };
}
