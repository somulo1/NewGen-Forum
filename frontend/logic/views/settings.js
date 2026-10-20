import { ADDONS, PLANS } from '~/logic/state/settings.js';
import { buildXlsx, downloadBlob } from '~/logic/xlsx.js';
import { buildDoc } from '~/logic/etr.js';
import * as settingsW from '~/logic/server/widgets/settings.js';

// renderVals step: settings. Organization profile, subscription, billing, security and data. Kept to what an org needs.
export function buildSettings(ctx) {
  const { can, s } = ctx;
  const canEdit = can('settings', 'Edit organization profile');
  const canSub = can('settings', 'Manage subscription');
  const draft = s.setDraft || {};
  const money = (n) => 'KES ' + Math.round(n).toLocaleString('en-US');
  const dayLabel = (d) => (d ? new Date(d + 'T00:00:00').toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) : '—');
  const audit = (action) => this.srv ? null : this.setState((st) => ({ extraLogs: [{ id: 'x' + Date.now() + Math.random().toString(36).slice(2, 5), ago: 0, user: 'Samwel Okoth Omulo', action, module: 'Settings', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }].concat(st.extraLogs) }));
  // Connected: the server stores the change (and writes its own audit entry); the screen then shows what it kept.
  const srvDo = (fn, okText) => fn().then(() => { if (okText) this.notify(okText, '#B8F36B'); }).catch((e) => this.notify((e && e.message) || 'That did not save.', '#FF9B8A'));
  const setDraft = (part, patch) => this.setState((st) => ({ setDraft: { ...(st.setDraft || {}), [part]: { ...((st.setDraft || {})[part] || {}), ...patch }, errs: { ...((st.setDraft || {}).errs || {}), ...Object.keys(patch).reduce((m, k) => { m[k] = ''; return m; }, {}) } } }));
  const clearPart = (part) => this.setState((st) => { const d = { ...(st.setDraft || {}) }; delete d[part]; return { setDraft: Object.keys(d).filter((k) => k !== 'errs').length ? d : null }; });
  const errs = draft.errs || {};

  // ---- generic form field builder ----
  const field = (part, saved, key, label, type, extra) => {
    const v = draft[part] && draft[part][key] !== undefined ? draft[part][key] : saved[key];
    return { key, label, type, value: v, isSelect: type === 'select', options: (extra && extra.options) || [], optLabel: (extra && extra.optLabel) || ((o) => o), placeholder: (extra && extra.placeholder) || '', wide: !!(extra && extra.wide), required: !!(extra && extra.required), err: errs[key] || '', hasErr: !!errs[key],
      disabled: !canEdit, onInput: (e) => setDraft(part, { [key]: e.target.value }) };
  };
  const dirtyOf = (part, saved) => !!draft[part] && Object.keys(draft[part]).some((k) => String(draft[part][k]) !== String(saved[k]));

  // ---- Profile ----
  const P = s.orgProfile;
  const pf = (k, l, t, e) => field('profile', P, k, l, t, e);
  const profileGroups = [
    { title: 'Organization', fields: [pf('name', 'Organization name', 'text', { required: true }), pf('short', 'Short name', 'text', { placeholder: 'e.g. ACME' }), pf('industry', 'Industry', 'select', { options: ['Financial services', 'Education', 'Health', 'Retail', 'Manufacturing', 'Government', 'Non-profit', 'Technology', 'Other'] }), pf('size', 'Number of employees', 'select', { options: ['1–50', '51–200', '201–500', '501–1,000', '1,000+'] }), pf('regNo', 'Registration number', 'text'), pf('taxId', 'Tax ID (PIN)', 'text')] },
    { title: 'Contact', fields: [pf('email', 'Official email', 'email'), pf('phone', 'Phone', 'tel'), pf('website', 'Website', 'url', { placeholder: 'https://', wide: true })] },
    { title: 'Address', fields: [pf('address', 'Street address', 'text', { wide: true }), pf('city', 'City', 'text'), pf('country', 'Country', 'select', { options: ['Kenya', 'Uganda', 'Tanzania', 'Rwanda', 'Ethiopia', 'Nigeria', 'Ghana', 'South Africa', 'Other'] })] },
    { title: 'Regional', fields: [pf('timezone', 'Time zone', 'select', { options: ['Africa/Nairobi', 'Africa/Kampala', 'Africa/Dar_es_Salaam', 'Africa/Lagos', 'Africa/Johannesburg', 'UTC'] }), pf('language', 'Language', 'select', { options: ['English', 'Kiswahili', 'French'] }), pf('currency', 'Currency', 'select', { options: ['KES', 'UGX', 'TZS', 'NGN', 'GHS', 'ZAR', 'USD'] })] }
  ];
  const profileDirty = dirtyOf('profile', P);
  const profile = { groups: profileGroups, dirty: profileDirty, canEdit,
    logo: { url: (draft.profile && draft.profile.logo !== undefined ? draft.profile.logo : P.logo), initials: (P.short || P.name || 'O').slice(0, 2).toUpperCase(),
      onFile: (e) => {
        const f = e.target.files && e.target.files[0]; e.target.value = '';
        if (!f) return;
        if (['image/png', 'image/jpeg', 'image/webp', 'image/svg+xml'].indexOf(f.type) === -1) { this.notify('The logo must be a PNG, JPEG, WebP or SVG image.', '#FFB86B'); return; }
        if (f.size > 1048576) { this.notify('The logo can be up to 1 MB. This one is ' + (f.size / 1048576).toFixed(1) + ' MB.', '#FFB86B'); return; }
        const r = new FileReader();
        r.onload = () => setDraft('profile', { logo: String(r.result) });
        r.readAsDataURL(f);
      },
      remove: () => setDraft('profile', { logo: '' }) },
    discard: () => clearPart('profile'),
    save: () => {
      const v = { ...P, ...(draft.profile || {}) }, e = {};
      if (!String(v.name || '').trim()) e.name = 'Organization name is required.';
      if (v.email && !/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(v.email)) e.email = 'That email does not look right.';
      if (v.website && !/^https?:\/\/\S+\.\S+/.test(v.website)) e.website = 'Start with http:// or https://';
      if (v.phone && !/^\+?[\d\s()-]{7,20}$/.test(v.phone)) e.phone = 'Use 7 to 20 digits, with an optional + at the start.';
      if (Object.keys(e).length) { this.setState((st) => ({ setDraft: { ...(st.setDraft || {}), errs: e } })); return; }
      if (this.srv) { const body = { ...P, ...(draft.profile || {}), name: v.name.trim() }; srvDo(() => settingsW.saveProfile(this, this.srv, body), 'Organization profile saved.').then(() => clearPart('profile')); return; }
      this.setState((st) => ({ orgProfile: { ...st.orgProfile, ...(st.setDraft || {}).profile, name: v.name.trim() } }));
      clearPart('profile'); audit('Updated the organization profile'); this.notify('Organization profile saved.', '#B8F36B');
    } };

  // ---- Subscription ----
  const S = s.orgSub;
  const plan = PLANS.find((p) => p.id === S.plan) || PLANS[1];
  const perSeat = (p, cycle) => (cycle === 'annual' ? p.price * 0.8 : p.price);
  const seatsCost = (p, seats, cycle) => seats * perSeat(p, cycle) * 100;
  const addonsOn = ADDONS.filter((a) => S.addons[a.id]);
  const total = (p, seats, cycle, ad) => seatsCost(p, seats, cycle) + ad.reduce((n, a) => n + a.price * 10, 0);
  const people = this.dir.length;
  const ask = (c) => this.setState({ setConfirm: c });
  // every subscription change sends the whole subscription; the server checks seats against the plan and the people in
  const subSave = (patch, text) => { if (!this.srv) return false; srvDo(() => settingsW.saveSubscription(this, this.srv, { ...S, ...patch }), text); return true; };
  const done = (text) => { this.setState({ setConfirm: null }); audit(text); this.notify(text + '.', '#B8F36B'); };
  const cf = s.setConfirm;
  const sub = { canSub, statusLabel: S.status === 'Cancelling' ? 'Cancels on ' + dayLabel(S.cancelAt) : S.status, statusColor: S.status === 'Cancelling' ? '#FFB86B' : '#B8F36B', cancelling: S.status === 'Cancelling',
    planName: plan.name, planBlurb: plan.blurb, cycle: S.cycle, cycleLabel: S.cycle === 'annual' ? 'Billed yearly' : 'Billed monthly', renews: dayLabel(S.renews), autoRenew: S.autoRenew,
    seats: S.seats, seatLimit: plan.seats, peopleNow: people, seatPct: Math.min(100, Math.round(S.seats / plan.seats * 100)), seatNote: people > S.seats ? people + ' people are in Hub but only ' + S.seats + ' seats are paid for.' : people + ' of ' + S.seats + ' seats in use.', seatWarn: people > S.seats,
    total: money(total(plan, S.seats, S.cycle, addonsOn)) + ' / month', totalNote: S.seats + ' seats × ' + money(perSeat(plan, S.cycle) * 100) + (addonsOn.length ? ' + ' + addonsOn.length + ' add-on' + (addonsOn.length > 1 ? 's' : '') : ''),
    setSeats: (d) => { const v = Math.max(1, Math.min(plan.seats, S.seats + d)); if (v !== S.seats && subSave({ seats: v })) return; if (v !== S.seats) { this.setState((st) => ({ orgSub: { ...st.orgSub, seats: v } })); audit('Changed paid seats to ' + v); } },
    toggleCycle: () => { if (subSave({ cycle: S.cycle === 'annual' ? 'monthly' : 'annual' })) return; this.setState((st) => ({ orgSub: { ...st.orgSub, cycle: st.orgSub.cycle === 'annual' ? 'monthly' : 'annual' } })); audit('Switched billing to ' + (S.cycle === 'annual' ? 'monthly' : 'yearly')); },
    toggleAuto: () => { if (subSave({ autoRenew: !S.autoRenew })) return; this.setState((st) => ({ orgSub: { ...st.orgSub, autoRenew: !st.orgSub.autoRenew } })); audit('Turned auto-renew ' + (S.autoRenew ? 'off' : 'on')); },
    plans: PLANS.map((p) => ({ id: p.id, name: p.name, blurb: p.blurb, features: p.features, price: money(perSeat(p, S.cycle) * 100) + ' / seat / month', current: p.id === S.plan, canChoose: canSub && p.id !== S.plan, action: PLANS.indexOf(p) > PLANS.indexOf(plan) ? 'Upgrade' : 'Switch',
      choose: () => ask({ kind: 'plan', id: p.id, text: (PLANS.indexOf(p) > PLANS.indexOf(plan) ? 'Upgrade' : 'Switch') + ' to ' + p.name + '? New total: ' + money(total(p, Math.min(S.seats, p.seats), S.cycle, addonsOn)) + ' / month.', confirmLabel: 'Confirm ' + p.name }) })),
    addons: ADDONS.map((a) => ({ ...a, priceLabel: money(a.price * 10) + ' / month', on: !!S.addons[a.id], toggle: () => { if (subSave({ addons: { ...S.addons, [a.id]: !S.addons[a.id] } })) return; this.setState((st) => ({ orgSub: { ...st.orgSub, addons: { ...st.orgSub.addons, [a.id]: !st.orgSub.addons[a.id] } } })); audit((S.addons[a.id] ? 'Removed' : 'Added') + ' add-on: ' + a.name); } })),
    cancel: () => ask({ kind: 'cancel', text: 'Cancel the subscription? Everyone keeps access until ' + dayLabel(S.renews) + ', then the organization moves to read-only.', confirmLabel: 'Cancel subscription' }),
    resume: () => { if (subSave({ status: 'Active', cancelAt: '' }, 'Subscription resumed.')) return; this.setState((st) => ({ orgSub: { ...st.orgSub, status: 'Active', cancelAt: '' } })); audit('Resumed the subscription'); this.notify('Subscription resumed.', '#B8F36B'); } };

  // ---- Billing ----
  const B = s.orgBilling;
  const bf = (k, l, t, e) => field('billing', B, k, l, t, e);
  const card = draft.card || { number: '', exp: '', name: '' };
  const billing = { canSub, cardLabel: B.cardBrand + ' ending ' + B.cardLast4, cardExp: 'Expires ' + B.cardExp, holder: B.holder,
    contact: [bf('contactEmail', 'Billing email (invoices go here)', 'email', { wide: true }), bf('holder', 'Name on invoices', 'text'), bf('vat', 'VAT / tax number', 'text')],
    dirty: dirtyOf('billing', B), discard: () => clearPart('billing'),
    save: () => { const v = { ...B, ...(draft.billing || {}) }; if (v.contactEmail && !/^[^\s@]+@[^\s@]+\.[^\s@]{2,}$/.test(v.contactEmail)) { this.setState((st) => ({ setDraft: { ...(st.setDraft || {}), errs: { contactEmail: 'That email does not look right.' } } })); return; } if (this.srv) { srvDo(() => settingsW.saveBilling(this, this.srv, { ...B, ...(draft.billing || {}) }), 'Billing details saved.').then(() => clearPart('billing')); return; } this.setState((st) => ({ orgBilling: { ...st.orgBilling, ...(st.setDraft || {}).billing } })); clearPart('billing'); audit('Updated billing details'); this.notify('Billing details saved.', '#B8F36B'); },
    cardOpen: !!draft.cardOpen, toggleCard: () => this.setState((st) => ({ setDraft: { ...(st.setDraft || {}), cardOpen: !(st.setDraft || {}).cardOpen, card: { number: '', exp: '', name: '' }, cardErr: '' } })),
    card: { number: card.number, exp: card.exp, name: card.name, err: draft.cardErr || '', hasErr: !!draft.cardErr,
      onNumber: (e) => this.setState((st) => ({ setDraft: { ...st.setDraft, card: { ...st.setDraft.card, number: e.target.value.replace(/[^\d ]/g, '').slice(0, 19) }, cardErr: '' } })),
      onExp: (e) => this.setState((st) => ({ setDraft: { ...st.setDraft, card: { ...st.setDraft.card, exp: e.target.value.replace(/[^\d/]/g, '').slice(0, 5) }, cardErr: '' } })),
      onName: (e) => this.setState((st) => ({ setDraft: { ...st.setDraft, card: { ...st.setDraft.card, name: e.target.value }, cardErr: '' } })),
      save: () => {
        const c = (this.state.setDraft || {}).card || card, digits = c.number.replace(/\D/g, '');
        if (digits.length < 13 || digits.length > 19) { this.setState((st) => ({ setDraft: { ...st.setDraft, cardErr: 'Enter the full card number.' } })); return; }
        if (!/^(0[1-9]|1[0-2])\/\d{2}$/.test(c.exp)) { this.setState((st) => ({ setDraft: { ...st.setDraft, cardErr: 'Expiry must look like 08/28.' } })); return; }
        const brand = digits[0] === '4' ? 'Visa' : digits[0] === '5' ? 'Mastercard' : 'Card';
        if (this.srv) {
          // only the brand, the last four digits and the expiry are sent and kept; the full number never leaves this form
          srvDo(() => settingsW.saveBilling(this, this.srv, { ...B, cardBrand: brand, cardLast4: digits.slice(-4), cardExp: c.exp }), 'Payment card updated.')
            .then(() => this.setState((st) => { const d = { ...(st.setDraft || {}) }; delete d.cardOpen; delete d.card; delete d.cardErr; return { setDraft: Object.keys(d).filter((k) => k !== 'errs').length ? d : null }; }));
          return;
        }
        this.setState((st) => ({ orgBilling: { ...st.orgBilling, cardBrand: brand, cardLast4: digits.slice(-4), cardExp: c.exp }, setDraft: (() => { const d = { ...(st.setDraft || {}) }; delete d.cardOpen; delete d.card; delete d.cardErr; return Object.keys(d).filter((k) => k !== 'errs').length ? d : null; })() }));
        audit('Changed the payment card to ' + brand + ' ending ' + digits.slice(-4)); this.notify('Payment card updated.', '#B8F36B');
      } },
    invoices: s.invoices.map((i) => ({ ...i, dateLabel: dayLabel(i.date), amountLabel: money(i.amount), color: i.status === 'Paid' ? '#B8F36B' : '#FFB86B', isPaid: i.status === 'Paid',
      openInvoice: () => this.setState({ etrOpen: { no: i.no, kind: 'invoice' } }), openReceipt: () => this.setState({ etrOpen: { no: i.no, kind: 'receipt' } }),
      download: () => { const rows = [['Invoice', 'Date', 'Plan', 'Amount', 'Status'], [i.no, i.date, plan.name, i.amount, i.status]]; downloadBlob(buildXlsx('Invoice', rows), i.no + '.xlsx'); } })),
    downloadAll: () => { downloadBlob(buildXlsx('Invoices', [['Invoice', 'Date', 'Amount (KES)', 'Status']].concat(s.invoices.map((i) => [i.no, i.date, i.amount, i.status]))), 'invoices.xlsx'); audit('Downloaded all invoices'); } };

  // ---- Security & data ----
  const X = s.orgSecurity;
  const xf = (k, l, t, e) => field('security', X, k, l, t, e);
  const flag = (k, label, note) => ({ key: k, label, note, on: (draft.security && draft.security[k] !== undefined ? draft.security[k] : X[k]), disabled: !canEdit, toggle: () => setDraft('security', { [k]: !(draft.security && draft.security[k] !== undefined ? draft.security[k] : X[k]) }) });
  const security = { canEdit, flags: [flag('mfa', 'Require two-step sign-in (MFA)', 'Everyone must confirm sign-in with a second step.'), flag('externalSharing', 'Allow sharing posts outside Hub', 'Turn off to stop WhatsApp, email and social sharing.')],
    fields: [xf('session', 'Sign out inactive people after', 'select', { options: ['15', '30', '45', '60', '120', '300', '720', '1440'], optLabel: (o) => (Number(o) < 60 ? o + ' minutes' : Number(o) === 60 ? '1 hour' : Number(o) / 60 + ' hours') }), xf('retention', 'Keep audit logs for', 'select', { options: ['3', '6', '12', '24'], optLabel: (o) => o + ' months' }), xf('maxFileMB', 'Largest file on a communication', 'select', { options: ['2', '5', '10', '25', '50'], optLabel: (o) => o + ' MB' }), xf('fileTypes', 'File types allowed on communications', 'text', { placeholder: 'e.g. pdf, docx, xlsx, png', wide: true }), xf('maxVideoMB', 'Largest video that can be posted', 'select', { options: ['5', '10', '25', '50', '100'], optLabel: (o) => o + ' MB' }),
      // how much each person may store in total; one person can be given a different amount below
      xf('userQuotaMB', 'Storage each person may use', 'select', { options: ['100', '250', '500', '1024', '2048', '5120', '10240', '20480', '51200', '102400', '0'], optLabel: (o) => (o === '0' ? 'No limit' : Number(o) >= 1024 ? Number(o) / 1024 + ' GB' : o + ' MB') })],
    dirty: dirtyOf('security', X) , discard: () => clearPart('security'),
    save: () => { if (this.srv) { srvDo(() => settingsW.saveSecurity(this, this.srv, { ...X, ...(draft.security || {}) }), 'Security and data settings saved.').then(() => clearPart('security')); return; } this.setState((st) => ({ orgSecurity: { ...st.orgSecurity, ...(st.setDraft || {}).security } })); clearPart('security'); audit('Updated security and data settings'); this.notify('Security and data settings saved.', '#B8F36B'); },
    exportData: () => { const data = { exportedAt: new Date().toISOString(), organization: P, people: this.dir.map((p) => ({ name: p.name, title: p.title, department: p.dept, branch: p.loc, email: p.email || '' })), subscription: S }; downloadBlob(new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' }), 'organization-data.json'); audit('Exported all organization data'); this.notify('Organization data exported.', '#7CC4FF'); },
    deleteAfter: s.orgDeleteAfter || '',
    deleteNote: s.orgDeleteAfter ? 'Deletion is scheduled for ' + new Date(s.orgDeleteAfter).toLocaleDateString('en-GB', { day: 'numeric', month: 'long', year: 'numeric' }) + '. Everything is erased then and cannot be recovered. You can cancel until that day.' : '',
    cancelDeletion: () => srvDo(() => settingsW.cancelDeletion(this, this.srv), 'Deletion cancelled. Your organization stays.'),
    del: () => ask({ kind: 'delete', text: 'Request deletion of ' + P.name + ' and all its data? Everything is erased after 30 days, and you can cancel any time before then. Type the organization name to confirm.', confirmLabel: 'Request deletion', needsName: true }) };
  security.dirtyAny = security.dirty;

  const confirm = cf ? { open: true, text: cf.text, label: cf.confirmLabel, needsName: !!cf.needsName, typed: draft.typed || '', onTyped: (e) => this.setState((st) => ({ setDraft: { ...(st.setDraft || {}), typed: e.target.value } })), ready: !cf.needsName || (draft.typed || '').trim() === P.name,
    cancel: () => { this.setState((st) => { const d = { ...(st.setDraft || {}) }; delete d.typed; return { setConfirm: null, setDraft: Object.keys(d).filter((k) => k !== 'errs').length ? d : null }; }); },
    run: () => {
      if (this.srv && (cf.kind === 'plan' || cf.kind === 'cancel')) {
        const p = PLANS.find((x) => x.id === cf.id);
        const patch = cf.kind === 'plan' ? { plan: p.id, seats: Math.min(S.seats, p.seats) } : { status: 'Cancelling', cancelAt: S.renews || new Date(Date.now() + 30 * 86400000).toISOString().slice(0, 10) };
        this.setState({ setConfirm: null });
        subSave(patch, cf.kind === 'plan' ? 'Subscription changed to ' + p.name + '.' : 'Subscription cancelled. Access continues until ' + dayLabel(patch.cancelAt) + '.');
        return;
      }
      if (cf.kind === 'plan') { const p = PLANS.find((x) => x.id === cf.id); this.setState((st) => ({ orgSub: { ...st.orgSub, plan: p.id, seats: Math.min(st.orgSub.seats, p.seats) } })); done('Subscription changed to ' + p.name); }
      if (cf.kind === 'cancel') { this.setState((st) => ({ orgSub: { ...st.orgSub, status: 'Cancelling', cancelAt: st.orgSub.renews } })); done('Subscription cancelled. Access continues until ' + dayLabel(S.renews)); }
      if (cf.kind === 'delete') {
        this.setState((st) => { const d = { ...(st.setDraft || {}) }; delete d.typed; return { setConfirm: null, setDraft: Object.keys(d).filter((k) => k !== 'errs').length ? d : null }; });
        if (this.srv) { srvDo(() => settingsW.requestDeletion(this, this.srv, P.name), 'Deletion requested. Everything is erased in 30 days unless you cancel.'); return; }
        audit('Requested deletion of the organization'); this.notify('Deletion requested. Everything is erased in 30 days unless you cancel.', '#FF9B8A');
      }
    } } : { open: false };

  // ---- ETR tax invoice / receipt ----
  const EO = s.etrOpen;
  const inv = EO ? s.invoices.find((x) => x.no === EO.no) : null;
  // Connected: the document is the server's (its numbers, VAT split and fiscal details). Otherwise the sample layout.
  const serverDoc = inv && this.srv ? settingsW.ensureDoc(this, this.srv, inv.no) : null;
  const doc = !inv ? null : this.srv ? (serverDoc ? { ...serverDoc, kind: EO.kind, title: EO.kind === 'receipt' ? 'PAYMENT RECEIPT' : 'TAX INVOICE', method: B.cardBrand + ' •••• ' + B.cardLast4 } : null)
    : buildDoc({ invoice: inv, kind: EO.kind, plan, seats: S.seats, cycle: S.cycle, addons: addonsOn,
      buyer: { name: P.name, pin: P.taxId || 'Not provided', address: [P.address, P.city, P.country].filter(Boolean).join(', ') || '—', email: B.contactEmail || P.email || '—', phone: P.phone || '' },
      method: B.cardBrand + ' •••• ' + B.cardLast4 });
  const etr = doc ? { open: true, ...doc,
    dateLabel: new Date(inv.date + 'T10:30:00').toLocaleString('en-GB', { day: '2-digit', month: '2-digit', year: 'numeric', hour: '2-digit', minute: '2-digit', hour12: false }),
    switchKind: (k) => this.setState({ etrOpen: { no: EO.no, kind: k } }), close: () => this.setState({ etrOpen: null }) } : { open: false };
  const tabDefs = [['profile', 'Organization profile'], ['subscription', 'Subscription'], ['billing', 'Billing & invoices'], ['security', 'Security & data']];
  const settingsPage = { canEdit, canSub, tab: s.setTab, tabs: tabDefs.map(([k, l]) => ({ label: l, on: s.setTab === k, go: () => this.setState({ setTab: k }) })), isProfile: s.setTab === 'profile', isSub: s.setTab === 'subscription', isBilling: s.setTab === 'billing', isSecurity: s.setTab === 'security',
    profile, sub, billing, security, confirm, etr, orgName: P.name };
  return { settingsPage, orgName: P.name, orgLogo: P.logo };
}
