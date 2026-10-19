// My profile (and colleagues' full profiles): status, cover, photo, phone, bio and skills are saved with PATCH /me;
// name, title, department, email and join date come from the directory.
export const name = 'profile';

const FIELDS = ['status', 'cover', 'photo', 'phone', 'bio', 'skills'];

export async function load(logic, srv) {
  const [me, row] = await Promise.all([srv.request('GET', '/me'), srv.request('GET', '/people/' + srv.me.person)]);
  srv.me = { ...srv.me, ...me };
  const pf = me.profile || {};
  const profile = { status: pf.status || 'Available', cover: pf.cover || '#1F3A2B', photo: pf.photo || '', phone: pf.phone || '', bio: pf.bio || '', skills: pf.skills || [] };
  srv.profileSaved = profile;
  logic.setState({
    profile,
    meInfo: {
      name: row.name, initials: row.initials, color: row.color || '#C2185B', email: row.email || '', title: row.title || '', dept: row.department ? row.department.name : '',
      branch: row.branch ? row.branch.name : '', unit: row.unit ? row.unit.name : '', role: row.role_name || '', joined: row.joined_on || '', reportsTo: row.reports_to || null
    }
  });
}

// Saves whichever profile fields differ from what the server has.
export async function save(logic, srv) {
  const cur = logic.state.profile;
  const body = {};
  FIELDS.forEach((k) => { if (JSON.stringify(cur[k]) !== JSON.stringify(srv.profileSaved[k])) body[k] = cur[k]; });
  if (!Object.keys(body).length) return false;
  await srv.request('PATCH', '/me', body);
  srv.profileSaved = { ...cur };
  return true;
}

export function revert(logic, srv) {
  logic.setState({ profile: { ...srv.profileSaved } });
}

// The server keeps photos up to 512 KB: shrink anything bigger to a 512 px JPEG first.
export function photoDataUrl(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('The profile photo could not be read.'));
    reader.onload = () => {
      const url = reader.result;
      if (file.size <= 480 * 1024) { resolve(url); return; }
      const img = new Image();
      img.onerror = () => reject(new Error('The profile photo could not be read.'));
      img.onload = () => {
        const k = Math.min(1, 512 / Math.max(img.width, img.height));
        const c = document.createElement('canvas');
        c.width = Math.round(img.width * k); c.height = Math.round(img.height * k);
        c.getContext('2d').drawImage(img, 0, 0, c.width, c.height);
        resolve(c.toDataURL('image/jpeg', 0.85));
      };
      img.src = url;
    };
    reader.readAsDataURL(file);
  });
}

// A colleague's full profile (bio, skills, contact details), fetched once when it is opened.
export async function loadPerson(logic, srv, uid) {
  srv.people = srv.people || {};
  if (srv.people[uid]) return;
  srv.people[uid] = 'loading';
  try { srv.people[uid] = await srv.request('GET', '/people/' + uid); } catch (e) { srv.people[uid] = null; }
  logic.forceUpdate();
}
