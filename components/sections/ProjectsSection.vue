<!-- Projects list and project view -->
<template><template v-if="showProjectList">
<div style="width: 100%; display: flex; flex-direction: column; gap: 18px">
<div style="display: flex; flex-wrap: wrap; align-items: flex-end; gap: 12px">
<div style="flex: 1 1 300px"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 28px; letter-spacing: -0.02em">Projects</h1><p style="margin: 4px 0 0; color: #9DB0A6">Every project's goal, progress, milestones and risks in one place.</p></div>
<div style="display: flex; flex-wrap: wrap; gap: 8px">
<template v-if="canMilestones"><button @click="newMilestone" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">New milestone</button></template>
<template v-if="canCreateProject"><button @click="requestProject" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">New project</button></template>
</div>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 10px">
<template v-for="(ps, ps__i) in projStats" :key="ps__i">
<div style="padding: 14px 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 2px"><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 28px; font-weight: 700; color: ${ps.color}`">{{ ps.n }}</span><span style="font-size: 13px; color: #C9D6CF">{{ ps.label }}</span></div>
</template>
</div>
<div style="display: flex; flex-wrap: wrap; gap: 12px; align-items: center; padding: 12px 16px; border-radius: 14px; background: #16201C; border: 1px solid #24322C">
<span style="width: 34px; height: 34px; border-radius: 9px; background: #22302A; color: #E8EFEA; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 16px">#</span>
<div style="flex: 1 1 260px; display: flex; flex-direction: column"><strong style="font-size: 14px">Slack · connected</strong><span style="font-size: 12px; color: #9DB0A6">Each project posts task moves, overdue reminders and milestones to its own Slack channel.</span></div>
<span style="padding: 4px 10px; border-radius: 12px; background: #26361A; color: #B8F36B; font-size: 12px; font-weight: 600">Active</span>
</div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 14px">
<template v-for="(pr, pr__i) in projects" :key="pr__i">
<article style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; gap: 12px; align-items: center">
<span :style="`width: 42px; height: 42px; flex: none; border-radius: 12px; background: ${pr.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 14px`">{{ pr.initials }}</span>
<div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 16px">{{ pr.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ pr.code }} · Led by {{ pr.owner }}</span></div>
<span :style="pr.statusStyle">{{ pr.status }}</span>
</div>
<p style="margin: 0; font-size: 13px; line-height: 1.5; color: #C9D6CF">{{ pr.goal }}</p>
<div style="display: flex; flex-direction: column; gap: 6px">
<div style="display: flex; justify-content: space-between; font-size: 12px"><span style="color: #9DB0A6">Progress</span><strong>{{ pr.progress }}%</strong></div>
<div style="height: 8px; border-radius: 4px; background: #22302A; overflow: hidden"><div :style="pr.barStyle"></div></div>
</div>
<div style="display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 8px; font-size: 12px">
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Next milestone</span><strong style="font-weight: 600">{{ pr.nextMilestone }}</strong></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Due</span><strong style="font-weight: 600">{{ pr.dueLabel }}</strong></div>
<div style="display: flex; flex-direction: column; gap: 2px"><span style="color: #9DB0A6">Open risks</span><strong :style="`font-weight: 600; color: ${pr.riskColor}`">{{ pr.riskCount }}</strong></div>
</div>
<div style="display: flex; align-items: center; gap: 10px">
<div style="display: flex">
<template v-for="(tm, tm__i) in pr.team" :key="tm__i"><span aria-hidden="true" :style="`width: 30px; height: 30px; margin-right: -8px; border-radius: 50%; border: 2px solid #16201C; background: ${tm.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 10px`">{{ tm.initials }}</span></template>
</div>
<span style="font-size: 12px; color: #9DB0A6; margin-left: 8px">{{ pr.teamSize }} people · {{ pr.sync }}</span>
<button @click="pr.open" style="margin-left: auto; height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">View</button>
</div>
</article>
</template>
</div>
</div>
</template>

<template v-if="showProjectView">
<div style="width: 100%; display: flex; flex-direction: column; gap: 16px">
<button @click="closeProject" style="align-self: flex-start; height: 38px; padding: 0 14px; border-radius: 19px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; gap: 6px"><svg aria-hidden="true" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg>All projects</button>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: center; padding: 20px; border-radius: 20px; background: #16201C; border: 1px solid #24322C">
<span :style="`width: 60px; height: 60px; flex: none; border-radius: 16px; background: ${pj.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-family: 'Space Grotesk', sans-serif; font-weight: 700; font-size: 20px`">{{ pj.initials }}</span>
<div style="flex: 1 1 280px; min-width: 0; display: flex; flex-direction: column; gap: 4px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><h1 style="margin: 0; font-family: 'Space Grotesk', sans-serif; font-size: 26px; letter-spacing: -0.02em">{{ pj.name }}</h1><span :style="pj.statusStyle">{{ pj.status }}</span></div>
<span style="font-size: 13px; color: #9DB0A6">{{ pj.code }} · Led by {{ pj.owner }} · {{ pj.startLabel }} → {{ pj.dueLabel }} · {{ pj.sync }}</span>
</div>
<div style="display: flex; gap: 8px; flex-wrap: wrap">
<template v-if="has.sessions"><button @click="pj.call" style="height: 40px; padding: 0 16px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer">Start standup</button></template>
<button style="height: 40px; padding: 0 16px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Share status update</button>
</div>
</div>

<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(180px, 1fr)); gap: 12px">
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Overall progress</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.progress }}%</span><div style="height: 6px; border-radius: 3px; background: #22302A; overflow: hidden"><div :style="pj.barStyle"></div></div></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Tasks done</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.doneCount }}<span style="font-size: 16px; color: #9DB0A6"> / {{ pj.taskCount }}</span></span><span style="font-size: 12px; color: #9DB0A6">{{ pj.inFlight }} in progress or review</span></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Time left</span><span style="font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700">{{ pj.daysLeft }}<span style="font-size: 16px; color: #9DB0A6"> days</span></span><span style="font-size: 12px; color: #9DB0A6">{{ pj.timeUsed }}% of the timeline used</span></div>
<div style="padding: 16px; border-radius: 16px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 8px"><span style="font-size: 12px; color: #9DB0A6">Budget used</span><span :style="`font-family: 'Space Grotesk', sans-serif; font-size: 30px; font-weight: 700; color: ${pj.budgetColor}`">{{ pj.budget }}%</span><span style="font-size: 12px; color: #9DB0A6">{{ pj.budgetNote }}</span></div>
</div>

