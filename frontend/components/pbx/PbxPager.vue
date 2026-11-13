<!-- Page buttons under a Calls management table: Previous, "11–20 of 54", Next. Hidden when everything fits on one page. -->
<template>
<nav v-if="total > size" class="pbx-pager" aria-label="Pages">
  <span class="pbx-sub">{{ from }}–{{ to }} of {{ total }}</span>
  <button type="button" class="pbx-btn small ghost" :disabled="page <= 1" @click="$emit('update:page', page - 1)">Previous</button>
  <span class="pbx-sub">Page {{ page }} of {{ pages }}</span>
  <button type="button" class="pbx-btn small ghost" :disabled="page >= pages" @click="$emit('update:page', page + 1)">Next</button>
</nav>
</template>

<script>
export default {
  name: 'PbxPager',
  props: { total: { type: Number, required: true }, page: { type: Number, default: 1 }, size: { type: Number, default: 10 } },
  emits: ['update:page'],
  computed: {
    pages() { return Math.max(1, Math.ceil(this.total / this.size)); },
    from() { return this.total ? (this.page - 1) * this.size + 1 : 0; },
    to() { return Math.min(this.total, this.page * this.size); }
  },
  // a filter that shrinks the list never leaves you on an empty page
  watch: { pages(n) { if (this.page > n) this.$emit('update:page', n); } }
};
</script>

<style>
.pbx-pager { display: flex; align-items: center; justify-content: flex-end; flex-wrap: wrap; gap: 8px; padding-top: 4px; }
</style>
