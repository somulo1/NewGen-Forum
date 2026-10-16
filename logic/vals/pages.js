// Values and handlers the templates read through hubBindings: pages.
export function valsPages(ctx) {
  const { addToRole, alertsTable, approvalsTable, cm, commNew, commsTable, flowsTable, logsTable, meetingsSummary, meetingsTable, membersTable, myUnits, ntDepts, openEv, org, ou, ouTable, peopleSummary, peopleTable, rightsHint, rightsTable, roleBar, rolesTable, rolesTabs, rolesView, rulesTable, s, setRoleSel, ticketsTable } = ctx;
  return {
    isDark: s.theme !== 'light', isLight: s.theme === 'light',
    themeClass: s.theme === 'light' ? 'hub-light' : 'hub-dark',
    themeFilter: s.theme === 'light' ? 'invert(1) hue-rotate(180deg)' : 'none',
    themeAria: s.theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode',
    toggleTheme: () => this.setState((x) => ({ theme: x.theme === 'light' ? 'dark' : 'light' })), newStandup: () => openEv('Standup'),
    meetingsSummary, meetingsTable, emptyList: [],
    org, myUnits, orgOpen: s.orgOpen,
    toggleOrg: () => this.setState((x) => ({ orgOpen: !x.orgOpen })),
    orgChevron: 'transition:transform 0.2s;transform:rotate(' + (s.orgOpen ? '0' : '-90') + 'deg)',
    ou, ouTable,
    cm, cmOpen: s.cmOpen, commNew, commsTable,
    ticketsTable, peopleTable, peopleSummary, logsTable, alertsTable, rulesTable,
    rolesTable, rightsTable, membersTable, flowsTable, rolesTabs, rolesView, rightsHint, ntDepts, approvalsTable, roleBar, setRoleSel, addToRole
  };
}