<div style="padding: 18px 20px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 16px">
<div style="display: flex; flex-wrap: wrap; justify-content: space-between; align-items: center; gap: 10px"><div style="display: flex; flex-direction: column; gap: 2px"><strong style="font-size: 15px">Milestones</strong><span style="font-size: 12px; color: #9DB0A6">{{ pj.msDone }} of {{ pj.msCount }} reached. Progress comes from the tasks under each objective.</span></div>
<template v-if="pj.canMilestones"><button @click="pj.addMilestone" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer; font-size: 13px">Add milestone</button></template></div>
<div style="display: grid; grid-template-columns: repeat(auto-fill, minmax(320px, 1fr)); gap: 12px">
<template v-for="(ms, ms__i) in pj.milestones" :key="ms__i">
<article :style="ms.cardStyle">
<div style="display: flex; flex-wrap: wrap; align-items: baseline; gap: 6px 10px"><strong style="flex: 1 1 140px; font-size: 14px">{{ ms.name }}</strong><span :style="`font-size: 12px; font-weight: 600; color: ${ms.color}`">{{ ms.stateLabel }}</span></div>
<span style="font-size: 12px; color: #9DB0A6">Due {{ ms.dateLabel }} · {{ ms.countLabel }}</span>
<div style="display: flex; align-items: center; gap: 8px"><div style="flex: 1; height: 8px; border-radius: 4px; background: #22302A; overflow: hidden"><div :style="ms.barStyle"></div></div><strong style="font-size: 12px; min-width: 34px; text-align: right">{{ ms.progress }}%</strong></div>
<div style="display: flex; flex-direction: column; gap: 8px">
<span style="font-size: 11px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Objectives</span>
<template v-for="(ob, ob__i) in ms.objectives" :key="ob__i">
<div style="display: flex; flex-direction: column; gap: 4px"><div style="display: flex; justify-content: space-between; gap: 8px; font-size: 13px"><span>{{ ob.title }}</span><span style="color: #9DB0A6; white-space: nowrap">{{ ob.label }}</span></div><div style="height: 5px; border-radius: 3px; background: #22302A; overflow: hidden"><div :style="ob.barStyle"></div></div></div>
</template>
<template v-if="ms.canAddObj"><label style="display: flex; gap: 6px"><span style="position: absolute; width: 1px; height: 1px; overflow: hidden; clip: rect(0 0 0 0)">New objective for {{ ms.name }}</span><input type="text" :value="ms.objDraft" @input="ms.onObjDraft" @keydown="ms.objKey" placeholder="Add an objective (Enter)" style="flex: 1; min-width: 0; box-sizing: border-box; height: 34px; padding: 0 10px; border-radius: 9px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 13px"><button type="button" @click="ms.addObj" style="height: 34px; padding: 0 12px; border-radius: 9px; border: 1px dashed #4A5C54; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px">Add</button></label></template>
</div>
<template v-if="ms.closed"><div style="padding: 10px 12px; border-radius: 10px; background: #1B2A1E; display: flex; flex-direction: column; gap: 4px"><span style="font-size: 12px; color: #B8F36B; font-weight: 600">{{ ms.closedLine }}</span><span style="font-size: 13px; line-height: 1.45; color: #C9D6CF">"{{ ms.closedComment }}"</span></div></template>
<template v-if="ms.hasHint"><span style="font-size: 12px; color: #9DB0A6">{{ ms.closeHint }}</span></template>
<div style="display: flex; gap: 8px; flex-wrap: wrap">
<template v-if="ms.canClose"><button @click="ms.close" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 13px">Mark as reached</button></template>
<template v-if="ms.canReopen"><button @click="ms.reopen" style="height: 36px; padding: 0 16px; border-radius: 18px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; font-weight: 600; cursor: pointer; font-size: 13px">Reopen</button></template>
</div>
</article>
</template>
</div>
</div>

