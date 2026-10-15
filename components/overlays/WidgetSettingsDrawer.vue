<!-- Widget settings drawer -->
<template><template v-if="settingsOpen">
<div style="position: fixed; inset: 0; z-index: 65; background: rgba(5,10,8,0.6); display: flex; justify-content: flex-end">
<div role="dialog" aria-label="Widget settings" style="width: 540px; max-width: 100%; height: 100%; background: #16201C; border-left: 1px solid #2A3832; box-shadow: -20px 0 50px rgba(0,0,0,0.4); display: flex; flex-direction: column">
<div style="flex: none; display: flex; align-items: center; gap: 12px; padding: 18px 20px; border-bottom: 1px solid #24322C">
<span :style="`width: 12px; height: 12px; border-radius: 50%; background: ${sw.tint}`"></span>
<div style="flex: 1; min-width: 0"><span style="font-size: 12px; color: #9DB0A6">Widget settings</span><h2 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 22px">{{ sw.baseTitle }}</h2></div>
<button aria-label="Close settings" @click="closeSettings" style="width: 40px; height: 40px; border-radius: 50%; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
<div style="flex: 1; min-height: 0; overflow-y: auto; padding: 20px; display: flex; flex-direction: column; gap: 22px">

<div style="display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">General</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Display name<input type="text" :value="sw.title" @input="sw.setTitle" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: #C9D6CF">Who can see it</span>
<div role="radiogroup" aria-label="Who can see it" style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(so, so__i) in sw.seeOpts" :key="so__i"><button role="radio" :aria-checked="so.on" @click="so.pick" :style="so.style">{{ so.label }}</button></template></div>
<template v-if="sw.isGroups">
<div :style="`padding: 12px; border-radius: 12px; background: #121B17; border: 1px solid ${sw.groupsBorder}; display: flex; flex-direction: column; gap: 10px`">
<span :style="`font-size: 12px; color: ${sw.groupsHintColor}`">{{ sw.groupsHint }}</span>
<template v-for="(gs, gs__i) in sw.groupSets" :key="gs__i"><div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: #9DB0A6">{{ gs.label }}</span><div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(gi, gi__i) in gs.items" :key="gi__i"><button :aria-pressed="gi.on" @click="gi.toggle" :style="gi.style">{{ gi.name }}</button></template></div></div></template>
</div>
</template>
</div>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 13px; color: #C9D6CF">Who can manage it <span style="color: #9DB0A6">(roles)</span></span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(mr2, mr2__i) in sw.managerRoles" :key="mr2__i"><button :aria-pressed="mr2.on" @click="mr2.toggle" :disabled="mr2.locked" :style="mr2.style">{{ mr2.name }}</button></template></div>
<span style="font-size: 11px; color: #9DB0A6">Super admin can always manage every widget.</span>
</div>
</div>

<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Rights in this widget</span>
<span style="font-size: 13px; color: #C9D6CF; line-height: 1.5">Tick what each role may do here. People can only act within these rights.</span>
<div style="overflow-x: auto; border-radius: 12px; border: 1px solid #24322C">
<table style="width: 100%; border-collapse: collapse; font-size: 13px">
<thead><tr><th style="text-align: left; padding: 10px 12px; background: #121B17; color: #9DB0A6; font-weight: 600; font-size: 12px">Role</th><template v-for="(rn, rn__i) in sw.rightNames" :key="rn__i"><th style="padding: 10px 6px; background: #121B17; color: #9DB0A6; font-weight: 600; font-size: 11px; text-align: center; min-width: 64px">{{ rn.name }}</th></template></tr></thead>
<tbody><template v-for="(mr, mr__i) in sw.matrix" :key="mr__i"><tr style="border-top: 1px solid #24322C"><td style="padding: 8px 12px; white-space: nowrap">{{ mr.role }}</td><template v-for="(mc, mc__i) in mr.cells" :key="mc__i"><td style="text-align: center; padding: 6px"><input type="checkbox" :aria-label="mc.label" :checked="mc.on" :disabled="mc.locked" @change="mc.toggle"><template v-if="mc.pending"><div style="font-size: 9px; font-weight: 700; color: #FFB86B">PENDING</div></template></td></template></tr></template></tbody>
</table>
</div>
</div>

