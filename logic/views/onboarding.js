import { hasLevelsFor } from '~/logic/actions/onboarding.js';
import { GENDERS, MANDATORY, csvColumns, structureFields, structureOptions } from '~/logic/onboarding.js';

// renderVals step: onboarding. The form only shows the structure levels this organization has set up.
export function buildOnboarding(ctx) {
  const { can, modEnabled, modInfo, ridOf, rightOf, rightsDef, s, visibleRoles } = ctx;
  const has = hasLevelsFor(s);
  const ob = s.onboard;
  const canOnboard = can('people', 'Onboard people');
  const base = { canOnboard, open: false, openOne: () => this.openOnboard('one'), openBulk: () => this.openOnboard('bulk') };
  if (!ob) return { onb: base };
  const f = ob.f, errs = ob.errs;
  // Role and individual rights, only when editing someone who is already in the directory.
  let access = null;
  if (ob.edit !== undefined && ob.edit !== null && this.dir[ob.edit]) {
    const person = this.dir[ob.edit];
    const canEditRights = can('roles', 'Manage roles & rights');
    const curRole = ob.roleDraft || ridOf(person.name);
    const saved = s.personRights[person.name] || {}, rd = ob.rightsDraft || {};
    const valueOf = (m, r) => { const k = m + '|' + r; if (rd[k] !== undefined) return rd[k] === null ? rightOf(curRole, m, r) : rd[k]; if (saved[k] !== undefined) return saved[k]; return rightOf(curRole, m, r); };
    const modules = Object.keys(rightsDef).filter(modEnabled).map((m) => ({ name: modInfo[m][0], tint: modInfo[m][1],
      rights: rightsDef[m].map((r) => { const on = valueOf(m, r), base = rightOf(curRole, m, r); return { label: r, on, custom: on !== base, disabled: !canEditRights, toggle: () => this.toggleOnboardRight(m + '|' + r, !on === base ? null : !on) }; }) }));
    const custom = modules.reduce((n, m) => n + m.rights.filter((x) => x.custom).length, 0);
    access = { canEdit: canEditRights, roleValue: curRole, roles: visibleRoles.map((r) => ({ v: r.id, l: r.name })), onRole: (e) => this.setOnboardRole(e.target.value), roleChanged: curRole !== ridOf(person.name),
      modules, customCount: custom, hasCustom: custom > 0, customNote: custom + (custom === 1 ? ' right differs' : ' rights differ') + ' from the role', reset: () => this.setState((st) => ({ onboard: { ...st.onboard, rightsDraft: Object.keys({ ...saved, ...rd }).reduce((m, k) => { m[k] = null; return m; }, {}) } })),
      open: !!ob.rightsOpen, toggleOpen: () => this.setState((st) => ({ onboard: { ...st.onboard, rightsOpen: !st.onboard.rightsOpen } })) };
  }
  const opts = structureOptions(this.org, has, f);
  const structure = structureFields(has);
  const rowsAll = ob.bulk ? ob.bulk.rows : [];
  const good = rowsAll.filter((r) => r.ok).length;
  const bad = rowsAll.length - good;
  const isEdit = ob.edit !== undefined && ob.edit !== null;
  const onb = { ...base, open: true, isEdit, heading: isEdit ? 'Edit ' + this.dir[ob.edit].name : 'Onboard people', submitLabel: isEdit ? 'Save changes' : 'Onboard person', isOne: ob.tab === 'one', isBulk: ob.tab === 'bulk' && !isEdit, showTabs: !isEdit, close: () => this.closeOnboard(),
    goOne: () => this.setOnboardTab('one'), goBulk: () => this.setOnboardTab('bulk'),
    done: ob.done, hasDone: !!ob.done, doneOne: !!(ob.done && ob.done.one), doneBulk: !!(ob.done && ob.done.count !== undefined),
    fields: MANDATORY.map((m) => ({ key: m.key, label: m.label, type: m.type, isSelect: m.type === 'select', value: f[m.key], err: errs[m.key] || '', hasErr: !!errs[m.key], options: GENDERS,
      onInput: (e) => this.setOnboardField(m.key, e.target.value) })),
    title: f.title, onTitle: (e) => this.setOnboardField('title', e.target.value),
    structure: structure.map((x) => ({ key: x.key, label: x.label + (x.required ? '' : ' (optional)'), value: f[x.key], options: isEdit && ob.orig[x.key] && opts[x.key].indexOf(ob.orig[x.key]) === -1 && f[x.key] === ob.orig[x.key] ? opts[x.key].concat([ob.orig[x.key]]) : opts[x.key], err: errs[x.key] || '', hasErr: !!errs[x.key],
      onChange: (e) => this.setOnboardField(x.key, e.target.value) })),
    hasStructure: structure.length > 0, access, hasAccess: !!access,
    structureNote: structure.length ? 'These come from your organization: ' + structure.map((x) => x.label.toLowerCase()).join(', ') + '. Each choice narrows the next one.' : 'This organization has no sectors, branches, units or departments set up, so only the required details are needed.',
    submit: () => this.submitOnboard(),
    bulk: { text: ob.bulkText, onText: (e) => this.parseBulk(e.target.value), columns: csvColumns(has).join(', '), template: () => this.downloadOnboardTemplate(),
      onFile: (e) => { const file = e.target.files && e.target.files[0]; e.target.value = ''; this.readBulkFile(file); }, fileName: ob.fileName,
      structureNote: structure.length ? 'Sector, branch, unit and department names must match the ones in your organization.' : 'No structure columns are needed because this organization has none set up.',
      hasMissing: !!(ob.bulk && ob.bulk.missing.length), missingText: ob.bulk && ob.bulk.missing.length ? 'The first row must have a column for: ' + ob.bulk.missing.join(', ') + '.' : '',
      hasUnknown: !!(ob.bulk && ob.bulk.unknown.length), unknownText: ob.bulk && ob.bulk.unknown.length ? 'These columns are not used and will be ignored: ' + ob.bulk.unknown.join(', ') + '.' : '',
      hasRows: rowsAll.length > 0, validCount: good, invalidCount: bad,
      summary: rowsAll.length ? rowsAll.length + ' rows read · ' + good + ' ready · ' + bad + ' with errors' : '',
      rows: rowsAll.slice(0, 200).map((r) => ({ n: r.n, name: r.name, email: r.email, ok: r.ok, statusLabel: r.ok ? 'Ready' : 'Fix', statusColor: r.ok ? '#B8F36B' : '#FF9B8A', errorText: Object.keys(r.errors).map((k) => r.errors[k]).join(' ') })),
      moreRows: rowsAll.length > 200 ? rowsAll.length - 200 : 0,
      canImport: good > 0, importLabel: good ? 'Onboard ' + good + (good === 1 ? ' person' : ' people') + (bad ? ' (skip ' + bad + ')' : '') : 'Nothing to onboard yet',
      import: () => this.importBulk() } };
  return { onb };
}
