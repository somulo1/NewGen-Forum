import { hasLevelsFor } from '~/logic/actions/onboarding.js';
import * as orgW from '~/logic/server/widgets/org.js';
import { GENDERS, MANDATORY, csvColumns, structureFields, structureOptions } from '~/logic/onboarding.js';

// renderVals step: onboarding. The form only shows the structure levels this organization has set up.
export function buildOnboarding(ctx) {
  const { can, modEnabled, modInfo, reportRoleName, ridOf, rightOf, rightsDef, roleHolders, s, targetRoleOf, visibleRoles } = ctx;
  const has = hasLevelsFor(s, this.srv ? this.org : null);
  const ob = s.onboard;
  const canOnboard = can('people', 'Onboard people');
  const base = { canOnboard, open: false, openOne: () => this.openOnboard('one'), openBulk: () => this.openOnboard('bulk') };
  const structureEd = buildStructureEd.call(this, s, can);
  if (!ob) return { onb: base, structureEd };
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
    // Who they report to: the role by default, or one specific person in that role.
    const tRole = targetRoleOf(person.name);
    const savedPick = s.personReports[person.name] || '';
    const pickNow = ob.reportDraft !== undefined && ob.reportDraft !== null ? ob.reportDraft : savedPick;
    const reports = tRole ? { show: true, roleName: reportRoleName(tRole), value: pickNow, canEdit: canEditRights,
      options: [{ v: '', l: 'Anyone in the ' + reportRoleName(tRole) + ' role (shown as the role)' }].concat(roleHolders(tRole).filter((h) => h.name !== person.name).slice(0, 80).map((h) => ({ v: h.name, l: h.name + (h.dept ? ' · ' + h.dept : '') }))),
      onChange: (e) => this.setOnboardReport(e.target.value) } : { show: false };
    access = { reports, canEdit: canEditRights, roleValue: curRole, roles: visibleRoles.map((r) => ({ v: r.id, l: r.name })), onRole: (e) => this.setOnboardRole(e.target.value), roleChanged: curRole !== ridOf(person.name),
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
  return { onb, structureEd };
}

// The Organization structure editor on the People page: which sectors, branches, units and departments the organization has.
// Each level appears across Hub (pages, onboarding, imports) as soon as it has one.
const KINDS = [
  { kind: 'sector', label: 'Sector', plural: 'Sectors', level: 'sectors', field: 'sector', blurb: 'The largest groupings, such as Business or Operations.' },
  { kind: 'branch', label: 'Branch', plural: 'Branches', level: 'branches', field: 'loc', blurb: 'Places or offices. They can sit under a sector.' },
  { kind: 'unit', label: 'Unit', plural: 'Units', level: 'units', field: 'unit', blurb: 'Teams inside a branch (or a sector, if there are no branches).' },
  { kind: 'department', label: 'Department', plural: 'Departments', level: 'departments', field: 'dept', blurb: 'Functions such as Finance or IT. They sit under a branch, unit or sector, or stand alone.' }
];
const PARENTS = { sector: [], branch: ['sector'], unit: ['branch', 'sector'], department: ['branch', 'unit', 'sector'] };

const FRESH_ED = { kind: 'sector', name: '', parent: '', err: '', ok: '', busy: false, edit: null, del: null, move: {}, sel: {}, into: {}, mergeAsk: null };

// The Manage organization structure page (People > Organization structure): add, rename, move and merge sectors,
// branches, units and departments, and remove empty ones. Everything is checked again by the server.
function buildStructureEd(s, can) {
  const canEdit = !!this.srv && can('settings', 'Edit organization profile');
  const openIt = () => this.setState({ section: 'orgedit', structEd: { ...FRESH_ED } });
  const closed = { canEdit, open: false, openIt };
  if (s.section !== 'orgedit' || !canEdit) return closed;
  const se = { ...FRESH_ED, ...(s.structEd || {}) };
  const units = this.srv.unitList || [];
  const byId = new Map(units.map((u) => [u.id, u]));
  const kindLabel = (k) => (KINDS.find((x) => x.kind === k) || {}).label || k;
  const set = (patch) => this.setState((st) => ({ structEd: { ...FRESH_ED, ...(st.structEd || {}), ...(typeof patch === 'function' ? patch({ ...FRESH_ED, ...(st.structEd || {}) }) : patch) } }));
  // one change at a time; the result (or the server's reason) is shown at the top of the page
  const run = (fn, after, ok) => {
    if (se.busy) return Promise.resolve();
    set({ busy: true, err: '', ok: '' });
    return fn().then(() => set({ busy: false, ok: ok || '', ...(after || {}) })).catch((e) => set({ busy: false, err: (e && e.message) || 'That did not work.' }));
  };
  const have = (k) => units.filter((u) => u.kind === k);
  const parentChoices = (kind, self) => PARENTS[kind].flatMap((k) => have(k).filter((u) => u.id !== self).map((u) => ({ id: u.id, label: kindLabel(k) + ' · ' + u.name })));
  const parentOpts = parentChoices(se.kind);
  const parent = parentOpts.some((o) => o.id === se.parent) ? se.parent : '';
  const levels = KINDS.map((K) => {
    const list = have(K.kind).sort((a, b) => a.name.localeCompare(b.name));
    const sel = (se.sel[K.kind] || []).filter((id) => list.some((u) => u.id === id));
    const rows = list.map((u) => {
      const par = byId.get(u.parent_id);
      const n = this.dir.filter((p) => p[K.field] === u.name).length;
      const editing = se.edit && se.edit.id === u.id;
      const moveOpts = parentChoices(K.kind, u.id);
      const moveTo = se.move[u.id] !== undefined ? se.move[u.id] : (u.parent_id || '');
      const moved = moveTo !== (u.parent_id || '');
      return {
        id: u.id, name: u.name, under: par ? kindLabel(par.kind) + ' ' + par.name : 'Top level', people: n + (n === 1 ? ' person' : ' people'), editing, editName: editing ? se.edit.name : u.name,
        startEdit: () => set({ edit: { id: u.id, name: u.name }, err: '', ok: '' }), cancelEdit: () => set({ edit: null }),
        onEdit: (e) => set({ edit: { id: u.id, name: e.target.value } }),
        saveEdit: () => { const nm = (se.edit.name || '').trim(); if (!nm) { set({ err: 'Enter a name.' }); return; } run(() => orgW.renameUnit(this, this.srv, u.id, nm), { edit: null }, 'Renamed to ' + nm + '.'); },
        confirming: se.del === u.id, askRemove: () => set({ del: u.id, err: '', ok: '' }), keep: () => set({ del: null }),
        remove: () => run(() => orgW.removeUnit(this, this.srv, u.id), { del: null }, u.name + ' was removed.'),
        // moving: only to a level this kind may sit under
        canMove: PARENTS[K.kind].length > 0, moveOpts, moveTo, moved,
        onMove: (e) => { const v = e.target.value; set((cur) => ({ move: { ...cur.move, [u.id]: v }, err: '', ok: '' })); },
        doMove: () => {
          const dest = moveTo ? byId.get(moveTo) : null;
          run(() => orgW.moveUnit(this, this.srv, u.id, moveTo || null), { move: { ...se.move, [u.id]: undefined } }, u.name + ' now sits under ' + (dest ? kindLabel(dest.kind) + ' ' + dest.name : 'the top level') + '.');
        },
        // merging: tick the ones to merge
        picked: sel.includes(u.id),
        onPick: (e) => { const on = e.target.checked; set((cur) => { const now = (cur.sel[K.kind] || []).filter((x) => x !== u.id).concat(on ? [u.id] : []); return { sel: { ...cur.sel, [K.kind]: now }, mergeAsk: null, err: '', ok: '' }; }); }
      };
    });
    const into = sel.includes(se.into[K.kind]) ? se.into[K.kind] : (sel[0] || '');
    const target = byId.get(into);
    const others = sel.filter((id) => id !== into).map((id) => byId.get(id)).filter(Boolean);
    return {
      ...K, rows, count: rows.length, none: rows.length === 0,
      canMerge: sel.length >= 2, picked: sel.length,
      mergeOpts: sel.map((id) => ({ id, label: (byId.get(id) || {}).name })), mergeInto: into,
      onInto: (e) => set((cur) => ({ into: { ...cur.into, [K.kind]: e.target.value }, mergeAsk: null })),
      asking: se.mergeAsk === K.kind,
      mergeText: target ? 'Merge ' + others.map((x) => x.name).join(', ') + ' into ' + target.name + '? Their people, parts, tickets and walls move to ' + target.name + ', and they are removed. This cannot be undone.' : '',
      askMerge: () => set({ mergeAsk: K.kind, err: '', ok: '' }), cancelMerge: () => set({ mergeAsk: null }),
      clearPick: () => set((cur) => ({ sel: { ...cur.sel, [K.kind]: [] }, mergeAsk: null })),
      merge: () => run(() => orgW.mergeUnits(this, this.srv, into, others.map((x) => x.id)), { sel: { ...se.sel, [K.kind]: [] }, mergeAsk: null }, others.map((x) => x.name).join(', ') + ' merged into ' + (target ? target.name : '') + '.')
    };
  });
  const K = KINDS.find((x) => x.kind === se.kind) || KINDS[0];
  const name = se.name.trim();
  return {
    canEdit, open: true, openIt, close: () => this.setState({ section: 'people', structEd: null }), busy: se.busy,
    err: se.err, hasErr: !!se.err, ok: se.ok, hasOk: !!se.ok && !se.err,
    levels,
    summary: levels.filter((l) => l.count).map((l) => l.count + ' ' + (l.count === 1 ? l.label.toLowerCase() : l.plural.toLowerCase())).join(' · ') || 'No structure yet. Everyone works at one level.',
    kinds: KINDS.map((x) => ({ kind: x.kind, label: x.label })), kind: se.kind, onKind: (e) => set({ kind: e.target.value, parent: '', err: '' }),
    kindBlurb: K.blurb, name: se.name, onName: (e) => set({ name: e.target.value, err: '' }),
    parentOpts, hasParents: PARENTS[se.kind].length > 0, parent, onParent: (e) => set({ parent: e.target.value }),
    canAdd: !!name && !se.busy,
    add: () => { if (!name) { set({ err: 'Enter a name.' }); return; } run(() => orgW.addUnit(this, this.srv, { kind: se.kind, name, parent }), { name: '' }, K.label + ' ' + name + ' was added.'); },
    note: 'Add, rename, move and merge the parts of your organization. Moving or merging re-places everyone inside straight away. A level only appears in Hub once it has at least one entry.'
  };
}
