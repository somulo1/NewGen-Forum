<!-- One row of tabs. When they do not all fit, the row scrolls sideways with arrows at its ends and a fade over the cut-off side, so the
     hidden ones are noticed. An arrow blinks when a tab with something new is hidden on its side. The chosen tab is always brought into view. -->
<template>
<div style="position: relative; display: flex; align-items: stretch; min-width: 0; border-bottom: 1px solid var(--k-24322C,#24322C)">
<button v-if="left" type="button" aria-label="Show earlier tabs" title="Earlier tabs" @click="move(-1)" :style="arrow"><span v-if="newLeft" class="hub-blink" aria-hidden="true"></span>‹</button>
<div ref="row" role="tablist" @scroll.passive="measure" @wheel="wheel" class="hub-scrolltabs" :style="`flex: 1; min-width: 0; display: flex; gap: 4px; overflow-x: auto; overscroll-behavior-x: contain; scroll-behavior: smooth; ${mask}`"><slot /></div>
<button v-if="right" type="button" aria-label="Show more tabs" title="More tabs" @click="move(1)" :style="arrow"><span v-if="newRight" class="hub-blink" aria-hidden="true"></span>›</button>
</div>
</template>

<script>
export default {
  name: 'ScrollTabs',
  props: { active: { type: [String, Number], default: '' } },
  data() { return { left: false, right: false, newLeft: false, newRight: false }; },
  computed: {
    arrow() { return 'position: relative; flex: none; width: 34px; align-self: stretch; min-height: 40px; padding: 0; border: 0; border-radius: 8px; background: var(--k-1F2C26,#1F2C26); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 20px; line-height: 1'; },
    mask() {
      if (!this.left && !this.right) return '';
      const m = `linear-gradient(to right, ${this.left ? 'transparent, #000 24px' : '#000'}, ${this.right ? '#000 calc(100% - 24px), transparent' : '#000'})`;
      return `mask-image: ${m}; -webkit-mask-image: ${m};`;
    }
  },
  watch: { active() { this.$nextTick(this.showActive); } },
  mounted() {
    this.ro = new ResizeObserver(() => this.measure());
    this.ro.observe(this.$refs.row);
    this.mo = new MutationObserver(() => this.measure());   // tabs and their "new" dots change with the data
    this.mo.observe(this.$refs.row, { childList: true, subtree: true });
    this.$nextTick(this.showActive);
  },
  beforeUnmount() { this.ro && this.ro.disconnect(); this.mo && this.mo.disconnect(); },
  methods: {
    measure() {
      const r = this.$refs.row;
      if (!r) return;
      this.left = r.scrollLeft > 2;
      this.right = r.scrollLeft + r.clientWidth < r.scrollWidth - 2;
      const box = r.getBoundingClientRect();
      const dots = [...r.querySelectorAll('.hub-blink')].map((d) => d.getBoundingClientRect());
      this.newLeft = dots.some((d) => d.right <= box.left + 4);
      this.newRight = dots.some((d) => d.left >= box.right - 4);
    },
    move(dir) { const r = this.$refs.row; r.scrollBy({ left: dir * Math.max(120, r.clientWidth * 0.7) }); },
    // a mouse wheel over the row moves it sideways, when there is somewhere to go
    wheel(e) {
      const r = this.$refs.row;
      if (r.scrollWidth <= r.clientWidth || Math.abs(e.deltaX) > Math.abs(e.deltaY)) return;
      e.preventDefault();
      r.scrollLeft += e.deltaY;
    },
    showActive() {
      const r = this.$refs.row;
      const on = r && r.querySelector('[aria-selected="true"]');
      if (on) on.scrollIntoView({ inline: 'nearest', block: 'nearest' });
      this.measure();
    }
  }
};
</script>

<style>
.hub-scrolltabs > button { flex: none; white-space: nowrap; }
</style>
