// Initial state: tickets.
export function ticketsState() {
  return {
    tickets: [], gallery: false, customize: false, acked: {}, rsvp: {}, wished: {}, hvote: null, note: '', forum: 'it', tab: 'discussions', menu: false, invite: false, create: false, link: false, liked: {}, vote: null
  };
}
