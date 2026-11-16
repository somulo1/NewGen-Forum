<!-- A project's GitHub link (optional): the repository, the webhook to paste into GitHub, an optional access token for private
     repositories, and whether GitHub may move tasks by itself or only suggest moves for the team to accept. -->
<template>
<section aria-label="GitHub" class="pg-card">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 8px"><strong style="flex: 1 1 200px; font-size: 14px">GitHub <span class="pg-mut">· optional</span></strong>
<span v-if="g && g.linked" class="pg-pill">Linked to {{ g.repo }}</span></div>
<span class="pg-mut">Link a repository and Hub follows its branches and pull requests. A branch named after a task (its stored branch, or one containing its key such as {{ example }}) is matched to that task: a branch or push suggests "in progress", an opened pull request "being checked", a merge "finished".</span>
<template v-if="g">
<template v-if="g.can_edit">
<label class="pg-lab">Repository<input type="text" v-model="repo" placeholder="owner/name, or its github.com address" class="pg-in"></label>
<label class="pg-lab">Access token <span class="pg-mut">(optional: only to list a private repository's branches; kept encrypted)</span><input type="password" v-model="token" :placeholder="g.has_token ? 'A token is kept. Type a new one to replace it.' : 'A fine-grained token with read access to the repository'" class="pg-in" autocomplete="off"></label>
<label style="display: flex; gap: 10px; align-items: flex-start; font-size: 13px; cursor: pointer"><input type="checkbox" v-model="auto" style="margin-top: 3px"><span>Move tasks automatically <span class="pg-mut">· otherwise each move waits for someone on the team to check it and accept it. Only forward moves the workflow allows.</span></span></label>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<button type="button" class="pg-btn pg-pri" :disabled="busy" @click="save(false)">{{ g.linked ? 'Save' : 'Link repository' }}</button>
<button v-if="g.linked" type="button" class="pg-btn" :disabled="busy" @click="save(true)">Make a new secret</button>
<button v-if="g.has_token" type="button" class="pg-btn" :disabled="busy" @click="dropToken">Remove token</button>
<button v-if="g.linked" type="button" class="pg-btn pg-danger" :disabled="busy" @click="unlink">Unlink</button>
</div>
</template>
<span v-else-if="!g.linked" class="pg-mut">Not linked. The project owner, or someone who manages project teams, can link it.</span>
<div v-if="g.linked" class="pg-box">
<strong style="font-size: 13px">In GitHub: the repository's Settings → Webhooks → Add webhook</strong>
<div class="pg-row"><span class="pg-k">Payload URL</span><code class="pg-code">{{ payloadURL }}</code><button type="button" class="pg-btn pg-sm" @click="copy(payloadURL)">Copy</button></div>
<span v-if="localOnly" role="alert" style="font-size: 12px; color: var(--k-FFB86B,#FFB86B)">This address is on this computer, so GitHub cannot reach it. Use the Hub that runs online (its address is set by NUXT_PUBLIC_API_BASE).</span>
<div class="pg-row"><span class="pg-k">Content type</span><code class="pg-code">application/json</code></div>
<div class="pg-row"><span class="pg-k">Secret</span><template v-if="secret"><code class="pg-code">{{ secret }}</code><button type="button" class="pg-btn pg-sm" @click="copy(secret)">Copy</button></template><span v-else class="pg-mut">Shown once, when it was made. Lost it? Make a new secret and paste it into GitHub.</span></div>
<div class="pg-row"><span class="pg-k">SSL verification</span><span>Enable</span></div>
<div class="pg-row"><span class="pg-k">Events</span><span>Let me select individual events: Branch or tag creation, Branch or tag deletion, Pushes, Pull requests. Keep it Active.</span></div>
<span v-if="secret" role="status" style="font-size: 12px; color: var(--k-FFB86B,#FFB86B)">Copy the secret now: Hub will not show it again.</span>
</div>
<div v-if="g.linked" style="display: flex; flex-direction: column; gap: 6px">
<strong style="font-size: 13px">Latest from GitHub</strong>
<span v-if="!g.events.length" class="pg-mut">Nothing yet. Once GitHub sends its first event it shows here.</span>
<div v-for="e in g.events.slice(0, 15)" :key="e.id" class="pg-ev"><span class="pg-k">{{ label(e.kind) }}</span><code class="pg-code">{{ e.branch }}</code><span style="flex: 1 1 160px; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ e.title }}</span><span class="pg-mut">{{ e.actor }} · {{ when(e.at) }}</span><span class="pg-mut">{{ e.task ? stateLabel(e.state) : 'no task' }}</span></div>
</div>
</template>
<span v-if="err" role="alert" style="font-size: 13px; color: #FF9B8A">{{ err }}</span>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

const LABELS = { branch_created: 'Branch', branch_deleted: 'Branch deleted', push: 'Push', pr_opened: 'Pull request', pr_draft: 'Draft PR', pr_merged: 'Merged', pr_closed: 'PR closed' };
export default {
  name: 'ProjectGitHub',
  props: { project: { type: String, required: true }, code: { type: String, default: '' } },
  emits: ['changed'],
  data() { return { g: null, repo: '', token: '', auto: false, secret: '', busy: false, err: '' }; },
  computed: {
    srv() { return useHubStore().logic.srv; },
    payloadURL() { const base = ((useRuntimeConfig().public.apiBase || '') + '').replace(/\s+/g, '').replace(/\/+$/, '') || (typeof location !== 'undefined' ? location.origin : ''); return this.g ? base + this.g.hook : ''; },
    localOnly() { return /^https?:\/\/(localhost|127\.|0\.0\.0\.0|\[::1\])/i.test(this.payloadURL); },
    example() { return (this.code || 'CODE') + '-12'; }
  },
  watch: { project: { immediate: true, handler() { this.load(); } } },
  methods: {
    label(k) { return LABELS[k] || k; },
    stateLabel(s) { return { pending: 'waiting for someone to accept', accepted: 'accepted', dismissed: 'dismissed', auto: 'moved by GitHub', none: 'recorded' }[s] || s; },
    when(at) { return new Date(at).toLocaleString('en-GB', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' }); },
    async load() {
      if (!this.srv || !this.project) return;
      try { this.g = await this.srv.request('GET', `/projects/${this.project}/github`); this.repo = this.g.repo || ''; this.auto = !!this.g.auto; this.err = ''; } catch (e) { this.err = (e && e.message) || 'Could not read the GitHub link.'; }
    },
    async save(newSecret, token) {
      this.busy = true; this.err = '';
      try {
        const body = { repo: this.repo, auto: this.auto, new_secret: !!newSecret };
        if (token !== undefined) body.token = token; else if (this.token.trim()) body.token = this.token.trim();
        const g = await this.srv.request('PUT', `/projects/${this.project}/github`, body);
        if (g.secret) this.secret = g.secret;
        this.g = g; this.token = '';
        this.$emit('changed', g.linked ? g.repo : '');
      } catch (e) { this.err = (e && e.message) || 'Could not save.'; }
      this.busy = false;
    },
    dropToken() { return this.save(false, ''); },
    async unlink() {
      if (!window.confirm('Unlink GitHub? Deliveries to its webhook are refused from now on.')) return;
      this.busy = true;
      try { await this.srv.request('DELETE', `/projects/${this.project}/github`); this.secret = ''; await this.load(); this.$emit('changed', ''); } catch (e) { this.err = (e && e.message) || 'Could not unlink.'; }
      this.busy = false;
    },
    copy(v) { try { navigator.clipboard.writeText(v); } catch (e) { /* select it by hand */ } }
  }
};
</script>
