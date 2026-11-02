<template>
<span :aria-hidden="!ariaLabel" :aria-label="ariaLabel || undefined" :role="ariaLabel ? 'img' : undefined" :style="`position: relative; display: flex; align-items: center; justify-content: center; flex: none; width: ${size}px; height: ${size}px; box-sizing: border-box; overflow: hidden; border-radius: ${radius}; background: ${color}; color: ${textColor}; font-weight: 700; font-size: ${fontSize}px; ${border ? `border: ${border};` : ''}`">
<img v-if="photo" :src="photo" :alt="ariaLabel ? '' : undefined" style="position: absolute; inset: 0; width: 100%; height: 100%; object-fit: cover">
<span v-else>{{ initials }}</span>
</span>
</template>

<script>
import { computed, ref, watchEffect } from 'vue';
import { useHubStore } from '~/composables/useHub';
import { photoUrl } from '~/logic/server/photos.js';

const norm = (n) => String(n || '').trim().toLowerCase();

export default {
  name: 'UserAvatar',
  props: {
    name: { type: String, default: '' },
    initials: { type: String, default: '' },
    color: { type: String, default: '#C9B6FF' },
    size: { type: Number, default: 36 },
    fontSize: { type: Number, default: 12 },
    radius: { type: String, default: '50%' },
    border: { type: String, default: '' },
    textColor: { type: String, default: '#0F1714' },
    ariaLabel: { type: String, default: '' },
    // Who this is, when the caller knows: the person's id, and whether it is me. With these the avatar never has to guess from a name.
    uid: { type: String, default: '' },
    self: { type: Boolean, default: undefined }
  },
  setup(props) {
    const { logic } = useHubStore();
    // the person this avatar is for, from the directory: by id when given, otherwise by name when exactly one person has that name
    const known = () => {
      const srv = logic.srv;
      if (!srv) return null;
      if (props.uid) return (srv.dirByUid && srv.dirByUid.get(props.uid)) || null;
      const n = norm(props.name);
      const hits = (logic.peopleDir ? logic.peopleDir() : logic.dir || []).filter((p) => norm(p.name) === n);
      return hits.length === 1 ? hits[0] : null;
    };
    // Mine means mine: the caller says so, or the id is mine, or the label is "You". A name alone counts only when nobody else shares it
    // (two people called the same must not both get my photo).
    const isSelf = computed(() => {
      if (props.self !== undefined) return props.self;
      const srv = logic.srv;
      if (props.uid) return !!srv && props.uid === srv.me.person;
      const n = norm(props.name);
      if (n === 'you') return true;
      if (!srv) return n === 'sam';
      if (n !== norm(srv.me.name)) return false;
      const others = (logic.dir || []).filter((p) => norm(p.name) === n);
      return others.length === 0;
    });
    const mine = computed(() => logic.state.profile.photo);
    const other = ref('');
    watchEffect((onCleanup) => {
      other.value = '';
      if (isSelf.value) return;
      const p = known();
      if (!p || !p.hasPhoto || !p.uid) return;
      let live = true;
      onCleanup(() => { live = false; });
      photoUrl(logic.srv, p.uid, p.photoVer).then((u) => { if (live) other.value = u; });
    });
    const photo = computed(() => (isSelf.value ? mine.value : other.value));
    return { photo };
  }
};
</script>
