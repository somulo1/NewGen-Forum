import { hasLevelsFor } from '~/logic/actions/onboarding.js';
import * as orgW from '~/logic/server/widgets/org.js';
import { GENDERS, MANDATORY, csvColumns, structureFields, structureOptions } from '~/logic/onboarding.js';

// renderVals step: onboarding. The form only shows the structure levels this organization has set up.
export function buildOnboarding(ctx) {
  const { can, modEnabled, modInfo, reportRoleName, ridOf, rightOf, rightsDef, roleHolders, s, targetRoleOf, visibleRoles } = ctx;
  const has = hasLevelsFor(s, this.srv ? this.org : null);
  // the levels this organization switched on (Organization structure > levels) always show on the person form, under the
  // organization's own names, even before any place of that level exists
  const lvCfg = (this.srv && this.srv.levels && this.srv.levels.levels) || [];
  const LVK = { sector: 'sectors', branch: 'branches', unit: 'units', department: 'departments' };
  lvCfg.forEach((l) => { if (l.enabled && LVK[l.kind]) has[LVK[l.kind]] = true; });
  const lvLabel = (key) => { const l = lvCfg.find((x) => x.kind === key); return l && l.label ? l.label : null; };
  const ob = s.onboard;
  const canOnboard = can('people', 'Onboard people');
  const base = { canOnboard, open: false, openOne: () => this.openOnboard('one'), openBulk: () => this.openOnboard('bulk') };
  const structureEd = buildStructureEd.call(this, s, can);
  if (!ob) return { onb: base, structureEd };
  const f = ob.f, errs = ob.errs;
  // Role and individual rights, only when editing someone who is already in the directory.
  let access = null;
  if (ob.edit !== undefined && ob.edit !== null && this.personById(ob.edit)) {
    const person = this.personById(ob.edit);
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
  // with the server's places (and their parents): each list only has the places under what is chosen above it, in whatever
  // order this organization nests its levels
  if (this.srv && (this.srv.unitList || []).length) {
    const U = this.srv.unitList, byId = new Map(U.map((u) => [u.id, u]));
    const ancestors = (u) => { const out = []; let p = u && byId.get(u.parent_id); while (p && out.length < 10) { out.push(p); p = byId.get(p.parent_id); } return out; };
    ['sector', 'branch', 'unit', 'department'].forEach((k) => {
      opts[k] = U.filter((u) => u.kind === k && ['sector', 'branch', 'unit', 'department'].every((j) => { if (j === k || !f[j]) return true; const a = ancestors(u).find((x) => x.kind === j); return !a || a.name === f[j]; }))
        .map((u) => u.name).sort((a, b) => a.localeCompare(b));
    });
  }
  const structure = structureFields(has).map((x) => ({ ...x, label: lvLabel(x.key) || x.label, required: x.required && (opts[x.key] || []).length > 0 }));
  const rowsAll = ob.bulk ? ob.bulk.rows : [];
  const good = rowsAll.filter((r) => r.ok).length;
  const bad = rowsAll.length - good;
  const isEdit = ob.edit !== undefined && ob.edit !== null;
  const onb = { ...base, open: true, isEdit, heading: isEdit ? 'Edit ' + this.personById(ob.edit).name : 'Onboard people', submitLabel: isEdit ? 'Save changes' : 'Onboard person', isOne: ob.tab === 'one', isBulk: ob.tab === 'bulk' && !isEdit, showTabs: !isEdit, close: () => this.closeOnboard(),
    goOne: () => this.setOnboardTab('one'), goBulk: () => this.setOnboardTab('bulk'),
    done: ob.done, hasDone: !!ob.done, doneOne: !!(ob.done && ob.done.one), doneBulk: !!(ob.done && ob.done.count !== undefined),
    fields: MANDATORY.map((m) => ({ key: m.key, label: m.label, type: m.type, isSelect: m.type === 'select', value: f[m.key], err: errs[m.key] || '', hasErr: !!errs[m.key], options: GENDERS,
      onInput: (e) => this.setOnboardField(m.key, e.target.value) })),
    roleOptions: visibleRoles.map((r) => ({ v: r.id, l: r.name })), role: f.role, roleErr: errs.role || '', hasRoleErr: !!errs.role, onRole: (e) => this.setOnboardField('role', e.target.value),
    title: f.title, onTitle: (e) => this.setOnboardField('title', e.target.value),
    structure: structure.map((x) => ({ key: x.key, label: x.label + (x.required ? '' : ' (optional)'), value: f[x.key], options: isEdit && ob.orig[x.key] && opts[x.key].indexOf(ob.orig[x.key]) === -1 && f[x.key] === ob.orig[x.key] ? opts[x.key].concat([ob.orig[x.key]]) : opts[x.key], err: errs[x.key] || '', hasErr: !!errs[x.key],
      onChange: (e) => this.setOnboardField(x.key, e.target.value) })),
    hasStructure: structure.length > 0, access, hasAccess: !!access,
    structureNote: structure.length ? 'These come from your organization: ' + structure.map((x) => x.label.toLowerCase()).join(', ') + '. Each choice narrows the next one.' : 'This organization has no sectors, branches, units or departments set up, so only the required details are needed.',
    submit: () => this.submitOnboard(), saving: !!ob.saving, formError: ob.formError || '',
    bulk: { role: ob.bulkRole || '', onRole: (e) => this.setOnboardBulkRole(e.target.value), roleErr: ob.bulkRoleErr || '', hasRoleErr: !!ob.bulkRoleErr, text: ob.bulkText, onText: (e) => this.parseBulk(e.target.value), columns: csvColumns(has).join(', '), template: () => this.downloadOnboardTemplate(),
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

const FRESH_ED = { kind: 'sector', name: '', parent: '', err: '', ok: '', busy: false, edit: null, del: null, move: {}, sel: {}, into: {}, mergeAsk: null, tab: 'sector', q: {} };

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
  const set = (patch) => this.setState((st) => ({ structEd: { ...FRESH_ED, ...(st.structEd || {}), ...(typeof patch === 'function' ? patch({ ...FRESH_ED, ...(st.structEd || {}) }) : patch) } }));
  // one change at a time; the result (or the server's reason) is shown at the top of the page
  const run = (fn, after, ok) => {
    if (se.busy) return Promise.resolve();
    set({ busy: true, err: '', ok: '' });
    return fn().then(() => set({ busy: false, ok: ok || '', ...(after || {}) })).catch((e) => set({ busy: false, err: (e && e.message) || 'That did not work.' }));
  };
  // The organization's own levels: only the ones it switched on can be made here, each under the one it chose, with its names.
  const cfg = (this.srv.levels && this.srv.levels.levels) || [];
  const cfgOf = (k) => cfg.find((l) => l.kind === k) || { kind: k, enabled: false, parent: '' };
  const kindLabel = (k) => cfgOf(k).label || (KINDS.find((x) => x.kind === k) || {}).label || k;
  const KS = KINDS.filter((K) => cfgOf(K.kind).enabled).map((K) => ({ ...K, label: cfgOf(K.kind).label || K.label, plural: cfgOf(K.kind).plural || K.plural }));
  const have = (k) => units.filter((u) => u.kind === k);
  const parentKind = (kind) => cfgOf(kind).parent || '';
  const parentChoices = (kind, self) => (parentKind(kind) ? have(parentKind(kind)).filter((u) => u.id !== self).map((u) => ({ id: u.id, label: kindLabel(parentKind(kind)) + ' · ' + u.name })) : []);
  const curKind = KS.some((K) => K.kind === se.kind) ? se.kind : ((KS[0] || {}).kind || '');
  const parentOpts = parentChoices(curKind);
  const parent = parentOpts.some((o) => o.id === se.parent) ? se.parent : '';
  // the roles that can head a unit (the role is what gives its holders their rights)
  // every role can head a place (the Owner role too)
  const roleChoices = (s.serverRoles || []).filter((r) => r.id).map((r) => ({ key: r.id, name: r.name }));
  const levels = KS.map((K) => {
    const list = have(K.kind).sort((a, b) => a.name.localeCompare(b.name));
    const sel = (se.sel[K.kind] || []).filter((id) => list.some((u) => u.id === id));
    const rows = list.map((u) => {
      const par = byId.get(u.parent_id);
      const n = this.peopleDir().filter((p) => p[K.field] === u.name).length;
      const editing = se.edit && se.edit.id === u.id;
      const moveOpts = parentChoices(K.kind, u.id);
      const moveTo = se.move[u.id] !== undefined ? se.move[u.id] : (u.parent_id || '');
      const moved = moveTo !== (u.parent_id || '');
      return {
        // the head role is decided for the whole level in "Levels your organization uses"
        headRole: u.head_role || '', headRoles: roleChoices, headRoleName: ((roleChoices.find((r) => r.key === (cfgOf(u.kind).head_role || u.head_role)) || {}).name) || '', heads: (u.heads || []).map((h) => h.name),
        // who leads it: people holding the level's head role, or the place's own members; choosing one also makes them a member
        ...(() => {
          const rk = cfgOf(u.kind).head_role || u.head_role || '';
          const rn = ((roleChoices.find((r) => r.key === rk) || {}).name) || '';
          const fieldOf = { sector: 'sector', branch: 'loc', unit: 'unit', department: 'dept' }[u.kind];
          const cands = (this.dir || []).filter((p) => p.uid && (p.accStatus || 'active') === 'active' && ((rk && (p.role === rk || p.role === rn)) || p[fieldOf] === u.name));
          const me = this.srv && this.srv.me;
          const mine = me ? [{ uid: me.person, name: me.name }] : [];
          const all = cands.map((p) => ({ uid: p.uid, name: p.name })).concat(mine.filter((m) => !cands.some((c) => c.uid === m.uid) && u.head_person === m.uid));
          return { headPerson: u.head_person || '', headOpts: all.sort((a, b) => a.name.localeCompare(b.name)).slice(0, 200),
            onHeadPerson: (e) => { const v = e.target.value || null; const who = (all.find((x) => x.uid === v) || {}).name; run(() => orgW.setHeadPerson(this, this.srv, u.id, v), null, v ? who + ' now leads ' + u.name + ' and is a member of it.' : 'Whoever holds the head role leads ' + u.name + ' again.'); } };
        })(),
        onHead: (e) => { const role = e.target.value; run(() => orgW.setHead(this, this.srv, u.id, role), null, role ? 'The ' + ((roleChoices.find((r) => r.key === role) || {}).name || role) + ' role now heads ' + u.name + '.' : u.name + ' no longer has a head role.'); },
        id: u.id, name: u.name, under: par ? kindLabel(par.kind) + ' ' + par.name : 'Top level', people: n + (n === 1 ? ' person' : ' people'), editing, editName: editing ? se.edit.name : u.name,
        startEdit: () => set({ edit: { id: u.id, name: u.name }, err: '', ok: '' }), cancelEdit: () => set({ edit: null }),
        onEdit: (e) => set({ edit: { id: u.id, name: e.target.value } }),
        saveEdit: () => { const nm = (se.edit.name || '').trim(); if (!nm) { set({ err: 'Enter a name.' }); return; } run(() => orgW.renameUnit(this, this.srv, u.id, nm), { edit: null }, 'Renamed to ' + nm + '.'); },
        confirming: se.del === u.id, askRemove: () => set({ del: u.id, err: '', ok: '' }), keep: () => set({ del: null }),
        remove: () => run(() => orgW.removeUnit(this, this.srv, u.id), { del: null }, u.name + ' was removed.'),
        // moving: only to a level this kind may sit under
        canMove: !!parentKind(K.kind), moveOpts, moveTo, moved,
        // adding people: tick them from the directory; they are placed here and in every part above it
        // its own dashboard, to view it: what can be changed there still follows this person's rights at that dashboard
        openDash: () => this.setState({ section: 'orgunit', ou: ({ sector: 'sectors', branch: 'branches', unit: 'units', department: 'departments' })[u.kind] + ':' + u.id, ouTab: 'comms', ouItem: null, ticketOpen: null, projectOpen: null }),
        adding: se.addTo === u.id, openAdd: () => set({ addTo: u.id, addSel: {}, addQ: '', err: '', ok: '' }), closeAdd: () => set({ addTo: null }),
        addQ: se.addTo === u.id ? se.addQ || '' : '', onAddQ: (e) => { const v = e.target.value; set({ addQ: v }); },
        addList: se.addTo === u.id ? this.peopleDir().filter((p) => p.uid && p[K.field] !== u.name && (!(se.addQ || '').trim() || p.name.toLowerCase().includes(se.addQ.trim().toLowerCase()))).slice(0, 60)
          .map((p) => ({ uid: p.uid, name: p.name, sub: [p.title, p.dept, p.loc].filter(Boolean).join(' · '), on: !!(se.addSel || {})[p.uid], toggle: () => set((cur) => ({ addSel: { ...(cur.addSel || {}), [p.uid]: !(cur.addSel || {})[p.uid] } })) })) : [],
        addCount: Object.values(se.addSel || {}).filter(Boolean).length,
        doAdd: () => {
          const ids = Object.keys(se.addSel || {}).filter((k) => se.addSel[k]);
          if (!ids.length) { set({ err: 'Tick the people to add.' }); return; }
          run(() => orgW.addPeople(this, this.srv, u.id, ids).then((failed) => { if (failed.length) throw new Error(failed.length + ' could not be added: ' + failed[0]); }), { addTo: null, addSel: {} }, ids.length + (ids.length === 1 ? ' person was' : ' people were') + ' added to ' + u.name + '.');
        },
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
    const q = ((se.q || {})[K.kind] || '').trim().toLowerCase();
    const shown = q ? rows.filter((r) => (r.name + ' ' + r.under).toLowerCase().includes(q)) : rows;
    return {
      ...K, rows: shown, count: rows.length, none: rows.length === 0, noMatch: rows.length > 0 && shown.length === 0, q: (se.q || {})[K.kind] || '',
      onQ: (e) => { const v = e.target.value; set((cur) => ({ q: { ...cur.q, [K.kind]: v } })); },
      selectShown: () => set((cur) => ({ sel: { ...cur.sel, [K.kind]: shown.map((r) => r.id) }, mergeAsk: null })),
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
  const K = KS.find((x) => x.kind === curKind) || KS[0] || KINDS[0];
  const name = se.name.trim();
  const tab = KS.some((x) => x.kind === se.tab) ? se.tab : curKind;
  const needParent = !!parentKind(curKind);
  // the levels panel: which levels the organization uses, what it calls them and which sits under which
  const draft = se.lvDraft || KINDS.map((x) => { const c = cfgOf(x.kind); return { kind: x.kind, enabled: !!c.enabled, label: c.label || x.label, plural: c.plural || x.plural, parent: c.parent || '', head_role: c.head_role || '' }; });
  const setDraft = (kind, patch) => set((cur) => ({ lvDraft: (cur.lvDraft || draft).map((l) => (l.kind === kind ? { ...l, ...patch } : l)), err: '', ok: '' }));
  const levelsPanel = {
    open: !!se.lvOpen || !cfg.some((l) => l.enabled), toggle: () => set({ lvOpen: !se.lvOpen, lvDraft: null }),
    rows: draft.map((l) => ({ ...l, onEnabled: (e) => setDraft(l.kind, { enabled: e.target.checked }), onLabel: (e) => setDraft(l.kind, { label: e.target.value }), onPlural: (e) => setDraft(l.kind, { plural: e.target.value }),
      onParent: (e) => setDraft(l.kind, { parent: e.target.value }), onHeadRole: (e) => setDraft(l.kind, { head_role: e.target.value }), headRoleOpts: roleChoices,
      parentOpts: [{ v: '', l: 'The organization itself' }].concat(draft.filter((x) => x.kind !== l.kind && x.enabled).map((x) => ({ v: x.kind, l: x.label }))) })),
    dirty: !!se.lvDraft,
    save: () => run(() => orgW.saveLevels(this, this.srv, draft), { lvDraft: null, lvOpen: false }, 'Your organization\'s levels were saved.')
  };
  return {
    canEdit, open: true, openIt, close: () => this.setState({ section: 'people', structEd: null }), busy: se.busy,
    err: se.err, hasErr: !!se.err, ok: se.ok, hasOk: !!se.ok && !se.err, levelsPanel, noLevels: KS.length === 0,
    levels, tabs: levels.map((l) => ({ kind: l.kind, label: l.plural, count: l.count, on: l.kind === tab, pick: () => set({ tab: l.kind, mergeAsk: null, err: '', ok: '' }) })), cur: levels.find((l) => l.kind === tab) || levels[0],
    summary: levels.filter((l) => l.count).map((l) => l.count + ' ' + (l.count === 1 ? l.label.toLowerCase() : l.plural.toLowerCase())).join(' · ') || (KS.length ? 'Nothing added yet.' : 'Choose the levels your organization uses to start.'),
    kinds: KS.map((x) => ({ kind: x.kind, label: x.label })), kind: curKind, onKind: (e) => set({ kind: e.target.value, parent: '', err: '' }),
    kindBlurb: needParent ? 'Each ' + K.label.toLowerCase() + ' sits under a ' + kindLabel(parentKind(curKind)).toLowerCase() + '.' : 'Each ' + K.label.toLowerCase() + ' sits directly under the organization.', name: se.name, onName: (e) => set({ name: e.target.value, err: '' }),
    parentOpts, hasParents: needParent, parentLabel: needParent ? kindLabel(parentKind(curKind)) : '', parent, onParent: (e) => set({ parent: e.target.value }),
    canAdd: !!name && !se.busy && (!needParent || !!parent) && KS.length > 0,
    add: () => { if (!name) { set({ err: 'Enter a name.' }); return; } if (needParent && !parent) { set({ err: 'Choose the ' + kindLabel(parentKind(curKind)).toLowerCase() + ' it sits under.' }); return; } run(() => orgW.addUnit(this, this.srv, { kind: curKind, name, parent }), { name: '' }, K.label + ' ' + name + ' was added.'); },
    note: 'Add, rename, move and merge the parts of your organization. Moving or merging re-places everyone inside straight away. A level only appears in Hub once it has at least one entry.'
  };
}
