// Initial state: profile.
export function profileState() {
  return {
    forumsOpen: false, folder: null, projectOpen: null, projectForm: null, msDialog: null, taskForm: null, sprintFilter: 'all', projTab: 'milestones', wgFilter: 'all', wgq: '', hiddenWidgets: [], standupForm: null,
    profileEdit: false, statusMenu: false, newSkill: '',
    profile: {
      status: 'Available', cover: '#1F3A2B', photo: '', phone: '',
      bio: '', skills: []
    }
  };
}
