// Initial state: notes (notebooks hold notes, like folders hold files).
export function notesState() {
  const h = 3600000, now = Date.now();
  return {
    noteBooks: [{ id: 'nb1', name: 'Quick notes', color: '#FFD66B' }, { id: 'nb2', name: 'Meetings', color: '#7CC4FF' }, { id: 'nb3', name: 'Ideas', color: '#C9B6FF' }],
    notes: [
      { id: 'n1', book: 'nb2', title: 'Risk review prep', body: 'Collect access review evidence.\nAsk Amina for the audit list.\nBook the room for Thursday.', updated: now - 3 * h, pinned: true },
      { id: 'n2', book: 'nb2', title: 'Town hall questions', body: 'Ask about the new leave policy and the office move.', updated: now - 30 * h, pinned: false },
      { id: 'n3', book: 'nb3', title: 'Self-service password reset', body: 'Let people reset their own password with a one-time code and security questions.', updated: now - 52 * h, pinned: false },
      { id: 'n4', book: 'nb1', title: 'Call Brian about the VPN certificate', body: 'Renew before Friday.', updated: now - 5 * h, pinned: false },
      { id: 'n5', owner: 'Grace Njeri', access: 'view', book: '', title: 'Leave policy FAQ', updated: now - 20 * h, pinned: false, shares: [{ id: 'sh1', kind: 'person', to: 'Sam', toName: 'Sam', perm: 'view', at: now - 20 * h }],
        body: '<h2>Leave policy FAQ</h2><p><strong>Annual leave</strong> is 21 working days. Request it at least <em>two weeks</em> ahead.</p><table><tbody><tr><th><p>Leave type</p></th><th><p>Days</p></th><th><p>Approval</p></th></tr><tr><td><p>Annual</p></td><td><p>21</p></td><td><p>Line manager</p></td></tr><tr><td><p>Sick</p></td><td><p>14</p></td><td><p>HR</p></td></tr><tr><td><p>Maternity</p></td><td><p>90</p></td><td><p>HR</p></td></tr></tbody></table><p>Questions? Ask HR.</p>' },
      { id: 'n6', owner: 'Brian Kiprop', access: 'edit', book: '', title: 'Maintenance window plan', updated: now - 8 * h, pinned: false, shares: [{ id: 'sh2', kind: 'person', to: 'Sam', toName: 'Sam', perm: 'edit', at: now - 8 * h }],
        body: '<h2>Maintenance window plan</h2><ul><li>Saturday 22:00: freeze deployments</li><li>Saturday 23:00: patch the mail server</li><li>Sunday 01:00: verify and reopen</li></ul><p>Add anything I missed. You can edit this note.</p>' }
    ],

    noteShare: null, noteBook: null, noteEdit: null, noteQ: '', noteNewBook: null, noteFlash: ''
  };
}
