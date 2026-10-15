// Tiny runtime that lets the prototype's logic classes run inside Vue 3 unchanged.
// State works like React: setState merges top-level keys, renderVals() feeds the template.
import { shallowReactive, shallowRef } from 'vue';

export class DCLogic {
  constructor(props) {
    this.props = props || {};
    this.state = {};
    this.__tick = shallowRef(0);
  }
  setState(update, done) {
    const patch = typeof update === 'function' ? update(this.state, this.props) : update;
    if (patch) Object.assign(this.state, patch);
    if (done) done();
  }
  forceUpdate() { this.__tick.value++; }
}

export function useDC(Logic, names) {
  const computed = {
    dcVals() {
      this.$.__dc.__tick.value; // track forceUpdate
      return this.$.__dc.renderVals() || {};
    }
  };
  names.forEach((n) => { computed[n] = function () { return this.dcVals[n]; }; });
  return {
    beforeCreate() {
      const logic = new Logic(this.$props);
      logic.props = this.$props;
      logic.state = shallowReactive(logic.state || {});
      this.$.__dc = logic;
    },
    computed,
    mounted() { if (this.$.__dc.componentDidMount) this.$.__dc.componentDidMount(); },
    updated() { if (this.$.__dc.componentDidUpdate) this.$.__dc.componentDidUpdate(); },
    beforeUnmount() { if (this.$.__dc.componentWillUnmount) this.$.__dc.componentWillUnmount(); }
  };
}
