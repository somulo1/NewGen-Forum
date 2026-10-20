// Initial state: calls.
export function callsState() {
  return {
    callLog: [],
    callOff: {}, ouItem: null, ticketFrom: null, secRange: { preset: 'all', from: '', to: '' }, alertNotes: {}, alertDlg: null, ouRecs: {}, ouThr: null,
    call: null, callMode: 'docked', mic: true, cam: true, sharing: false, hand: false, rec: false, callPanel: null, callDraft: '', callFind: '', callLocked: false,
    callMsgs: [],
    ticketOpen: null, escOpen: false, esc: { type: '', id: '', cat: '', person: '', why: '' }, tcomment: '', tagOpen: false, ntOpen: false, nt: { title: '', cat: '', prio: 'Medium', desc: '', files: [], toType: '', toName: '' }, unitCats: {}, catEdit: null, catWork: null, catDraft: { name: '', tat: '4', prio: 'Medium', err: '' },
    notifs: [],
    notifSeen: false, notifGone: [], flowRole: {}, notifOpen: false, toast: '', toastColor: '#B8F36B',
    settingsFor: null, wcfg: {},
    tset: {
      cats: [],
      start: '08:00', end: '17:00',
      pause: { weekends: true, holidays: true, offhours: true, onhold: true, awaiting: false },
      notify: { solved: true, overdue: true, nearing: true, assigned: true },
      nearPct: 80, escalate: 'Head of IT Operations', newCat: '', newTat: '', newUnit: 'hours'
    }
  };
}
