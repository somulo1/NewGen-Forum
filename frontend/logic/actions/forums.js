import { chime, hangup, startRing, stopRing } from '~/logic/tones.js';
import * as callsW from '~/logic/server/calls.js';
import { MY_NAME, defaultPolicy, othersVoted, pollCap } from '~/logic/polls.js';
import * as forumsW from '~/logic/server/widgets/forums.js';

// HubLogic actions: forum polls, forum chat rules and private-chat privacy.
// Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const forumActions = {
  // Other members vote while the page is open, so results move without a refresh.
  startPollTimer() {
    clearInterval(this._pollT);
    this._pollT = setInterval(() => this.pollTick(), 4000);
  },
  pollTick() {
    if (this.srv) return;   // on the server the votes are real people's
    const open = this.state.polls.filter((p) => !p.locked && othersVoted(p) < pollCap(p));
    if (!open.length) return;
    const p = open[Math.floor(Math.random() * open.length)];
    const weights = p.options.map((o) => o.votes + 1);
    let r = Math.random() * weights.reduce((a, b) => a + b, 0), pick = 0;
    while (pick < weights.length - 1 && (r -= weights[pick]) > 0) pick++;
    const oid = p.options[pick].id;
    this.setState((st) => ({ polls: st.polls.map((x) => (x.id === p.id ? { ...x, options: x.options.map((o) => (o.id === oid ? { ...o, votes: o.votes + 1 } : o)) } : x)) }));
  },
  updPoll(pid, fn) {
    this.setState((st) => ({ polls: st.polls.map((p) => (p.id === pid ? fn(p) : p)) }));
  },
  // Polls in a forum on the server: the vote, lock, delete and new-poll steps go to it; the screens redraw from the reload.
  liveFail(e) { this.notify((e && e.message) || 'That did not go through.', '#FF9B8A'); },
  votePoll(pid, oid) {
    const live = this.srv && forumsW.findPoll(this, pid);
    if (live) {
      if (live.poll.locked) { this.notify('Voting is closed on this poll.', '#FFB86B'); return; }
      forumsW.votePoll(this, this.srv, pid, oid).catch((e) => this.liveFail(e));
      return;
    }
    const p = this.state.polls.find((x) => x.id === pid);
    if (!p) return;
    if (p.locked) { this.notify('Voting is closed on this poll.', '#FFB86B'); return; }
    if (p.blocked.indexOf(MY_NAME) !== -1) { this.notify('You have been blocked from voting on this poll.', '#FF9B8A'); return; }
    this.updPoll(pid, (x) => ({ ...x, myVote: x.myVote === oid ? null : oid }));
  },
  togglePollLock(pid) {
    const live = this.srv && forumsW.findPoll(this, pid);
    if (live) { forumsW.lockPoll(this, this.srv, pid).then(() => this.notify(live.poll.locked ? 'Voting is open again.' : 'Voting is now blocked on "' + live.poll.question + '".', '#7CC4FF')).catch((e) => this.liveFail(e)); return; }
    const p = this.state.polls.find((x) => x.id === pid);
    this.updPoll(pid, (x) => ({ ...x, locked: !x.locked }));
    this.notify(p && !p.locked ? 'Voting is now blocked on "' + p.question + '".' : 'Voting is open again.', '#7CC4FF');
  },
  deletePoll(pid) {
    if (this.srv && forumsW.findPoll(this, pid)) { this.setState({ pollMenu: null }); forumsW.deletePoll(this, this.srv, pid).catch((e) => this.liveFail(e)); return; }
    this.setState((st) => ({ polls: st.polls.filter((p) => p.id !== pid), pollMenu: null }));
  },
  toggleBlockVoter(pid, name) {
    this.updPoll(pid, (x) => ({ ...x, blocked: x.blocked.indexOf(name) !== -1 ? x.blocked.filter((n) => n !== name) : x.blocked.concat([name]) }));
  },
  createPoll() {
    const f = this.state.pollForm;
    if (!f) return;
    const options = f.options.map((o) => o.trim()).filter(Boolean);
    let err = '';
    if (!f.question.trim()) err = 'Write the question first.';
    else if (options.length < 2) err = 'Give at least two options.';
    else if (new Set(options.map((o) => o.toLowerCase())).size !== options.length) err = 'Each option must be different.';
    if (err) { this.setState({ pollForm: { ...f, err } }); return; }
    if (this.srv && forumsW.isLive(this.srv, f.forum)) {
      this.setState({ pollForm: null });
      forumsW.createPoll(this, this.srv, f.forum, f.question.trim(), options)
        .then(() => this.notify('Poll created. Members can vote now and the results update live.', '#B8F36B'))
        .catch((e) => { this.setState({ pollForm: f }); this.liveFail(e); });
      return;
    }
    const poll = { id: 'poll' + Date.now(), forum: f.forum, question: f.question.trim(), options: options.map((l, i) => ({ id: 'o' + (i + 1), label: l, votes: 0 })), by: 'You', at: Date.now(), locked: false, blocked: [], myVote: null, members: f.members };
    this.setState((st) => ({ polls: [poll].concat(st.polls), pollForm: null }));
    this.notify('Poll created. Members can vote now and the results update live.', '#B8F36B');
  },

  // ---- Forum chat rules ----
  setForumChat(fid, patch) {
    if (this.srv && forumsW.isLive(this.srv, fid)) { forumsW.saveRules(this, this.srv, fid, patch).catch((e) => this.liveFail(e)); return; }
    this.setState((st) => ({ forumChat: { ...st.forumChat, [fid]: { ...(st.forumChat[fid] || defaultPolicy()), ...patch } } }));
  },
  toggleChatRole(fid, role) {
    const f = this.srv && forumsW.liveForum(this, fid);
    const cur = f ? forumsW.policyOf(f) : (this.state.forumChat[fid] || defaultPolicy());
    this.setForumChat(fid, { roles: cur.roles.indexOf(role) !== -1 ? cur.roles.filter((r) => r !== role) : cur.roles.concat([role]) });
  },
  // Flips one person: someone who can chat is blocked; someone who cannot is let in (removed from the blocked list, or allowed by name).
  toggleMemberChat(fid, name, canNow) {
    const f = this.srv && forumsW.liveForum(this, fid);
    const cur = f ? forumsW.policyOf(f) : (this.state.forumChat[fid] || defaultPolicy());
    if (canNow) this.setForumChat(fid, { blocked: cur.blocked.concat([name]), allowed: cur.allowed.filter((n) => n !== name) });
    else if (cur.blocked.indexOf(name) !== -1) this.setForumChat(fid, { blocked: cur.blocked.filter((n) => n !== name) });
    else this.setForumChat(fid, { allowed: cur.allowed.concat([name]) });
  },

  // ---- Private chat privacy ----
  setDmPrivacy(name, patch) {
    this.setState((st) => ({ dmPrivacy: { ...st.dmPrivacy, [name]: { ...(st.dmPrivacy[name] || {}), ...patch } } }));
  },
  // Something arrives from this person. Your privacy settings decide whether you ever see it.
  simulateIncoming(name, kind) {
    const pr = this.state.dmPrivacy[name] || {};
    const what = kind === 'chat' ? 'message' : kind === 'call' ? 'call' : 'video call';
    const blocked = pr.blocked || (kind === 'chat' && pr.noChat) || (kind === 'call' && pr.noCall) || (kind === 'video' && (pr.noVideo || pr.noCall));
    if (blocked) { if (kind !== 'chat') this.logCall({ ctx: 'dm', ref: name, kind: kind === 'video' ? 'video' : 'voice', dir: 'blocked' }); this.notify(name + ' tried to ' + (kind === 'chat' ? 'message you' : kind === 'call' ? 'call you' : 'video call you') + ', but your privacy settings stopped it.', '#FFB86B'); return; }
    if (kind === 'chat') {
      this.setState((st) => {
        const open = st.panelKind === 'dm' && st.dms.find((d) => d.id === st.dmId && d.with === name) && st.mode !== 'closed' && st.mode !== 'min';
        const exists = st.dms.some((d) => d.with === name);
        const msg = { id: crypto.randomUUID(), from: name, text: 'Hi Sam, do you have a minute?', time: 'Just now' };
        const dms = exists ? st.dms.map((d) => (d.with === name ? { ...d, unread: open ? 0 : d.unread + 1, msgs: d.msgs.concat([msg]) } : d)) : [{ id: 'dm-' + Date.now(), with: name, unread: 1, msgs: [msg] }].concat(st.dms);
        return { dms };
      });
      chime();
      return;
    }
    if (this.state.call) { this.notify('Missed ' + what + ' from ' + name + ' while you were in a call.', '#FFB86B'); return; }
    this.setState({ incoming: { name, kind: kind === 'video' ? 'video' : 'voice' } });
    startRing();
    clearTimeout(this._incT);
    this._incT = setTimeout(() => { if (this.state.incoming && this.state.incoming.name === name) { stopRing(); this.logCall({ ctx: 'dm', ref: name, kind: kind === 'video' ? 'video' : 'voice', dir: 'missed' }); this.setState({ incoming: null }); this.notify('Missed ' + what + ' from ' + name + '.', '#FFB86B'); } }, 20000);
  },
  acceptIncoming() {
    const inc = this.state.incoming;
    if (!inc) return;
    clearTimeout(this._incT);
    stopRing();
    this.setState({ incoming: null });
    if (this.srv && inc.callId) {
      this._joinCallId = inc.callId;
      if (inc.context === 'dm') { const d0 = this.dir.find((x) => x.name === inc.name) || { initials: inc.name.split(' ').map((x) => x[0]).join(''), color: '#C9B6FF' }; this.startDmCall(inc.kind, inc.name, d0.initials, d0.color, 'in'); }
      else this.startCall(inc.kind, (inc.context === 'forum' ? 'Forum call' : 'Meeting') + ' with ' + inc.name, inc.name.split(' ').map((x) => x[0]).join('').slice(0, 2), '#7CC4FF', [], inc.context === 'forum' ? 'forum' : null, inc.context === 'forum' && this.srv.forumLocalId ? this.srv.forumLocalId.get(inc.conversation) : inc.conversation, 'in');
      return;
    }
    const d = this.dir.find((x) => x.name === inc.name) || { initials: inc.name.split(' ').map((x) => x[0]).join(''), color: '#C9B6FF' };
    this.startDmCall(inc.kind, inc.name, d.initials, d.color, 'in');
  },
  declineIncoming() {
    clearTimeout(this._incT);
    stopRing(); hangup();
    const inc = this.state.incoming;
    if (this.srv && inc && inc.callId) { if (inc.context === 'dm') callsW.decline(this.srv, inc.callId); this.setState({ incoming: null }); return; }
    if (inc) this.logCall({ ctx: 'dm', ref: inc.name, kind: inc.kind, dir: 'declined' });
    this.setState({ incoming: null });
  }
};
