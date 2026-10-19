// Onboarding rules. First name, second name, date of birth, gender, email and phone are always required.
// Sector, branch, unit and department only exist on the form when the organization has set that level up.
export const GENDERS = ['Female', 'Male', 'Other', 'Prefer not to say'];
export const MANDATORY = [
  { key: 'first', label: 'First name', type: 'text', csv: 'first_name' },
  { key: 'last', label: 'Second name', type: 'text', csv: 'second_name' },
  { key: 'dob', label: 'Date of birth', type: 'date', csv: 'date_of_birth' },
  { key: 'gender', label: 'Gender', type: 'select', csv: 'gender' },
  { key: 'email', label: 'Email', type: 'email', csv: 'email' },
  { key: 'phone', label: 'Phone number', type: 'tel', csv: 'phone' }
];
// Parent first, so a choice can narrow the next one.
const STRUCTURE = [
  { key: 'sector', label: 'Sector', level: 'sectors', csv: 'sector' },
  { key: 'branch', label: 'Branch', level: 'branches', csv: 'branch' },
  { key: 'unit', label: 'Unit', level: 'units', csv: 'unit', optional: true },
  { key: 'department', label: 'Department', level: 'departments', csv: 'department' }
];
const ALIASES = {
  first: ['first', 'firstname', 'first name', 'first_name', 'given name', 'givenname'],
  last: ['second', 'secondname', 'second name', 'second_name', 'last', 'lastname', 'last name', 'last_name', 'surname'],
  dob: ['dob', 'date of birth', 'date_of_birth', 'dateofbirth', 'birthdate', 'birth date', 'birthday'],
  gender: ['gender', 'sex'], email: ['email', 'e-mail', 'email address'], phone: ['phone', 'phone number', 'phone_number', 'phonenumber', 'mobile', 'telephone', 'tel'],
  sector: ['sector'], branch: ['branch', 'location'], unit: ['unit'], department: ['department', 'dept'], title: ['title', 'job title', 'job_title', 'role', 'position']
};

export const structureFields = (has) => STRUCTURE.filter((f) => has[f.level]).map((f) => ({ ...f, required: !f.optional }));
export const csvColumns = (has) => MANDATORY.map((f) => f.csv).concat(structureFields(has).map((f) => f.csv));

const byName = (list) => (list || []).reduce((m, r) => { m[r.name.toLowerCase()] = r; return m; }, {});

// Options for each structure field, narrowed by what is already chosen above it.
export function structureOptions(org, has, v) {
  const sector = has.sectors && v.sector ? (org.sectors.find((x) => x.name === v.sector) || null) : null;
  const branch = has.branches && v.branch ? (org.branches.find((x) => x.name === v.branch) || null) : null;
  return {
    sector: org.sectors.map((x) => x.name),
    branch: org.branches.filter((b) => !sector || b.sector === sector.id).map((b) => b.name),
    unit: org.units.filter((u) => !branch || u.branch === branch.id).map((u) => u.name),
    department: org.departments.filter((d) => (!branch || d.branch === branch.id) && (!sector || !has.branches || branch || d.sector === sector.id)).map((d) => d.name)
  };
}

export function normalizeDob(raw) {
  const s = String(raw || '').trim();
  let m = s.match(/^(\d{4})-(\d{1,2})-(\d{1,2})$/);
  let y, mo, d;
  if (m) { y = +m[1]; mo = +m[2]; d = +m[3]; } else {
    m = s.match(/^(\d{1,2})[/.-](\d{1,2})[/.-](\d{4})$/);
    if (!m) return '';
    d = +m[1]; mo = +m[2]; y = +m[3];
  }
  const dt = new Date(Date.UTC(y, mo - 1, d));
  if (dt.getUTCFullYear() !== y || dt.getUTCMonth() !== mo - 1 || dt.getUTCDate() !== d) return '';
  return y + '-' + String(mo).padStart(2, '0') + '-' + String(d).padStart(2, '0');
}

export function normalizeGender(raw) {
  const g = String(raw || '').trim().toLowerCase();
  if (['f', 'female', 'woman', 'girl'].indexOf(g) !== -1) return 'Female';
  if (['m', 'male', 'man', 'boy'].indexOf(g) !== -1) return 'Male';
  if (['other', 'non-binary', 'nonbinary', 'x'].indexOf(g) !== -1) return 'Other';
  if (['prefer not to say', 'not specified', 'n/a', 'unspecified'].indexOf(g) !== -1) return 'Prefer not to say';
  return '';
}

export function normalizePhone(raw) {
  const s = String(raw || '').trim();
  const digits = s.replace(/[^\d]/g, '');
  return /^\+?[\d\s()-]{7,20}$/.test(s) && digits.length >= 7 && digits.length <= 15 ? (s.startsWith('+') ? '+' : '') + digits : '';
}

