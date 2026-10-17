<!-- Generated from Table.dc.html — exact port of the Claude Design prototype. -->
<template>
<div class="dc-root" style="display: contents">
<div style="font-family: 'DM Sans', system-ui, sans-serif; color: #E8EFEA; font-size: 13px; display: flex; flex-direction: column; gap: 10px">
<div style="overflow-x: auto; border-radius: 14px; border: 1px solid #24322C; background: #16201C">
<table :style="`width: 100%; min-width: ${minW}; border-collapse: collapse`">
<thead>
<tr><template v-for="(c, c__i) in cols" :key="c__i"><th scope="col" :style="c.thStyle">{{ c.label }}</th></template></tr>
<tr><template v-for="(c, c__i) in cols" :key="c__i"><td style="padding: 0 8px 10px; background: #121B17; vertical-align: top; border-bottom: 1px solid #24322C">
<template v-if="c.isText"><input type="search" :aria-label="`Filter ${c.label}`" placeholder="Filter…" :value="c.val" @input="c.onVal" style="width: 100%; min-width: 80px; box-sizing: border-box; height: 32px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 12px"></template>
<template v-if="c.isSelect"><select :aria-label="`Filter ${c.label}`" :value="c.val" @change="c.onVal" style="width: 100%; min-width: 90px; box-sizing: border-box; height: 32px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 12px"><option value="">All</option><template v-for="(o, o__i) in c.opts" :key="o__i"><option :value="o.v">{{ o.v }}</option></template></select></template>
<template v-if="c.isDate"><div style="display: flex; flex-direction: column; gap: 4px; min-width: 130px"><input type="date" :aria-label="`${c.label} from`" title="From" :value="c.from" @input="c.onFrom" style="height: 30px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 12px; color-scheme: dark"><input type="date" :aria-label="`${c.label} to`" title="To" :value="c.to" @input="c.onTo" style="height: 30px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 12px; color-scheme: dark"></div></template>
</td></template></tr>
</thead>
<tbody>
<template v-for="(r, r__i) in rows" :key="r__i">
<tr :style="r.style"><template v-for="(cl, cl__i) in r.cells" :key="cl__i"><td :style="cl.tdStyle">
<template v-if="cl.isText"><div style="display: flex; flex-direction: column; gap: 2px"><span :style="cl.textStyle">{{ cl.text }}</span><template v-if="cl.hasSub"><span style="font-size: 11px; color: #9DB0A6; line-height: 1.35">{{ cl.sub }}</span></template></div></template>
<template v-if="cl.isBadge"><span :style="cl.badgeStyle">{{ cl.text }}</span><template v-if="cl.hasSub"><div style="font-size: 11px; color: #9DB0A6; margin-top: 4px">{{ cl.sub }}</div></template></template>
<template v-if="cl.isPerson"><div style="display: flex; align-items: center; gap: 10px"><span style="position: relative; flex: none"><UserAvatar :name="cl.text" :initials="cl.initials" :color="cl.color" :size="32" :font-size="11"/><template v-if="cl.hasDot"><span :style="`position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: ${cl.dot}; border: 2px solid #16201C`"></span></template></span><div style="display: flex; flex-direction: column; min-width: 0"><strong style="font-size: 13px; font-weight: 600; white-space: nowrap">{{ cl.text }}</strong><template v-if="cl.hasSub"><span style="font-size: 11px; color: #9DB0A6">{{ cl.sub }}</span></template></div></div></template>
<template v-if="cl.isAvatars"><div style="display: flex; align-items: center"><template v-for="(av, av__i) in cl.list" :key="av__i"><UserAvatar :name="av.name" :initials="av.initials" :color="av.color" :size="26" :font-size="9" border="2px solid #16201C" :aria-label="av.name" style="margin-right: -7px"/></template><span style="font-size: 11px; color: #9DB0A6; margin-left: 14px; white-space: nowrap">{{ cl.sub }}</span></div></template>
<template v-if="cl.isCheck"><div style="display: flex; flex-direction: column; align-items: center; gap: 2px"><input type="checkbox" :aria-label="cl.label" :checked="cl.on" :disabled="cl.disabled" @change="cl.toggle" style="width: 18px; height: 18px; accent-color: #B8F36B"><template v-if="cl.pending"><span style="font-size: 9px; font-weight: 700; color: #FFB86B">PENDING</span></template></div></template>
<template v-if="cl.isSelect"><select :aria-label="cl.label" :value="cl.value" :disabled="cl.disabled" @change="cl.onChange" :style="cl.selectStyle"><template v-for="(op, op__i) in cl.options" :key="op__i"><option :value="op.v">{{ op.l }}</option></template></select></template>
<template v-if="cl.isBar"><div style="display: flex; align-items: center; gap: 8px; min-width: 120px"><div style="flex: 1; height: 6px; border-radius: 3px; background: #22302A; overflow: hidden"><div :style="cl.barStyle"></div></div><span style="font-size: 11px; font-weight: 600; min-width: 34px; text-align: right">{{ cl.text }}</span></div></template>
<template v-if="cl.isActions"><div style="display: flex; gap: 6px; justify-content: flex-end"><template v-for="(b, b__i) in cl.buttons" :key="b__i"><button @click="b.run" :disabled="b.disabled" :aria-label="b.aria" :style="b.style">{{ b.label }}</button></template></div></template>
</td></template></tr>
</template>
<template v-if="empty"><tr><td :colspan="colCount" style="padding: 30px; text-align: center; color: #9DB0A6">{{ emptyText }}</td></tr></template>
</tbody>
</table>
</div>
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; font-size: 12px; color: #9DB0A6">
<span style="flex: 1 1 160px">{{ countLabel }}</span>
<template v-if="filtered"><button @click="clear" style="height: 32px; padding: 0 12px; border-radius: 16px; border: 0; background: transparent; color: #B8F36B; cursor: pointer; font-size: 12px; font-weight: 600">Clear filters</button></template>
<button @click="prev" :disabled="atStart" :style="prevStyle">Previous</button>
<span style="color: #C9D6CF">Page {{ pageNo }} of {{ pages }}</span>
<button @click="next" :disabled="atEnd" :style="nextStyle">Next</button>
</div>
</div>
</div>
</template>

