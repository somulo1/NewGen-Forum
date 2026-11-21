// Moving between chats on a phone, where the chat fills the screen: the list button opens the other chats over it, and a swipe
// across the chat's header goes to the next (swipe left) or previous (swipe right) chat in the list.
const MIN_DX = 60;

export const chatSwitchMixin = {
  data() { return { railOpen: false }; },
  methods: {
    // From a reply's quoted card to the message it answers: scrolled into view and lit up for a moment. A message in the part of the
    // chat not shown yet is brought in first.
    jumpTo(mid, tries = 0) {
      if (!mid) return;
      const find = () => Array.from(document.querySelectorAll('[data-mid]')).find((x) => x.getAttribute('data-mid') === mid && x.offsetParent !== null);
      const el = find();
      if (el) { el.scrollIntoView({ block: 'center', behavior: 'smooth' }); el.classList.add('cw-flash'); setTimeout(() => el.classList.remove('cw-flash'), 1600); return; }
      if (tries < 12 && ((this.olderCount || 0) > 0 || this.moreOlder)) { if (this.showOlder) this.showOlder(); setTimeout(() => this.jumpTo(mid, tries + 1), 250); }
    },
    swipeStart(e) {
      if (e.pointerType === 'mouse') return;
      this.swipeAt = { x: e.clientX, y: e.clientY, id: e.pointerId };
    },
    swipeEnd(e) {
      const at = this.swipeAt;
      this.swipeAt = null;
      if (!at || at.id !== e.pointerId) return;
      const dx = e.clientX - at.x, dy = e.clientY - at.y;
      if (Math.abs(dx) < MIN_DX || Math.abs(dy) > Math.abs(dx) * 0.6) return;
      const go = dx < 0 ? this.chatSwitch.next : this.chatSwitch.prev;
      if (go) go();
    }
  }
};
