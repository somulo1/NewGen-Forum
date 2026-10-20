// Initial state: comms.
export function commsState() {
  return {
    cmOpen: false, cmd: { type: 'Memo', title: '', body: '', wall: false, forum: '', sel: {} },
    comms: []
  };
}