<div style="display: flex; flex-wrap: wrap; gap: 16px; align-items: flex-start">
<div style="flex: 999 1 560px; min-width: 0; padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 14px">
<div style="display: flex; flex-wrap: wrap; gap: 10px; align-items: center"><strong style="flex: 1; font-size: 15px">Task board</strong>
<label style="display: flex; align-items: center; gap: 8px; font-size: 12px; color: #9DB0A6">Sprint<select :value="pj.sprintValue" @change="pj.setSprint" aria-label="Filter tasks by sprint" style="height: 38px; max-width: 260px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #121B17; color: #E8EFEA; font-size: 13px"><template v-for="(sp, sp__i) in pj.sprintOpts" :key="sp__i"><option :value="sp.v">{{ sp.l }}</option></template></select></label>
<template v-if="pj.canNewTask"><button @click="pj.openTaskForm" style="height: 38px; padding: 0 18px; border-radius: 19px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer; font-size: 14px">New task</button></template>
</div>
<template v-if="pj.taskForm.open">
<form novalidate @submit.prevent="pj.taskForm.submit" aria-label="New task" style="padding: 16px; border-radius: 14px; background: #121B17; border: 1px solid #B8F36B; display: flex; flex-direction: column; gap: 12px">
<strong style="font-size: 14px">New task</strong>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Title<input type="text" :value="pj.taskForm.title" @input="pj.taskForm.onTitle" placeholder="What needs to be done?" style="height: 40px; padding: 0 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Description<textarea rows="3" :value="pj.taskForm.desc" @input="pj.taskForm.onDesc" placeholder="Add the details someone needs to do this task" style="padding: 10px 12px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px; line-height: 1.45; resize: vertical"></textarea></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Objective<select :value="pj.taskForm.objective" @change="pj.taskForm.onObjective" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><option value="">Choose an objective…</option><template v-for="(oo, oo__i) in pj.taskForm.objOpts" :key="oo__i"><option :value="oo.v">{{ oo.l }}</option></template></select></label>
<div style="display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 12px">
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Owner<select :value="pj.taskForm.owner" @change="pj.taskForm.onOwner" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"><template v-for="(ow, ow__i) in pj.taskForm.ownerOpts" :key="ow__i"><option :value="ow.v">{{ ow.l }}</option></template></select></label>
<label style="display: flex; flex-direction: column; gap: 6px; font-size: 12px; color: #9DB0A6">Due date<input type="date" :value="pj.taskForm.due" :min="pj.taskForm.minDue" :max="pj.taskForm.maxDue" @input="pj.taskForm.onDue" style="height: 40px; padding: 0 10px; border-radius: 10px; border: 1px solid #2A3832; background: #16201C; color: #E8EFEA; font-size: 14px"></label>
</div>
<div style="padding: 10px 12px; border-radius: 10px; background: #16201C; display: flex; flex-direction: column; gap: 4px"><span style="font-size: 12px; color: #C9D6CF">{{ pj.taskForm.planLine }}</span><template v-if="pj.taskForm.lateWarn"><span style="font-size: 12px; color: #FFB86B">{{ pj.taskForm.warnLine }}</span></template></div>
<template v-if="pj.taskForm.noObjectives"><span role="alert" style="font-size: 13px; color: #FFB86B">No open milestone has an objective yet. Add a milestone or an objective first.</span></template>
<template v-if="pj.taskForm.hasErr"><span role="alert" style="font-size: 13px; color: #FF9B8A">{{ pj.taskForm.err }}</span></template>
<div style="display: flex; gap: 8px"><button type="submit" style="height: 40px; padding: 0 20px; border-radius: 20px; border: 0; background: #B8F36B; color: #0F1714; font-weight: 600; cursor: pointer">Create task</button><button type="button" @click="pj.taskForm.cancel" style="height: 40px; padding: 0 18px; border-radius: 20px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer">Cancel</button></div>
</form>
</template>
<div style="display: grid; grid-template-columns: repeat(4, minmax(170px, 1fr)); gap: 10px; overflow-x: auto">
<template v-for="(col, col__i) in pj.columns" :key="col__i">
<div style="min-width: 170px; padding: 10px; border-radius: 14px; background: #121B17; display: flex; flex-direction: column; gap: 8px">
<div style="display: flex; align-items: center; gap: 8px; padding: 2px 4px 4px"><span :style="`width: 8px; height: 8px; border-radius: 50%; background: ${col.color}`"></span><strong style="font-size: 13px; flex: 1">{{ col.name }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ col.count }}</span></div>
<template v-for="(tk2, tk2__i) in col.tasks" :key="tk2__i">
<div :style="tk2.cardStyle">
<button @click="tk2.open" style="padding: 0; border: 0; background: transparent; color: #E8EFEA; text-align: left; cursor: pointer; font-size: 13px; font-weight: 500; line-height: 1.35">{{ tk2.title }}</button>
<div style="display: flex; flex-wrap: wrap; gap: 4px"><span :style="tk2.dueStyle">{{ tk2.dueLabel }}</span><span :title="tk2.ctxTitle" style="padding: 2px 7px; border-radius: 6px; background: #1B2A33; color: #7CC4FF; font-size: 10px; font-weight: 600">{{ tk2.ctxLabel }}</span><template v-if="tk2.hasLinks"><span style="padding: 2px 7px; border-radius: 6px; background: #22302A; color: #C9D6CF; font-size: 10px; font-weight: 600">{{ tk2.linkLabel }}</span></template></div>
<div style="display: flex; align-items: center; gap: 4px">
<span :title="`Primary owner: ${tk2.primary}`" :style="`width: 24px; height: 24px; flex: none; border-radius: 50%; background: ${tk2.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 9px`">{{ tk2.initials }}</span>
<template v-if="tk2.hasSecondary"><span :title="`Secondary owner: ${tk2.secondary}`" :style="`width: 18px; height: 18px; flex: none; margin-left: -8px; border-radius: 50%; border: 2px solid #1D2A25; background: ${tk2.secColor}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 7px`">{{ tk2.secInitials }}</span></template>
<span style="flex: 1"></span>
<template v-if="tk2.canBack"><button :aria-label="tk2.backLabel" :title="tk2.backLabel" @click="tk2.back" style="width: 28px; height: 28px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M15 6l-6 6 6 6"></path></svg></button></template>
<template v-if="tk2.canMove"><button :aria-label="tk2.moveLabel" :title="tk2.moveLabel" @click="tk2.move" style="width: 28px; height: 28px; border-radius: 8px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; display: flex; align-items: center; justify-content: center"><svg aria-hidden="true" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M9 6l6 6-6 6"></path></svg></button></template>
</div>
</div>
</template>
</div>
</template>
</div>
</div>

