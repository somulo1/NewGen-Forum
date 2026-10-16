// Values and handlers the Community wall reads through hubBindings: moments, polls, audience, hashtags, wishes.
export function valsWall(ctx) {
  const { moments, openEv, openPersonName, people, s } = ctx;
  const set = (o) => this.setState(o);
  const COLORS = ['#2B4A63', '#6B4A2B', '#5A2E47', '#3B5A2B', '#4A3A6B', '#6B2B2B'];
  const wp = s.wpoll;
  const setOpts = (fn) => this.setState((st) => ({ wpoll: { ...st.wpoll, opts: fn(st.wpoll.opts) } }));
  const tags = ['CharityRun', 'NewJoiners', 'TournamentNight'];
  const m = s.momentOpen !== null && s.momentOpen !== undefined ? moments[s.momentOpen] : null;
  return {
    onlineList: people.map((p) => ({ ...p, open: () => openPersonName(p.name) })),
    trending: tags.map((tg) => ({ tag: tg, on: s.wallTag === tg, pick: () => set({ wallTag: s.wallTag === tg ? '' : tg }) })),
    wallFilter: { active: !!s.wallTag, tag: s.wallTag, clear: () => set({ wallTag: '' }) },
    // Moments (stories)
    momentView: { open: !!m, item: m || {}, close: () => set({ momentOpen: null }),
      next: () => set({ momentOpen: s.momentOpen + 1 < moments.length ? s.momentOpen + 1 : null }), prev: () => set({ momentOpen: s.momentOpen > 0 ? s.momentOpen - 1 : 0 }),
      counter: m ? (s.momentOpen + 1) + ' / ' + moments.length : '' },
    momentForm: { open: s.addMomentOpen, toggle: () => set({ addMomentOpen: !s.addMomentOpen }), colors: COLORS },
    addMoment: (caption, bg) => {
      const c = (caption || '').trim();
      if (!c) return false;
      set({ myMoments: [{ initials: 'S', color: '#F49AC1', bg: bg || COLORS[0], caption: c.slice(0, 60) }].concat(s.myMoments), addMomentOpen: false });
      this.notify('Your moment is up.', '#B8F36B');
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
    wishBirthday: () => set({ draft: 'Happy birthday Faith Achieng and everyone celebrating today! 🎂 Wishing you a wonderful day.' })
  };
}
