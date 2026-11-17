// The app: the Hub web interface (its phone layout, exactly as the web shows it) running full screen, with the phone doing what a
// browser would: the key that binds the sign-in to this phone, notifications, files, links that leave Hub, the back button and
// the colour of the system bars. Everything comes from the Hub backend through that interface; the app keeps no data of its own.
import * as Application from 'expo-application';
import * as Clipboard from 'expo-clipboard';
import Constants, { ExecutionEnvironment } from 'expo-constants';
import * as SplashScreen from 'expo-splash-screen';
import { StatusBar } from 'expo-status-bar';
import * as SystemUI from 'expo-system-ui';
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { AppState, BackHandler, KeyboardAvoidingView, Linking, Platform, Pressable, StyleSheet, Text, View } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { WebView, type WebViewMessageEvent, type WebViewProps } from 'react-native-webview';
import type { ShouldStartLoadRequest, WebViewErrorEvent, WebViewHttpErrorEvent, WebViewNavigation } from 'react-native-webview/lib/WebViewTypes';

import { wakeBackend } from './backend';
import { bridgeScript } from './bridge';
import { COLORS, isSecure, WEB_ORIGIN, WEB_URL } from './config';
import { proof, ready } from './deviceKey';
import { saveAndShare } from './files';
import * as notify from './notify';
import * as push from './push';
import { Logo, parseColor } from './ui';

// react-native-webview's class typing does not line up with React 19's ref types; this names the same component with its props
type WebViewHandle = { goBack: () => void; reload: () => void; injectJavaScript: (js: string) => void };
const HubWebView = WebView as unknown as React.ComponentType<WebViewProps & { ref?: React.Ref<WebViewHandle> }>;

SplashScreen.preventAutoHideAsync().catch(() => {});
SplashScreen.setOptions({ fade: true, duration: 250 });

type Problem = { title: string; detail: string; retry: boolean } | null;

const here = (u: string) => !!WEB_ORIGIN && (u === WEB_ORIGIN || u.toLowerCase().startsWith(WEB_ORIGIN + '/'));
const EXTERNAL = /^(https?:|mailto:|tel:|sms:)/i;

