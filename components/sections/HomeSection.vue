<!-- Home: awards ribbon + widget masonry -->
<template><template v-if="show.home">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: flex-end">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Good afternoon, Sam</h1><p style="margin: 4px 0 0; color: #9DB0A6">Your home page. Add, remove and arrange widgets to suit your organization.</p></div>
<template v-if="canCustomize">
<button @click="toggleCustomize" :style="customizeBtnStyle">{{ customizeLabel }}</button>
<button v-if="canWidgetsPage" @click="openGallery" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 8px"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg>Add widget</button>
</template>
</div>

<template v-if="noWidgets">
<div style="padding: 40px 20px; border-radius: 18px; border: 1px dashed #4A5C54; text-align: center; color: #9DB0A6">No widgets yet. Use Add widget to build your home page.</div>
</template>

<template v-if="awardsOn && (!awardStrip.hidden || customizing)">
<section aria-label="Performance awards" style="border-radius: 20px; background: #1B1A12; border: 1px solid #4A4426; overflow: hidden; display: flex; flex-direction: column">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center; padding: 14px 18px 4px">
<svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#FFD66B" stroke-width="1.8" stroke-linejoin="round"><path d="M8 4h8v5a4 4 0 0 1-8 0z"></path><path d="M8 6H5a3 3 0 0 0 3 4M16 6h3a3 3 0 0 1-3 4M12 13v4M8 20h8M10 17h4"></path></svg>
<div style="display: flex; flex-direction: column"><strong style="font-size: 15px; color: #F3EBC8">{{ awardStrip.title }}</strong><span style="font-size: 12px; color: #C9BE8F">{{ awardStrip.sub }}</span></div>
<span style="flex: 1"></span>
<template v-if="customizing"><button @click="awardStrip.toggleHide" style="height: 32px; padding: 0 10px; border-radius: 8px; border: 1px solid #4A4426; background: transparent; color: #F3EBC8; cursor: pointer; font-size: 13px">{{ awardStrip.hideLabel }}</button><button @click="awardStrip.openSettings" style="height: 32px; padding: 0 10px; border-radius: 8px; border: 1px solid #4A4426; background: transparent; color: #F3EBC8; cursor: pointer; font-size: 12px">Settings</button><button @click="awardStrip.remove" aria-label="Remove awards from Home" style="height: 32px; padding: 0 10px; border-radius: 8px; border: 1px solid #5A3A36; background: transparent; color: #FF9B8A; cursor: pointer; font-size: 12px">Remove</button></template>
<template v-if="awardStrip.canGive"><button @click="openAward" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #FFD66B; color: #0F1714; font-weight: 700; cursor: pointer; font-size: 13px">Give an award</button></template>
</div>
<template v-if="awardStrip.empty"><div style="padding: 20px 18px 22px; color: #C9BE8F; font-size: 13px">No approved awards yet. Awards show here once they are approved.</div></template>
<div style="overflow: hidden; padding: 10px 0 16px">
<div class="hub-marquee" :style="awardStrip.trackStyle">
<template v-for="(ap, ap__i) in awardStrip.items" :key="ap__i">
<button @click="ap.open" :aria-label="`Open ${ap.name}'s profile, ${ap.award}`" style="flex: none; width: 190px; padding: 14px 12px; border-radius: 16px; background: #24220F; border: 1px solid #4A4426; color: #F3EBC8; cursor: pointer; display: flex; flex-direction: column; align-items: center; gap: 8px; text-align: center; font-family: inherit">
<span style="width: 76px; height: 76px; border-radius: 50%; padding: 3px; background: #FFD66B; box-sizing: border-box; display: flex"><span :style="`flex: 1; border-radius: 50%; background: ${ap.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 24px; overflow: hidden`"><template v-if="ap.hasPhoto"><img :src="ap.photo" alt="" style="width: 100%; height: 100%; object-fit: cover"></template><template v-if="ap.noPhoto">{{ ap.initials }}</template></span></span>
<strong style="font-size: 14px">{{ ap.name }}</strong>
<span style="font-size: 12px; font-weight: 700; color: #FFD66B">{{ ap.award }}</span>
<span style="font-size: 11px; color: #C9BE8F">{{ ap.when }}</span>
</button>
</template>
</div>
</div>
</section>
</template>
<div style="columns: 300px; column-gap: 16px">
<template v-for="(w, w__i) in widgets" :key="w__i">
<section :aria-label="w.title" :style="w.cardStyle + (w.hidden ? ';opacity: 0.5' : '')">
<div style="display: flex; align-items: center; gap: 10px">
<span :style="`width: 10px; height: 10px; flex: none; border-radius: 50%; background: ${w.tint}`"></span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 15px">{{ w.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ w.sub }}</span></div>
<template v-if="w.customizing">
<button :aria-label="`${w.hideLabel} ${w.title}`" @click="w.toggleHide" style="height: 32px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">{{ w.hideLabel }}</button>
<button :aria-label="`Settings for ${w.title}`" @click="w.openSettings" style="height: 32px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 6px; font-size: 12px"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><circle cx="12" cy="12" r="3.5"></circle><path d="M12 2.5v3M12 18.5v3M2.5 12h3M18.5 12h3M5.3 5.3l2.1 2.1M16.6 16.6l2.1 2.1M5.3 18.7l2.1-2.1M16.6 7.4l2.1-2.1"></path></svg>Settings</button>
<button :aria-label="`Move ${w.title} earlier`" @click="w.moveUp" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg></button>
<button :aria-label="`Remove ${w.title}`" @click="w.remove" style="width: 32px; height: 32px; border-radius: 8px; border: 1px solid #5A3A36; background: transparent; color: #FF9B8A; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</template>
</div>

