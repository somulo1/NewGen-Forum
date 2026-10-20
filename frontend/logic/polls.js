// Forum rules: who may chat, and how polls count votes.
export const MY_NAME = 'Samwel Okoth Omulo';
export const FORUM_ROLES = ['Admin', 'Moderator', 'Member', 'Guest'];
// The signed-in user's role inside each forum. Forums not listed here make them a Member.
export const MY_FORUM_ROLE = { risk: 'Admin' };

export const defaultPolicy = () => ({ mode: 'everyone', roles: ['Admin', 'Moderator'], blocked: [], allowed: [] });

// mode "everyone": anyone, except people on the blocked list.
// mode "roles": only the chosen roles, plus people on the allowed list, minus people on the blocked list.
// mode "locked": only admins, plus the allowed list, minus the blocked list.
export function chatAllowed(policy, name, role) {
  if (policy.blocked.indexOf(name) !== -1) return { allowed: false, reason: 'You have been blocked from chatting in this forum.' };
  if (policy.allowed.indexOf(name) !== -1) return { allowed: true, reason: '' };
  if (policy.mode === 'roles' && policy.roles.indexOf(role) === -1) return { allowed: false, reason: 'Chat in this forum is limited to ' + (policy.roles.length ? policy.roles.join(', ') : 'nobody') + '.' };
  if (policy.mode === 'locked' && role !== 'Admin') return { allowed: false, reason: 'Chat is locked in this forum. Only admins can post.' };
  return { allowed: true, reason: '' };
}

export function policySummary(policy) {
  const extras = (policy.blocked.length ? ' · ' + policy.blocked.length + ' blocked' : '') + (policy.allowed.length ? ' · ' + policy.allowed.length + ' allowed by name' : '');
  if (policy.mode === 'roles') return 'Only ' + (policy.roles.length ? policy.roles.join(', ') : 'nobody') + ' can chat' + extras;
  if (policy.mode === 'locked') return 'Chat is locked. Only admins can post' + extras;
  return 'Everyone can chat' + extras;
}

export const pollCount = (poll, option) => option.votes + (poll.myVote === option.id ? 1 : 0);
export const pollTotal = (poll) => poll.options.reduce((n, o) => n + pollCount(poll, o), 0);
// How many other members may still vote. Blocked voters and the signed-in user are left out, and a few never vote.
export const pollCap = (poll) => Math.max(0, Math.floor((poll.members - 1 - poll.blocked.length) * 0.8));
export const othersVoted = (poll) => poll.options.reduce((n, o) => n + o.votes, 0);

// "5 min ago", "3 h ago", "2 d ago" for a time in milliseconds.
export function agoLabel(at, nowMs) {
  const m = Math.max(0, Math.round((nowMs - at) / 60000));
  if (m < 1) return 'Just now';
  if (m < 60) return m + ' min ago';
  const hr = Math.round(m / 60);
  return hr < 24 ? hr + ' h ago' : Math.round(hr / 24) + ' d ago';
}
