// Initial state: admin.
function readPeopleAdmin() {
  if (typeof localStorage === 'undefined') return {};
  try {
    return JSON.parse(localStorage.getItem('hub-people-admin') || '{}');
  } catch (error) {
    console.warn('Could not load saved people account statuses.', error);
    return {};
  }
}

export function adminState() {
  return {
    orgOpen: false, ou: null, ouTab: 'comms', targets: {}, rolesTab: 'roles', memberRoles: {}, reportsTo: {}, personRights: {}, personReports: {}, reportsPerson: {}, reportPDraft: null, reportDraft: null,
    theme: 'dark',
    commRules: {
      departments: { Memo: { on: true, to: ['own', 'departments'], sender: '', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: '', approver: 'none' } },
      branches: { Memo: { on: true, to: ['own', 'units', 'departments'], sender: '', approver: 'none' }, Announcement: { on: false, to: ['own'], sender: '', approver: 'none' } },
      sectors: { Memo: { on: true, to: ['own', 'branches', 'departments'], sender: '', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: '', approver: 'none' } },
      units: { Memo: { on: true, to: ['own'], sender: '', approver: 'none' }, Announcement: { on: false, to: ['own'], sender: '', approver: 'none' } }
    },
    pq: '', pdept: 'All', ploc: 'All', pAvail: false, pview: 'grid', ppage: 0, personOpen: null, userActionOpen: null, userResetTemp: '', peopleAdmin: readPeopleAdmin(), taskOpen: null, taskMove: null, taskReminders: {},
    secTab: 'alerts', alertSt: { 'AL-112': 'Investigating' }, openOnly: true, lq: '', lmod: 'All', lres: 'All', lflag: false, lshow: 25,
    rulesOn: { r1: true, r2: true, r3: true, r4: true, r5: true, r6: true, r7: false }, ruleSev: {}, extraLogs: [],
    profileOf: null, roleSel: '', roleTab: 'rights', rightsOv: {}, pendingRights: {}, actingRole: '', myRole: '', serverRoles: [], catalog: { modules: [], kinds: [] }, newRole: { name: '', desc: '', err: '' }, newRight: { module: '', name: '', kind: '', err: '' }, newModule: { name: '', err: '' }, apTab: 'waiting',
    flows: [
      { id: 'f1', mod: 'announcements', action: 'Publish an announcement to everyone', levels: 2, approvers: ['Head of department', 'Communications lead', 'Managing Director'] },
      { id: 'f2', mod: 'projects', action: 'Create a new project', levels: 3, approvers: ['Head of department', 'Finance approver', 'Managing Director'] },
      { id: 'f3', mod: 'files', action: 'Share a file outside Hub', levels: 1, approvers: ['Line manager', 'Security officer', 'Managing Director'] },
      { id: 'f4', mod: 'wall', action: 'Share a wall post outside Hub', levels: 1, approvers: ['Line manager', 'Communications lead', 'Managing Director'] },
      { id: 'f5', mod: 'tickets', action: 'Close a High priority ticket', levels: 1, approvers: ['Head of IT Operations', 'Head of department', 'Managing Director'] },
      { id: 'f6', mod: 'secureshare', action: 'Create a link that lasts 7 days', levels: 2, approvers: ['Line manager', 'Security officer', 'Managing Director'] },
      { id: 'f7', mod: 'awards', action: 'Give a performance award', levels: 2, approvers: ['Head of department', 'HR Manager', 'Managing Director'] },
      { id: 'f8', mod: 'roles', action: 'Change what a role is allowed to do', levels: 3, approvers: ['Security officer', 'Head of IT Operations', 'Managing Director'] },
      { id: 'f9', mod: 'files', action: 'Upload to the Finance folder', levels: 0, approvers: ['Finance approver', 'Head of department', 'Managing Director'] },
      { id: 'f10', mod: 'memos', action: 'Send a memo or announcement under a communication rule', levels: 1, approvers: ['Admin', 'Head of department', 'Managing Director'] },
      { id: 'f11', mod: 'people', action: 'Change a person\'s details, role or rights', levels: 2, approvers: ['Line manager', 'HR Manager', 'Managing Director'] }
    ],
    requests: [
      { id: 'AP-33', flowId: 'f3', title: 'Share "Q3 budget tracker.xlsx" with the external auditor', by: 'Amina Wanjiru', when: 'Today, 11:40', level: 1, status: 'Pending', history: [], note: 'The auditor needs it before Friday\'s review.' },
      { id: 'AP-32', flowId: 'f1', title: 'Announcement: Office closed Friday for maintenance', by: 'Faith Achieng', when: 'Today, 09:15', level: 2, status: 'Pending', history: [{ level: 1, who: 'Grace Njeri', decision: 'Approved', when: 'Today, 10:02' }] },
      { id: 'AP-31', flowId: 'f2', title: 'New project: Customer self-service portal', by: 'Kevin Mutua', when: 'Yesterday, 16:20', level: 3, status: 'Pending', history: [{ level: 1, who: 'Grace Njeri', decision: 'Approved', when: 'Yesterday, 17:00' }, { level: 2, who: 'Amina Wanjiru', decision: 'Approved', when: 'Today, 08:30' }], note: 'Estimated budget [AMOUNT]. Six months.' },
      { id: 'AP-29', flowId: 'f7', title: 'Employee of the month: Brian Kiprop', by: 'Grace Njeri', when: 'Monday, 10:00', level: 3, status: 'Approved', history: [{ level: 1, who: 'David Otieno', decision: 'Approved', when: 'Monday, 11:00' }, { level: 2, who: 'Grace Njeri', decision: 'Approved', when: 'Monday, 14:30' }] }
    ]
  };
}