<script>
import { DCLogic, useDC } from '../../utils/dcLogic.js';

class Component extends DCLogic {
  constructor(props) {
    super(props);
    this.state = { tid: null, f: {}, from: {}, to: {}, page: 0 };
  }
  renderVals() {
    const sample = {
      id: 'sample', cols: [{ label: 'Name', filter: 'text' }, { label: 'Team', filter: 'select' }, { label: 'Date', filter: 'date' }, { label: '', align: 'right' }],
      rows: [
        { cells: [{ kind: 'person', text: 'Amina Wanjiru', initials: 'AW', color: '#6EE7C8', sub: 'Finance Analyst' }, { text: 'Finance' }, { text: '14 Oct', fv: '2026-10-14' }, { kind: 'actions', buttons: [{ label: 'View', run: () => {} }] }] },
        { cells: [{ kind: 'person', text: 'Brian Kiprop', initials: 'BK', color: '#7CC4FF', sub: 'Systems Engineer' }, { text: 'IT Operations' }, { text: '15 Oct', fv: '2026-10-15' }, { kind: 'actions', buttons: [{ label: 'View', run: () => {} }] }] }
      ]
    };
    const t = this.props.t || sample;
    const st = this.state;
    const same = st.tid === t.id;
    const F = same ? st.f : {}, FR = same ? st.from : {}, TO = same ? st.to : {};
    const cols = t.cols || [];
    const all = t.rows || [];
    const ps = t.pageSize || 10;
    const fvOf = (cell) => String(cell && cell.fv != null ? cell.fv : (cell && cell.text != null ? cell.text : ''));
    const rows = all.filter((r) => cols.every((c, i) => {
      const cell = (r.cells || [])[i] || {};
      const fv = fvOf(cell);
      if (c.filter === 'text') { const q = String(F[i] || '').trim().toLowerCase(); return !q || (fv + ' ' + (cell.sub || '')).toLowerCase().indexOf(q) !== -1; }
      if (c.filter === 'select') { const v = F[i] || ''; return !v || fv === v; }
      if (c.filter === 'date') { const a = FR[i] || '', b = TO[i] || ''; return (!a || fv >= a) && (!b || fv <= b); }
      return true;
    }));
    const pages = Math.max(1, Math.ceil(rows.length / ps));
    const page = Math.min(same ? st.page : 0, pages - 1);
    const upd = (patch) => this.setState((x) => {
      const keep = x.tid === t.id;
      return Object.assign({ tid: t.id, f: keep ? x.f : {}, from: keep ? x.from : {}, to: keep ? x.to : {}, page: 0 }, typeof patch === 'function' ? patch(keep ? x : { f: {}, from: {}, to: {} }) : patch);
    });
    const filtered = Object.keys(F).some((k) => F[k]) || Object.keys(FR).some((k) => FR[k]) || Object.keys(TO).some((k) => TO[k]);
    const btnStyle = (b) => 'height:32px;padding:0 12px;border-radius:8px;font-size:12px;font-weight:600;white-space:nowrap;cursor:' + (b.disabled ? 'default;opacity:0.45' : 'pointer') + ';' +
      (b.kind === 'primary' ? 'border:0;background:#B8F36B;color:#0F1714' : b.kind === 'danger' ? 'border:1px solid #5A3A36;background:transparent;color:#FF9B8A' : 'border:1px solid #2A3832;background:transparent;color:#E8EFEA');
    const pager = (dis) => 'height:32px;padding:0 12px;border-radius:16px;border:1px solid #2A3832;background:transparent;font-size:12px;cursor:' + (dis ? 'default;color:#4A5C54' : 'pointer;color:#E8EFEA');
    return {
      minW: t.minWidth || '760px',
      colCount: cols.length,
      cols: cols.map((c, i) => {
        const opts = c.filter === 'select' ? (c.options || Array.from(new Set(all.map((r) => fvOf((r.cells || [])[i])).filter(Boolean))).sort()) : [];
        return {
          label: c.label,
          thStyle: 'padding:10px 12px 6px;background:#121B17;text-align:' + (c.align || 'left') + ';color:#9DB0A6;font-size:11px;font-weight:600;letter-spacing:0.05em;text-transform:uppercase;white-space:nowrap',
          isText: c.filter === 'text', isSelect: c.filter === 'select', isDate: c.filter === 'date',
          val: F[i] || '', from: FR[i] || '', to: TO[i] || '',
          opts: opts.map((v) => ({ v: v })),
          onVal: (e) => { const v = e.target.value; upd((x) => ({ f: Object.assign({}, x.f, { [i]: v }) })); },
          onFrom: (e) => { const v = e.target.value; upd((x) => ({ from: Object.assign({}, x.from, { [i]: v }) })); },
          onTo: (e) => { const v = e.target.value; upd((x) => ({ to: Object.assign({}, x.to, { [i]: v }) })); }
        };
      }),
      rows: rows.slice(page * ps, (page + 1) * ps).map((r) => ({
        style: 'border-top:1px solid #24322C;' + (r.tint ? 'background:' + r.tint + '14' : ''),
        cells: cols.map((c, i) => {
          const cl = (r.cells || [])[i] || { text: '' };
          const k = cl.kind || 'text';
          return Object.assign({}, cl, {
            tdStyle: 'padding:10px 12px;vertical-align:middle;text-align:' + (c.align || 'left'),
            isText: k === 'text', isBadge: k === 'badge', isPerson: k === 'person', isAvatars: k === 'avatars', isCheck: k === 'check', isSelect: k === 'select', isBar: k === 'bar', isActions: k === 'actions',
            hasSub: !!cl.sub, hasDot: !!cl.dot,
            textStyle: 'font-size:13px;line-height:1.4;color:' + (cl.color || '#E8EFEA') + ';' + (cl.strong ? 'font-weight:600;' : '') + (cl.mono ? 'font-family:Space Grotesk,sans-serif;' : '') + (cl.nowrap ? 'white-space:nowrap;' : ''),
            badgeStyle: 'display:inline-block;padding:3px 10px;border-radius:12px;font-size:11px;font-weight:600;white-space:nowrap;background:' + (cl.color || '#9DB0A6') + '22;color:' + (cl.color || '#9DB0A6') + ';border:1px solid ' + (cl.color || '#9DB0A6') + '55',
            barStyle: 'height:100%;width:' + Math.max(2, Math.min(100, cl.pct || 0)) + '%;background:' + (cl.color || '#B8F36B') + ';border-radius:3px',
            selectStyle: 'height:34px;min-width:' + (cl.minWidth || '130px') + ';padding:0 6px;border-radius:8px;border:1px solid #2A3832;background:#121B17;color:#E8EFEA;font-size:12px;' + (cl.disabled ? 'opacity:0.4' : ''),
            buttons: (cl.buttons || []).map((b) => Object.assign({}, b, { aria: b.aria || b.label, style: btnStyle(b), disabled: !!b.disabled }))
          });
        })
      })),
      empty: rows.length === 0,
      emptyText: t.emptyText || 'Nothing matches these filters.',
      countLabel: rows.length ? ('Showing ' + (page * ps + 1) + '–' + Math.min(rows.length, (page + 1) * ps) + ' of ' + rows.length + (rows.length !== all.length ? ' (filtered from ' + all.length + ')' : '')) : '0 results',
      filtered,
      clear: () => this.setState({ tid: t.id, f: {}, from: {}, to: {}, page: 0 }),
      pageNo: page + 1, pages,
      atStart: page === 0, atEnd: page >= pages - 1,
      prevStyle: pager(page === 0), nextStyle: pager(page >= pages - 1),
      prev: () => this.setState({ tid: t.id, f: F, from: FR, to: TO, page: Math.max(0, page - 1) }),
      next: () => this.setState({ tid: t.id, f: F, from: FR, to: TO, page: Math.min(pages - 1, page + 1) })
    };
  }
}

export default {
  name: 'DataTable',
  props: { t: { type: Object, default: null } },
  ...useDC(Component, ["atEnd","atStart","clear","colCount","cols","countLabel","empty","emptyText","filtered","minW","next","nextStyle","pageNo","pages","prev","prevStyle","rows"])
};
</script>
