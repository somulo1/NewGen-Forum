<!-- Tickets list and ticket details -->
<template><template v-if="showTicketList">
<div style="max-width: 1120px; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Tickets</h1><p style="margin: 4px 0 0; color: #9DB0A6">Raise, track and solve requests within their turnaround time.</p></div>
<template v-if="canRaise"><button @click="toggleNewTicket" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">New ticket</button></template>
<div style="flex: 1 1 100%; display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px">
<template v-for="(ts, ts__i) in ticketStats" :key="ts__i">
<div style="padding: 14px 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; color: ${ts.color}`">{{ ts.n }}</span><span style="font-size: 13px; color: #C9D6CF">{{ ts.label }}</span></div>
</template>
</div>
<template v-if="ntOpen">
<div style="flex: 1 1 100%; padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #B8F36B; display: flex; flex-direction: column; gap: 14px">
<strong style="font-size: 16px">Raise a new ticket</strong>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">What's the issue?<input type="text" :value="nt.title" @input="ntSetTitle" placeholder="e.g. Printer on 3rd floor not printing" style="height: 42px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"></label>
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: #C9D6CF">Category (sets the TAT)</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(nc, nc__i) in ntCats" :key="nc__i"><button @click="nc.pick" :style="nc.style">{{ nc.label }}</button></template></div></div>
<template v-if="has.departments"><div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: #C9D6CF">Send to department</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(nd, nd__i) in ntDepts" :key="nd__i"><button @click="nd.pick" :style="nd.style">{{ nd.label }}</button></template></div></div></template>
<div style="display: flex; flex-direction: column; gap: 6px"><span style="font-size: 13px; color: #C9D6CF">Priority</span><div style="display: flex; flex-wrap: wrap; gap: 8px"><template v-for="(np, np__i) in ntPrios" :key="np__i"><button @click="np.pick" :style="np.style">{{ np.label }}</button></template></div></div>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Details<textarea rows="3" :value="nt.desc" @input="ntSetDesc" placeholder="Describe what happened and where" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; resize: vertical"></textarea></label>
<div style="display: flex; gap: 8px; justify-content: flex-end"><button @click="toggleNewTicket" style="height: 40px; padding: 0 16px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button><button @click="createTicket" :style="ntBtnStyle">Raise ticket</button></div>
</div>
</template>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<span style="padding: 6px 12px; border-radius: 16px; background: #1A2520; border: 1px solid #2A3832; font-size: 13px">Jira · connected</span>
<span style="padding: 6px 12px; border-radius: 16px; background: #1A2520; border: 1px solid #2A3832; font-size: 13px">GitHub · connected</span>
<span style="padding: 6px 12px; border-radius: 16px; background: #1A2520; border: 1px solid #2A3832; font-size: 13px">Trello · connected</span>
<button style="padding: 6px 12px; border-radius: 16px; background: transparent; border: 1px dashed #4A5C54; color: #C9D6CF; font-size: 13px; cursor: pointer">+ Connect Asana, ClickUp…</button>
</div>
<DataTable :t="ticketsTable" />
<div style="display: none">
<table style="width: 100%; min-width: 720px; border-collapse: collapse; font-size: 14px">
<thead><tr style="text-align: left; color: #9DB0A6; font-size: 12px; text-transform: uppercase; letter-spacing: 0.05em"><th style="padding: 12px 16px">ID</th><th style="padding: 12px 16px">Title</th><th style="padding: 12px 16px">Category</th><th style="padding: 12px 16px">Priority</th><th style="padding: 12px 16px">Status</th><th style="padding: 12px 16px">TAT</th><th style="padding: 12px 16px">Handling</th><th style="padding: 12px 16px"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Open</span></th></tr></thead>
<tbody>
<template v-for="(t, t__i) in tickets" :key="t__i">
<tr style="border-top: 1px solid #24322C"><td style="padding: 14px 16px; color: #9DB0A6; font-family: 'Space Grotesk', sans-serif; white-space: nowrap">{{ t.id }}</td><td style="padding: 8px 10px"><button @click="t.open" style="min-height: 40px; padding: 0 6px; border: 0; background: transparent; color: #E8EFEA; font-size: 14px; font-weight: 600; text-align: left; cursor: pointer; text-decoration: underline; text-decoration-color: #3F524A; text-underline-offset: 4px">{{ t.title }}</button></td><td style="padding: 14px 16px; color: #C9D6CF">{{ t.cat }}</td><td style="padding: 14px 16px"><span :style="t.prioStyle">{{ t.prio }}</span></td><td style="padding: 14px 16px; white-space: nowrap">{{ t.statusLabel }}</td><td style="padding: 14px 16px; white-space: nowrap"><span :style="t.slaBadgeStyle">{{ t.slaLabel }}</span><div style="font-size: 11px; color: #9DB0A6; margin-top: 4px">{{ t.remainingLabel }}</div></td><td style="padding: 14px 16px; white-space: nowrap">{{ t.handlerLabel }}</td><td style="padding: 8px 16px 8px 0"><button @click="t.open" :aria-label="`View ${t.id}`" style="height: 34px; padding: 0 14px; border-radius: 17px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px; font-weight: 600">View</button></td></tr>
</template>
</tbody>
</table>
</div>
</div>
</template>