<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Approvals</span>
<template v-if="sw.noFlows"><span style="font-size: 13px; color: #9DB0A6">Nothing done in this widget needs approval.</span></template>
<template v-for="(fw, fw__i) in sw.flows" :key="fw__i">
<div style="padding: 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; gap: 10px; align-items: center; flex-wrap: wrap"><strong style="flex: 1 1 200px; font-size: 14px; font-weight: 600">{{ fw.action }}</strong>
<select :aria-label="`Approval levels for ${fw.action}`" :value="fw.levelsVal" @change="fw.setLevels" style="height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><option value="0">No approval</option><option value="1">Direct (1 level)</option><option value="2">2 levels</option><option value="3">3 levels</option></select></div>
<template v-if="fw.hasLevels"><div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 8px"><template v-for="(lv, lv__i) in fw.levels" :key="lv__i"><label style="display: flex; flex-direction: column; gap: 4px; font-size: 11px; color: #9DB0A6">Level {{ lv.n }} approver<select :value="lv.who" @change="lv.set" style="height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><option>Line manager</option><option>Head of department</option><option>Finance approver</option><option>HR Manager</option><option>Communications lead</option><option>Security officer</option><option>Head of IT Operations</option><option>Managing Director</option></select></label></template></div></template>
</div>
</template>
</div>

<template v-if="sw.isLevel">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Communication rules</span>
<span style="font-size: 13px; color: #C9D6CF; line-height: 1.5">Decide whether each {{ sw.levelSingular }} can send memos and announcements, where they can go, which role sends them and which role approves.</span>
<template v-for="(rl, rl__i) in sw.commRules" :key="rl__i">
<div style="padding: 14px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<label style="display: flex; gap: 10px; align-items: center; cursor: pointer"><input type="checkbox" :checked="rl.on" @change="rl.toggle"><strong style="font-size: 14px">{{ rl.label }}</strong></label>
<template v-if="rl.on">
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 12px; color: #9DB0A6">Can send to</span><div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(tg, tg__i) in rl.targets" :key="tg__i"><button :aria-pressed="tg.on" @click="tg.toggle" :style="tg.style">{{ tg.label }}</button></template></div></div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(170px, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #9DB0A6">Role that sends<select :value="rl.sender" @change="rl.setSender" style="height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><template v-for="(ro2, ro2__i) in rl.roleOpts" :key="ro2__i"><option :value="ro2.v">{{ ro2.l }}</option></template></select></label>
<label style="display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #9DB0A6">Role that approves<select :value="rl.approver" @change="rl.setApprover" style="height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><option value="none">No approval needed</option><template v-for="(ro3, ro3__i) in rl.roleOpts" :key="ro3__i"><option :value="ro3.v">{{ ro3.l }}</option></template></select></label>
</div>
</template>
</div>
</template>
</div>
</template>

