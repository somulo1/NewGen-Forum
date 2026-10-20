// Initial state: organization settings (profile, subscription, billing, security and data).
export const PLANS = [
  { id: 'starter', name: 'Starter', price: 4, seats: 50, storage: 50, blurb: 'Core intranet for small teams.', features: ['Wall, forums and chat', 'Files and calendar', '50 GB storage'] },
  { id: 'business', name: 'Business', price: 9, seats: 500, storage: 500, blurb: 'Everything most organizations need.', features: ['Tickets, projects and approvals', 'Meetings and standups', '500 GB storage', 'Security logs'] },
  { id: 'enterprise', name: 'Enterprise', price: 15, seats: 5000, storage: 5000, blurb: 'Scale, control and support.', features: ['Unlimited structure levels', 'Single sign-on and audit exports', '5 TB storage', 'Priority support'] }
];
export const ADDONS = [
  { id: 'storage', name: 'Extra storage', desc: '+500 GB of files', price: 20 },
  { id: 'video', name: 'Video minutes', desc: '+5,000 meeting recording minutes', price: 15 },
  { id: 'sms', name: 'SMS alerts', desc: 'Text alerts for security and approvals', price: 10 }
];

export function settingsState() {
  const day = 86400000, now = Date.now();
  const iso = (t) => new Date(t).toISOString().slice(0, 10);
  return {
    setTab: 'profile',
    orgProfile: { name: '[YOUR ORGANIZATION]', short: 'Org', logo: '', industry: 'Financial services', size: '201–500', website: '', email: '', phone: '', address: '', city: '', country: 'Kenya', timezone: 'Africa/Nairobi', language: 'English', currency: 'KES', regNo: '', taxId: '' },
    orgSecurity: { mfa: false, session: '8', domains: '', retention: '12', externalSharing: true, maxVideoMB: '50', userQuotaMB: '5120' },
    setDraft: null,
    orgSub: { plan: 'business', cycle: 'monthly', seats: 300, autoRenew: true, renews: iso(now + 23 * day), status: 'Active', cancelAt: '', addons: { storage: false, video: true, sms: false } },
    orgBilling: { cardBrand: 'Visa', cardLast4: '4242', cardExp: '08/28', holder: '', contactEmail: '', vat: '' },
    invoices: [
      { no: 'INV-2026-010', date: iso(now - 7 * day), amount: 1980, status: 'Paid' },
      { no: 'INV-2026-009', date: iso(now - 37 * day), amount: 1980, status: 'Paid' },
      { no: 'INV-2026-008', date: iso(now - 67 * day), amount: 1960, status: 'Paid' },
      { no: 'INV-2026-007', date: iso(now - 97 * day), amount: 1940, status: 'Paid' }
    ],
    setConfirm: null,
    etrOpen: null
  };
}
