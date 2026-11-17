import type { ConfigContext, ExpoConfig } from 'expo/config';

const { existsSync } = require('fs') as { existsSync: (path: string) => boolean };

// iOS lets the in-app web view keep Hub's pages for offline use (service workers) only for the app's own domains, its
// "App-Bound Domains", which also stop the web view from navigating anywhere else. They are the two addresses in .env.
const host = (u?: string) => {
  const m = /^https:\/\/([^/?#:]+)/i.exec((u || '').trim());
  return m ? m[1].toLowerCase() : '';
};

// Push notifications need the app's Firebase file (Android) and Apple's production push entitlement (iOS). EAS hands the file
// over as a file secret (GOOGLE_SERVICES_JSON); a local build reads ./google-services.json. Without it the app builds and runs,
// only without pushes.
const googleServices = [process.env.GOOGLE_SERVICES_JSON, './google-services.json'].find((f) => !!f && existsSync(f));
const apsMode = process.env.HUB_APS_ENV === 'development' ? 'development' : 'production';

export default ({ config }: ConfigContext): ExpoConfig => {
  const domains = [...new Set([host(process.env.EXPO_PUBLIC_HUB_WEB_URL), host(process.env.EXPO_PUBLIC_HUB_API_URL)].filter(Boolean))];
  const plugins = (config.plugins || []).map((p) => (Array.isArray(p) && p[0] === 'expo-notifications' ? [p[0], { ...p[1], mode: apsMode }] : p)) as ExpoConfig['plugins'];
  return {
    ...(config as ExpoConfig),
    plugins,
    android: { ...config.android, ...(googleServices ? { googleServicesFile: googleServices } : {}) },
    ios: { ...config.ios, infoPlist: { ...config.ios?.infoPlist, ...(domains.length ? { WKAppBoundDomains: domains } : {}) } }
  };
};
