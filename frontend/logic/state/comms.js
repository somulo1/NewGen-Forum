// Initial state: comms.
export function commsState() {
  return {
    cmOpen: false, cmd: { type: 'Memo', title: '', summary: '', body: '', fields: {}, files: [], wall: false, forum: '', sel: {} }, commOpen: null, commPage: '', commEng: { loading: false, err: '', rows: [], at: 0 }, commTab: '', commView: 'read', commIns: null,
    comms: []
  };
}
