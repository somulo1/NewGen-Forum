import * as profileW from '~/logic/server/widgets/profile.js';
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
    previewOptions: (s.serverRoles || []).map((r) => ({ v: r.id, l: r.name + (r.id === s.myRole ? ' (you)' : '') })),
    previewing: !!s.myRole && s.actingRole !== s.myRole, actingRole: s.actingRole, actingRoleName: roleNames[s.actingRole] || s.actingRole,
    setActingRole: (e) => { const v = e.target.value; this.setState({ actingRole: v }); if (v !== s.myRole) this.notify('Previewing Hub as ' + (roleNames[v] || v) + '.', '#FFD66B'); },
    exitPreview: () => this.setState({ actingRole: this.state.myRole }),
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
    cancelProfile: () => { if (this.srv) profileW.revert(this, this.srv); this.setState({ profileEdit: false }); },
    saveProfile: () => {
      if (!this.srv) return;
      profileW.save(this, this.srv)
        .then(() => { this.setState({ profileEdit: false }); this.notify('Profile updated.', '#B8F36B'); })
        .catch((e) => this.notify(e.message || 'Could not save your profile.', '#FF9B8A'));
    },
    copyEmail: () => {
      const email = me.email || '';
      if (!email || email.startsWith('Not')) { this.notify('No email to copy.', '#FFB86B'); return; }
      const done = () => this.notify('Email copied: ' + email, '#7CC4FF');
      if (navigator.clipboard) navigator.clipboard.writeText(email).then(done).catch(() => this.notify(email, '#7CC4FF')); else this.notify(email, '#7CC4FF');
    },
    shareProfile: () => {
      const url = window.location.origin + '/profile';
      const done = () => this.notify('Profile link copied. Paste it anywhere in Hub.', '#7CC4FF');
      if (navigator.clipboard) navigator.clipboard.writeText(url).then(done).catch(() => this.notify(url, '#7CC4FF')); else this.notify(url, '#7CC4FF');
    },
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
      profileW.photoDataUrl(f).then((url) => {
        setPf({ photo: url });
        // outside edit mode the photo is saved straight away
        if (!this.state.profileEdit && this.srv) profileW.save(this, this.srv).then(() => this.notify('Profile photo updated.', '#B8F36B')).catch((x) => this.notify(x.message, '#FF9B8A'));
      }).catch((x) => this.notify(x.message, '#FF9B8A'));
    },
    has, cal, ev, evOpen: s.evOpen,
    awardsOn, awardStrip, aw, awOpen: s.awOpen, openAward: set({ awOpen: true }),
    newEvent: () => openEv('Personal'), newMeeting: () => openEv('Meeting')
  };
}
