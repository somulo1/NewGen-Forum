// Initial state: awards.
export function awardsState() {
  return {
    awOpen: false,
    awd: { type: 'Employee of the month', people: [], role: 'all', q: '', citation: '', poster: '', wall: true, forums: [] },
    awardList: []
  };
}
