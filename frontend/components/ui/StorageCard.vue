<!-- My storage: how much of my space I use, and my biggest files so I can delete what I no longer need. -->
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
<template v-if="files.length">
<strong style="font-size: 14px">Your biggest files</strong>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6); margin-top: -8px">Deleting a file frees its space straight away. It is removed for everyone it was shared with.</span>
<div style="display: flex; flex-direction: column; gap: 6px">
<div v-for="f in files" :key="f.id" style="display: flex; flex-wrap: wrap; align-items: center; gap: 10px; padding: 10px 14px; border-radius: 12px; border: 1px solid var(--k-24322C,#24322C)">
<div style="flex: 1 1 220px; min-width: 0; display: flex; flex-direction: column">
<strong style="font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ f.name }}</strong>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">{{ f.where }} · {{ f.when }}</span>
</div>
<span style="font-size: 13px; font-weight: 600; min-width: 70px; text-align: right">{{ size(f.size) }}</span>
<template v-if="asking === f.id">
<button type="button" @click="remove(f)" :disabled="busy" style="height: 32px; padding: 0 14px; border-radius: 16px; border: 0; background: #FF9B8A; color: #0F1714; font-weight: 600; cursor: pointer">Delete</button>
<button type="button" @click="asking = ''" style="height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Keep</button>
</template>
<button v-else type="button" @click="asking = f.id" :aria-label="`Delete ${f.name}`" style="height: 32px; padding: 0 14px; border-radius: 16px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer">Delete…</button>
</div>
</div>
</template>
<span v-else style="font-size: 13px; color: var(--k-9DB0A6,#9DB0A6)">You have not uploaded any files yet.</span>
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
