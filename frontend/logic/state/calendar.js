// Initial state: calendar.
export function calendarState() {
  return {
    calMonth: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }).slice(0, 7), calSel: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), calHide: {},
    evOpen: false, ev: { title: '', type: 'Meeting', date: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), start: '10:00', end: '11:00', repeat: 'None', forum: '', people: [], pq: '' },
    events: []
  };
}
