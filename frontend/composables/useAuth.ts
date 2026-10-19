// Signed-in state comes from the backend session (see useBackend). user() keeps its old shape for existing callers.
export const useAuth = () => {
  const be = useBackend();
  return {
    user: () => { const s = be.session(); return s ? { name: s.name, email: s.email } : null; },
    signIn: (email: string, password: string, totp = '') => be.login(email, password, totp),
    // A full reload also clears everything the app holds in memory.
    async signOut() {
      await be.logout();
      window.location.assign('/login');
    }
  };
};
