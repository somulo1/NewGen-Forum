import { mapHeaders, parseCsv, structureOptions, templateCsv, toDirectoryEntry, validatePerson, MANDATORY } from '~/logic/onboarding.js';
import { todayIso } from '~/logic/projects.js';

const COLORS = ['#7CC4FF', '#FFB86B', '#F49AC1', '#6EE7C8', '#B8F36B', '#C9B6FF', '#FFD66B'];
const BLANK = { first: '', last: '', dob: '', gender: '', email: '', phone: '', title: '', sector: '', branch: '', unit: '', department: '' };

// Which structure levels the organization has, from the widgets it switched on.
export const hasLevels = (st) => ({ departments: st.widgets.indexOf('departments') !== -1, sectors: st.widgets.indexOf('sectors') !== -1, branches: st.widgets.indexOf('branches') !== -1, units: st.widgets.indexOf('units') !== -1 });

// HubLogic actions: onboarding people one at a time or in bulk. Mixed into HubLogic.prototype.
export const onboardingActions = {
  openOnboard(tab) {
    this.setState({ onboard: { tab: tab || 'one', f: { ...BLANK }, errs: {}, bulkText: '', bulk: null, fileName: '', done: null } });
  },
  closeOnboard() { this.setState({ onboard: null }); },
  setOnboardTab(tab) { this.setState((st) => ({ onboard: { ...st.onboard, tab, done: null } })); },

  // Changing sector, branch or department clears choices below it that no longer fit.
  setOnboardField(k, v) {
    this.setState((st) => {
      const has = hasLevels(st);
      const f = { ...st.onboard.f, [k]: v };
      const opts = structureOptions(this.org, has, f);
      ['branch', 'unit', 'department'].forEach((c) => { if (c !== k && f[c] && opts[c].indexOf(f[c]) === -1) f[c] = ''; });
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
    const { errors, clean } = validatePerson(ob.f, this.org, hasLevels(this.state), this.takenEmails(), todayIso(Date.now()));
    if (Object.keys(errors).length) { this.setState({ onboard: { ...ob, errs: errors, done: null } }); return; }
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
