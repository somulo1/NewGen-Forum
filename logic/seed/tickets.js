// Sample data derived on start-up: tickets. Called with the HubLogic instance as `this`.
export function seedTickets() {
  // Tickets now belong to departments
  const tMeta = { 'IT-208': ['IT Operations', 'Customer Care', '2026-10-15', false], 'DEV-91': ['Software Development', 'HR', '2026-10-12', false], 'IT-211': ['IT Operations', 'Sales', '2026-10-14', true], 'HR-14': ['HR', 'Operations', '2026-10-12', false], 'DEV-95': ['Software Development', 'Risk & Compliance', '2026-10-15', false] };
  this.state.tickets = this.state.tickets.map((t) => { const m = tMeta[t.id] || ['IT Operations', 'Risk & Compliance', '2026-10-15', false]; return { ...t, dept: m[0], fromDept: m[1], createdAt: m[2], escalated: m[3] }; });
}
