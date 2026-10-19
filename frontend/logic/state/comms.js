// Initial state: comms.
export function commsState() {
  return {
    announcementItems: [], cmOpen: false, cmd: { type: 'Memo', title: '', body: '', wall: false, forum: '', sel: {} },
    comms: [
      { id: 'C-210', type: 'Announcement', title: 'Updated leave policy takes effect 1 November', body: 'Please read the new leave policy in HR files.', by: 'Grace Njeri', date: '2026-10-13', to: ['Community wall'], status: 'Published', read: 78 },
      { id: 'C-209', type: 'Memo', title: 'Year-end audit timelines', body: 'Evidence is due by 14 November.', by: 'Amina Wanjiru', date: '2026-10-12', to: ['Risk & Compliance', 'Finance'], status: 'Published', read: 91 },
      { id: 'C-208', type: 'Newsletter', title: 'Hub Monthly · October', body: 'New joiners, the charity run and what\'s next.', by: 'Kevin Mutua', date: '2026-10-01', to: ['Community wall'], status: 'Published', read: 64 },
      { id: 'C-207', type: 'Memo', title: 'Branch opening hours over the holiday', body: 'Branches open 9 to 1 on the public holiday.', by: 'Samwel Okoth Omulo', date: '2026-10-10', to: ['North Branch', 'Coast Branch'], status: 'Published', read: 88 },
      { id: 'C-206', type: 'Announcement', title: 'Office closed Friday for maintenance', body: 'The head office is closed this Friday.', by: 'Faith Achieng', date: '2026-10-15', to: ['Community wall', 'IT Operations forum'], status: 'Awaiting approval', read: 0 },
      { id: 'C-205', type: 'Memo', title: 'New password rules from Monday', body: 'Passwords now need 12 characters.', by: 'Samwel Okoth Omulo', date: '2026-10-06', to: ['Community wall'], status: 'Published', read: 96 },
      { id: 'C-204', type: 'Newsletter', title: 'Hub Monthly · September', body: 'Q3 highlights.', by: 'Kevin Mutua', date: '2026-09-01', to: ['Community wall'], status: 'Published', read: 71 }
    ]
  };
}
