// Initial state: comms.
export function commsState() {
  return {
    cmOpen: false, cmd: { type: 'Memo', title: '', summary: '', body: '', fields: {}, files: [], wall: false, forum: '', sel: {} }, commOpen: null, commTab: '', commView: 'read', commIns: null,
    comms: []
  };
}
