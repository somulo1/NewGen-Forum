// Signed-out visitors go to the login page; signed-in users have no reason to see it.
const OPEN = ['/login', '/signup', '/reset-password'];

export default defineNuxtRouteMiddleware((to) => {
  if (import.meta.server) return;
  const signedIn = !!useAuth().user();
  if (!signedIn && OPEN.indexOf(to.path) === -1) return navigateTo(to.fullPath === '/' ? '/login' : { path: '/login', query: { next: to.fullPath } });
  if (signedIn && OPEN.indexOf(to.path) !== -1) return navigateTo('/');
});
