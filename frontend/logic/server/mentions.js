// Tagging someone (@Name) in a ticket comment, a wall post or a forum message: the message is encrypted, so the server is only
// told who was tagged and where. It sends each of them a bell notification and an email.
export function names(srv, text) {
  const dir = [...((srv.dirByUid && srv.dirByUid.values()) || [])];
  return dir.filter((d) => d.name && text.indexOf('@' + d.name) !== -1);
}

export function tell(srv, where, ref, people) {
  const ids = [...new Set(people.map((p) => p.uid).filter(Boolean))].slice(0, 20);
  if (!ids.length) return Promise.resolve();
  return srv.request('POST', '/mentions', { where, ref: String(ref).slice(0, 80), people: ids }).catch(() => {});
}