<template v-if="sw.t.tickets">
<div style="display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Categories and TAT</span>
<span style="font-size: 13px; color: #C9D6CF; line-height: 1.5">Each category has a turnaround time. A ticket must be solved within it or it goes overdue.</span>
<template v-for="(tc, tc__i) in tsetCats" :key="tc__i">
<div style="display: grid; grid-template-columns: minmax(0, 1fr) 72px 132px 36px; gap: 8px; align-items: center; padding: 8px 10px; border-radius: 12px; background: #121B17; border: 1px solid #24322C">
<span style="font-size: 14px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ tc.name }}</span>
<input type="number" min="1" :aria-label="`TAT for ${tc.name}`" :value="tc.tat" @input="tc.setTat" style="height: 36px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px">
<select :aria-label="`Unit for ${tc.name}`" :value="tc.unit" @change="tc.setUnit" style="height: 36px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><option value="hours">hours</option><option value="days">working days</option></select>
<button :aria-label="`Remove ${tc.name}`" @click="tc.remove" style="width: 36px; height: 36px; border-radius: 8px; border: 0; background: transparent; color: #FF9B8A; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><path d="M6 6l12 12M18 6L6 18"></path></svg></button>
</div>
</template>
<div style="display: grid; grid-template-columns: minmax(0, 1fr) 72px 132px; gap: 8px">
<input type="text" aria-label="New category name" placeholder="New category" :value="tset.newCat" @input="setNewCat" style="height: 38px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<input type="number" min="1" aria-label="New category TAT" placeholder="TAT" :value="tset.newTat" @input="setNewTat" style="height: 38px; padding: 0 8px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<select aria-label="New category unit" :value="tset.newUnit" @change="setNewUnit" style="height: 38px; padding: 0 6px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"><option value="hours">hours</option><option value="days">working days</option></select>
</div>
<button @click="addCat" style="align-self: flex-start; height: 38px; padding: 0 16px; border-radius: 10px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">+ Add category</button>
</div>

<div style="display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Working hours</span>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Day starts<input type="time" :value="tset.start" @input="setStart" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; color-scheme: dark"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Day ends<input type="time" :value="tset.end" @input="setEnd" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; color-scheme: dark"></label>
</div>
<span style="font-size: 12px; color: #9DB0A6">One working day = {{ dayHoursLabel }} of TAT.</span>
</div>

<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">TAT does not count during</span>
<template v-for="(pr, pr__i) in pauseRules" :key="pr__i">
<label style="display: flex; gap: 12px; align-items: flex-start; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :checked="pr.checked" @change="pr.toggle"><span style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ pr.label }}</span><span style="font-size: 12px; color: #9DB0A6">{{ pr.desc }}</span></span></label>
</template>
</div>

<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Notifications</span>
<template v-for="(nr, nr__i) in notifyRules" :key="nr__i">
<label style="display: flex; gap: 12px; align-items: flex-start; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :checked="nr.checked" @change="nr.toggle"><span style="display: flex; flex-direction: column; gap: 2px"><span style="font-size: 14px">{{ nr.label }}</span><span style="font-size: 12px; color: #9DB0A6">{{ nr.desc }}</span></span></label>
</template>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Warn when this much of the TAT is used: {{ tset.nearPct }}%<input type="range" min="50" max="95" step="5" :value="tset.nearPct" @input="setNearPct"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Escalate overdue tickets to<select :value="tset.escalate" @change="setEscalate" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option value="Head of IT Operations">Head of IT Operations</option><option value="Head of Software Development">Head of Software Development</option><option value="HR Manager">HR Manager</option><option value="The requester's department head">The requester's department head</option></select></label>
</div>
</template>

<template v-if="sw.t.announcements">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Announcements</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Who can post<select defaultValue="HODs and admins" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>Admins only</option><option>HODs and admins</option><option>Everyone</option></select></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Ask staff to mark each announcement as read</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Send a notification for every new announcement</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox"><span style="font-size: 14px">Expire announcements after 30 days</span></label>
</div>
</template>

<template v-if="sw.t.awards">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Awards</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><span style="padding: 6px 12px; border-radius: 16px; background: #2A2716; border: 1px solid #4A4426; font-size: 13px">Employee of the month</span><span style="padding: 6px 12px; border-radius: 16px; background: #2A2716; border: 1px solid #4A4426; font-size: 13px">Star performer</span><span style="padding: 6px 12px; border-radius: 16px; background: #2A2716; border: 1px solid #4A4426; font-size: 13px">Customer hero</span><span style="padding: 6px 12px; border-radius: 16px; background: #2A2716; border: 1px solid #4A4426; font-size: 13px">Team of the quarter</span></div>
<div style="display: flex; gap: 8px"><input type="text" aria-label="New award" placeholder="Add an award, e.g. Innovation" style="flex: 1; min-width: 0; height: 38px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><button style="height: 38px; padding: 0 14px; border-radius: 8px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">Add</button></div>
<div style="display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">How often<select defaultValue="Monthly" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>Monthly</option><option>Quarterly</option><option>Yearly</option></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Who approves<select defaultValue="HR" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>HR</option><option>Management</option><option>HODs</option></select></label>
</div>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Nominations are open to all staff</span></label>
</div>
</template>

