// Signed-out visitors go to the login page; signed-in users have no reason to see it.
const OPEN = ['/login', '/signup', '/reset-password'];
// Legal documents are readable by anyone, signed in or not.
const PUBLIC = ['/legal', '/terms', '/privacy', '/cookies', '/acceptable-use', '/data-processing', '/security-overview'];

export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return;
  const signedIn = !!useAuth().user();
  if (PUBLIC.indexOf(to.path) !== -1) return;
  if (!signedIn && OPEN.indexOf(to.path) === -1) return navigateTo(to.fullPath === '/' ? '/login' : { path: '/login', query: { next: to.fullPath } });
  if (signedIn && OPEN.indexOf(to.path) !== -1) return navigateTo('/');
});
