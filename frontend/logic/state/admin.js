// Initial state: admin.
export function adminState() {
  return {
    orgOpen: false, ou: null, ouTab: 'comms', targets: {}, rolesTab: 'roles', memberRoles: {}, reportsTo: {}, personRights: {}, personReports: {}, reportsPerson: {}, reportPDraft: null, reportDraft: null,
    // the last theme this browser used, so the page opens in it at once; the server's saved choice replaces it when it arrives
    theme: (() => { try { return localStorage.getItem('hub-theme') === 'light' ? 'light' : 'dark'; } catch (e) { return 'dark'; } })(),
    commRules: {
      departments: { Memo: { on: true, to: ['own', 'departments'], sender: '', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: '', approver: 'none' } },
      branches: { Memo: { on: true, to: ['own', 'units', 'departments'], sender: '', approver: 'none' }, Announcement: { on: false, to: ['own'], sender: '', approver: 'none' } },
      sectors: { Memo: { on: true, to: ['own', 'branches', 'departments'], sender: '', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: '', approver: 'none' } },
      units: { Memo: { on: true, to: ['own'], sender: '', approver: 'none' }, Announcement: { on: false, to: ['own'], sender: '', approver: 'none' } }
    },
    pq: '', pdept: 'All', ploc: 'All', pAvail: false, pview: 'grid', ppage: 0, personOpen: null, userActionOpen: null, userResetTemp: '', taskOpen: null, taskMove: null, taskReminders: {},
    secTab: 'alerts', logOpen: '', alertSt: {}, openOnly: true, lq: '', lmod: 'All', lres: 'All', lflag: false, lshow: 25,
    rulesOn: { r1: true, r2: true, r3: true, r4: true, r5: true, r6: true, r7: false }, ruleSev: {}, extraLogs: [],
    profileOf: null, roleSel: '', roleTab: 'rights', rightsOv: {}, pendingRights: {}, actingRole: '', myRole: '', serverRoles: [], serverGroups: [], groupRights: {}, catalog: { modules: [], kinds: [] }, newRole: { name: '', desc: '', err: '' }, newRight: { module: '', name: '', kind: '', err: '' }, newModule: { name: '', err: '' }, apTab: 'waiting',
    flows: [],
    requests: []
  };
}