export default function HubShell() {
  const web = useRef<WebViewHandle>(null);
  const canGoBack = useRef(false);
  const [bg, setBg] = useState({ hex: COLORS.bg, dark: true });
  const [boot, setBoot] = useState<{ permission: string } | null>(null);
  const [problem, setProblem] = useState<Problem>(null);
  const [round, setRound] = useState(0);   // bumped by Retry: a fresh page
  const asked = useRef(false);

  const version = Application.nativeApplicationVersion || Constants.expoConfig?.version || '1.0.0';

  // before the page loads: the phone's key must be usable (a sign-in here is bound to it), and the page is told whether
  // notifications are allowed
  useEffect(() => {
    let live = true;
    (async () => {
      if (!WEB_URL || (!isSecure(WEB_URL) && !__DEV__)) {
        setProblem({ title: 'Hub is not set up', detail: 'This build has no Hub address. Set EXPO_PUBLIC_HUB_WEB_URL to the https address of your Hub site and build again.', retry: false });
        SplashScreen.hideAsync().catch(() => {});
        return;
      }
      wakeBackend();
      if (!(await ready())) {
        setProblem({ title: 'This phone cannot keep Hub secure', detail: 'Hub binds your sign-in to this phone with a key kept in its secure storage, and that storage is not available. Set a screen lock on the phone, then open Hub again.', retry: true });
        SplashScreen.hideAsync().catch(() => {});
        return;
      }
      const permission = await notify.permission();
      if (live) setBoot({ permission });
    })();
    const t = setTimeout(() => SplashScreen.hideAsync().catch(() => {}), 12000);
    return () => { live = false; clearTimeout(t); };
  }, [round]);

  useEffect(() => { SystemUI.setBackgroundColorAsync(bg.hex).catch(() => {}); }, [bg.hex]);

  // Android back: back through the page's own history, then leave the app
  useEffect(() => {
    const sub = BackHandler.addEventListener('hardwareBackPress', () => {
      if (canGoBack.current && web.current) { web.current.goBack(); return true; }
      return false;
    });
    return () => sub.remove();
  }, []);

  // pushes: a tap opens what it was about (kept until the page is ready to ask), a new address is registered by the page
  const pushTap = useRef<push.PushData | null>(null);
  const pushOn = useRef(false);
  const giveAddress = useCallback((a: push.PushAddress | null) => { if (a) web.current?.injectJavaScript(`window.__hubPushAddress && window.__hubPushAddress(${JSON.stringify(a)}); true;`); }, []);
  useEffect(() => {
    const offTap = push.onTap((d) => { pushTap.current = d; web.current?.injectJavaScript(`window.__hubPushTap && window.__hubPushTap(); true;`); });
    const offNew = push.onNewAddress(giveAddress);
    return () => { offTap(); offNew(); };
  }, [giveAddress]);

  useEffect(() => notify.onTap((nid) => {
    web.current?.injectJavaScript(`window.__hubNotifyTap && window.__hubNotifyTap(${JSON.stringify(nid)}); true;`);
  }), []);

  const reply = useCallback((id: number, ok: boolean, value: unknown) => {
    web.current?.injectJavaScript(`window.__hubNativeReply && window.__hubNativeReply(${id}, ${ok ? 'true' : 'false'}, ${JSON.stringify(value ?? null)}); true;`);
  }, []);

  const onMessage = useCallback(async (e: WebViewMessageEvent) => {
    // only the Hub site may talk to the app
    if (!here(e.nativeEvent.url || '')) return;
    let m: any;
    try { m = JSON.parse(e.nativeEvent.data); } catch (x) { return; }
    if (!m || typeof m.type !== 'string') return;
    const id = Number(m.id) | 0;
    switch (m.type) {
      case 'proof':
        try { reply(id, true, await proof(String(m.kind), String(m.subject))); } catch (x) { reply(id, false, 'refused'); }
        break;
      case 'notifyPermission': {
        const p = await notify.ask();
        reply(id, true, p);
        if (p === 'granted') giveAddress(await push.address());
        break;
      }
      case 'pushAddress':
        reply(id, true, await push.address());
        break;
      case 'pushOpened':
        reply(id, true, pushTap.current); pushTap.current = null;
        break;
      case 'pushReady':
        pushOn.current = !!m.on;   // the backend now pushes to this phone: the page's own alerts would only repeat them
        break;
      case 'notify':
        // the page only announces what arrives while Hub is not in front; inside the app that means the app is in the background
        if (AppState.currentState !== 'active' && !pushOn.current) notify.show(String(m.nid), String(m.title || 'Hub'), String(m.body || ''), String(m.tag || '')).catch(() => {});
        break;
      case 'notifyClose':
        notify.close(String(m.nid));
        break;
      case 'file':
        if (typeof m.data === 'string' && m.data) saveAndShare(String(m.name || ''), String(m.mime || ''), m.data).catch(() => {});
        break;
      case 'external':
        if (EXTERNAL.test(String(m.url || ''))) Linking.openURL(String(m.url)).catch(() => {});
        break;
      case 'copy':
        Clipboard.setStringAsync(String(m.text ?? '')).catch(() => {});
        break;
      case 'theme': {
        const c = parseColor(String(m.bg || ''));
        if (c) setBg(c);
        break;
      }
      case 'ready':
        SplashScreen.hideAsync().catch(() => {});
        if (!asked.current && boot?.permission === 'default') {
          asked.current = true;
          const p = await notify.ask();
          web.current?.injectJavaScript(`try { window.Notification.permission = ${JSON.stringify(p)}; } catch (e) {} true;`);
          if (p === 'granted') giveAddress(await push.address());
        }
        break;
    }
  }, [reply, boot, giveAddress]);

  const onShouldStart = useCallback((req: ShouldStartLoadRequest) => {
    const u = req.url || '';
    if (here(u)) return true;
    if (/^(about:|blob:|data:)/i.test(u)) return true;
    if (Platform.OS === 'ios' && req.isTopFrame === false) return true;   // something embedded in a page stays in it
    if (EXTERNAL.test(u)) Linking.openURL(u).catch(() => {});
    return false;
  }, []);

  const injected = useMemo(() => (boot ? bridgeScript({ platform: Platform.OS, version, permission: boot.permission }) : ''), [boot, version]);

  const retry = () => { setProblem(null); setBoot(null); setRound((r) => r + 1); };

  if (problem) {
    return (
      <SafeAreaView style={[s.fill, { backgroundColor: COLORS.bg }]}>
        <StatusBar style="light" />
        <View style={s.center}>
          <Logo />
          <Text style={s.title}>{problem.title}</Text>
          <Text style={s.detail}>{problem.detail}</Text>
          {problem.retry ? (
            <Pressable onPress={retry} style={({ pressed }) => [s.button, pressed && { opacity: 0.8 }]} accessibilityRole="button">
              <Text style={s.buttonText}>Try again</Text>
            </Pressable>
          ) : null}
        </View>
      </SafeAreaView>
    );
  }

  return (
    <SafeAreaView style={[s.fill, { backgroundColor: bg.hex }]} edges={['top', 'bottom', 'left', 'right']}>
      <StatusBar style={bg.dark ? 'light' : 'dark'} />
      <KeyboardAvoidingView style={s.fill} behavior={Platform.OS === 'android' ? 'padding' : undefined} enabled={Platform.OS === 'android'}>
        {boot ? (
          <HubWebView
            key={round}
            ref={web}
            source={{ uri: WEB_URL + '/' }}
            style={[s.fill, { backgroundColor: bg.hex }]}
            containerStyle={{ backgroundColor: bg.hex }}
            originWhitelist={['*']}
            injectedJavaScriptBeforeContentLoaded={injected}
            injectedJavaScriptBeforeContentLoadedForMainFrameOnly
            onMessage={onMessage}
            onShouldStartLoadWithRequest={onShouldStart}
            onNavigationStateChange={(n: WebViewNavigation) => { canGoBack.current = n.canGoBack; }}
            onLoadEnd={() => SplashScreen.hideAsync().catch(() => {})}
            onError={(e: WebViewErrorEvent) => setProblem({ title: "Can't reach Hub", detail: 'Check your internet connection, then try again. (' + (e.nativeEvent.description || 'no connection') + ')', retry: true })}
            onHttpError={(e: WebViewHttpErrorEvent) => {
              if (e.nativeEvent.statusCode >= 500 && here(e.nativeEvent.url || '') && !/\.(js|css|png|svg|ico|json|woff2?)(\?|$)/i.test(e.nativeEvent.url || '')) {
                setProblem({ title: 'Hub is starting', detail: 'The Hub server is waking up or busy. Give it a moment, then try again.', retry: true });
              }
            }}
            onRenderProcessGone={() => setRound((r) => r + 1)}
            onContentProcessDidTerminate={() => web.current?.reload()}
            applicationNameForUserAgent={'HubApp/' + version}
            javaScriptEnabled
            domStorageEnabled
            cacheEnabled
            allowFileAccess={false}
            allowsInlineMediaPlayback
            mediaPlaybackRequiresUserAction={false}
            mediaCapturePermissionGrantType="grant"
            allowsBackForwardNavigationGestures
            setSupportMultipleWindows={false}
            // iOS keeps pages offline only on the app's own domains (app.config.ts); Expo Go has none of its own
            limitsNavigationsToAppBoundDomains={Platform.OS === 'ios' && Constants.executionEnvironment !== ExecutionEnvironment.StoreClient}
            javaScriptCanOpenWindowsAutomatically={false}
            textZoom={100}
            overScrollMode="never"
            bounces={false}
            keyboardDisplayRequiresUserAction={false}
            contentInsetAdjustmentBehavior="never"
            automaticallyAdjustContentInsets={false}
            webviewDebuggingEnabled={__DEV__}
          />
        ) : null}
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}

const s = StyleSheet.create({
  fill: { flex: 1 },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 28, gap: 14 },
  title: { color: COLORS.ink, fontSize: 22, fontWeight: '700', textAlign: 'center' },
  detail: { color: COLORS.muted, fontSize: 15, lineHeight: 22, textAlign: 'center', maxWidth: 420 },
  button: { marginTop: 10, height: 44, paddingHorizontal: 24, borderRadius: 10, backgroundColor: COLORS.accent, alignItems: 'center', justifyContent: 'center' },
  buttonText: { color: COLORS.bg, fontSize: 15, fontWeight: '600' }
});
