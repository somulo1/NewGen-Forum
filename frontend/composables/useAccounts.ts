// Demo accounts kept in the browser (no server yet): name, hashed password and three security questions with hashed answers.
const KEY = 'hub-accounts';

export const SECURITY_QUESTIONS = [
  'What was the name of your first pet?', 'In which town did you grow up?', 'What was the name of your first school?',
  "What is your mother's maiden name?", 'What was your first car?', 'What is the name of your favourite teacher?'
] as const;

const readAll = (): Record<string, any> => { try { return JSON.parse(localStorage.getItem(KEY) || '{}'); } catch (e) { return {}; } };
const writeAll = (o: Record<string, any>) => { localStorage.setItem(KEY, JSON.stringify(o)); };
const sha = async (s: string) => { const b = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(s)); return Array.from(new Uint8Array(b)).map((x) => x.toString(16).padStart(2, '0')).join(''); };
const norm = (a: string) => a.trim().toLowerCase().replace(/\s+/g, ' ');
const keyOf = (email: string) => email.trim().toLowerCase();

export const useAccounts = () => ({
  get: (email: string) => readAll()[keyOf(email)] || null,
  async create(name: string, email: string, password: string, quiz: { q: string; a: string }[]) {
    const all = readAll(); const k = keyOf(email);
    all[k] = { name, email: k, pass: await sha(k + ':' + password), quiz: await Promise.all(quiz.map(async (x) => ({ q: x.q, h: await sha(k + ':' + norm(x.a)) }))) };
    writeAll(all);
  },
  // Every answer has to match. Which one is wrong is never revealed.
  async checkQuiz(email: string, answers: string[]) {
    const acc = readAll()[keyOf(email)]; if (!acc) return false;
    const hs = await Promise.all(answers.map((a) => sha(keyOf(email) + ':' + norm(a))));
    return acc.quiz.length > 0 && acc.quiz.every((x: any, i: number) => x.h === hs[i]);
  },
  async verifyPassword(email: string, password: string) {
    const acc = readAll()[keyOf(email)];
    return !!acc && acc.pass === await sha(keyOf(email) + ':' + password);
  },
  async setPassword(email: string, password: string) {
    const all = readAll(); const k = keyOf(email);
    if (!all[k]) throw new Error('No local account exists for this email address.');
    all[k].pass = await sha(k + ':' + password); writeAll(all);
  },
  async resetPassword(email: string, password: string) {
    const all = readAll(); const k = keyOf(email);
    const acc = all[k] || { email: k, name: k.split('@')[0], quiz: [] };
    acc.pass = await sha(k + ':' + password);
    all[k] = acc;
    writeAll(all);
  },
  // A ready-made account so the reset flow can be tried without signing up first.
  async seedDemo() {
    const all = readAll(); if (all['sam@hub.example']) return;
    await this.create('Sam', 'sam@hub.example', 'Password1', [{ q: SECURITY_QUESTIONS[0], a: 'rex' }, { q: SECURITY_QUESTIONS[1], a: 'nairobi' }, { q: SECURITY_QUESTIONS[2], a: 'hub primary' }]);
  }
});
