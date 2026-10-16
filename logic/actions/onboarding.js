import { mapHeaders, parseCsv, structureOptions, templateCsv, toDirectoryEntry, validatePerson, MANDATORY } from '~/logic/onboarding.js';
import { todayIso } from '~/logic/projects.js';

const COLORS = ['#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#B8F36B', '#C9B6FF', '#FFD66B'];
const BLANK = { first: '', last: '', dob: '', gender: '', email: '', phone: '', title: '', sector: '', branch: '', unit: '', department: '' };

// Which structure levels the organization has, from the widgets it switched on.
export const hasLevels = (st) => ({ departments: st.widgets.indexOf('departments') !== -1, sectors: st.widgets.indexOf('sectors') !== -1, branches: st.widgets.indexOf('branches') !== -1, units: st.widgets.indexOf('units') !== -1 });

// When editing, a level the person already has a value for stays on the form even if its widget is off, so nothing they have is hidden.
export const hasLevelsFor = (st) => {
  const h = hasLevels(st), o = st.onboard;
  if (o && o.orig) { if (o.orig.sector) h.sectors = true; if (o.orig.branch) h.branches = true; if (o.orig.unit) h.units = true; if (o.orig.department) h.departments = true; }
  return h;
};

// HubLogic actions: onboarding people one at a time or in bulk. Mixed into HubLogic.prototype.
export const onboardingActions = {
  openOnboard(tab) {
    this.setState({ onboard: { tab: tab || 'one', f: { ...BLANK }, errs: {}, bulkText: '', bulk: null, fileName: '', done: null } });
  },
  // Edit an existing person from the People table, with the same fields the organization uses to onboard.
  openEditPerson(id) {
    const p = this.dir[id];
    if (!p) return;
    const parts = (p.name || '').split(' ');
    const first = p.first || parts[0] || '', last = p.last || parts.slice(1).join(' ');
    const f = { ...BLANK, first, last, dob: p.dob || '', gender: p.gender || '', email: p.email || '', phone: p.phone || '', title: p.title || '', sector: p.sector || '', branch: p.loc || '', unit: p.unit || '', department: p.dept || '' };
    this.setState({ onboard: { tab: 'one', edit: id, orig: { ...f }, f, roleDraft: null, rightsDraft: {}, errs: {}, bulkText: '', bulk: null, fileName: '', done: null } });
  },
  // Role and rights drafts while editing a person; saved together with the details.
  setOnboardRole(v) { this.setState((st) => ({ onboard: { ...st.onboard, roleDraft: v } })); },
  toggleOnboardRight(key, value) { this.setState((st) => ({ onboard: { ...st.onboard, rightsDraft: { ...st.onboard.rightsDraft, [key]: value } } })); },
  resetOnboardRights() { this.setState((st) => ({ onboard: { ...st.onboard, rightsDraft: Object.keys({ ...(st.personRights[this.dir[st.onboard.edit].name] || {}) }).reduce((m, k) => { m[k] = null; return m; }, {}) } })); },
  closeOnboard() { this.setState({ onboard: null }); },
  setOnboardTab(tab) { this.setState((st) => ({ onboard: { ...st.onboard, tab, done: null } })); },

  // Changing sector, branch or department clears choices below it that no longer fit.
  setOnboardField(k, v) {
    this.setState((st) => {
      const has = hasLevelsFor(st);
      const f = { ...st.onboard.f, [k]: v };
      const opts = structureOptions(this.org, has, f);
      if (['sector', 'branch', 'unit', 'department'].indexOf(k) !== -1) ['branch', 'unit', 'department'].forEach((c) => { if (c !== k && f[c] && opts[c].indexOf(f[c]) === -1) f[c] = ''; });
      if (k === 'department' && v && has.branches && !f.branch) { const d = this.org.departments.find((x) => x.name === v); const b = d && this.org.branches.find((x) => x.id === d.branch); if (b) f.branch = b.name; }
      if (k === 'branch' && v && has.sectors && !f.sector) { const b = this.org.branches.find((x) => x.name === v); const sc = b && this.org.sectors.find((x) => x.id === b.sector); if (sc) f.sector = sc.name; }
      const errs = { ...st.onboard.errs }; delete errs[k];
      return { onboard: { ...st.onboard, f, errs, done: null } };
    });
  },

  takenEmails() { return this.dir.map((p) => p.email).filter(Boolean).map((e) => e.toLowerCase()); },

  // Adds a validated person to the directory and the audit log.
  addPerson(clean) {
    const today = todayIso(Date.now());
    const entry = toDirectoryEntry(clean, this.dir.length, COLORS[this.dir.length % COLORS.length], today);
    this.dir.push(entry);
    this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now() + entry.id, ago: 0, user: 'Samwel Okoth Omulo', action: 'Onboarded ' + entry.name, module: 'People', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }].concat(st.extraLogs) }));
    return entry;
  },

  submitOnboard() {
    const ob = this.state.onboard;
    if (!ob) return;
    const editing = ob.edit !== undefined && ob.edit !== null;
    const own = editing ? (this.dir[ob.edit].email || '').toLowerCase() : '';
    const { errors, clean } = validatePerson(ob.f, this.org, hasLevelsFor(this.state), this.takenEmails().filter((e) => e !== own), todayIso(Date.now()));
    // Records created before these details were collected may have gaps; only insist on what was already filled in.
    if (editing) ['sector', 'branch', 'unit', 'department'].forEach((k) => { if (ob.orig[k] && ob.f[k] === ob.orig[k]) delete errors[k]; });
    if (editing) ['dob', 'gender', 'phone', 'email'].forEach((k) => { if (!ob.orig[k] && !String(ob.f[k] || '').trim()) delete errors[k]; });
    if (Object.keys(errors).length) { this.setState({ onboard: { ...ob, errs: errors, done: null } }); return; }
    if (editing) {
      const p = this.dir[ob.edit];
      Object.assign(p, { name: clean.first + ' ' + clean.last, first: clean.first, last: clean.last, initials: (clean.first[0] + clean.last[0]).toUpperCase(), title: clean.title || p.title, dept: clean.department || p.dept, loc: clean.branch || p.loc, sector: clean.sector, unit: clean.unit, email: clean.email, phone: clean.phone, dob: clean.dob, gender: clean.gender });
      const logs = [];
      if (ob.roleDraft) { this.setState((st) => ({ memberRoles: { ...st.memberRoles, [p.name]: ob.roleDraft } })); logs.push('Changed the role of ' + p.name + ' to ' + ob.roleDraft); }
      const rd = ob.rightsDraft || {};
      if (Object.keys(rd).length) {
        this.setState((st) => { const cur = { ...(st.personRights[p.name] || {}) }; Object.keys(rd).forEach((k) => { if (rd[k] === null) delete cur[k]; else cur[k] = rd[k]; }); const pr = { ...st.personRights }; if (Object.keys(cur).length) pr[p.name] = cur; else delete pr[p.name]; return { personRights: pr }; });
        logs.push('Changed individual rights of ' + p.name + ' (' + Object.keys(rd).length + ' change' + (Object.keys(rd).length === 1 ? '' : 's') + ')');
      }
      this.setState((st) => ({ onboard: null, extraLogs: logs.map((a, i) => ({ id: 'xr' + Date.now() + i, ago: 0, user: 'Samwel Okoth Omulo', action: a, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' })).concat([{ id: 'x' + Date.now() + p.id, ago: 0, user: 'Samwel Okoth Omulo', action: 'Edited details of ' + p.name, module: 'People', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }].concat(st.extraLogs)) }));
      this.forceUpdate();
      this.notify(p.name + '\'s details were saved.', '#B8F36B');
      return;
    }
    const entry = this.addPerson(clean);
    this.forceUpdate();
    this.setState({ onboard: { ...ob, f: { ...BLANK }, errs: {}, done: { one: entry.name } } });
    this.notify(entry.name + ' was onboarded and added to the directory.', '#B8F36B');
  },

  parseBulk(text) {
    const has = hasLevels(this.state);
    const rows = parseCsv(text);
    if (!rows.length) { this.setState((st) => ({ onboard: { ...st.onboard, bulkText: text, bulk: null } })); return; }
    const { map, unknown } = mapHeaders(rows[0]);
    const have = map.filter(Boolean);
    const needed = MANDATORY.map((m) => m.key);
    const missing = needed.filter((k) => have.indexOf(k) === -1);
    const bulk = { unknown, missing: missing.map((k) => MANDATORY.find((m) => m.key === k).label), rows: [] };
    if (!missing.length) {
      const seen = this.takenEmails();
      rows.slice(1).forEach((cells, i) => {
        const raw = {};
        map.forEach((k, ci) => { if (k) raw[k] = (cells[ci] || '').trim(); });
        const { errors, clean } = validatePerson(raw, this.org, has, seen, todayIso(Date.now()));
        if (!errors.email && clean.email) seen.push(clean.email);
        bulk.rows.push({ n: i + 2, clean, errors, ok: Object.keys(errors).length === 0, name: ((raw.first || '') + ' ' + (raw.last || '')).trim() || '(no name)', email: raw.email || '' });
      });
    }
    this.setState((st) => ({ onboard: { ...st.onboard, bulkText: text, bulk, done: null } }));
  },

  readBulkFile(file) {
    if (!file) return;
    const reader = new FileReader();
    reader.onload = () => { this.setState((st) => ({ onboard: { ...st.onboard, fileName: file.name } })); this.parseBulk(String(reader.result || '')); };
    reader.readAsText(file);
  },

  importBulk() {
    const ob = this.state.onboard;
    if (!ob || !ob.bulk) return;
    const good = ob.bulk.rows.filter((r) => r.ok);
    if (!good.length) return;
    good.forEach((r) => this.addPerson(r.clean));
    this.forceUpdate();
    const skipped = ob.bulk.rows.length - good.length;
    this.setState({ onboard: { ...ob, bulkText: '', bulk: null, fileName: '', done: { count: good.length, skipped } } });
    this.notify(good.length + (good.length === 1 ? ' person was' : ' people were') + ' onboarded' + (skipped ? '. ' + skipped + ' row' + (skipped === 1 ? ' was' : 's were') + ' skipped because of errors.' : '.'), '#B8F36B');
  },

  downloadOnboardTemplate() {
    const blob = new Blob([templateCsv(hasLevels(this.state))], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'onboarding-template.csv';
    document.body.appendChild(a); a.click(); a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
};
