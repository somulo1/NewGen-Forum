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

  async epochKeys(conv, force = false) {
    let m = this.keys.get(conv);
    if (m && !force) return m;
    m = m || new Map();
    const { keys } = await this.request('GET', `/conversations/${conv}/keys`);
    for (const k of keys) {
      if (k.device !== this.deviceId || m.has(k.epoch)) continue;
      try { m.set(k.epoch, await this.device.unwrap(unb64(k.wrapped_key))); } catch (e) { /* not wrapped for this device */ }
    }
    this.keys.set(conv, m);
    return m;
  }
  async keyFor(conv, epoch) {
    let m = await this.epochKeys(conv);
    if (!m.has(epoch)) m = await this.epochKeys(conv, true);
    return m.get(epoch) || null;
  }

  async signingKey(accountId, personId, deviceId) {
    let byDev = this.signing.get(accountId);
    if (!byDev || (deviceId && !byDev.has(deviceId))) {
      byDev = byDev || new Map();
      if (accountId === this.account) {
        for (const d of (await this.request('GET', '/e2ee/devices')).devices) byDev.set(d.id, unb64(d.signing_key));
      } else if (personId) {
        for (const d of (await this.request('GET', `/e2ee/people/${personId}/devices`)).devices) byDev.set(d.id, unb64(d.signing_key));
      }
      this.signing.set(accountId, byDev);
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
    return this.openItem({ conversation: r.container, id: r.id, sender: r.author_account, senderPerson: personOf ? personOf(r.author_account) : null, senderDevice: r.author_device, epoch: r.epoch, envelope: r.envelope });
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
