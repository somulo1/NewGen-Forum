// Initial state: wall.
export function wallState() {
  return {
    pw: 380, dragging: false, draft: '', fdraft: '', added: {}, emojiF: false, emojiW: false, wfiles: [], shareFor: null, shareNote: {}, shareCount: {},
    posts: [
      { id: 1, author: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', meta: 'Finance · 12 min ago', text: 'Huge thanks to everyone who joined the charity run on Saturday. We raised [AMOUNT] for the children\'s home!', likes: 48, extra: 9, shares: 3, photoLabel: 'Charity run group photo', photoBg: '#2E5B4A',
        comments: [{ who: 'Brian K.', initials: 'BK', color: '#7CC4FF', text: 'What a morning! Same time next year.' }, { who: 'Grace N.', initials: 'GN', color: '#FFB86B', text: 'Proud of this team.' }] },
      { id: 2, author: 'David Otieno', initials: 'DO', color: '#F49AC1', meta: 'Weekend Gamers · 1 h ago', text: 'Which game for this month\'s tournament?', likes: 21, extra: 28, shares: 1, hasPoll: true,
        comments: [{ who: 'Kevin M.', initials: 'KM', color: '#C9B6FF', text: 'FIFA or nothing.' }] },
      { id: 3, author: 'Grace Njeri', initials: 'GN', color: '#FFB86B', meta: 'HR · Yesterday', text: 'Welcome to our 6 new team members joining this week. Say hi below and help them settle in.', likes: 73, extra: 24, shares: 5,
        comments: [{ who: 'Faith A.', initials: 'FA', color: '#B8F36B', text: 'Karibu everyone!' }] }
    ]
  };
}
