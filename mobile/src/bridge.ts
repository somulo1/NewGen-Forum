// The script that runs in the web interface before its own code, inside the app only. It adds `window.HubNative` (the phone's
// key, notifications, files) and fills in what a page inside an app does not get from a browser: system notifications, file
// downloads, links that leave Hub, the clipboard, and the colour of the bars around the page. The interface itself is untouched.
export function bridgeScript(o: { platform: string; version: string; permission: string }): string {
  const cfg = JSON.stringify(o);
  return `(function () {
  if (window.HubNative) return;
  var CFG = ${cfg};
  var seq = 0, waits = {};
  function post(m) { try { window.ReactNativeWebView.postMessage(JSON.stringify(m)); } catch (e) {} }
  function call(type, data) {
    return new Promise(function (res, rej) {
      var id = ++seq; waits[id] = { res: res, rej: rej };
      var m = { id: id, type: type }; for (var k in (data || {})) m[k] = data[k];
      post(m);
      setTimeout(function () { if (waits[id]) { delete waits[id]; rej(new Error('The app did not answer.')); } }, 20000);
    });
  }
  window.__hubNativeReply = function (id, ok, v) { var w = waits[id]; if (!w) return; delete waits[id]; ok ? w.res(v) : w.rej(new Error(v || 'Refused')); };

  window.HubNative = Object.freeze({
    app: true, platform: CFG.platform, version: CFG.version,
    deviceProof: function (kind, subject) { return call('proof', { kind: String(kind), subject: String(subject) }); },
    share: function (name, mime, base64) { post({ type: 'file', name: name, mime: mime, data: base64 }); }
  });

  // ---- system notifications: the browser Notification API, answered by the phone
  var notes = {};
  function HubNotification(title, opts) {
    opts = opts || {};
    this.title = String(title || 'Hub'); this.body = String(opts.body || ''); this.tag = String(opts.tag || '');
    this.onclick = null; this.onclose = null; this._id = 'n' + (++seq);
    notes[this._id] = this;
    post({ type: 'notify', nid: this._id, title: this.title, body: this.body.slice(0, 300), tag: this.tag });
  }
  HubNotification.prototype.close = function () { post({ type: 'notifyClose', nid: this._id }); delete notes[this._id]; };
  HubNotification.prototype.addEventListener = function (t, f) { if (t === 'click') this.onclick = f; };
  HubNotification.prototype.removeEventListener = function () {};
  HubNotification.permission = CFG.permission;
  HubNotification.requestPermission = function (cb) {
    return call('notifyPermission').then(function (p) { HubNotification.permission = p; if (cb) cb(p); return p; });
  };
  try { Object.defineProperty(window, 'Notification', { value: HubNotification, writable: true, configurable: true }); } catch (e) { window.Notification = HubNotification; }
  window.__hubNotifyTap = function (nid) { var n = notes[nid]; if (n && n.onclick) { try { n.onclick({ target: n, preventDefault: function () {} }); } catch (e) {} } };

  // ---- files: blob and data links (downloads, exports, attachments) go to the phone's open/save/send sheet
  function sendBlob(href, name) {
    return fetch(href).then(function (r) { return r.blob(); }).then(function (b) {
      return new Promise(function (res) {
        var fr = new FileReader();
        fr.onload = function () { var s = String(fr.result || ''); var i = s.indexOf(','); post({ type: 'file', name: name || '', mime: b.type || 'application/octet-stream', data: s.slice(i + 1) }); res(); };
        fr.onerror = function () { res(); };
        fr.readAsDataURL(b);
      });
    }).catch(function () {});
  }
  function abs(u) { try { return new URL(String(u || ''), location.href).href; } catch (e) { return String(u || ''); } }
  function isFile(h) { return /^(blob:|data:)/i.test(h); }
  function isHere(h) { return h.indexOf(location.origin + '/') === 0 || h === location.origin; }
  function external(h) { post({ type: 'external', url: h }); }

  var aClick = HTMLAnchorElement.prototype.click;
  HTMLAnchorElement.prototype.click = function () {
    var h = abs(this.getAttribute('href'));
    if (isFile(h)) { sendBlob(h, this.getAttribute('download') || ''); return; }
    if (h && !isHere(h) && /^(https?:|mailto:|tel:|sms:)/i.test(h)) { external(h); return; }
    return aClick.call(this);
  };
  document.addEventListener('click', function (e) {
    var a = e.target && e.target.closest ? e.target.closest('a[href]') : null;
    if (!a || e.defaultPrevented) return;
    var h = abs(a.getAttribute('href'));
    if (isFile(h)) { e.preventDefault(); sendBlob(h, a.getAttribute('download') || ''); }
    // ordinary links that leave Hub are sent to the phone's browser by the app itself (onShouldStartLoadWithRequest)
  });
  var fakeWin = function () { return { closed: false, opener: null, close: function () {}, focus: function () {}, blur: function () {}, postMessage: function () {}, document: null, location: {} }; };
  window.open = function (u) {
    var h = abs(u);
    if (!u || h === 'about:blank') return fakeWin();
    if (isFile(h)) { sendBlob(h, ''); return fakeWin(); }
    if (isHere(h)) { location.assign(h); return fakeWin(); }
    external(h);
    return fakeWin();
  };

  // ---- clipboard: always through the phone, which works without the page being "focused"
  try {
    var clip = navigator.clipboard || {};
    var writeText = function (t) { post({ type: 'copy', text: String(t == null ? '' : t) }); return Promise.resolve(); };
    if (navigator.clipboard) navigator.clipboard.writeText = writeText;
    else Object.defineProperty(navigator, 'clipboard', { value: { writeText: writeText, readText: function () { return Promise.resolve(''); } }, configurable: true });
  } catch (e) {}

  // ---- the colour behind the status bar and the home indicator follows the page (dark, light, the sign-in pages)
  var lastBg = '', tq = 0;
  function theme() {
    var el = document.querySelector('.hub-shell') || document.body;
    if (!el) return;
    var c = getComputedStyle(el).backgroundColor;
    if (!c || c === 'rgba(0, 0, 0, 0)' || c === 'transparent') c = getComputedStyle(document.documentElement).backgroundColor;
    if (c && c !== lastBg && c !== 'rgba(0, 0, 0, 0)') { lastBg = c; post({ type: 'theme', bg: c }); }
  }
  function soon() { clearTimeout(tq); tq = setTimeout(theme, 120); }
  var watch = function () {
    try { new MutationObserver(soon).observe(document.documentElement, { subtree: true, childList: true, attributes: true, attributeFilter: ['class'] }); } catch (e) {}
    soon();
  };
  if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', watch); else watch();
  ['pushState', 'replaceState'].forEach(function (k) { var f = history[k]; history[k] = function () { var r = f.apply(this, arguments); soon(); return r; }; });
  window.addEventListener('popstate', soon);
  window.addEventListener('load', function () { post({ type: 'ready' }); soon(); });
})(); true;`;
}
