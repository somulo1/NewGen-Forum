// Holds this device's keys and turns sealed server data into plain objects (and back). Transport-agnostic:
// `request(method, path, body)` returns parsed JSON (useBackend().request in the app, fetch in tests).
import { Device, unb64, b64, open, wrap, newConversationKey, newId, toText, fromText } from './e2ee.js';

export class Vault {
  constructor(request, { org, account, person, device, deviceId }) {
    this.request = request;
    this.org = org; this.account = account; this.person = person;
    this.device = device; this.deviceId = deviceId;
    this.keys = new Map();      // conversation id -> Map(epoch -> 32-byte key)
    this.signing = new Map();   // account id -> Map(device id -> signing key)
    this.convs = new Map();     // conversation id -> latest known conversation (epoch, members)
  }

  // One request per conversation at a time: opening a page of messages asks for the same keys hundreds of times at once.
  async epochKeys(conv, force = false) {
    let m = this.keys.get(conv);
    if (m && !force) return m;
    this.inflight = this.inflight || new Map();
    const k = 'keys:' + conv;
    if (this.inflight.has(k)) return this.inflight.get(k);
    const job = (async () => {
      const have = this.keys.get(conv) || new Map();
      const { keys } = await this.request('GET', `/conversations/${conv}/keys`);
      for (const kk of keys) {
        if (kk.device !== this.deviceId || have.has(kk.epoch)) continue;
        try { have.set(kk.epoch, await this.device.unwrap(unb64(kk.wrapped_key))); } catch (e) { /* not wrapped for this device */ }
      }
      this.keys.set(conv, have);
      return have;
    })().finally(() => this.inflight.delete(k));
    this.inflight.set(k, job);
    return job;
  }
  async keyFor(conv, epoch) {
    let m = await this.epochKeys(conv);
    if (!m.has(epoch)) m = await this.epochKeys(conv, true);
    return m.get(epoch) || null;
  }

  // Signing keys of other people's devices are fetched in batches: everyone asked for in the same tick goes in one request.
  batchSigning(personId) {
    this.signQ = this.signQ || { people: new Map(), timer: 0 };
    const q = this.signQ;
    return new Promise((resolve, reject) => {
      const w = q.people.get(personId) || [];
      w.push({ resolve, reject });
      q.people.set(personId, w);
      if (!q.timer) q.timer = setTimeout(() => this.flushSigning(), 0);
    });
  }
  async flushSigning() {
    const q = this.signQ;
    this.signQ = null;
    const ids = [...q.people.keys()];
    try {
      for (let i = 0; i < ids.length; i += 1000) {
        const { devices } = await this.request('POST', '/e2ee/people/devices', { people: ids.slice(i, i + 1000) });
        for (const d of devices) {
          const have = this.signing.get(d.account) || new Map();
          have.set(d.id, unb64(d.signing_key));
          this.signing.set(d.account, have);
        }
      }
      q.people.forEach((ws) => ws.forEach((w) => w.resolve()));
    } catch (e) { q.people.forEach((ws) => ws.forEach((w) => w.reject(e))); }
  }

  async signingKey(accountId, personId, deviceId) {
    let byDev = this.signing.get(accountId);
    if (!byDev || (deviceId && !byDev.has(deviceId))) {
      this.inflight = this.inflight || new Map();
      const k = 'sign:' + accountId + ':' + (byDev ? 'all' : 'first');
      if (!this.inflight.has(k)) {
        const job = (async () => {
          const have = this.signing.get(accountId) || new Map();
          if (accountId === this.account) {
            for (const d of (await this.request('GET', '/e2ee/devices')).devices) have.set(d.id, unb64(d.signing_key));
          } else if (personId) {
            // first look: the batch (active devices). A device that is not there (revoked since) is asked for by itself.
            if (!byDev) { await this.batchSigning(personId); for (const [id, key] of this.signing.get(accountId) || []) have.set(id, key); }
            if (!have.size || (deviceId && !have.has(deviceId))) {
              for (const d of (await this.request('GET', `/e2ee/people/${personId}/devices`)).devices) have.set(d.id, unb64(d.signing_key));
            }
          }
          this.signing.set(accountId, have);
          return have;
        })().finally(() => this.inflight.delete(k));
        this.inflight.set(k, job);
      }
      byDev = await this.inflight.get(k);
    }
    return deviceId ? byDev.get(deviceId) || [...byDev.values()][0] : [...byDev.values()][0];
  }

