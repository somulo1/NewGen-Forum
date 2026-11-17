import type { ConfigContext, ExpoConfig } from 'expo/config';

// iOS lets the in-app web view keep Hub's pages for offline use (service workers) only for the app's own domains, its
// "App-Bound Domains", which also stop the web view from navigating anywhere else. They are the two addresses in .env.
const host = (u?: string) => {
  const m = /^https:\/\/([^/?#:]+)/i.exec((u || '').trim());
  return m ? m[1].toLowerCase() : '';
};

export default ({ config }: ConfigContext): ExpoConfig => {
  const domains = [...new Set([host(process.env.EXPO_PUBLIC_HUB_WEB_URL), host(process.env.EXPO_PUBLIC_HUB_API_URL)].filter(Boolean))];
  return {
    ...(config as ExpoConfig),
    ios: { ...config.ios, infoPlist: { ...config.ios?.infoPlist, ...(domains.length ? { WKAppBoundDomains: domains } : {}) } }
  };
};
