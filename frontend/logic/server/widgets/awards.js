// Performance awards: the award strip on Home and the "Give an award" dialog. Awards follow approval flow f7 on the server.
// An award is announced on the community wall and/or in forums the nominator picked. The posts are end-to-end encrypted, so
// the nominator's browser writes them: straight away when the award needs no approval, or as soon as it is approved.
import * as wallW from './wall.js';
import * as forumsW from './forums.js';
import { connectRealtime } from '../realtime.js';

export const name = 'awards';

const mapAward = (a) => ({
  id: a.ref, uid: a.id, type: a.type, people: a.names || [], citation: a.citation || '', date: a.date,
  status: a.status === 'Pending approval' ? 'Pending' : a.status, by: a.by, byPerson: a.by_person, targets: a.targets || []
});

export async function load(logic, srv) {
  const { awards } = await srv.request('GET', '/awards');
  const list = awards.map(mapAward);
  logic.setState({ awardList: list });
  if (!srv.awardsWatch) {
    srv.awardsWatch = true;
    const rt = srv.rt || (srv.rt = connectRealtime(srv));
    let t = 0;
    rt.on('approval.changed', () => { clearTimeout(t); t = setTimeout(() => load(logic, srv).catch(() => {}), 400); });   // my award may have just been approved
  }
  // announce my approved awards that still have places to go (an approval came in while I was away, or I just gave one)
  list.filter((a) => a.status === 'Approved' && a.targets.length && a.byPerson === srv.me.person).forEach((a) => announce(logic, srv, a).catch(() => {}));
}

const wordsFor = (a) => 'Congratulations ' + a.people.map((n) => '@' + n).join(', ') + ' on the ' + a.type + ' award!' + (a.citation ? '\n' + a.citation : '');

// Writes the award's post to each place that has not had it yet. Each place is claimed first, so it is posted once even if two
// of my tabs notice at the same moment.
export async function announce(logic, srv, a) {
  srv.announcing = srv.announcing || new Set();
  if (srv.announcing.has(a.uid)) return;
  srv.announcing.add(a.uid);
  try {
    for (let i = 0; i < 80 && !srv.bootDone; i++) await new Promise((r) => setTimeout(r, 250));   // the wall and forums must be loaded first
    const poster = srv.awardPoster && srv.awardPoster[a.uid];   // the picture picked in the dialog, while this tab remembers it
    const when = new Date().toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
    for (const target of a.targets) {
      try { await srv.request('POST', `/awards/${a.uid}/claim`, { target }); } catch (e) { continue; }   // someone else already did it
      try {
        if (target === 'wall') {
          await wallW.post(logic, srv, { text: wordsFor(a), files: poster ? [poster] : [], audience: 'Everyone', tags: a.people, poster: poster ? undefined : { award: a.type, names: a.people.join(', '), when } });
        } else {
          const local = srv.forumLocalId && srv.forumLocalId.get(target);
          if (local) await forumsW.send(logic, srv, local, { text: 'Award: ' + a.type + ' · ' + wordsFor(a).replace(/^Congratulations /, 'Congratulations '), files: poster ? [poster] : [] });
        }
      } catch (e) { logic.notify('The award post to ' + (target === 'wall' ? 'the wall' : 'a forum') + ' did not go through: ' + (e.message || 'try again'), '#FF9B8A'); }
    }
  } finally { srv.announcing.delete(a.uid); }
}

// a = { type, people (names), citation, wall, forums (local forum ids), poster (File) }
export async function give(logic, srv, a) {
  const people = a.people.map((n) => (n === 'Sam' || n === srv.me.name ? srv.me.person : srv.personOf.get(n))).filter(Boolean);
  const targets = (a.wall ? ['wall'] : []).concat((a.forums || []).map((f) => srv.forumConv && srv.forumConv[f]).filter(Boolean));
  const res = await srv.request('POST', '/awards', { type: a.type, people, citation: a.citation, targets });
  srv.awardPoster = srv.awardPoster || {};
  if (a.poster) srv.awardPoster[res.award.id] = a.poster;
  await load(logic, srv);   // an award that needs no approval is announced from here, at once
  return res;
}
