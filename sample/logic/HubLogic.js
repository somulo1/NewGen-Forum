import { DCLogic } from '~/utils/dcLogic.js';

// All Hub state, sample data, business rules and actions.
// renderVals() computes everything the templates show; actions call setState().
export class HubLogic extends DCLogic {
  constructor(props) {
    super(props);
    this.state = { section: 'home', mode: 'docked',
      widgets: ['awards', 'announcements', 'chat', 'calendar', 'tickets', 'projects', 'sessions', 'departments', 'branches', 'performance', 'memos'],
      navOpen: false, panelKind: 'forum', dmId: null, dmsOpen: true, newDmOpen: false, dmq: '',
      dms: [
        { id: 'dm1', with: 'Brian Kiprop', unread: 2, msgs: [{ from: 'Brian Kiprop', text: 'Hi Sam, the VPN certificate issue is fixed.', time: '2:10 PM' }, { from: 'me', text: 'Great, thanks!', time: '2:12 PM' }, { from: 'Brian Kiprop', text: 'Can we do a quick call about the SIEM rollout?', time: '3:31 PM' }, { from: 'Brian Kiprop', text: 'I have 15 minutes before 4.', time: '3:32 PM' }] },
        { id: 'dm2', with: 'Amina Wanjiru', unread: 1, msgs: [{ from: 'Amina Wanjiru', text: 'Do you have the audit evidence list?', time: '1:05 PM' }] },
        { id: 'dm3', with: 'Grace Njeri', unread: 0, msgs: [{ from: 'Grace Njeri', text: 'The access review sign-offs are in the HR folder.', time: '11:20 AM' }, { from: 'me', text: 'Perfect, I\'ll check.', time: '11:24 AM' }] },
        { id: 'dm4', with: 'David Otieno', unread: 0, msgs: [{ from: 'me', text: 'Is DEV-95 on your radar?', time: 'Yesterday' }, { from: 'David Otieno', text: 'Yes, picking it up tomorrow.', time: 'Yesterday' }] }
      ],
      awOpen: false,
      awd: { type: 'Employee of the month', people: [], role: 'all', q: '', citation: '', poster: '', wall: true, forums: [] },
      awardList: [
        { id: 'AW-12', type: 'Employee of the month', people: ['Brian Kiprop'], citation: '99.9% uptime across all branches in October.', date: '2026-10-01', status: 'Approved' },
        { id: 'AW-11', type: 'Customer hero', people: ['Faith Achieng'], citation: 'Highest customer satisfaction score this quarter.', date: '2026-09-30', status: 'Approved' },
        { id: 'AW-10', type: 'Star performer · Q3', people: ['Amina Wanjiru', 'David Otieno', 'Grace Njeri'], citation: 'Delivered the Q3 close and the new HR onboarding ahead of time.', date: '2026-09-28', status: 'Approved' },
        { id: 'AW-09', type: 'Innovation award', people: ['Kevin Mutua'], citation: 'Built the new sales pipeline tracker.', date: '2026-09-15', status: 'Approved' },
        { id: 'AW-08', type: 'Team player', people: ['John Kamau', 'Lilian Chebet'], citation: 'Always first to help other branches.', date: '2026-09-01', status: 'Approved' },
        { id: 'AW-07', type: 'Security champion', people: ['Samwel Okoth Omulo', 'Brian Kiprop'], citation: 'Zero critical findings in the Q2 audit.', date: '2026-07-10', status: 'Approved' }
      ],
      calMonth: '2026-10', calSel: '2026-10-15', calHide: {},
      evOpen: false, ev: { title: '', type: 'Meeting', date: '2026-10-15', start: '10:00', end: '11:00', repeat: 'None', forum: '', people: [], pq: '' },
      events: [
        { id: 'e1', title: 'Sprint review', type: 'Meeting', date: '2026-10-15', start: '15:00', end: '16:00', people: ['David Otieno', 'Amina Wanjiru', 'Brian Kiprop', 'Kevin Mutua'], forum: 'Product Builders', repeat: 'None', by: 'David Otieno' },
        { id: 'e2', title: 'IT daily standup', type: 'Standup', date: '2026-10-13', start: '09:15', end: '09:30', people: ['Brian Kiprop', 'David Otieno', 'John Kamau'], forum: 'IT Operations', repeat: 'Weekdays', by: 'Brian Kiprop' },
        { id: 'e3', title: 'Access review with department heads', type: 'Meeting', date: '2026-10-17', start: '11:00', end: '12:00', people: ['Grace Njeri', 'Amina Wanjiru', 'Kevin Mutua'], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' },
        { id: 'e4', title: 'Draft the Q4 risk report', type: 'Task', date: '2026-10-20', start: '', end: '', people: [], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' },
        { id: 'e5', title: 'Team building', type: 'Event', date: '2026-10-18', start: '08:00', end: '17:00', people: [], forum: '', repeat: 'None', by: 'Grace Njeri' },
        { id: 'e6', title: 'Quarterly town hall', type: 'Meeting', date: '2026-10-22', start: '10:00', end: '11:30', people: ['Brian Kiprop', 'Grace Njeri', 'David Otieno', 'Amina Wanjiru', 'Faith Achieng', 'Kevin Mutua'], forum: '', repeat: 'None', by: 'Grace Njeri' },
        { id: 'e7', title: 'Risk & Compliance weekly sync', type: 'Standup', date: '2026-10-14', start: '08:30', end: '08:45', people: ['Amina Wanjiru', 'Kevin Mutua'], forum: 'Risk & Compliance', repeat: 'Weekly', by: 'Samwel Okoth Omulo' },
        { id: 'e8', title: 'Submit SIEM rollout plan', type: 'Task', date: '2026-10-16', start: '17:00', end: '', people: [], forum: '', repeat: 'None', by: 'Samwel Okoth Omulo' }
      ],
      cmOpen: false, cmd: { type: 'Memo', title: '', body: '', wall: false, forum: '', sel: {} },
      comms: [
        { id: 'C-210', type: 'Announcement', title: 'Updated leave policy takes effect 1 November', body: 'Please read the new leave policy in HR files.', by: 'Grace Njeri', date: '2026-10-13', to: ['Community wall'], status: 'Published', read: 78 },
        { id: 'C-209', type: 'Memo', title: 'Year-end audit timelines', body: 'Evidence is due by 14 November.', by: 'Amina Wanjiru', date: '2026-10-12', to: ['Risk & Compliance', 'Finance'], status: 'Published', read: 91 },
        { id: 'C-208', type: 'Newsletter', title: 'Hub Monthly · October', body: 'New joiners, the charity run and what\'s next.', by: 'Kevin Mutua', date: '2026-10-01', to: ['Community wall'], status: 'Published', read: 64 },
        { id: 'C-207', type: 'Memo', title: 'Branch opening hours over the holiday', body: 'Branches open 9 to 1 on the public holiday.', by: 'Samwel Okoth Omulo', date: '2026-10-10', to: ['North Branch', 'Coast Branch'], status: 'Published', read: 88 },
        { id: 'C-206', type: 'Announcement', title: 'Office closed Friday for maintenance', body: 'The head office is closed this Friday.', by: 'Faith Achieng', date: '2026-10-15', to: ['Community wall', 'IT Operations forum'], status: 'Awaiting approval', read: 0 },
        { id: 'C-205', type: 'Memo', title: 'New password rules from Monday', body: 'Passwords now need 12 characters.', by: 'Samwel Okoth Omulo', date: '2026-10-06', to: ['Community wall'], status: 'Published', read: 96 },
        { id: 'C-204', type: 'Newsletter', title: 'Hub Monthly · September', body: 'Q3 highlights.', by: 'Kevin Mutua', date: '2026-09-01', to: ['Community wall'], status: 'Published', read: 71 }
      ],
      orgOpen: true, ou: null, ouTab: 'comms', targets: {}, rolesTab: 'roles', memberRoles: {},
      theme: 'dark',
      commRules: {
        departments: { Memo: { on: true, to: ['own', 'departments'], sender: 'hod', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: 'hod', approver: 'admin' } },
        branches: { Memo: { on: true, to: ['own', 'units', 'departments'], sender: 'branch', approver: 'none' }, Announcement: { on: false, to: ['own'], sender: 'branch', approver: 'admin' } },
        sectors: { Memo: { on: true, to: ['own', 'branches', 'departments'], sender: 'sector', approver: 'none' }, Announcement: { on: true, to: ['own', 'wall'], sender: 'sector', approver: 'admin' } },
        units: { Memo: { on: true, to: ['own'], sender: 'unit', approver: 'branch' }, Announcement: { on: false, to: ['own'], sender: 'unit', approver: 'branch' } }
      },
      pq: '', pdept: 'All', ploc: 'All', pAvail: false, pview: 'grid', ppage: 0, personOpen: null, taskOpen: null, taskReminders: {},
      secTab: 'alerts', alertSt: { 'AL-112': 'Investigating' }, openOnly: true, lq: '', lmod: 'All', lres: 'All', lflag: false, lshow: 25,
      rulesOn: { r1: true, r2: true, r3: true, r4: true, r5: true, r6: true, r7: false }, ruleSev: {}, extraLogs: [],
      roleSel: 'staff', roleTab: 'rights', rightsOv: {}, pendingRights: {}, actingRole: 'admin', apTab: 'waiting',
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
        { id: 'f10', mod: 'memos', action: 'Send a memo or announcement under a communication rule', levels: 1, approvers: ['Admin', 'Head of department', 'Managing Director'] }
      ],
      requests: [
        { id: 'AP-33', flowId: 'f3', title: 'Share "Q3 budget tracker.xlsx" with the external auditor', by: 'Amina Wanjiru', when: 'Today, 11:40', level: 1, status: 'Pending', history: [], note: 'The auditor needs it before Friday\'s review.' },
        { id: 'AP-32', flowId: 'f1', title: 'Announcement: Office closed Friday for maintenance', by: 'Faith Achieng', when: 'Today, 09:15', level: 2, status: 'Pending', history: [{ level: 1, who: 'Grace Njeri', decision: 'Approved', when: 'Today, 10:02' }] },
        { id: 'AP-31', flowId: 'f2', title: 'New project: Customer self-service portal', by: 'Kevin Mutua', when: 'Yesterday, 16:20', level: 3, status: 'Pending', history: [{ level: 1, who: 'Grace Njeri', decision: 'Approved', when: 'Yesterday, 17:00' }, { level: 2, who: 'Amina Wanjiru', decision: 'Approved', when: 'Today, 08:30' }], note: 'Estimated budget [AMOUNT]. Six months.' },
        { id: 'AP-29', flowId: 'f7', title: 'Employee of the month: Brian Kiprop', by: 'Grace Njeri', when: 'Monday, 10:00', level: 3, status: 'Approved', history: [{ level: 1, who: 'David Otieno', decision: 'Approved', when: 'Monday, 11:00' }, { level: 2, who: 'Grace Njeri', decision: 'Approved', when: 'Monday, 14:30' }] }
      ],
      forumsOpen: true, folder: null, projectOpen: null, newTask: '',
      profileEdit: false, statusMenu: false, newSkill: '',
      profile: {
        status: 'Available', cover: '#1F3A2B', photo: '', phone: '',
        bio: 'I look after cybersecurity and technology risk across our systems, from access reviews and audits to making sure new tools go live safely.\nHappy to help with anything security related. Just tag me in a ticket or ping me in Risk & Compliance.',
        skills: ['Cybersecurity', 'Technology risk', 'IT audit', 'Access governance', 'Business continuity', 'Laravel / PHP', 'Vue 3', 'SIEM monitoring']
      },
      folders: [
        { id: 'hr', name: 'HR Policies', color: '#F49AC1', audience: 'everyone', updated: '2 days ago', files: [
          { name: 'Leave policy 2026.pdf', kind: 'PDF', size: '420 KB', by: 'Grace Njeri', when: '2 days ago' },
          { name: 'Code of conduct.docx', kind: 'DOC', size: '96 KB', by: 'Grace Njeri', when: 'Last week' },
          { name: 'Onboarding checklist.xlsx', kind: 'XLS', size: '38 KB', by: 'Mercy Wambui', when: 'Yesterday' },
          { name: 'Staff handbook.pdf', kind: 'PDF', size: '2.1 MB', by: 'Grace Njeri', when: 'Last month' },
          { name: 'Team day photo.jpg', kind: 'IMG', size: '3.4 MB', by: 'Faith Achieng', when: 'Last week', imgBg: '#2E5B4A' }
        ] },
        { id: 'finance', name: 'Finance', color: '#6EE7C8', audience: 'Finance and managers', updated: 'today', files: [
          { name: 'Q3 budget tracker.xlsx', kind: 'XLS', size: '210 KB', by: 'Amina Wanjiru', when: 'Today' },
          { name: 'Expense claim form.pdf', kind: 'PDF', size: '120 KB', by: 'Amina Wanjiru', when: 'Last month' },
          { name: 'Audit committee pack.pptx', kind: 'PPT', size: '5.8 MB', by: 'Amina Wanjiru', when: 'Yesterday' },
          { name: 'Payroll calendar.xlsx', kind: 'XLS', size: '24 KB', by: 'Grace Njeri', when: 'Last week' }
        ] },
        { id: 'it', name: 'IT Operations', color: '#7CC4FF', audience: 'IT team', updated: '3 h ago', files: [
          { name: 'Network diagram v3.png', kind: 'IMG', size: '880 KB', by: 'Brian Kiprop', when: '3 h ago', imgBg: '#24405A' },
          { name: 'Backup procedure.docx', kind: 'DOC', size: '64 KB', by: 'Brian Kiprop', when: 'Last week' },
          { name: 'Asset register.xlsx', kind: 'XLS', size: '340 KB', by: 'David Otieno', when: 'Yesterday' },
          { name: 'Router configs.zip', kind: 'ZIP', size: '1.2 MB', by: 'Brian Kiprop', when: '2 days ago' }
        ] },
        { id: 'marketing', name: 'Marketing', color: '#FFB86B', audience: 'everyone', updated: 'yesterday', files: [
          { name: 'Town hall slides.pptx', kind: 'PPT', size: '7.2 MB', by: 'Kevin Mutua', when: 'Yesterday' },
          { name: 'Campaign poster.png', kind: 'IMG', size: '2.6 MB', by: 'Kevin Mutua', when: 'Last week', imgBg: '#6B4A2B' },
          { name: 'Social media calendar.xlsx', kind: 'XLS', size: '52 KB', by: 'Kevin Mutua', when: '3 days ago' },
          { name: 'Brand assets.zip', kind: 'ZIP', size: '48 MB', by: 'Kevin Mutua', when: 'Last month' }
        ] },
        { id: 'projects', name: 'Projects', color: '#C9B6FF', audience: 'project teams', updated: '5 h ago', files: [
          { name: 'Mobile app roadmap.pptx', kind: 'PPT', size: '3.1 MB', by: 'David Otieno', when: '5 h ago' },
          { name: 'Network upgrade plan.pdf', kind: 'PDF', size: '760 KB', by: 'Brian Kiprop', when: 'Last week' },
          { name: 'Onboarding revamp brief.docx', kind: 'DOC', size: '88 KB', by: 'Grace Njeri', when: '4 days ago' }
        ] }
      ],
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
      ],
      call: null, mic: true, cam: true, sharing: false, hand: false, rec: false, callPanel: null, callDraft: '',
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
      },
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
      ], gallery: false, customize: false, acked: {}, rsvp: {}, wished: {}, hvote: null, note: '', forum: 'it', tab: 'discussions', menu: false, invite: false, create: false, link: false, liked: {}, vote: null,
      pw: 380, dragging: false, draft: '', fdraft: '', added: {}, emojiF: false, emojiW: false, wfiles: [], shareFor: null, shareNote: {}, shareCount: {},
      posts: [
        { id: 1, author: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', meta: 'Finance · 12 min ago', text: 'Huge thanks to everyone who joined the charity run on Saturday. We raised [AMOUNT] for the children\'s home!', likes: 48, extra: 9, shares: 3, photoLabel: 'Charity run group photo', photoBg: '#2E5B4A',
          comments: [{ who: 'Brian K.', initials: 'BK', color: '#7CC4FF', text: 'What a morning! Same time next year.' }, { who: 'Grace N.', initials: 'GN', color: '#FFB86B', text: 'Proud of this team.' }] },
        { id: 2, author: 'David Otieno', initials: 'DO', color: '#F49AC1', meta: 'Weekend Gamers · 1 h ago', text: 'Which game for this month\'s tournament?', likes: 21, extra: 28, shares: 1, hasPoll: true,
          comments: [{ who: 'Kevin M.', initials: 'KM', color: '#C9B6FF', text: 'FIFA or nothing.' }] },
        { id: 3, author: 'Grace Njeri', initials: 'GN', color: '#FFB86B', meta: 'HR · Yesterday', text: 'Welcome to our 6 new team members joining this week. Say hi below and help them settle in.', likes: 73, extra: 24, shares: 5,
          comments: [{ who: 'Faith A.', initials: 'FA', color: '#B8F36B', text: 'Karibu everyone!' }] }
      ] };
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
      tasks: p.tasks.map((t, i) => { const m = (M[p.id] || [])[i] || [t.who, '', '2026-10-01', '2026-10-31', '', []]; return { title: t.title, col: t.col, primary: m[0], secondary: m[1], start: m[2], due: m[3], desc: m[4], links: m[5] }; }) }));

    // Organization structure. Levels only show when their widget is on Home.
    this.org = {
      sectors: [{ id: 's1', name: 'Business' }, { id: 's2', name: 'Operations' }, { id: 's3', name: 'Corporate' }],
      branches: [{ id: 'b1', name: 'Head Office', sector: 's3' }, { id: 'b2', name: 'North Branch', sector: 's1' }, { id: 'b3', name: 'Coast Branch', sector: 's1' }, { id: 'b4', name: 'West Branch', sector: 's2' }],
      units: [{ id: 'u1', name: 'Retail Desk', branch: 'b2' }, { id: 'u2', name: 'Collections', branch: 'b2' }, { id: 'u3', name: 'Field Team', branch: 'b3' }, { id: 'u4', name: 'Service Desk', branch: 'b1' }, { id: 'u5', name: 'Treasury', branch: 'b1' }, { id: 'u6', name: 'Fleet', branch: 'b4' }],
      departments: [
        { id: 'd1', name: 'IT Operations', branch: 'b1', sector: 's2' }, { id: 'd2', name: 'Software Development', branch: 'b1', sector: 's2' },
        { id: 'd3', name: 'Finance', branch: 'b1', sector: 's3' }, { id: 'd4', name: 'HR', branch: 'b1', sector: 's3' },
        { id: 'd5', name: 'Risk & Compliance', branch: 'b1', sector: 's3' }, { id: 'd6', name: 'Sales', branch: 'b2', sector: 's1' },
        { id: 'd7', name: 'Customer Care', branch: 'b3', sector: 's1' }, { id: 'd8', name: 'Operations', branch: 'b4', sector: 's2' },
        { id: 'd9', name: 'Marketing', branch: 'b1', sector: 's1' }, { id: 'd10', name: 'Logistics', branch: 'b4', sector: 's2' }, { id: 'd11', name: 'Legal', branch: 'b1', sector: 's3' }
      ],
      mine: { departments: 'd5', branches: 'b1', sectors: 's3', units: 'u4' }
    };
    // Tickets now belong to departments
    const tMeta = { 'IT-208': ['IT Operations', 'Customer Care', '2026-10-15', false], 'DEV-91': ['Software Development', 'HR', '2026-10-12', false], 'IT-211': ['IT Operations', 'Sales', '2026-10-14', true], 'HR-14': ['HR', 'Operations', '2026-10-12', false], 'DEV-95': ['Software Development', 'Risk & Compliance', '2026-10-15', false] };
    this.state.tickets = this.state.tickets.map((t) => { const m = tMeta[t.id] || ['IT Operations', 'Risk & Compliance', '2026-10-15', false]; return { ...t, dept: m[0], fromDept: m[1], createdAt: m[2], escalated: m[3] }; });

    // People directory (generated sample)
    const fn = ['Amina', 'Brian', 'Grace', 'David', 'Faith', 'Kevin', 'Mercy', 'John', 'Lilian', 'Peter', 'Esther', 'Daniel', 'Joy', 'Samuel', 'Ruth', 'Paul', 'Cynthia', 'Collins', 'Diana', 'Victor', 'Sharon', 'Dennis', 'Purity', 'George', 'Agnes', 'Felix', 'Irene', 'Moses', 'Naomi', 'Patrick', 'Rose', 'Stephen', 'Tabitha', 'Hassan', 'Fatuma', 'Ali', 'Zawadi', 'Kelvin', 'Beatrice', 'Allan'];
    const ln = ['Achieng', 'Wanjiru', 'Otieno', 'Kiprop', 'Njeri', 'Mutua', 'Wambui', 'Kamau', 'Chebet', 'Odhiambo', 'Akinyi', 'Mwangi', 'Nduta', 'Kibet', 'Atieno', 'Muthoni', 'Ochieng', 'Jeptoo', 'Wekesa', 'Nyambura', 'Omondi', 'Kariuki', 'Rotich', 'Barasa', 'Langat', 'Onyango', 'Gitau', 'Korir', 'Wafula', 'Maina'];
    const depts = [['Customer Care', ['Customer Care Agent', 'Team Lead, Customer Care']], ['Sales', ['Sales Executive', 'Relationship Manager', 'Sales Lead']], ['Finance', ['Accountant', 'Finance Analyst', 'Credit Officer']], ['Operations', ['Operations Officer', 'Branch Supervisor']], ['IT Operations', ['Systems Engineer', 'IT Support Officer', 'Network Engineer']], ['Software Development', ['Developer', 'QA Engineer', 'Product Designer']], ['HR', ['HR Officer', 'Recruiter']], ['Risk & Compliance', ['Risk Officer', 'Compliance Analyst', 'Internal Auditor']], ['Marketing', ['Marketing Officer', 'Content Creator']], ['Logistics', ['Driver', 'Logistics Coordinator']], ['Legal', ['Legal Officer']]];
    const locs = ['Head Office', 'North Branch', 'Coast Branch', 'West Branch'];
    const cols = ['#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#B8F36B', '#C9B6FF', '#FFD66B'];
    const sts = ['Available', 'Available', 'In a meeting', 'Available', 'Focusing', 'Away', 'Available'];
    this.dir = [];
    for (let i = 0; i < 248; i++) {
      const d = depts[(i * 7 + (i >> 2)) % depts.length];
      const first = fn[(i * 13) % fn.length], last = ln[(i * 17 + 3) % ln.length];
      const loc = locs[(i * 5 + (i >> 3)) % (i % 3 === 0 ? 1 : 4)];
      const br = this.org.branches.find((b) => b.name === loc);
      const sec = br ? this.org.sectors.find((x) => x.id === br.sector) : null;
      const us = br ? this.org.units.filter((u) => u.branch === br.id) : [];
      const unit = us.length ? (i % (us.length + 1) === us.length ? '' : us[i % (us.length + 1)].name) : '';
      this.dir.push({ id: i, name: first + ' ' + last, initials: first[0] + last[0], dept: d[0], title: d[1][i % d[1].length], loc: loc, sector: sec ? sec.name : '', unit: unit, color: cols[i % cols.length], status: sts[(i * 3) % sts.length] });
    }

    // Named colleagues are part of the directory too
    [['Brian Kiprop', 'IT Operations', 'Systems Engineer', 'Head Office', '#7CC4FF'], ['David Otieno', 'Software Development', 'Developer', 'Head Office', '#F49AC1'], ['Amina Wanjiru', 'Finance', 'Finance Analyst', 'Head Office', '#6EE7C8'],
      ['Grace Njeri', 'HR', 'HR Officer', 'Head Office', '#FFB86B'], ['Faith Achieng', 'Customer Care', 'Customer Care Agent', 'Coast Branch', '#B8F36B'], ['Kevin Mutua', 'Sales', 'Sales Lead', 'North Branch', '#C9B6FF'],
      ['Mercy Wambui', 'Finance', 'Accountant', 'Head Office', '#FFB86B'], ['John Kamau', 'Logistics', 'Driver', 'West Branch', '#7CC4FF'], ['Lilian Chebet', 'Marketing', 'Marketing Officer', 'Head Office', '#F49AC1']].forEach(([name, dept, title, loc, color]) => {
      const br = this.org.branches.find((b) => b.name === loc);
      const sec = br ? this.org.sectors.find((x) => x.id === br.sector) : null;
      this.dir.push({ id: this.dir.length, name: name, initials: name.split(' ').map((x) => x[0]).join(''), dept: dept, title: title, loc: loc, sector: sec ? sec.name : '', unit: '', color: color, status: 'Available' });
    });

    // Audit log sample (minutes ago)
    const users = ['Brian Kiprop', 'Grace Njeri', 'David Otieno', 'Amina Wanjiru', 'Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'Samwel Okoth Omulo'];
    const normal = [['Signed in', 'Auth'], ['Viewed an announcement', 'Announcements'], ['Updated ticket IT-208', 'Tickets'], ['Uploaded "Q3 budget tracker.xlsx"', 'Files'], ['Posted on the wall', 'Wall'], ['Started a video meeting', 'Meetings'], ['Created a secure link', 'Secure share'], ['Moved a task to Review', 'Projects'], ['Downloaded "Staff handbook.pdf"', 'Files'], ['Signed out', 'Auth']];
    const logs = [];
    for (let i = 0; i < 70; i++) {
      const n = normal[(i * 3) % normal.length];
      logs.push({ id: 'n' + i, ago: 4 + i * 97, user: users[(i * 5) % users.length], action: n[0], module: n[1], ip: '10.0.' + ((i * 7) % 12) + '.' + (20 + (i * 11) % 200), loc: 'Office network', result: 'Success', rule: null });
    }
    for (let k = 0; k < 6; k++) logs.push({ id: 'f' + k, ago: 12 + k, user: 'Peter Kamau', action: 'Failed sign-in (wrong password)', module: 'Auth', ip: '41.90.12.7', loc: 'Nairobi, KE', result: 'Failed', rule: 'r1' });
    logs.push({ id: 'g1', ago: 38, user: 'Diana Rotich', action: 'Signed in', module: 'Auth', ip: '102.89.3.44', loc: 'Lagos, NG', result: 'Success', rule: 'r2' });
    logs.push({ id: 'g2', ago: 80, user: 'Collins Barasa', action: 'Downloaded 64 files from Finance', module: 'Files', ip: '10.0.6.88', loc: 'Office network', result: 'Success', rule: 'r3' });
    logs.push({ id: 'g3', ago: 120, user: 'Kevin Mutua', action: 'Shared 12 posts outside Hub in one hour', module: 'Wall', ip: '10.0.3.41', loc: 'Office network', result: 'Success', rule: 'r7' });
    logs.push({ id: 'g4', ago: 180, user: 'Samwel Okoth Omulo', action: 'Changed rights for the Staff role', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' });
    logs.push({ id: 'g5', ago: 300, user: 'External viewer', action: 'Opened secure link "Payroll login"', module: 'Secure share', ip: '185.220.101.4', loc: 'Unknown', result: 'Success', rule: 'r6' });
    logs.push({ id: 'g6', ago: 1220, user: 'Brian Kiprop', action: 'Reset MFA for Peter Kamau at 23:40', module: 'Roles', ip: '10.0.2.15', loc: 'VPN', result: 'Success', rule: 'r4' });
    logs.push({ id: 'g7', ago: 640, user: 'Unknown', action: 'Sign-in blocked from a disallowed country', module: 'Auth', ip: '91.240.118.9', loc: 'Unknown', result: 'Blocked', rule: null });
    logs.push({ id: 'g8', ago: 900, user: 'Unknown', action: 'Sign-in blocked: account locked', module: 'Auth', ip: '41.90.12.7', loc: 'Nairobi, KE', result: 'Blocked', rule: null });
    this.baseLogs = logs.sort((a, b) => a.ago - b.ago);
  }
  postMsg(text, file) {
    if (this.state.panelKind === 'dm' && this.state.dmId) this.addDmMsg(this.state.dmId, text, file);
    else this.addForumMsg(this.state.forum, text, file);
  }
  addDmMsg(id, text, file) {
    this.setState((st) => ({ dms: st.dms.map((d) => (d.id === id ? { ...d, msgs: d.msgs.concat([{ from: 'me', text: text, file: file, time: 'Just now' }]) } : d)) }));
    if (!text) return;
    const replies = ['Got it, thanks!', 'Sure, give me a minute.', 'Okay, I\'ll check and get back to you.', 'Thanks for the update.', 'Can we jump on a quick call?'];
    clearTimeout(this._dmT);
    this._dmT = setTimeout(() => {
      this.setState((st) => ({ dms: st.dms.map((d) => (d.id === id ? { ...d, msgs: d.msgs.concat([{ from: d.with, text: replies[d.msgs.length % replies.length], time: 'Just now' }]) } : d)) }));
    }, 2600);
  }
  openDm(name) {
    if (!name || name === 'Samwel Okoth Omulo') return;
    this.setState((st) => {
      let conv = st.dms.find((d) => d.with === name);
      const dms = conv ? st.dms.map((d) => (d.id === conv.id ? { ...d, unread: 0 } : d)) : [{ id: 'dm-' + Date.now(), with: name, unread: 0, msgs: [] }].concat(st.dms);
      if (!conv) conv = dms[0];
      return { dms: dms, dmId: conv.id, panelKind: 'dm', mode: st.mode === 'full' ? 'full' : 'docked', menu: false, personOpen: null, newDmOpen: false, dmq: '', fdraft: '' };
    });
  }
  moveTask(pid, ti, dir) {
    const names = ['To do', 'In progress', 'Review', 'Done'];
    this.setState((st) => ({ projects: st.projects.map((p) => {
      if (p.id !== pid) return p;
      const t = p.tasks[ti];
      const nc = Math.max(0, Math.min(3, t.col + dir));
      if (nc === t.col) return p;
      return { ...p, tasks: p.tasks.map((x, j) => (j === ti ? { ...x, col: nc } : x)), activity: [{ text: 'You moved "' + t.title + '" ' + (dir > 0 ? 'forward' : 'back') + ' to ' + names[nc], time: 'Just now' }].concat(p.activity) };
    }) }));
  }
  startDrag(e) {
    e.preventDefault();
    const startX = e.clientX, startW = this.state.pw;
    const move = (ev) => {
      const w = Math.min(640, Math.max(320, startW + (startX - ev.clientX)));
      this.setState({ pw: w });
    };
    const up = () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', up);
      this.setState({ dragging: false });
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
    this.setState({ dragging: true });
  }
  componentDidMount() { try { if (window.innerWidth <= 900 && this.state.mode === 'docked') this.setState({ mode: 'min' }); } catch (err) {} this.scrollChat(); this._tick = setInterval(() => this.forceUpdate(), 30000); }
  componentWillUnmount() { clearInterval(this._tick); clearInterval(this._ct); clearTimeout(this._toastT); }
  notify(text, color, ticket) {
    this.setState((st) => ({ notifs: [{ text: text, color: color || '#B8F36B', time: 'Just now', ticket: ticket || null }].concat(st.notifs), notifSeen: false, toast: text, toastColor: color || '#B8F36B' }));
    clearTimeout(this._toastT);
    this._toastT = setTimeout(() => this.setState({ toast: '' }), 3800);
  }
  startCall(mode, title, initials, color, people) {
    clearInterval(this._ct);
    this._ct = setInterval(() => this.forceUpdate(), 1000);
    this.setState({ call: { mode: mode, title: title, initials: initials, color: color, start: Date.now(), people: people || null }, personOpen: null, mic: true, cam: mode === 'video', sharing: false, hand: false, rec: false, callPanel: null, menu: false });
  }
  endCall() {
    clearInterval(this._ct);
    this.setState({ call: null });
  }
  updTicket(id, fn) {
    this.setState((st) => ({ tickets: st.tickets.map((t) => (t.id === id ? fn(t) : t)) }));
  }
  act(t, entry) {
    return t.activity.concat([Object.assign({ who: 'Sam', time: 'Just now', kind: 'event' }, entry)]);
  }
  componentDidUpdate() { this.scrollChat(); }
  scrollChat() {
    if (this.chatEl && this._scrolledKey !== this._chatKey) {
      this.chatEl.scrollTop = this.chatEl.scrollHeight;
      this._scrolledKey = this._chatKey;
    }
  }
  fmtSize(n) {
    if (n > 1048576) return (n / 1048576).toFixed(1) + ' MB';
    return Math.max(1, Math.round(n / 1024)) + ' KB';
  }
  addForumMsg(fid, text, file) {
    this.setState((st) => ({ added: { ...st.added, [fid]: [...(st.added[fid] || []), { who: 'Sam', initials: 'S', color: '#F49AC1', time: 'Just now', text: text, file: file, replies: 0 }] } }));
  }
  renderVals() {
    const s = this.state;
    const set = (o) => () => this.setState(o);
    const openPersonName = (n) => {
      if (n === 'Sam' || n === 'Samwel Okoth Omulo') { this.setState({ section: 'profile', personOpen: null }); return; }
      const d = this.dir.find((x) => x.name === n);
      if (d) this.setState({ personOpen: d.id });
    };

    // ---- Roles & rights ----
    const rightsDef = {
      home: ['Customize home', 'Manage widget settings'],
      announcements: ['View', 'Post', 'Edit', 'Delete', 'See who read'],
      wall: ['View', 'Post', 'Comment', 'Share outside Hub', 'Moderate posts'],
      forums: ['View', 'Create forum', 'Invite members', 'Start calls', 'Moderate'],
      tickets: ['View', 'Raise', 'Assign & hand over', 'Solve', 'Close', 'Manage TAT settings'],
      projects: ['View', 'Create project', 'Edit tasks', 'Move tasks', 'Manage team'],
      files: ['View', 'Upload', 'Delete', 'Share outside Hub'],
      secureshare: ['Create link', 'View link history'],
      sessions: ['Join', 'Start meeting', 'Record'],
      awards: ['View', 'Nominate', 'Approve awards'],
      people: ['View directory', 'See contact details', 'Edit profiles'],
      security: ['View alerts', 'View logs', 'Resolve alerts', 'Manage rules', 'Export logs'],
      roles: ['View roles', 'Manage roles & rights', 'Configure approvals'],
      calendar: ['View', 'Add events', 'Schedule meetings'],
      memos: ['Read', 'Send memos'],
      newsletter: ['Read', 'Publish newsletters'],
      departments: ['View dashboard', 'Manage department', 'See performance'],
      sectors: ['View dashboard', 'Manage sector', 'See performance'],
      branches: ['View dashboard', 'Manage branch', 'See performance'],
      units: ['View dashboard', 'Manage unit', 'See performance'],
      performance: ['View', 'Set targets'],
      chat: ['Send private messages', 'Private calls']
    };
    const W = s.widgets;
    const has = {
      departments: W.indexOf('departments') !== -1, sectors: W.indexOf('sectors') !== -1, branches: W.indexOf('branches') !== -1, units: W.indexOf('units') !== -1,
      chat: W.indexOf('chat') !== -1, calendar: W.indexOf('calendar') !== -1, memos: W.indexOf('memos') !== -1, newsletter: W.indexOf('newsletter') !== -1, announcements: W.indexOf('announcements') !== -1
    };
    has.structure = has.departments || has.sectors || has.branches || has.units;
    has.comms = has.memos || has.newsletter || has.announcements;
    ['tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards'].forEach((k) => { has[k] = W.indexOf(k) !== -1; });
    if (!has.sessions) { rightsDef.calendar = ['View', 'Add events']; rightsDef.forums = rightsDef.forums.filter((x) => x !== 'Start calls'); }
    const optionalMods = ['calendar', 'memos', 'newsletter', 'announcements', 'departments', 'sectors', 'branches', 'units', 'performance', 'tickets', 'projects', 'files', 'secureshare', 'sessions', 'awards', 'chat'];
    const modEnabled = (m) => optionalMods.indexOf(m) === -1 || (m === 'performance' ? has.structure : W.indexOf(m) !== -1);
    const modInfo = {
      home: ['Home & widgets', '#B8F36B'], announcements: ['Announcements', '#B8F36B'], wall: ['Community wall', '#F49AC1'], forums: ['Forums', '#7CC4FF'],
      tickets: ['Tickets', '#FF9B8A'], projects: ['Projects', '#7CC4FF'], files: ['Files', '#C9B6FF'], secureshare: ['Secure share', '#6EE7C8'],
      sessions: ['Meetings & standups', '#6EE7C8'], awards: ['Performance awards', '#FFD66B'], people: ['People', '#C9B6FF'],
      security: ['Security & logs', '#FF9B8A'], roles: ['Roles & rights', '#FFB86B'],
      calendar: ['Calendar', '#7CC4FF'], memos: ['Memos', '#FFD66B'], newsletter: ['Newsletter', '#F49AC1'],
      departments: ['Departments', '#B8F36B'], sectors: ['Sectors', '#FFB86B'], branches: ['Branches', '#7CC4FF'], units: ['Units', '#C9B6FF'], performance: ['Performance', '#6EE7C8'], chat: ['Private messages', '#C9B6FF']
    };
    const genericRights = ['View', 'Edit content', 'Manage settings'];
    const roleDefs = [
      { id: 'superadmin', name: 'Super admin', desc: 'Full control of Hub. This role cannot be limited.', count: 2, locked: true },
      { id: 'admin', name: 'Admin', desc: 'Runs Hub day to day and manages widgets.', count: 5 },
      { id: 'hod', name: 'Head of department', desc: 'Leads a department and approves its requests.', count: 11 },
      { id: 'it', name: 'IT support', desc: 'Handles tickets and watches security.', count: 9 },
      { id: 'finance', name: 'Finance approver', desc: 'Approves spending and finance documents.', count: 4 },
      { id: 'staff', name: 'Staff', desc: 'Everyone working in the organization.', count: 212 },
      { id: 'guest', name: 'Guest', desc: 'People from outside. View only.', count: 5 },
      { id: 'sector', name: 'Sector manager', desc: 'Runs a sector and every branch in it.', count: 3, needs: 'sectors' },
      { id: 'branch', name: 'Branch manager', desc: 'Runs a branch and its units.', count: 4, needs: 'branches' },
      { id: 'unit', name: 'Unit lead', desc: 'Leads a unit inside a branch.', count: 6, needs: 'units' }
    ];
    const roleScope = { superadmin: 'Whole organization', admin: 'Whole organization', hod: 'Each department', it: 'Whole organization', finance: 'Whole organization', staff: 'Their own work', guest: 'Invited spaces only', sector: 'Each sector', branch: 'Each branch', unit: 'Each unit' };
    const visibleRoles = roleDefs.filter((r) => (r.needs ? has[r.needs] : true) && (r.id !== 'hod' || has.departments || !has.structure));
    const staffAllow = { home: ['Customize home'], announcements: ['View'], wall: ['View', 'Post', 'Comment', 'Share outside Hub'], forums: ['View', 'Start calls'],
      tickets: ['View', 'Raise'], projects: ['View', 'Edit tasks', 'Move tasks'], files: ['View', 'Upload'], secureshare: ['Create link'], sessions: ['Join', 'Start meeting'],
      awards: ['View', 'Nominate'], people: ['View directory', 'See contact details'], security: [], roles: [],
      calendar: ['View', 'Add events', 'Schedule meetings'], memos: ['Read'], newsletter: ['Read'], departments: ['View dashboard'], sectors: ['View dashboard'], branches: ['View dashboard'], units: ['View dashboard'], performance: ['View'], chat: ['Send private messages', 'Private calls'] };
    const extraAllow = {
      hod: { announcements: ['Post', 'Edit', 'See who read'], wall: ['Moderate posts'], forums: ['Create forum', 'Invite members', 'Moderate'], tickets: ['Assign & hand over', 'Solve', 'Close'], projects: ['Create project', 'Manage team'], files: ['Share outside Hub'], sessions: ['Record'], awards: ['Approve awards'], roles: ['View roles'] },
      it: { tickets: ['Assign & hand over', 'Solve', 'Close', 'Manage TAT settings'], security: ['View alerts', 'View logs', 'Resolve alerts'], files: ['Delete'], forums: ['Create forum', 'Invite members'], secureshare: ['View link history'] },
      finance: { files: ['Share outside Hub'], projects: ['Create project'] },
      sector: { sectors: ['Manage sector', 'See performance'], memos: ['Send memos'], performance: ['Set targets'], announcements: ['Post'], tickets: ['Assign & hand over'] },
      branch: { branches: ['Manage branch', 'See performance'], memos: ['Send memos'], performance: ['Set targets'], tickets: ['Assign & hand over', 'Solve'] },
      unit: { units: ['Manage unit', 'See performance'], memos: ['Send memos'], tickets: ['Solve'] }
    };
    extraAllow.hod.memos = ['Send memos']; extraAllow.hod.departments = ['Manage department', 'See performance']; extraAllow.hod.performance = ['Set targets'];
    const defaultRight = (role, mod, right) => {
      if (role === 'superadmin' || role === 'admin') return true;
      if (!rightsDef[mod]) return (role === 'hod' || role === 'sector' || role === 'branch') ? right !== 'Manage settings' : right === 'View';
      if (role === 'guest') return ['announcements', 'wall', 'forums', 'sessions'].indexOf(mod) !== -1 && (right === 'View' || right === 'Join');
      const base = (staffAllow[mod] || []).indexOf(right) !== -1;
      const extra = ((extraAllow[role] || {})[mod] || []).indexOf(right) !== -1;
      return base || extra;
    };
    const rKey = (role, mod, right) => role + '|' + mod + '|' + right;
    const rightOf = (role, mod, right) => { const k = rKey(role, mod, right); return Object.prototype.hasOwnProperty.call(s.rightsOv, k) ? s.rightsOv[k] : defaultRight(role, mod, right); };
    const can = (mod, right) => rightOf(s.actingRole, mod, right);
    const flowById = (id) => s.flows.find((f) => f.id === id);
    const nextReqId = () => 'AP-' + (34 + s.requests.length);
    const makeRequest = (flowId, title, extra) => {
      const f = flowById(flowId);
      const req = Object.assign({ id: nextReqId(), flowId: flowId, title: title, by: 'Samwel Okoth Omulo', when: 'Just now', level: 1, status: 'Pending', history: [] }, extra || {});
      this.setState((st) => ({ requests: [req].concat(st.requests) }));
      this.notify('Sent for approval: ' + title + ' (' + (f.levels === 1 ? 'direct approval' : f.levels + ' levels') + ').', '#FFB86B');
    };
    const toggleRight = (role, mod, right) => {
      if (role === 'superadmin') return;
      const k = rKey(role, mod, right);
      const nv = !rightOf(role, mod, right);
      const f = flowById('f8');
      const roleName = (roleDefs.find((r) => r.id === role) || {}).name;
      const logIt = (txt) => this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: txt, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(st.extraLogs) }));
      const label = roleName + ' · ' + (modInfo[mod] ? modInfo[mod][0] : mod) + ' · ' + right + (nv ? ' (allow)' : ' (remove)');
      if (f && f.levels > 0) {
        if (s.pendingRights[k] !== undefined) return;
        this.setState((st) => ({ pendingRights: { ...st.pendingRights, [k]: nv } }));
        makeRequest('f8', 'Change rights: ' + label, { kind: 'rights', key: k, val: nv });
        logIt('Requested rights change: ' + label);
      } else {
        this.setState((st) => ({ rightsOv: { ...st.rightsOv, [k]: nv } }));
        logIt('Changed rights: ' + label);
        this.notify('Rights updated: ' + label + '. Logged for review.', '#FFB86B');
      }
    };
    const forumsData = [
      { id: 'it', name: 'IT Operations', initials: 'IT', color: '#B8F36B', unread: 3, members: 42, online: 9, privacy: 'Invite only', about: 'Day-to-day IT support, outages and change announcements.',
        updates: [
          { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '2 min ago', text: 'Email server patch is done. Restart Outlook if you see sync errors.', replies: 4 },
          { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '14 min ago', text: 'Ticket IT-208 moved to In Progress in Jira.', replies: 1 },
          { who: 'David O.', initials: 'DO', color: '#F49AC1', time: '1 h ago', text: 'Reminder: VPN certificates expire Friday. Renew from the self-service page.', replies: 7 },
          { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: '3 h ago', text: 'Shared "Network diagram v3" in Files.', replies: 0 }
        ] },
      { id: 'dev', name: 'Product Builders', initials: 'PB', color: '#7CC4FF', unread: 5, members: 18, online: 6, privacy: 'Members of Engineering', about: 'Developers, designers and testers shipping the product.',
        updates: [
          { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: 'Just now', text: 'PR #312 merged into main via GitHub. Staging deploy running.', replies: 2 },
          { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '20 min ago', text: 'Sprint review session starts at 3pm. Join from the Sessions tab.', replies: 5 },
          { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '1 h ago', text: 'Shared staging credentials via a one-time link. Check your DMs.', replies: 0 }
        ] },
      { id: 'game', name: 'Weekend Gamers', initials: 'WG', color: '#FFB86B', unread: 0, members: 64, online: 12, privacy: 'Open to everyone', about: 'FIFA nights, tournaments and banter. Everyone welcome.',
        updates: [
          { who: 'David O.', initials: 'DO', color: '#F49AC1', time: '10 min ago', text: 'Tournament bracket is up. Vote on the game in the feed poll!', replies: 11 },
          { who: 'Brian K.', initials: 'BK', color: '#7CC4FF', time: '2 h ago', text: 'Who is on for Saturday 6pm?', replies: 9 }
        ] },
      { id: 'choir', name: 'Choir & Worship', initials: 'CW', color: '#F49AC1', unread: 1, members: 27, online: 3, privacy: 'Request to join', about: 'Rehearsal schedules, song lists and announcements.',
        updates: [
          { who: 'Grace N.', initials: 'GN', color: '#FFB86B', time: '30 min ago', text: 'Rehearsal moved to Thursday 5:30pm, main hall.', replies: 3 }
        ] },
      { id: 'risk', name: 'Risk & Compliance', initials: 'RC', color: '#6EE7C8', unread: 0, members: 11, online: 2, privacy: 'Invite only', about: 'Policies, audits and control reviews.',
        updates: [
          { who: 'Amina W.', initials: 'AW', color: '#6EE7C8', time: '1 h ago', text: 'New access-review policy uploaded. Please acknowledge by Monday.', replies: 2 }
        ] }
    ];
    const chatters = [
      { who: 'Brian K.', initials: 'BK', color: '#7CC4FF' },
      { who: 'Grace N.', initials: 'GN', color: '#FFB86B' },
      { who: 'David O.', initials: 'DO', color: '#F49AC1' },
      { who: 'Amina W.', initials: 'AW', color: '#6EE7C8' },
      { who: 'Faith A.', initials: 'FA', color: '#B8F36B' },
      { who: 'Kevin M.', initials: 'KM', color: '#C9B6FF' }
    ];
    const lines = {
      it: ['Printer on 2nd floor is back online.', 'Anyone else seeing slow Wi-Fi in the boardroom?', 'Yes, checking the access point now.', 'Fixed. It was a firmware hang.', 'Please lock your screens when stepping away.', 'New laptops arrive Monday, list is in Files.', 'Backup job finished successfully overnight.', 'Password reset requests go through Tickets please.', 'Thanks for the quick fix this morning!', 'Antivirus update rolling out at 1pm.', 'Server room AC was serviced today.', 'Who has the spare HDMI adapter?', 'It is in the IT cabinet, top drawer.', 'Scheduled downtime Saturday 8 to 10pm.'],
      dev: ['Build is green again.', 'Can someone review PR #309?', 'On it after standup.', 'Staging DB was refreshed from yesterday\'s snapshot.', 'Found the bug: timezone offset on reports.', 'Nice catch!', 'Merging the hotfix now.', 'Design for the mobile settings screen is in Files.', 'Who is on call this weekend?', 'Me, ping me on the forum.', 'Unit test coverage is now 78%.', 'Let\'s demo the new export at the sprint review.', 'Release notes draft is ready for comments.', 'API rate limit raised to 600/min.'],
      game: ['GG last night!', 'Rematch on Saturday?', 'I\'m in.', 'Bringing my own controller this time.', 'Who won the last round?', 'Kevin, again.', 'We need a rule against Kevin.', 'Snacks are on David this week.', 'Projector or the big TV?', 'Big TV, less lag.', 'Bracket goes up Thursday.', 'Can guests join?', 'Yes, just invite them to the forum.', 'Let\'s go!'],
      choir: ['Song list for Sunday is in Files.', 'Altos, please practice the bridge.', 'Can we start 15 minutes early?', 'Fine by me.', 'Keyboard needs new batteries.', 'I\'ll bring some.', 'Great rehearsal everyone.', 'Recording from Thursday is uploaded.', 'Uniform colour this week is white.', 'Who is leading the opening song?', 'Faith is.', 'Practice the harmonies at home please.', 'Thank you all!', 'See you Sunday.'],
      risk: ['Q3 control testing starts next week.', 'Please submit evidence by Friday.', 'Access review template is in Files.', 'Any exceptions to report?', 'Two, logged in the register.', 'Good, let\'s discuss Monday.', 'Policy acknowledgment is at 82%.', 'Reminder sent to the remaining staff.', 'Audit committee pack is due on the 30th.', 'Draft is ready for review.', 'Incident from last week is closed.', 'Root cause documented.', 'Thanks team.', 'Training session booked for Tuesday.']
    };
    const ago = ['4 h ago', '5 h ago', '6 h ago', 'Yesterday', 'Yesterday', 'Yesterday', '2 days ago', '2 days ago', '3 days ago', '4 days ago', 'Last week'];
    forumsData.forEach((f, fi) => {
      const pool = lines[f.id];
      for (let i = 0; i < 26; i++) {
        const c = chatters[(i + fi) % chatters.length];
        f.updates.push({ who: c.who, initials: c.initials, color: c.color, time: ago[Math.min(ago.length - 1, Math.floor(i / 2.5))], text: pool[i % pool.length], replies: (i * 3 + fi) % 6 });
      }
    });
    const cur = forumsData.find((f) => f.id === s.forum) || forumsData[0];
    const openMode = s.mode === 'full' ? 'full' : 'docked';
    const forums = forumsData.map((f) => {
      const active = f.id === cur.id && s.mode !== 'closed';
      return {
        ...f,
        hasUnread: f.unread > 0,
        open: () => this.setState({ forum: f.id, panelKind: 'forum', mode: openMode, menu: false, tab: 'discussions', invite: false }),
        rowStyle: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;color:#E8EFEA;background:' + (active ? '#22302A' : 'transparent'),
        chipStyle: 'flex:none;width:36px;height:36px;border-radius:10px;cursor:pointer;font-family:Space Grotesk,sans-serif;font-weight:700;font-size:12px;color:#0F1714;background:' + f.color + ';border:2px solid ' + (f.id === cur.id ? '#E8EFEA' : 'transparent') + ';opacity:' + (f.id === cur.id ? '1' : '0.65')
      };
    });
    const norm = (u) => ({ replies: 0, ...u, hasText: !!u.text, hasFile: !!u.file, fileName: u.file ? u.file.name : '', fileMeta: u.file ? u.file.meta : '' });
    const addedHere = (s.added[cur.id] || []).map(norm);
    const baseMsgs = cur.updates.map(norm);
    const fStyle = (m) => ({ ...m, rowStyle: 'display:flex;gap:10px', bubbleStyle: 'flex:1;min-width:0;padding:10px 12px;border-radius:4px 12px 12px 12px;background:#1A2520' });
    const forumView = { ...cur, chat: baseMsgs.slice().reverse().concat(addedHere).map(fStyle), topics: addedHere.slice().reverse().concat(baseMsgs), subline: cur.online + ' online · live updates', dotColor: '#B8F36B' };
    // Private messages (share the forum panel)
    const dmStatusCol = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
    const dmPerson = (n) => this.dir.find((d) => d.name === n) || { name: n, initials: n.split(' ').map((x) => x[0]).join(''), color: '#C9B6FF', status: 'Available', title: '' };
    const curDm = s.panelKind === 'dm' ? s.dms.find((d) => d.id === s.dmId) : null;
    const isDM = !!curDm;
    let dmView = null;
    if (curDm) {
      const pr = dmPerson(curDm.with);
      dmView = { id: curDm.id, name: curDm.with, initials: pr.initials, color: pr.color, online: '', about: pr.title, subline: (pr.status || 'Available') + ' · private chat', dotColor: dmStatusCol[pr.status] || '#B8F36B', title: pr.title,
        chat: curDm.msgs.map((m) => { const mine = m.from === 'me'; const who = mine ? 'You' : m.from; const pp = mine ? { initials: 'S', color: '#C2185B' } : dmPerson(m.from);
          return { who: who, initials: pp.initials, color: pp.color, time: m.time, text: m.text || '', hasText: !!m.text, hasFile: !!m.file, fileName: m.file ? m.file.name : '', fileMeta: m.file ? m.file.meta : '', replies: 0, mine: mine,
            rowStyle: 'display:flex;gap:10px;' + (mine ? 'flex-direction:row-reverse' : ''), bubbleStyle: 'max-width:78%;min-width:0;padding:10px 14px;border-radius:' + (mine ? '14px 4px 14px 14px;background:#2C3D34' : '4px 14px 14px 14px;background:#1A2520') }; }),
        topics: [] };
    }
    const curView = isDM ? dmView : forumView;
    this._chatKey = (isDM ? curDm.id : cur.id) + ':' + curView.chat.length;
    const emojiSet = [['😀','grinning'],['😂','laughing'],['😍','heart eyes'],['😎','cool'],['🤔','thinking'],['😢','sad'],['🙏','thank you'],['👍','thumbs up'],['👏','clap'],['🎉','party'],['🔥','fire'],['❤️','heart'],['💯','hundred'],['🚀','rocket'],['✅','done'],['👀','eyes']];
    const emojisForum = emojiSet.map(([ch, label]) => ({ ch, label, add: () => this.setState((st) => ({ fdraft: st.fdraft + ch })) }));
    const emojisWall = emojiSet.map(([ch, label]) => ({ ch, label, add: () => this.setState((st) => ({ draft: st.draft + ch })) }));
    const wfiles = s.wfiles.map((w) => ({ ...w, notImg: !w.isImg, remove: () => this.setState((st) => ({ wfiles: st.wfiles.filter((x) => x.id !== w.id) })) }));
    const navBase ='display:flex;align-items:center;gap:12px;width:100%;min-height:44px;padding:0 12px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-size:14px;';
    const sections = ['home', 'profile', 'feed', 'meetings', 'projects', 'approvals', 'security', 'roles', 'calendar', 'comms', 'structure', 'orgunit', 'tickets', 'files', 'share', 'people'];
    const nav = {}, go = {}, show = {};
    const secNeeds = { meetings: 'sessions', projects: 'projects', calendar: 'calendar', comms: 'comms', structure: 'structure', orgunit: 'structure', tickets: 'tickets', files: 'files', share: 'secureshare' };
    const secOk = (k) => !secNeeds[k] || !!has[secNeeds[k]];
    const curSec = secOk(s.section) ? s.section : 'home';
    sections.forEach((k) => {
      const on = curSec === k;
      nav[k] = navBase + (on ? 'background:#22302A;color:#E8EFEA;font-weight:600' : 'background:transparent;color:#C9D6CF');
      go[k] = set({ section: k, ticketOpen: null, projectOpen: null, folder: null, notifOpen: false, statusMenu: false });
      show[k] = on;
    });
    const tabs = ['discussions', 'sessions', 'members', 'files', 'about'];
    const tab = {}, tabTo = {}, ftab = {};
    tabs.forEach((k) => {
      const on = s.tab === k;
      tab[k] = 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;border-bottom:2px solid ' + (on ? '#B8F36B' : 'transparent') + ';color:' + (on ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (on ? '600' : '500');
      tabTo[k] = set({ tab: k });
      ftab[k] = on && (k !== 'sessions' || has.sessions);
    });
    const pollOpts = [
      { id: 'a', label: 'FIFA 26', base: 18 },
      { id: 'b', label: 'Call of Duty', base: 11 },
      { id: 'c', label: 'Mario Kart', base: 7 }
    ];
    const total = pollOpts.reduce((n, o) => n + o.base, 0) + (s.vote ? 1 : 0);
    const poll = pollOpts.map((o) => {
      const v = o.base + (s.vote === o.id ? 1 : 0);
      const pct = Math.round((v / total) * 100);
      const mine = s.vote === o.id;
      return {
        label: o.label + (mine ? '  ✓' : ''),
        pct: s.vote ? pct + '%' : '',
        vote: () => this.setState({ vote: o.id }),
        style: 'display:flex;align-items:center;min-height:44px;padding:0 14px;border-radius:10px;cursor:pointer;font-size:14px;color:#E8EFEA;text-align:left;border:1px solid ' + (mine ? '#B8F36B' : '#2A3832') + ';background:' + (s.vote ? 'linear-gradient(90deg,#2C3D34 ' + pct + '%,#121B17 ' + pct + '%)' : '#121B17')
      };
    });
    const posts = s.posts.map((p) => {
      const liked = !!s.liked[p.id];
      return {
        ...p,
        hasPoll: !!p.hasPoll,
        hasPhoto: !!p.photoBg,
        poll: poll,
        pollNote: s.vote ? total + ' votes · thanks for voting' : 'Tap an option to vote · closes Friday',
        likes: p.likes + (liked ? 1 : 0),
        commentList: p.comments,
        images: (p.media || []).filter((m) => m.isImg),
        docs: (p.media || []).filter((m) => !m.isImg),
        hasImages: (p.media || []).some((m) => m.isImg),
        hasDocs: (p.media || []).some((m) => !m.isImg),
        shares: p.shares + (s.shareCount[p.id] || 0),
        shareOpen: s.shareFor === p.id,
        shareNote: s.shareNote[p.id] || '',
        hasShareNote: !!s.shareNote[p.id],
        hasPoster: !!p.poster, posterAward: p.poster ? p.poster.award : '', posterNames: p.poster ? p.poster.names : '', posterWhen: p.poster ? p.poster.when : '',
        hasTags: !!(p.tags && p.tags.length), tagList: (p.tags || []).map((n) => ({ name: n, open: () => openPersonName(n) })),
        previewImg: ((p.media || []).find((m) => m.isImg) || {}).url || '',
        hasPreviewImg: (p.media || []).some((m) => m.isImg),
        hasPreviewBg: !!p.photoBg && !(p.media || []).some((m) => m.isImg),
        previewText: (p.text || 'Shared a file').length > 110 ? p.text.slice(0, 110) + '…' : (p.text || 'Shared a file'),
        shareOutBlocked: !can('wall', 'Share outside Hub'),
        shareOut: can('wall', 'Share outside Hub') ? ['WhatsApp', 'Email', 'LinkedIn', 'X', 'Facebook'].map((app) => ({ label: app, run: () => {
          const f = flowById('f4');
          if (f && f.levels > 0) {
            makeRequest('f4', 'Share ' + p.author + '\'s post to ' + app, { kind: 'share', postId: p.id, app: app });
            this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Waiting for approval before sharing to ' + app + '.' } }));
          } else {
            this.setState((st) => ({ shareFor: null, shareCount: { ...st.shareCount, [p.id]: (st.shareCount[p.id] || 0) + 1 }, shareNote: { ...st.shareNote, [p.id]: 'Shared to ' + app + ' with the image and text.' } }));
          }
        } })) : [],
        toggleShare: () => this.setState((st) => ({ shareFor: st.shareFor === p.id ? null : p.id })),
        copyLink: () => this.setState((st) => ({ shareFor: null, shareNote: { ...st.shareNote, [p.id]: 'Link copied. Paste it anywhere in Hub.' } })),
        shareTargets: forumsData.map((f) => ({ name: f.name, initials: f.initials, color: f.color, pick: () => {
          const snippet = p.text.length > 60 ? p.text.slice(0, 60) + '…' : p.text;
          this.addForumMsg(f.id, 'Shared a wall post from ' + p.author + ': "' + snippet + '"');
          this.setState((st) => ({ shareFor: null, shareCount: { ...st.shareCount, [p.id]: (st.shareCount[p.id] || 0) + 1 }, shareNote: { ...st.shareNote, [p.id]: 'Shared to ' + f.name } }));
        } })),
        commentCount: p.extra + p.comments.length,
        heartFill: liked ? 'currentColor' : 'none',
        like: () => this.setState({ liked: { ...this.state.liked, [p.id]: !liked } }),
        likeStyle: 'height:40px;border-radius:10px;border:0;cursor:pointer;font-size:14px;display:flex;align-items:center;justify-content:center;gap:8px;background:transparent;font-weight:' + (liked ? '600' : '400') + ';color:' + (liked ? '#FF9B8A' : '#C9D6CF'),
        onComment: (e) => {
          const v = (e.target.value || '').trim();
          if (e.key !== 'Enter' || !v) return;
          e.target.value = '';
          this.setState({ posts: this.state.posts.map((q) => q.id === p.id ? { ...q, comments: [...q.comments, { who: 'Sam', initials: 'S', color: '#F49AC1', text: v }] } : q) });
        }
      };
    });
    const moments = [
      { initials: 'BK', color: '#7CC4FF', bg: '#2B4A63', caption: 'Brian at the server room' },
      { initials: 'GN', color: '#FFB86B', bg: '#6B4A2B', caption: 'Team lunch' },
      { initials: 'DO', color: '#F49AC1', bg: '#5A2E47', caption: 'Game night setup' },
      { initials: 'FA', color: '#B8F36B', bg: '#3B5A2B', caption: 'First day!' }
    ];
    const canPost = s.draft.trim().length > 0 || s.wfiles.length > 0;
    const pw = s.pw;
    const panelStyle = 'position:relative;flex:0 0 ' + pw + 'px;width:' + pw + 'px;max-width:60%;min-height:0;box-sizing:border-box;margin:12px;border-radius:22px;border:1px solid #2A3832;background:#131D19;box-shadow:0 12px 36px rgba(0,0,0,0.35);display:flex;flex-direction:column;overflow:hidden';
    const sessions = [
      { title: 'IT daily standup', forum: 'IT Operations', initials: 'IT', color: '#B8F36B', when: 'Weekdays 9:15', live: false, btn: 'Remind me' },
      { title: 'Sprint review', forum: 'Product Builders', initials: 'PB', color: '#7CC4FF', when: 'Now', live: true, btn: 'Join now' },
      { title: 'Choir rehearsal (online)', forum: 'Choir & Worship', initials: 'CW', color: '#F49AC1', when: 'Thu 5:30pm', live: false, btn: 'RSVP' },
      { title: 'IT town hall', forum: 'IT Operations', initials: 'IT', color: '#B8F36B', when: 'Fri 10:00', live: false, btn: 'RSVP' }
    ].map((x) => ({ ...x, join: x.live ? () => this.startCall('video', x.title, x.initials, x.color) : () => this.notify('RSVP sent for ' + x.title + ' (' + x.when + ').', '#7CC4FF'), btnStyle: 'height:40px;padding:0 18px;border-radius:20px;cursor:pointer;font-weight:600;' + (x.live ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA') }));
    const prio = (p) => {
      const c = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' }[p];
      return 'padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;border:1px solid ' + c + ';color:' + c;
    };
    // ---- Tickets & TAT ----
    const ts = s.tset;
    const toMin = (v) => { const p = String(v || '0:0').split(':'); return (Number(p[0]) || 0) * 60 + (Number(p[1]) || 0); };
    const dayH = Math.max(1, (toMin(ts.end) - toMin(ts.start)) / 60);
    const pColor = { 'Brian Kiprop': '#7CC4FF', 'David Otieno': '#F49AC1', 'Amina Wanjiru': '#6EE7C8', 'Grace Njeri': '#FFB86B', 'Faith Achieng': '#B8F36B', 'Kevin Mutua': '#C9B6FF', 'Mercy Wambui': '#FFB86B', 'Sam': '#F49AC1', 'John Kamau': '#7CC4FF', 'Lilian Chebet': '#F49AC1' };
    const ini = (n) => (n === 'Sam' ? 'S' : n.split(' ').map((x) => x[0]).join('').slice(0, 2));
    const fmtH = (h) => {
      let m = Math.round(Math.abs(h) * 60);
      const dm = Math.round(dayH * 60);
      const d = Math.floor(m / dm); m -= d * dm;
      const hh = Math.floor(m / 60), mm = m % 60;
      const parts = [];
      if (d) parts.push(d + 'd');
      if (hh) parts.push(hh + 'h');
      if (mm && !d) parts.push(mm + 'm');
      return parts.join(' ') || '0m';
    };
    const catOf = (name) => ts.cats.find((c) => c.name === name) || { name: name, tat: 1, unit: 'days' };
    const tatHours = (name) => { const c = catOf(name); return c.unit === 'days' ? c.tat * dayH : c.tat; };
    const tatText = (name) => { const c = catOf(name); return c.tat + (c.unit === 'days' ? (c.tat === 1 ? ' working day' : ' working days') : (c.tat === 1 ? ' hour' : ' hours')); };
    const stepNames = ['Raised', 'Assigned', 'In progress', 'Resolved', 'Closed'];
    const badge = (c) => 'display:inline-block;padding:3px 10px;border-radius:12px;font-size:12px;font-weight:600;white-space:nowrap;background:' + c + '22;color:' + c + ';border:1px solid ' + c + '66';
    const calc = (t) => {
      const tat = tatHours(t.cat);
      const paused = t.step === 2 && t.hold && ts.pause.onhold;
      const rem = tat - t.elapsed;
      const pct = Math.min(100, Math.round((t.elapsed / tat) * 100));
      let sla, color, remainingLabel;
      if (t.step >= 3) { sla = t.late ? 'Solved late' : 'Solved on time'; color = t.late ? '#FFB86B' : '#B8F36B'; remainingLabel = 'Solved in ' + fmtH(t.elapsed); }
      else if (paused) { sla = 'TAT paused'; color = '#7CC4FF'; remainingLabel = fmtH(Math.max(0, rem)) + ' left'; }
      else if (rem < 0) { sla = 'Overdue'; color = '#FF9B8A'; remainingLabel = fmtH(-rem) + ' overdue'; }
      else if (pct >= ts.nearPct) { sla = 'Due soon'; color = '#FFB86B'; remainingLabel = fmtH(rem) + ' left'; }
      else { sla = 'On time'; color = '#B8F36B'; remainingLabel = fmtH(rem) + ' left'; }
      return { tat, rem, pct, paused, sla, color, remainingLabel };
    };
    const tickets = s.tickets.map((t) => {
      const c = calc(t);
      return { ...t, prioStyle: prio(t.prio), statusLabel: t.step === 2 && t.hold ? 'On hold' : stepNames[t.step], slaLabel: c.sla, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, handlerLabel: t.handler || 'Unassigned',
        open: () => this.setState({ ticketOpen: t.id, tagOpen: false, tcomment: '' }), _c: c };
    });
    const nOpen = tickets.filter((t) => t.step < 3).length;
    const nOver = tickets.filter((t) => t._c.sla === 'Overdue').length;
    const nSoon = tickets.filter((t) => t._c.sla === 'Due soon').length;
    const nDone = tickets.filter((t) => t.step >= 3).length;
    const ticketStats = [
      { label: 'Open', n: nOpen, color: '#E8EFEA' },
      { label: 'Overdue', n: nOver, color: '#FF9B8A' },
      { label: 'Due soon', n: nSoon, color: '#FFB86B' },
      { label: 'Solved', n: nDone, color: '#B8F36B' }
    ];
    const cur_t = s.tickets.find((t) => t.id === s.ticketOpen);
    let tk = {};
    if (cur_t) {
      const c = calc(cur_t);
      const id = cur_t.id;
      const primary = 'height:42px;border-radius:10px;border:0;background:#B8F36B;color:#0F1714;font-weight:600;cursor:pointer;font-size:14px';
      const secondary = 'height:42px;border-radius:10px;border:1px solid #2A3832;background:transparent;color:#E8EFEA;cursor:pointer;font-size:14px';
      const setStep = (step, text, extra) => this.updTicket(id, (t) => Object.assign({}, t, { step: step, hold: false, activity: this.act(t, { text: text }) }, extra || {}));
      const actions = [];
      if (cur_t.step === 1) actions.push({ label: 'Start work', style: primary, run: () => setStep(2, 'Started work') });
      if (cur_t.step === 2) {
        actions.push({ label: 'Mark as solved', style: primary, run: () => {
          const late = cur_t.elapsed > c.tat;
          setStep(3, 'Marked it solved', { late: late });
          if (ts.notify.solved) this.notify(id + ' solved' + (late ? ' (late)' : ' on time') + '. ' + cur_t.raisedBy + ' has been notified.', '#B8F36B', id);
        } });
        actions.push({ label: cur_t.hold ? 'Resume work' : 'Put on hold', style: secondary, run: () => this.updTicket(id, (t) => Object.assign({}, t, { hold: !t.hold, activity: this.act(t, { text: t.hold ? 'Resumed work' : 'Put the ticket on hold' }) })) });
      }
      if (cur_t.step === 3) {
        actions.push({ label: 'Close ticket', style: primary, run: () => setStep(4, 'Closed the ticket') });
        actions.push({ label: 'Reopen', style: secondary, run: () => setStep(2, 'Reopened the ticket') });
      }
      if (cur_t.step === 4) actions.push({ label: 'Reopen', style: secondary, run: () => setStep(2, 'Reopened the ticket') });
      const pauses = [ts.pause.weekends && 'weekends', ts.pause.holidays && 'public holidays', ts.pause.offhours && ('outside ' + ts.start + '–' + ts.end), ts.pause.onhold && 'on-hold time', ts.pause.awaiting && 'waiting for the requester'].filter(Boolean);
      tk = {
        ...cur_t,
        prioStyle: prio(cur_t.prio),
        slaLabel: c.sla, slaColor: c.color, slaBadgeStyle: badge(c.color), remainingLabel: c.remainingLabel, paused: c.paused,
        slaBarStyle: 'height:100%;width:' + Math.max(4, c.pct) + '%;background:' + c.color + ';border-radius:4px',
        tatLabel: tatText(cur_t.cat), usedLabel: fmtH(cur_t.elapsed),
        pauseNote: pauses.length ? 'Not counted: ' + pauses.join(', ') + '.' : 'TAT counts all the time.',
        hasHandler: !!cur_t.handler, noHandler: !cur_t.handler,
        handlerColor: pColor[cur_t.handler] || '#9DB0A6', handlerInitials: cur_t.handler ? ini(cur_t.handler) : '',
        handlerValue: cur_t.handler || '',
        handoverLabel: cur_t.handler ? 'Hand over to' : 'Assign to',
        wasEscalated: !!cur_t.escalatedTo, escalatedToLabel: cur_t.escalatedTo || '',
        assignees: cur_t.assignees.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })),
        actions: actions, noActions: actions.length === 0,
        steps: stepNames.map((label, i) => {
          const done = i < cur_t.step, now = i === cur_t.step;
          const notes = [cur_t.created, cur_t.handler ? 'To ' + cur_t.handler : 'Waiting for someone', cur_t.step === 2 && cur_t.hold ? 'On hold' : (cur_t.step >= 2 ? 'Work started' : ''), cur_t.step >= 3 ? (cur_t.late ? 'After the TAT' : 'Within the TAT') : 'TAT ' + tatText(cur_t.cat), cur_t.step === 4 ? 'Done' : ''];
          return {
            label: label, num: done ? '✓' : String(i + 1), note: notes[i],
            barStyle: 'display:block;height:4px;border-radius:2px;background:' + (done || now ? '#B8F36B' : '#2A3832'),
            dotStyle: 'width:22px;height:22px;flex:none;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;' + (done ? 'background:#B8F36B;color:#0F1714' : now ? 'border:2px solid #B8F36B;color:#B8F36B;box-sizing:border-box' : 'border:1px solid #3F524A;color:#9DB0A6;box-sizing:border-box'),
            labelStyle: 'font-weight:' + (now ? '700' : '500') + ';color:' + (done || now ? '#E8EFEA' : '#9DB0A6')
          };
        }),
        activity: cur_t.activity.map((a) => ({ ...a, initials: ini(a.who), color: pColor[a.who] || '#9DB0A6', isEvent: a.kind === 'event', isComment: a.kind === 'comment', isFile: a.kind === 'file', meta: a.meta || '' }))
      };
    }
    const tagNames = ['Brian Kiprop', 'David Otieno', 'Amina Wanjiru', 'Grace Njeri', 'Faith Achieng', 'Kevin Mutua'];
    const tagPeople = tagNames.map((n) => ({ name: n, tag: () => this.setState((st) => ({ tcomment: (st.tcomment && !/\s$/.test(st.tcomment) ? st.tcomment + ' ' : st.tcomment) + '@' + n + ' ', tagOpen: false })) }));
    const autoNotifs = [];
    if (has.tickets) tickets.forEach((t) => {
      if (t._c.sla === 'Overdue' && ts.notify.overdue) autoNotifs.push({ text: t.id + ' is overdue by ' + fmtH(-t._c.rem) + '. Escalated to ' + ts.escalate + '.', color: '#FF9B8A', time: 'Now', ticket: t.id });
      if (t._c.sla === 'Due soon' && ts.notify.nearing) autoNotifs.push({ text: t.id + ' is due soon: ' + fmtH(t._c.rem) + ' left on its TAT.', color: '#FFB86B', time: 'Now', ticket: t.id });
    });
    const notifList = autoNotifs.concat(s.notifs).map((n) => ({ ...n, open: () => this.setState({ notifOpen: false, section: n.ticket ? 'tickets' : this.state.section, ticketOpen: n.ticket || this.state.ticketOpen }) }));
    const chip = (on) => 'height:36px;padding:0 14px;border-radius:18px;cursor:pointer;font-size:13px;' + (on ? 'border:1px solid #B8F36B;background:#2C3D34;color:#E8EFEA;font-weight:600' : 'border:1px solid #2A3832;background:#121B17;color:#C9D6CF');
    const ntCats = ts.cats.map((c) => ({ label: c.name + ' · ' + tatText(c.name), pick: () => this.setState((st) => ({ nt: { ...st.nt, cat: c.name } })), style: chip(s.nt.cat === c.name) }));
    const ntPrios = ['High', 'Medium', 'Low'].map((p) => ({ label: p, pick: () => this.setState((st) => ({ nt: { ...st.nt, prio: p } })), style: chip(s.nt.prio === p) }));
    const setTs = (patch) => this.setState((st) => ({ tset: Object.assign({}, st.tset, typeof patch === 'function' ? patch(st.tset) : patch) }));
    const tsetCats = ts.cats.map((c, i) => ({
      ...c,
      setTat: (e) => setTs((t) => ({ cats: t.cats.map((x, j) => (j === i ? { ...x, tat: Math.max(1, Number(e.target.value) || 1) } : x)) })),
      setUnit: (e) => setTs((t) => ({ cats: t.cats.map((x, j) => (j === i ? { ...x, unit: e.target.value } : x)) })),
      remove: () => setTs((t) => ({ cats: t.cats.filter((x, j) => j !== i) }))
    }));
    const pauseRules = [
      { key: 'weekends', label: 'Weekends', desc: 'Saturday and Sunday do not count.' },
      { key: 'holidays', label: 'Public holidays', desc: 'Days on the holiday calendar do not count.' },
      { key: 'offhours', label: 'Outside working hours', desc: 'Only time between the start and end of the working day counts.' },
      { key: 'onhold', label: 'While a ticket is on hold', desc: 'Putting a ticket on hold stops its clock.' },
      { key: 'awaiting', label: 'While waiting for the requester', desc: 'The clock stops until the person who raised it replies.' }
    ].map((r) => ({ ...r, checked: !!ts.pause[r.key], toggle: () => setTs((t) => ({ pause: { ...t.pause, [r.key]: !t.pause[r.key] } })) }));
    const notifyRules = [
      { key: 'solved', label: 'When a ticket is solved', desc: 'Tell the person who raised it.' },
      { key: 'overdue', label: 'When a ticket goes overdue', desc: 'Alert the handler and escalate.' },
      { key: 'nearing', label: 'Before the TAT runs out', desc: 'Warn the handler when time is nearly up.' },
      { key: 'assigned', label: 'When a ticket is assigned or handed over', desc: 'Tell the new handler.' }
    ].map((r) => ({ ...r, checked: !!ts.notify[r.key], toggle: () => setTs((t) => ({ notify: { ...t.notify, [r.key]: !t.notify[r.key] } })) }));
    const files = [
      { kind: 'PDF', color: '#FF9B8A', name: 'Leave policy 2026', meta: 'HR · v3 · 2 days ago' },
      { kind: 'XLS', color: '#B8F36B', name: 'Q3 budget tracker', meta: 'Finance · v12 · today' },
      { kind: 'DOC', color: '#7CC4FF', name: 'Onboarding guide', meta: 'HR · v5 · last week' },
      { kind: 'PNG', color: '#FFB86B', name: 'Network diagram v3', meta: 'IT Operations · 3 h ago' },
      { kind: 'PPT', color: '#F49AC1', name: 'Town hall slides', meta: 'Management · yesterday' },
      { kind: 'ZIP', color: '#6EE7C8', name: 'Brand assets', meta: 'Marketing · last month' }
    ];
    const people = [
      { name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', role: 'Finance Analyst', forumRole: 'Admin' },
      { name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', role: 'Systems Engineer', forumRole: 'Moderator' },
      { name: 'Grace Njeri', initials: 'GN', color: '#FFB86B', role: 'HR Officer', forumRole: 'Member' },
      { name: 'David Otieno', initials: 'DO', color: '#F49AC1', role: 'Developer', forumRole: 'Member' },
      { name: 'Faith Achieng', initials: 'FA', color: '#B8F36B', role: 'Customer Care', forumRole: 'Member' },
      { name: 'Kevin Mutua', initials: 'KM', color: '#C9B6FF', role: 'Sales Lead', forumRole: 'Guest' }
    ];
    // ---- Projects ----
    const nowMs = Date.now();
    const dateOf = (d) => new Date(d + 'T00:00:00+03:00').getTime();
    const shortDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
    const longDate = (d) => new Date(d + 'T00:00:00+03:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
    const statusColor = { 'On track': '#B8F36B', 'At risk': '#FFB86B', 'Off track': '#FF9B8A' };
    const levelColor = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' };
    const colNames = [['To do', '#9DB0A6'], ['In progress', '#7CC4FF'], ['Review', '#FFB86B'], ['Done', '#B8F36B']];
    const dayStart = new Date(new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) + 'T00:00:00+03:00').getTime();
    const taskInfo = (t) => {
      const dueMs = dateOf(t.due), startMs = dateOf(t.start);
      const dd = Math.round((dueMs - dayStart) / 86400000);
      const overdue = t.col < 3 && dd < 0;
      const soon = t.col < 3 && dd >= 0 && dd <= 2;
      const color = t.col === 3 ? '#B8F36B' : overdue ? '#FF9B8A' : soon ? '#FFB86B' : '#9DB0A6';
      const label = t.col === 3 ? 'Done' : overdue ? 'Overdue ' + (-dd) + 'd' : dd === 0 ? 'Due today' : soon ? 'Due in ' + dd + 'd' : 'Due ' + shortDate(t.due);
      const pct = Math.min(100, Math.max(0, Math.round(((nowMs - startMs) / Math.max(1, dueMs + 86400000 - startMs)) * 100)));
      return { dd, overdue, soon, color, label, pct, overDays: (-dd) + (dd === -1 ? ' day' : ' days') };
    };
    const tBadge = (c) => 'display:inline-block;padding:2px 7px;border-radius:6px;font-size:10px;font-weight:700;background:' + c + '22;color:' + c;
    const canMoveT = can('projects', 'Move tasks');
    const projCalc = (p) => {
      const total = p.tasks.length, done = p.tasks.filter((t) => t.col === 3).length;
      const progress = total ? Math.round((done / total) * 100) : 0;
      const sc = statusColor[p.status] || '#9DB0A6';
      const next = p.milestones.find((m) => m.state !== 'done');
      return { total, done, progress, sc, next };
    };
    const projects = s.projects.map((p) => {
      const c = projCalc(p);
      return { ...p, initials: p.name.split(' ').map((x) => x[0]).join('').slice(0, 2), progress: c.progress,
        statusStyle: badge(c.sc), statusColor: c.sc,
        barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:4px',
        nextMilestone: c.next ? c.next.name : 'All done', dueLabel: longDate(p.due), dueShort: shortDate(p.due),
        riskCount: p.risks.length, riskColor: p.risks.some((r) => r.level === 'High') ? '#FF9B8A' : '#FFB86B',
        team: p.team.slice(0, 4).map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6' })), teamSize: p.team.length,
        open: () => this.setState({ section: 'projects', projectOpen: p.id, ticketOpen: null, newTask: '' }) };
    });
    const projStats = [
      { label: 'Active projects', n: projects.length, color: '#E8EFEA' },
      { label: 'On track', n: projects.filter((p) => p.status === 'On track').length, color: '#B8F36B' },
      { label: 'At risk', n: projects.filter((p) => p.status === 'At risk').length, color: '#FFB86B' },
      { label: 'Off track', n: projects.filter((p) => p.status === 'Off track').length, color: '#FF9B8A' }
    ];
    const cur_p = s.projects.find((p) => p.id === s.projectOpen);
    let pj = {};
    if (cur_p) {
      const c = projCalc(cur_p);
      const pid = cur_p.id;
      const startMs = dateOf(cur_p.start), dueMs = dateOf(cur_p.due);
      const daysLeft = Math.max(0, Math.ceil((dueMs - nowMs) / 86400000));
      const timeUsed = Math.min(100, Math.max(0, Math.round(((nowMs - startMs) / (dueMs - startMs)) * 100)));
      const openTask = (ti) => this.setState({ taskOpen: { pid: pid, ti: ti } });
      const msColor = { done: '#B8F36B', late: '#FF9B8A', next: '#FFB86B', upcoming: '#9DB0A6' };
      const msLabel = { done: 'Reached', late: 'Late', next: 'Next up', upcoming: 'Upcoming' };
      pj = {
        ...cur_p, initials: cur_p.name.split(' ').map((x) => x[0]).join('').slice(0, 2),
        statusStyle: badge(c.sc), progress: c.progress, doneCount: c.done, taskCount: c.total,
        inFlight: cur_p.tasks.filter((t) => t.col === 1 || t.col === 2).length,
        barStyle: 'height:100%;width:' + Math.max(3, c.progress) + '%;background:' + c.sc + ';border-radius:3px',
        daysLeft, timeUsed, startLabel: longDate(cur_p.start), dueLabel: longDate(cur_p.due),
        budgetColor: cur_p.budget > timeUsed + 10 ? '#FF9B8A' : '#E8EFEA',
        budgetNote: cur_p.budget > timeUsed + 10 ? 'Spending ahead of the timeline' : 'In line with the timeline',
        msCount: cur_p.milestones.length, msDone: cur_p.milestones.filter((m) => m.state === 'done').length,
        milestones: cur_p.milestones.map((m) => ({ ...m, color: msColor[m.state], stateLabel: msLabel[m.state],
          barStyle: 'display:block;height:4px;border-radius:2px;background:' + (m.state === 'done' ? '#B8F36B' : m.state === 'late' ? '#FF9B8A' : '#2A3832'),
          dotStyle: 'width:12px;height:12px;flex:none;border-radius:50%;box-sizing:border-box;' + (m.state === 'done' ? 'background:#B8F36B' : 'border:2px solid ' + msColor[m.state]) })),
        columns: colNames.map(([name, color], ci) => {
          const tasks = cur_p.tasks.map((t, ti) => ({ t, ti })).filter((x) => x.t.col === ci).map(({ t, ti }) => {
            const inf = taskInfo(t);
            return {
              title: t.title, primary: t.primary, secondary: t.secondary, initials: ini(t.primary), color: pColor[t.primary] || '#9DB0A6',
              hasSecondary: !!t.secondary, secInitials: t.secondary ? ini(t.secondary) : '', secColor: pColor[t.secondary] || '#9DB0A6',
              dueLabel: inf.label, dueStyle: tBadge(inf.color),
              hasLinks: t.links.length > 0, linkLabel: t.links.length + ' linked',
              cardStyle: 'padding:10px;border-radius:10px;background:#1D2A25;border:1px solid #2A3832;display:flex;flex-direction:column;gap:8px;' + (inf.overdue ? 'border-left:3px solid #FF9B8A' : ''),
              open: () => openTask(ti),
              canMove: canMoveT && ci < 3, moveLabel: ci < 3 ? 'Move to ' + colNames[ci + 1][0] : '', move: () => this.moveTask(pid, ti, 1),
              canBack: canMoveT && ci > 0, backLabel: ci > 0 ? 'Move back to ' + colNames[ci - 1][0] : '', back: () => this.moveTask(pid, ti, -1)
            };
          });
          return { name, color, tasks, count: tasks.length };
        }),
        teamCount: cur_p.team.length,
        team: cur_p.team.map((n) => ({ name: n, initials: ini(n), color: pColor[n] || '#9DB0A6', load: cur_p.tasks.filter((t) => (t.primary === n || t.secondary === n) && t.col < 3).length })),
        risks: cur_p.risks.map((r) => ({ ...r, style: 'flex:none;' + badge(levelColor[r.level]) })),
        linked: cur_p.tickets.map((id) => tickets.find((t) => t.id === id)).filter(Boolean).map((t) => ({ id: t.id, title: t.title, statusLabel: t.statusLabel, slaLabel: t.slaLabel, slaColor: t._c.color, open: () => this.setState({ section: 'tickets', ticketOpen: t.id, projectOpen: null }) })),
        noTickets: cur_p.tickets.length === 0,
        call: () => this.startCall('video', cur_p.name + ' standup', cur_p.name.split(' ').map((x) => x[0]).join('').slice(0, 2), cur_p.color)
      };
    }
    const addTask = () => {
      const title = this.state.newTask.trim();
      if (!title || !cur_p) return;
      this.setState((st) => ({ newTask: '', projects: st.projects.map((p) => (p.id === cur_p.id ? { ...p, tasks: p.tasks.concat([{ title: title, col: 0, primary: p.owner, secondary: '', start: new Date(nowMs).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), due: new Date(nowMs + 7 * 86400000).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), desc: 'Added from the task board. Open the task to add details.', links: [] }]), activity: [{ text: 'You added "' + title + '"', time: 'Just now' }].concat(p.activity) } : p)) }));
    };

    // ---- Task detail ----
    let td = {};
    const to = s.taskOpen;
    const tp = to && s.projects.find((p) => p.id === to.pid);
    const tt = tp && tp.tasks[to.ti];
    if (tt) {
      const inf = taskInfo(tt);
      const updT = (fnT) => this.setState((st) => ({ projects: st.projects.map((p) => (p.id === tp.id ? { ...p, tasks: p.tasks.map((x, j) => (j === to.ti ? fnT(x) : x)) } : p)) }));
      const rk = tp.id + ':' + to.ti;
      const extraRem = s.taskReminders[rk] || [];
      td = {
        project: tp.name, pcolor: tp.color, pinitials: tp.name.split(' ').map((x) => x[0]).join('').slice(0, 2),
        title: tt.title, desc: tt.desc || 'No description yet.',
        col: colNames[tt.col][0], colStyle: 'height:36px;padding:0 14px;border-radius:18px;display:flex;align-items:center;font-size:13px;font-weight:700;background:' + colNames[tt.col][1] + '22;color:' + colNames[tt.col][1],
        canBack: canMoveT && tt.col > 0, prevCol: tt.col > 0 ? colNames[tt.col - 1][0] : '', back: () => this.moveTask(tp.id, to.ti, -1),
        canFwd: canMoveT && tt.col < 3, nextCol: tt.col < 3 ? colNames[tt.col + 1][0] : '', fwd: () => this.moveTask(tp.id, to.ti, 1),
        overdue: inf.overdue, overDays: inf.overDays,
        reminderNote: 'An automatic reminder went to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' today.',
        remind: () => { this.setState((st) => ({ taskReminders: { ...st.taskReminders, [rk]: (st.taskReminders[rk] || []).concat(['You reminded ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' just now']) } })); this.notify('Reminder sent to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' about "' + tt.title + '".', '#FFB86B'); },
        startLabel: longDate(tt.start), dueLabel: longDate(tt.due), dueColor: inf.color,
        timeBar: 'height:100%;width:' + Math.max(3, tt.col === 3 ? 100 : inf.pct) + '%;background:' + inf.color + ';border-radius:4px',
        timeNote: tt.col === 3 ? 'Finished.' : inf.overdue ? inf.overDays + ' past the due date.' : inf.dd === 0 ? 'Due today.' : inf.dd + ' days left · ' + inf.pct + '% of the time used.',
        primary: tt.primary, pIni: ini(tt.primary), pColor: pColor[tt.primary] || '#9DB0A6',
        secondary: tt.secondary || '',
        teamOpts: tp.team.filter((n) => n !== tt.primary).map((n) => ({ name: n })),
        setSecondary: (e) => { const v = e.target.value; updT((x) => ({ ...x, secondary: v })); },
        links: tt.links.map((li) => ({ li, t: tp.tasks[li] })).filter((x) => x.t).map(({ li, t }) => ({ title: t.title, col: colNames[t.col][0], due: shortDate(t.due),
          open: () => this.setState({ taskOpen: { pid: tp.id, ti: li } }),
          unlink: () => updT((x) => ({ ...x, links: x.links.filter((y) => y !== li) })) })),
        noLinks: tt.links.length === 0,
        linkOpts: tp.tasks.map((t, i) => ({ i: String(i), title: t.title })).filter((o) => Number(o.i) !== to.ti && tt.links.indexOf(Number(o.i)) === -1),
        addLink: (e) => { const v = e.target.value; if (v === '') return; updT((x) => ({ ...x, links: x.links.concat([Number(v)]) })); },
        reminders: (inf.overdue ? ['Automatic overdue reminder sent to ' + tt.primary + (tt.secondary ? ' and ' + tt.secondary : '') + ' today'] : inf.soon ? ['Due-soon reminder sent to ' + tt.primary + ' today'] : ['Reminders go out 2 days before the due date and every day once overdue']).concat(extraRem).map((x) => ({ text: x }))
      };
    }
    const taskNotifs = [];
    if (has.projects) s.projects.forEach((p) => p.tasks.forEach((t, i) => {
      const inf = taskInfo(t);
      if (inf.overdue) taskNotifs.push({ text: 'Task "' + t.title + '" (' + p.name + ') is ' + inf.overDays + ' overdue. Reminded ' + t.primary + (t.secondary ? ' and ' + t.secondary : '') + '.', color: '#FF9B8A', time: 'Today', ticket: null, task: { pid: p.id, ti: i } });
    }));

    // ---- People directory ----
    const PAGE = 24;
    const q = s.pq.trim().toLowerCase();
    const statusCol = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
    const base = this.dir.filter((p) => (!q || (p.name + ' ' + p.title + ' ' + p.dept).toLowerCase().indexOf(q) !== -1) && (!s.pAvail || p.status === 'Available'));
    const inLoc = base.filter((p) => s.ploc === 'All' || p.loc === s.ploc);
    const results = inLoc.filter((p) => s.pdept === 'All' || p.dept === s.pdept);
    const pages = Math.max(1, Math.ceil(results.length / PAGE));
    const pg = Math.min(s.ppage, pages - 1);
    const deptNames = Array.from(new Set(this.dir.map((p) => p.dept))).sort();
    const facetStyle = (on) => 'display:flex;align-items:center;gap:8px;width:100%;min-height:36px;padding:0 10px;border-radius:8px;border:0;cursor:pointer;font-size:13px;' + (on ? 'background:#2C3D34;color:#E8EFEA;font-weight:600' : 'background:transparent;color:#C9D6CF');
    const chipS = (on) => 'height:32px;padding:0 12px;border-radius:16px;cursor:pointer;font-size:12px;' + (on ? 'border:1px solid #B8F36B;background:#2C3D34;color:#E8EFEA;font-weight:600' : 'border:1px solid #2A3832;background:transparent;color:#C9D6CF');
    const segS = (on) => 'height:34px;padding:0 14px;border-radius:9px;border:0;cursor:pointer;font-size:13px;font-weight:600;' + (on ? 'background:#E8EFEA;color:#0F1714' : 'background:transparent;color:#C9D6CF');
    const pagerS = (dis) => 'height:38px;padding:0 16px;border-radius:19px;border:1px solid #2A3832;background:transparent;font-size:13px;cursor:' + (dis ? 'default;color:#4A5C54' : 'pointer;color:#E8EFEA');
    const dir = {
      total: this.dir.length, deptCount: deptNames.length, onlineCount: this.dir.filter((p) => p.status === 'Available').length,
      q: s.pq, onQ: (e) => this.setState({ pq: e.target.value, ppage: 0 }),
      avail: s.pAvail, toggleAvail: () => this.setState((st) => ({ pAvail: !st.pAvail, ppage: 0 })),
      hasFilters: s.pdept !== 'All' || s.ploc !== 'All' || s.pAvail || !!q,
      clear: set({ pdept: 'All', ploc: 'All', pAvail: false, pq: '', ppage: 0 }),
      depts: [{ name: 'All departments', n: inLoc.length, key: 'All' }].concat(deptNames.map((d) => ({ name: d, key: d, n: inLoc.filter((p) => p.dept === d).length }))).map((d) => ({ ...d, style: facetStyle(s.pdept === d.key), pick: () => this.setState({ pdept: d.key, ppage: 0 }) })),
      locs: ['All', 'Head Office', 'North Branch', 'Coast Branch', 'West Branch'].map((l) => ({ name: l === 'All' ? 'All locations' : l, style: chipS(s.ploc === l), pick: () => this.setState({ ploc: l, ppage: 0 }) })),
      isGrid: s.pview === 'grid', isList: s.pview === 'list', gridBtn: segS(s.pview === 'grid'), listBtn: segS(s.pview === 'list'),
      setGrid: set({ pview: 'grid' }), setList: set({ pview: 'list' }),
      none: results.length === 0,
      showing: results.length ? 'Showing ' + (pg * PAGE + 1) + '–' + Math.min(results.length, (pg + 1) * PAGE) + ' of ' + results.length + ' people' : '',
      page: results.slice(pg * PAGE, (pg + 1) * PAGE).map((p) => ({ ...p, statusColor: statusCol[p.status], open: () => this.setState({ personOpen: p.id }) })),
      pageNo: pg + 1, pages, atStart: pg === 0, atEnd: pg >= pages - 1,
      prev: () => this.setState({ ppage: Math.max(0, pg - 1) }), next: () => this.setState({ ppage: Math.min(pages - 1, pg + 1) }),
      prevStyle: pagerS(pg === 0), nextStyle: pagerS(pg >= pages - 1)
    };
    const po = s.personOpen !== null && s.personOpen !== undefined ? this.dir[s.personOpen] : null;
    const skillsByDept = { 'IT Operations': ['Networking', 'Windows server', 'Help desk'], 'Software Development': ['Laravel', 'Vue', 'Testing'], 'Finance': ['Reporting', 'Excel', 'Budgeting'], 'Risk & Compliance': ['Audit', 'Policy', 'Controls'], 'Sales': ['Negotiation', 'CRM'], 'HR': ['Recruiting', 'Onboarding'], 'Customer Care': ['Customer service', 'Complaints handling'], 'Marketing': ['Content', 'Social media'], 'Operations': ['Process improvement'], 'Logistics': ['Fleet', 'Scheduling'], 'Legal': ['Contracts', 'Compliance'] };
    const psn = po ? { ...po, statusColor: statusCol[po.status], cover: po.color + '55', manager: 'Head of ' + po.dept, hubRole: /Lead|Supervisor|Manager/.test(po.title) ? 'Head of department' : 'Staff',
      email: '[' + po.name.toLowerCase().replace(' ', '.') + '@company.com]', skills: (skillsByDept[po.dept] || ['Teamwork']).map((x) => ({ name: x })),
      message: () => this.openDm(po.name),
      awards: has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf(po.name) !== -1).map((a) => ({ type: a.type, when: new Date(a.date + 'T00:00:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }), citation: a.citation })) : [],
      hasAwards: has.awards && s.awardList.some((a) => a.status === 'Approved' && a.people.indexOf(po.name) !== -1),
      canMessage: has.chat,
      voice: () => this.startCall('voice', 'Private call with ' + po.name, po.initials, po.color, [po.name]), video: () => this.startCall('video', 'Private call with ' + po.name, po.initials, po.color, [po.name]) } : {};

    // ---- Security & logs ----
    const ruleDefs = [
      { id: 'r1', name: 'Repeated failed sign-ins', desc: '5 or more failed sign-ins on one account within 10 minutes.', sev: 'High' },
      { id: 'r2', name: 'Sign-in from a new country', desc: 'Someone signs in from a country they have never used before.', sev: 'High' },
      { id: 'r3', name: 'Mass file download', desc: 'More than 50 files downloaded within 15 minutes.', sev: 'High' },
      { id: 'r4', name: 'After-hours admin action', desc: 'Admin changes made outside working hours.', sev: 'Medium' },
      { id: 'r5', name: 'Role or rights change', desc: 'Any change to roles, rights or approval rules.', sev: 'Medium' },
      { id: 'r6', name: 'Secure link opened from an unusual IP', desc: 'A one-time link is opened from outside the usual IP ranges.', sev: 'Medium' },
      { id: 'r7', name: 'Outside sharing spike', desc: 'More than 10 posts shared outside Hub within an hour.', sev: 'Low' }
    ];
    const ruleSev = (r) => s.ruleSev[r.id] || r.sev;
    const ruleOn = (id) => !!s.rulesOn[id];
    const sevColor = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#7CC4FF' };
    const agoLabel = (m) => { const d = new Date(nowMs - m * 60000); const t = d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi' }); const day = d.toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) === new Date(nowMs).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) ? 'Today' : 'Yesterday'; return day + ' ' + t; };
    const allLogs = s.extraLogs.concat(this.baseLogs);
    const alertDefs = [
      { id: 'AL-118', rule: 'r1', ago: 12, title: '6 failed sign-ins for Peter Kamau in 5 minutes', detail: 'From 41.90.12.7 (Nairobi, KE). The account was locked after the 6th attempt.' },
      { id: 'AL-117', rule: 'r2', ago: 38, title: 'Diana Rotich signed in from Lagos, Nigeria for the first time', detail: 'Her usual location is Nairobi, KE. IP 102.89.3.44.' },
      { id: 'AL-116', rule: 'r3', ago: 80, title: 'Collins Barasa downloaded 64 files from Finance in 9 minutes', detail: 'The limit is 50 files in 15 minutes.' },
      { id: 'AL-115', rule: 'r7', ago: 120, title: 'Kevin Mutua shared 12 posts outside Hub in one hour', detail: 'The limit is 10 an hour.' },
      { id: 'AL-114', rule: 'r5', ago: 180, title: 'Rights changed for the Staff role', detail: 'Changed by Samwel Okoth Omulo from the office network.' },
      { id: 'AL-113', rule: 'r6', ago: 300, title: 'Secure link "Payroll login" opened from 185.220.101.4', detail: 'Created by Grace Njeri. The IP is outside the usual ranges.' },
      { id: 'AL-112', rule: 'r4', ago: 1220, title: 'Brian Kiprop reset MFA for Peter Kamau at 23:40', detail: 'Admin action outside working hours (' + ts.start + '–' + ts.end + ').' }
    ];
    const newRightsAlerts = s.extraLogs.filter((l) => l.rule === 'r5').slice(0, 3).map((l, i) => ({ id: 'AL-' + (119 + i), rule: 'r5', ago: 0, title: l.action, detail: 'Made by ' + l.user + '. Logged automatically.' }));
    const alertsAll = newRightsAlerts.concat(alertDefs).filter((a) => ruleOn(a.rule)).map((a) => {
      const r = ruleDefs.find((x) => x.id === a.rule);
      const sv = ruleSev(r);
      const st = s.alertSt[a.id] || 'Open';
      const stC = st === 'Open' ? '#FF9B8A' : st === 'Investigating' ? '#FFB86B' : '#9DB0A6';
      const setSt = (v) => () => { this.setState((x) => ({ alertSt: { ...x.alertSt, [a.id]: v } })); this.notify(a.id + ' marked ' + v.toLowerCase() + '.', '#7CC4FF'); };
      return { ...a, sev: sv, color: sevColor[sv], sevStyle: badge(sevColor[sv]), ruleName: r.name, when: a.ago === 0 ? 'Just now' : agoLabel(a.ago), status: st, statusStyle: badge(stC),
        investigate: setSt('Investigating'), resolve: setSt('Resolved'), falsePos: setSt('False positive'), isOpen: st === 'Open' || st === 'Investigating' };
    });
    const openAlerts = alertsAll.filter((a) => a.isOpen);
    const lq = s.lq.trim().toLowerCase();
    const logsF = allLogs.filter((l) => (s.lmod === 'All' || l.module === s.lmod) && (s.lres === 'All' || l.result === s.lres) && (!s.lflag || (l.rule && ruleOn(l.rule))) && (!lq || (l.user + ' ' + l.action + ' ' + l.ip).toLowerCase().indexOf(lq) !== -1));
    const flaggedHours = {};
    allLogs.forEach((l) => { if (l.rule && ruleOn(l.rule)) flaggedHours[Math.floor(l.ago / 60)] = true; });
    const tabS = segS;
    const sec = {
      kpis: [
        { label: 'Open alerts', n: openAlerts.length, color: '#FF9B8A', sub: openAlerts.filter((a) => a.sev === 'High').length + ' high severity' },
        { label: 'Failed sign-ins (24 h)', n: allLogs.filter((l) => l.result === 'Failed' && l.ago < 1440).length, color: '#FFB86B', sub: 'Across ' + new Set(allLogs.filter((l) => l.result === 'Failed').map((l) => l.user)).size + ' account' },
        { label: 'Blocked attempts', n: allLogs.filter((l) => l.result === 'Blocked').length, color: '#7CC4FF', sub: 'Country and lockout rules' },
        { label: 'Active sessions', n: 186, color: '#B8F36B', sub: '3 on new devices' }
      ],
      bars: Array.from({ length: 24 }, (_, i) => { const h = 23 - i; const v = [3, 2, 2, 1, 1, 2, 4, 9, 14, 16, 15, 13, 12, 14, 15, 13, 11, 8, 6, 5, 4, 4, 3, 3][(i + 5) % 24]; return { style: 'flex:1;border-radius:3px 3px 0 0;height:' + (12 + v * 4.6) + '%;background:' + (flaggedHours[h] ? '#FF9B8A' : '#2C3D34') }; }),
      alerts: (s.openOnly ? openAlerts : alertsAll), noAlerts: (s.openOnly ? openAlerts : alertsAll).length === 0,
      toggleOpenOnly: () => this.setState((x) => ({ openOnly: !x.openOnly })), openOnlyStyle: chipS(s.openOnly),
      lq: s.lq, onLq: (e) => this.setState({ lq: e.target.value, lshow: 25 }),
      lmod: s.lmod, onLmod: (e) => this.setState({ lmod: e.target.value, lshow: 25 }),
      lres: s.lres, onLres: (e) => this.setState({ lres: e.target.value, lshow: 25 }),
      lflag: s.lflag, toggleLflag: () => this.setState((x) => ({ lflag: !x.lflag, lshow: 25 })),
      export: () => this.notify('Exported ' + logsF.length + ' log entries to CSV. The export was logged.', '#7CC4FF'),
      logCountLabel: logsF.length + ' entries · showing ' + Math.min(logsF.length, s.lshow),
      logs: logsF.slice(0, s.lshow).map((l) => {
        const fl = !!(l.rule && ruleOn(l.rule));
        const r = fl ? ruleDefs.find((x) => x.id === l.rule) : null;
        const c = fl ? sevColor[ruleSev(r)] : '';
        return { ...l, time: l.ago === 0 ? 'Just now' : agoLabel(l.ago), flagged: fl, ruleName: r ? r.name : '', flagStyle: fl ? badge(c) : '',
          resColor: l.result === 'Success' ? '#C9D6CF' : l.result === 'Failed' ? '#FF9B8A' : '#FFB86B',
          rowStyle: 'border-top:1px solid #24322C;' + (fl ? 'background:' + c + '12' : '') };
      }),
      hasMore: logsF.length > s.lshow, more: () => this.setState((x) => ({ lshow: x.lshow + 25 })),
      rules: ruleDefs.map((r) => ({ ...r, on: ruleOn(r.id), sev: ruleSev(r), hits: allLogs.filter((l) => l.rule === r.id).length + ' matches today',
        toggle: () => this.setState((x) => ({ rulesOn: { ...x.rulesOn, [r.id]: !x.rulesOn[r.id] } })),
        setSev: (e) => { const v = e.target.value; this.setState((x) => ({ ruleSev: { ...x.ruleSev, [r.id]: v } })); } }))
    };
    const secTab = { alerts: tabS(s.secTab === 'alerts'), logs: tabS(s.secTab === 'logs'), rules: tabS(s.secTab === 'rules'),
      alertsGo: set({ secTab: 'alerts' }), logsGo: set({ secTab: 'logs' }), rulesGo: set({ secTab: 'rules' }),
      isAlerts: s.secTab === 'alerts', isLogs: s.secTab === 'logs', isRules: s.secTab === 'rules' };

    // ---- Roles page ----
    const memberSamples = { superadmin: ['Samwel Okoth Omulo', 'Brian Kiprop'], admin: ['Samwel Okoth Omulo', 'David Otieno', 'Grace Njeri'], hod: ['Grace Njeri', 'Amina Wanjiru', 'Kevin Mutua'], it: ['Brian Kiprop', 'David Otieno', 'John Kamau'], finance: ['Amina Wanjiru', 'Mercy Wambui'], staff: ['Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'John Kamau', 'Lilian Chebet'], guest: ['External auditor'] };
    const roleSelDef = roleDefs.find((r) => r.id === s.roleSel) || roleDefs[5];
    const roleList = roleDefs.map((r) => ({ ...r, pick: () => this.setState({ roleSel: r.id }), style: 'display:flex;align-items:center;gap:8px;width:100%;min-height:52px;padding:6px 12px;border-radius:10px;cursor:pointer;color:#E8EFEA;border:1px solid ' + (s.roleSel === r.id ? '#3F524A' : 'transparent') + ';background:' + (s.roleSel === r.id ? '#22302A' : 'transparent') }));
    const roleSel = { ...roleSelDef,
      members: (memberSamples[roleSelDef.id] || []).map((n) => ({ name: n, initials: n === 'Samwel Okoth Omulo' ? 'S' : ini(n), color: pColor[n] || '#C9B6FF' })),
      modules: Object.keys(rightsDef).map((m) => {
        const rights = rightsDef[m].map((r) => { const k = rKey(roleSelDef.id, m, r); return { name: r, on: rightOf(roleSelDef.id, m, r), locked: !!roleSelDef.locked, pending: s.pendingRights[k] !== undefined, toggle: () => toggleRight(roleSelDef.id, m, r) }; });
        return { label: modInfo[m][0], tint: modInfo[m][1], rights, onCount: rights.filter((x) => x.on).length, total: rights.length };
      }) };
    const roleTab = { rights: tabS(s.roleTab === 'rights'), members: tabS(s.roleTab === 'members'), flows: tabS(s.roleTab === 'flows'),
      rightsGo: set({ roleTab: 'rights' }), membersGo: set({ roleTab: 'members' }), flowsGo: set({ roleTab: 'flows' }),
      isRights: s.roleTab === 'rights', isMembers: s.roleTab === 'members', isFlows: s.roleTab === 'flows' };
    const flowItem = (f) => ({ ...f, modName: modInfo[f.mod] ? modInfo[f.mod][0] : f.mod, levelsVal: String(f.levels), hasLevels: f.levels > 0,
      setLevels: (e) => { const v = Number(e.target.value); this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, levels: v } : x)) })); },
      levels: f.approvers.slice(0, f.levels).map((who, i) => ({ n: i + 1, who, set: (e) => { const v = e.target.value; this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, approvers: x.approvers.map((a, j) => (j === i ? v : a)) } : x)) })); } })) });
    const flowsAll = s.flows.map(flowItem);
    const f8 = flowById('f8');

    // ---- Approvals inbox ----
    const reqItems = s.requests.map((rq) => {
      const f = flowById(rq.flowId) || { levels: 1, approvers: ['Line manager'], mod: 'home' };
      const L = Math.max(1, f.levels);
      const pending = rq.status === 'Pending';
      const stC = rq.status === 'Approved' ? '#B8F36B' : rq.status === 'Rejected' ? '#FF9B8A' : '#FFB86B';
      const steps = Array.from({ length: L }, (_, i) => {
        const lvl = i + 1;
        const h = rq.history.find((x) => x.level === lvl);
        const done = !!h && h.decision === 'Approved';
        const rej = !!h && h.decision === 'Rejected';
        const curL = pending && lvl === rq.level;
        const c = done ? '#B8F36B' : rej ? '#FF9B8A' : curL ? '#FFB86B' : '#9DB0A6';
        return { n: done ? '✓' : String(lvl), who: f.approvers[i] || 'Approver', color: c,
          state: h ? h.decision + ' by ' + h.who + ' · ' + h.when : curL ? 'Waiting now' : 'Not reached yet',
          bar: 'display:block;height:4px;border-radius:2px;background:' + (done ? '#B8F36B' : rej ? '#FF9B8A' : curL ? '#FFB86B' : '#2A3832'),
          dot: 'width:22px;height:22px;flex:none;border-radius:50%;display:flex;align-items:center;justify-content:center;font-size:11px;font-weight:700;box-sizing:border-box;' + (done ? 'background:#B8F36B;color:#0F1714' : 'border:2px solid ' + c + ';color:' + c) };
      });
      const decide = (ok) => () => {
        const cur = this.state.requests.find((x) => x.id === rq.id);
        const lvl = cur.level;
        const hist = cur.history.concat([{ level: lvl, who: 'You', decision: ok ? 'Approved' : 'Rejected', when: 'Just now' }]);
        if (!ok) {
          this.setState((st) => ({ requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, status: 'Rejected' } : x)), pendingRights: rq.kind === 'rights' ? Object.fromEntries(Object.entries(st.pendingRights).filter(([k]) => k !== rq.key)) : st.pendingRights }));
          this.notify(rq.id + ' rejected. ' + rq.by + ' has been told.', '#FF9B8A');
          return;
        }
        if (lvl < L) {
          this.setState((st) => ({ requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, level: lvl + 1 } : x)) }));
          this.notify(rq.id + ' approved at level ' + lvl + '. Sent to ' + (f.approvers[lvl] || 'the next approver') + '.', '#B8F36B');
        } else {
          this.setState((st) => {
            const out = { requests: st.requests.map((x) => (x.id === rq.id ? { ...x, history: hist, status: 'Approved' } : x)) };
            if (rq.kind === 'rights') { out.rightsOv = { ...st.rightsOv, [rq.key]: rq.val }; out.pendingRights = Object.fromEntries(Object.entries(st.pendingRights).filter(([k]) => k !== rq.key)); }
            if (rq.kind === 'award') {
              const a = st.awardList.find((x) => x.id === rq.awardId);
              out.awardList = st.awardList.map((x) => (x.id === rq.awardId ? { ...x, status: 'Approved' } : x));
              if (a && a.wall) {
                const names = a.people.join(', ');
                out.posts = [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: 'Performance award · Just now', text: 'Congratulations ' + a.people.map((n) => '@' + n).join(', ') + ' on the ' + a.type + ' award!' + (a.citation ? '\n' + a.citation : ''), likes: 0, extra: 0, shares: 0, comments: [],
                  media: a.poster ? [{ id: 'pst' + Date.now(), name: 'Award poster', meta: '', isImg: true, url: a.poster }] : [],
                  poster: a.poster ? null : { award: a.type, names: names, when: 'Today' }, tags: a.people.slice() }].concat(out.posts || st.posts);
              }
            }
            if (rq.kind === 'comm') {
              const cmItem = st.comms.find((c) => c.id === rq.commId);
              out.comms = st.comms.map((c) => (c.id === rq.commId ? { ...c, status: 'Published', read: 0 } : c));
              if (cmItem && cmItem.to.indexOf('Community wall') !== -1) out.posts = [{ id: Date.now(), author: cmItem.by === 'Samwel Okoth Omulo' ? 'Sam' : cmItem.by, initials: cmItem.by === 'Samwel Okoth Omulo' ? 'S' : ini(cmItem.by), color: '#F49AC1', meta: cmItem.type + ' · Just now', text: cmItem.title + '\n' + cmItem.body, likes: 0, extra: 0, shares: 0, comments: [] }].concat(st.posts);
            }
            if (rq.kind === 'share') { out.shareCount = { ...st.shareCount, [rq.postId]: (st.shareCount[rq.postId] || 0) + 1 }; out.shareNote = { ...st.shareNote, [rq.postId]: 'Approved and shared to ' + rq.app + ' with the image and text.' }; }
            return out;
          });
          if (rq.kind === 'award') {
            const a = this.state.awardList.find((x) => x.id === rq.awardId);
            if (a) (a.forums || []).forEach((fid) => this.addForumMsg(fid, 'Award: ' + a.type + ' · Congratulations ' + a.people.map((n) => '@' + n).join(', ') + '!' + (a.poster ? ' (poster attached)' : ''), a.poster ? { name: 'Award poster.png', meta: 'Image' } : undefined));
          }
          this.notify(rq.id + ' fully approved: ' + rq.title + (rq.kind === 'award' ? '. Everyone awarded was tagged and notified.' : '.'), '#B8F36B');
        }
      };
      return { ...rq, modName: modInfo[f.mod] ? modInfo[f.mod][0] : '', steps, statusLabel: pending ? 'Level ' + rq.level + ' of ' + L : rq.status, statusStyle: badge(stC),
        levelsLabel: L === 1 ? 'Direct approval' : L + '-level approval', hasNote: !!rq.note,
        canAct: pending, currentApprover: f.approvers[rq.level - 1] || 'approver', approveLabel: rq.level < L ? 'Approve & send on' : 'Approve',
        approve: decide(true), reject: decide(false) };
    });
    const pendingCount = reqItems.filter((r) => r.status === 'Pending').length;
    const requestsShown = s.apTab === 'waiting' ? reqItems.filter((r) => r.status === 'Pending') : reqItems;

    // =============== Shared helpers for tables ===============
    const tz9 = { timeZone: 'Africa/Nairobi' };
    const todayISO = new Date(nowMs).toLocaleDateString('en-CA', tz9);
    const isoAdd = (iso, d) => new Date(new Date(iso + 'T12:00:00+03:00').getTime() + d * 86400000).toLocaleDateString('en-CA', tz9);
    const dowOf = (iso) => new Date(iso + 'T12:00:00+03:00').getUTCDay();
    const fmtDay = (iso) => new Date(iso + 'T12:00:00+03:00').toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short', timeZone: 'Africa/Nairobi' });
    const personOf = (n) => ({ name: n, initials: n === 'Samwel Okoth Omulo' ? 'S' : ini(n), color: pColor[n] || ((this.dir.find((d) => d.name === n) || {}).color) || '#C9B6FF' });
    const C = {
      t: (text, o) => Object.assign({ text: String(text == null ? '' : text) }, o || {}),
      b: (text, color, o) => Object.assign({ kind: 'badge', text: text, color: color }, o || {}),
      p: (name, o) => { const pp = personOf(name); return Object.assign({ kind: 'person', text: name, initials: pp.initials, color: pp.color }, o || {}); },
      a: (buttons) => ({ kind: 'actions', buttons: buttons, fv: '' }),
      bar: (pct, color, text) => ({ kind: 'bar', pct: pct, color: color, text: text || (pct + '%'), fv: String(pct) }),
      chk: (on, toggle, o) => Object.assign({ kind: 'check', on: on, toggle: toggle, fv: on ? 'Yes' : 'No' }, o || {}),
      sel: (value, options, onChange, o) => Object.assign({ kind: 'select', value: value, options: options, onChange: onChange, fv: value }, o || {})
    };
    const corePeople = ['Brian Kiprop', 'David Otieno', 'Amina Wanjiru', 'Grace Njeri', 'Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'John Kamau', 'Lilian Chebet'];

    // =============== Calendar ===============
    const typeColor = { Meeting: '#7CC4FF', Standup: '#6EE7C8', Task: '#FFB86B', Event: '#C9B6FF', Personal: '#F49AC1', Deadline: '#FF9B8A' };
    const occurs = (e, iso) => {
      if (e.repeat === 'None' || !e.repeat) return e.date === iso;
      if (iso < e.date) return false;
      const d = dowOf(iso);
      if (e.repeat === 'Daily') return true;
      if (e.repeat === 'Weekdays') return d >= 1 && d <= 5;
      if (e.repeat === 'Weekly') return d === dowOf(e.date);
      return false;
    };
    const deadlines = [];
    if (has.projects) s.projects.forEach((p) => p.tasks.forEach((t) => { if (t.col < 3) deadlines.push({ id: 'dl-' + p.id + t.title, title: t.title, type: 'Deadline', date: t.due, start: '', end: '', people: [t.primary].concat(t.secondary ? [t.secondary] : []), forum: '', repeat: 'None', project: p.name }); }));
    const itemsOn = (iso) => s.events.filter((e) => !s.calHide[e.type] && (has.sessions || (e.type !== 'Meeting' && e.type !== 'Standup')) && occurs(e, iso)).concat(s.calHide.Deadline ? [] : deadlines.filter((d) => d.date === iso)).sort((a, b) => (a.start || '99').localeCompare(b.start || '99'));
    const cy = Number(s.calMonth.slice(0, 4)), cmn = Number(s.calMonth.slice(5, 7));
    const firstISO = s.calMonth + '-01';
    const offset = (dowOf(firstISO) + 6) % 7;
    const dim = new Date(Date.UTC(cy, cmn, 0)).getUTCDate();
    const weeks = Math.ceil((offset + dim) / 7);
    const monthShift = (d) => { let y = cy, m = cmn + d; if (m < 1) { m = 12; y--; } if (m > 12) { m = 1; y++; } return y + '-' + (m < 10 ? '0' : '') + m; };
    const chipStyleC = (c) => 'display:block;max-width:100%;overflow:hidden;text-overflow:ellipsis;white-space:nowrap;padding:2px 6px;border-radius:5px;font-size:10px;font-weight:600;background:' + c + '26;color:' + c;
    const joinEvent = (e) => this.startCall(e.type === 'Standup' ? 'voice' : 'video', e.title, e.title.split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase(), typeColor[e.type], e.people);
    const cal = {
      monthLabel: new Date(firstISO + 'T12:00:00+03:00').toLocaleDateString('en-GB', { month: 'long', year: 'numeric', timeZone: 'Africa/Nairobi' }),
      prevMonth: () => this.setState({ calMonth: monthShift(-1) }), nextMonth: () => this.setState({ calMonth: monthShift(1) }),
      today: () => this.setState({ calMonth: todayISO.slice(0, 7), calSel: todayISO }),
      weekdays: ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'].map((l) => ({ l: l })),
      legend: [['Personal', 'My own'], ['Meeting', 'Meetings'], ['Standup', 'Standups'], ['Task', 'Tasks'], ['Event', 'Events'], ['Deadline', 'Project deadlines']].filter(([k]) => (k !== 'Deadline' || has.projects) && ((k !== 'Meeting' && k !== 'Standup') || has.sessions)).map(([k, l]) => ({ label: l, color: typeColor[k], on: !s.calHide[k],
        toggle: () => this.setState((x) => ({ calHide: { ...x.calHide, [k]: !x.calHide[k] } })),
        style: 'height:32px;padding:0 12px;border-radius:16px;cursor:pointer;font-size:12px;display:flex;align-items:center;gap:6px;' + (s.calHide[k] ? 'border:1px dashed #3F524A;background:transparent;color:#9DB0A6' : 'border:1px solid #2A3832;background:#16201C;color:#E8EFEA') })),
      days: Array.from({ length: weeks * 7 }, (_, i) => {
        const iso = isoAdd(firstISO, i - offset);
        const inMonth = iso.slice(0, 7) === s.calMonth;
        const its = itemsOn(iso);
        const isSel = iso === s.calSel, isToday = iso === todayISO;
        return { n: Number(iso.slice(8, 10)), aria: fmtDay(iso) + ', ' + its.length + ' items',
          pick: () => this.setState({ calSel: iso }),
          style: 'min-height:108px;padding:6px;border:0;border-right:1px solid #24322C;border-bottom:1px solid #24322C;display:flex;flex-direction:column;gap:3px;text-align:left;cursor:pointer;font-family:inherit;color:#E8EFEA;overflow:hidden;min-width:0;background:' + (isSel ? '#22302A' : inMonth ? '#16201C' : '#121B17') + ';' + (inMonth ? '' : 'opacity:0.5'),
          numStyle: isToday ? 'align-self:flex-start;min-width:24px;height:24px;padding:0 6px;box-sizing:border-box;border-radius:12px;background:#B8F36B;color:#0F1714;font-weight:700;font-size:12px;display:flex;align-items:center;justify-content:center' : 'font-size:12px;font-weight:600;color:#C9D6CF;padding:2px 4px',
          items: its.slice(0, 3).map((e) => ({ label: (e.start ? e.start + ' ' : '') + e.title, style: chipStyleC(typeColor[e.type]) })),
          hasMore: its.length > 3, more: its.length - 3 };
      }),
      selLabel: (s.calSel === todayISO ? 'Today · ' : '') + fmtDay(s.calSel),
      agenda: itemsOn(s.calSel).map((e) => ({
        title: e.type === 'Deadline' ? 'Due: ' + e.title : e.title, type: e.type === 'Deadline' ? 'Project deadline' : e.type, color: typeColor[e.type],
        typeStyle: 'padding:2px 8px;border-radius:8px;font-size:11px;font-weight:600;background:' + typeColor[e.type] + '22;color:' + typeColor[e.type],
        time: e.start ? e.start + (e.end ? '–' + e.end : '') : 'All day',
        sub: [e.project, e.forum ? e.forum + ' forum' : '', e.repeat && e.repeat !== 'None' ? 'Repeats ' + e.repeat.toLowerCase() : '', e.notes || '', e.visibility ? 'Visible to: ' + e.visibility.toLowerCase() : '', e.remind && e.remind !== 'None' ? 'Reminder: ' + e.remind.toLowerCase() : ''].filter(Boolean).join(' · '),
        hasSub: !!(e.project || e.forum || (e.repeat && e.repeat !== 'None') || e.notes || e.visibility),
        people: e.people.slice(0, 6).map(personOf), hasPeople: e.people.length > 0, peopleLabel: e.people.length + (e.people.length === 1 ? ' person' : ' people'),
        canJoin: (e.type === 'Meeting' || e.type === 'Standup') && s.calSel === todayISO, join: () => joinEvent(e),
        canRemove: e.type !== 'Deadline', removeLabel: e.type === 'Task' ? 'Mark done' : 'Cancel',
        remove: () => { this.setState((x) => ({ events: x.events.filter((y) => y.id !== e.id) })); this.notify(e.type === 'Task' ? '"' + e.title + '" marked done.' : '"' + e.title + '" cancelled. Participants were told.', '#7CC4FF'); }
      }))
    };
    cal.noAgenda = cal.agenda.length === 0;
    const openEv = (type) => this.setState({ evOpen: true, ev: { title: '', type: type, date: s.calSel || todayISO, start: type === 'Standup' ? '09:15' : '10:00', end: type === 'Standup' ? '09:30' : '11:00', repeat: type === 'Standup' ? 'Weekdays' : 'None', forum: '', people: [], pq: '', notes: '', visibility: 'Only me', remind: '15 minutes before' } });
    const evS = s.ev;
    const setEv = (patch) => this.setState((x) => ({ ev: Object.assign({}, x.ev, patch) }));
    const evNeeds = evS.type === 'Meeting' || evS.type === 'Standup';
    const pqv = evS.pq.trim().toLowerCase();
    const candidates = pqv ? corePeople.concat(this.dir.map((d) => d.name)).filter((n, i, a) => a.indexOf(n) === i && n.toLowerCase().indexOf(pqv) !== -1).slice(0, 8) : corePeople.slice(0, 8);
    const titleOf = (n) => ((this.dir.find((d) => d.name === n) || {}).title) || ({ 'Brian Kiprop': 'Systems Engineer', 'David Otieno': 'Developer', 'Amina Wanjiru': 'Finance Analyst', 'Grace Njeri': 'HR Officer', 'Faith Achieng': 'Customer Care', 'Kevin Mutua': 'Sales Lead' }[n] || 'Staff');
    const ev = {
      heading: 'New ' + evS.type.toLowerCase(), title: evS.title, date: evS.date, start: evS.start, end: evS.end, repeat: evS.repeat, forum: evS.forum, pq: evS.pq,
      setTitle: (e) => setEv({ title: e.target.value }), setDate: (e) => setEv({ date: e.target.value }), setStart: (e) => setEv({ start: e.target.value }), setEnd: (e) => setEv({ end: e.target.value }),
      setRepeat: (e) => setEv({ repeat: e.target.value }), setForum: (e) => setEv({ forum: e.target.value }), setPq: (e) => setEv({ pq: e.target.value }),
      types: [['Personal', 'Just for me'], ['Task', 'Task'], ['Event', 'Event'], ['Meeting', 'Meeting'], ['Standup', 'Standup']].filter(([t]) => (t !== 'Meeting' && t !== 'Standup') || has.sessions).map(([t, l]) => ({ label: l, pick: () => setEv({ type: t, repeat: t === 'Standup' ? 'Weekdays' : evS.repeat }), style: chip(evS.type === t) })),
      needsPeople: evNeeds, isOwn: !evNeeds, count: evS.people.length,
      notes: evS.notes || '', visibility: evS.visibility || 'Only me', remind: evS.remind || 'None',
      setNotes: (e) => setEv({ notes: e.target.value }), setVisibility: (e) => setEv({ visibility: e.target.value }), setRemind: (e) => setEv({ remind: e.target.value }),
      chosen: evS.people.map((n) => ({ ...personOf(n), remove: () => setEv({ people: evS.people.filter((x) => x !== n) }) })),
      matches: candidates.map((n) => { const on = evS.people.indexOf(n) !== -1; return { ...personOf(n), title: titleOf(n), on: on, mark: on ? 'Added' : '+ Add',
        toggle: () => setEv({ people: on ? evS.people.filter((x) => x !== n) : evS.people.concat([n]) }),
        style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;cursor:pointer;color:#E8EFEA;font-family:inherit;background:' + (on ? '#22302A' : 'transparent') }; }),
      close: set({ evOpen: false }),
      saveLabel: evNeeds ? 'Schedule & send invites' : 'Add to calendar',
      saveStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (evS.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
      save: () => {
        const e = this.state.ev;
        if (!e.title.trim()) { this.notify('Give it a title first.', '#FFB86B'); return; }
        if ((e.type === 'Meeting' || e.type === 'Standup') && e.people.length === 0) { this.notify('Pick at least one participant.', '#FFB86B'); return; }
        const item = { id: 'e' + Date.now(), title: e.title.trim(), type: e.type, date: e.date || todayISO, start: e.start, end: e.end, people: (e.type === 'Meeting' || e.type === 'Standup') ? e.people : [], forum: e.forum, repeat: (e.type === 'Meeting' || e.type === 'Standup') ? e.repeat : 'None', by: 'Samwel Okoth Omulo', notes: (e.type === 'Meeting' || e.type === 'Standup') ? '' : (e.notes || '').trim(), visibility: (e.type === 'Meeting' || e.type === 'Standup') ? '' : e.visibility, remind: e.remind };
        this.setState((x) => ({ evOpen: false, events: x.events.concat([item]), calSel: item.date, calMonth: item.date.slice(0, 7) }));
        this.notify((e.type === 'Meeting' || e.type === 'Standup') ? e.type + ' scheduled for ' + fmtDay(item.date) + '. Invites sent to ' + e.people.length + (e.people.length === 1 ? ' person.' : ' people.') : 'Added to your calendar for ' + fmtDay(item.date) + '.', '#B8F36B');
      }
    };
    const calUpcoming = [];
    for (let k = 0; k < 14 && calUpcoming.length < 4; k++) { const iso = isoAdd(todayISO, k); itemsOn(iso).forEach((e) => { if (calUpcoming.length < 4) calUpcoming.push({ ...e, iso: iso }); }); }
    const calItems = calUpcoming.map((e) => ({ title: e.type === 'Deadline' ? 'Due: ' + e.title : e.title, color: typeColor[e.type], time: e.start || 'All day', day: e.iso === todayISO ? 'Today' : e.iso === isoAdd(todayISO, 1) ? 'Tomorrow' : fmtDay(e.iso).split(' ').slice(0, 2).join(' '), type: e.type === 'Deadline' ? 'Project deadline' : e.type, peopleLabel: e.people.length ? ' · ' + e.people.length + ' people' : '' }));
    const meetingEvents = s.events.filter((e) => e.type === 'Meeting' || e.type === 'Standup').slice().sort((a, b) => (a.date + a.start).localeCompare(b.date + b.start));
    const meetingsTable = { id: 'meetings', pageSize: 8, minWidth: '960px', emptyText: 'No meetings scheduled yet.',
      cols: [{ label: 'Meeting', filter: 'text' }, { label: 'Type', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Time' }, { label: 'Repeats', filter: 'select' }, { label: 'Participants' }, { label: 'Organizer', filter: 'select' }, { label: '', align: 'right' }],
      rows: meetingEvents.map((e) => ({ cells: [
        C.t(e.title, { strong: true, sub: e.forum ? e.forum + ' forum' : '' }), C.b(e.type, typeColor[e.type]), C.t(fmtDay(e.date), { fv: e.date, nowrap: true }),
        C.t(e.start + '–' + e.end, { nowrap: true, mono: true }), C.t(e.repeat === 'None' ? 'Once' : e.repeat),
        { kind: 'avatars', list: e.people.slice(0, 5).map(personOf), sub: e.people.length + (e.people.length === 1 ? ' person' : ' people'), fv: '' },
        C.t(e.by === 'Samwel Okoth Omulo' ? 'You' : e.by, { fv: e.by === 'Samwel Okoth Omulo' ? 'You' : e.by }),
        C.a([{ label: 'Join', kind: 'primary', run: () => joinEvent(e) }, { label: 'Cancel', kind: 'danger', run: () => { this.setState((x) => ({ events: x.events.filter((y) => y.id !== e.id) })); this.notify('"' + e.title + '" cancelled. Participants were told.', '#7CC4FF'); } }])
      ] })) };

    // =============== Organization structure ===============
    const O = this.org;
    const typeName = { sectors: 'Sector', branches: 'Branch', units: 'Unit', departments: 'Department' };
    const typeTint = { sectors: '#FFB86B', branches: '#7CC4FF', units: '#C9B6FF', departments: '#B8F36B' };
    const mgrRole = { departments: 'Head of department', sectors: 'Sector manager', branches: 'Branch manager', units: 'Unit lead' };
    const levels = ['sectors', 'branches', 'units', 'departments'].filter((t) => has[t]);
    const parentFor = (type, r) => {
      if (type === 'sectors') return null;
      if (type === 'branches') return has.sectors ? ['sectors', r.sector] : null;
      if (type === 'units') return has.branches ? ['branches', r.branch] : null;
      return has.branches ? ['branches', r.branch] : has.sectors ? ['sectors', r.sector] : null;
    };
    const recOf = (type, id) => (O[type] || []).find((x) => x.id === id);
    const pathOf = (type, r) => { const out = []; let p = parentFor(type, r); while (p) { const pr = recOf(p[0], p[1]); if (!pr) break; out.unshift(pr.name); p = parentFor(p[0], pr); } return ['[YOUR ORGANIZATION]'].concat(out).join(' › '); };
    const membersOf = (type, name) => this.dir.filter((p) => (type === 'departments' ? p.dept : type === 'branches' ? p.loc : type === 'sectors' ? p.sector : p.unit) === name);
    const shortOf = (n) => n.split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase();
    // ---- Ticket escalation ----
    const escSeed = (str) => { let h = 11; for (let i = 0; i < str.length; i++) h = (h * 33 + str.charCodeAt(i)) >>> 0; return h; };
    const deptCats = { 'IT Operations': ['Network & Wi-Fi', 'Hardware', 'Access request'], 'Software Development': ['Software bug', 'Change request'], 'HR': ['HR request', 'Payroll query'], 'Finance': ['Payment query', 'Budget approval'], 'Risk & Compliance': ['Security incident', 'Compliance query'], 'Customer Care': ['Customer complaint'], 'Legal': ['Contract review'] };
    const escLevels = ['departments', 'units', 'branches', 'sectors'].filter((t) => has[t]);
    const escTypes = escLevels.length ? escLevels : ['desk'];
    const escRecs = (type) => type === 'desk' ? [{ id: 'desk', name: 'Ticket team' }] : (O[type] || []).filter((r) => !(cur_t && type === 'departments' && r.name === cur_t.dept));
    const escCats = (type, r) => type === 'desk' ? ts.cats.map((c) => c.name) : type === 'departments' ? (deptCats[r.name] || [r.name + ' request', 'General enquiry']) : ['Escalated tickets', 'Service complaint', 'Approval & sign-off'];
    const escPeople = (type, r, cat) => {
      const pool = type === 'desk' ? this.dir.filter((p) => tagNames.indexOf(p.name) !== -1) : membersOf(type, r.name);
      if (!pool.length) return [];
      const named = pool.filter((p) => tagNames.indexOf(p.name) !== -1);
      const h = escSeed(r.name + cat), n = Math.min(pool.length, 3);
      const out = named.slice(0, 1);
      for (let i = 0; out.length < n && i < pool.length; i++) { const p = pool[(h + i * 7) % pool.length]; if (out.indexOf(p) === -1) out.push(p); }
      return out;
    };
    const ES = s.esc;
    const escType = escTypes.indexOf(ES.type) !== -1 ? ES.type : escTypes[0];
    const escRecList = escRecs(escType);
    const escRec = escRecList.find((r) => r.id === ES.id) || (escType === 'desk' ? escRecList[0] : null);
    const escCatList = escRec ? escCats(escType, escRec) : [];
    const escCat = escCatList.indexOf(ES.cat) !== -1 ? ES.cat : '';
    const escPpl = escRec && escCat ? escPeople(escType, escRec, escCat) : [];
    const escPerson = escPpl.find((p) => p.name === ES.person) ? ES.person : '';
    const setEsc = (patch) => this.setState((st) => ({ esc: Object.assign({}, st.esc, patch) }));
    const escLabel = (t) => (t === 'desk' ? 'Ticket team' : typeName[t]);
    const escReady = !!(escRec && escCat && escPerson);
    const esc = {
      open: s.escOpen, btnLabel: s.escOpen ? 'Cancel' : 'Escalate ticket', canEscalate: !!cur_t && cur_t.step < 3,
      toggle: () => this.setState((st) => ({ escOpen: !st.escOpen, esc: { type: '', id: '', cat: '', person: '', why: '' } })),
      types: escTypes.map((t) => ({ label: escLabel(t), pick: () => setEsc({ type: t, id: '', cat: '', person: '' }), style: 'height:34px;padding:0 12px;border-radius:17px;cursor:pointer;font-size:13px;' + (t === escType ? 'border:1px solid #FF9B8A;background:#3A2420;color:#FFD3CA;font-weight:600' : 'border:1px solid #2A3832;background:#121B17;color:#C9D6CF') })),
      typeLabel: escLabel(escType), showRec: escType !== 'desk',
      recs: escRecList.map((r) => ({ id: r.id, name: r.name })), recValue: escRec ? escRec.id : '',
      onRec: (e) => setEsc({ id: e.target.value, cat: '', person: '' }),
      hasRec: !!escRec, cats: escCatList.map((c) => ({ name: c, tat: tatText(c) })), catValue: escCat,
      onCat: (e) => setEsc({ cat: e.target.value, person: '' }),
      hasCat: !!escCat, people: escPpl.map((p) => ({ name: p.name, title: p.title })), personValue: escPerson,
      noPeople: !!escCat && escPpl.length === 0,
      onPerson: (e) => setEsc({ person: e.target.value }),
      why: ES.why, onWhy: (e) => setEsc({ why: e.target.value }),
      btnStyle: 'height:42px;border-radius:10px;border:0;font-weight:600;cursor:pointer;font-size:14px;' + (escReady ? 'background:#FF9B8A;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
      submit: () => {
        if (!cur_t || !escReady) { this.notify('Pick where to escalate, the category and the person.', '#FFB86B'); return; }
        const dest = escType === 'desk' ? 'the ticket team' : typeName[escType] + ' ' + escRec.name;
        const why = (this.state.esc.why || '').trim();
        this.updTicket(cur_t.id, (t) => Object.assign({}, t, {
          escalated: true, escalatedTo: (escType === 'desk' ? 'Ticket team' : typeName[escType] + ': ' + escRec.name) + ' · ' + escCat,
          dept: escType === 'departments' ? escRec.name : t.dept, cat: escCat, handler: escPerson, handlerSince: 'Just now',
          assignees: t.assignees.indexOf(escPerson) === -1 ? t.assignees.concat([escPerson]) : t.assignees, step: Math.max(1, Math.min(t.step, 2)), hold: false,
          activity: this.act(t, { text: 'Escalated to ' + dest + ' (' + escCat + '), assigned to ' + escPerson + (why ? '. Reason: ' + why : '') })
        }));
        this.setState({ escOpen: false, esc: { type: '', id: '', cat: '', person: '', why: '' } });
        this.notify(cur_t.id + ' escalated to ' + dest + '. ' + escPerson + ' has been notified.', '#FF9B8A', cur_t.id);
      }
    };
    const seedOf = (str) => { let h = 7; for (let i = 0; i < str.length; i++) h = (h * 31 + str.charCodeAt(i)) >>> 0; return h; };
    const unitData = (type, r) => {
      let h = seedOf(type + r.id);
      const rnd = () => { h = (h * 1664525 + 1013904223) >>> 0; return h / 4294967296; };
      const pick = (arr) => arr[Math.floor(rnd() * arr.length)];
      const pOn = 0.86 + (seedOf(r.id + 'p') % 14) / 100;
      const others = levels.reduce((acc, t) => acc.concat(O[t].filter((x) => x.name !== r.name).map((x) => x.name)), []);
      const cps = others.length ? others : ['Head Office', 'Finance', 'IT Operations'];
      const comms = Array.from({ length: 26 }, (_, i) => {
        const dirn = rnd() < 0.6 ? 'Received' : 'Sent';
        const ok = rnd() < pOn;
        return { date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, type: pick(['Memo', 'Announcement', 'Newsletter', 'Email']), subject: pick(['Updated leave policy', 'Quarter-end close dates', 'System maintenance window', 'New password rules', 'Branch opening hours', 'Audit evidence request', 'Customer complaint trends', 'Training schedule', 'Budget submission reminder', 'Holiday rota']), cp: pick(cps),
          status: dirn === 'Received' ? (ok ? 'Acknowledged on time' : pick(['Acknowledged late', 'Not acknowledged'])) : pick(['Delivered', 'Read by all', 'Partly read']), ok: ok };
      });
      const reqs = Array.from({ length: 22 }, (_, i) => {
        const x = rnd(); const dirn = x < 0.55 ? 'Received' : x < 0.85 ? 'Sent' : 'Escalated';
        const ok = rnd() < pOn;
        return { ref: 'BR-' + (1000 + (seedOf(r.id) % 800) + i), date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, subject: pick(['Approve purchase of laptops', 'Customer refund review', 'Access to finance reports', 'Change to branch rota', 'New vendor onboarding', 'Loan document check', 'Office supplies order', 'Fix branch signage', 'Data correction request', 'Budget reallocation']), cp: pick(cps),
          status: dirn === 'Sent' ? pick(['Resolved on time', 'In progress', 'Waiting on them']) : (ok ? 'Resolved on time' : pick(['Resolved late', 'Overdue', 'In progress'])), ok: ok };
      });
      let tks = Array.from({ length: 14 }, (_, i) => {
        const x = rnd(); const dirn = x < 0.55 ? 'Received' : x < 0.85 ? 'Sent' : 'Escalated';
        const ok = rnd() < pOn;
        return { id: 'TK-' + (500 + (seedOf(r.id + 't') % 400) + i), date: isoAdd(todayISO, -Math.floor(rnd() * 30)), dir: dirn, title: pick(['Printer not working', 'Cannot log in to email', 'Slow network', 'Request new monitor', 'Report not loading', 'Password reset', 'Update contact list', 'Phone line down']), cp: pick(cps),
          status: dirn === 'Sent' ? pick(['Solved on time', 'In progress', 'Assigned']) : (ok ? 'Solved on time' : pick(['Solved late', 'Overdue', 'In progress'])), ok: ok, real: null };
      });
      if (type === 'departments') {
        const real = tickets.filter((t) => t.dept === r.name || t.fromDept === r.name).map((t) => {
          const dirn = t.escalated && t.dept === r.name ? 'Escalated' : t.dept === r.name ? 'Received' : 'Sent';
          return { id: t.id, date: t.createdAt, dir: dirn, title: t.title, cp: dirn === 'Sent' ? t.dept : t.fromDept, status: t.slaLabel === 'Overdue' ? 'Overdue' : t.statusLabel, ok: t._c.sla !== 'Overdue' && t._c.sla !== 'Solved late', real: t };
        });
        tks = real.concat(tks);
      }
      const recv = (arr) => arr.filter((x) => x.dir !== 'Sent');
      const pct = (arr) => { const rr = recv(arr); return rr.length ? Math.round(rr.filter((x) => x.ok).length / rr.length * 100) : 100; };
      const all = recv(comms).concat(recv(reqs)).concat(recv(tks));
      const perf = all.length ? Math.round(all.filter((x) => x.ok).length / all.length * 100) : 100;
      return { comms, reqs, tks, perf, pc: pct(comms), pr: pct(reqs), pt: pct(tks) };
    };
    const targetOf = (key) => s.targets[key] || 95;
    const perfColor = (v, tg) => (v >= tg ? '#B8F36B' : v >= tg - 10 ? '#FFB86B' : '#FF9B8A');
    const openOu = (type, id) => () => this.setState({ section: 'orgunit', ou: type + ':' + id, ouTab: 'comms', ticketOpen: null, projectOpen: null });
    const orgRows = [{ depth: 0, type: 'Organization', name: '[YOUR ORGANIZATION]', sub: this.dir.length + ' people · ' + levels.map((t) => O[t].length + ' ' + typeName[t].toLowerCase() + (O[t].length === 1 ? '' : (t === 'branches' ? 'es' : 's'))).join(' · '), color: '#E8EFEA', canOpen: false }];
    const walk = (ptype, pid, depth) => {
      levels.forEach((t) => O[t].forEach((r) => {
        const p = parentFor(t, r);
        const match = ptype === null ? p === null : (p && p[0] === ptype && p[1] === pid);
        if (!match) return;
        orgRows.push({ depth: depth, type: typeName[t], name: r.name, sub: mgrRole[t] + ' · ' + membersOf(t, r.name).length + ' people', color: typeTint[t], canOpen: true, open: openOu(t, r.id) });
        walk(t, r.id, depth + 1);
      }));
    };
    if (has.structure) walk(null, null, 1);
    const org = {
      chain: ['Organization'].concat(levels.filter((t) => t !== 'departments').map((t) => typeName[t] + (t === 'branches' ? 'es' : 's'))).map((l, i, a) => ({ label: l, arrow: i < a.length - 1, style: 'padding:6px 12px;border-radius:16px;background:#16201C;border:1px solid #2A3832;font-size:13px' })),
      deptNote: has.departments, deptNoteText: 'Departments sit under ' + (has.branches ? 'their branch' : has.sectors ? 'their sector' : 'the organization') + '.',
      rows: orgRows.map((o, i) => ({ ...o, rowStyle: 'display:flex;flex-wrap:wrap;align-items:center;gap:12px;padding:12px 16px;' + (i ? 'border-top:1px solid #24322C' : 'background:#121B17'),
        indent: 'width:' + (o.depth * 26) + 'px;flex:none;align-self:stretch;' + (o.depth ? 'border-right:2px solid #2A3832' : ''), badge: badge(o.color) }))
    };
    const allUnits = levels.reduce((acc, t) => acc.concat(O[t].map((r) => ({ type: t, r: r }))), []);
    const unitPerf = {};
    allUnits.forEach((u) => { unitPerf[u.type + ':' + u.r.id] = unitData(u.type, u.r); });
    const myUnits = ['departments', 'branches', 'sectors', 'units'].filter((t) => has[t]).map((t) => {
      const r = recOf(t, O.mine[t]);
      const key = t + ':' + r.id;
      const on = s.section === 'orgunit' && s.ou === key;
      return { name: r.name, type: typeName[t], typeLower: typeName[t].toLowerCase(), short: shortOf(r.name), color: typeTint[t], key: key, perf: unitPerf[key].perf, open: openOu(t, r.id),
        style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:46px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;color:#E8EFEA;font-family:inherit;background:' + (on ? '#22302A' : 'transparent') };
    });
    const roleNameOf = (id) => (roleDefs.find((r) => r.id === id) || {}).name || id;
    const targetLabel = (lvl, k) => (k === 'own' ? 'its own ' + typeName[lvl].toLowerCase() : k === 'wall' ? 'the Community wall' : 'other ' + (typeName[k] + (k === 'branches' ? 'es' : 's')).toLowerCase());
    const ruleSentence = (lvl, ty) => {
      const r = (s.commRules[lvl] || {})[ty];
      if (!r || !r.on) return ty + 's: switched off.';
      const to = r.to.filter((k) => k === 'own' || k === 'wall' || has[k]).map((k) => targetLabel(lvl, k));
      return ty + 's go to ' + (to.length ? to.join(', ') : 'nowhere yet') + '. Sent by ' + roleNameOf(r.sender) + '. ' + (r.approver === 'none' ? 'No approval.' : 'Approved by ' + roleNameOf(r.approver) + '.');
    };
    let ou = {}, ouTable = { id: 'none', cols: [], rows: [] };
    const ouTab = (s.ouTab === 'tks' && !has.tickets) || (s.ouTab === 'projects' && !has.projects) ? 'comms' : s.ouTab;
    const ouParts = s.ou ? s.ou.split(':') : null;
    const ouRec = ouParts && has[ouParts[0]] ? recOf(ouParts[0], ouParts[1]) : null;
    if (ouRec) {
      const t = ouParts[0], key = s.ou, D = unitPerf[key] || unitData(t, ouRec), tg = targetOf(key);
      const mem = membersOf(t, ouRec.name);
      const lead = mem.find((p) => /Lead|Supervisor|Manager/.test(p.title)) || mem[0];
      const dirC = { Received: '#7CC4FF', Sent: '#C9B6FF', Escalated: '#FF9B8A' };
      const stC = (st) => (/on time|Delivered|Read by all|Solved on time/.test(st) ? '#B8F36B' : /late|Partly|Waiting|In progress|Assigned/.test(st) ? '#FFB86B' : /Overdue|Not ack/.test(st) ? '#FF9B8A' : '#9DB0A6');
      const cnt = (arr, d) => arr.filter((x) => x.dir === d).length;
      const deptProj = { 'IT Operations': ['network'], 'Software Development': ['mobile'], 'HR': ['onboard'], 'Risk & Compliance': ['audit'], 'Finance': ['audit'], 'Operations': ['network'], 'Customer Care': ['mobile'] };
      const pids = t === 'departments' ? (deptProj[ouRec.name] || []) : s.projects.filter((p, i) => (seedOf(ouRec.id) + i) % 2 === 0).map((p) => p.id);
      const ps = projects.filter((p) => pids.indexOf(p.id) !== -1);
      const mk = (label, pc, detail) => ({ label: label, pct: pc, detail: detail, color: perfColor(pc, tg), bar: 'height:100%;width:' + pc + '%;background:' + perfColor(pc, tg) + ';border-radius:5px', marker: 'position:absolute;top:-3px;left:calc(' + tg + '% - 1px);width:2px;height:16px;background:#E8EFEA;border-radius:1px' });
      ou = {
        name: ouRec.name, type: typeName[t], short: shortOf(ouRec.name), color: typeTint[t], badge: badge(typeTint[t]), path: pathOf(t, ouRec),
        managerRole: mgrRole[t], manager: lead ? lead.name : 'Not set', members: mem.length,
        typePlural: typeName[t] + (t === 'branches' ? 'es' : 's'),
        commRules: ['Memo', 'Announcement'].map((ty) => ({ text: ruleSentence(t, ty) })),
        sendMemo: () => this.setState({ cmOpen: true, cmd: { type: has.memos ? 'Memo' : 'Announcement', title: '', body: '', wall: false, forum: '', sel: { [t + ':' + ouRec.name]: true } } }),
        kpis: [
          { label: 'Communications', parts: [{ n: cnt(D.comms, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.comms, 'Sent'), l: 'Sent', color: '#C9B6FF' }] },
          { label: 'Business requests', parts: [{ n: cnt(D.reqs, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.reqs, 'Sent'), l: 'Sent', color: '#C9B6FF' }, { n: cnt(D.reqs, 'Escalated'), l: 'Escalated', color: '#FF9B8A' }] },
          { label: 'Tickets', parts: [{ n: cnt(D.tks, 'Received'), l: 'Received', color: '#7CC4FF' }, { n: cnt(D.tks, 'Sent'), l: 'Sent', color: '#C9B6FF' }, { n: cnt(D.tks, 'Escalated'), l: 'Escalated', color: '#FF9B8A' }] },
          { label: 'Projects', parts: [{ n: ps.length, l: 'Active', color: '#E8EFEA' }, { n: ps.filter((p) => p.status !== 'On track').length, l: 'Need attention', color: '#FFB86B' }] }
        ],
        target: tg, setTarget: (e) => { const v = Math.max(50, Math.min(100, Number(e.target.value) || 95)); this.setState((x) => ({ targets: { ...x.targets, [key]: v } })); },
        perf: D.perf, perfColor: perfColor(D.perf, tg), perfNote: D.perf >= tg ? 'Target met' : (tg - D.perf) + ' points below target',
        breakdown: [mk('Communications acknowledged on time', D.pc, D.comms.filter((x) => x.dir !== 'Sent').length + ' received'), mk('Business requests resolved on time', D.pr, D.reqs.filter((x) => x.dir !== 'Sent').length + ' received'), mk('Tickets resolved on time', D.pt, D.tks.filter((x) => x.dir !== 'Sent').length + ' received')],
        tabs: [['comms', 'Communications', D.comms.length], ['reqs', 'Business requests', D.reqs.length], ['tks', 'Tickets', D.tks.length], ['projects', 'Projects', ps.length]].map(([k, l, n]) => ({ label: l, count: n, go: () => this.setState({ ouTab: k }),
          style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;white-space:nowrap;font-family:inherit;border-bottom:2px solid ' + (ouTab === k ? '#B8F36B' : 'transparent') + ';color:' + (ouTab === k ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (ouTab === k ? '600' : '500') }))
      };
      const ouOff = (l) => (/Ticket/.test(l) && !has.tickets) || (/Project/.test(l) && !has.projects);
      ou.kpis = ou.kpis.filter((k) => !ouOff(k.label)); ou.tabs = ou.tabs.filter((k) => !ouOff(k.label)); ou.breakdown = ou.breakdown.filter((k) => !ouOff(k.label));
      const dirCell = (d) => C.b(d, dirC[d]);
      if (ouTab === 'comms') ouTable = { id: 'ou-c-' + key, pageSize: 10, minWidth: '900px', cols: [{ label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Type', filter: 'select' }, { label: 'Subject', filter: 'text' }, { label: 'From / to', filter: 'text' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
        rows: D.comms.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ cells: [C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.type), C.t(c.subject, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)), C.a([{ label: 'Open', run: () => this.notify('Opening "' + c.subject + '".', '#7CC4FF') }])] })) };
      if (ouTab === 'reqs') ouTable = { id: 'ou-r-' + key, pageSize: 10, minWidth: '940px', cols: [{ label: 'Ref', filter: 'text' }, { label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Request', filter: 'text' }, { label: 'With', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
        rows: D.reqs.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ tint: c.status === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(c.ref, { mono: true }), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.subject, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)),
          C.a([{ label: 'View', run: () => this.notify('Opening ' + c.ref + '.', '#7CC4FF') }].concat(c.dir === 'Received' && c.status !== 'Resolved on time' && c.status !== 'Resolved late' ? [{ label: 'Escalate', kind: 'danger', run: () => this.notify(c.ref + ' escalated to the ' + (has.branches ? 'branch manager' : 'head of department') + '.', '#FF9B8A') }] : []))] })) };
      if (ouTab === 'tks') ouTable = { id: 'ou-t-' + key, pageSize: 10, minWidth: '900px', cols: [{ label: 'Ticket', filter: 'text' }, { label: 'Date', filter: 'date' }, { label: 'Direction', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'With', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
        rows: D.tks.slice().sort((a, b) => b.date.localeCompare(a.date)).map((c) => ({ tint: c.status === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(c.id, { mono: true }), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }), dirCell(c.dir), C.t(c.title, { strong: true }), C.t(c.cp), C.b(c.status, stC(c.status)),
          C.a([{ label: 'View', run: c.real ? () => this.setState({ section: 'tickets', ticketOpen: c.id }) : () => this.notify('Opening ' + c.id + '.', '#7CC4FF') }])] })) };
      if (ouTab === 'projects') ouTable = { id: 'ou-p-' + key, pageSize: 10, minWidth: '760px', emptyText: 'No projects for this ' + typeName[t].toLowerCase() + ' yet.', cols: [{ label: 'Project', filter: 'text' }, { label: 'Status', filter: 'select' }, { label: 'Progress' }, { label: 'Due', filter: 'date' }, { label: '', align: 'right' }],
        rows: ps.map((p) => ({ cells: [C.t(p.name, { strong: true, sub: 'Led by ' + p.owner }), C.b(p.status, p.statusColor), C.bar(p.progress, p.statusColor), C.t(p.dueLabel, { fv: p.due, nowrap: true }), C.a([{ label: 'View', kind: 'primary', run: p.open }])] })) };
    }
    const orgWidgetItems = {};
    ['departments', 'units'].forEach((t) => { if (!has[t]) return; orgWidgetItems[t] = O[t].map((r) => { const k = t + ':' + r.id; const pf = unitPerf[k].perf; const pp = parentFor(t, r); return { name: r.name, sub: (pp ? recOf(pp[0], pp[1]).name + ' · ' : '') + membersOf(t, r.name).length + ' people', perf: pf, perfColor: perfColor(pf, targetOf(k)), open: openOu(t, r.id) }; }); });
    if (has.branches) orgWidgetItems.branches = O.branches.map((b) => ({ name: b.name, initials: shortOf(b.name), place: has.sectors ? recOf('sectors', b.sector).name + ' sector' : 'Branch', head: (membersOf('branches', b.name).find((p) => /Supervisor|Lead|Manager/.test(p.title)) || {}).name || 'Not set', staff: membersOf('branches', b.name).length, statusColor: '#B8F36B', statusLabel: unitPerf['branches:' + b.id].perf + '% on time', open: openOu('branches', b.id) }));
    if (has.sectors) orgWidgetItems.sectors = O.sectors.map((sc, i) => ({ name: sc.name, color: ['#FFB86B', '#7CC4FF', '#F49AC1'][i % 3], depts: (has.branches ? O.branches.filter((b) => b.sector === sc.id) : O.departments.filter((d) => d.sector === sc.id)).map((x) => x.name).join(' · '), staff: membersOf('sectors', sc.name).length, open: openOu('sectors', sc.id) }));
    orgWidgetItems.performance = myUnits.map((u) => { const tg = targetOf(u.key); return { name: u.name, type: u.type, perf: u.perf, color: perfColor(u.perf, tg), open: u.open, bar: 'height:100%;width:' + u.perf + '%;background:' + perfColor(u.perf, tg) + ';border-radius:4px', marker: 'position:absolute;top:-3px;left:calc(' + tg + '% - 1px);width:2px;height:14px;background:#E8EFEA' }; });

    // =============== Communications ===============
    const commTypes = [has.announcements && 'Announcement', has.memos && 'Memo', has.newsletter && 'Newsletter'].filter(Boolean);
    const commColor = { Announcement: '#B8F36B', Memo: '#FFD66B', Newsletter: '#F49AC1' };
    const openCm = (type) => this.setState({ cmOpen: true, cmd: { type: type, title: '', body: '', wall: type !== 'Memo', forum: '', sel: {} } });
    const cmS = s.cmd;
    const setCm = (patch) => this.setState((x) => ({ cmd: Object.assign({}, x.cmd, patch) }));
    const f1 = flowById('f1');
    // Communication rules: the user's own level decides where memos/announcements may go
    const ruleLevel = ['departments', 'units', 'branches', 'sectors'].find((t) => has[t]);
    const ruleObj = ruleLevel && (cmS.type === 'Memo' || cmS.type === 'Announcement') ? (s.commRules[ruleLevel] || {})[cmS.type] : null;
    const ruleRec = ruleLevel ? recOf(ruleLevel, O.mine[ruleLevel]) : null;
    const ruleTo = ruleObj ? (ruleObj.on ? ruleObj.to : []) : null;
    const senderOk = !ruleObj || ['admin', 'superadmin'].indexOf(s.actingRole) !== -1 || s.actingRole === ruleObj.sender;
    const ruleNeeds = !!(ruleObj && ruleObj.on && ruleObj.approver !== 'none');
    const wallAllowed = !ruleTo || ruleTo.indexOf('wall') !== -1;
    const cmNeedsApproval = ruleNeeds || (cmS.type === 'Announcement' && f1 && f1.levels > 0);
    const cm = {
      heading: 'New ' + cmS.type.toLowerCase(), title: cmS.title, body: cmS.body, wall: cmS.wall && wallAllowed, forum: cmS.forum, wallAllowed: wallAllowed,
      types: commTypes.map((t) => ({ label: t, pick: () => setCm({ type: t }), style: chip(cmS.type === t) })),
      setTitle: (e) => setCm({ title: e.target.value }), setBody: (e) => setCm({ body: e.target.value }),
      toggleWall: () => setCm({ wall: !this.state.cmd.wall }), setForum: (e) => setCm({ forum: e.target.value }),
      groups: ['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]).map((t) => {
        let recs = O[t];
        if (ruleTo) recs = ruleTo.indexOf(t) !== -1 ? O[t] : (ruleTo.indexOf('own') !== -1 && t === ruleLevel ? [ruleRec] : []);
        return { label: typeName[t] + (t === 'branches' ? 'es' : 's'),
          items: recs.map((r) => { const k = t + ':' + r.name; const on = !!cmS.sel[k]; return { name: r.name, on: on, toggle: () => setCm({ sel: Object.assign({}, this.state.cmd.sel, { [k]: !on }) }), style: chipS(on) }; }) };
      }).filter((g) => g.items.length),
      hasRuleNote: !!ruleObj,
      ruleNote: ruleObj ? (ruleObj.on ? 'Rule for ' + ruleRec.name + ' (' + typeName[ruleLevel].toLowerCase() + '): ' + ruleSentence(ruleLevel, cmS.type) + (senderOk ? '' : ' Your role is not the sending role, so you cannot send this.') : cmS.type + 's from ' + typeName[ruleLevel].toLowerCase() + 's are switched off in the ' + typeName[ruleLevel] + (ruleLevel === 'branches' ? 'es' : 's') + ' widget settings.') : '',
      noStructure: !has.structure,
      needsApproval: cmNeedsApproval, approvalNote: ruleNeeds ? 'This needs approval by ' + roleNameOf(ruleObj.approver) + ' before it goes out.' : cmNeedsApproval ? 'Announcements need approval (' + (f1.levels === 1 ? 'direct' : f1.levels + ' levels') + ') before they go out.' : '',
      publishLabel: cmNeedsApproval ? 'Send for approval' : 'Publish',
      publishStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (cmS.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
      close: set({ cmOpen: false }),
      publish: () => {
        const c = this.state.cmd;
        if (!c.title.trim()) { this.notify('Add a title first.', '#FFB86B'); return; }
        if (ruleObj && !ruleObj.on) { this.notify(c.type + 's are switched off for your ' + typeName[ruleLevel].toLowerCase() + '.', '#FF9B8A'); return; }
        if (!senderOk) { this.notify('Only the ' + roleNameOf(ruleObj.sender) + ' role can send ' + c.type.toLowerCase() + 's here.', '#FF9B8A'); return; }
        const fo = forums.find((f) => f.id === c.forum);
        const to = (c.wall && wallAllowed ? ['Community wall'] : []).concat(fo ? [fo.name + ' forum'] : []).concat(Object.keys(c.sel).filter((k) => c.sel[k]).map((k) => k.split(':')[1]));
        if (!to.length) { this.notify('Choose at least one place to send it.', '#FFB86B'); return; }
        const id = 'C-' + (211 + this.state.comms.length);
        const needs = ruleNeeds || (c.type === 'Announcement' && f1 && f1.levels > 0);
        const item = { id: id, type: c.type, title: c.title.trim(), body: c.body.trim(), by: 'Samwel Okoth Omulo', date: todayISO, to: to, status: needs ? 'Awaiting approval' : 'Published', read: 0 };
        const patch = { cmOpen: false, comms: [item].concat(this.state.comms) };
        if (!needs && c.wall) patch.posts = [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: c.type + ' · Just now', text: item.title + (item.body ? '\n' + item.body : ''), likes: 0, extra: 0, shares: 0, comments: [] }].concat(this.state.posts);
        this.setState(patch);
        if (needs && ruleNeeds) {
          const apName = roleNameOf(ruleObj.approver);
          this.setState((x) => ({ flows: x.flows.map((fl) => (fl.id === 'f10' ? { ...fl, levels: 1, approvers: [apName, fl.approvers[1], fl.approvers[2]] } : fl)) }));
          makeRequest('f10', c.type + ' from ' + ruleRec.name + ': ' + item.title, { kind: 'comm', commId: id });
        } else if (needs) makeRequest('f1', c.type + ': ' + item.title, { kind: 'comm', commId: id });
        else { if (fo) this.addForumMsg(fo.id, c.type + ': ' + item.title); this.notify(c.type + ' published to ' + to.join(', ') + '.', '#B8F36B'); }
      }
    };
    const commNew = commTypes.map((t, i) => ({ label: 'New ' + t.toLowerCase(), run: () => openCm(t), style: 'height:40px;padding:0 16px;border-radius:20px;font-weight:600;cursor:pointer;' + (i === 0 ? 'border:0;background:#B8F36B;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA') }));
    const commsShown = s.comms.filter((c) => commTypes.indexOf(c.type) !== -1);
    const commsTable = { id: 'comms', pageSize: 10, minWidth: '1000px', emptyText: 'Nothing sent yet.',
      cols: [{ label: 'Type', filter: 'select' }, { label: 'Title', filter: 'text' }, { label: 'Sent to', filter: 'text' }, { label: 'From', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: 'Read' }, { label: '', align: 'right' }],
      rows: commsShown.map((c) => ({ cells: [C.b(c.type, commColor[c.type]), C.t(c.title, { strong: true, sub: c.body }), C.t(c.to.join(', ')), C.t(c.by === 'Samwel Okoth Omulo' ? 'You' : c.by), C.t(fmtDay(c.date), { fv: c.date, nowrap: true }),
        C.b(c.status, c.status === 'Published' ? '#B8F36B' : '#FFB86B'), C.bar(c.read, c.read >= 80 ? '#B8F36B' : '#FFB86B'),
        C.a(c.status === 'Published' ? [{ label: 'Remind unread', run: () => this.notify('Reminder sent to everyone who has not read "' + c.title + '".', '#7CC4FF') }] : [{ label: 'See approval', kind: 'primary', run: () => this.setState({ section: 'approvals' }) }])] })) };

    // =============== Tables for existing pages ===============
    const ticketsTable = { id: 'tickets', pageSize: 10, minWidth: has.departments ? '1180px' : '1060px', emptyText: 'No tickets match.',
      cols: [{ label: 'ID', filter: 'text' }, { label: 'Title', filter: 'text' }].concat(has.departments ? [{ label: 'Department', filter: 'select' }, { label: 'From', filter: 'select' }] : []).concat([{ label: 'Category', filter: 'select' }, { label: 'Priority', filter: 'select' }, { label: 'Status', filter: 'select' }, { label: 'TAT', filter: 'select' }, { label: 'Handling', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: '', align: 'right' }]),
      rows: tickets.map((t) => ({ tint: t._c.sla === 'Overdue' ? '#FF9B8A' : '', cells: [C.t(t.id, { mono: true, nowrap: true }), C.t(t.title, { strong: true })].concat(has.departments ? [C.t(t.dept), C.t(t.fromDept)] : []).concat([
        C.t(t.cat), C.b(t.prio, { High: '#FF9B8A', Medium: '#FFB86B', Low: '#9DB0A6' }[t.prio]), C.t(t.statusLabel + (t.escalated ? ' · escalated' : ''), { fv: t.statusLabel }), C.b(t.slaLabel, t._c.color, { sub: t.remainingLabel }), C.t(t.handlerLabel), C.t(fmtDay(t.createdAt), { fv: t.createdAt, nowrap: true }),
        C.a([{ label: 'View', kind: 'primary', run: t.open }])]) })) };
    const statusColP = { 'Available': '#B8F36B', 'In a meeting': '#FFB86B', 'Focusing': '#7CC4FF', 'Away': '#9DB0A6' };
    const peopleTable = { id: 'people', pageSize: 15, minWidth: '880px',
      cols: [{ label: 'Person', filter: 'text' }].concat(has.departments ? [{ label: 'Department', filter: 'select' }] : []).concat(has.sectors ? [{ label: 'Sector', filter: 'select' }] : []).concat(has.branches ? [{ label: 'Branch', filter: 'select' }] : []).concat(has.units ? [{ label: 'Unit', filter: 'select' }] : []).concat([{ label: 'Status', filter: 'select' }, { label: '', align: 'right' }]),
      rows: this.dir.map((p) => ({ cells: [C.p(p.name, { sub: p.title, dot: statusColP[p.status], fv: p.name })].concat(has.departments ? [C.t(p.dept)] : []).concat(has.sectors ? [C.t(p.sector)] : []).concat(has.branches ? [C.t(p.loc)] : []).concat(has.units ? [C.t(p.unit || '—', { fv: p.unit || 'No unit' })] : []).concat([
        C.b(p.status, statusColP[p.status]),
        C.a((has.chat ? [{ label: 'Message', run: () => this.openDm(p.name) }, { label: 'Call', run: () => this.startCall('voice', 'Private call with ' + p.name, p.initials, p.color, [p.name]) }] : []).concat([{ label: 'Profile', kind: 'primary', run: () => this.setState({ personOpen: p.id }) }]))]) })) };
    const peopleSummary = this.dir.length + ' people' + (has.departments ? ' in ' + O.departments.length + ' departments' : '') + (has.branches ? ' across ' + O.branches.length + ' branches' : '') + ' · ' + this.dir.filter((p) => p.status === 'Available').length + ' available now';
    const isoFromAgo = (m) => new Date(nowMs - m * 60000).toLocaleDateString('en-CA', tz9);
    const logsTable = { id: 'logs', pageSize: 12, minWidth: '1120px', emptyText: 'No log entries match.',
      cols: [{ label: 'When', filter: 'date' }, { label: 'User', filter: 'text' }, { label: 'Action', filter: 'text' }, { label: 'Module', filter: 'select' }, { label: 'IP · Location', filter: 'text' }, { label: 'Result', filter: 'select' }, { label: 'Flag', filter: 'select' }, { label: '', align: 'right' }],
      rows: allLogs.map((l) => {
        const fl = !!(l.rule && ruleOn(l.rule)); const r = fl ? ruleDefs.find((x) => x.id === l.rule) : null; const c = fl ? sevColor[ruleSev(r)] : '';
        return { tint: fl ? c : '', cells: [C.t(l.ago === 0 ? 'Just now' : agoLabel(l.ago), { fv: isoFromAgo(l.ago), nowrap: true, mono: true }), C.t(l.user, { nowrap: true }), C.t(l.action), C.t(l.module), C.t(l.ip + ' · ' + l.loc, { nowrap: true }),
          C.b(l.result, l.result === 'Success' ? '#9DB0A6' : l.result === 'Failed' ? '#FF9B8A' : '#FFB86B'), fl ? C.b(r.name, c, { fv: 'Flagged: ' + r.name }) : C.t('—', { fv: 'Not flagged', color: '#4A5C54' }),
          C.a(fl ? [{ label: 'View alert', kind: 'primary', run: () => this.setState({ secTab: 'alerts' }) }] : l.result !== 'Success' ? [{ label: 'Block IP', kind: 'danger', run: () => this.notify(l.ip + ' blocked. The block was logged.', '#FF9B8A') }] : [{ label: 'Details', run: () => this.notify(l.user + ' · ' + l.action + ' · ' + l.ip, '#7CC4FF') }])] };
      }) };
    const alertsTable = { id: 'alerts', pageSize: 8, minWidth: '1100px', emptyText: 'No alerts. Rules that are switched off do not raise alerts.',
      cols: [{ label: 'ID', filter: 'text' }, { label: 'Severity', filter: 'select' }, { label: 'Alert', filter: 'text' }, { label: 'Rule', filter: 'select' }, { label: 'Raised', filter: 'date' }, { label: 'Status', filter: 'select' }, { label: '', align: 'right' }],
      rows: alertsAll.map((a) => ({ tint: a.isOpen ? a.color : '', cells: [C.t(a.id, { mono: true, nowrap: true }), C.b(a.sev, a.color), C.t(a.title, { strong: true, sub: a.detail }), C.t(a.ruleName), C.t(a.when, { fv: isoFromAgo(a.ago), nowrap: true }),
        C.b(a.status, a.status === 'Open' ? '#FF9B8A' : a.status === 'Investigating' ? '#FFB86B' : '#9DB0A6'),
        C.a([{ label: 'Investigate', run: a.investigate, disabled: a.status === 'Investigating' }, { label: 'Resolve', kind: 'primary', run: a.resolve, disabled: !a.isOpen }, { label: 'False positive', run: a.falsePos, disabled: !a.isOpen }])] })) };
    const sevOpts = [{ v: 'High', l: 'High' }, { v: 'Medium', l: 'Medium' }, { v: 'Low', l: 'Low' }];
    const rulesTable = { id: 'rules', pageSize: 10, minWidth: '860px',
      cols: [{ label: 'Rule', filter: 'text' }, { label: 'Severity', filter: 'select' }, { label: 'Matches today' }, { label: 'On', filter: 'select', align: 'center' }],
      rows: sec.rules.map((r) => ({ cells: [C.t(r.name, { strong: true, sub: r.desc }), C.sel(r.sev, sevOpts, r.setSev, { label: 'Severity for ' + r.name, minWidth: '110px' }), C.t(r.hits), C.chk(r.on, r.toggle, { label: 'Turn ' + r.name + (r.on ? ' off' : ' on') })] })) };
    const rolesTable = { id: 'roles', pageSize: 10, minWidth: '900px',
      cols: [{ label: 'Role', filter: 'text' }, { label: 'Scope', filter: 'select' }, { label: 'People' }, { label: 'Rights granted' }, { label: '', align: 'right' }],
      rows: visibleRoles.map((r) => { const mods = Object.keys(rightsDef).filter(modEnabled); let on = 0, tot = 0; mods.forEach((m) => rightsDef[m].forEach((x) => { tot++; if (rightOf(r.id, m, x)) on++; })); const pc = Math.round(on / tot * 100);
        return { cells: [C.t(r.name, { strong: true, sub: r.desc }), C.t(roleScope[r.id] || 'Whole organization'), C.t(String(r.count)), C.bar(pc, pc > 80 ? '#FFB86B' : '#B8F36B', on + ' / ' + tot),
          C.a([{ label: 'Edit rights', kind: 'primary', run: () => this.setState({ rolesTab: 'rights' }) }, { label: 'Members', run: () => this.setState({ rolesTab: 'members' }) }])] }; }) };
    const rightsRows = [];
    Object.keys(rightsDef).filter(modEnabled).forEach((m) => rightsDef[m].forEach((x) => {
      rightsRows.push({ cells: [C.t(modInfo[m][0], { fv: modInfo[m][0] }), C.t(x, { strong: true })].concat(visibleRoles.map((r) => C.chk(rightOf(r.id, m, x), () => toggleRight(r.id, m, x), { disabled: !!r.locked, pending: s.pendingRights[rKey(r.id, m, x)] !== undefined, label: r.name + ': ' + modInfo[m][0] + ' · ' + x }))) });
    }));
    const rightsTable = { id: 'rights', pageSize: 12, minWidth: (300 + visibleRoles.length * 96) + 'px',
      cols: [{ label: 'Widget', filter: 'select' }, { label: 'Right', filter: 'text' }].concat(visibleRoles.map((r) => ({ label: r.name, align: 'center' }))), rows: rightsRows };
    const rightsHint = 'Tick a box to let that role do that action. ' + (f8 && f8.levels > 0 ? 'Changes need approval (' + (f8.levels === 1 ? 'direct' : f8.levels + ' levels') + ') and show as PENDING until approved.' : 'Changes apply straight away and are logged.') + ' Widgets your organization does not use are hidden.';
    const defaultRole = { 'Samwel Okoth Omulo': 'admin', 'Brian Kiprop': 'it', 'David Otieno': 'staff', 'Amina Wanjiru': 'finance', 'Grace Njeri': 'hod', 'Faith Achieng': 'staff', 'Kevin Mutua': 'staff', 'Mercy Wambui': 'staff', 'John Kamau': 'staff', 'Lilian Chebet': 'staff' };
    const memberList = ['Samwel Okoth Omulo'].concat(corePeople).concat(this.dir.slice(0, 40).map((d) => d.name)).filter((n, i, a) => a.indexOf(n) === i);
    const roleOpts = visibleRoles.map((r) => ({ v: r.id, l: r.name }));
    const membersTable = { id: 'members', pageSize: 12, minWidth: '860px',
      cols: [{ label: 'Person', filter: 'text' }, { label: 'Role', filter: 'select' }, { label: 'Department', filter: 'select' }, { label: '', align: 'right' }],
      rows: memberList.map((n) => {
        const d = this.dir.find((x) => x.name === n);
        let rid = s.memberRoles[n] || defaultRole[n] || (d && /Lead|Supervisor/.test(d.title) ? (has.branches ? 'branch' : 'hod') : 'staff');
        if (!visibleRoles.find((r) => r.id === rid)) rid = 'staff';
        const rname = (visibleRoles.find((r) => r.id === rid) || {}).name;
        return { cells: [C.p(n, { sub: d ? d.title : titleOf(n), fv: n }), C.sel(rid, roleOpts, (e) => { const v = e.target.value; this.setState((x) => ({ memberRoles: { ...x.memberRoles, [n]: v }, extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: 'Gave ' + n + ' the ' + (visibleRoles.find((r) => r.id === v) || {}).name + ' role', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' }].concat(x.extraLogs) })); this.notify(n + ' is now ' + (visibleRoles.find((r) => r.id === v) || {}).name + '. Logged.', '#B8F36B'); }, { fv: rname, label: 'Role for ' + n, minWidth: '170px', disabled: n === 'Samwel Okoth Omulo' }),
          C.t(d ? d.dept : ({ 'Brian Kiprop': 'IT Operations', 'David Otieno': 'Software Development', 'Amina Wanjiru': 'Finance', 'Grace Njeri': 'HR', 'Faith Achieng': 'Customer Care', 'Kevin Mutua': 'Sales', 'Samwel Okoth Omulo': 'Risk & Compliance' }[n] || 'Operations')),
          C.a([{ label: 'Remove role', kind: 'danger', disabled: n === 'Samwel Okoth Omulo', run: () => { this.setState((x) => ({ memberRoles: { ...x.memberRoles, [n]: 'staff' } })); this.notify(n + ' moved back to Staff.', '#7CC4FF'); } }])] };
      }) };
    const approverOpts = ['Line manager', 'Head of department', 'Finance approver', 'HR Manager', 'Communications lead', 'Security officer', 'Head of IT Operations', 'Managing Director'].concat(has.sectors ? ['Sector manager'] : []).concat(has.branches ? ['Branch manager'] : []).concat(has.units ? ['Unit lead'] : []).map((x) => ({ v: x, l: x }));
    const levelOpts = [{ v: '0', l: 'No approval' }, { v: '1', l: 'Direct (1 level)' }, { v: '2', l: '2 levels' }, { v: '3', l: '3 levels' }];
    const flowsTable = { id: 'flows', pageSize: 10, minWidth: '1180px',
      cols: [{ label: 'Action', filter: 'text' }, { label: 'Widget', filter: 'select' }, { label: 'Approval', filter: 'select' }, { label: 'Level 1' }, { label: 'Level 2' }, { label: 'Level 3' }],
      rows: s.flows.filter((f) => modEnabled(f.mod)).map((f) => {
        const fi = flowItem(f);
        const lv = (i) => C.sel(f.approvers[i] || 'Line manager', approverOpts, (e) => { const v = e.target.value; this.setState((st) => ({ flows: st.flows.map((x) => (x.id === f.id ? { ...x, approvers: x.approvers.map((a, j) => (j === i ? v : a)) } : x)) })); }, { disabled: f.levels <= i, label: 'Level ' + (i + 1) + ' approver for ' + f.action, minWidth: '160px' });
        return { cells: [C.t(f.action, { strong: true }), C.t(fi.modName), C.sel(String(f.levels), levelOpts, fi.setLevels, { fv: levelOpts[f.levels].l, label: 'Approval levels for ' + f.action, minWidth: '150px' }), lv(0), lv(1), lv(2)] };
      }) };
    const rolesTabs = [['roles', 'Roles'], ['rights', 'Rights'], ['members', 'Members'], ['flows', 'Approval rules']].map(([k, l]) => ({ label: l, go: () => this.setState({ rolesTab: k }),
      style: 'min-height:44px;padding:0 16px;border:0;background:transparent;cursor:pointer;font-size:14px;font-family:inherit;border-bottom:2px solid ' + (s.rolesTab === k ? '#B8F36B' : 'transparent') + ';color:' + (s.rolesTab === k ? '#E8EFEA' : '#9DB0A6') + ';font-weight:' + (s.rolesTab === k ? '600' : '500') }));
    const rolesView = { isRoles: s.rolesTab === 'roles', isRights: s.rolesTab === 'rights', isMembers: s.rolesTab === 'members', isFlows: s.rolesTab === 'flows' };
    const ntDepts = has.departments ? O.departments.map((d) => ({ label: d.name, pick: () => this.setState((x) => ({ nt: { ...x.nt, dept: d.name } })), style: chip((s.nt.dept || 'IT Operations') === d.name) })) : [];

    // =============== Performance awards ===============
    const awardsOn = s.widgets.indexOf('awards') !== -1;
    const approvedAwards = s.awardList.filter((a) => a.status === 'Approved');
    const stripBase = [];
    approvedAwards.forEach((a) => a.people.forEach((n) => {
      const isMe = n === 'Samwel Okoth Omulo';
      const pp = personOf(n);
      stripBase.push({ name: n, initials: isMe ? 'S' : pp.initials, color: pp.color, award: a.type, when: fmtDay(a.date), photo: isMe ? s.profile.photo : '', hasPhoto: isMe && !!s.profile.photo, noPhoto: !(isMe && s.profile.photo), open: () => openPersonName(n) });
    }));
    const awardStrip = {
      title: 'Performance awards', sub: stripBase.length + ' people recognised recently · tap anyone to see their profile',
      items: stripBase.length ? stripBase.concat(stripBase) : [],
      empty: stripBase.length === 0,
      trackStyle: 'display:flex;gap:14px;width:max-content;padding:0 7px;animation:hubMarquee ' + Math.max(30, stripBase.length * 6) + 's linear infinite',
      canGive: can('awards', 'Nominate') || can('awards', 'Approve awards'),
      openSettings: () => this.setState({ settingsFor: 'awards' }),
      remove: () => this.setState((x) => ({ widgets: x.widgets.filter((w) => w !== 'awards') }))
    };
    const awS = s.awd;
    const setAw = (patch) => this.setState((x) => ({ awd: Object.assign({}, x.awd, typeof patch === 'function' ? patch(x.awd) : patch) }));
    const personRole = (n) => s.memberRoles[n] || ({ 'Samwel Okoth Omulo': 'admin', 'Brian Kiprop': 'it', 'Amina Wanjiru': 'finance', 'Grace Njeri': 'hod' }[n]) || (/Lead|Supervisor/.test(((this.dir.find((d) => d.name === n) || {}).title) || '') ? (has.branches ? 'branch' : 'hod') : 'staff');
    const awq = awS.q.trim().toLowerCase();
    const awPool = this.dir.filter((d) => (awS.role === 'all' || personRole(d.name) === awS.role) && (!awq || (d.name + ' ' + d.title + ' ' + d.dept).toLowerCase().indexOf(awq) !== -1));
    const awMatches = awPool.slice(0, 40);
    const f7 = flowById('f7');
    const awTypes = ['Employee of the month', 'Star performer', 'Customer hero', 'Team player', 'Innovation award', 'Long service'];
    const aw = {
      type: awS.type, q: awS.q, role: awS.role, citation: awS.citation, poster: awS.poster, hasPoster: !!awS.poster, noPoster: !awS.poster, wall: awS.wall,
      types: awTypes.map((t) => ({ label: t, pick: () => setAw({ type: t }), style: chip(awS.type === t) })),
      roleOpts: visibleRoles.filter((r) => r.id !== 'superadmin').map((r) => ({ v: r.id, l: r.name })),
      setQ: (e) => setAw({ q: e.target.value }), setRole: (e) => setAw({ role: e.target.value }), setCitation: (e) => setAw({ citation: e.target.value }),
      setPoster: (e) => { const f = (e.target.files || [])[0]; e.target.value = ''; if (f) setAw({ poster: URL.createObjectURL(f) }); },
      clearPoster: () => setAw({ poster: '' }),
      toggleWall: () => setAw((x) => ({ wall: !x.wall })),
      forumChips: forums.map((f) => { const on = awS.forums.indexOf(f.id) !== -1; return { name: f.name, on: on, style: chipS(on), toggle: () => setAw((x) => ({ forums: on ? x.forums.filter((y) => y !== f.id) : x.forums.concat([f.id]) })) }; }),
      countLabel: awS.people.length ? awS.people.length + (awS.people.length === 1 ? ' person selected' : ' people selected') : 'Nobody selected yet',
      chosen: awS.people.map((n) => ({ ...personOf(n), remove: () => setAw((x) => ({ people: x.people.filter((y) => y !== n) })) })),
      matchLabel: awPool.length > 40 ? 'Showing 40 of ' + awPool.length + ' matches. Search to narrow it down.' : awPool.length + (awPool.length === 1 ? ' match' : ' matches'),
      selectAll: () => setAw((x) => ({ people: x.people.concat(awMatches.map((d) => d.name).filter((n) => x.people.indexOf(n) === -1)) })),
      matches: awMatches.map((d) => { const on = awS.people.indexOf(d.name) !== -1; return { name: d.name, initials: d.initials, color: d.color, sub: d.title + ' · ' + (roleDefs.find((r) => r.id === personRole(d.name)) || {}).name, on: on, mark: on ? 'Selected' : '+ Add',
        toggle: () => setAw((x) => ({ people: on ? x.people.filter((y) => y !== d.name) : x.people.concat([d.name]) })),
        style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:44px;padding:4px 10px;border:0;cursor:pointer;color:#E8EFEA;font-family:inherit;background:' + (on ? '#24220F' : 'transparent') }; }),
      posterNames: awS.people.length ? (awS.people.length > 3 ? awS.people.slice(0, 3).join(', ') + ' +' + (awS.people.length - 3) : awS.people.join(', ')) : 'Awardee names',
      approvalNote: f7 && f7.levels > 0 ? 'Awards are approved before they appear (' + (f7.levels === 1 ? 'direct approval by ' + f7.approvers[0] : f7.levels + ' levels: ' + f7.approvers.slice(0, f7.levels).join(' → ')) + ').' : 'Awards appear straight away.',
      submitLabel: f7 && f7.levels > 0 ? 'Send for approval' : 'Give award',
      submitStyle: 'height:42px;padding:0 20px;border-radius:10px;border:0;font-weight:700;cursor:pointer;' + (awS.people.length ? 'background:#FFD66B;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
      close: set({ awOpen: false }),
      submit: () => {
        const a0 = this.state.awd;
        if (!a0.people.length) { this.notify('Pick at least one person to award.', '#FFB86B'); return; }
        const id = 'AW-' + (13 + this.state.awardList.length);
        const needs = f7 && f7.levels > 0;
        const rec = { id: id, type: a0.type, people: a0.people.slice(), citation: a0.citation.trim(), date: todayISO, status: needs ? 'Pending' : 'Approved', poster: a0.poster, wall: a0.wall, forums: a0.forums.slice() };
        this.setState((x) => ({ awOpen: false, awardList: [rec].concat(x.awardList), awd: { type: 'Employee of the month', people: [], role: 'all', q: '', citation: '', poster: '', wall: true, forums: [] } }));
        const label = a0.type + ' for ' + (a0.people.length > 2 ? a0.people.length + ' people' : a0.people.join(' and '));
        if (needs) makeRequest('f7', label, { kind: 'award', awardId: id });
        else this.notify(label + ' is live on Home.', '#B8F36B');
      }
    };

    // ---- Files ----
    const kindColor = { PDF: '#FF9B8A', DOC: '#7CC4FF', XLS: '#B8F36B', PPT: '#FFB86B', IMG: '#F49AC1', ZIP: '#C9B6FF' };
    const folders = s.folders.map((f) => ({ ...f, count: f.files.length,
      samples: f.files.slice(0, 3).map((x, i) => ({ kind: x.kind, color: kindColor[x.kind] || '#9DB0A6', h: [86, 74, 80][i] + 'px' })),
      open: () => this.setState({ folder: f.id }) }));
    const cur_f = s.folders.find((f) => f.id === s.folder);
    const fv = cur_f ? { ...cur_f, count: cur_f.files.length, files: cur_f.files.map((x) => ({ ...x, color: kindColor[x.kind] || '#9DB0A6',
      isDoc: x.kind === 'PDF' || x.kind === 'DOC', isSheet: x.kind === 'XLS', isSlide: x.kind === 'PPT', isImg: x.kind === 'IMG', isZip: x.kind === 'ZIP' || !kindColor[x.kind],
      hasUrl: !!x.url, noUrl: !x.url, imgBg: x.imgBg || '#3A3150' })) } : {};

    const catalog = {
      announcements: { title: 'Announcements', sub: 'From management', tint: '#B8F36B', desc: 'Pinned company news. Staff mark each one as read, so admins can see who has seen it.' },
      links: { title: 'Quick links', sub: 'Your daily tools', tint: '#7CC4FF', desc: 'Shortcuts to email, payroll, HR and any other system your staff use.' },
      events: { title: 'Upcoming events', sub: 'Next few weeks', tint: '#FFB86B', desc: 'Events across the organization with one-tap RSVP.' },
      clocks: { title: 'World clock', sub: 'Team time zones', tint: '#6EE7C8', desc: 'Live time in the cities your teams work from.' },
      birthdays: { title: 'Celebrations', sub: 'Birthdays & anniversaries', tint: '#F49AC1', desc: 'Who is celebrating today, with a quick way to send wishes.' },
      tickets: { title: 'Tickets', sub: 'Turnaround time tracking', tint: '#FF9B8A', desc: 'Open, overdue and due-soon tickets, the most urgent ones to view, and a quick way to raise a new one.' },
      projects: { title: 'Projects', sub: 'Progress and health', tint: '#7CC4FF', desc: 'Every project with its status, progress and due date, plus a View button into each one.' },
      awards: { title: 'Performance awards', sub: 'Recognition', tint: '#FFD66B', desc: 'Spotlight top performers, show award winners and let staff nominate colleagues.' },
      secureshare: { title: 'Secure share', sub: 'One-time links', tint: '#6EE7C8', desc: 'Share a password or credential through a link that self-destructs after one view.' },
      wall: { title: 'Community wall', sub: 'Latest posts', tint: '#F49AC1', desc: 'The newest posts from the wall, so nobody misses what colleagues are sharing.' },
      forums: { title: 'My forums', sub: 'Unread activity', tint: '#7CC4FF', desc: 'Your forums with unread counts. Open any one straight into the side panel.' },
      people: { title: 'People finder', sub: 'Directory', tint: '#C9B6FF', desc: 'Search anyone in the organization and see who is online.' },
      files: { title: 'Recent files', sub: 'Across your forums', tint: '#C9B6FF', desc: 'The latest documents shared with you.' },
      joiners: { title: 'New joiners', sub: 'This week', tint: '#B8F36B', desc: 'Welcome new colleagues and help them settle in.' },
      sessions: { title: 'Meetings & standups', sub: 'Live & upcoming', tint: '#6EE7C8', desc: 'Live meetings and daily standups you can join, plus a button to start one.' },
      poll: { title: 'Quick poll', sub: 'Closes Friday', tint: '#FFB86B', desc: 'Ask everyone a question and show live results.' },
      notes: { title: 'My notes', sub: 'Only you can see this', tint: '#FFD66B', desc: 'A private sticky note for your own reminders.' },
      countdown: { title: 'Countdown', sub: 'Big day ahead', tint: '#6EE7C8', desc: 'Count down the days to an event, launch or deadline.' },
      embed: { title: 'Embed', sub: 'Any link', tint: '#C9B6FF', desc: 'Show a dashboard, video, form or website inside your home page.' },
      branches: { title: 'Branches', sub: 'Your locations', tint: '#7CC4FF', desc: 'For organizations with several locations: each branch, its head, staff count and contacts.' },
      sectors: { title: 'Sectors', sub: 'Top level of the structure', tint: '#FFB86B', desc: 'Groups branches (or departments) into sectors under the organization. Turns on sector dashboards and the Sector manager role.' },
      departments: { title: 'Departments', sub: 'Teams and how they perform', tint: '#B8F36B', desc: 'Departments sit under their branch or sector. Tickets go to departments, and each department gets a dashboard and a Head of department role.' },
      units: { title: 'Units', sub: 'Teams inside branches', tint: '#C9B6FF', desc: 'Units are the smallest level and sit inside a branch. Turns on unit dashboards and the Unit lead role.' },
      performance: { title: 'Our performance', sub: 'Against a 95% target', tint: '#6EE7C8', desc: 'How your department, branch, sector or unit resolves what it receives on time, against its target.' },
      calendar: { title: 'Calendar', sub: 'Coming up', tint: '#7CC4FF', desc: 'Your meetings, standups, tasks and events, with project deadlines. Schedule straight from here.' },
      memos: { title: 'Memos', sub: 'Latest memos', tint: '#FFD66B', desc: 'Send formal memos to the wall, a forum or chosen departments, branches, sectors or units.' },
      newsletter: { title: 'Newsletter', sub: 'Latest issues', tint: '#F49AC1', desc: 'Publish a regular newsletter and see how many people read it.' },
      chat: { title: 'Private messages', sub: 'One-to-one chats', tint: '#C9B6FF', desc: 'Private one-to-one chats with private voice and video calls. Opens in the same side panel, full screen or bubble as forums.' }
    };
    if (catalog.branches) catalog.branches.desc = 'Branches sit under their sector (if you use sectors) and hold units and departments. Turns on branch dashboards and the Branch manager role.';
    const pillBtn = (on) => 'flex:none;height:30px;padding:0 12px;border-radius:15px;cursor:pointer;font-size:12px;font-weight:600;' + (on ? 'border:0;background:#2C3D34;color:#B8F36B' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA');
    const tz = (zone) => { try { return new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: zone }); } catch (err) { return '--:--'; } };
    const tzDay = (zone) => { try { return new Date().toLocaleDateString('en-GB', { weekday: 'long', timeZone: zone }); } catch (err) { return ''; } };
    const hpollBase = [{ id: 'a', label: 'Hotel ballroom', base: 14 }, { id: 'b', label: 'Beach resort', base: 22 }, { id: 'c', label: 'Our own office rooftop', base: 9 }];
    const hTotal = hpollBase.reduce((n, o) => n + o.base, 0) + (s.hvote ? 1 : 0);
    const itemsFor = {
      chat: s.dms.slice(0, 4).map((d) => { const pr = this.dir.find((x) => x.name === d.with) || { initials: ini(d.with), color: '#C9B6FF' }; const last = d.msgs[d.msgs.length - 1];
        return { name: d.with, initials: pr.initials, color: pr.color, unread: d.unread, hasUnread: d.unread > 0, last: last ? (last.from === 'me' ? 'You: ' : '') + (last.text || 'Sent a file') : 'Say hello', open: () => this.openDm(d.with) }; }),
      announcements: [
        { id: 'a1', tag: 'Policy', title: 'Updated leave policy takes effect 1 November', meta: 'HR · 2 days ago' },
        { id: 'a2', tag: 'Office', title: 'Office closed Friday for maintenance', meta: 'Admin · Today' },
        { id: 'a3', tag: 'IT', title: 'Planned downtime Saturday 8 to 10pm', meta: 'IT Operations · Yesterday' }
      ].map((a) => ({ ...a, ackLabel: s.acked[a.id] ? 'Read' : 'Mark read', ackStyle: pillBtn(!!s.acked[a.id]), ack: () => this.setState((st) => ({ acked: { ...st.acked, [a.id]: true } })) })),
      links: [['Email', 'E', '#7CC4FF'], ['Payroll', 'P', '#B8F36B'], ['HR system', 'H', '#F49AC1'], ['IT helpdesk', 'I', '#FFB86B'], ['Policies', 'P', '#6EE7C8'], ['Leave request', 'L', '#C9B6FF']].map(([label, letter, color]) => ({ label, letter, color })),
      events: [
        { id: 'e1', mon: 'OCT', day: '18', title: 'Team building', meta: 'Saturday · Karura Forest' },
        { id: 'e2', mon: 'OCT', day: '22', title: 'Quarterly town hall', meta: 'Wednesday · 10:00 · Online' },
        { id: 'e3', mon: 'OCT', day: '31', title: 'Dress-up day', meta: 'Friday · All offices' }
      ].map((e) => ({ ...e, rsvpLabel: s.rsvp[e.id] ? 'Going' : 'RSVP', rsvpStyle: pillBtn(!!s.rsvp[e.id]), rsvp: () => this.setState((st) => ({ rsvp: { ...st.rsvp, [e.id]: !st.rsvp[e.id] } })) })),
      clocks: [['Nairobi', 'Africa/Nairobi'], ['London', 'Europe/London'], ['New York', 'America/New_York'], ['Dubai', 'Asia/Dubai']].map(([city, zone]) => ({ city, time: tz(zone), day: tzDay(zone) })),
      birthdays: [
        { id: 'b1', name: 'Faith Achieng', initials: 'FA', color: '#B8F36B', what: 'Birthday today' },
        { id: 'b2', name: 'Kevin Mutua', initials: 'KM', color: '#C9B6FF', what: '3 years with the team' }
      ].map((b) => ({ ...b, wishLabel: s.wished[b.id] ? 'Sent' : 'Send wish', wishStyle: pillBtn(!!s.wished[b.id]), wish: () => this.setState((st) => ({ wished: { ...st.wished, [b.id]: true } })) })),
      tickets: [{ n: nOpen, label: 'Open', color: '#E8EFEA' }, { n: nOver, label: 'Overdue', color: '#FF9B8A' }, { n: nSoon, label: 'Due soon', color: '#FFB86B' }],
      files: files.slice(0, 4),
      joiners: [
        { name: 'Mercy Wambui', role: 'Accountant · Finance', initials: 'MW', color: '#FFB86B' },
        { name: 'John Kamau', role: 'Driver · Logistics', initials: 'JK', color: '#7CC4FF' },
        { name: 'Lilian Chebet', role: 'Marketing Officer', initials: 'LC', color: '#F49AC1' }
      ],
      sessions: sessions,
      projects: projects,
      awards: [
        { award: 'Star performer · Q3', name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF' },
        { award: 'Customer hero', name: 'Faith Achieng', initials: 'FA', color: '#B8F36B' },
        { award: 'Team of the quarter', name: 'IT Operations', initials: 'IT', color: '#6EE7C8' }
      ],
      wall: s.posts.slice(0, 2).map((p) => ({ author: p.author, initials: p.initials, color: p.color, text: (p.text || 'Shared a file').length > 90 ? p.text.slice(0, 90) + '…' : (p.text || 'Shared a file'), likes: p.likes + (s.liked[p.id] ? 1 : 0) })),
      forums: forums.filter((f) => f.hasUnread).concat(forums.filter((f) => !f.hasUnread)).slice(0, 4),
      people: people.slice(0, 4),
      branches: [
        { name: 'Head Office', place: '[CITY]', head: 'Amina Wanjiru', staff: 84, open: true },
        { name: 'North Branch', place: '[CITY]', head: 'Brian Kiprop', staff: 32, open: true },
        { name: 'Coast Branch', place: '[CITY]', head: 'Grace Njeri', staff: 27, open: true },
        { name: 'West Branch', place: '[CITY]', head: 'David Otieno', staff: 19, open: false }
      ].map((b) => ({ ...b, initials: b.name.split(' ').map((x) => x[0]).join(''), statusLabel: b.open ? 'Open' : 'Closed today', statusColor: b.open ? '#B8F36B' : '#FF9B8A' })),
      sectors: [
        { name: 'Business', color: '#FFB86B', depts: 'Sales · Marketing · Customer Care', staff: 61 },
        { name: 'Operations', color: '#7CC4FF', depts: 'IT · Logistics · Facilities', staff: 48 },
        { name: 'Corporate', color: '#F49AC1', depts: 'Finance · HR · Legal · Risk', staff: 37 }
      ],
      poll: hpollBase.map((o) => {
        const v = o.base + (s.hvote === o.id ? 1 : 0);
        const pct = Math.round((v / hTotal) * 100);
        const mine = s.hvote === o.id;
        return { label: o.label, pct: s.hvote ? pct + '%' : '', vote: () => this.setState({ hvote: o.id }),
          style: 'display:flex;align-items:center;min-height:42px;padding:0 12px;border-radius:10px;cursor:pointer;font-size:13px;color:#E8EFEA;text-align:left;border:1px solid ' + (mine ? '#B8F36B' : '#2A3832') + ';background:' + (s.hvote ? 'linear-gradient(90deg,#2C3D34 ' + pct + '%,#121B17 ' + pct + '%)' : '#121B17') };
      })
    };
    const daysLeft = Math.max(0, Math.ceil((new Date('2026-12-18T00:00:00+03:00') - new Date()) / 86400000));
    const ids = s.widgets;
    const widgets = ids.filter((x) => x !== 'awards').map((id, i) => ({
      ...catalog[id],
      title: (s.wcfg[id] && s.wcfg[id].title) || catalog[id].title,
      openSettings: () => this.setState({ settingsFor: id }),
      startMeeting: () => this.startCall('video', 'Instant meeting', 'S', '#F49AC1', []),
      t: { [id]: true },
      items: id === 'calendar' ? calItems : (id === 'memos' || id === 'newsletter') ? s.comms.filter((c) => c.type === (id === 'memos' ? 'Memo' : 'Newsletter') && c.status === 'Published').slice(0, 3).map((c) => ({ ...c, by: c.by === 'Samwel Okoth Omulo' ? 'You' : c.by, dateLabel: fmtDay(c.date), audience: c.to.length > 2 ? c.to.slice(0, 2).join(', ') + ' +' + (c.to.length - 2) : c.to.join(', '), readBar: 'height:100%;width:' + c.read + '%;background:#B8F36B' })) : (orgWidgetItems[id] || itemsFor[id] || []),
      isCommList: id === 'memos' || id === 'newsletter',
      noComms: (id === 'memos' || id === 'newsletter') && !s.comms.some((c) => c.type === (id === 'memos' ? 'Memo' : 'Newsletter')),
      compose: () => openCm(id === 'memos' ? 'Memo' : id === 'newsletter' ? 'Newsletter' : 'Announcement'),
      composeLabel: id === 'memos' ? 'New memo' : id === 'newsletter' ? 'New issue' : 'New announcement',
      goComms: () => this.setState({ section: 'comms' }),
      newEvent: () => openEv('Personal'), goCalendar: () => this.setState({ section: 'calendar' }),
      isOrgList: id === 'departments' || id === 'units',
      orgNote: id === 'departments' ? 'Each sits under ' + (has.branches ? 'its branch' : has.sectors ? 'its sector' : 'the organization') + '. % = received items resolved on time.' : 'Units sit inside ' + (has.branches ? 'their branch' : 'the organization') + '.',
      goStructure: () => this.setState({ section: 'structure' }),
      target: 95, noUnits: id === 'performance' && myUnits.length === 0,
      customizing: s.customize,
      note: s.note,
      onNote: (e) => this.setState({ note: e.target.value }),
      days: daysLeft,
      goTickets: () => this.setState({ section: 'tickets', ticketOpen: null }),
      goProjects: () => this.setState({ section: 'projects', projectOpen: null }),
      urgent: tickets.filter((t) => t.step < 3).slice().sort((a, b) => ['Overdue', 'Due soon', 'On time', 'TAT paused'].indexOf(a.slaLabel) - ['Overdue', 'Due soon', 'On time', 'TAT paused'].indexOf(b.slaLabel)).slice(0, 3),
      pstats: projStats.slice(1),
      goMeetings: () => this.setState({ section: 'meetings' }),
      goWall: () => this.setState({ section: 'feed' }),
      goPeople: () => this.setState({ section: 'people' }),
      nominated: !!s.nominated,
      nominateLabel: s.nominated ? 'Nomination sent' : 'Nominate a colleague',
      nominate: () => this.setState({ nominated: true }),
      shareMade: !!s.wlink,
      shareIdle: !s.wlink,
      makeShare: () => this.setState({ wlink: true }),
      resetShare: () => this.setState({ wlink: false }),
      cardStyle: 'break-inside:avoid;margin:0 0 16px;box-sizing:border-box;width:100%;padding:18px;border-radius:18px;background:#16201C;display:flex;flex-direction:column;gap:14px;border:' + (s.customize ? '1px dashed #4A5C54' : '1px solid #24322C'),
      remove: () => this.setState((st) => ({ widgets: st.widgets.filter((x) => x !== id) })),
      moveUp: () => this.setState((st) => {
        const w = st.widgets.slice(); const k = w.indexOf(id);
        if (k > 0) { const tmp = w[k - 1]; w[k - 1] = w[k]; w[k] = tmp; }
        return { widgets: w };
      })
    }));
    const gallery = Object.keys(catalog).filter((id) => id !== 'performance' || has.structure).map((id) => {
      const added = ids.indexOf(id) !== -1;
      return { ...catalog[id], btnLabel: added ? 'Added' : 'Add to home',
        openSettings: () => this.setState({ settingsFor: id }),
        btnStyle: 'flex:1;height:38px;border-radius:10px;font-weight:600;cursor:' + (added ? 'default' : 'pointer') + ';' + (added ? 'border:1px solid #2A3832;background:transparent;color:#9DB0A6' : 'border:0;background:#B8F36B;color:#0F1714'),
        add: () => { if (!added) this.setState((st) => ({ widgets: st.widgets.concat(id) })); } };
    });
    // ---- My profile ----
    nav.profileCard = 'margin-top:auto;display:flex;align-items:center;gap:10px;width:100%;min-height:58px;padding:8px 10px;border-radius:12px;cursor:pointer;color:#E8EFEA;font-family:inherit;border:1px solid ' + (s.section === 'profile' ? '#3F524A' : '#24322C') + ';background:' + (s.section === 'profile' ? '#22302A' : '#16201C');
    const pf = s.profile;
    const statusDefs = [
      { label: 'Available', color: '#B8F36B', note: 'Available for chats and calls' },
      { label: 'In a meeting', color: '#FFB86B', note: 'In a meeting, replies may be slow' },
      { label: 'Focusing', color: '#7CC4FF', note: 'Focusing, notifications paused' },
      { label: 'Away', color: '#9DB0A6', note: 'Away from desk' }
    ];
    const curStatus = statusDefs.find((x) => x.label === pf.status) || statusDefs[0];
    const setPf = (patch) => this.setState((st) => ({ profile: Object.assign({}, st.profile, patch) }));
    let localTime = '';
    try { localTime = new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi' }); } catch (err) { localTime = ''; }
    const checks = [!!pf.photo, (pf.bio || '').trim().length > 20, pf.skills.length >= 3, true, !!pf.phone.trim()];
    const complete = Math.round((checks.filter(Boolean).length / checks.length) * 100);
    const tip = !pf.photo ? 'Add a profile photo so colleagues recognise you.' : !pf.phone.trim() ? 'Add your phone number to reach 100%.' : (pf.bio || '').trim().length <= 20 ? 'Write a short bio.' : 'Your profile is complete.';
    const myPosts = s.posts.filter((p) => p.author === 'Sam');
    const myRaised = s.tickets.filter((t) => t.raisedBy === 'Sam');
    const myActivity = myPosts.slice(0, 2).map((p) => ({ tag: 'W', color: '#F49AC1', text: 'Posted on the wall: "' + ((p.text || 'a file').length > 70 ? p.text.slice(0, 70) + '…' : (p.text || 'a file')) + '"', time: p.meta.split('· ').pop() }))
      .concat(myRaised.slice(0, 2).map((t) => ({ tag: 'T', color: '#FF9B8A', text: 'Raised ' + t.id + ': ' + t.title, time: t.created })))
      .concat([
        { tag: 'F', color: '#6EE7C8', text: 'Shared "Access review template" in Risk & Compliance', time: 'Yesterday' },
        { tag: 'A', color: '#7CC4FF', text: 'Assigned IT-208 to Brian Kiprop', time: 'Today, 08:40' },
        { tag: 'B', color: '#FFD66B', text: 'Earned the Security Champion badge', time: 'Last month' }
      ]);
    const me = {
      name: 'Samwel Okoth Omulo', role: 'Cybersecurity & Technology Risk Officer', dept: 'Risk & Compliance',
      status: curStatus.label, statusColor: curStatus.color, statusNote: curStatus.note, statusMenu: s.statusMenu,
      statuses: statusDefs.map((x) => ({ ...x, pick: () => this.setState((st) => ({ statusMenu: false, profile: { ...st.profile, status: x.label } })) })),
      photo: pf.photo, hasPhoto: !!pf.photo, noPhoto: !pf.photo,
      editing: s.profileEdit, notEditing: !s.profileEdit,
      coverStyle: 'position:relative;height:190px;background:' + pf.cover + ';overflow:hidden',
      covers: [['#1F3A2B', 'Forest'], ['#1E3347', 'Ocean'], ['#3A2433', 'Plum'], ['#3D3320', 'Sand']].map(([c, label]) => ({ label: label + ' cover', pick: () => setPf({ cover: c }),
        style: 'width:26px;height:26px;border-radius:50%;cursor:pointer;background:' + c + ';border:2px solid ' + (pf.cover === c ? '#E8EFEA' : 'rgba(255,255,255,0.25)') })),
      complete, completeTip: tip, completeBar: 'height:100%;width:' + complete + '%;background:' + (complete === 100 ? '#B8F36B' : '#FFB86B') + ';border-radius:3px',
      bio: pf.bio, phone: pf.phone, newSkill: s.newSkill,
      skills: pf.skills.map((n, i) => ({ name: n, editing: s.profileEdit, remove: () => this.setState((st) => ({ profile: { ...st.profile, skills: st.profile.skills.filter((x, j) => j !== i) } })) })),
      facts: [
        { label: 'Department', value: 'Risk & Compliance', sub: 'Cybersecurity team' },
        { label: 'Reports to', value: 'Head of Risk Management', sub: 'Risk & Compliance' },
        { label: 'Location', value: 'Kenya', sub: 'GMT+3 · ' + localTime + ' now' },
        { label: 'Member since', value: '[JOIN DATE]', sub: 'Hub member' }
      ],
      stats: [
        { label: 'Tickets raised', n: myRaised.length, color: '#FF9B8A' },
        { label: 'Tickets handled', n: s.tickets.filter((t) => t.handler === 'Sam' || t.assignees.indexOf('Sam') !== -1).length, color: '#7CC4FF' },
        { label: 'Wall posts', n: myPosts.length, color: '#F49AC1' },
        { label: 'Kudos received', n: 14, color: '#FFD66B' }
      ],
      projects: projects.filter((p) => p.id === 'audit' || p.id === 'network').map((p) => ({ ...p, myRole: p.id === 'audit' ? 'Controls reviewer' : 'Security reviewer' })),
      activity: myActivity,
      contact: [
        { label: 'Work email', value: '[you@company.com]', color: '#E8EFEA' },
        { label: 'Phone', value: pf.phone.trim() || 'Not added yet', color: pf.phone.trim() ? '#E8EFEA' : '#FFB86B' },
        { label: 'Office', value: '[OFFICE / FLOOR]', color: '#E8EFEA' }
      ],
      localTime, hours: ts.start + ' to ' + ts.end,
      badges: has.awards ? s.awardList.filter((a) => a.status === 'Approved' && a.people.indexOf('Samwel Okoth Omulo') !== -1).map((a) => ({ name: a.type, when: fmtDay(a.date), color: '#FFD66B', bg: '#2A2716' })) : [],
      hasBadges: has.awards && s.awardList.some((a) => a.status === 'Approved' && a.people.indexOf('Samwel Okoth Omulo') !== -1),
      peers: ['Brian Kiprop', 'David Otieno', 'Amina Wanjiru', 'Grace Njeri'].map((n) => ({ name: n, initials: ini(n), color: pColor[n] })),
      forums: forums.filter((f) => f.id === 'risk' || f.id === 'it' || f.id === 'dev').map((f) => ({ ...f, myRole: f.id === 'risk' ? 'Admin' : 'Member' }))
    };
    const addSkill = () => {
      const v = this.state.newSkill.trim();
      if (!v) return;
      this.setState((st) => ({ newSkill: '', profile: { ...st.profile, skills: st.profile.skills.indexOf(v) === -1 ? st.profile.skills.concat([v]) : st.profile.skills } }));
    };

    // ---- Widget settings ----
    const sid = s.settingsFor;
    const special = ['tickets', 'announcements', 'awards', 'secureshare', 'sessions', 'clocks', 'branches', 'sectors'];
    let sw = { t: {} };
    if (sid && catalog[sid]) {
      const cfg = Object.assign({ title: catalog[sid].title, see: 'Everyone', manage: 'Admins' }, s.wcfg[sid] || {});
      const setCfg = (k) => (e) => { const v = e.target.value; this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: catalog[sid].title, see: 'Everyone', manage: 'Admins' }, st.wcfg[sid] || {}, { [k]: v }) } })); };
      const wmod = sid;
      const rn = rightsDef[wmod] || genericRights;
      const wflows = s.flows.filter((f) => f.mod === wmod).map(flowItem);
      sw = { ...catalog[sid], baseTitle: catalog[sid].title, t: { [sid]: true }, genericOnly: special.indexOf(sid) === -1,
        rightNames: rn.map((x) => ({ name: x })),
        matrix: visibleRoles.map((r) => ({ role: r.name, cells: rn.map((x) => ({ label: r.name + ': ' + x, on: rightOf(r.id, wmod, x), locked: !!r.locked, pending: s.pendingRights[rKey(r.id, wmod, x)] !== undefined, toggle: () => toggleRight(r.id, wmod, x) })) })),
        flows: wflows, noFlows: wflows.length === 0,
        title: cfg.title, see: cfg.see, manage: cfg.manage, setTitle: setCfg('title'), setSee: setCfg('see'), setManage: setCfg('manage') };
      // Visibility (groups) and management (roles)
      const setW = (patch) => this.setState((st) => ({ wcfg: { ...st.wcfg, [sid]: Object.assign({ title: catalog[sid].title, see: 'Everyone', groups: [], managers: ['superadmin', 'admin'] }, st.wcfg[sid] || {}, typeof patch === 'function' ? patch(Object.assign({ groups: [], managers: ['superadmin', 'admin'] }, st.wcfg[sid] || {})) : patch) } }));
      const groups = cfg.groups || [];
      const managers = cfg.managers || ['superadmin', 'admin'];
      const isGroups = cfg.see === 'Specific groups';
      const groupSets = ['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]).map((t) => ({ label: typeName[t] + (t === 'branches' ? 'es' : 's'), items: O[t].map((r) => ({ key: t + ':' + r.id, name: r.name })) }))
        .concat([{ label: 'Forums', items: forums.map((f) => ({ key: 'forum:' + f.id, name: f.name })) }])
        .map((gs) => ({ label: gs.label, items: gs.items.map((it) => { const on = groups.indexOf(it.key) !== -1; return { name: it.name, on: on, style: chipS(on), toggle: () => setW((c) => ({ groups: on ? c.groups.filter((x) => x !== it.key) : c.groups.concat([it.key]) })) }; }) }));
      Object.assign(sw, {
        seeOpts: ['Everyone', 'Specific groups'].map((v) => ({ label: v, on: cfg.see === v, style: chip(cfg.see === v), pick: () => setW({ see: v }) })),
        isGroups: isGroups, groupSets: groupSets,
        groupsHint: groups.length ? groups.length + (groups.length === 1 ? ' group selected. Only its members see this widget.' : ' groups selected. Only their members see this widget.') : 'Pick at least one group. You can\'t save until you do.',
        groupsHintColor: groups.length ? '#C9D6CF' : '#FFB86B', groupsBorder: groups.length ? '#24322C' : '#6A5426',
        managerRoles: visibleRoles.map((r) => { const on = r.id === 'superadmin' || managers.indexOf(r.id) !== -1; return { name: r.name, on: on, locked: r.id === 'superadmin', style: chipS(on), toggle: () => setW((c) => ({ managers: on ? c.managers.filter((x) => x !== r.id) : c.managers.concat([r.id]) })) }; })
      });
      // Communication rules for structure widgets
      if (['departments', 'sectors', 'branches', 'units'].indexOf(sid) !== -1) {
        const lvl = sid;
        const roleOpts = visibleRoles.filter((r) => r.id !== 'guest').map((r) => ({ v: r.id, l: r.name }));
        const setRule = (ty, patch) => this.setState((st) => ({ commRules: { ...st.commRules, [lvl]: { ...st.commRules[lvl], [ty]: Object.assign({}, st.commRules[lvl][ty], typeof patch === 'function' ? patch(st.commRules[lvl][ty]) : patch) } } }));
        const targetKeys = ['own', 'wall'].concat(['departments', 'sectors', 'branches', 'units'].filter((t) => has[t]));
        Object.assign(sw, {
          isLevel: true, levelSingular: typeName[lvl].toLowerCase(),
          commRules: ['Memo', 'Announcement'].filter((ty) => (ty === 'Memo' ? has.memos : has.announcements) || true).map((ty) => {
            const r = s.commRules[lvl][ty];
            return { label: typeName[lvl] + 's can send ' + ty.toLowerCase() + 's', on: r.on, toggle: () => setRule(ty, (x) => ({ on: !x.on })),
              targets: targetKeys.map((k) => { const on = r.to.indexOf(k) !== -1; return { label: k === 'own' ? 'Its own ' + typeName[lvl].toLowerCase() : k === 'wall' ? 'Community wall' : typeName[k] + (k === 'branches' ? 'es' : 's'), on: on, style: chipS(on),
                toggle: () => setRule(ty, (x) => ({ to: on ? x.to.filter((y) => y !== k) : x.to.concat([k]) })) }; }),
              sender: r.sender, approver: r.approver, roleOpts: roleOpts,
              setSender: (e) => { const v = e.target.value; setRule(ty, { sender: v }); },
              setApprover: (e) => { const v = e.target.value; setRule(ty, { approver: v }); } };
          })
        });
      }
    }

    // ---- Meeting ----
    let call = {};
    if (s.call) {
      const secs = Math.max(0, Math.floor((Date.now() - s.call.start) / 1000));
      const pad = (n) => (n < 10 ? '0' : '') + n;
      const timer = (secs >= 3600 ? Math.floor(secs / 3600) + ':' : '') + pad(Math.floor((secs % 3600) / 60)) + ':' + pad(secs % 60);
      const isVideo = s.call.mode === 'video';
      const othersAll = [
        { name: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', bg: '#1B2E40', camBg: '#24405A', muted: false, camOn: true },
        { name: 'Grace Njeri', initials: 'GN', color: '#FFB86B', bg: '#33271A', camBg: '#4A3826', muted: true, camOn: false },
        { name: 'David Otieno', initials: 'DO', color: '#F49AC1', bg: '#33202C', camBg: '#4A2E40', muted: false, camOn: true },
        { name: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', bg: '#16332B', camBg: '#1F4A3E', muted: true, camOn: false },
        { name: 'Faith Achieng', initials: 'FA', color: '#B8F36B', bg: '#22301A', camBg: '#324624', muted: false, camOn: true }
      ];
      const pick = s.call.people;
      const others = pick ? pick.filter((n) => n && n !== 'Everyone' && n !== 'Samwel Okoth Omulo').slice(0, 8).map((n, i) => {
        const c = pColor[n] || ((this.dir.find((d) => d.name === n) || {}).color) || '#C9B6FF';
        return { name: n, initials: ini(n), color: c, bg: c + '26', camBg: c + '40', muted: i % 3 === 1, camOn: i % 2 === 0 };
      }) : (isVideo ? othersAll : othersAll.slice(0, 3));
      const me = { name: 'You', initials: 'S', color: '#F49AC1', bg: '#2B2A1E', camBg: '#3D3B29', muted: !s.mic, camOn: isVideo && s.cam, you: true };
      const all = [me].concat(others);
      const talkers = all.filter((p) => !p.muted);
      const speakerName = talkers.length ? talkers[Math.floor(secs / 4) % talkers.length].name : '';
      const tiles = all.map((p) => {
        const speaking = p.name === speakerName;
        const hand = p.you ? s.hand : false;
        return { ...p, speaking, hand, camOff: !p.camOn || !isVideo, camOn: p.camOn && isVideo,
          stateLabel: p.muted ? 'Muted' : speaking ? 'Speaking' : 'Listening',
          style: 'position:relative;aspect-ratio:16/9;min-height:150px;border-radius:18px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;overflow:hidden;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
          stripStyle: 'position:relative;flex:none;width:160px;height:96px;border-radius:12px;background:' + p.bg + ';display:flex;align-items:center;justify-content:center;box-sizing:border-box;border:2px solid ' + (speaking ? '#B8F36B' : 'transparent'),
          voiceStyle: 'width:104px;height:104px;border-radius:50%;background:' + p.color + ';color:#0F1714;display:flex;align-items:center;justify-content:center;font-weight:700;font-size:34px;' + (speaking ? 'animation:ring 1.2s ease-out infinite' : 'opacity:0.85') };
      });
      const ctl = (on, danger) => 'width:66px;height:60px;border-radius:16px;border:0;cursor:pointer;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;' + (danger ? 'background:#FF9B8A;color:#0F1714' : on ? 'background:#B8F36B;color:#0F1714' : 'background:#1D2A25;color:#E8EFEA');
      call = {
        ...s.call, timer, isVideo, isVoice: !isVideo, count: all.length,
        modeLabel: isVideo ? 'Video meeting' : 'Voice call',
        rec: s.rec, sharing: s.sharing && isVideo, notSharing: !(s.sharing && isVideo), tiles,
        panelOpen: !!s.callPanel, showChat: s.callPanel === 'chat', showPeople: s.callPanel === 'people',
        panelTitle: s.callPanel === 'people' ? 'People (' + all.length + ')' : 'Meeting chat',
        msgs: s.callMsgs, draft: s.callDraft,
        micLabel: s.mic ? 'Mute' : 'Unmute', micStyle: ctl(false, !s.mic),
        camLabel: s.cam ? 'Stop video' : 'Start video', camStyle: ctl(false, !s.cam),
        shareStyle: ctl(s.sharing, false), handStyle: ctl(s.hand, false), recStyle: ctl(false, s.rec),
        chatStyle: ctl(s.callPanel === 'chat', false), peopleStyle: ctl(s.callPanel === 'people', false)
      };
    }
    const sendCall = () => {
      const t = this.state.callDraft.trim();
      if (!t) return;
      this.setState((st) => ({ callMsgs: st.callMsgs.concat([{ who: 'You', text: t }]), callDraft: '' }));
    };
    const allNotifs = taskNotifs.map((n) => ({ ...n, open: () => this.setState({ notifOpen: false, section: 'projects', projectOpen: n.task.pid, taskOpen: n.task }) })).concat(notifList);
    const roleNames = { admin: 'Admin', hod: 'Head of department', it: 'IT support', finance: 'Finance approver', staff: 'Staff', guest: 'Guest' };
    return {
      canSecurity: can('security', 'View alerts') || can('security', 'View logs'),
      canRoles: can('roles', 'View roles'),
      canSeeAdmin: can('security', 'View alerts') || can('security', 'View logs') || can('roles', 'View roles'),
      canCustomize: can('home', 'Customize home'),
      canRaise: can('tickets', 'Raise'),
      canCreateProject: can('projects', 'Create project'),
      canPostWall: can('wall', 'Post'), cannotPostWall: !can('wall', 'Post'),
      previewing: s.actingRole !== 'admin', actingRole: s.actingRole, actingRoleName: roleNames[s.actingRole] || s.actingRole,
      setActingRole: (e) => { const v = e.target.value; this.setState({ actingRole: v }); if (v !== 'admin') this.notify('Previewing Hub as ' + (roleNames[v] || v) + '.', '#FFD66B'); },
      exitPreview: set({ actingRole: 'admin' }),
      dir, personOpen: !!po, psn, closePerson: set({ personOpen: null }),
      taskOpen: !!tt, td, closeTask: set({ taskOpen: null }),
      sec, secTab, hasOpenAlerts: openAlerts.length > 0, openAlertCount: String(openAlerts.length),
      roleList, roleSel, roleTab, flowsAll,
      rightsNeedApproval: !!(f8 && f8.levels > 0), rightsFlowLabel: f8 ? (f8.levels === 1 ? 'direct approval by ' + f8.approvers[0] : f8.levels + ' levels: ' + f8.approvers.slice(0, f8.levels).join(' → ')) : '',
      assignRole: () => this.notify('Role assigned. The change was logged.', '#B8F36B'),
      requests: requestsShown, noRequests: requestsShown.length === 0,
      pendingApprovals: String(pendingCount), hasPendingApprovals: pendingCount > 0,
      apTab: { waiting: segS(s.apTab === 'waiting'), all: segS(s.apTab === 'all'), waitingGo: set({ apTab: 'waiting' }), allGo: set({ apTab: 'all' }) },
      requestProject: () => { const f = flowById('f2'); if (f && f.levels > 0) makeRequest('f2', 'New project: Untitled project', {}); else this.notify('Project created.', '#B8F36B'); },
      me,
      toggleStatusMenu: () => this.setState((st) => ({ statusMenu: !st.statusMenu })),
      startEditProfile: set({ profileEdit: true, statusMenu: false }),
      cancelProfile: set({ profileEdit: false }),
      saveProfile: () => { this.setState({ profileEdit: false }); this.notify('Profile updated.', '#B8F36B'); },
      shareProfile: () => this.notify('Profile link copied. Paste it anywhere in Hub.', '#7CC4FF'),
      onBio: (e) => setPf({ bio: e.target.value }),
      onPhone: (e) => setPf({ phone: e.target.value }),
      onNewSkill: (e) => this.setState({ newSkill: e.target.value }),
      newSkillKey: (e) => { if (e.key === 'Enter') addSkill(); },
      addSkill,
      onPhoto: (e) => { const f = (e.target.files || [])[0]; e.target.value = ''; if (f) setPf({ photo: URL.createObjectURL(f) }); },
      has, cal, ev, evOpen: s.evOpen,
      awardsOn, awardStrip, aw, awOpen: s.awOpen, openAward: set({ awOpen: true }),
      newEvent: () => openEv('Personal'), newMeeting: () => openEv('Meeting'),
      isDark: s.theme !== 'light', isLight: s.theme === 'light',
      themeClass: s.theme === 'light' ? 'hub-light' : 'hub-dark',
      themeFilter: s.theme === 'light' ? 'invert(1) hue-rotate(180deg)' : 'none',
      themeAria: s.theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode',
      toggleTheme: () => this.setState((x) => ({ theme: x.theme === 'light' ? 'dark' : 'light' })), newStandup: () => openEv('Standup'),
      meetingsTable, emptyList: [],
      org, myUnits, orgOpen: s.orgOpen,
      toggleOrg: () => this.setState((x) => ({ orgOpen: !x.orgOpen })),
      orgChevron: 'transition:transform 0.2s;transform:rotate(' + (s.orgOpen ? '0' : '-90') + 'deg)',
      ou, ouTable,
      cm, cmOpen: s.cmOpen, commNew, commsTable,
      ticketsTable, peopleTable, peopleSummary, logsTable, alertsTable, rulesTable,
      rolesTable, rightsTable, membersTable, flowsTable, rolesTabs, rolesView, rightsHint, ntDepts,
      forumsOpen: s.forumsOpen,
      toggleForums: () => this.setState((st) => ({ forumsOpen: !st.forumsOpen })),
      forumsChevron: 'transition:transform 0.2s;transform:rotate(' + (s.forumsOpen ? '0' : '-90') + 'deg)',
      forumCount: s.forumsOpen ? String(forumsData.length) : String(forumsData.reduce((n, f) => n + f.unread, 0)) + ' new',
      projects, projStats, pj,
      showProjectList: s.section === 'projects' && !cur_p,
      showProjectView: s.section === 'projects' && !!cur_p,
      closeProject: set({ projectOpen: null }),
      newTask: s.newTask,
      onNewTask: (e) => this.setState({ newTask: e.target.value }),
      newTaskKey: (e) => { if (e.key === 'Enter') addTask(); },
      folders, fv,
      showFolderList: s.section === 'files' && !cur_f,
      showFolderView: s.section === 'files' && !!cur_f,
      closeFolder: set({ folder: null }),
      uploadToFolder: (e) => {
        const list = Array.from(e.target.files || []);
        e.target.value = '';
        if (!cur_f || !list.length) return;
        const kindOf = (f) => { const n = f.name.toLowerCase(); if ((f.type || '').indexOf('image/') === 0) return 'IMG'; if (/\.pdf$/.test(n)) return 'PDF'; if (/\.(docx?|txt|rtf)$/.test(n)) return 'DOC'; if (/\.(xlsx?|csv)$/.test(n)) return 'XLS'; if (/\.(pptx?|key)$/.test(n)) return 'PPT'; return 'ZIP'; };
        const added = list.map((f) => { const k = kindOf(f); return { name: f.name, kind: k, size: this.fmtSize(f.size), by: 'Sam', when: 'Just now', url: k === 'IMG' ? URL.createObjectURL(f) : '' }; });
        this.setState((st) => ({ folders: st.folders.map((f) => (f.id === cur_f.id ? { ...f, updated: 'just now', files: added.concat(f.files) } : f)) }));
        this.notify(added.length + (added.length === 1 ? ' file' : ' files') + ' uploaded to ' + cur_f.name + '.', '#B8F36B');
      },
      sw, settingsOpen: !!(sid && catalog[sid]),
      closeSettings: set({ settingsFor: null }),
      saveSettings: () => { const wc = this.state.wcfg[sid] || {}; if (wc.see === 'Specific groups' && !(wc.groups || []).length) { this.notify('Pick at least one group, or choose Everyone.', '#FFB86B'); return; } this.setState({ settingsFor: null }); this.notify('Settings saved for ' + ((s.wcfg[sid] && s.wcfg[sid].title) || catalog[sid].title) + '.', '#B8F36B'); },
      tset: ts, tsetCats, pauseRules, notifyRules,
      dayHoursLabel: (Math.round(dayH * 10) / 10) + ' hours',
      setNewCat: (e) => setTs({ newCat: e.target.value }),
      setNewTat: (e) => setTs({ newTat: e.target.value }),
      setNewUnit: (e) => setTs({ newUnit: e.target.value }),
      addCat: () => setTs((t) => (t.newCat.trim() ? { cats: t.cats.concat([{ name: t.newCat.trim(), tat: Math.max(1, Number(t.newTat) || 1), unit: t.newUnit }]), newCat: '', newTat: '' } : {})),
      setStart: (e) => setTs({ start: e.target.value }),
      setEnd: (e) => setTs({ end: e.target.value }),
      setNearPct: (e) => setTs({ nearPct: Number(e.target.value) }),
      setEscalate: (e) => setTs({ escalate: e.target.value }),
      tickets, ticketStats, tk, tagPeople,
      showTicketList: s.section === 'tickets' && !cur_t,
      showTicketView: s.section === 'tickets' && !!cur_t,
      closeTicket: set({ ticketOpen: null }),
      esc, tcomment: s.tcomment, tagOpen: s.tagOpen,
      onTComment: (e) => this.setState({ tcomment: e.target.value }),
      toggleTag: () => this.setState((st) => ({ tagOpen: !st.tagOpen })),
      postTComment: () => {
        const text = this.state.tcomment.trim();
        if (!text || !cur_t) return;
        const tagged = tagNames.filter((n) => text.indexOf('@' + n) !== -1);
        this.updTicket(cur_t.id, (t) => Object.assign({}, t, { activity: this.act(t, { kind: 'comment', text: text }) }));
        this.setState({ tcomment: '', tagOpen: false });
        if (tagged.length) this.notify('You tagged ' + tagged.join(', ') + ' on ' + cur_t.id + '. They have been notified.', '#C9B6FF', cur_t.id);
      },
      attachTicket: (e) => {
        const list = Array.from(e.target.files || []);
        e.target.value = '';
        if (!cur_t || !list.length) return;
        this.updTicket(cur_t.id, (t) => Object.assign({}, t, { activity: list.reduce((acc, f) => acc.concat([{ who: 'Sam', time: 'Just now', kind: 'file', text: f.name, meta: this.fmtSize(f.size) }]), t.activity) }));
      },
      handover: (e) => {
        const name = e.target.value;
        if (!name || !cur_t || name === cur_t.handler) return;
        const verb = cur_t.handler ? 'Handed it over to ' : 'Assigned it to ';
        this.updTicket(cur_t.id, (t) => Object.assign({}, t, { handler: name, handlerSince: 'Just now', assignees: t.assignees.indexOf(name) === -1 ? t.assignees.concat([name]) : t.assignees, step: Math.max(1, t.step), activity: this.act(t, { text: verb + name }) }));
        if (ts.notify.assigned) this.notify(cur_t.id + (cur_t.handler ? ' handed over to ' : ' assigned to ') + name + '. They have been notified.', '#7CC4FF', cur_t.id);
      },
      ntOpen: s.ntOpen, nt: s.nt, ntCats, ntPrios,
      toggleNewTicket: () => this.setState((st) => ({ ntOpen: !st.ntOpen })),
      ntSetTitle: (e) => { const v = e.target.value; this.setState((st) => ({ nt: { ...st.nt, title: v } })); },
      ntSetDesc: (e) => { const v = e.target.value; this.setState((st) => ({ nt: { ...st.nt, desc: v } })); },
      ntBtnStyle: 'height:40px;padding:0 20px;border-radius:10px;border:0;font-weight:600;cursor:pointer;' + (s.nt.title.trim() ? 'background:#B8F36B;color:#0F1714' : 'background:#2A3832;color:#9DB0A6'),
      createTicket: () => {
        const nt = this.state.nt;
        if (!nt.title.trim()) return;
        const pre = /Software/.test(nt.cat) ? 'DEV-' : /HR/.test(nt.cat) ? 'HR-' : /Network|Hardware|Access/.test(nt.cat) ? 'IT-' : 'TK-';
        const id = pre + (300 + this.state.tickets.length);
        const t = { id: id, title: nt.title.trim(), cat: nt.cat, prio: nt.prio, step: 0, hold: false, raisedBy: 'Sam', assignees: [], handler: '', handlerSince: '', created: 'Just now', elapsed: 0, sync: 'Not synced', dept: nt.dept || 'IT Operations', fromDept: 'Risk & Compliance', createdAt: new Date().toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }), escalated: false, desc: nt.desc.trim() || 'No details given.', activity: [{ who: 'Sam', kind: 'event', text: 'Raised the ticket', time: 'Just now' }] };
        this.setState((st) => ({ tickets: [t].concat(st.tickets), ntOpen: false, nt: { title: '', cat: st.nt.cat, prio: 'Medium', desc: '' }, ticketOpen: id }));
        this.notify(id + ' raised in ' + nt.cat + '. TAT ' + tatText(nt.cat) + '.', '#7CC4FF', id);
      },
      notifs: allNotifs, noNotifs: allNotifs.length === 0,
      notifOpen: s.notifOpen,
      hasBadge: !s.notifSeen && allNotifs.length > 0,
      badgeCount: allNotifs.length > 9 ? '9+' : String(allNotifs.length),
      toggleNotifs: () => this.setState((st) => ({ notifOpen: !st.notifOpen, notifSeen: true })),
      hasToast: !!s.toast, toast: s.toast, toastColor: s.toastColor,
      inCall: !!s.call, call,
      navOpen: s.navOpen, navClass: 'hub-nav' + (s.navOpen ? ' open' : ''),
      openNav: () => this.setState({ navOpen: true, notifOpen: false }), closeNav: () => this.setState({ navOpen: false }),
      navClick: (e) => { if (!this.state.navOpen) return; const b = e.target && e.target.closest ? e.target.closest('button') : null; if (b && !b.closest('[data-keep]')) this.setState({ navOpen: false }); },
      showTabbar: s.mode !== 'full' && !s.call,
      tabbar: (() => {
        const unread = s.dms.reduce((a, d) => a + d.unread, 0);
        const st = (on) => 'flex:1;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:3px;border:0;background:transparent;cursor:pointer;font-family:inherit;color:' + (on ? '#B8F36B' : '#9DB0A6') + ';font-weight:' + (on ? '600' : '500');
        const items = [
          { label: 'Home', icon: 'M3 11l9-7 9 7v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z', on: curSec === 'home', go: () => this.setState({ section: 'home', navOpen: false, ticketOpen: null, projectOpen: null }) },
          { label: 'Wall', icon: 'M5 3h14v18H5zM8 8h8M8 12h8M8 16h5', on: curSec === 'feed', go: () => this.setState({ section: 'feed', navOpen: false }) }
        ];
        if (has.chat) items.push({ label: 'Messages', icon: 'M4 5h16v11H9l-5 4z', on: isDM && (s.mode === 'docked'), dot: unread, go: () => { if (s.dms.length) this.openDm((s.dms.find((d) => d.unread) || s.dms[0]).with); else this.setState({ navOpen: true, dmsOpen: true, newDmOpen: true }); } });
        if (has.tickets) items.push({ label: 'Tickets', icon: 'M3 7h18v4a2 2 0 0 0 0 4v4H3v-4a2 2 0 0 0 0-4z', on: curSec === 'tickets', go: () => this.setState({ section: 'tickets', ticketOpen: null, navOpen: false }) });
        else if (has.calendar) items.push({ label: 'Calendar', icon: 'M3 5h18v16H3zM3 10h18M8 3v4M16 3v4', on: curSec === 'calendar', go: () => this.setState({ section: 'calendar', navOpen: false }) });
        items.push({ label: 'Menu', icon: 'M4 7h16M4 12h16M4 17h16', on: s.navOpen, go: () => this.setState((x) => ({ navOpen: !x.navOpen })) });
        return items.map((t) => ({ ...t, hasDot: !!t.dot, dot: t.dot || 0, style: st(t.on) }));
      })(),
      startInstant: () => this.startCall('video', 'Instant meeting', 'S', '#F49AC1', []),
      voiceCall: isDM ? () => this.startCall('voice', 'Private call with ' + curDm.with, dmView.initials, dmView.color, [curDm.with]) : () => this.startCall('voice', cur.name, cur.initials, cur.color),
      videoCall: isDM ? () => this.startCall('video', 'Private call with ' + curDm.with, dmView.initials, dmView.color, [curDm.with]) : () => this.startCall('video', cur.name, cur.initials, cur.color),
      endCall: () => this.endCall(),
      toggleMic: () => this.setState((st) => ({ mic: !st.mic })),
      toggleCam: () => this.setState((st) => ({ cam: !st.cam })),
      toggleShare: () => this.setState((st) => ({ sharing: !st.sharing })),
      toggleHand: () => this.setState((st) => ({ hand: !st.hand })),
      toggleRec: () => this.setState((st) => ({ rec: !st.rec })),
      openCallChat: () => this.setState((st) => ({ callPanel: st.callPanel === 'chat' ? null : 'chat' })),
      openCallPeople: () => this.setState((st) => ({ callPanel: st.callPanel === 'people' ? null : 'people' })),
      closeCallPanel: set({ callPanel: null }),
      onCallDraft: (e) => this.setState({ callDraft: e.target.value }),
      callKey: (e) => { if (e.key === 'Enter') sendCall(); },
      sendCallMsg: sendCall,
      widgets, gallery,
      noWidgets: ids.length === 0,
      galleryOpen: s.gallery,
      openGallery: set({ gallery: true }),
      closeGallery: set({ gallery: false }),
      customizing: s.customize,
      toggleCustomize: () => this.setState((st) => ({ customize: !st.customize })),
      customizeLabel: s.customize ? 'Done' : 'Customize',
      customizeBtnStyle: 'height:40px;padding:0 18px;border-radius:20px;font-weight:600;cursor:pointer;' + (s.customize ? 'border:0;background:#E8EFEA;color:#0F1714' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA'),
      forums, cur: curView, nav, go, show, tab, tabTo, ftab, posts, sessions, tickets, files, people, moments,
      emojisForum, emojisWall, wfiles,
      hasWFiles: s.wfiles.length > 0,
      emojiForumOpen: s.emojiF,
      emojiWallOpen: s.emojiW,
      toggleEmojiForum: () => this.setState((st) => ({ emojiF: !st.emojiF })),
      toggleEmojiWall: () => this.setState((st) => ({ emojiW: !st.emojiW })),
      chatRef: (el) => { this.chatEl = el; },
      fdraft: s.fdraft,
      onFDraft: (e) => this.setState({ fdraft: e.target.value }),
      sendForum: () => {
        const t = this.state.fdraft.trim();
        if (!t) return;
        this.postMsg(t);
        this.setState({ fdraft: '', emojiF: false });
      },
      fKey: (e) => {
        if (e.key !== 'Enter') return;
        const t = this.state.fdraft.trim();
        if (!t) return;
        this.postMsg(t);
        this.setState({ fdraft: '', emojiF: false });
      },
      attachForum: (e) => {
        const list = Array.from(e.target.files || []);
        list.forEach((f) => this.postMsg('', { name: f.name, meta: this.fmtSize(f.size) }));
        e.target.value = '';
      },
      onWallFiles: (e) => {
        const list = Array.from(e.target.files || []).map((f, i) => {
          const isImg = (f.type || '').indexOf('image/') === 0;
          return { id: Date.now() + '-' + i, name: f.name, meta: this.fmtSize(f.size), isImg: isImg, url: isImg ? URL.createObjectURL(f) : '' };
        });
        e.target.value = '';
        this.setState((st) => ({ wfiles: st.wfiles.concat(list) }));
      },
      draft: s.draft,
      onDraft: (e) => this.setState({ draft: e.target.value }),
      onPost: () => {
        const t = this.state.draft.trim();
        const media = this.state.wfiles;
        if (!t && !media.length) return;
        this.setState({ draft: '', wfiles: [], emojiW: false, posts: [{ id: Date.now(), author: 'Sam', initials: 'S', color: '#F49AC1', meta: 'Everyone · Just now', text: t, likes: 0, extra: 0, shares: 0, comments: [], media: media }, ...this.state.posts] });
      },
      postBtnStyle: 'height:38px;padding:0 22px;border-radius:19px;border:0;font-weight:600;cursor:' + (canPost ? 'pointer' : 'default') + ';background:' + (canPost ? '#B8F36B' : '#2A3832') + ';color:' + (canPost ? '#0F1714' : '#9DB0A6'),
      panelStyle,
      gripStyle: 'display:block;width:4px;height:44px;border-radius:2px;background:' + (s.dragging ? '#B8F36B' : '#3F524A'),
      selectMode: s.dragging ? 'none' : 'auto',
      startDrag: (e) => this.startDrag(e),
      resizeKey: (e) => {
        if (e.key === 'ArrowLeft') this.setState({ pw: Math.min(640, this.state.pw + 20) });
        if (e.key === 'ArrowRight') this.setState({ pw: Math.max(320, this.state.pw - 20) });
      },
      isDocked: s.mode === 'docked',
      isFull: s.mode === 'full', isFullForum: s.mode === 'full' && !isDM, isFullDm: s.mode === 'full' && isDM, isDM: isDM,
      panelChips: isDM ? s.dms.map((d) => { const pr = dmPerson(d.with); const on = d.id === s.dmId; return { name: d.with, initials: pr.initials, open: () => this.openDm(d.with),
        chipStyle: 'flex:none;width:36px;height:36px;border-radius:50%;cursor:pointer;font-weight:700;font-size:12px;color:#0F1714;font-family:inherit;background:' + pr.color + ';border:2px solid ' + (on ? '#E8EFEA' : 'transparent') + ';opacity:' + (on ? '1' : '0.65') }; }) : forums,
      dmList: s.dms.map((d) => { const pr = dmPerson(d.with); const last = d.msgs[d.msgs.length - 1]; const on = isDM && d.id === s.dmId && s.mode !== 'closed';
        return { name: d.with, initials: pr.initials, color: pr.color, dot: dmStatusCol[pr.status] || '#B8F36B', unread: d.unread, hasUnread: d.unread > 0,
          last: last ? (last.from === 'me' ? 'You: ' : '') + (last.text || (last.file ? 'Sent a file' : '')) : 'Say hello', open: () => this.openDm(d.with),
          style: 'display:flex;align-items:center;gap:10px;width:100%;min-height:50px;padding:4px 10px;border:0;border-radius:10px;cursor:pointer;text-align:left;font-family:inherit;color:#E8EFEA;background:' + (on ? '#22302A' : 'transparent') }; }),
      dmsOpen: s.dmsOpen, toggleDms: () => this.setState((x) => ({ dmsOpen: !x.dmsOpen })),
      dmsChevron: 'transition:transform 0.2s;transform:rotate(' + (s.dmsOpen ? '0' : '-90') + 'deg)',
      dmUnreadLabel: (() => { const n = s.dms.reduce((a, d) => a + d.unread, 0); return n ? n + ' new' : String(s.dms.length); })(),
      newDmOpen: s.newDmOpen, toggleNewDm: () => this.setState((x) => ({ newDmOpen: !x.newDmOpen, dmsOpen: true, dmq: '' })),
      dmq: s.dmq, onDmq: (e) => this.setState({ dmq: e.target.value }),
      dmMatches: (s.dmq.trim() ? this.dir.filter((d) => d.name.toLowerCase().indexOf(s.dmq.trim().toLowerCase()) !== -1) : this.dir.slice(-9)).slice(0, 6).map((d) => ({ name: d.name, initials: d.initials, color: d.color, title: d.title, open: () => this.openDm(d.name) })),
      isMin: s.mode === 'min',
      menuOpen: s.menu,
      inviteOpen: s.invite,
      createOpen: s.create,
      linkMade: s.link,
      toDocked: set({ mode: 'docked', menu: false }),
      toFull: set({ mode: 'full', menu: false }),
      toMin: set({ mode: 'min', menu: false }),
      toClosed: set({ mode: 'closed', menu: false }),
      toggleMenu: () => this.setState({ menu: !s.menu }),
      toggleInvite: () => this.setState({ invite: !s.invite }),
      toggleCreate: () => this.setState({ create: !s.create }),
      genLink: set({ link: true })
    };
  }
}
