// Emoji library: the full Unicode set from the `emojibase-data` package, loaded on first use.
// Components never hold emoji of their own; they ask the picker, which reads from here.
let cache = null;

// Emojibase group numbers -> the tabs we show. Group 2 is skin-tone modifiers, which we leave out.
export const CATEGORIES = [
  { id: 0, label: 'Smileys', icon: '😀' }, { id: 1, label: 'People', icon: '👋' }, { id: 3, label: 'Animals & nature', icon: '🐶' },
  { id: 4, label: 'Food & drink', icon: '🍔' }, { id: 6, label: 'Activities', icon: '⚽' }, { id: 5, label: 'Travel & places', icon: '✈️' },
  { id: 7, label: 'Objects', icon: '💡' }, { id: 8, label: 'Symbols', icon: '❤️' }, { id: 9, label: 'Flags', icon: '🏁' }
];

export async function loadEmoji() {
  if (cache) return cache;
  const mod = await import('emojibase-data/en/data.json');
  const list = (mod.default || mod).filter((e) => e.group !== undefined && e.group !== 2 && e.emoji);
  list.sort((a, b) => (a.order || 0) - (b.order || 0));
  cache = { list, byGroup: CATEGORIES.reduce((m, c) => { m[c.id] = list.filter((e) => e.group === c.id); return m; }, {}) };
  return cache;
}

export function searchEmoji(data, q, limit = 120) {
  const s = q.trim().toLowerCase();
  if (!s) return [];
  return data.list.filter((e) => e.label.toLowerCase().indexOf(s) !== -1 || (e.tags || []).some((t) => t.toLowerCase().indexOf(s) !== -1)).slice(0, limit);
}

// Recently used, kept per browser.
const KEY = 'hub-emoji-recent';
export function recentEmoji() { try { return JSON.parse(localStorage.getItem(KEY) || '[]'); } catch (e) { return []; } }
export function rememberEmoji(ch) { try { localStorage.setItem(KEY, JSON.stringify([ch].concat(recentEmoji().filter((x) => x !== ch)).slice(0, 24))); } catch (e) { /* storage blocked: skip */ } }
