// On screen, rights are called permissions. Text the server keeps (catalog labels, request titles) may still say "rights";
// it goes through here wherever it is shown.
const SAY = { right: 'permission', rights: 'permissions', Right: 'Permission', Rights: 'Permissions' };
export const sayPermissions = (t) => String(t == null ? '' : t).replace(/\b[Rr]ights?\b/g, (w) => SAY[w] || w);
