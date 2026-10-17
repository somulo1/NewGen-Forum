// Values and handlers the templates read through hubBindings: access.
export function valsAccess(ctx) {
  const { addSkill, aw, awardStrip, awardsOn, cal, can, dir, ev, f8, flowsAll, has, me, openAlerts, openEv, pendingCount, po, psn, requestsShown, roleList, roleNames, roleSel, roleTab, s, sec, secTab, segS, set, setPf, td, tt } = ctx;
  return {
    canSecurity: can('security', 'View alerts') || can('security', 'View logs'),
    canRoles: can('roles', 'View roles'), canSettings: can('settings', 'View settings'),
    canSeeAdmin: can('security', 'View alerts') || can('security', 'View logs') || can('roles', 'View roles'),
    canCustomize: can('home', 'Customize home'),
    canWidgetsPage: can('home', 'Open widgets page'), canManageWidgets: can('home', 'Add & remove widgets'),
    canRaise: can('tickets', 'Raise'),
    canCreateProject: can('projects', 'Create project'),
    canPostWall: can('wall', 'Post'), cannotPostWall: !can('wall', 'Post'),
    previewing: s.actingRole !== 'admin', actingRole: s.actingRole, actingRoleName: roleNames[s.actingRole] || s.actingRole,
    setActingRole: (e) => { const v = e.target.value; this.setState({ actingRole: v }); if (v !== 'admin') this.notify('Previewing Hub as ' + (roleNames[v] || v) + '.', '#FFD66B'); },
    exitPreview: set({ actingRole: 'admin' }),
    dir, personOpen: !!po, psn, closePerson: set({ personOpen: null }),
    taskOpen: !!ctx.has.projects && !!tt, td, closeTask: set({ taskOpen: null }),
    sec, secTab, hasOpenAlerts: openAlerts.length > 0, openAlertCount: String(openAlerts.length),
    roleList, roleSel, roleTab, flowsAll,
    rightsNeedApproval: !!(f8 && f8.levels > 0), rightsFlowLabel: f8 ? (f8.levels === 1 ? 'direct approval by ' + f8.approvers[0] : f8.levels + ' levels: ' + f8.approvers.slice(0, f8.levels).join(' → ')) : '',
    assignRole: () => this.notify('Role assigned. The change was logged.', '#B8F36B'),
    requests: requestsShown, noRequests: requestsShown.length === 0,
    pendingApprovals: String(pendingCount), hasPendingApprovals: pendingCount > 0,
    apTab: { waiting: segS(s.apTab === 'waiting'), all: segS(s.apTab === 'all'), waitingGo: set({ apTab: 'waiting' }), allGo: set({ apTab: 'all' }) },
    me, profileBack: () => this.setState({ section: 'people' }),
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
    onPhoto: (e) => {
      const f = (e.target.files || [])[0];
      e.target.value = '';
      if (!f) return;
      if (!f.type.startsWith('image/')) { this.notify('Choose an image file for your profile photo.', '#FF9B8A'); return; }
      if (f.size > 5 * 1024 * 1024) { this.notify('Profile photos must be 5 MB or smaller.', '#FF9B8A'); return; }
      const reader = new FileReader();
      reader.onload = () => {
        if (typeof reader.result !== 'string') { this.notify('The profile photo could not be read.', '#FF9B8A'); return; }
        setPf({ photo: reader.result });
      };
      reader.onerror = () => this.notify('The profile photo could not be read.', '#FF9B8A');
      reader.readAsDataURL(f);
    },
    has, cal, ev, evOpen: s.evOpen,
    awardsOn, awardStrip, aw, awOpen: s.awOpen, openAward: set({ awOpen: true }),
    newEvent: () => openEv('Personal'), newMeeting: () => openEv('Meeting')
  };
}
