// Performance awards: the award strip on Home and the "Give an award" dialog. Awards follow approval flow f7 on the server.
export const name = 'awards';

const mapAward = (a) => ({
  id: a.ref, uid: a.id, type: a.type, people: a.names || [], citation: a.citation || '', date: a.date,
  status: a.status === 'Pending approval' ? 'Pending' : a.status, by: a.by
});

export async function load(logic, srv) {
  const { awards } = await srv.request('GET', '/awards');
  logic.setState({ awardList: awards.map(mapAward) });
}

// a = { type, people (names), citation }
export async function give(logic, srv, a) {
  const people = a.people.map((n) => (n === 'Sam' || n === srv.me.name ? srv.me.person : srv.personOf.get(n))).filter(Boolean);
  const res = await srv.request('POST', '/awards', { type: a.type, people, citation: a.citation });
  await load(logic, srv);
  return res;
}