<template v-if="sw.t.secureshare">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Secure share</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Link expires<select defaultValue="After 1 view" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>After 1 view</option><option>After 1 hour</option><option>After 24 hours</option><option>After 7 days</option></select></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Always require a password</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Notify the sender when the link is opened</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox"><span style="font-size: 14px">Allow sharing files, not just text</span></label>
</div>
</template>

<template v-if="sw.t.sessions">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Meetings &amp; standups</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Most people in one meeting<select defaultValue="100" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>25</option><option>100</option><option>300</option></select></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Allow recording</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox"><span style="font-size: 14px">Use a waiting room</span></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox"><span style="font-size: 14px">Allow guests from outside the organization</span></label>
</div>
</template>

<template v-if="sw.t.clocks">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Cities shown</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><span style="padding: 6px 12px; border-radius: 16px; background: #121B17; border: 1px solid #2A3832; font-size: 13px">Nairobi</span><span style="padding: 6px 12px; border-radius: 16px; background: #121B17; border: 1px solid #2A3832; font-size: 13px">London</span><span style="padding: 6px 12px; border-radius: 16px; background: #121B17; border: 1px solid #2A3832; font-size: 13px">New York</span><span style="padding: 6px 12px; border-radius: 16px; background: #121B17; border: 1px solid #2A3832; font-size: 13px">Dubai</span></div>
<div style="display: flex; gap: 8px"><input type="text" aria-label="Add a city" placeholder="Add a city" style="flex: 1; min-width: 0; height: 38px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><button style="height: 38px; padding: 0 14px; border-radius: 8px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">Add</button></div>
</div>
</template>

<template v-if="sw.t.branches">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Branches</span>
<div style="display: flex; gap: 8px"><input type="text" aria-label="Branch name" placeholder="Branch name and city" style="flex: 1; min-width: 0; height: 38px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><button style="height: 38px; padding: 0 14px; border-radius: 8px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">Add branch</button></div>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Show open or closed status</span></label>
</div>
</template>

<template v-if="sw.t.sectors">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Sectors</span>
<div style="display: flex; gap: 8px"><input type="text" aria-label="Sector name" placeholder="Sector name" style="flex: 1; min-width: 0; height: 38px; padding: 0 10px; border-radius: 8px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><button style="height: 38px; padding: 0 14px; border-radius: 8px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer">Add sector</button></div>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Show staff count per sector</span></label>
</div>
</template>

<template v-if="sw.genericOnly">
<div style="display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Display</span>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Items to show<select defaultValue="4" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option>3</option><option>4</option><option>6</option><option>10</option></select></label>
<label style="display: flex; gap: 12px; align-items: center; padding: 10px 12px; border-radius: 12px; background: #121B17; border: 1px solid #24322C; cursor: pointer"><input type="checkbox" :defaultChecked="true"><span style="font-size: 14px">Show a "View all" link</span></label>
</div>
</template>
</div>
<div style="flex: none; display: flex; justify-content: flex-end; gap: 8px; padding: 14px 20px; border-top: 1px solid #24322C">
<button @click="closeSettings" style="height: 42px; padding: 0 18px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button>
<button @click="saveSettings" style="height: 42px; padding: 0 22px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Save settings</button>
</div>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'WidgetSettingsDrawer',
  ...hubBindings(["addCat","closeSettings","dayHoursLabel","notifyRules","pauseRules","saveSettings","setEnd","setEscalate","setNearPct","setNewCat","setNewTat","setNewUnit","setStart","settingsOpen","sw","tset","tsetCats"])
};
</script>
