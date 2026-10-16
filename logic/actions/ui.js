import { chime } from '~/logic/tones.js';
// HubLogic actions: ui. Mixed into HubLogic.prototype, so `this` is the HubLogic instance.
export const uiActions = {
  componentDidMount() { try { if (window.innerWidth <= 900 && this.state.mode === 'docked') this.setState({ mode: 'min' }); } catch (err) {} this.scrollChat(); this._tick = setInterval(() => this.forceUpdate(), 15000); this.startPollTimer(); },

  componentWillUnmount() { clearInterval(this._tick); clearInterval(this._ct); clearInterval(this._pollT); clearTimeout(this._toastT); clearTimeout(this._incT); },

  // Widest the docked chat panel may be. On the Community wall it stops before it would push the wall's
  // right-hand cards out: wall column (460) + gap (24) + cards (250), the page padding, the sidebar and the panel's margin.
  panelMax() {
    const base = 640;
    if (this.state.section !== 'feed' || typeof window === 'undefined') return base;
    const w = window.innerWidth;
    if (w <= 900) return base;
    const pad = Math.min(40, Math.max(16, w * 0.03)) + Math.min(8, Math.max(4, w * 0.005));
    return Math.max(320, Math.min(base, w - 240 - 24 - pad - 734));
  },

  startDrag(e) {
    e.preventDefault();
    const startX = e.clientX, startW = this.state.pw;
    const move = (ev) => {
      const w = Math.min(this.panelMax(), Math.max(320, startW + (startX - ev.clientX)));
      this.setState({ pw: w });
    };
    const up = () => {
      window.removeEventListener('pointermove', move);
      window.removeEventListener('pointerup', up);
      this.setState({ dragging: false });
    };
    window.addEventListener('pointermove', move);
    window.addEventListener('pointerup', up);
    this.setState({ dragging: true });
  },

  notify(text, color, ticket) {
    chime();
    this.setState((st) => ({ notifs: [{ text: text, color: color || '#B8F36B', time: 'Just now', ticket: ticket || null }].concat(st.notifs), notifSeen: false, toast: text, toastColor: color || '#B8F36B' }));
    clearTimeout(this._toastT);
    this._toastT = setTimeout(() => this.setState({ toast: '' }), 3800);
  },

  updTicket(id, fn) {
    this.setState((st) => ({ tickets: st.tickets.map((t) => (t.id === id ? fn(t) : t)) }));
  },

  act(t, entry) {
    return t.activity.concat([Object.assign({ who: 'Sam', time: 'Just now', kind: 'event' }, entry)]);
  },

  componentDidUpdate() { this.scrollChat(); },

  scrollChat() {
    if (this.chatEl && this._scrolledKey !== this._chatKey) {
      this.chatEl.scrollTop = this.chatEl.scrollHeight;
      this._scrolledKey = this._chatKey;
    }
  },

  // At most 5 files per post, message or ticket step; tells the person when some were left out.
  capFiles(already, list) {
    const room = Math.max(0, 5 - already);
    if (list.length > room) this.notify('You can attach up to 5 files at a time. ' + (room ? 'The first ' + room + ' were added.' : 'Remove one first.'), '#FFB86B');
    return list.slice(0, room);
  },

  // A picked file as the {name, meta, url, isImg} record every attachment view reads.
  fileInfo(f) {
    return { name: f.name, meta: this.fmtSize(f.size), url: URL.createObjectURL(f), isImg: (f.type || '').indexOf('image/') === 0, isVideo: (f.type || '').indexOf('video/') === 0 };
  },

  fmtSize(n) {
    if (n > 1048576) return (n / 1048576).toFixed(1) + ' MB';
    return Math.max(1, Math.round(n / 1024)) + ' KB';
  }
};
