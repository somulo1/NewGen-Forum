<!-- My storage: how much of my space I use and how much is left. -->
<template>
<section v-if="ready" aria-label="My storage" style="padding: 20px 24px; border-radius: 22px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-24322C,#24322C); display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; align-items: baseline; gap: 8px">
<h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 18px; flex: 1">My storage</h2>
<span style="font-size: 14px; color: var(--k-C9D6CF,#C9D6CF)"><strong>{{ size(used) }}</strong> used {{ quota ? 'of ' + size(quota) : '(no limit)' }}</span>
</div>
<template v-if="quota">
<div role="progressbar" :aria-valuenow="pct" aria-valuemin="0" aria-valuemax="100" :aria-label="`${pct}% of your storage used`" style="height: 10px; border-radius: 5px; background: var(--k-2A3832,#2A3832); overflow: hidden">
<div :style="`height: 100%; width: ${Math.max(pct, used ? 2 : 0)}%; background: ${color}; border-radius: 5px; transition: width .3s`"></div>
</div>
<span :style="`font-size: 13px; color: ${color}; line-height: 1.5`">{{ advice }}</span>
</template>
<template v-if="err"><div role="alert" style="padding: 10px 14px; border-radius: 12px; background: var(--k-3A2420,#3A2420); color: #FFD3CA; font-size: 13px">{{ err }}</div></template>
</section>
</template>

<script>
import { useHubStore } from '~/composables/useHub';

const KIND = { folder: 'Files', forum: 'Forum', dm: 'Private chat', wall: 'Wall', ticket: 'Ticket', project: 'Project', event: 'Calendar' };

export default {
  name: 'StorageCard',
  data() { return { ready: false, used: 0, quota: 0, files: [], err: '', asking: '', busy: false }; },
  computed: {
    pct() { return this.quota ? Math.min(100, Math.round((this.used / this.quota) * 100)) : 0; },
    color() { return this.pct >= 90 ? '#FF9B8A' : this.pct >= 75 ? '#FFB86B' : '#B8F36B'; },
    advice() {
      const left = Math.max(0, this.quota - this.used);
      if (this.pct >= 100) return 'Your storage is full. New uploads are stopped until you delete files below, or an administrator gives you more space.';
      if (this.pct >= 90) return 'Almost full: ' + this.size(left) + ' left. Delete the files below that you no longer need.';
      if (this.pct >= 75) return this.size(left) + ' left. Consider deleting large files you no longer need.';
      return this.size(left) + ' free.';
    }
  },
  mounted() { this.load(); },
  methods: {
    srv() { return useHubStore().logic.srv; },
    size(b) { return b >= 1073741824 ? (b / 1073741824).toFixed(1) + ' GB' : b >= 1048576 ? (b / 1048576).toFixed(1) + ' MB' : Math.max(1, Math.round(b / 1024)) + ' KB'; },
    where(f) {
      const st = useHubStore().logic.state;
      if (f.kind === 'forum') { const fo = (st.forumsLive || []).find((x) => x.sid === f.folder); if (fo) return 'Forum · ' + fo.name; }
      if (f.kind === 'dm') { const d = (st.dms || []).find((x) => x.id === f.folder); if (d) return 'Chat with ' + d.with; }
      return KIND[f.kind] || 'Files';
    },
    async load() {
      const srv = this.srv();
      if (!srv) { setTimeout(() => this.load(), 1000); return; }   // still signing in
      try {
        const [u, l] = await Promise.all([srv.request('GET', '/storage/usage'), srv.request('GET', '/storage/files?limit=20')]);
        this.used = u.used_bytes; this.quota = u.quota_mb * 1048576;
        // the file names are sealed; this device opens them
        this.files = await Promise.all(l.files.map(async (f) => {
          const pl = await srv.vault.openItem({ conversation: f.folder, id: f.id, sender: f.author_account, senderPerson: srv.me.person, senderDevice: f.author_device, epoch: f.epoch, envelope: f.envelope }).catch(() => null);
          return { id: f.id, size: f.size, name: (pl && pl.name) || 'File', where: this.where(f), when: new Date(f.created_at).toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' }) };
        }));
        this.ready = true; this.err = '';
      } catch (e) { this.err = (e && e.message) || 'Could not read your storage.'; this.ready = true; }
    },
    async remove(f) {
      this.busy = true;
      try { await this.srv().request('DELETE', '/files/' + f.id); this.asking = ''; await this.load(); } catch (e) { this.err = (e && e.message) || 'Could not delete the file.'; }
      this.busy = false;
    }
  }
};
</script>
