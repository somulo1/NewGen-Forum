// Ratings: on calls, meetings and standups (Organization settings → Feedback), and on closed tickets (for the people who
// worked on them and the heads of the place each ticket was with). Both are read a page at a time.

export async function loadCallFeedback(logic, srv, page, context) {
  const key = page + '|' + context;
  srv.fbCallKey = key;
  logic.setState((st) => ({ callFbAll: { ...(st.callFbAll || {}), page, context, loading: true, err: '' } }));
  try {
    const d = await srv.request('GET', `/calls/feedback/all?page=${page}&size=10${context ? '&context=' + context : ''}`);
    if (srv.fbCallKey !== key) return;   // a newer page was asked for meanwhile
    logic.setState({ callFbAll: { ...d, page, context, loading: false, err: '', loaded: true } });
  } catch (e) {
    if (srv.fbCallKey === key) logic.setState((st) => ({ callFbAll: { ...(st.callFbAll || {}), loading: false, loaded: true, err: e.message || 'Could not load the feedback.' } }));
  }
}

export async function loadTicketFeedback(logic, srv, page) {
  srv.fbTkPage = page;
  logic.setState((st) => ({ tkFbList: { ...(st.tkFbList || {}), page, loading: true, err: '' } }));
  try {
    const d = await srv.request('GET', `/tickets/feedback?page=${page}&size=10`);
    if (srv.fbTkPage !== page) return;
    logic.setState({ tkFbList: { ...d, page, loading: false, err: '', loaded: true } });
  } catch (e) {
    if (srv.fbTkPage === page) logic.setState((st) => ({ tkFbList: { ...(st.tkFbList || {}), loading: false, loaded: true, err: e.message || 'Could not load the ratings.' } }));
  }
}

export const starsText = (n) => { const r = Math.max(0, Math.min(5, Math.round(n || 0))); return '★★★★★'.slice(0, r) + '☆☆☆☆☆'.slice(r); };