<template v-if="w.t.announcements">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(a, a__i) in w.items" :key="a__i">
<div style="display: flex; gap: 10px; align-items: flex-start; padding: 10px; border-radius: 12px; background: #121B17">
<span style="flex: none; padding: 2px 8px; border-radius: 8px; font-size: 11px; font-weight: 600; background: #22302A; color: #B8F36B">{{ a.tag }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 14px; font-weight: 600; line-height: 1.35">{{ a.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ a.meta }}</span></div>
<button @click="a.ack" :style="a.ackStyle">{{ a.ackLabel }}</button>
</div>
</template>
</div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px"><button @click="w.compose" style="height: 38px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">New announcement</button><button @click="w.goComms" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">See all</button></div>
</template>

<template v-if="w.t.links">
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px">
<template v-for="(l, l__i) in w.items" :key="l__i">
<button style="min-height: 76px; padding: 10px 6px; border-radius: 12px; border: 1px solid #24322C; background: #121B17; color: #E8EFEA; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px; font-size: 12px"><span :style="`width: 32px; height: 32px; border-radius: 9px; background: ${l.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 13px`">{{ l.letter }}</span>{{ l.label }}</button>
</template>
</div>
</template>

<template v-if="w.t.events">
<div style="display: flex; flex-direction: column; gap: 10px">
<template v-for="(ev, ev__i) in w.items" :key="ev__i">
<div style="display: flex; gap: 12px; align-items: center">
<div style="width: 48px; height: 52px; flex: none; border-radius: 10px; background: #22302A; display: flex; flex-direction: column; align-items: center; justify-content: center"><span style="font-size: 10px; font-weight: 700; color: #FFB86B">{{ ev.mon }}</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 20px; font-weight: 700; line-height: 1">{{ ev.day }}</span></div>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ ev.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ ev.meta }}</span></div>
<button @click="ev.rsvp" :style="ev.rsvpStyle">{{ ev.rsvpLabel }}</button>
</div>
</template>
</div>
</template>

<template v-if="w.t.clocks">
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px">
<template v-for="(c, c__i) in w.items" :key="c__i">
<div style="padding: 12px; border-radius: 12px; background: #121B17; display: flex; flex-direction: column; gap: 2px"><span style="font-size: 12px; color: #9DB0A6">{{ c.city }}</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 26px; font-weight: 600">{{ c.time }}</span><span style="font-size: 11px; color: #9DB0A6">{{ c.day }}</span></div>
</template>
</div>
</template>

<template v-if="w.t.birthdays">
<div style="display: flex; flex-direction: column; gap: 10px">
<template v-for="(b, b__i) in w.items" :key="b__i">
<div style="display: flex; gap: 10px; align-items: center">
<span aria-hidden="true" :style="`width: 38px; height: 38px; flex: none; border-radius: 50%; background: ${b.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 13px`">{{ b.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ b.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ b.what }}</span></div>
<button @click="b.wish" :style="b.wishStyle">{{ b.wishLabel }}</button>
</div>
</template>
</div>
</template>

<template v-if="w.t.tickets">
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px">
<template v-for="(k, k__i) in w.items" :key="k__i">
<div style="padding: 12px; border-radius: 12px; background: #121B17; display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 26px; font-weight: 700; color: ${k.color}`">{{ k.n }}</span><span style="font-size: 12px; color: #C9D6CF; line-height: 1.3">{{ k.label }}</span></div>
</template>
</div>
<div style="display: flex; flex-direction: column; gap: 6px">
<span style="font-size: 12px; color: #9DB0A6">Needs attention</span>
<template v-for="(ut, ut__i) in w.urgent" :key="ut__i">
<div style="display: flex; gap: 10px; align-items: center; padding: 8px 10px; border-radius: 10px; background: #121B17">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px"><strong style="font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ ut.id }} · {{ ut.title }}</strong><span style="display: flex; gap: 6px; align-items: center; font-size: 11px; color: #9DB0A6"><span :style="ut.slaBadgeStyle">{{ ut.slaLabel }}</span>{{ ut.remainingLabel }}</span></div>
<button @click="ut.open" :aria-label="`View ${ut.id}`" style="flex: none; height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px; font-weight: 600">View</button>
</div>
</template>
</div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px"><button @click="w.goTickets" style="height: 38px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">New ticket</button><button @click="w.goTickets" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">View all</button></div>
</template>

