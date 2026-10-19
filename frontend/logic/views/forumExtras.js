import { FORUM_ROLES, MY_FORUM_ROLE, MY_NAME, agoLabel, chatAllowed, defaultPolicy, policySummary, pollCount, pollTotal } from '~/logic/polls.js';
import { richTextHtml } from '~/logic/richText.js';
import * as forumsW from '~/logic/server/widgets/forums.js';

// renderVals step: forumExtras. Polls, forum chat rules and private-chat privacy for whatever is open in the chat panel.
export function buildForumExtras(ctx) {
  const { can, cur, curDm, isDM, people: sampleMembers, s, set } = ctx;
  const nowMs = Date.now();
  const fid = cur.id;
  // A forum on the server has real members, posting rules (enforced by the server on every send) and polls; otherwise the samples.
  const liveF = !curDm && cur.live ? cur : null;
  const people = liveF ? liveF.memberList.map((m) => ({ name: m.name, initials: m.initials, color: m.color, role: m.title || m.role, forumRole: m.forumRole, person: m.person, roleKey: m.roleKey })) : sampleMembers;
  const myRole = liveF ? liveF.myForumRole : (MY_FORUM_ROLE[fid] || 'Member');
  const policy = liveF ? forumsW.policyOf(liveF) : (s.forumChat[fid] || defaultPolicy());
  const gateOf = (m) => (liveF ? forumsW.mayPost(liveF.rules, m.person, m.roleKey) : chatAllowed(policy, m.name, m.forumRole));
  const priv = curDm ? (s.dmPrivacy[curDm.with] || {}) : {};

  // postMsg() reads this, so a blocked chat can never send, whichever composer is used.
  const chatGate = isDM
    ? (priv.blocked ? { allowed: false, reason: 'You blocked ' + curDm.with + '. Unblock them to send messages.' } : { allowed: true, reason: '' })
    : liveF ? forumsW.mayPost(liveF.rules, this.srv.me.person, this.srv.me.role) : chatAllowed(policy, MY_NAME, myRole);
  this._chatGate = chatGate;

  const canManage = !isDM && (myRole === 'Admin' || myRole === 'Moderator' || can('forums', 'Moderate'));

  // ---- Chat rules for the open forum ----
  const chatCtl = { canManage, mode: policy.mode, summary: policySummary(policy), youLine: chatGate.allowed ? 'You can chat in this forum.' : chatGate.reason,
    modeOpts: [{ v: 'everyone', l: 'Everyone can chat' }, { v: 'roles', l: 'Only selected roles can chat' }, { v: 'locked', l: 'Locked: only admins can chat' }],
    setMode: (e) => this.setForumChat(fid, { mode: e.target.value }), isRoles: policy.mode === 'roles',
    roles: FORUM_ROLES.map((r) => ({ name: r, on: policy.roles.indexOf(r) !== -1, toggle: () => this.toggleChatRole(fid, r) })),
    members: people.map((m) => {
      const g = gateOf(m);
      const listed = policy.blocked.indexOf(m.name) !== -1;
      return { name: m.name, initials: m.initials, color: m.color, forumRole: m.forumRole, can: g.allowed,
        stateLabel: g.allowed ? (policy.allowed.indexOf(m.name) !== -1 ? 'Allowed by name' : 'Can chat') : listed ? 'Blocked' : 'Not allowed',
        stateColor: g.allowed ? '#B8F36B' : '#FF9B8A', actionLabel: g.allowed ? 'Block chat' : listed ? 'Unblock' : 'Allow chat',
        toggle: () => this.toggleMemberChat(fid, m.name, g.allowed) };
    }) };

  // ---- Polls in the open forum ----
  const forumPolls = isDM ? [] : (liveF ? liveF.polls : s.polls.filter((p) => p.forum === fid)).slice().sort((a, b) => b.at - a.at);
  const pf = s.pollForm;
  const setOpt = (i) => (e) => { const v = typeof e === 'string' ? e : e.target.value; this.setState((st) => ({ pollForm: { ...st.pollForm, err: '', options: st.pollForm.options.map((o, j) => (j === i ? v : o)) } })); };
  const pollView = { count: forumPolls.length, noPolls: forumPolls.length === 0, canCreate: chatGate.allowed, createHint: chatGate.allowed ? '' : chatGate.reason,
    openForm: () => this.setState({ pollForm: { forum: fid, question: '', options: ['', ''], err: '', members: cur.members } }),
    // The poll button next to attach and emoji: jump to the Polls view with the new-poll form open.
    compose: () => this.setState({ tab: 'polls', panelTab: 'polls', pollForm: { forum: fid, question: '', options: ['', ''], err: '', members: cur.members } }),
    form: pf ? { open: true, question: pf.question, err: pf.err, hasErr: !!pf.err, canAdd: pf.options.length < 6,
      onQuestion: (e) => { const v = typeof e === 'string' ? e : e.target.value; this.setState((st) => ({ pollForm: { ...st.pollForm, question: v, err: '' } })); },
      options: pf.options.map((v, i) => ({ value: v, html: richTextHtml(v), placeholder: 'Option ' + (i + 1), onInput: setOpt(i), onValue: (value) => this.setState((st) => ({ pollForm: { ...st.pollForm, err: '', options: st.pollForm.options.map((option, index) => index === i ? value : option) } })), canRemove: pf.options.length > 2,
        remove: () => this.setState((st) => ({ pollForm: { ...st.pollForm, options: st.pollForm.options.filter((x, j) => j !== i) } })) })),
      addOption: () => this.setState((st) => ({ pollForm: { ...st.pollForm, options: st.pollForm.options.concat(['']) } })),
      submit: () => this.createPoll(), cancel: set({ pollForm: null }) } : { open: false },
    polls: forumPolls.map((p) => {
      const total = pollTotal(p);
      const blockedMe = p.blocked.indexOf(MY_NAME) !== -1;
      const open = !p.locked && !blockedMe;
      return { id: p.id, question: p.question, questionHtml: richTextHtml(p.question), by: p.by, when: agoLabel(p.at, nowMs), locked: p.locked,
        statusLabel: p.locked ? 'Voting blocked' : 'Live', statusColor: p.locked ? '#FFB86B' : '#B8F36B',
        totalLabel: total + (total === 1 ? ' vote' : ' votes') + ' · ' + p.members + ' members',
        note: blockedMe ? 'You are blocked from voting on this poll.' : p.locked ? 'Voting is blocked. These results are final.' : p.myVote ? 'You voted. Tap it again to take it back, or tap another option to change.' : 'Tap an option to vote. Results update live.',
        options: p.options.map((o) => {
          const c = pollCount(p, o), pct = total ? Math.round((c / total) * 100) : 0, mine = p.myVote === o.id;
          return { id: o.id, label: o.label, labelHtml: richTextHtml(o.label), count: c, pct, mine, canVote: open,
            rowStyle: 'position:relative;display:flex;flex-direction:column;gap:6px;width:100%;padding:10px 12px;border-radius:10px;text-align:left;font-family:inherit;color:var(--k-E8EFEA,#E8EFEA);background:var(--k-121B17,#121B17);border:1px solid ' + (mine ? '#B8F36B' : '#24322C') + ';cursor:' + (open ? 'pointer' : 'default') + ';opacity:' + (open || mine ? '1' : '0.85'),
            barStyle: 'height:100%;width:' + pct + '%;background:' + (mine ? '#B8F36B' : 'var(--k-3F524A,#3F524A)') + ';border-radius:3px;transition:width .6s ease',
            vote: () => this.votePoll(p.id, o.id) };
        }),
        canControl: canManage || p.by === 'You', lockLabel: p.locked ? 'Allow voting' : 'Block voting', toggleLock: () => this.togglePollLock(p.id), remove: () => this.deletePoll(p.id),
        blockedCount: p.blocked.length, menuOpen: s.pollMenu === p.id, toggleMenu: () => this.setState((st) => ({ pollMenu: st.pollMenu === p.id ? null : p.id })),
        voters: p.live ? [] : people.map((m) => { const b = p.blocked.indexOf(m.name) !== -1; return { name: m.name, initials: m.initials, color: m.color, blocked: b, label: b ? 'Unblock' : 'Block from voting', toggle: () => this.toggleBlockVoter(p.id, m.name) }; }) };
    }) };

  // ---- Private chat privacy ----
  let dmPriv = { open: false };
  if (curDm) {
    const name = curDm.with;
    const flag = (k, label, desc) => ({ key: k, label, desc, on: !!priv[k] || (k !== 'blocked' && !!priv.blocked), locked: k !== 'blocked' && !!priv.blocked,
      toggle: () => this.setDmPrivacy(name, { [k]: !priv[k] }) });
    dmPriv = { open: s.privacyOpen, toggleOpen: () => this.setState((st) => ({ privacyOpen: !st.privacyOpen })), name, blocked: !!priv.blocked,
      anyOn: !!(priv.blocked || priv.noChat || priv.noCall || priv.noVideo),
      summary: priv.blocked ? 'Blocked. Nothing gets through in either direction.' : [priv.noChat && 'chats', priv.noCall && 'calls', priv.noVideo && 'video'].filter(Boolean).length ? 'Not receiving ' + [priv.noChat && 'chats', priv.noCall && 'calls', priv.noVideo && 'video'].filter(Boolean).join(', ') + ' from ' + name : 'You receive everything from ' + name,
      flags: [flag('blocked', 'Block ' + name, 'No messages, calls or video, and you cannot message or call them.'), flag('noChat', 'Do not receive their chats', 'Their messages are not delivered to you.'), flag('noCall', 'Do not receive their calls', 'Their voice calls will not ring.'), flag('noVideo', 'Do not receive their video', 'Their video calls and video requests are stopped.')],
      tries: [{ label: 'Try a message', run: () => this.simulateIncoming(name, 'chat') }, { label: 'Try a call', run: () => this.simulateIncoming(name, 'call') }, { label: 'Try a video call', run: () => this.simulateIncoming(name, 'video') }] };
  }
  const inc = s.incoming;
  const incomingView = inc ? { name: inc.name, kindLabel: inc.kind === 'video' ? 'Video call' : 'Voice call', accept: () => this.acceptIncoming(), decline: () => this.declineIncoming() } : null;

  return { chatCtl, chatGate, dmPriv, incomingView, pollView };
}
