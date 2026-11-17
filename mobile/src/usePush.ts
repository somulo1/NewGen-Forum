// Pushes in the shell: a tap opens what it was about (kept until the page is ready to ask for it), and a new push address is
// handed to the page, which registers it with the backend for this sign-in.
import { useCallback, useEffect, useRef, type RefObject } from 'react';

import * as push from './push';

type Page = RefObject<{ injectJavaScript: (js: string) => void } | null>;

export function usePush(web: Page) {
  const tap = useRef<push.PushData | null>(null);
  const on = useRef(false);
  const giveAddress = useCallback((a: push.PushAddress | null) => {
    if (a) web.current?.injectJavaScript(`window.__hubPushAddress && window.__hubPushAddress(${JSON.stringify(a)}); true;`);
  }, [web]);
  useEffect(() => {
    const offTap = push.onTap((d) => { tap.current = d; web.current?.injectJavaScript(`window.__hubPushTap && window.__hubPushTap(); true;`); });
    const offNew = push.onNewAddress(giveAddress);
    return () => { offTap(); offNew(); };
  }, [giveAddress, web]);
  return { tap, on, giveAddress };
}
