// Initial state: projects.
export function projectsState() {
  return {
    projects: [
      { id: 'mobile', name: 'Mobile App 2.0', code: 'PRJ-04', color: '#7CC4FF', owner: 'David Otieno', team: ['David Otieno', 'Amina Wanjiru', 'Brian Kiprop', 'Kevin Mutua'], status: 'On track',
        goal: 'Give every staff member a fast mobile app with offline mode and a dark theme, so field teams can keep working without signal.',
        start: '2026-09-01', due: '2026-12-15', budget: 58, sync: 'GitHub · mobile-app', tickets: ['DEV-91', 'DEV-95'],
        milestones: [{ name: 'Design sign-off', date: '20 Sep', state: 'done' }, { name: 'Beta to 50 staff', date: '10 Oct', state: 'done' }, { name: 'Offline mode', date: '30 Oct', state: 'next' }, { name: 'Store release', date: '15 Dec', state: 'upcoming' }],
        risks: [{ text: 'Offline sync conflicts are not fully tested yet.', level: 'Medium' }, { text: 'One developer is on leave for two weeks in November.', level: 'Low' }],
        tasks: [
          { title: 'Login with company SSO', col: 3, who: 'Brian Kiprop', due: '2 Oct' }, { title: 'Crash reporting', col: 3, who: 'Brian Kiprop', due: '30 Sep' },
          { title: 'Beta feedback round', col: 3, who: 'David Otieno', due: '12 Oct' }, { title: 'Push notifications', col: 2, who: 'David Otieno', due: '20 Oct' },
          { title: 'Offline cache for forms', col: 1, who: 'Amina Wanjiru', due: '24 Oct' }, { title: 'Sync conflict tests', col: 1, who: 'Amina Wanjiru', due: '28 Oct' },
          { title: 'Dark theme colours', col: 0, who: 'Kevin Mutua', due: '5 Nov' }, { title: 'App store screenshots', col: 0, who: 'Kevin Mutua', due: '1 Dec' }
        ],
        activity: [{ text: 'David moved "Push notifications" to Review', time: '2 h ago' }, { text: 'Amina commented on "Offline cache for forms"', time: '4 h ago' }, { text: 'Milestone "Beta to 50 staff" reached', time: '5 days ago' }] },
      { id: 'network', name: 'Branch Network Upgrade', code: 'PRJ-02', color: '#6EE7C8', owner: 'Brian Kiprop', team: ['Brian Kiprop', 'David Otieno', 'Faith Achieng'], status: 'At risk',
        goal: 'Replace old routers and Wi-Fi at every branch so staff get stable connections and IT gets fewer network tickets.',
        start: '2026-08-15', due: '2026-11-30', budget: 81, sync: 'Jira · NET board', tickets: ['IT-208', 'IT-211'],
        milestones: [{ name: 'Site survey', date: '5 Sep', state: 'done' }, { name: 'Head office', date: '25 Sep', state: 'done' }, { name: 'North branch', date: '10 Oct', state: 'late' }, { name: 'Coast branch', date: '10 Nov', state: 'upcoming' }, { name: 'West branch', date: '25 Nov', state: 'upcoming' }],
        risks: [{ text: 'Router delivery for the Coast branch is delayed by two weeks.', level: 'High' }, { text: '81% of the budget is used with two branches left.', level: 'Medium' }],
        tasks: [
          { title: 'Site survey report', col: 3, who: 'Faith Achieng', due: '5 Sep' }, { title: 'Head office rollout', col: 3, who: 'Brian Kiprop', due: '25 Sep' },
          { title: 'Configure guest Wi-Fi', col: 2, who: 'David Otieno', due: '17 Oct' }, { title: 'Install North branch routers', col: 1, who: 'Brian Kiprop', due: '18 Oct' },
          { title: 'Order Coast branch equipment', col: 0, who: 'Faith Achieng', due: '22 Oct' }, { title: 'Test failover links', col: 0, who: 'David Otieno', due: '30 Oct' }
        ],
        activity: [{ text: 'Milestone "North branch" is 5 days late', time: 'Today' }, { text: 'Brian raised a risk about router delivery', time: 'Yesterday' }, { text: 'Faith uploaded "Network upgrade plan.pdf"', time: 'Last week' }] },
      { id: 'onboard', name: 'Onboarding Revamp', code: 'PRJ-05', color: '#F49AC1', owner: 'Grace Njeri', team: ['Grace Njeri', 'Faith Achieng', 'Brian Kiprop'], status: 'On track',
        goal: 'Get new staff fully set up on their first day instead of their third.',
        start: '2026-09-20', due: '2026-11-20', budget: 34, sync: 'Trello · HR board', tickets: ['HR-14'],
        milestones: [{ name: 'New checklist', date: '12 Oct', state: 'done' }, { name: 'Account automation', date: '30 Oct', state: 'next' }, { name: 'Buddy programme', date: '15 Nov', state: 'upcoming' }, { name: 'Launch', date: '20 Nov', state: 'upcoming' }],
        risks: [{ text: 'Depends on IT automating account creation.', level: 'Medium' }],
        tasks: [
          { title: 'Update onboarding checklist', col: 3, who: 'Grace Njeri', due: '12 Oct' }, { title: 'Welcome pack design', col: 2, who: 'Faith Achieng', due: '20 Oct' },
          { title: 'Automate account creation', col: 1, who: 'Brian Kiprop', due: '30 Oct' }, { title: 'Day-one schedule', col: 1, who: 'Faith Achieng', due: '25 Oct' },
          { title: 'Pick buddy volunteers', col: 0, who: 'Grace Njeri', due: '8 Nov' }
        ],
        activity: [{ text: 'Grace marked "Update onboarding checklist" done', time: 'Yesterday' }, { text: 'Milestone "New checklist" reached', time: '3 days ago' }] },
      { id: 'audit', name: 'Year-end Audit Prep', code: 'PRJ-06', color: '#FFB86B', owner: 'Amina Wanjiru', team: ['Amina Wanjiru', 'Grace Njeri', 'Kevin Mutua'], status: 'Off track',
        goal: 'Have every control, policy and evidence file ready before the external auditors arrive.',
        start: '2026-09-15', due: '2026-12-18', budget: 45, sync: 'Not synced', tickets: [],
        milestones: [{ name: 'Evidence list', date: '1 Oct', state: 'done' }, { name: 'Control testing', date: '14 Oct', state: 'late' }, { name: 'Policy sign-offs', date: '20 Nov', state: 'upcoming' }, { name: 'Audit pack', date: '10 Dec', state: 'upcoming' }],
        risks: [{ text: 'Control testing is a week behind schedule.', level: 'High' }, { text: 'Two policies are still waiting for sign-off.', level: 'Medium' }],
        tasks: [
          { title: 'Evidence list', col: 3, who: 'Amina Wanjiru', due: '1 Oct' }, { title: 'Policy acknowledgment report', col: 2, who: 'Grace Njeri', due: '18 Oct' },
          { title: 'Collect access review evidence', col: 1, who: 'Amina Wanjiru', due: '14 Oct' }, { title: 'Test payment controls', col: 0, who: 'Kevin Mutua', due: '14 Oct' },
          { title: 'Audit committee pack', col: 0, who: 'Amina Wanjiru', due: '10 Dec' }
        ],
        activity: [{ text: 'Milestone "Control testing" missed its date', time: 'Yesterday' }, { text: 'Amina uploaded "Audit committee pack.pptx"', time: 'Yesterday' }] }
    ]
  };
}
