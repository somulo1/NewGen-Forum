// Initial state: tickets.
export function ticketsState() {
  return {
    tickets: [
      { id: 'IT-208', title: 'Laptop not connecting to office Wi-Fi', cat: 'Network & Wi-Fi', prio: 'High', step: 2, hold: false, raisedBy: 'Faith Achieng', assignees: ['Brian Kiprop'], handler: 'Brian Kiprop', handlerSince: 'Today, 08:40', created: 'Today, 08:10', elapsed: 2.5, sync: 'Jira',
        desc: 'Since this morning my laptop sees the office Wi-Fi but will not connect. Other devices on the same floor work fine. I have restarted twice.',
        activity: [
          { who: 'Faith Achieng', kind: 'event', text: 'Raised the ticket', time: 'Today, 08:10' },
          { who: 'Sam', kind: 'event', text: 'Assigned it to Brian Kiprop', time: 'Today, 08:40' },
          { who: 'Brian Kiprop', kind: 'event', text: 'Started work', time: 'Today, 09:05' },
          { who: 'Brian Kiprop', kind: 'comment', text: 'Looks like an expired network certificate on the laptop. Renewing it now.', time: 'Today, 09:30' }
        ] },
      { id: 'DEV-91', title: 'Export button fails on reports page', cat: 'Software bug', prio: 'Medium', step: 2, hold: false, raisedBy: 'Grace Njeri', assignees: ['David Otieno', 'Amina Wanjiru'], handler: 'Amina Wanjiru', handlerSince: 'Yesterday, 15:00', created: 'Monday, 11:20', elapsed: 23, sync: 'GitHub #312',
        desc: 'Clicking Export on the monthly report shows a spinner and nothing downloads. Happens in Chrome and Edge.',
        activity: [
          { who: 'Grace Njeri', kind: 'event', text: 'Raised the ticket', time: 'Monday, 11:20' },
          { who: 'Sam', kind: 'event', text: 'Assigned it to David Otieno', time: 'Monday, 11:45' },
          { who: 'David Otieno', kind: 'comment', text: 'Reproduced. A timezone offset breaks the date filter before the export runs.', time: 'Monday, 14:10' },
          { who: 'David Otieno', kind: 'file', text: 'export-error.png', meta: '84 KB', time: 'Monday, 14:12' },
          { who: 'David Otieno', kind: 'event', text: 'Handed it over to Amina Wanjiru', time: 'Yesterday, 15:00' }
        ] },
      { id: 'IT-211', title: 'Access request: shared finance folder', cat: 'Access request', prio: 'Medium', step: 1, hold: false, raisedBy: 'Kevin Mutua', assignees: ['Brian Kiprop'], handler: 'Brian Kiprop', handlerSince: 'Today, 07:50', created: 'Yesterday, 15:30', elapsed: 9.5, sync: 'Not synced',
        desc: 'I need read access to the Finance > Reports shared folder for the quarterly sales review.',
        activity: [
          { who: 'Kevin Mutua', kind: 'event', text: 'Raised the ticket', time: 'Yesterday, 15:30' },
          { who: 'Sam', kind: 'event', text: 'Assigned it to Brian Kiprop', time: 'Today, 07:50' }
        ] },
      { id: 'HR-14', title: 'New staff onboarding checklist update', cat: 'HR request', prio: 'Low', step: 3, hold: false, late: false, raisedBy: 'Mercy Wambui', assignees: ['Grace Njeri'], handler: 'Grace Njeri', handlerSince: 'Monday, 09:20', created: 'Monday, 09:00', elapsed: 14, sync: 'Trello',
        desc: 'The onboarding checklist is missing the step for setting up system accounts on day one.',
        activity: [
          { who: 'Mercy Wambui', kind: 'event', text: 'Raised the ticket', time: 'Monday, 09:00' },
          { who: 'Sam', kind: 'event', text: 'Assigned it to Grace Njeri', time: 'Monday, 09:20' },
          { who: 'Grace Njeri', kind: 'event', text: 'Started work', time: 'Monday, 10:00' },
          { who: 'Grace Njeri', kind: 'comment', text: 'Checklist updated with the account setup step. Shared in HR files.', time: 'Yesterday, 16:40' },
          { who: 'Grace Njeri', kind: 'event', text: 'Marked it resolved', time: 'Yesterday, 16:41' }
        ] },
      { id: 'DEV-95', title: 'Add dark mode to mobile app', cat: 'Software bug', prio: 'Low', step: 0, hold: false, raisedBy: 'Sam', assignees: [], handler: '', handlerSince: '', created: 'Today, 10:15', elapsed: 1, sync: 'Jira',
        desc: 'Staff on night shifts asked for a dark theme in the mobile app.',
        activity: [{ who: 'Sam', kind: 'event', text: 'Raised the ticket', time: 'Today, 10:15' }] }
    ], gallery: false, customize: false, acked: {}, rsvp: {}, wished: {}, hvote: null, note: '', forum: 'it', tab: 'discussions', menu: false, invite: false, create: false, link: false, liked: {}, vote: null
  };
}
