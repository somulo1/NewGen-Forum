<!-- A task's GitHub activity: its branch (pick one of the repository's), what GitHub reported for it, and moves it suggests. A
     suggestion is accepted by moving the task through the usual move form (with its comment), after checking it is right. -->
<template>
<section v-if="g && g.linked" aria-label="GitHub activity" style="display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px"><span style="flex: 1; font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: var(--k-9DB0A6,#9DB0A6)">GitHub · {{ g.repo }}</span>
<button v-if="canEdit" type="button" class="pg-btn pg-sm" :disabled="loadingB" @click="loadBranches">{{ loadingB ? 'Reading branches…' : 'Pick a branch' }}</button></div>
<div v-if="branches" style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center">
<select v-model="pick" aria-label="Branch from GitHub" class="pg-in" style="flex: 1 1 220px; height: 34px"><option value="">Choose one of {{ branches.length }} branches…</option><option v-for="b in branches" :key="b" :value="b">{{ b }}</option></select>
<button type="button" class="pg-btn pg-sm pg-pri" :disabled="!pick" @click="$emit('use-branch', pick); branches = null; pick = ''">Use for this task</button>
<button type="button" class="pg-btn pg-sm" @click="branches = null">Cancel</button></div>
<div v-for="e in pending" :key="e.id" style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px; padding: 10px 12px; border-radius: 12px; background: rgba(184,243,107,0.08); border: 1px solid var(--k-B8F36B,#B8F36B)">
<span style="flex: 1 1 220px; font-size: 13px">{{ label(e.kind) }} on <code class="pg-code">{{ e.branch }}</code> by {{ e.actor || 'someone' }}. Move it to <strong>{{ statusName(e.suggest) }}</strong>? Check it is right first.</span>
<a v-if="e.url" :href="e.url" target="_blank" rel="noopener noreferrer" style="font-size: 12px; color: var(--k-7CC4FF,#7CC4FF)">Open in GitHub</a>
<button type="button" class="pg-btn pg-sm pg-pri" @click="$emit('accept', { status: e.suggest, event: e.id, note: 'GitHub: ' + label(e.kind) + ' on ' + e.branch + (e.title ? ' (' + e.title + ')' : '') })">Accept and move</button>
<button type="button" class="pg-btn pg-sm" @click="dismiss(e)">Dismiss</button></div>
<div v-for="e in rest" :key="e.id" class="pg-ev"><span class="pg-k">{{ label(e.kind) }}</span><code class="pg-code">{{ e.branch }}</code><span style="flex: 1 1 120px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ e.title }}</span><a v-if="e.url" :href="e.url" target="_blank" rel="noopener noreferrer" style="font-size: 12px; color: var(--k-7CC4FF,#7CC4FF)">Open</a><span class="pg-mut">{{ state(e) }} · {{ when(e.at) }}</span></div>
<span v-if="!g.events.length" class="pg-mut">Nothing from GitHub for this task yet.</span>
<span v-if="err" role="alert" style="font-size: 12px; color: #FF9B8A">{{ err }}</span>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

const LABELS = { branch_created: 'Branch created', branch_deleted: 'Branch deleted', push: 'Push', pr_opened: 'Pull request opened', pr_draft: 'Draft pull request', pr_merged: 'Pull request merged', pr_closed: 'Pull request closed' };
export default {
  name: 'TaskGitHub',
  props: { project: { type: String, required: true }, task: { type: String, required: true }, statuses: { type: Array, default: () => [] }, canEdit: { type: Boolean, default: false }, version: { type: [String, Number], default: '' } },
  emits: ['accept', 'use-branch'],
  data() { return { g: null, branches: null, pick: '', loadingB: false, err: '' }; },
  computed: {
    srv() { return useHubStore().logic.srv; },
    pending() { return this.g ? this.g.events.filter((e) => e.state === 'pending') : []; },
    rest() { return this.g ? this.g.events.filter((e) => e.state !== 'pending').slice(0, 8) : []; }
  },
  watch: { task: { immediate: true, handler() { this.load(); } }, version() { this.load(); } },
  methods: {
    label(k) { return LABELS[k] || k; },
    statusName(k) { return (this.statuses.find((s) => s.v === k) || {}).l || k; },
    state(e) { return { pending: 'waiting', accepted: 'accepted', dismissed: 'dismissed', auto: 'moved to ' + this.statusName(e.suggest) + ' by GitHub', none: 'recorded' }[e.state] || e.state; },
    when(at) { return new Date(at).toLocaleString('en-GB', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }); },
    async load() {
      if (!this.srv || !this.project || !this.task || !/^[0-9a-f-]{36}$/.test(this.task)) { this.g = null; return; }
      try { this.g = await this.srv.request('GET', `/projects/${this.project}/github?task=${this.task}`); } catch (e) { this.g = null; }
    },
    async loadBranches() {
      this.loadingB = true; this.err = '';
      try { this.branches = (await this.srv.request('GET', `/projects/${this.project}/github/branches`)).branches; } catch (e) { this.err = (e && e.message) || 'Could not read the branches.'; }
      this.loadingB = false;
    },
    async dismiss(e) {
      try { await this.srv.request('POST', `/projects/${this.project}/github/events/${e.id}`, { decision: 'dismissed' }); await this.load(); } catch (x) { this.err = (x && x.message) || 'Could not dismiss it.'; }
    }
  }
};
</script>
