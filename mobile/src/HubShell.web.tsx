// The app in a browser (Expo web). The phone build wraps the Hub site in a WebView and adds what a phone does (the device key,
// system notifications, the share sheet); a browser already does all of that itself, and the Hub site refuses to be shown inside
// another page (X-Frame-Options: DENY), so here the app simply opens the Hub site. Metro picks this file over HubShell.tsx for web,
// so none of the phone-only modules are bundled.
import React, { useEffect } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { COLORS, WEB_URL } from './config';

const target = WEB_URL ? WEB_URL + '/' : '';

export default function HubShell() {
  useEffect(() => {
    if (target && typeof window !== 'undefined') window.location.replace(target);
  }, []);

  return (
    <View style={s.fill}>
      <View style={s.logo}>
        <View style={[s.bar, { left: 18.7, top: 16.5, width: 8.8, height: 37.4 }]} />
        <View style={[s.bar, { left: 42.9, top: 16.5, width: 8.8, height: 37.4 }]} />
        <View style={[s.bar, { left: 18.7, top: 30.8, width: 33, height: 8.8 }]} />
      </View>
      {target ? (
        <>
          <Text style={s.title}>Opening Hub…</Text>
          <Pressable onPress={() => window.location.replace(target)} accessibilityRole="link" style={({ pressed }) => [s.button, pressed && { opacity: 0.8 }]}>
            <Text style={s.buttonText}>Open Hub</Text>
          </Pressable>
        </>
      ) : (
        <>
          <Text style={s.title}>Hub is not set up</Text>
          <Text style={s.detail}>This build has no Hub address. Set EXPO_PUBLIC_HUB_WEB_URL in mobile/.env to the address of your Hub site, then restart Expo.</Text>
        </>
      )}
    </View>
  );
}

const s = StyleSheet.create({
  fill: { flex: 1, minHeight: '100%', alignItems: 'center', justifyContent: 'center', padding: 28, gap: 14, backgroundColor: COLORS.bg },
  logo: { width: 70.4, height: 70.4, borderRadius: 16.5, backgroundColor: COLORS.accent, marginBottom: 8 },
  bar: { position: 'absolute', backgroundColor: COLORS.bg, borderRadius: 2.75 },
  title: { color: COLORS.ink, fontSize: 22, fontWeight: '700', textAlign: 'center' },
  detail: { color: COLORS.muted, fontSize: 15, lineHeight: 22, textAlign: 'center', width: '100%', maxWidth: 420 },
  button: { marginTop: 10, height: 44, paddingHorizontal: 24, borderRadius: 10, backgroundColor: COLORS.accent, alignItems: 'center', justifyContent: 'center' },
  buttonText: { color: COLORS.bg, fontSize: 15, fontWeight: '600' }
});
