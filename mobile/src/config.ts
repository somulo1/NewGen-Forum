// Addresses come from mobile/.env (see .env.example). Both must be https in release builds: end-to-end encryption needs a
// secure page, and nothing the app sends may travel in the clear.
const clean = (v: string | undefined) => (v || '').trim().replace(/\/+$/, '');
const originOf = (u: string) => {
  const m = /^(https?:\/\/[^/?#]+)/i.exec(u);
  return m ? m[1].toLowerCase() : '';
};
export const isSecure = (u: string) => /^https:\/\//i.test(u);

// the Hub web interface the app shows
export const WEB_URL = clean(process.env.EXPO_PUBLIC_HUB_WEB_URL);
export const WEB_ORIGIN = originOf(WEB_URL);

// the Hub backend that web interface talks to
export const API_URL = clean(process.env.EXPO_PUBLIC_HUB_API_URL);

// The app's colours: the same as the web's dark theme (the web reports its live colour once a page is up).
export const COLORS = { bg: '#0F1714', ink: '#E8EFEA', muted: '#9DB0A6', accent: '#B8F36B', line: '#22302A', card: '#16201C' };
