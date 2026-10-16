// Initial state: calls.
export function callsState() {
  return {
    call: null, callMode: 'docked', mic: true, cam: true, sharing: false, hand: false, rec: false, callPanel: null, callDraft: '',
    callMsgs: [{ who: 'Grace Njeri', text: 'Joining from the Coast branch. Can everyone hear me?' }, { who: 'Brian Kiprop', text: 'Loud and clear.' }],
    ticketOpen: null, escOpen: false, esc: { type: '', id: '', cat: '', person: '', why: '' }, tcomment: '', tagOpen: false, ntOpen: false, nt: { title: '', cat: 'Network & Wi-Fi', prio: 'Medium', desc: '' },
    notifs: [{ text: 'DEV-91 was handed over to Amina Wanjiru.', color: '#7CC4FF', time: '1 h ago', ticket: 'DEV-91' }],
    notifSeen: false, notifOpen: false, toast: '', toastColor: '#B8F36B',
    settingsFor: null, wcfg: {},
    tset: {
      cats: [
        { name: 'Network & Wi-Fi', tat: 4, unit: 'hours' },
        { name: 'Hardware', tat: 2, unit: 'days' },
        { name: 'Software bug', tat: 3, unit: 'days' },
        { name: 'Access request', tat: 8, unit: 'hours' },
        { name: 'HR request', tat: 5, unit: 'days' }
      ],
      start: '08:00', end: '17:00',
      pause: { weekends: true, holidays: true, offhours: true, onhold: true, awaiting: false },
      notify: { solved: true, overdue: true, nearing: true, assigned: true },
      nearPct: 80, escalate: 'Head of IT Operations', newCat: '', newTat: '', newUnit: 'hours'
    }
  };
}
