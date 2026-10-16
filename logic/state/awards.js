// Initial state: awards.
export function awardsState() {
  return {
    awOpen: false,
    awd: { type: 'Employee of the month', people: [], role: 'all', q: '', citation: '', poster: '', wall: true, forums: [] },
    awardList: [
      { id: 'AW-12', type: 'Employee of the month', people: ['Brian Kiprop'], citation: '99.9% uptime across all branches in October.', date: '2026-10-01', status: 'Approved' },
      { id: 'AW-11', type: 'Customer hero', people: ['Faith Achieng'], citation: 'Highest customer satisfaction score this quarter.', date: '2026-09-30', status: 'Approved' },
      { id: 'AW-10', type: 'Star performer · Q3', people: ['Amina Wanjiru', 'David Otieno', 'Grace Njeri'], citation: 'Delivered the Q3 close and the new HR onboarding ahead of time.', date: '2026-09-28', status: 'Approved' },
      { id: 'AW-09', type: 'Innovation award', people: ['Kevin Mutua'], citation: 'Built the new sales pipeline tracker.', date: '2026-09-15', status: 'Approved' },
      { id: 'AW-08', type: 'Team player', people: ['John Kamau', 'Lilian Chebet'], citation: 'Always first to help other branches.', date: '2026-09-01', status: 'Approved' },
      { id: 'AW-07', type: 'Security champion', people: ['Samwel Okoth Omulo', 'Brian Kiprop'], citation: 'Zero critical findings in the Q2 audit.', date: '2026-07-10', status: 'Approved' }
    ]
  };
}