<template v-if="showTicketView">
<div style="max-width: 1180px; display: flex; flex-direction: column; gap: 16px">
<button @click="closeTicket" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 6px"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg>All tickets</button>
<div style="display: flex; flex-wrap: wrap; gap: 10px 14px; align-items: center">
<span style="font-family: 'Space Grotesk', sans-serif; color: #9DB0A6; font-size: 15px">{{ tk.id }}</span>
<h1 style="margin: 0; flex: 1 1 320px; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em; line-height: 1.2">{{ tk.title }}</h1>
<span :style="tk.prioStyle">{{ tk.prio }}</span>
<span :style="tk.slaBadgeStyle">{{ tk.slaLabel }}</span>
</div>

<ol aria-label="Ticket progress" style="list-style: none; margin: 0; padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: grid; grid-template-columns: repeat(5, minmax(0, 1fr)); gap: 10px">
<template v-for="(st, st__i) in tk.steps" :key="st__i">
<li style="display: flex; flex-direction: column; gap: 8px; min-width: 0">
<span :style="st.barStyle"></span>
<span style="display: flex; align-items: center; gap: 8px; font-size: 13px"><span :style="st.dotStyle">{{ st.num }}</span><span :style="st.labelStyle">{{ st.label }}</span></span>
<span style="font-size: 11px; color: #9DB0A6; line-height: 1.35">{{ st.note }}</span>
</li>
</template>
</ol>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 460px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><strong style="font-size: 15px">Description</strong><p style="margin: 0; font-size: 14px; line-height: 1.6; color: #C9D6CF">{{ tk.desc }}</p></div>