<template v-if="w.t.files">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(d, d__i) in w.items" :key="d__i">
<div style="display: flex; gap: 10px; align-items: center"><span :style="`width: 36px; height: 36px; flex: none; border-radius: 9px; background: #22302A; color: ${d.color}; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 10px`">{{ d.kind }}</span><div style="min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ d.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ d.meta }}</span></div></div>
</template>
</div>
</template>

<template v-if="w.t.joiners">
<div style="display: flex; flex-direction: column; gap: 10px">
<template v-for="(j, j__i) in w.items" :key="j__i">
<div style="display: flex; gap: 10px; align-items: center">
<span aria-hidden="true" :style="`width: 38px; height: 38px; flex: none; border-radius: 50%; background: ${j.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 13px`">{{ j.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ j.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ j.role }}</span></div>
<button style="height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">Say hi</button>
</div>
</template>
</div>
</template>

<template v-if="w.t.sessions">
<div style="display: flex; flex-direction: column; gap: 10px">
<template v-for="(x, x__i) in w.items" :key="x__i">
<div style="display: flex; gap: 10px; align-items: center">
<span :style="`width: 36px; height: 36px; flex: none; border-radius: 10px; background: ${x.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 12px`">{{ x.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ x.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ x.when }}</span></div>
<template v-if="x.live"><span style="font-size: 11px; font-weight: 700; color: #FF9B8A">LIVE</span></template>
</div>
</template>
</div>
<div style="display: grid; grid-template-columns: minmax(0, 1fr); gap: 8px"><button @click="w.goMeetings" style="height: 38px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Schedule a meeting</button></div>
</template>

<template v-if="w.t.awards">
<div style="padding: 16px; border-radius: 14px; background: #2A2716; border: 1px solid #4A4426; display: flex; gap: 14px; align-items: center">
<span aria-hidden="true" style="width: 52px; height: 52px; flex: none; border-radius: 50%; background: #7CC4FF; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 17px; box-shadow: 0 0 0 3px #FFD66B">BK</span>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 11px; font-weight: 700; letter-spacing: 0.06em; text-transform: uppercase; color: #FFD66B">Employee of the month</span><strong style="font-size: 16px">Brian Kiprop</strong><span style="font-size: 12px; color: #E3D9AE">Systems Engineer · 99.9% uptime this month</span></div>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(aw, aw__i) in w.items" :key="aw__i">
<div style="display: flex; gap: 10px; align-items: center">
<svg aria-hidden="true" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FFD66B" stroke-width="1.8" stroke-linejoin="round"><circle cx="12" cy="9" r="6"></circle><path d="M8.5 14l-1.5 7 5-3 5 3-1.5-7"></path></svg>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ aw.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ aw.award }}</span></div>
</div>
</template>
</div>
<button @click="w.nominate" style="height: 38px; border-radius: 10px; border: 1px solid #4A4426; background: transparent; color: #FFD66B; font-weight: 600; cursor: pointer">{{ w.nominateLabel }}</button>
</template>

<template v-if="w.t.secureshare">
<span style="font-size: 13px; line-height: 1.45; color: #C9D6CF">Send a password-protected one-time secret to a colleague or copy a share link.</span>
<button @click="w.makeShare" style="height: 40px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Open Secure Share</button>
</template>

