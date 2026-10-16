// Every chat starts with 3 received, 3 missed and 3 blocked calls (plus one outgoing), spread over the last days.
function seedCallLog() {
  const H = 3600000, now = Date.now(), out = [];
  const chats = [['dm', 'Brian Kiprop'], ['dm', 'Amina Wanjiru'], ['dm', 'Grace Njeri'], ['dm', 'David Otieno'], ['forum', 'it'], ['forum', 'dev'], ['forum', 'game'], ['forum', 'choir'], ['forum', 'risk']];
  const plan = [['in', 2, 372], ['missed', 5, 0], ['blocked', 9, 0], ['in', 24, 840], ['missed', 29, 0], ['blocked', 33, 0], ['in', 52, 95], ['missed', 60, 0], ['blocked', 74, 0], ['out', 28, 1180]];
  chats.forEach(([c, ref], ci) => plan.forEach(([dir, h, dur], i) => out.push({ id: 'cl-' + ci + '-' + i, ctx: c, ref, kind: (i + ci) % 3 === 0 ? 'video' : 'voice', dir, at: now - (h + ci) * H, dur })));
  return out;
}

// Initial state: calls.
export function callsState() {
  return {
    callLog: seedCallLog(),
    callOff: {}, ouItem: null, ticketFrom: null, secRange: { preset: 'all', from: '', to: '' }, alertNotes: { 'AL-112': [{ status: 'Investigating', comment: 'Checking with Brian whether the MFA reset was approved.', who: 'Sam', at: Date.now() - 3600000 }] }, alertDlg: null, ouRecs: {}, ouThr: null,
    call: null, callMode: 'docked', mic: true, cam: true, sharing: false, hand: false, rec: false, callPanel: null, callDraft: '',
    callMsgs: [{ who: 'Grace Njeri', text: 'Joining from the Coast branch. Can everyone hear me?' }, { who: 'Brian Kiprop', text: 'Loud and clear.' }],
    ticketOpen: null, escOpen: false, esc: { type: '', id: '', cat: '', person: '', why: '' }, tcomment: '', tagOpen: false, ntOpen: false, nt: { title: '', cat: 'Network & Wi-Fi', prio: 'Medium', desc: '', files: [], toType: '', toName: '' }, unitCats: {}, catEdit: null, catWork: null, catDraft: { name: '', tat: '4', prio: 'Medium', err: '' },
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
