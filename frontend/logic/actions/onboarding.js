import * as peopleW from '~/logic/server/widgets/people.js';
import { mapHeaders, parseCsv, structureOptions, templateCsv, toDirectoryEntry, validatePerson, MANDATORY } from '~/logic/onboarding.js';
import { todayIso } from '~/logic/projects.js';

const COLORS = ['#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#B8F36B', '#C9B6FF', '#FFD66B'];
const BLANK = { first: '', last: '', dob: '', gender: '', email: '', phone: '', title: '', sector: '', branch: '', unit: '', department: '' };

// Which structure levels the organization has: the ones it has created (People > Organization structure). Without the
// server connected, the widgets it switched on decide, as in the sample data.
export const hasLevels = (st, org) => {
  const on = (k) => (org ? (org[k] || []).length > 0 : st.widgets.indexOf(k) !== -1);
  return { departments: on('departments'), sectors: on('sectors'), branches: on('branches'), units: on('units') };
};

// When editing, a level the person already has a value for stays on the form even if the organization has since removed it,
// so nothing they have is hidden.
export const hasLevelsFor = (st, org) => {
  const h = hasLevels(st, org), o = st.onboard;
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
    if (this.srv && p.uid) peopleW.loadPersonRights(this, this.srv, p).catch(() => {});   // their personal rights, from the server
  },
  // Role and rights drafts while editing a person; saved together with the details.
  setOnboardReport(v) { this.setState((st) => ({ onboard: { ...st.onboard, reportDraft: v } })); },
  setOnboardRole(v) { this.setState((st) => ({ onboard: { ...st.onboard, roleDraft: v } })); },
  toggleOnboardRight(key, value) { this.setState((st) => ({ onboard: { ...st.onboard, rightsDraft: { ...st.onboard.rightsDraft, [key]: value } } })); },
  resetOnboardRights() { this.setState((st) => ({ onboard: { ...st.onboard, rightsDraft: Object.keys({ ...(st.personRights[this.dir[st.onboard.edit].name] || {}) }).reduce((m, k) => { m[k] = null; return m; }, {}) } })); },
  closeOnboard() { this.setState({ onboard: null }); },
  setOnboardTab(tab) { this.setState((st) => ({ onboard: { ...st.onboard, tab, done: null } })); },

  // Changing sector, branch or department clears choices below it that no longer fit.
  setOnboardField(k, v) {
    this.setState((st) => {
      const has = hasLevelsFor(st, this.srv ? this.org : null);
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

  // Applies an approved (or approval-free) edit of a person: details, role, individual rights and who they report to.
  applyPersonEdit(ch) {
    const p = this.dir[ch.id];
    if (!p) return;
    const before = p.name;
    Object.assign(p, ch.details);
    const logs = [];
    if (ch.role) { this.setState((st) => ({ memberRoles: { ...st.memberRoles, [p.name]: ch.role } })); logs.push('Changed the role of ' + p.name + ' to ' + ch.role); }
    if (ch.report !== null && ch.report !== undefined) {
      this.setState((st) => { const pr = { ...st.personReports }; if (ch.report) pr[p.name] = ch.report; else delete pr[p.name]; return { personReports: pr }; });
      logs.push(ch.report ? 'Set ' + ch.report + ' as the person ' + p.name + ' reports to' : 'Cleared the specific manager of ' + p.name + ' (reports to the role)');
    }
    const rd = ch.rights || {};
    if (Object.keys(rd).length) {
      this.setState((st) => { const cur = { ...(st.personRights[p.name] || {}) }; Object.keys(rd).forEach((k) => { if (rd[k] === null) delete cur[k]; else cur[k] = rd[k]; }); const pr = { ...st.personRights }; if (Object.keys(cur).length) pr[p.name] = cur; else delete pr[p.name]; return { personRights: pr }; });
      logs.push('Changed individual rights of ' + p.name + ' (' + Object.keys(rd).length + ' change' + (Object.keys(rd).length === 1 ? '' : 's') + ')');
    }
    this.setState((st) => ({ extraLogs: logs.map((a, i) => ({ id: 'xr' + Date.now() + i, ago: 0, user: 'Samwel Okoth Omulo', action: a, module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' })).concat([{ id: 'x' + Date.now() + p.id, ago: 0, user: 'Samwel Okoth Omulo', action: 'Edited details of ' + before, module: 'People', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }], st.extraLogs) }));
    this.forceUpdate();
    this.notify(p.name + '\'s changes are now in effect.', '#B8F36B');
  },

  submitOnboard() {
    const ob = this.state.onboard;
    if (!ob) return;
    const editing = ob.edit !== undefined && ob.edit !== null;
    const own = editing ? (this.dir[ob.edit].email || '').toLowerCase() : '';
    const { errors, clean } = validatePerson(ob.f, this.org, hasLevelsFor(this.state, this.srv ? this.org : null), this.takenEmails().filter((e) => e !== own), todayIso(Date.now()));
    // Records created before these details were collected may have gaps; only insist on what was already filled in.
    if (editing) ['sector', 'branch', 'unit', 'department'].forEach((k) => { if (ob.orig[k] && ob.f[k] === ob.orig[k]) delete errors[k]; });
    if (editing) ['dob', 'gender', 'phone', 'email'].forEach((k) => { if (!ob.orig[k] && !String(ob.f[k] || '').trim()) delete errors[k]; });
    if (Object.keys(errors).length) { this.setState({ onboard: { ...ob, errs: errors, done: null } }); return; }
    if (editing) {
      const p = this.dir[ob.edit];
      const details = { name: clean.first + ' ' + clean.last, first: clean.first, last: clean.last, initials: (clean.first[0] + clean.last[0]).toUpperCase(), title: clean.title || p.title, dept: clean.department || p.dept, loc: clean.branch || p.loc, sector: clean.sector, unit: clean.unit, email: clean.email, phone: clean.phone, dob: clean.dob, gender: clean.gender };
      const lines = [];
      [['name', 'Name'], ['title', 'Job title'], ['dept', 'Department'], ['loc', 'Branch'], ['sector', 'Sector'], ['unit', 'Unit'], ['email', 'Email'], ['phone', 'Phone'], ['dob', 'Date of birth'], ['gender', 'Gender']].forEach(([k, l]) => {
        if (String(details[k] || '') !== String(p[k] || '')) lines.push(l + ': ' + (p[k] || '—') + ' → ' + (details[k] || '—'));
      });
      if (ob.roleDraft && ob.roleDraft !== (this.state.memberRoles[p.name] || ob.roleDraft)) lines.push('Role changed to ' + ob.roleDraft);
      else if (ob.roleDraft) lines.push('Role set to ' + ob.roleDraft);
      const rd = ob.rightsDraft || {};
      if (Object.keys(rd).length) lines.push(Object.keys(rd).length + ' individual right change' + (Object.keys(rd).length === 1 ? '' : 's'));
      if (ob.reportDraft !== undefined && ob.reportDraft !== null && ob.reportDraft !== (this.state.personReports[p.name] || '')) lines.push(ob.reportDraft ? 'Reports to ' + ob.reportDraft : 'Reports to the role');
      if (!lines.length) { this.setState({ onboard: null }); this.notify('No changes to ' + p.name + '.', '#9DB0A6'); return; }
      if (this.srv && p.uid) {
        // saved on the server (its approval flow may hold the change for sign-off). The form stays open while it saves, and keeps
        // everything typed if the server refuses, so nothing has to be typed again.
        if (ob.saving) return;
        this.setState({ onboard: { ...ob, saving: true, formError: '' } });
        peopleW.update(this, this.srv, p, clean, ob.orig, ob.roleDraft || null, { rights: rd, report: ob.reportDraft })
          .then((r) => {
            this.setState({ onboard: null });
            this.notify(r === 'pending' ? 'Changes to ' + p.name + ' were sent for approval.' : r === 'unchanged' ? 'No changes to save for ' + p.name + '.' : p.name + '\'s changes were saved.', r === 'pending' ? '#FFB86B' : r === 'unchanged' ? '#9DB0A6' : '#B8F36B');
          })
          .catch((e) => {
            const msg = (e && e.message) || 'Could not save the changes.';
            this.setState((st) => ({ onboard: st.onboard ? { ...st.onboard, saving: false, formError: msg } : st.onboard }));
            this.notify(msg, '#FF9B8A');
          });
        return;
      }
      const pending = this.state.requests.find((r) => r.kind === 'personEdit' && r.personId === ob.edit && r.status === 'Pending');
      if (pending) { this.notify('Changes to ' + p.name + ' are already waiting for approval (' + pending.id + '). Wait for it to finish first.', '#FFB86B'); return; }
      const changes = { id: ob.edit, name: p.name, details, role: ob.roleDraft || null, rights: rd, report: ob.reportDraft === undefined ? null : ob.reportDraft };
      const flow = this._flowById ? this._flowById('f11') : null;
      this.setState({ onboard: null });
      if (flow && flow.levels > 0 && this._makeRequest) {
        // Nothing changes yet: the edit takes effect when the last approver signs off.
        this._makeRequest('f11', 'Edit ' + p.name + ': ' + lines.join('; '), { kind: 'personEdit', personId: ob.edit, changes, note: lines.join('\n') });
        return;
      }
      this.applyPersonEdit(changes);
      return;
    }
    if (this.srv) {
      // saved on the server: the person gets an active account and an email with their first password
      peopleW.onboard(this, this.srv, [clean]).then((res) => {
        const r = res[0] || {};
        if (r.status !== 'invited') { this.notify(r.error || 'Could not onboard this person.', '#FF9B8A'); return; }
        const name = clean.first + ' ' + clean.last;
        this.setState({ onboard: { ...ob, f: { ...BLANK }, errs: {}, done: { one: name } } });
        this.notify(name + ' was onboarded. Their first password was emailed to them.', '#B8F36B');
      }).catch((e) => this.notify((e && e.message) || 'Could not onboard this person.', '#FF9B8A'));
      return;
    }
    const entry = this.addPerson(clean);
    this.forceUpdate();
    this.setState({ onboard: { ...ob, f: { ...BLANK }, errs: {}, done: { one: entry.name } } });
    this.notify(entry.name + ' was onboarded and added to the directory.', '#B8F36B');
  },

  parseBulk(text) {
    const has = hasLevels(this.state, this.srv ? this.org : null);
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
    if (this.srv) {
      peopleW.onboard(this, this.srv, good.map((r) => r.clean)).then((res) => {
        const ok = res.filter((r) => r.status === 'invited').length;
        const skipped = ob.bulk.rows.length - ok;
        this.setState({ onboard: { ...ob, bulkText: '', bulk: null, fileName: '', done: { count: ok, skipped } } });
        const why = res.find((r) => r.status !== 'invited');
        this.notify(ok + (ok === 1 ? ' person was' : ' people were') + ' onboarded and emailed their first password' + (skipped ? '. ' + skipped + ' skipped' + (why && why.error ? ': ' + why.error : '') : '.'), skipped ? '#FFB86B' : '#B8F36B');
      }).catch((e) => this.notify((e && e.message) || 'Could not import these people.', '#FF9B8A'));
      return;
    }
    good.forEach((r) => this.addPerson(r.clean));
    this.forceUpdate();
    const skipped = ob.bulk.rows.length - good.length;
    this.setState({ onboard: { ...ob, bulkText: '', bulk: null, fileName: '', done: { count: good.length, skipped } } });
    this.notify(good.length + (good.length === 1 ? ' person was' : ' people were') + ' onboarded' + (skipped ? '. ' + skipped + ' row' + (skipped === 1 ? ' was' : 's were') + ' skipped because of errors.' : '.'), '#B8F36B');
  },

  downloadOnboardTemplate() {
    const blob = new Blob([templateCsv(hasLevels(this.state, this.srv ? this.org : null))], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url; a.download = 'onboarding-template.csv';
    document.body.appendChild(a); a.click(); a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
  }
};
