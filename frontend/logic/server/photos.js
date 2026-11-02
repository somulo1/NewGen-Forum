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

export function photoUrl(srv, uid, version) {
  const key = uid + ':' + (version || 0);
  if (!urls.has(key)) {
    urls.set(key, srv.bytes('GET', '/people/' + uid + '/photo').then((b) => {
      const type = b ? typeOf(b) : '';
      return type ? URL.createObjectURL(new Blob([b], { type })) : '';
    }).catch(() => { setTimeout(() => urls.delete(key), 60000); return ''; }));   // a failed one is tried again in a minute
  }
  return urls.get(key);
}