  // Opens a message or record; returns the parsed JSON payload, or null when it cannot be opened.
  async openItem({ conversation, id, sender, senderPerson, senderDevice, epoch, envelope }) {
    if (!envelope) return null;
    try {
      const key = await this.keyFor(conversation, epoch);
      const sk = await this.signingKey(sender, senderPerson, senderDevice);
      if (!key || !sk) return null;
      const pt = await open(key, unb64(envelope), sk, { org: this.org, conversation, sender, id, epoch });
      const text = toText(pt);
      try { return JSON.parse(text); } catch (e) { return { text }; }
    } catch (e) { return null; }
  }
  openMessage(m) {
    return this.openItem({ conversation: m.conversation, id: m.id, sender: m.sender_account, senderPerson: m.sender_person, senderDevice: m.sender_device, epoch: m.epoch, envelope: m.envelope });
  }
  openRecord(r, personOf) {
    return this.openItem({ conversation: r.container, id: r.id, sender: r.author_account, senderPerson: r.author_person || (personOf ? personOf(r.author_account) : null), senderDevice: r.author_device, epoch: r.epoch, envelope: r.envelope });
  }

  async conversation(id, force = false) {
    if (!force && this.convs.has(id)) return this.convs.get(id);
    const c = await this.request('GET', `/conversations/${id}`);
    this.convs.set(id, c);
    return c;
  }

  // Seals a payload for a conversation at its current epoch. Returns the fields the API wants.
  async seal(conv, payload, id = newId()) {
    const c = await this.conversation(conv, true);
    const key = await this.keyFor(conv, c.epoch);
    if (!key) throw new Error('This device does not have the key for this conversation yet.');
    const env = await this.device.seal(key, fromText(typeof payload === 'string' ? payload : JSON.stringify(payload)), { org: this.org, conversation: conv, sender: this.account, id, epoch: c.epoch });
    return { id, epoch: c.epoch, device: this.deviceId, envelope: b64(env) };
  }

  // Wraps a conversation key for every device of the given accounts that lacks one at this epoch (client-side key custody).
  async wrapFor(conv, epoch, key, deviceKeys) {
    return Promise.all(deviceKeys.map(async (d) => ({ epoch, device: d.id, wrapped_key: b64(await wrap(key, unb64(d.identity_key))) })));
  }

  // Runs the pending key work for this person's devices: new members/devices need a key, evictions need a new epoch.
  async custody() {
    const { todo } = await this.request('GET', '/e2ee/todo');
    let done = 0;
    for (const t of todo) {
      try {
        if (t.rotate) await this.rotate(t.conversation);
        else await this.fillMissing(t.conversation);
        done++;
      } catch (e) { /* another member's device may handle it */ }
    }
    return done;
  }
  async fillMissing(conv) {
    const c = await this.conversation(conv, true);
    const key = await this.keyFor(conv, c.epoch);
    if (!key) return;
    const { missing } = await this.request('GET', `/conversations/${conv}/missing-keys`);
    if (!missing || !missing.length) return;
    const wraps = await this.wrapFor(conv, c.epoch, key, missing.map((m) => ({ id: m.device || m.id, identity_key: m.identity_key })));
    await this.request('POST', `/conversations/${conv}/keys`, { device: this.deviceId, keys: wraps });
    this.keys.delete(conv);
  }
  async rotate(conv) {
    const c = await this.conversation(conv, true);
    const members = await this.request('GET', `/conversations/${conv}/missing-keys`).catch(() => null); // devices needing keys at next epoch are computed below
    void members;
    const next = c.epoch + 1;
    const key = newConversationKey();
    const devices = [];
    for (const m of c.members || []) {
      const r = await this.request('GET', `/e2ee/people/${m.person}/devices`);
      devices.push(...r.devices);
    }
    const wraps = await this.wrapFor(conv, next, key, devices);
    await this.request('POST', `/conversations/${conv}/rotate`, { device: this.deviceId, keys: wraps });
    this.keys.delete(conv); this.convs.delete(conv);
  }
}

export { newId, b64, unb64, Device, newConversationKey, wrap };