<aside aria-label="Project overview" style="flex: 1 1 320px; min-width: 0; display: flex; flex-direction: column; gap: 16px">
<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Why this project</span>
<p style="margin: 0; font-size: 14px; line-height: 1.55">{{ pj.goal }}</p>
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6; margin-top: 6px">Team · {{ pj.teamCount }}</span>
<div style="max-height: 200px; overflow-y: auto; display: flex; flex-direction: column; gap: 8px; padding-right: 4px">
<template v-for="(tm, tm__i) in pj.team" :key="tm__i"><div style="display: flex; gap: 10px; align-items: center"><span aria-hidden="true" :style="`width: 30px; height: 30px; border-radius: 50%; background: ${tm.color}; color: #0F1714; display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 11px`">{{ tm.initials }}</span><span style="font-size: 14px; flex: 1">{{ tm.name }}</span><span style="font-size: 12px; color: #9DB0A6">{{ tm.load }} tasks</span></div></template>
</div>
</div>

<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Risks &amp; issues</span>
<template v-for="(rk, rk__i) in pj.risks" :key="rk__i"><div style="display: flex; gap: 10px; align-items: flex-start; padding: 10px; border-radius: 10px; background: #121B17"><span :style="rk.style">{{ rk.level }}</span><span style="font-size: 13px; line-height: 1.45">{{ rk.text }}</span></div></template>
</div>

