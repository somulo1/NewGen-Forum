import * as foldersW from '../server/widgets/folders.js';
// Values and handlers the templates read through hubBindings: projects.
export function valsProjects(ctx) {
  const { cur_f, cur_p, folders, forumsData, fv, msd, openProjectForm, projectForm, pj, projStats, projects, s, set } = ctx;
  return {
    forumsOpen: s.forumsOpen,
    toggleForums: () => this.setState((st) => ({ forumsOpen: !st.forumsOpen })),
    forumsChevron: 'transition:transform 0.2s;transform:rotate(' + (s.forumsOpen ? '0' : '-90') + 'deg)',
    forumCount: s.forumsOpen ? String(forumsData.length) : String(forumsData.reduce((n, f) => n + f.unread, 0)) + ' new',
    projects, projStats, pj, projectForm, openProjectForm,
    showProjectList: ctx.has.projects && s.section === 'projects' && !cur_p,
    showProjectView: ctx.has.projects && s.section === 'projects' && !!cur_p,
    closeProject: set({ projectOpen: null }),
    msd,
    folders, fv,
    showFolderList: s.section === 'files' && !cur_f,
    showFolderView: s.section === 'files' && !!cur_f,
    closeFolder: set({ folder: null }),
    deleteFile: (f) => { if (this.srv && f.fileId) foldersW.remove(this, this.srv, f).then(() => this.notify('File deleted.', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not delete it.', '#FF9B8A')); },
    shareFile: (f) => { if (this.srv && f.fileId) foldersW.share(this.srv, f).then((r) => this.notify(r.pending ? 'Link sent for approval.' : 'Link copied (works for 24 hours).', '#B8F36B')).catch((e) => this.notify(e.message || 'Could not share it.', '#FF9B8A')); },
    openFile: (f) => { if (this.srv && f.fileId) foldersW.open(this.srv, f).catch((e) => this.notify(e.message || 'Could not open it.', '#FF9B8A')); },
    uploadToFolder: (e) => {
      const list = Array.from(e.target.files || []);
      e.target.value = '';
      if (!cur_f || cur_f.isNotes || !list.length) return;
      if (this.srv && /^[0-9a-f]{8}-/.test(cur_f.id)) { foldersW.upload(this, this.srv, cur_f.id, list).then(() => this.notify(list.length + (list.length === 1 ? ' file' : ' files') + ' uploaded to ' + cur_f.name + '.', '#B8F36B')).catch((x) => this.notify(x.message || 'Upload failed.', '#FF9B8A')); return; }
      const kindOf = (f) => { const n = f.name.toLowerCase(); if ((f.type || '').indexOf('image/') === 0) return 'IMG'; if (/\.pdf$/.test(n)) return 'PDF'; if (/\.(docx?|txt|rtf)$/.test(n)) return 'DOC'; if (/\.(xlsx?|csv)$/.test(n)) return 'XLS'; if (/\.(pptx?|key)$/.test(n)) return 'PPT'; return 'ZIP'; };
      const added = list.map((f) => { const k = kindOf(f); return { name: f.name, kind: k, size: this.fmtSize(f.size), by: 'Sam', when: 'Just now', url: k === 'IMG' ? URL.createObjectURL(f) : '' }; });
      this.setState((st) => ({ folders: st.folders.map((f) => (f.id === cur_f.id ? { ...f, updated: 'just now', files: added.concat(f.files) } : f)) }));
      this.notify(added.length + (added.length === 1 ? ' file' : ' files') + ' uploaded to ' + cur_f.name + '.', '#B8F36B');
    }
  };
}
