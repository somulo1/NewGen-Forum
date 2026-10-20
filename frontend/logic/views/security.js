import * as securityW from '~/logic/server/widgets/security.js';
import { buildXlsx, downloadBlob } from '~/logic/xlsx.js';
// renderVals step: security. Receives the names built by earlier steps and returns the ones later steps need.
export function buildSecurity(ctx) {
  const { badge, chipS, nowMs, s, segS, set, ts } = ctx;
  const iso9 = (m) => new Date(nowMs - m * 60000).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' });
  // ---- Security & logs ----
  const ruleDefs = s.secRuleDefs && s.secRuleDefs.length ? s.secRuleDefs : [
    { id: 'r1', name: 'Repeated failed sign-ins', desc: '5 or more failed sign-ins on one account within 10 minutes.', sev: 'High' },
    { id: 'r2', name: 'Sign-in from a new country', desc: 'Someone signs in from a country they have never used before.', sev: 'High' },
    { id: 'r3', name: 'Mass file download', desc: 'More than 50 files downloaded within 15 minutes.', sev: 'High' },
    { id: 'r4', name: 'After-hours admin action', desc: 'Admin changes made outside working hours.', sev: 'Medium' },
    { id: 'r5', name: 'Role or rights change', desc: 'Any change to roles, rights or approval rules.', sev: 'Medium' },
    { id: 'r6', name: 'Secure link opened from an unusual IP', desc: 'A one-time link is opened from outside the usual IP ranges.', sev: 'Medium' },
    { id: 'r7', name: 'Outside sharing spike', desc: 'More than 10 posts shared outside Hub within an hour.', sev: 'Low' }
  ];
  const ruleSev = (r) => s.ruleSev[r.id] || r.sev;
  const ruleOn = (id) => !!s.rulesOn[id];
  const sevColor = { High: '#FF9B8A', Medium: '#FFB86B', Low: '#7CC4FF' };
  const agoLabel = (m) => { const d = new Date(nowMs - m * 60000); const t = d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', timeZone: 'Africa/Nairobi' }); const day = d.toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) === new Date(nowMs).toLocaleDateString('en-CA', { timeZone: 'Africa/Nairobi' }) ? 'Today' : 'Yesterday'; return day + ' ' + t; };
  const allLogs = s.extraLogs.concat(this.baseLogs);
  const alertDefs = [];   // alerts come from the server (s.alertsLive)
  const alertList = s.alertsLive || alertDefs;
  const newRightsAlerts = this.srv ? [] : s.extraLogs.filter((l) => l.rule === 'r5').slice(0, 3).map((l, i) => ({ id: 'AL-' + (119 + i), rule: 'r5', ago: 0, title: l.action, detail: 'Made by ' + l.user + '. Logged automatically.' }));
  const alertsAll = newRightsAlerts.concat(alertList).filter((a) => ruleOn(a.rule)).map((a) => {
    const r = ruleDefs.find((x) => x.id === a.rule) || { name: a.rule, sev: 'Medium' };
    const sv = ruleSev(r);
    const st = s.alertSt[a.id] || 'Open';
    const stC = st === 'Open' ? '#FF9B8A' : st === 'Investigating' ? '#FFB86B' : '#9DB0A6';
    const notes = s.alertNotes[a.id] || [];
    const closed = st !== 'Open' && st !== 'Investigating';
    const openDlg = (mode) => () => this.setState({ alertDlg: { id: a.id, mode, outcome: 'Resolved', comment: '', err: '' } });
    return { ...a, sev: sv, color: sevColor[sv], sevStyle: badge(sevColor[sv]), ruleName: r.name, when: a.ago === 0 ? 'Just now' : agoLabel(a.ago), status: st, statusStyle: badge(stC),
      notes, lastNote: notes.length ? notes[notes.length - 1] : null,
      investigate: openDlg('investigate'), closeAlert: openDlg('close'), falsePos: openDlg('falsepos'), history: openDlg('history'), isOpen: st === 'Open' || st === 'Investigating', canInvestigate: st === 'Open', canClose: st === 'Investigating', canFalsePos: st === 'Open' || st === 'Investigating', isClosed: closed };
  });
  const openAlerts = alertsAll.filter((a) => a.isOpen);
  const lq = s.lq.trim().toLowerCase();
  const logsF = allLogs.filter((l) => (s.lmod === 'All' || l.module === s.lmod) && (s.lres === 'All' || l.result === s.lres) && (!s.lflag || (l.rule && ruleOn(l.rule))) && (!lq || (l.user + ' ' + l.action + ' ' + l.ip).toLowerCase().indexOf(lq) !== -1));
  const flaggedHours = {};
  allLogs.forEach((l) => { if (l.rule && ruleOn(l.rule)) flaggedHours[Math.floor(l.ago / 60)] = true; });
  // ---- Time range: the tables and the Excel export use the same window ----
  const R = s.secRange || { preset: 'all', from: '', to: '' };
  const PRESET = { '24h': 1440, '7d': 10080, '30d': 43200 };
  const inRange = (ago) => {
    if (R.preset === 'custom') { const d = iso9(ago); return (!R.from || d >= R.from) && (!R.to || d <= R.to); }
    return PRESET[R.preset] ? ago <= PRESET[R.preset] : true;
  };
  const alertsRange = alertsAll.filter((a) => inRange(a.ago));
  const logsRange = allLogs.filter((l) => inRange(l.ago));
  const setRange = (patch) => this.setState((x) => ({ secRange: { ...x.secRange, ...patch } }));
  const range = { preset: R.preset, from: R.from, to: R.to, isCustom: R.preset === 'custom',
    presets: [['all', 'All time'], ['24h', 'Last 24 hours'], ['7d', 'Last 7 days'], ['30d', 'Last 30 days'], ['custom', 'Custom']].map(([k, l]) => ({ label: l, on: R.preset === k, pick: () => setRange({ preset: k }) })),
    onFrom: (e) => setRange({ preset: 'custom', from: e.target.value }), onTo: (e) => setRange({ preset: 'custom', to: e.target.value }),
    label: R.preset === 'custom' ? (R.from || 'start') + ' to ' + (R.to || 'today') : ({ all: 'All time', '24h': 'Last 24 hours', '7d': 'Last 7 days', '30d': 'Last 30 days' })[R.preset],
    summary: alertsRange.length + ' alerts · ' + logsRange.length + ' log entries in this range' };
  const stamp = (ago) => new Date(nowMs - ago * 60000).toLocaleString('en-GB', { timeZone: 'Africa/Nairobi', hour12: false });
  const exportXlsx = () => {
    const tab = s.secTab;
    let name, head, body;
    if (tab === 'alerts') {
      name = 'Alerts'; head = ['ID', 'Severity', 'Alert', 'Details', 'Rule', 'Raised', 'Status', 'Comments'];
      body = alertsRange.map((a) => [a.id, a.sev, a.title, a.detail, a.ruleName, stamp(a.ago), a.status, a.notes.map((n) => new Date(n.at).toLocaleString('en-GB', { hour12: false }) + ' · ' + n.who + ' · ' + n.status + ': ' + n.comment).join('\n')]);
    } else if (tab === 'logs') {
      name = 'Audit logs'; head = ['When', 'User', 'Action', 'Module', 'IP', 'Location', 'Result', 'Flagged by rule'];
      body = logsRange.map((l) => { const r = l.rule && ruleOn(l.rule) ? ruleDefs.find((x) => x.id === l.rule) : null; return [stamp(l.ago), l.user, l.action, l.module, l.ip, l.loc, l.result, r ? r.name : '']; });
    } else {
      name = 'Detection rules'; head = ['Rule', 'Description', 'Severity', 'On', 'Matches today'];
      body = ruleDefs.map((r) => [r.name, r.desc, ruleSev(r), ruleOn(r.id) ? 'Yes' : 'No', allLogs.filter((l) => l.rule === r.id).length]);
    }
    const day = (d) => d || '';
    const rangeTag = tab === 'rules' ? 'all' : R.preset === 'custom' ? (day(R.from) || 'start') + '_to_' + (day(R.to) || 'today') : R.preset;
    downloadBlob(buildXlsx(name, [head].concat(body)), 'security-' + tab + '-' + rangeTag + '.xlsx');
    if (this.srv) this.srv.request('GET', '/security/logs/export?limit=1').catch(() => {});   // records the export in the audit log
    if (!this.srv) this.setState((x) => ({ extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: 'Exported ' + body.length + ' ' + name.toLowerCase() + ' rows to Excel (' + range.label + ')', module: 'Security', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }].concat(x.extraLogs) }));
    this.notify('Exported ' + body.length + ' ' + name.toLowerCase() + (tab === 'rules' ? '' : ' (' + range.label.toLowerCase() + ')') + ' to Excel. The export was logged.', '#7CC4FF');
  };
  // ---- Status change dialog: every change needs a comment; Open -> Investigating -> Closed ----
  const AD = s.alertDlg;
  const adAlert = AD ? alertsAll.find((x) => x.id === AD.id) : null;
  const setDlg = (patch) => this.setState((x) => ({ alertDlg: { ...x.alertDlg, ...patch } }));
  const alertDialog = adAlert ? { open: true, mode: AD.mode, isHistory: AD.mode === 'history', isClose: AD.mode === 'close',
    title: AD.mode === 'investigate' ? 'Start investigating ' + adAlert.id : AD.mode === 'close' ? 'Close ' + adAlert.id : AD.mode === 'falsepos' ? 'Mark ' + adAlert.id + ' as a false positive' : adAlert.id + ' history', alertTitle: adAlert.title, status: adAlert.status, statusStyle: adAlert.statusStyle,
    outcomes: ['Resolved', 'False positive'].map((o) => ({ label: o, on: AD.outcome === o, pick: () => setDlg({ outcome: o }) })),
    commentLabel: (AD.mode === 'falsepos' || (AD.mode === 'close' && AD.outcome === 'False positive')) ? 'Comment (optional)' : 'Comment (required)',
    comment: AD.comment, err: AD.err, hasErr: !!AD.err, onComment: (e) => setDlg({ comment: e.target.value, err: '' }), addEmoji: (ch) => setDlg({ comment: AD.comment + ch }),
    history: adAlert.notes.map((n) => ({ ...n, when: new Date(n.at).toLocaleString('en-GB', { hour12: false }) })).reverse(), hasHistory: adAlert.notes.length > 0,
    confirmLabel: AD.mode === 'investigate' ? 'Start investigating' : AD.mode === 'falsepos' ? 'Mark false positive' : 'Close alert',
    cancel: () => this.setState({ alertDlg: null }),
    confirm: () => {
      const c = (this.state.alertDlg.comment || '').trim();
      const to = AD.mode === 'investigate' ? 'Investigating' : AD.mode === 'falsepos' ? 'False positive' : AD.outcome;
      if (to !== 'False positive' && c.length < 3) { setDlg({ err: 'Add a comment explaining this change.' }); return; }
      if (this.srv) { this.setState({ alertDlg: null }); securityW.setAlert(this, this.srv, adAlert.id, to, c).then(() => this.notify(adAlert.id + ' ' + (to === 'Investigating' ? 'is now being investigated.' : 'closed as ' + to.toLowerCase() + '.'), '#7CC4FF')).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); return; }
      this.setState((x) => ({ alertDlg: null, alertSt: { ...x.alertSt, [adAlert.id]: to }, alertNotes: { ...x.alertNotes, [adAlert.id]: (x.alertNotes[adAlert.id] || []).concat([{ status: to, comment: c, who: 'Sam', at: Date.now() }]) },
        extraLogs: [{ id: 'x' + Date.now(), ago: 0, user: 'Samwel Okoth Omulo', action: adAlert.id + ' moved to ' + to + ': ' + c, module: 'Security', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: '' }].concat(x.extraLogs) }));
      this.notify(adAlert.id + ' ' + (to === 'Investigating' ? 'is now being investigated.' : 'closed as ' + to.toLowerCase() + '.'), '#7CC4FF');
    } } : { open: false };
  const tabS = segS;
  const sec = {
    kpis: [
      { label: 'Open alerts', n: openAlerts.length, color: '#FF9B8A', sub: openAlerts.filter((a) => a.sev === 'High').length + ' high severity' },
      { label: 'Failed sign-ins (24 h)', n: allLogs.filter((l) => l.result === 'Failed' && l.ago < 1440).length, color: '#FFB86B', sub: 'Across ' + new Set(allLogs.filter((l) => l.result === 'Failed').map((l) => l.user)).size + ' account' },
      this.srv ? (() => {
        // sign-ins that were stopped in the last 24 hours: lockouts after wrong passwords, and addresses that kept trying
        const day = (s.signBlocks || []).filter((b) => Date.now() - new Date(b.at).getTime() < 86400000);
        const now = day.filter((b) => new Date(b.until).getTime() > Date.now()).length;
        return { label: 'Blocked sign-ins (24 h)', n: day.length, color: '#7CC4FF', sub: now ? now + ' blocked right now' : 'None blocked now' };
      })() : { label: 'Blocked attempts', n: allLogs.filter((l) => l.result === 'Blocked').length, color: '#7CC4FF', sub: 'Country and lockout rules' },
      this.srv ? { label: 'Sign-ins (24 h)', n: allLogs.filter((l) => l.action === 'Signed in' && l.ago < 1440).length, color: '#B8F36B', sub: new Set(allLogs.filter((l) => l.action === 'Signed in' && l.ago < 1440).map((l) => l.user)).size + ' people' } : { label: 'Active sessions', n: 186, color: '#B8F36B', sub: '3 on new devices' }
    ],
    bars: Array.from({ length: 24 }, (_, i) => { const h = 23 - i; const v = this.srv ? Math.min(18, allLogs.filter((l) => Math.floor(l.ago / 60) === h).length) : [3, 2, 2, 1, 1, 2, 4, 9, 14, 16, 15, 13, 12, 14, 15, 13, 11, 8, 6, 5, 4, 4, 3, 3][(i + 5) % 24]; return { style: 'flex:1;border-radius:3px 3px 0 0;height:' + (12 + v * 4.6) + '%;background:' + (flaggedHours[h] ? '#FF9B8A' : '#2C3D34') }; }),
    alerts: (s.openOnly ? openAlerts : alertsAll), noAlerts: (s.openOnly ? openAlerts : alertsAll).length === 0,
    toggleOpenOnly: () => this.setState((x) => ({ openOnly: !x.openOnly })), openOnlyStyle: chipS(s.openOnly),
    lq: s.lq, onLq: (e) => this.setState({ lq: e.target.value, lshow: 25 }),
    lmod: s.lmod, onLmod: (e) => this.setState({ lmod: e.target.value, lshow: 25 }),
    lres: s.lres, onLres: (e) => this.setState({ lres: e.target.value, lshow: 25 }),
    lflag: s.lflag, toggleLflag: () => this.setState((x) => ({ lflag: !x.lflag, lshow: 25 })),
    exportXlsx, range, alertDialog,
    logCountLabel: logsF.length + ' entries · showing ' + Math.min(logsF.length, s.lshow),
    logs: logsF.slice(0, s.lshow).map((l) => {
      const fl = !!(l.rule && ruleOn(l.rule));
      const r = fl ? ruleDefs.find((x) => x.id === l.rule) : null;
      const c = fl ? sevColor[ruleSev(r)] : '';
      return { ...l, time: l.ago === 0 ? 'Just now' : agoLabel(l.ago), flagged: fl, ruleName: r ? r.name : '', flagStyle: fl ? badge(c) : '',
        resColor: l.result === 'Success' ? '#C9D6CF' : l.result === 'Failed' ? '#FF9B8A' : '#FFB86B',
        rowStyle: 'border-top:1px solid var(--k-24322C,#24322C);' + (fl ? 'background:' + c + '12' : '') };
    }),
    hasMore: logsF.length > s.lshow, more: () => this.setState((x) => ({ lshow: x.lshow + 25 })),
    rules: ruleDefs.map((r) => ({ ...r, on: ruleOn(r.id), sev: ruleSev(r), hits: allLogs.filter((l) => l.rule === r.id).length + ' matches today',
      toggle: () => { const next = !ruleOn(r.id); this.setState((x) => ({ rulesOn: { ...x.rulesOn, [r.id]: next } })); if (this.srv) securityW.setRule(this, this.srv, r.id, next, ruleSev(r)).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); },
      setSev: (e) => { const v = e.target.value; this.setState((x) => ({ ruleSev: { ...x.ruleSev, [r.id]: v } })); if (this.srv) securityW.setRule(this, this.srv, r.id, ruleOn(r.id), v).catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A')); } }))
  };
  const secTab = { alerts: tabS(s.secTab === 'alerts'), logs: tabS(s.secTab === 'logs'), rules: tabS(s.secTab === 'rules'), blocks: tabS(s.secTab === 'blocks'),
    alertsGo: set({ secTab: 'alerts' }), logsGo: set({ secTab: 'logs' }), rulesGo: set({ secTab: 'rules' }), blocksGo: set({ secTab: 'blocks' }),
    isAlerts: s.secTab === 'alerts', isLogs: s.secTab === 'logs', isRules: s.secTab === 'rules', isBlocks: s.secTab === 'blocks' };
  return { agoLabel, alertsAll, alertsRange, logsRange, allLogs, openAlerts, ruleDefs, ruleOn, ruleSev, sec, secTab, sevColor, tabS };
}
