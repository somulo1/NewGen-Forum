// Where the Hub web interface lives. The app shows that interface (its mobile layout, exactly as on a phone browser) and talks to
// the same backend it uses. Set EXPO_PUBLIC_HUB_WEB_URL in mobile/.env (see .env.example). It must be https: end-to-end
// encryption uses the browser crypto API, which only works on secure pages.
const raw = (process.env.EXPO_PUBLIC_HUB_WEB_URL || '').trim().replace(/\/+$/, '');

export const WEB_URL = raw;
export const WEB_ORIGIN = (() => {
  const m = /^(https?:\/\/[^/?#]+)/i.exec(raw);
  return m ? m[1].toLowerCase() : '';
})();

// The app's colours: the same as the web's dark theme (the web reports its live colour once a page is up).
export const COLORS = { bg: '#0F1714', ink: '#E8EFEA', muted: '#9DB0A6', accent: '#B8F36B', line: '#22302A', card: '#16201C' };
