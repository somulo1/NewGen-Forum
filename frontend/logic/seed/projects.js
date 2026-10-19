import { normalizeProject } from '~/logic/projects.js';
// Sample data derived on start-up: projects. Called with the HubLogic instance as `this`.
export function seedProjects() {
  // Task owners, timelines, descriptions and links
  const M = {
    mobile: [
      ['Brian Kiprop', 'David Otieno', '2026-09-15', '2026-10-02', 'Staff sign in to the app with their normal company account. No separate password.', []],
      ['Brian Kiprop', 'Amina Wanjiru', '2026-09-20', '2026-09-30', 'Send app crashes to the dev team automatically with the screen and device details.', []],
      ['David Otieno', 'Kevin Mutua', '2026-10-01', '2026-10-12', 'Collect feedback from the 50 beta staff and turn it into tasks.', []],
      ['David Otieno', 'Brian Kiprop', '2026-10-05', '2026-10-20', 'Notify staff of new announcements, mentions and ticket updates.', []],
      ['Amina Wanjiru', 'David Otieno', '2026-10-06', '2026-10-14', 'Let field staff fill in forms without signal. Forms send once the phone is back online.', []],
      ['Amina Wanjiru', 'Brian Kiprop', '2026-10-10', '2026-10-28', 'Test what happens when two people edit the same record offline.', [4]],
      ['Kevin Mutua', 'David Otieno', '2026-10-20', '2026-11-05', 'Choose dark theme colours that pass contrast checks.', []],
      ['Kevin Mutua', '', '2026-11-20', '2026-12-01', 'Prepare screenshots and text for the app stores.', [6]]
    ],
    network: [
      ['Faith Achieng', 'Brian Kiprop', '2026-08-20', '2026-09-05', 'Visit each branch and record the current routers, cabling and coverage.', []],
      ['Brian Kiprop', 'David Otieno', '2026-09-10', '2026-09-25', 'Replace routers and access points at head office.', [0]],
      ['David Otieno', 'Brian Kiprop', '2026-10-08', '2026-10-17', 'Separate guest Wi-Fi from staff Wi-Fi at every branch.', []],
      ['Brian Kiprop', 'John Kamau', '2026-10-01', '2026-10-10', 'Install the new routers at North branch after hours.', [4]],
      ['Faith Achieng', 'Amina Wanjiru', '2026-10-12', '2026-10-22', 'Order routers and access points for the Coast branch.', []],
      ['David Otieno', 'Brian Kiprop', '2026-10-20', '2026-10-30', 'Check each branch switches to the backup link when the main one fails.', [3]]
    ],
    onboard: [
      ['Grace Njeri', 'Faith Achieng', '2026-09-25', '2026-10-12', 'Add the missing account setup step to the onboarding checklist.', []],
      ['Faith Achieng', 'Lilian Chebet', '2026-10-05', '2026-10-20', 'Design a welcome pack for new staff.', []],
      ['Brian Kiprop', 'Grace Njeri', '2026-10-10', '2026-10-30', 'Create email and system accounts automatically when HR adds a new hire.', [0]],
      ['Faith Achieng', 'Grace Njeri', '2026-10-08', '2026-10-13', 'Plan a new hire\'s first day hour by hour.', []],
      ['Grace Njeri', 'Faith Achieng', '2026-10-25', '2026-11-08', 'Find volunteers to buddy with new staff for their first month.', []]
    ],
    audit: [
      ['Amina Wanjiru', 'Grace Njeri', '2026-09-15', '2026-10-01', 'List all the evidence the auditors will ask for.', []],
      ['Grace Njeri', 'Amina Wanjiru', '2026-10-05', '2026-10-18', 'Report who has and hasn\'t acknowledged each policy.', []],
      ['Amina Wanjiru', 'Kevin Mutua', '2026-10-01', '2026-10-14', 'Collect access review sign-offs from every department head.', [0]],
      ['Kevin Mutua', 'Amina Wanjiru', '2026-10-07', '2026-10-14', 'Test the payment approval controls on a sample of transactions.', []],
      ['Amina Wanjiru', 'Grace Njeri', '2026-11-20', '2026-12-10', 'Put together the pack for the audit committee.', [1, 2]]
    ]
  };
  const syncMap = { mobile: 'Slack · #mobile-app', network: 'Slack · #network-upgrade', onboard: 'Slack · #onboarding', audit: 'Slack · #audit-prep' };
  const teamMap = { network: ['Brian Kiprop', 'David Otieno', 'Faith Achieng', 'John Kamau', 'Amina Wanjiru', 'Kevin Mutua', 'Grace Njeri', 'Mercy Wambui'] };
  this.state.projects = this.state.projects.map((p) => ({ ...p, sync: syncMap[p.id] || p.sync, team: teamMap[p.id] || p.team,
    tasks: p.tasks.map((t, i) => { const m = (M[p.id] || [])[i] || [t.who, '', '2026-10-01', '2026-10-31', '', []]; return { title: t.title, col: t.col, primary: m[0], secondary: m[1], start: m[2], due: m[3], desc: m[4], links: m[5] }; }) })).map(normalizeProject);
}
