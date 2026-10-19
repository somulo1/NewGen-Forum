import { wallSamples } from '~/logic/samples.js';

// Initial state: wall.
export function wallState() {
  return {
    pw: 380, dragging: false, draft: '', fdraft: '', added: {}, emojiF: false, emojiW: false, wfiles: [], wallAudience: 'Everyone', wallTag: '', hiddenPosts: {}, savedPosts: {}, moments: [
      { id: 'm0', owner: 'Sam', initials: 'S', color: '#F49AC1', bg: '#4A3A6B', caption: 'Late shift at the office, quiet and productive', media: { url: '/samples/photo-4.svg', isImg: true, name: 'Product demo day.svg' }, likes: ['Brian Kiprop', 'Grace Njeri', 'David Otieno'], loves: ['Amina Wanjiru', 'Faith Achieng'] },
      { id: 'm1', owner: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', bg: '#2B4A63', caption: 'Brian at the server room', likes: ['Grace Njeri'], loves: [] },
      { id: 'm2', owner: 'Grace Njeri', initials: 'GN', color: '#FFB86B', bg: '#6B4A2B', caption: 'Team lunch', media: { url: '/samples/photo-2.svg', isImg: true, name: 'Team offsite.svg' }, likes: ['David Otieno', 'Faith Achieng'], loves: ['Brian Kiprop'] },
      { id: 'm3', owner: 'David Otieno', initials: 'DO', color: '#F49AC1', bg: '#5A2E47', caption: 'Game night setup', media: { url: '/samples/Demo-clip.webm', isVideo: true, name: 'Demo clip.webm' }, likes: [], loves: ['Kevin Mutua'] },
      { id: 'm4', owner: 'Faith Achieng', initials: 'FA', color: '#B8F36B', bg: '#3B5A2B', caption: 'First day!', likes: ['Amina Wanjiru', 'Brian Kiprop'], loves: [] }
    ], momentOpen: null, reactorsOpen: false, addMomentOpen: false, postMenu: null, wpoll: null, pollVotes: {}, replyTo: null, sharedComments: {}, commentsOpen: {}, replyFor: null, shareFor: null, shareNote: {}, shareCount: {},
    posts: [
      { id: 1, author: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', meta: 'Finance · 12 min ago', text: 'Huge thanks to everyone who joined the charity run on Saturday. We raised [AMOUNT] for the children\'s home!', likes: 48, extra: 9, shares: 3, photoLabel: 'Charity run group photo', photoBg: '#2E5B4A',
        comments: [{ who: 'Brian K.', initials: 'BK', color: '#7CC4FF', text: 'What a morning! Same time next year.' }, { who: 'Grace N.', initials: 'GN', color: '#FFB86B', text: 'Proud of this team.' }] },
      { id: 2, author: 'David Otieno', initials: 'DO', color: '#F49AC1', meta: 'Weekend Gamers · 1 h ago', text: 'Which game for this month\'s tournament?', likes: 21, extra: 28, shares: 1, hasPoll: true,
        comments: [{ who: 'Kevin M.', initials: 'KM', color: '#C9B6FF', text: 'FIFA or nothing.' }] },
      { id: 3, author: 'Grace Njeri', initials: 'GN', color: '#FFB86B', meta: 'HR · Yesterday', text: 'Welcome to our 6 new team members joining this week. Say hi below and help them settle in.', likes: 73, extra: 24, shares: 5,
        comments: [{ who: 'Faith A.', initials: 'FA', color: '#B8F36B', text: 'Karibu everyone!' }] }
    ].concat(wallSamples())
  };
}
