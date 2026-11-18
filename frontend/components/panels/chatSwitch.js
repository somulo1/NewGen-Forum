// Moving between chats on a phone, where the chat fills the screen: the list button opens the other chats over it, and a swipe
// across the chat's header goes to the next (swipe left) or previous (swipe right) chat in the list.
const MIN_DX = 60;

export const chatSwitchMixin = {
  data() { return { railOpen: false }; },
  methods: {
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
