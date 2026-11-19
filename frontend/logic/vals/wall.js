import { hashtagsOf } from '~/logic/richText.js';
// Values and handlers the Community wall reads through hubBindings: moments, polls, audience, hashtags, wishes.
import * as wallW from '~/logic/server/widgets/wall.js';
import * as wallPages from '~/logic/server/widgets/wallPages.js';

const TREND_DAYS = 7;
const TREND_MAX = 5;
export function valsWall(ctx) {
  const { wallBirthdays, moments, openEv, openPersonName, people, s, sharedViews, settingsPage, orgName, orgLogo, notesPage, noteWidget, posts } = ctx;
  const set = (o) => this.setState(o);
  const COLORS = ['#2B4A63', '#6B4A2B', '#5A2E47', '#3B5A2B', '#4A3A6B', '#6B2B2B'];
  const wp = s.wpoll;
  const setOpts = (fn) => this.setState((st) => ({ wpoll: { ...st.wpoll, opts: fn(st.wpoll.opts) } }));
  // Trending (and Popular, under the post box): the five hashtags most used on the wall in the last week, ties going to the newest.
  // A hashtag leaves when five others pass it or when its posts are more than a week old; the posts themselves stay on the wall.
  const weekAgo = Date.now() - TREND_DAYS * 24 * 3600 * 1000;
  const counts = new Map();
  (s.posts || []).forEach((p, i) => {
    if (!(Date.parse(p.at) >= weekAgo)) return;
    hashtagsOf(p.text).forEach((t) => { const k = t.toLowerCase(); const c = counts.get(k) || { tag: t, n: 0, first: i }; c.n++; c.first = Math.min(c.first, i); counts.set(k, c); });
  });
  const tags = [...counts.values()].sort((a, b) => b.n - a.n || a.first - b.first).slice(0, TREND_MAX).map((c) => c.tag);
  const m = s.momentOpen !== null && s.momentOpen !== undefined ? moments[s.momentOpen] : null;
  const mi = s.meInfo || {};
  const meAv = { name: mi.name || '', initials: mi.initials || ((mi.name || '').split(' ').map((x) => x[0]).join('').slice(0, 2).toUpperCase()), color: mi.color || '#C2185B', first: (mi.name || '').split(' ')[0] };
  return {
    meAv, greeting: 'Good ' + (new Date().getHours() < 12 ? 'morning' : new Date().getHours() < 17 ? 'afternoon' : 'evening') + (meAv.first ? ', ' + meAv.first : ''), sharedViews, settingsPage, orgName, orgLogo, notesPage, noteWidget,
    onlineList: people.map((p) => ({ ...p, open: () => openPersonName(p.name) })),
    trending: tags.map((tg) => ({ tag: tg, count: counts.get(tg.toLowerCase()).n, on: (s.wallTag || '').toLowerCase() === tg.toLowerCase(), pick: () => set({ wallTag: (s.wallTag || '').toLowerCase() === tg.toLowerCase() ? '' : tg }) })),
    noTrending: tags.length === 0,
    // one tap adds a popular hashtag to what you are writing
    addHashtag: (tg) => this.setState((st) => ({ draft: (st.draft && !/\s$/.test(st.draft) ? st.draft + ' ' : st.draft || '') + '#' + tg + ' ' })),
    wallFilter: { active: !!s.wallTag, tag: s.wallTag, clear: () => set({ wallTag: '' }) },
    // the wall a batch at a time: the posts revealed so far (all matching ones while a hashtag is picked), and the next batch
    wallPosts: (() => {
      if (s.wallTag || !this.srv) return posts || [];
      const keep = new Set(wallPages.visible(s).map((p) => p.id));
      return (posts || []).filter((p) => keep.has(p.id));
    })(),
    wallMore: (() => {
      const held = (s.posts || []).length, shown = wallPages.shownCount(s);
      const done = !!this.srv && shown >= held && this.srv.wallEnd === this.srv.wallId;
      return { loading: !!s.wallLoadingMore, done, active: !!this.srv && !s.wallTag && !done && held > 0,
        next: () => { if (this.srv) wallW.showMore(this, this.srv).catch(() => {}); } };
    })(),
    // Moments (stories): photo, video or text; like, love (double tap), and the owner sees who reacted.
    momentView: (() => {
      const ME = this.srv ? this.srv.me.name : 'Sam';   // the name a reaction of mine carries
      const react = (kind) => {
        if (!m) return;
        const other = kind === 'likes' ? 'loves' : null;
        this.setState((st) => ({ moments: st.moments.map((x, i) => (i !== s.momentOpen ? x : { ...x, [kind]: x[kind].indexOf(ME) !== -1 ? x[kind].filter((n) => n !== ME) : x[kind].concat([ME]) })) }));
        wallW.toggleReaction(this, this.srv, m.id, kind === 'likes' ? 'like' : 'love').catch((e) => this.notify(e.message || 'That did not work.', '#FF9B8A'));
      };
      const loveOn = () => { if (m && m.loves.indexOf(ME) === -1) react('loves'); };
      const who = m && m.mine ? m.likes.map((n) => ({ name: n, kind: 'Liked' })).concat(m.loves.map((n) => ({ name: n, kind: 'Loved' }))) : [];
      return { open: !!m, item: m || {}, close: () => set({ momentOpen: null, reactorsOpen: false }),
        next: () => set({ momentOpen: s.momentOpen + 1 < moments.length ? s.momentOpen + 1 : null, reactorsOpen: false }), prev: () => set({ momentOpen: s.momentOpen > 0 ? s.momentOpen - 1 : 0, reactorsOpen: false }),
        counter: m ? (s.momentOpen + 1) + ' / ' + moments.length : '', index: s.momentOpen === null || s.momentOpen === undefined ? -1 : s.momentOpen, total: moments.length,
        iLiked: !!m && m.likes.indexOf(ME) !== -1, iLoved: !!m && m.loves.indexOf(ME) !== -1,
        like: () => react('likes'), love: () => react('loves'), loveOn,
        mine: !!m && m.mine, who, hasWho: who.length > 0, whoOpen: !!s.reactorsOpen, toggleWho: () => set({ reactorsOpen: !s.reactorsOpen }) };
    })(),
    momentForm: { open: s.addMomentOpen, toggle: () => set({ addMomentOpen: !s.addMomentOpen }), colors: COLORS },
    addMoment: (caption, bg, file) => {
      const c = (caption || '').trim();
      if (!c && !file) return false;
      this.setState({ addMomentOpen: false });
      wallW.addMoment(this, this.srv, { caption: c.slice(0, 120), bg: bg || COLORS[0], file })
        .then(() => this.notify('Your moment is up.', '#B8F36B'))
        .catch((e) => this.notify(e.message || 'Could not add the moment.', '#FF9B8A'));
      return true;
    },
    // Poll builder in the composer
    wpollView: { on: !!wp, toggle: () => set({ wpoll: wp ? null : { opts: ['', ''] } }),
      opts: wp ? wp.opts.map((v, i) => ({ value: v, n: i + 1, onInput: (e) => setOpts((o) => o.map((x, k) => (k === i ? e.target.value : x))), canRemove: wp.opts.length > 2, remove: () => setOpts((o) => o.filter((_, k) => k !== i)) })) : [],
      canAdd: !!wp && wp.opts.length < 5, add: () => setOpts((o) => o.concat([''])) },
    openEventFromWall: () => openEv('Event'),
    // Audience
    wallAudience: s.wallAudience, wallAudiences: ['Everyone', 'My department', 'Management'],
    onAudience: (e) => set({ wallAudience: e.target.value }),
    // Birthdays
    wishBirthday: () => set({ draft: 'Happy birthday ' + (wallBirthdays.names.length ? wallBirthdays.names.join(', ') + ' and everyone' : 'to everyone') + ' celebrating today! 🎂 Wishing you a wonderful day.' }),
    wallBirthdays
  };
}
