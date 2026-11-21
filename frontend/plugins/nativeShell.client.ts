// Inside the phone app the native shell already keeps the page above the system buttons, so the page must not add that space again.
export default defineNuxtPlugin(() => {
  const native = (window as any).HubNative;
  if (native && native.app) document.documentElement.style.setProperty('--hub-safe-bottom', '0px');
});
