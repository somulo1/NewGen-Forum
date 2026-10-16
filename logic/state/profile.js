// Initial state: profile.
export function profileState() {
  return {
    forumsOpen: true, folder: null, projectOpen: null, newTask: '', msDialog: null, taskForm: null, sprintFilter: 'all', objDraft: {},
    profileEdit: false, statusMenu: false, newSkill: '',
    profile: {
      status: 'Available', cover: '#1F3A2B', photo: '', phone: '',
      bio: 'I look after cybersecurity and technology risk across our systems, from access reviews and audits to making sure new tools go live safely.\nHappy to help with anything security related. Just tag me in a ticket or ping me in Risk & Compliance.',
      skills: ['Cybersecurity', 'Technology risk', 'IT audit', 'Access governance', 'Business continuity', 'Laravel / PHP', 'Vue 3', 'SIEM monitoring']
    }
  };
}
