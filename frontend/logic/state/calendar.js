// Initial state: calendar.
export function calendarState() {
  return {
    calMonth: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }).slice(0, 7), calSel: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), calHide: {},
    evOpen: false, ev: { title: '', type: 'Meeting', date: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), start: '10:00', end: '11:00', repeat: 'None', forum: '', people: [], pq: '' },
    events: [
      { id: 'e1', title: 'Sprint review', type: 'Meeting', date: '2026-10-15', start: '15:00', end: '16:00', people: ['David Otieno', 'Amina Wanjiru', 'Brian Kiprop', 'Kevin Mutua'], forum: 'Product Builders', repeat: 'None', by: 'David Otieno' },
      { id: 'e2', title: 'IT daily standup', type: 'Standup', date: '2026-10-13', start: '09:15', end: '09:30', people: ['Brian Kiprop', 'David Otieno', 'John Kamau'], forum: 'IT Operations', repeat: 'Weekdays', by: 'Brian Kiprop' },
      { id: 'e3', title: 'Access review with department heads', type: 'Meeting', date: '2026-10-17', start: '11:00', end: '12:00', people: ['Grace Njeri', 'Amina Wanjiru', 'Kevin Mutua'], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' },
      { id: 'e4', title: 'Draft the Q4 risk report', type: 'Task', date: '2026-10-20', start: '', end: '', people: [], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' },
      { id: 'e5', title: 'Team building', type: 'Event', date: '2026-10-18', start: '08:00', end: '17:00', people: [], forum: '', repeat: 'None', by: 'Grace Njeri' },
      { id: 'e6', title: 'Quarterly town hall', type: 'Meeting', date: '2026-10-22', start: '10:00', end: '11:30', people: ['Brian Kiprop', 'Grace Njeri', 'David Otieno', 'Amina Wanjiru', 'Faith Achieng', 'Kevin Mutua'], forum: '', repeat: 'None', by: 'Grace Njeri' },
      { id: 'e7', title: 'Risk & Compliance weekly sync', type: 'Standup', date: '2026-10-14', start: '08:30', end: '08:45', people: ['Amina Wanjiru', 'Kevin Mutua'], forum: 'Risk & Compliance', repeat: 'Weekly', by: 'Samwel Okoth Omulo' },
      { id: 'e8', title: 'Submit SIEM rollout plan', type: 'Task', date: '2026-10-16', start: '17:00', end: '', people: [], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' }
    ]
  };
}
