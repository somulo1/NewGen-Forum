// Demo sign-in: no backend yet, so anything signs you in. The "session" is one entry in the browser's storage.
const KEY = 'hub-auth';

export const useAuth = () => {
  const read = () => {
    try { const v = localStorage.getItem(KEY); return v ? JSON.parse(v) : null; } catch (e) { return null; }
  };
  return {
    user: read,
    signIn(name: string, email: string) {
      try { localStorage.setItem(KEY, JSON.stringify({ name: name || (email ? email.split('@')[0] : 'Sam'), email: email || '' })); } catch (e) { /* storage blocked: stay signed out */ }
    },
    // A full reload also clears everything the app holds in memory.
    signOut() {
      try { localStorage.removeItem(KEY); } catch (e) { /* nothing to clear */ }
      window.location.assign('/login');
    }
  };
};