// Returns { errors: { field: message }, clean }. `taken` is the list of emails already in the directory (lower-case).
export function validatePerson(p, org, has, taken, todayIso) {
  const errors = {};
  const clean = { ...p };
  clean.first = (p.first || '').trim(); clean.last = (p.last || '').trim();
  if (!clean.first) errors.first = 'First name is required.';
  if (!clean.last) errors.last = 'Second name is required.';
  clean.dob = normalizeDob(p.dob);
  if (!String(p.dob || '').trim()) errors.dob = 'Date of birth is required.';
  else if (!clean.dob) errors.dob = 'Use a real date, like 1994-03-21 or 21/03/1994.';
  else if (clean.dob > todayIso) errors.dob = 'Date of birth cannot be in the future.';
  else if (clean.dob > String(Number(todayIso.slice(0, 4)) - 16) + todayIso.slice(4)) errors.dob = 'A person must be at least 16 years old.';
  clean.gender = normalizeGender(p.gender);
  if (!String(p.gender || '').trim()) errors.gender = 'Gender is required.';
  else if (!clean.gender) errors.gender = 'Use Female, Male, Other or Prefer not to say.';
  clean.email = (p.email || '').trim().toLowerCase();
  if (!clean.email) errors.email = 'Email is required.';
  else if (!/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(clean.email)) errors.email = 'That email does not look right.';
  else if (taken.indexOf(clean.email) !== -1) errors.email = 'Someone with this email is already in the directory.';
  clean.phone = normalizePhone(p.phone);
  if (!String(p.phone || '').trim()) errors.phone = 'Phone number is required.';
  else if (!clean.phone) errors.phone = 'Use 7 to 15 digits, with an optional + at the start.';

  // Structure: only what the organization has, and each choice must exist and sit under its parent.
  const sectors = byName(org.sectors), branches = byName(org.branches), depts = byName(org.departments), units = byName(org.units);
  const sec = has.sectors && p.sector ? sectors[String(p.sector).trim().toLowerCase()] : null;
  let br = has.branches && p.branch ? branches[String(p.branch).trim().toLowerCase()] : null;
  const un = has.units && p.unit ? units[String(p.unit).trim().toLowerCase()] : null;
  const dp = has.departments && p.department ? depts[String(p.department).trim().toLowerCase()] : null;
  if (has.sectors) { if (!String(p.sector || '').trim()) errors.sector = 'Choose a sector.'; else if (!sec) errors.sector = 'This organization has no sector called "' + p.sector + '".'; }
  if (has.branches) {
    if (!String(p.branch || '').trim()) errors.branch = 'Choose a branch.';
    else if (!br) errors.branch = 'This organization has no branch called "' + p.branch + '".';
    else if (sec && br.sector !== sec.id) errors.branch = br.name + ' is not in the ' + sec.name + ' sector.';
  }
  if (has.departments) {
    if (!String(p.department || '').trim()) errors.department = 'Choose a department.';
    else if (!dp) errors.department = 'This organization has no department called "' + p.department + '".';
    else if (br && dp.branch !== br.id) errors.department = dp.name + ' is not in ' + br.name + '.';
  }
  if (has.units && String(p.unit || '').trim()) {
    if (!un) errors.unit = 'This organization has no unit called "' + p.unit + '".';
    else if (br && un.branch !== br.id) errors.unit = un.name + ' is not in ' + br.name + '.';
  }
  // Fill the parents from the child when they were left out.
  if (!br && dp && has.branches) br = org.branches.find((b) => b.id === dp.branch) || null;
  const sector = sec || (br && has.sectors ? org.sectors.find((x) => x.id === br.sector) : null);
  clean.sector = sector ? sector.name : '';
  clean.branch = br ? br.name : '';
  clean.unit = un ? un.name : '';
  clean.department = dp ? dp.name : '';
  clean.title = (p.title || '').trim();
  return { errors, clean };
}

// Directory entry for a validated person. `loc` is the branch, as everywhere else in the directory.
export function toDirectoryEntry(clean, id, color, todayIso) {
  return { id, name: clean.first + ' ' + clean.last, first: clean.first, last: clean.last, initials: (clean.first[0] + clean.last[0]).toUpperCase(), color,
    title: clean.title || 'New joiner', dept: clean.department, loc: clean.branch, sector: clean.sector, unit: clean.unit, status: 'Available',
    email: clean.email, phone: clean.phone, dob: clean.dob, gender: clean.gender, joined: todayIso };
}

// ---- CSV ----
export function parseCsv(text) {
  const rows = []; let row = [], cell = '', q = false;
  const s = String(text || '').replace(/^﻿/, '');
  for (let i = 0; i < s.length; i++) {
    const c = s[i];
    if (q) { if (c === '"' && s[i + 1] === '"') { cell += '"'; i++; } else if (c === '"') q = false; else cell += c; }
    else if (c === '"') q = true;
    else if (c === ',' || c === ';' || c === '\t') { row.push(cell); cell = ''; }
    else if (c === '\n' || c === '\r') { if (c === '\r' && s[i + 1] === '\n') i++; row.push(cell); cell = ''; if (row.some((x) => x.trim() !== '')) rows.push(row); row = []; }
    else cell += c;
  }
  row.push(cell); if (row.some((x) => x.trim() !== '')) rows.push(row);
  return rows;
}

// Maps the header row to field keys. Unknown columns are reported so nothing is silently dropped.
export function mapHeaders(headerRow) {
  const map = []; const unknown = [];
  headerRow.forEach((h, i) => {
    const norm = String(h).trim().toLowerCase();
    const key = Object.keys(ALIASES).find((k) => ALIASES[k].indexOf(norm) !== -1);
    if (key) map[i] = key; else if (norm) unknown.push(String(h).trim());
  });
  return { map, unknown };
}

export function templateCsv(has) {
  const cols = csvColumns(has);
  const sample = { first_name: 'Amani', second_name: 'Kariuki', date_of_birth: '1994-03-21', gender: 'Female', email: 'amani.kariuki@company.com', phone: '+254700123456', sector: '', branch: '', unit: '', department: '' };
  return cols.join(',') + '\n' + cols.map((c) => sample[c] || '').join(',') + '\n';
}
