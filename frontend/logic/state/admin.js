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
    secTab: 'alerts', alertSt: {}, openOnly: true, lq: '', lmod: 'All', lres: 'All', lflag: false, lshow: 25,
    rulesOn: { r1: true, r2: true, r3: true, r4: true, r5: true, r6: true, r7: false }, ruleSev: {}, extraLogs: [],
    profileOf: null, roleSel: '', roleTab: 'rights', rightsOv: {}, pendingRights: {}, actingRole: '', myRole: '', serverRoles: [], serverGroups: [], groupRights: {}, catalog: { modules: [], kinds: [] }, newRole: { name: '', desc: '', err: '' }, newRight: { module: '', name: '', kind: '', err: '' }, newModule: { name: '', err: '' }, apTab: 'waiting',
    flows: [],
    requests: []
  };
}