<template v-if="w.t.wall">
<div style="display: flex; flex-direction: column; gap: 10px">
<template v-for="(wp, wp__i) in w.items" :key="wp__i">
<div style="display: flex; gap: 10px; padding: 10px; border-radius: 12px; background: #121B17">
<span aria-hidden="true" :style="`width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${wp.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ wp.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 13px">{{ wp.author }}</strong><span style="font-size: 13px; line-height: 1.4; color: #C9D6CF">{{ wp.text }}</span><span style="font-size: 11px; color: #9DB0A6">{{ wp.likes }} likes</span></div>
</div>
</template>
</div>
<button @click="w.goWall" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Open the wall</button>
</template>

<template v-if="w.t.chat">
<div style="display: flex; flex-direction: column; gap: 6px">
<template v-for="(dm, dm__i) in w.items" :key="dm__i">
<button @click="dm.open" style="display: flex; align-items: center; gap: 10px; min-height: 50px; padding: 4px 10px; border-radius: 10px; border: 1px solid #24322C; background: #121B17; color: #E8EFEA; cursor: pointer; text-align: left">
<span :style="`width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${dm.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ dm.initials }}</span>
<span style="flex: 1; min-width: 0; display: flex; flex-direction: column"><span style="font-size: 14px; font-weight: 600">{{ dm.name }}</span><span style="font-size: 12px; color: #9DB0A6; white-space: nowrap; overflow: hidden; text-overflow: ellipsis">{{ dm.last }}</span></span>
<template v-if="dm.hasUnread"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #C9B6FF; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ dm.unread }}</span></template>
</button>
</template>
</div>
<button data-keep="1" @click="toggleNewDm" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">New message</button>
</template>

<template v-if="w.t.forums">
<div style="display: flex; flex-direction: column; gap: 6px">
<template v-for="(fm, fm__i) in w.items" :key="fm__i">
<button @click="fm.open" style="display: flex; align-items: center; gap: 10px; min-height: 44px; padding: 4px 10px; border-radius: 10px; border: 1px solid #24322C; background: #121B17; color: #E8EFEA; cursor: pointer; text-align: left">
<span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: ${fm.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 11px`">{{ fm.initials }}</span>
<span style="flex: 1; min-width: 0; font-size: 14px">{{ fm.name }}</span>
<template v-if="fm.hasUnread"><span style="min-width: 20px; height: 20px; padding: 0 6px; box-sizing: border-box; border-radius: 10px; background: #B8F36B; color: #0F1714; font-size: 11px; font-weight: 700; display: flex; align-items: center; justify-content: center">{{ fm.unread }}</span></template>
</button>
</template>
</div>
</template>

<template v-if="w.t.people">
<label style="display: flex; align-items: center; gap: 8px; height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #9DB0A6"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><circle cx="11" cy="11" r="7"></circle><path d="M20 20l-4-4"></path></svg><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Search people</span><input type="search" placeholder="Name, role or department" style="flex: 1; min-width: 0; border: 0; outline: none; background: transparent; color: #E8EFEA"></label>
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(pp, pp__i) in w.items" :key="pp__i">
<div style="display: flex; gap: 10px; align-items: center">
<span aria-hidden="true" :style="`position: relative; width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${pp.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ pp.initials }}<span style="position: absolute; right: -1px; bottom: -1px; width: 10px; height: 10px; border-radius: 50%; background: #B8F36B; border: 2px solid #16201C"></span></span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ pp.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ pp.role }}</span></div>
</div>
</template>
</div>
<button @click="w.goPeople" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Open directory</button>
</template>

<template v-if="w.t.poll">
<span style="font-size: 14px; font-weight: 500">Where should we hold the end-year party?</span>
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(o, o__i) in w.items" :key="o__i">
<button @click="o.vote" :style="o.style"><span>{{ o.label }}</span><span style="margin-left: auto; font-weight: 600">{{ o.pct }}</span></button>
</template>
</div>
</template>

<template v-if="w.t.notes">
<label><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">My notes</span><textarea rows="5" :value="w.note" @input="w.onNote" placeholder="Jot down reminders for yourself…" style="width: 100%; box-sizing: border-box; padding: 12px; border-radius: 12px; border: 1px solid #4A4426; background: #2A2716; color: #F3EBC8; font-size: 14px; line-height: 1.5; resize: vertical"></textarea></label>
</template>