<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<strong style="font-size: 15px">Activity</strong>
<template v-for="(ac, ac__i) in tk.activity" :key="ac__i">
<div style="display: flex; gap: 10px">
<span aria-hidden="true" :style="`width: 34px; height: 34px; flex: none; border-radius: 50%; background: ${ac.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 12px`">{{ ac.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; gap: 8px; align-items: baseline; flex-wrap: wrap"><strong style="font-size: 13px">{{ ac.who }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ ac.time }}</span></div>
<template v-if="ac.isEvent"><span style="font-size: 13px; color: #C9D6CF">{{ ac.text }}</span></template>
<template v-if="ac.isComment"><div style="padding: 10px 12px; border-radius: 4px 14px 14px 14px; background: #1F2C26; font-size: 14px; line-height: 1.5; white-space: pre-line">{{ ac.text }}</div></template>
<template v-if="ac.isFile"><div style="align-self: flex-start; display: flex; gap: 10px; align-items: center; padding: 8px 12px; border-radius: 10px; background: #121B17; border: 1px solid #2A3832"><span style="width: 30px; height: 30px; border-radius: 8px; background: #22302A; color: #B8F36B; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"><path d="M6 3h8l4 4v14H6z"></path><path d="M14 3v4h4"></path></svg></span><div style="display: flex; flex-direction: column"><strong style="font-size: 13px">{{ ac.text }}</strong><span style="font-size: 11px; color: #9DB0A6">{{ ac.meta }}</span></div></div></template>
</div>
</div>
</template>

<div style="border-top: 1px solid #24322C; padding-top: 14px; display: flex; flex-direction: column; gap: 10px">
<template v-if="tagOpen">
<div style="display: flex; flex-wrap: wrap; gap: 6px; align-items: center"><span style="font-size: 12px; color: #9DB0A6">Tag:</span><template v-for="(tp, tp__i) in tagPeople" :key="tp__i"><button @click="tp.tag" style="height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; cursor: pointer; font-size: 13px">@{{ tp.name }}</button></template></div>
</template>
<label><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">Comment</span><textarea rows="3" :value="tcomment" @input="onTComment" placeholder="Write a comment. Use Tag to mention someone and notify them." style="width: 100%; box-sizing: border-box; padding: 10px 12px; border-radius: 12px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; resize: vertical"></textarea></label>
<div style="display: flex; flex-wrap: wrap; gap: 8px; align-items: center">
<label style="position: relative; height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 8px; font-size: 13px"><input type="file" multiple="" aria-label="Attach files to ticket" @change="attachTicket" style="position: absolute; inset: 0; width: 100%; height: 100%; opacity: 0; cursor: pointer"><svg aria-hidden="true" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M20 11.5l-8.2 8.2a5 5 0 0 1-7.1-7.1L13.2 4a3.4 3.4 0 0 1 4.8 4.8l-8.5 8.5a1.8 1.8 0 0 1-2.5-2.5L14.8 7"></path></svg>Attach file</label>
<button @click="toggleTag" style="height: 38px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 13px">@ Tag someone</button>
<button @click="postTComment" style="margin-left: auto; height: 38px; padding: 0 18px; border-radius: 10px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Comment</button>
</div>
</div>
</div>
</div>

<aside aria-label="Ticket details" style="flex: 1 1 300px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Turnaround time</span>
<span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 26px; font-weight: 700; color: ${tk.slaColor}`">{{ tk.remainingLabel }}</span>
<div style="height: 8px; border-radius: 4px; background: #22302A; overflow: hidden"><div :style="tk.slaBarStyle"></div></div>
<span style="font-size: 13px; color: #C9D6CF">{{ tk.cat }} TAT: {{ tk.tatLabel }} · {{ tk.usedLabel }} used</span>
<template v-if="tk.paused"><span style="font-size: 12px; color: #7CC4FF">Clock paused while the ticket is on hold.</span></template>
<span style="font-size: 12px; color: #9DB0A6">{{ tk.pauseNote }}</span>
</div>

<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 12px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Currently handling</span>
<template v-if="tk.hasHandler">
<div style="display: flex; gap: 10px; align-items: center"><span aria-hidden="true" :style="`width: 42px; height: 42px; border-radius: 50%; background: ${tk.handlerColor}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700`">{{ tk.handlerInitials }}</span><div style="display: flex; flex-direction: column"><strong style="font-size: 15px">{{ tk.handler }}</strong><span style="font-size: 12px; color: #9DB0A6">Since {{ tk.handlerSince }}</span></div></div>
</template>
<template v-if="tk.noHandler"><span style="font-size: 14px; color: #FFB86B">Not assigned yet</span></template>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">{{ tk.handoverLabel }}
<select :value="tk.handlerValue" @change="handover" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px">
<option value="">Choose a person…</option>
<option value="Sam">Me (Sam)</option>
<option value="Brian Kiprop">Brian Kiprop · IT</option>
<option value="David Otieno">David Otieno · Development</option>
<option value="Amina Wanjiru">Amina Wanjiru · Finance</option>
<option value="Grace Njeri">Grace Njeri · HR</option>
<option value="Faith Achieng">Faith Achieng · Customer Care</option>
<option value="Kevin Mutua">Kevin Mutua · Sales</option>
</select></label>
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; margin-top: 4px">Everyone assigned</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(as, as__i) in tk.assignees" :key="as__i"><span style="height: 30px; padding: 0 10px 0 4px; border-radius: 15px; background: #121B17; border: 1px solid #2A3832; display: flex; align-items: center; gap: 6px; font-size: 12px"><span :style="`width: 22px; height: 22px; border-radius: 50%; background: ${as.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 9px`">{{ as.initials }}</span>{{ as.name }}</span></template></div>
<span style="font-size: 12px; color: #9DB0A6">Raised by {{ tk.raisedBy }}</span>
</div>

<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Actions</span>
<template v-for="(act, act__i) in tk.actions" :key="act__i"><button @click="act.run" :style="act.style">{{ act.label }}</button></template>
<template v-if="tk.noActions"><span style="font-size: 13px; color: #9DB0A6">Assign someone above to start this ticket.</span></template>
</div>

<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: grid; grid-template-columns: auto minmax(0, 1fr); gap: 8px 14px; font-size: 13px">
<span style="color: #9DB0A6">Category</span><span>{{ tk.cat }}</span>
<template v-if="tk.wasEscalated"><span style="color: #9DB0A6">Escalated to</span><span style="color: #FF9B8A">{{ tk.escalatedToLabel }}</span></template>
<span style="color: #9DB0A6">Created</span><span>{{ tk.created }}</span>
<span style="color: #9DB0A6">Synced with</span><span>{{ tk.sync }}</span>
</div>
<template v-if="esc.canEscalate">
<div style="padding: 18px; border-radius: 16px; background: #16201C; border: 1px solid #3F2A26; display: flex; flex-direction: column; gap: 10px">
<div style="display: flex; align-items: center; gap: 8px"><span style="flex: 1; font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #FF9B8A">Escalate</span>
<button @click="esc.toggle" style="height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #5A3A33; background: transparent; color: #FFD3CA; cursor: pointer; font-size: 13px">{{ esc.btnLabel }}</button></div>
<template v-if="esc.open">
<span style="font-size: 13px; color: #C9D6CF">1 · Escalate to</span>
<div style="display: flex; flex-wrap: wrap; gap: 6px"><template v-for="(et, et__i) in esc.types" :key="et__i"><button @click="et.pick" :style="et.style">{{ et.label }}</button></template></div>
<template v-if="esc.showRec">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">2 · Which {{ esc.typeLabel }}
<select :value="esc.recValue" @change="esc.onRec" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option value="">Choose…</option><template v-for="(er, er__i) in esc.recs" :key="er__i"><option :value="er.id">{{ er.name }}</option></template></select></label>
</template>
<template v-if="esc.hasRec">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">3 · Category they handle
<select :value="esc.catValue" @change="esc.onCat" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option value="">Choose a category…</option><template v-for="(ec, ec__i) in esc.cats" :key="ec__i"><option :value="ec.name">{{ ec.name }} · TAT {{ ec.tat }}</option></template></select></label>
</template>
<template v-if="esc.hasCat">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">4 · Person assigned to this category
<select :value="esc.personValue" @change="esc.onPerson" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px"><option value="">Choose a person…</option><template v-for="(ep, ep__i) in esc.people" :key="ep__i"><option :value="ep.name">{{ ep.name }} · {{ ep.title }}</option></template></select></label>
<template v-if="esc.noPeople"><span style="font-size: 12px; color: #FFB86B">Nobody is assigned to this category there yet.</span></template>
</template>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 13px; color: #C9D6CF">Reason (optional)<textarea rows="2" :value="esc.why" @input="esc.onWhy" placeholder="Why is this being escalated?" style="padding: 8px 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 14px; resize: vertical"></textarea></label>
<button @click="esc.submit" :style="esc.btnStyle">Escalate</button>
<span style="font-size: 12px; color: #9DB0A6">The ticket moves to the chosen person, takes the new category's TAT, and the change is logged in the activity.</span>
</template>
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
  name: 'TicketsSection',
  ...hubBindings(["attachTicket","canRaise","closeTicket","createTicket","esc","handover","has","nt","ntBtnStyle","ntCats","ntDepts","ntOpen","ntPrios","ntSetDesc","ntSetTitle","onTComment","postTComment","showTicketList","showTicketView","tagOpen","tagPeople","tcomment","ticketStats","tickets","ticketsTable","tk","toggleNewTicket","toggleTag"])
};
</script>
