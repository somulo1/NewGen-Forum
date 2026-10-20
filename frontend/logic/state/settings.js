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

// Neutral until the server's values arrive (see logic/server/widgets/settings.js).
export function settingsState() {
  return {
    setTab: 'profile',
    orgProfile: { name: '', short: '', logo: '', industry: '', size: '', website: '', email: '', phone: '', address: '', city: '', country: '', timezone: 'Africa/Nairobi', language: 'English', currency: 'KES', regNo: '', taxId: '' },
    orgDeleteAfter: '', orgSecurity: { mfa: false, session: '8', domains: '', retention: '12', externalSharing: true, maxVideoMB: '50', maxFileMB: '10', fileTypes: 'pdf,doc,docx,xls,xlsx,ppt,pptx,csv,txt,png,jpg,jpeg', userQuotaMB: '5120' },
    setDraft: null,
    orgSub: { plan: 'business', cycle: 'monthly', seats: 1, autoRenew: true, renews: '', status: 'Active', cancelAt: '', addons: { storage: false, video: false, sms: false } },
    orgBilling: { cardBrand: '', cardLast4: '', cardExp: '', holder: '', contactEmail: '', vat: '' },
    invoices: [],
    setConfirm: null,
    etrOpen: null
  };
}