<template v-if="w.t.countdown">
<div style="display: flex; align-items: baseline; gap: 10px"><span style="font-family: 'Space Grotesk', sans-serif; font-size: 56px; font-weight: 700; line-height: 1; color: #B8F36B">{{ w.days }}</span><span style="color: #C9D6CF">days to go</span></div>
<span style="font-size: 14px">End-year party · Friday 18 December</span>
</template>

<template v-if="w.t.projects">
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px">
<template v-for="(ps, ps__i) in w.pstats" :key="ps__i"><div style="padding: 10px; border-radius: 12px; background: #121B17; display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 22px; font-weight: 700; color: ${ps.color}`">{{ ps.n }}</span><span style="font-size: 11px; color: #C9D6CF">{{ ps.label }}</span></div></template>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(pw, pw__i) in w.items" :key="pw__i">
<div style="padding: 10px 12px; border-radius: 12px; background: #121B17; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; gap: 10px; align-items: center">
<span :style="`width: 30px; height: 30px; flex: none; border-radius: 9px; background: ${pw.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 11px`">{{ pw.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ pw.name }}</strong><span :style="`font-size: 11px; color: ${pw.statusColor}`">{{ pw.status }} · due {{ pw.dueShort }}</span></div>
<button @click="pw.open" :aria-label="`View ${pw.name}`" style="flex: none; height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px; font-weight: 600">View</button>
</div>
<div style="display: flex; align-items: center; gap: 8px"><div style="flex: 1; height: 6px; border-radius: 3px; background: #22302A; overflow: hidden"><div :style="pw.barStyle"></div></div><span style="font-size: 11px; font-weight: 600; width: 34px; text-align: right">{{ pw.progress }}%</span></div>
</div>
</template>
</div>
<button @click="w.goProjects" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">All projects</button>
</template>

<template v-if="w.t.calendar">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(ce, ce__i) in w.items" :key="ce__i">
<div :style="`display: flex; gap: 10px; align-items: center; padding: 8px 10px; border-radius: 10px; background: #121B17; border: 1px solid ${ce.color}`">
<div style="width: 52px; flex: none; display: flex; flex-direction: column"><span style="font-size: 11px; color: #9DB0A6">{{ ce.day }}</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 13px">{{ ce.time }}</span></div>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ ce.title }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ ce.type }}{{ ce.peopleLabel }}</span></div>
</div>
</template>
</div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px"><button @click="w.newEvent" style="height: 38px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">+ New</button><button @click="w.goCalendar" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Open calendar</button></div>
</template>

<template v-if="w.isCommList">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(cm, cm__i) in w.items" :key="cm__i">
<div style="padding: 10px 12px; border-radius: 10px; background: #121B17; display: flex; flex-direction: column; gap: 3px">
<strong style="font-size: 13px; font-weight: 600">{{ cm.title }}</strong>
<span style="font-size: 11px; color: #9DB0A6">{{ cm.by }} · {{ cm.dateLabel }} · to {{ cm.audience }}</span>
<div style="display: flex; align-items: center; gap: 8px"><div style="flex: 1; height: 4px; border-radius: 2px; background: #22302A; overflow: hidden"><div :style="cm.readBar"></div></div><span style="font-size: 10px; color: #9DB0A6">{{ cm.read }}% read</span></div>
</div>
</template>
<template v-if="w.noComms"><span style="font-size: 13px; color: #9DB0A6">Nothing sent yet.</span></template>
</div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 8px"><button @click="w.compose" style="height: 38px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">{{ w.composeLabel }}</button><button @click="w.goComms" style="height: 38px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">See all</button></div>
</template>

<template v-if="w.isOrgList">
<span style="font-size: 12px; color: #9DB0A6">{{ w.orgNote }}</span>
<div style="display: flex; flex-direction: column; gap: 6px; max-height: 300px; overflow-y: auto">
<template v-for="(og, og__i) in w.items" :key="og__i">
<div style="display: flex; gap: 10px; align-items: center; padding: 8px 10px; border-radius: 10px; background: #121B17">
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; font-weight: 600">{{ og.name }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ og.sub }}</span></div>
<span :style="`font-size: 12px; font-weight: 700; color: ${og.perfColor}`">{{ og.perf }}%</span>
<button @click="og.open" :aria-label="`Open ${og.name} dashboard`" style="height: 30px; padding: 0 10px; border-radius: 15px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">View</button>
</div>
</template>
</div>
</template>

