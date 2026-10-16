<!-- Calendar -->
<template><template v-if="show.calendar">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: center; gap: 12px">
<h1 style="margin: 0; flex: 1 1 200px; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Calendar</h1>
<div style="display: flex; align-items: center; gap: 6px">
<button aria-label="Previous month" @click="cal.prevMonth" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg></button>
<strong style="min-width: 150px; text-align: center; font-size: 16px">{{ cal.monthLabel }}</strong>
<button aria-label="Next month" @click="cal.nextMonth" style="width: 38px; height: 38px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6"></path></svg></button>
<button @click="cal.today" style="height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Today</button>
</div>
<button @click="newEvent" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">+ New</button>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(lg, lg__i) in cal.legend" :key="lg__i"><button @click="lg.toggle" :aria-pressed="lg.on" :style="lg.style"><span :style="`width: 9px; height: 9px; border-radius: 50%; background: ${lg.color}`"></span>{{ lg.label }}</button></template></div>
<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 620px; min-width: 0; overflow-x: auto">
<div class="hub-cal" style="min-width: 640px; border-radius: 16px; border: 1px solid #24322C; background: #16201C; overflow: hidden">
<div style="display: grid; grid-template-columns: repeat(7, minmax(0, 1fr)); background: #121B17"><template v-for="(wd, wd__i) in cal.weekdays" :key="wd__i"><span style="padding: 10px; font-size: 11px; font-weight: 600; letter-spacing: 0.05em; text-transform: uppercase; color: #9DB0A6">{{ wd.l }}</span></template></div>
<div style="display: grid; grid-template-columns: repeat(7, minmax(0, 1fr))">
<template v-for="(dy, dy__i) in cal.days" :key="dy__i">
<button @click="dy.pick" :aria-label="dy.aria" :style="dy.style"><span :style="dy.numStyle">{{ dy.n }}</span><template v-for="(it, it__i) in dy.items" :key="it__i"><span :style="it.style">{{ it.label }}</span></template><template v-if="dy.hasMore"><span style="font-size: 10px; color: #9DB0A6">+{{ dy.more }} more</span></template></button>
</template>
</div>
</div>
</div>
<aside aria-label="Agenda" style="flex: 1 1 300px; min-width: 0; padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 16px">{{ cal.selLabel }}</strong>
<template v-if="cal.noAgenda"><span style="font-size: 13px; color: #9DB0A6">Nothing planned. Use + New to add something.</span></template>
<template v-for="(ag, ag__i) in cal.agenda" :key="ag__i">
<div :style="`padding: 12px; border-radius: 12px; background: #121B17; border: 1px solid ${ag.color}; display: flex; flex-direction: column; gap: 6px`">
<div style="display: flex; gap: 8px; align-items: center"><span style="font-family: 'Space Grotesk', sans-serif; font-size: 13px; color: #C9D6CF">{{ ag.time }}</span><span :style="ag.typeStyle">{{ ag.type }}</span></div>
<strong style="font-size: 14px; font-weight: 600">{{ ag.title }}</strong>
<template v-if="ag.hasSub"><span style="font-size: 12px; color: #9DB0A6">{{ ag.sub }}</span></template>
<template v-if="ag.hasPeople"><div style="display: flex; align-items: center"><template v-for="(pp, pp__i) in ag.people" :key="pp__i"><span :title="pp.name" :style="`width: 24px; height: 24px; margin-right: -6px; border-radius: 50%; border: 2px solid #121B17; background: ${pp.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 9px`">{{ pp.initials }}</span></template><span style="font-size: 11px; color: #9DB0A6; margin-left: 12px">{{ ag.peopleLabel }}</span></div></template>
<div style="display: flex; gap: 6px"><template v-if="ag.canJoin"><button @click="ag.join" style="height: 32px; padding: 0 12px; border-radius: 8px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 12px">Join</button></template><template v-if="ag.canRemove"><button @click="ag.remove" style="height: 32px; padding: 0 12px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">{{ ag.removeLabel }}</button></template></div>
</div>
</template>
</aside>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'CalendarSection',
  ...hubBindings(["cal","newEvent","show"])
};
</script>
