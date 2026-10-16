// Any error Vue catches while rendering or running a handler shows the 500 page instead of a blank screen.
export default defineNuxtPlugin((nuxtApp) => {
  const fatal = useState<{ message: string } | null>('hubFatal', () => null);
  nuxtApp.hook('vue:error', (err: any) => {
    console.error(err);
    fatal.value = { message: String(err && err.message ? err.message : err) };
  });
});
