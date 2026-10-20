// Signed-out visitors go to the login page; signed-in users have no reason to see it.
const OPEN = ['/login', '/signup', '/reset-password'];
// Legal documents are readable by anyone, signed in or not.
const PUBLIC = ['/legal', '/terms', '/privacy', '/cookies', '/acceptable-use', '/data-processing', '/security-overview'];

export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return;
  const signedIn = !!useAuth().user();
  if (PUBLIC.indexOf(to.path) !== -1) return;
  if (!signedIn && OPEN.indexOf(to.path) === -1) return navigateTo(to.fullPath === '/' ? '/login' : { path: '/login', query: { next: to.fullPath } });
  // a temporary password must be replaced before the app can be used
  const mustChange = signedIn && !!useBackend().session()?.mustChange;
  if (mustChange && to.path !== '/change-password') return navigateTo('/change-password');
  if (!signedIn && to.path === '/change-password') return navigateTo('/login');
  if (signedIn && !mustChange && (OPEN.indexOf(to.path) !== -1 || to.path === '/change-password')) return navigateTo('/');
});