<template v-if="w.t.performance">
<span style="font-size: 12px; color: #9DB0A6">Received items resolved on time. Target {{ w.target }}%.</span>
<div style="display: flex; flex-direction: column; gap: 12px">
<template v-for="(pf, pf__i) in w.items" :key="pf__i">
<button @click="pf.open" style="padding: 0; border: 0; background: transparent; color: #E8EFEA; cursor: pointer; text-align: left; display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; justify-content: space-between; gap: 8px; width: 100%; font-size: 13px"><span><span style="color: #9DB0A6">{{ pf.type }} ·</span> {{ pf.name }}</span><strong :style="`color: ${pf.color}`">{{ pf.perf }}%</strong></div>
<div style="position: relative; width: 100%; height: 8px; border-radius: 4px; background: #22302A"><div :style="pf.bar"></div><span :style="pf.marker"></span></div>
</button>
</template>
</div>
<template v-if="w.noUnits"><span style="font-size: 13px; color: #9DB0A6">Add a Departments, Sectors, Branches or Units widget to track performance.</span></template>
</template>

<template v-if="w.t.branches">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(br, br__i) in w.items" :key="br__i">
<button @click="br.open" style="display: flex; gap: 10px; align-items: center; padding: 10px; border-radius: 12px; border: 1px solid #24322C; background: #121B17; color: #E8EFEA; cursor: pointer; text-align: left">
<span style="width: 36px; height: 36px; flex: none; border-radius: 10px; background: #22302A; color: #7CC4FF; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 12px">{{ br.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 14px; font-weight: 600">{{ br.name }} · {{ br.place }}</strong><span style="font-size: 12px; color: #9DB0A6">Head: {{ br.head }} · {{ br.staff }} staff</span></div>
<span :style="`flex: none; font-size: 11px; font-weight: 600; color: ${br.statusColor}`">{{ br.statusLabel }}</span>
</button>
</template>
</div>
</template>

<template v-if="w.t.sectors">
<div style="display: flex; flex-direction: column; gap: 8px">
<template v-for="(sc, sc__i) in w.items" :key="sc__i">
<button @click="sc.open" style="display: flex; gap: 12px; align-items: stretch; padding: 12px; border-radius: 12px; border: 1px solid #24322C; background: #121B17; color: #E8EFEA; cursor: pointer; text-align: left">
<span :style="`width: 4px; flex: none; border-radius: 2px; background: ${sc.color}`"></span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 3px"><strong style="font-size: 14px; font-weight: 600">{{ sc.name }}</strong><span style="font-size: 12px; color: #C9D6CF; line-height: 1.4">{{ sc.depts }}</span></div>
<span style="flex: none; align-self: center; font-family: 'Space Grotesk', sans-serif; font-size: 18px; font-weight: 700">{{ sc.staff }}<span style="font-family: 'DM Sans', sans-serif; font-size: 11px; font-weight: 400; color: #9DB0A6"> staff</span></span>
</button>
</template>
</div>
</template>

<template v-if="w.t.embed">
<div style="height: 140px; border-radius: 12px; border: 1px dashed #4A5C54; display: flex; align-items: center; justify-content: center; text-align: center; padding: 12px; color: #9DB0A6; font-size: 13px">Paste a link to show a dashboard, video, form or another website here.</div>
<div style="display: flex; gap: 8px"><label style="flex: 1; min-width: 0"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Link to embed</span><input type="url" placeholder="https://…" style="width: 100%; box-sizing: border-box; height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA"></label><button style="height: 38px; padding: 0 14px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Embed</button></div>
</template>
</section>
</template>
<template v-if="customizing">
<button v-if="canWidgetsPage" @click="openGallery" style="width: 100%; break-inside: avoid; margin: 0 0 16px; min-height: 160px; border-radius: 18px; border: 2px dashed #3F524A; background: transparent; color: #C9D6CF; cursor: pointer; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; font-size: 14px"><svg aria-hidden="true" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M12 5v14M5 12h14"></path></svg>Add a widget</button>
</template>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'HomeSection',
  ...hubBindings(["awardStrip","awardsOn","canCustomize","canWidgetsPage","customizeBtnStyle","customizeLabel","customizing","noWidgets","openAward","openGallery","secureShare","show","toggleCustomize","toggleNewDm","widgets"])
};
</script>
