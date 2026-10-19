// Sample data derived on start-up: logs. Called with the HubLogic instance as `this`.
export function seedLogs() {
  // Audit log sample (minutes ago)
  const users = ['Brian Kiprop', 'Grace Njeri', 'David Otieno', 'Amina Wanjiru', 'Faith Achieng', 'Kevin Mutua', 'Mercy Wambui', 'Samwel Okoth Omulo'];
  const normal = [['Signed in', 'Auth'], ['Viewed an announcement', 'Announcements'], ['Updated ticket IT-208', 'Tickets'], ['Uploaded "Q3 budget tracker.xlsx"', 'Files'], ['Posted on the wall', 'Wall'], ['Started a video meeting', 'Meetings'], ['Created a secure link', 'Secure share'], ['Moved a task to Review', 'Projects'], ['Downloaded "Staff handbook.pdf"', 'Files'], ['Signed out', 'Auth']];
  const logs = [];
  for (let i = 0; i < 70; i++) {
    const n = normal[(i * 3) % normal.length];
    logs.push({ id: 'n' + i, ago: 4 + i * 97, user: users[(i * 5) % users.length], action: n[0], module: n[1], ip: '10.0.' + ((i * 7) % 12) + '.' + (20 + (i * 11) % 200), loc: 'Office network', result: 'Success', rule: null });
  }
  for (let k = 0; k < 6; k++) logs.push({ id: 'f' + k, ago: 12 + k, user: 'Peter Kamau', action: 'Failed sign-in (wrong password)', module: 'Auth', ip: '41.90.12.7', loc: 'Nairobi, KE', result: 'Failed', rule: 'r1' });
  logs.push({ id: 'g1', ago: 38, user: 'Diana Rotich', action: 'Signed in', module: 'Auth', ip: '102.89.3.44', loc: 'Lagos, NG', result: 'Success', rule: 'r2' });
  logs.push({ id: 'g2', ago: 80, user: 'Collins Barasa', action: 'Downloaded 64 files from Finance', module: 'Files', ip: '10.0.6.88', loc: 'Office network', result: 'Success', rule: 'r3' });
  logs.push({ id: 'g3', ago: 120, user: 'Kevin Mutua', action: 'Shared 12 posts outside Hub in one hour', module: 'Wall', ip: '10.0.3.41', loc: 'Office network', result: 'Success', rule: 'r7' });
  logs.push({ id: 'g4', ago: 180, user: 'Samwel Okoth Omulo', action: 'Changed rights for the Staff role', module: 'Roles', ip: '10.0.4.21', loc: 'Office network', result: 'Success', rule: 'r5' });
  logs.push({ id: 'g5', ago: 300, user: 'External viewer', action: 'Opened secure link "Payroll login"', module: 'Secure share', ip: '185.220.101.4', loc: 'Unknown', result: 'Success', rule: 'r6' });
  logs.push({ id: 'g6', ago: 1220, user: 'Brian Kiprop', action: 'Reset MFA for Peter Kamau at 23:40', module: 'Roles', ip: '10.0.2.15', loc: 'VPN', result: 'Success', rule: 'r4' });
  logs.push({ id: 'g7', ago: 640, user: 'Unknown', action: 'Sign-in blocked from a disallowed country', module: 'Auth', ip: '91.240.118.9', loc: 'Unknown', result: 'Blocked', rule: null });
  logs.push({ id: 'g8', ago: 900, user: 'Unknown', action: 'Sign-in blocked: account locked', module: 'Auth', ip: '41.90.12.7', loc: 'Nairobi, KE', result: 'Blocked', rule: null });
  this.baseLogs = logs.sort((a, b) => a.ago - b.ago);
}
