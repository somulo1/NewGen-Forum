<!-- Calendar: new event / meeting / personal pin dialog -->
<template><template v-if="evOpen">
<div class="hub-overlay" style="position: fixed; inset: 0; z-index: 66; background: rgba(5,10,8,0.65); display: flex; align-items: center; justify-content: center; padding: 16px; box-sizing: border-box">
<div role="dialog" :aria-label="ev.heading" style="width: 100%; max-width: 640px; max-height: 92vh; display: flex; flex-direction: column; border-radius: 20px; background: var(--k-16201C,#16201C); border: 1px solid var(--k-2A3832,#2A3832); box-shadow: 0 24px 60px rgba(0,0,0,0.5); overflow: hidden">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 16px 20px; border-bottom: 1px solid var(--k-24322C,#24322C)"><h2 style="margin: 0; flex: 1; font-family: 'Space Grotesk', sans-serif; font-size: 20px">{{ ev.heading }}</h2><button aria-label="Close" @click="ev.close" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 18px 20px; display: flex; flex-direction: column; gap: 14px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Title<input type="text" :value="ev.title" @input="ev.setTitle" placeholder="What is it?" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 15px"></label>
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Type</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(ty, ty__i) in ev.types" :key="ty__i"><button @click="ty.pick" :style="ty.style">{{ ty.label }}</button></template></div></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Date<input type="date" :value="ev.date" @input="ev.setDate" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; color-scheme: dark"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Starts<input type="time" :value="ev.start" @input="ev.setStart" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; color-scheme: dark"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Ends<input type="time" :value="ev.end" @input="ev.setEnd" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; color-scheme: dark"></label>
</div>
<template v-if="ev.isOwn">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Notes<textarea rows="3" :value="ev.notes" @input="ev.setNotes" placeholder="Anything you want to remember" style="padding: 10px 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px; resize: vertical"></textarea></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Who can see it<select :value="ev.visibility" @change="ev.setVisibility" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="Only me">Only me</option><option value="My team">My team</option><option value="Everyone">Everyone</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Remind me<select :value="ev.remind" @change="ev.setRemind" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="None">No reminder</option><option value="At the time">At the time</option><option value="15 minutes before">15 minutes before</option><option value="1 hour before">1 hour before</option><option value="1 day before">1 day before</option></select></label>
</div>
</template>
<template v-if="ev.needsPeople">
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Repeats<select :value="ev.repeat" @change="ev.setRepeat" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="None">Does not repeat</option><option value="Daily">Every day</option><option value="Weekdays">Every weekday</option><option value="Weekly">Every week</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Forum (optional)<select :value="ev.forum" @change="ev.setForum" style="height: 42px; padding: 0 10px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"><option value="">None</option><template v-for="(f, f__i) in forums" :key="f__i"><option :value="f.name">{{ f.name }}</option></template></select></label>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: var(--k-C9D6CF,#C9D6CF)">Participants ({{ ev.count }} selected)</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(pc, pc__i) in ev.chosen" :key="pc__i"><span style="height: 32px; padding: 0 4px 0 4px; border-radius: 16px; background: var(--k-1F2C26,#1F2C26); border: 1px solid var(--k-2A3832,#2A3832); display: flex; align-items: center; gap: 6px; font-size: 12px"><UserAvatar :name="pc.name" :initials="pc.initials" :color="pc.color" :size="24" :font-size="9"/>{{ pc.name }}<button :aria-label="`Remove ${pc.name}`" @click="pc.remove" style="width: 22px; height: 22px; border-radius: 50%; border: 0; background: var(--k-2C3D34,#2C3D34); color: var(--k-E8EFEA,#E8EFEA); cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="9" height="9" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button></span></template></div>
<label style="display: flex; align-items: center; gap: 8px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: var(--k-121B17,#121B17); color: var(--k-9DB0A6,#9DB0A6)"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"></circle><path d="M20 20l-4-4"></path></svg><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Find people</span><input type="search" :value="ev.pq" @input="ev.setPq" placeholder="Find people to invite" style="flex: 1; min-width: 0; border: 0; outline: none; background: transparent; color: var(--k-E8EFEA,#E8EFEA); font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 2px; max-height: 210px; overflow-y: auto; border-radius: 10px; border: 1px solid var(--k-24322C,#24322C)">
<template v-for="(pm, pm__i) in ev.matches" :key="pm__i"><button @click="pm.toggle" :aria-pressed="pm.on" :style="pm.style"><UserAvatar :name="pm.name" :initials="pm.initials" :color="pm.color" :size="28" :font-size="10"/><span style="flex: 1; min-width: 0; display: flex; flex-direction: column; text-align: left"><span style="font-size: 13px">{{ pm.name }}</span><span style="font-size: 11px; color: var(--k-9DB0A6,#9DB0A6)">{{ pm.title }}</span></span><span style="font-size: 12px; font-weight: 700; color: var(--k-B8F36B,#B8F36B)">{{ pm.mark }}</span></button></template>
</div>
</div>
</template>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid var(--k-24322C,#24322C)"><button @click="ev.close" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid var(--k-2A3832,#2A3832); background: transparent; color: var(--k-E8EFEA,#E8EFEA); cursor: pointer">Cancel</button><button @click="ev.save" :style="ev.saveStyle">{{ ev.saveLabel }}</button></div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'EventDialog',
  ...hubBindings(["ev","evOpen","forums"])
};
</script>
