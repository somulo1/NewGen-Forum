// Announcements widget: published announcements, who sent them, and the "Mark read" button.
export const name = 'announcements';

const ago = (iso) => {
  const m = Math.max(0, Math.round((Date.now() - new Date(iso).getTime()) / 60000));
  if (m < 1) return 'Just now';
  if (m < 60) return m + ' min ago';
  const h = Math.round(m / 60);
  if (h < 24) return h + ' h ago';
  const d = Math.round(h / 24);
  return d === 1 ? 'Yesterday' : d + ' days ago';
};

export async function load(logic, srv) {
  const { comms } = await srv.request('GET', '/comms?type=Announcement');
  const published = comms.filter((c) => c.status === 'Published').slice(0, 20);
  const items = [];
  await Promise.all(published.map(async (c) => {
    const sealed = await srv.request('GET', `/comms/${c.id}/content`);
    const body = await srv.vault.openItem({ conversation: sealed.conversation, id: sealed.message, sender: sealed.sender_account, senderPerson: sealed.sender_person, senderDevice: sealed.device, epoch: sealed.epoch, envelope: sealed.envelope });
    const by = srv.dirByUid && srv.dirByUid.get(c.by);
    items.push({ id: c.id, ref: c.ref, tag: by && by.dept ? by.dept.split(' ')[0] : 'News', title: body ? body.title : '(sealed)', body: body ? body.body || '' : '', meta: (c.by_name || '') + ' · ' + ago(c.published_at || c.created_at), read: !!c.i_read, at: c.published_at || c.created_at });
  }));
  items.sort((a, b) => (a.at < b.at ? 1 : -1));
  logic.setState({ announcementItems: items });
}

export async function markRead(logic, srv, id) {
  await srv.request('POST', `/comms/${id}/read`);
  logic.setState((st) => ({ announcementItems: st.announcementItems.map((a) => (a.id === id ? { ...a, read: true } : a)) }));
}