<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Linked tickets</span>
<template v-if="pj.noTickets"><span style="font-size: 13px; color: #9DB0A6">No tickets linked yet.</span></template>
<template v-for="(lt, lt__i) in pj.linked" :key="lt__i"><div style="display: flex; gap: 10px; align-items: center"><div style="flex: 1; min-width: 0; display: flex; flex-direction: column"><strong style="font-size: 13px; font-weight: 600; overflow: hidden; text-overflow: ellipsis; white-space: nowrap">{{ lt.id }} · {{ lt.title }}</strong><span style="font-size: 12px; color: #9DB0A6">{{ lt.statusLabel }} · <span :style="`color: ${lt.slaColor}`">{{ lt.slaLabel }}</span></span></div><button @click="lt.open" style="height: 32px; padding: 0 12px; border-radius: 16px; border: 1px solid #2A3832; background: transparent; color: #E8EFEA; cursor: pointer; font-size: 12px; font-weight: 600">View</button></div></template>
</div>

<div style="padding: 18px; border-radius: 18px; background: #16201C; border: 1px solid #24322C; display: flex; flex-direction: column; gap: 10px">
<span style="font-size: 12px; font-weight: 600; letter-spacing: 0.06em; text-transform: uppercase; color: #9DB0A6">Recent activity</span>
<div style="max-height: 230px; overflow-y: auto; display: flex; flex-direction: column; gap: 10px; padding-right: 4px">
<template v-for="(pa, pa__i) in pj.activity" :key="pa__i"><div style="display: flex; gap: 10px; align-items: flex-start"><span style="width: 6px; height: 6px; margin-top: 7px; flex: none; border-radius: 50%; background: #B8F36B"></span><span style="font-size: 13px; line-height: 1.45; color: #C9D6CF">{{ pa.text }} <span style="color: #9DB0A6">· {{ pa.time }}</span></span></div></template>
</div>
</div>
</aside>
</div>
</div>
</template>
</template>

<script>
import { hubBindings } from '~/composables/useHub';

export default {
  name: 'ProjectsSection',
  ...hubBindings(["canCreateProject","canMilestones","closeProject","has","newMilestone","pj","projStats","projects","requestProject","showProjectList","showProjectView"])
};
</script>
