// Initial state: files.
export function filesState() {
  return {
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
    ]
  };
}
