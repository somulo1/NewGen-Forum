// Sample attachments so every place that shows files can be seen with real-looking content.
// The files live in /public/samples. Replace with real uploads once there is a backend.
const IMG = (n, name) => ({ name, meta: '240 KB', url: '/samples/photo-' + n + '.svg', isImg: true });
export const sampleImages = () => [IMG(1, 'Office launch.svg'), IMG(2, 'Team offsite.svg'), IMG(3, 'Charity run.svg')];
export const sampleVideo = () => ({ name: 'Demo clip.webm', meta: '0:03 · 89 KB', url: '/samples/Demo-clip.webm', isVideo: true });
export const sampleDocs = () => [{ name: 'Policy handbook.pdf', meta: '400 B', url: '/samples/Policy-handbook.pdf' }, { name: 'Q3 board report.pdf', meta: '400 B', url: '/samples/Q3-board-report.pdf' }];
export const sampleFiles = () => [{ name: 'Brand assets.zip', meta: '132 B', url: '/samples/Brand-assets.zip' }, { name: 'Budget tracker.csv', meta: '40 B', url: '/samples/Budget-tracker.csv' }, { name: 'Meeting notes.txt', meta: '22 B', url: '/samples/Meeting-notes.txt' }];

// Four messages for a forum, newest first (the forum list is stored newest first).
export const forumSamples = () => [
  { who: 'Kevin M.', initials: 'KM', color: '#C9B6FF', time: 'Yesterday', text: 'Brand assets, the budget tracker and last meeting\'s notes.', file: sampleFiles(), replies: 0 },
  { who: 'Faith A.', initials: 'FA', color: '#B8F36B', time: '2 days ago', text: 'Policy handbook and the Q3 board report, for review.', file: sampleDocs(), replies: 2 },
  { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '3 days ago', text: 'Quick demo clip from the dry run.', file: sampleVideo(), replies: 1 },
  { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: 'Last week', text: 'Photos from the last event.', file: sampleImages(), replies: 4 }
];

// Four messages for a private chat, oldest first.
export const dmSamples = (name) => [
  { from: name, text: 'Photos from the last event.', file: sampleImages(), time: 'Mon 9:12 AM' },
  { from: 'me', text: 'Demo clip from the dry run.', file: sampleVideo(), time: 'Mon 9:40 AM' },
  { from: name, text: 'Policy handbook and the Q3 board report.', file: sampleDocs(), time: 'Tue 11:05 AM' },
  { from: 'me', text: 'Brand assets, budget tracker and notes.', file: sampleFiles(), time: 'Tue 4:30 PM' }
];

// Four wall posts with media.
export const wallSamples = () => [
  { id: 101, author: 'Kevin Mutua', initials: 'KM', color: '#C9B6FF', meta: 'Everyone · Last week', text: 'Shared the brand assets, budget tracker and meeting notes.', likes: 9, extra: 1, shares: 0, comments: [], media: sampleFiles().map((m, i) => ({ id: 'sf' + i, ...m })) },
  { id: 102, author: 'Faith Achieng', initials: 'FA', color: '#B8F36B', meta: 'Everyone · 5 days ago', text: 'Policy handbook and the Q3 board report are out for review.', likes: 14, extra: 3, shares: 2, comments: [], media: sampleDocs().map((m, i) => ({ id: 'sd' + i, ...m })) },
  { id: 103, author: 'Grace Njeri', initials: 'GN', color: '#FFB86B', meta: 'Everyone · 4 days ago', text: 'A short clip from the dry run.', likes: 31, extra: 6, shares: 4, comments: [], media: [{ id: 'sv0', ...sampleVideo() }] },
  { id: 104, author: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', meta: 'Everyone · 3 days ago', text: 'Photos from the last event. What a day!', likes: 52, extra: 11, shares: 3, comments: [], media: sampleImages().map((m, i) => ({ id: 'si' + i, ...m })) }
];
