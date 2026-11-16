// Other people's profile pictures. The server keeps each picture for its owner and hands it to colleagues at GET /people/{id}/photo; the
// directory says who has one (hasPhoto) and a version number that changes when it does. A picture is downloaded once per version and
// shown from a blob address, so a page with fifty posts by ten people makes ten downloads, not fifty.
const urls = new Map();   // "uid:version" -> Promise<string | ''>

const typeOf = (b) => {
  if (b[0] === 0x89 && b[1] === 0x50) return 'image/png';
  if (b[0] === 0xFF && b[1] === 0xD8) return 'image/jpeg';
  if (b[0] === 0x47 && b[1] === 0x49) return 'image/gif';
  if (b[0] === 0x52 && b[1] === 0x49 && b[8] === 0x57) return 'image/webp';
  return '';
};

// Pictures asked for at about the same moment go to the server together (one request for up to 40 people), not one each.
let queue = new Map();   // uid -> [{ key, resolve }]
let timer = null;
function flush(srv) {
  timer = null;
  const batch = queue; queue = new Map();
  const ids = [...batch.keys()];
  for (let i = 0; i < ids.length; i += 40) {
    const part = ids.slice(i, i + 40);
    srv.request('GET', '/people/photos?ids=' + part.join(',')).then(({ photos }) => {
      part.forEach((uid) => {
        const p = photos && photos[uid];
        let url = '';
        if (p && p.data) {
          const bin = atob(p.data); const b = new Uint8Array(bin.length);
          for (let k = 0; k < bin.length; k++) b[k] = bin.charCodeAt(k);
          url = typeOf(b) ? URL.createObjectURL(new Blob([b], { type: typeOf(b) })) : '';
        }
        batch.get(uid).forEach((w) => w.resolve(url));
      });
    }).catch(() => part.forEach((uid) => batch.get(uid).forEach((w) => { w.resolve(''); setTimeout(() => urls.delete(w.key), 60000); })));
  }
}

export function photoUrl(srv, uid, version) {
  const key = uid + ':' + (version || 0);
  if (!urls.has(key)) {
    urls.set(key, new Promise((resolve) => {
      if (!queue.has(uid)) queue.set(uid, []);
      queue.get(uid).push({ key, resolve });
      if (!timer) timer = setTimeout(() => flush(srv), 30);
    }));
  }
  return urls.get(key);
}
