<!-- Give an award dialog -->
<template><template v-if="awOpen">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 66; background: rgba(5,10,8,0.65); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box">
<div role="dialog" aria-label="Give an award" style="width: 100%; max-width: 720px; max-height: 92vh; display: flex; flex-direction: column; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 16px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)"><h2 style="margin: 0; flex: 1; font-family: 'Space Grotesk', sans-serif; font-size: 20px">Give an award</h2><button aria-label="Close" @click="aw.close" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Award</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(at, at__i) in aw.types" :key="at__i"><button @click="at.pick" :style="at.style">{{ at.label }}</button></template></div></div>

<div style="display: flex; flex-direction: column; gap: 8px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C)">
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center"><strong style="flex: 1; font-size: 14px">Who gets it? <span style="color: var(--k-9DB0A6,#9DB0A6); font-weight: 400">One person or many</span></strong><span style="font-size: 12px; color: var(--k-FFD66B,#FFD66B); font-weight: 600">{{ aw.countLabel }}</span></div>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(ac, ac__i) in aw.chosen" :key="ac__i"><span style="height: 32px; padding: 0 4px; border-radius: 16px; background: var(--k-24220F,#24220F); border: 1px solid var(--k-4A4426,#4A4426); display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--k-F3EBC8,#F3EBC8)"><UserAvatar :name="ac.name" :initials="ac.initials" :color="ac.color" :size="24" :font-size="9"/>{{ ac.name }}<button :aria-label="`Remove ${ac.name}`" @click="ac.remove" style="width: 22px; height: 22px; border-radius: 50%; border: 0; background: var(--k-3A3418,#3A3418); color: var(--k-F3EBC8,#F3EBC8); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></span></template></div>
<div style="display: grid; grid-template-columns: minmax(0, 1fr) 190px; gap: 8px">
<label style="display: flex; align-items: center; gap: 8px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-9DB0A6,#9DB0A6)"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"></circle><path d="M20 20l-4-4"></path></svg><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Search people</span><input type="search" :value="aw.q" @input="aw.setQ" placeholder="Search by name or job" style="flex: 1; min-width: 0; border: 0; outline: none; background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<label style="display: flex; flex-direction: column"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Filter by role</span><select :value="aw.role" @change="aw.setRole" style="height: 40px; padding: 0 8px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-16201C,#16201C); color: var(--k-E8EFEA,#E8EFEA); font-size: 13px"><option value="all">All roles</option><template v-for="(rop, rop__i) in aw.roleOpts" :key="rop__i"><option :value="rop.v">{{ rop.l }}</option></template></select></label>
</div>
<div style="display: flex; justify-content: space-between; align-items: center; font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)"><span>{{ aw.matchLabel }}</span><button @click="aw.selectAll" style="height: 30px; padding: 0 10px; border-radius: 15px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; font-size: 12px">Select all shown</button></div>
<div style="display: flex; flex-direction: column; max-height: 220px; overflow-y: auto; border-radius: 10px; border: 1px solid var(--k-24322C,#24322C)">
<template v-for="(am, am__i) in aw.matches" :key="am__i"><button @click="am.toggle" :aria-pressed="am.on" :style="am.style"><UserAvatar :name="am.name" :initials="am.initials" :color="am.color" :size="28" :font-size="10"/><span style="flex: 1; min-width: 0; display: flex; flex-direction: column; text-align: left"><span style="font-size: 13px">{{ am.name }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ am.sub }}</span></span><span style="font-size: 12px; font-weight: 700; color: var(--k-FFD66B,#FFD66B)">{{ am.mark }}</span></button></template>
</div>
</div>

<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Why are they being recognised?<textarea rows="3" :value="aw.citation" @input="aw.setCitation" placeholder="e.g. Kept every branch online during the network upgrade" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; resize: vertical"></textarea></label>

<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Poster <span style="color: var(--k-9DB0A6,#9DB0A6)">(optional — we make one if you don't upload)</span></span>
<div style="display: flex; flex-wrap: wrap; gap: 12px; align-items: flex-start">
<template v-if="aw.hasPoster"><img :src="aw.poster" alt="Uploaded poster" style="width: 180px; height: 120px; object-fit: cover; border-radius: 10px; border: 1px solid var(--k-4A4426,#4A4426)"></template>
<template v-if="aw.noPoster"><div aria-hidden="true" style="width: 180px; height: 120px; border-radius: 10px; background: var(--k-24220F,#24220F); border: 1px solid var(--k-4A4426,#4A4426); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 4px; text-align: center; padding: 8px; box-sizing: border-box"><span style="font-size: 9px; letter-spacing: 0.1em; text-transform: uppercase; color: var(--k-FFD66B,#FFD66B); font-weight: 700">{{ aw.type }}</span><strong style="font-size: 12px; color: var(--k-F3EBC8,#F3EBC8); line-height: 1.3">{{ aw.posterNames }}</strong></div></template>
<div style="display: flex; flex-direction: column; gap: 8px">
<label style="position: relative; height: 38px; padding: 0 14px; border-radius: 10px; border: 1px solid var(--k-4A4426,#4A4426); color: var(--k-F3EBC8,#F3EBC8); cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px"><input type="file" accept="image/*" aria-label="Upload a poster" @change="aw.setPoster" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer">Upload poster</label>
<template v-if="aw.hasPoster"><button @click="aw.clearPoster" style="height: 34px; padding: 0 12px; border-radius: 10px; border: 0; background: transparent; color: var(--k-FF9B8A,#FF9B8A); cursor: pointer; font-size: 12px">Remove poster</button></template>
</div>
</div>
</div>

<div style="display: flex; flex-direction: column; gap: 10px; padding: 14px; border-radius: 14px; background: var(--k-121B17,#121B17); border: 1px solid var(--k-24322C,#24322C)">
<strong style="font-size: 14px">Also post it to</strong>
<label style="display: flex; gap: 10px; align-items: center; font-size: 14px; cursor: pointer"><input type="checkbox" :checked="aw.wall" @change="aw.toggleWall">Community wall</label>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(afc, afc__i) in aw.forumChips" :key="afc__i"><button :aria-pressed="afc.on" @click="afc.toggle" :style="afc.style">{{ afc.name }}</button></template></div>
<span style="font-size: 12px; color: var(--k-9DB0A6,#9DB0A6)">The poster goes with the post and every awardee is tagged and notified.</span>
</div>
<span style="font-size: 13px; color: var(--k-FFB86B,#FFB86B)">{{ aw.approvalNote }}</span>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid var(--k-24322C,#24322C)"><button @click="aw.close" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button><button @click="aw.submit" :style="aw.submitStyle">{{ aw.submitLabel }}</button></div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'AwardDialog',
  ...hubBindings(["aw","awOpen"])
};
</script>
